package client

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xackery/launcheq/config"
	"github.com/xackery/launcheq/gui"
	"gopkg.in/yaml.v3"

	"github.com/fynelabs/selfupdate"
)

// Client wraps the entire UI
type Client struct {
	ctx           context.Context
	cancel        context.CancelFunc
	baseName      string
	patcherUrl    string
	currentPath   string
	clientVersion string
	isPatched     bool
	patchSummary  string
	cfg           *config.Config
	cacheFileList *FileList
	version       string
	cacheLog      string
	httpClient    *http.Client
}

// New creates a new client
func New(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, version string, patcherUrl string) (*Client, error) {
	var err error
	c := &Client{
		ctx:           ctx,
		cancel:        cancel,
		cfg:           cfg,
		clientVersion: "rof",
		patcherUrl:    patcherUrl,
		version:       version,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
	exeName, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("executable: %w", err)
	}

	c.baseName = filepath.Base(exeName)
	if strings.Contains(c.baseName, ".") {
		c.baseName = c.baseName[0:strings.Index(c.baseName, ".")]
	}

	fmt.Printf("Starting %s %s\n", c.baseName, c.version)
	c.currentPath, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("wd invalid: %w", err)
	}

	gui.SubscribePatchButton(func() { c.Patch() })
	gui.SubscribePlayButton(func() { c.Play() })
	gui.SubscribeAutoPatch(func() {
		c.cfg.IsAutoPatch = gui.IsAutoPatch()
		c.cfg.Save()
	})
	gui.SubscribeAutoPlay(func() {
		c.cfg.IsAutoPlay = gui.IsAutoPlay()
		c.cfg.Save()
	})

	return c, nil
}

// AutoPlay will automatically patch then play the game. It is designed to be called after New
func (c *Client) AutoPlay() error {
	gui.SetAutoMode(true)
	defer gui.SetAutoMode(false)

	isCleanAutoPlay := true
	if c.cfg.IsAutoPatch {
		fmt.Println("Autopatch is enabled, patching...")
		err := c.Patch()
		if err != nil {
			isCleanAutoPlay = false
		}
	}

	if c.cfg.IsAutoPlay {
		fmt.Println("Autoplay is enabled, playing...")
		if c.isPatched {
			c.log("Since files were patched, waiting 5 seconds before launching EverQuest")
			time.Sleep(5 * time.Second)
		}
		err := c.Play()
		if err != nil {
			c.log("Failed to play: %s", err)
			isCleanAutoPlay = false
		}
	}
	fmt.Println("Autoplay complete")
	if isCleanAutoPlay {
		return nil
	}
	return fmt.Errorf("autoplay finished with errors")
}

func (c *Client) Play() error {
	gui.LogClear()
	c.log("Launching EverQuest from %s", c.currentPath)
	username, err := c.fetchUsername()
	if err != nil {
		c.log("Failed grabbing username from eqlsPlayerData.ini: %s", err)
		//this error is not critical
	}
	if username == "" {
		username = "x"
	}
	cmd := c.createCommand(true, fmt.Sprintf("%s/eqgame.exe", c.currentPath), "patchme", "/login:"+username)
	cmd.Dir = c.currentPath
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start eqgame.exe: %w", err)
	}

	time.Sleep(1000 * time.Millisecond)
	isStarted := false
	// poll for process to be started
	for i := 0; i < 10; i++ {
		if cmd.Process == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		// check if process is running
		_, err := os.FindProcess(cmd.Process.Pid)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		isStarted = true
		c.logf("EverQuest started with process ID %d\n", cmd.Process.Pid)
		break
	}
	if !isStarted {
		return fmt.Errorf("failed to start eqgame.exe")
	}
	return nil
}

func (c *Client) Patch() error {
	start := time.Now()
	gui.LogClear()

	gui.SetPatchMode(true)
	defer gui.SetPatchMode(false)
	gui.SetProgress(0)

	_, err := os.Stat("eqgame.exe")
	if err != nil {
		c.log("eqgame.exe must be in the same directory as %s.", c.baseName)
		return fmt.Errorf("stat failed")
	}

	err = c.selfUpdateAndPatch()
	if err != nil {
		c.log("Failed to self update and patch: %s", err)
		return fmt.Errorf("self update and patch: %w", err)
	}

	if c.isPatched {
		c.log(c.patchSummary)
	}

	c.log("Finished in %0.2f seconds", time.Since(start).Seconds())
	return nil
}

func (c *Client) selfUpdateAndPatch() error {
	var err error

	err = c.fetchFileList()
	if err != nil {
		c.log("Failed fetch file list, skipping: %s", err)
		return nil
	}

	err = c.patch()
	if err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	err = c.selfUpdate()
	if err != nil {
		c.log("Failed self update, skipping: %s", err)
	}

	return nil
}

