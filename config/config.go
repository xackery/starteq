package config

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// Config represents a configuration parse
type Config struct {
	Version     string
	baseName    string
	IsAutoPlay  bool
	IsAutoPatch bool
	IsTorrentOK bool
}

// New creates a new configuration
func New(ctx context.Context, baseName string) (*Config, error) {
	var f *os.File
	cfg := &Config{
		baseName: baseName,
	}
	path := baseName + ".ini"

	isNewConfig := false
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("config info: %w", err)
		}
		f, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create %s.ini: %w", baseName, err)
		}
		fi, err = os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("new config info: %w", err)
		}
		isNewConfig = true
	}
	if !isNewConfig {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open config: %w", err)
		}
	}

	defer f.Close()
	if fi.IsDir() {
		return nil, fmt.Errorf("%s.ini is a directory, should be a file", baseName)
	}

	if isNewConfig {
		cfg = &Config{
			baseName: baseName,
		}
		err = cfg.Save()
		if err != nil {
			return nil, fmt.Errorf("save config: %w", err)
		}
		return cfg, nil
	}

	err = decode(f, cfg)
	if err != nil {
		return nil, fmt.Errorf("decode %s.ini: %w", baseName, err)
	}

	return cfg, nil
}

// Verify returns an error if configuration appears off
func (c *Config) Verify() error {

	return nil
}

func decode(r io.Reader, cfg *Config) error {
	reader := bufio.NewScanner(r)
	for reader.Scan() {
		line := reader.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		switch key {
		case "version":
			cfg.Version = value
		case "auto_patch":
			if strings.ToLower(value) == "true" {
				cfg.IsAutoPatch = true
			}
			if value == "1" {
				cfg.IsAutoPatch = true
			}
		case "auto_play":
			if strings.ToLower(value) == "true" {
				cfg.IsAutoPlay = true
			}
			if value == "1" {
				cfg.IsAutoPlay = true
			}
		case "torrent_ok":
			if strings.ToLower(value) == "true" {
				cfg.IsTorrentOK = true
			}
			if value == "1" {
				cfg.IsTorrentOK = true
			}

		}
	}
	return nil
}

// Save saves the config
func (c *Config) Save() error {

	fi, err := os.Stat(c.baseName + ".ini")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s.ini:  %w", c.baseName, err)
		}
		w, err := os.Create(c.baseName + ".ini")
		if err != nil {
			return fmt.Errorf("create %s.ini: %w", c.baseName, err)
		}
		w.Close()
	}
	if fi != nil && fi.IsDir() {
		return fmt.Errorf("dirCheck %s.ini: is a directory", c.baseName)
	}

	r, err := os.Open(c.baseName + ".ini")
	if err != nil {
		return err
	}
	defer r.Close()

	tmpConfig := &Config{}

	out := ""
	reader := bufio.NewScanner(r)
	for reader.Scan() {
		line := reader.Text()
		if strings.HasPrefix(line, "#") {
			out += line + "\n"
			continue
		}
		if !strings.Contains(line, "=") {
			out += line + "\n"
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "version":
			if tmpConfig.Version == "1" {
				continue
			}
			out += fmt.Sprintf("%s = %s\n", key, c.Version)
			tmpConfig.Version = "1"
			continue
		case "auto_patch":
			if tmpConfig.IsAutoPatch {
				continue
			}

			if c.IsAutoPatch {
				value = "true"
			} else {
				value = "false"
			}
			tmpConfig.IsAutoPatch = true
		case "auto_play":
			if tmpConfig.IsAutoPlay {
				continue
			}
			if c.IsAutoPlay {
				value = "true"
			} else {
				value = "false"
			}
			tmpConfig.IsAutoPlay = true
		case "torrent_ok":
			if tmpConfig.IsTorrentOK {
				continue
			}
			if c.IsTorrentOK {
				value = "true"
			} else {
				value = "false"
			}
			tmpConfig.IsTorrentOK = true
		}
		line = fmt.Sprintf("%s = %s", key, value)
		out += line + "\n"
	}

	if tmpConfig.Version != "1" && c.Version != "" {
		out += fmt.Sprintf("version = %s\n", c.Version)
	}
	if !tmpConfig.IsAutoPatch {
		if c.IsAutoPatch {
			out += "auto_patch = true\n"
		} else {
			out += "auto_patch = false\n"
		}
	}
	if !tmpConfig.IsAutoPlay {
		if c.IsAutoPlay {
			out += "auto_play = true\n"
		} else {
			out += "auto_play = false\n"
		}
	}
	if !tmpConfig.IsTorrentOK {
		if c.IsTorrentOK {
			out += "torrent_ok = true\n"
		}
		// no need to flag torrent ok if false
	}

	err = os.WriteFile(c.baseName+".ini", []byte(out), 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
