package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/xackery/starteq/client"
	"github.com/xackery/starteq/config"
	"github.com/xackery/starteq/gui"
	"github.com/xackery/starteq/slog"
	"github.com/xackery/wlk/walk"
)

//go:embed splash.png
var starteqSplash []byte

var (
	Version    string
	PatcherUrl string
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exeName, err := os.Executable()
	if err != nil {
		gui.MessageBox("Error", "Failed to get executable name", true)
		os.Exit(1)
	}
	baseName := filepath.Base(exeName)
	if strings.Contains(baseName, ".") {
		baseName = baseName[0:strings.Index(baseName, ".")]
	}
	if baseName == "" {
		baseName = "starteq"
	}
	cfg, err := config.New(context.Background(), baseName)
	if err != nil {
		gui.MessageBox("Error", "Failed to load config: "+err.Error(), true)
		os.Exit(1)
	}

	err = gui.NewMainWindow(ctx, cancel, cfg, starteqSplash)
	if err != nil {
		gui.MessageBox("Error", "Failed to create main window: "+err.Error(), true)
		os.Exit(1)
	}
	PatcherUrl = strings.TrimSuffix(PatcherUrl, "/")
	if Version == "" {
		Version = "dev"
	}

	c, err := client.New(ctx, cancel, cfg, Version, PatcherUrl)
	if err != nil {
		gui.MessageBox("Error", "Failed to create client: "+err.Error(), true)
		os.Exit(1)
	}
	defer slog.Dump(baseName + ".txt")
	defer c.Done()

	gui.SubscribeClose(func(canceled *bool, reason walk.CloseReason) {
		if ctx.Err() != nil {
			fmt.Println("Accepting exit")
			return
		}
		*canceled = true
		fmt.Println("Got close message")
		gui.SetTitle("Closing...")
		cancel()
	})

	go func() {
		<-ctx.Done()
		fmt.Println("Doing clean up process...")
		c.Done() // close client
		gui.Close()
		walk.App().Exit(0)
		fmt.Println("Done, exiting")
		slog.Dump(baseName + ".txt")
		os.Exit(0)
	}()

	err = c.AutoPlay()
	if err == nil {
		// no gui needed if auto play worked with zero errors
		fmt.Println("Autoplay worked cleanly, exiting")
		return
	}
	slog.Dump(baseName + ".txt")

	errCode := gui.Run()
	if errCode != 0 {
		fmt.Println("Failed to run:", errCode)
		os.Exit(1)
	}

}