func (c *Client) fetchFileList() error {
	client := c.httpClient
	url := fmt.Sprintf("%s/filelist_%s.yml", c.patcherUrl, c.clientVersion)
	c.log("Downloading %s", url)
	resp, err := client.Get(url)
	if err != nil {
		url := fmt.Sprintf("%s/%s/filelist_%s.yml", c.patcherUrl, c.clientVersion, c.clientVersion)
		c.log("Downloading legacy %s", url)
		resp, err = client.Get(url)
		if err != nil {
			return fmt.Errorf("download %s: %w", url, err)
		}
	}
	if resp.StatusCode != 200 {
		c.cacheFileList = &FileList{}
		return fmt.Errorf("download %s responded %d (not 200)", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	fileList := &FileList{}

	err = yaml.NewDecoder(resp.Body).Decode(fileList)
	if err != nil {
		return fmt.Errorf("decode filelist: %w", err)
	}
	//c.log("patch version is", fileList.Version, "and we are version", c.cfg.ClientVersion)
	c.cacheFileList = fileList
	return nil
}

func (c *Client) selfUpdate() error {
	client := c.httpClient

	exeName, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable: %w", err)
	}

	baseName := c.baseName

	err = os.Remove(baseName + ".bat")
	if err != nil {
		if !os.IsNotExist(err) {
			c.log("Failed to remove %s.bat: %s", baseName, err)
		}
	} else {
		c.log("Removed %s.bat", baseName)
	}

	err = os.Remove("." + baseName + ".exe.old")
	if err != nil {
		if !os.IsNotExist(err) {
			c.log("Failed to remove .%s.exe.old: %s", baseName, err)
		}
	} else {
		c.log("Removed .%s.exe.old", baseName)
	}

	myHash, err := md5Checksum(exeName)
	if err != nil {
		return fmt.Errorf("checksum: %w", err)
	}
	url := fmt.Sprintf("%s/launcheq-hash.txt", c.patcherUrl)
	c.log("Checking for self update at %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s responded %d (not 200)", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("read %s: %w", url, err)
	}

	remoteHash := strings.TrimSpace(string(data))

	if remoteHash == "Not Found" {
		c.log("Remote site down, ignoring self update")
		return nil
	}

	if strings.EqualFold(myHash, remoteHash) {
		c.log("Self update not needed")
		return nil
	}

	c.log("Updating %s... %s vs %s", c.baseName, myHash, remoteHash)

	url = fmt.Sprintf("%s/%s.exe", c.patcherUrl, c.baseName)
	c.log("Downloading %s at %s", c.baseName, url)
	resp, err = client.Get(url)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s responded %d (not 200)", url, resp.StatusCode)
	}
	c.log("Applying update (will be used next launch)")
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		return fmt.Errorf("apply: %w", err)
	}

	//isErrored := false

	// c.log("Creating %s.bat", c.baseName)
	// err = os.WriteFile(fmt.Sprintf("%s.bat", c.baseName), []byte(fmt.Sprintf("timeout 1\n%s.exe", c.baseName)), os.ModePerm)
	// if err != nil {
	// 	fmt.Printf("Failed to write %s.bat: %s\n", c.baseName, err)
	// 	isErrored = true
	// }

	// c.log("Writing log")
	// err = os.WriteFile(fmt.Sprintf("%s.txt", c.baseName), []byte(c.cacheLog), os.ModePerm)
	// if err != nil {
	// 	fmt.Println("Failed to write log:", err)
	// 	isErrored = true
	// }

	// cmd := c.createCommand(false, fmt.Sprintf("%s/%s.bat", c.currentPath, c.baseName))
	// cmd.Dir = c.currentPath
	// err = cmd.Start()
	// if err != nil {
	// 	fmt.Printf("Failed to start %s.bat: %s\n", c.baseName, err)
	// 	isErrored = true
	// }

	// if isErrored && runtime.GOOS == "windows" {
	// 	fmt.Printf("There was an error while self updating %s. Review above or %s.txt to see why.\n", c.baseName, c.baseName)
	// 	fmt.Println("Automatically exiting in 10 seconds...")
	// 	time.Sleep(10 * time.Second)
	// 	os.Exit(1)
	// }

	// c.log("Successfully updated. Restarting %s and starting EverQuest...", c.baseName)
	// time.Sleep(1 * time.Second)
	// os.Exit(0)
	return nil
}

func (c *Client) log(format string, a ...interface{}) {
	c.logf(format+"\n", a...)
}

func (c *Client) logf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	gui.Logf(format, a...)
	c.cacheLog += fmt.Sprintf(format, a...)
}

func (c *Client) patch() error {
	var err error
	start := time.Now()

	fileList := c.cacheFileList

	if c.cfg.Version == fileList.Version {
		if len(fileList.Version) < 8 {
			c.log("We are up to date")
			return nil
		}
		c.log("We are up to date latest patch %s", fileList.Version[0:8])
		return nil
	}

	totalSize := int64(0)

	for _, entry := range fileList.Downloads {
		totalSize += int64(entry.Size)
	}

	progressSize := int64(1)

	totalDownloaded := int64(0)

	if len(fileList.Version) < 8 {
		c.log("Total patch size: %s", generateSize(int(totalSize)))
	} else {
		c.log("Total patch size: %s, version: %s", generateSize(int(totalSize)), fileList.Version[0:8])
	}

	ratio := float64(totalSize / 100)
	gui.SetProgress(0)

	for _, entry := range fileList.Downloads {
		if strings.Contains(entry.Name, "..") {
			c.log("Skipping %s, has .. inside it", entry.Name)
			continue
		}

		if strings.Contains(entry.Name, "/") {
			newPath := strings.TrimSuffix(entry.Name, filepath.Base(entry.Name))
			err = os.MkdirAll(newPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("mkdir %s: %w", newPath, err)
			}
		}
		_, err := os.Stat(entry.Name)
		if err != nil {
			if os.IsNotExist(err) {
				err = c.downloadPatchFile(entry)
				if err != nil {
					return fmt.Errorf("download new file: %w", err)
				}
				totalDownloaded += int64(entry.Size)
				progressSize += int64(entry.Size)
				gui.SetProgress(int(ratio * float64(progressSize)))
				c.isPatched = true
				continue
			}
			return fmt.Errorf("stat %s: %w", entry.Name, err)
		}

		hash, err := md5Checksum(entry.Name)
		if err != nil {
			return fmt.Errorf("md5checksum: %w", err)
		}

		if hash == entry.Md5 {
			c.log("%s skipped (up to date)", entry.Name)
			progressSize += int64(entry.Size)
			gui.SetProgress(int(ratio * float64(progressSize)))
			continue
		}

		err = c.downloadPatchFile(entry)
		if err != nil {
			return fmt.Errorf("download new file: %w", err)
		}
		progressSize += int64(entry.Size)
		totalDownloaded += int64(entry.Size)
		gui.SetProgress(int(ratio * float64(progressSize)))
		c.isPatched = true
	}

	for _, entry := range fileList.Deletes {
		if strings.Contains(entry.Name, "..") {
			c.log("Skipping %s, has .. inside it", entry.Name)
			continue
		}
		fi, err := os.Stat(entry.Name)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("stat %s: %w", entry.Name, err)
		}
		if fi.IsDir() {
			c.log("Skipping deleting %s, it is a directory", entry.Name)
			continue
		}
		err = os.Remove(entry.Name)
		if err != nil {
			c.log("Failed to delete %s: %s", entry.Name, err)
			continue
		}
		c.log("%s removed", entry.Name)
	}
	gui.SetProgress(100)

	c.cfg.Version = fileList.Version
	err = c.cfg.Save()
	if err != nil {
		c.log("Failed to save version to %s.ini: %s", c.baseName, err)
	}

	if totalDownloaded == 0 {
		c.patchSummary = fmt.Sprintf("Finished patch in %0.2f seconds", time.Since(start).Seconds())
		return nil
	}
	c.patchSummary = fmt.Sprintf("Finished patch of %s in %0.2f seconds", generateSize(int(totalDownloaded)), time.Since(start).Seconds())

	return nil
}

func (c *Client) downloadPatchFile(entry FileEntry) error {
	c.log("%s (%s)", entry.Name, generateSize(entry.Size))
	w, err := os.Create(entry.Name)
	if err != nil {
		return fmt.Errorf("create %s: %w", entry.Name, err)
	}
	defer w.Close()
	client := c.httpClient

	url := fmt.Sprintf("%s/%s/%s", c.cacheFileList.DownloadPrefix, c.clientVersion, entry.Name)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download %s responded %d (not 200)", url, resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("write %s: %w", entry.Name, err)
	}
	return nil
}

func md5Checksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("new: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generateSize(in int) string {
	val := float64(in)
	if val < 1024 {
		return fmt.Sprintf("%0.2f bytes", val)
	}
	val /= 1024
	if val < 1024 {
		return fmt.Sprintf("%0.2f KB", val)
	}
	val /= 1024
	if val < 1024 {
		return fmt.Sprintf("%0.2f MB", val)
	}
	val /= 1024
	if val < 1024 {
		return fmt.Sprintf("%0.2f GB", val)
	}
	val /= 1024
	return fmt.Sprintf("%0.2f TB", val)
}

func (c *Client) fetchUsername() (string, error) {

	r, err := os.Open("eqlsPlayerData.ini")
	if err != nil {
		return "", err
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Username=") {
			line = strings.TrimPrefix(line, "Username=")
			return line, nil
		}
	}
	return "", nil
}

func (c *Client) DumpLog() error {
	if len(c.cacheLog) == 0 {
		return nil
	}
	err := os.WriteFile(fmt.Sprintf("%s.txt", c.baseName), []byte(c.cacheLog), os.ModePerm)
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	c.cacheLog = ""
	return nil
}
