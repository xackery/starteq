package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/xackery/launcheq/client"
	"github.com/xackery/launcheq/config"
	"github.com/xackery/launcheq/gui"
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
	cfg, err := config.New(context.Background(), c.baseName)
	if err != nil {
		fmt.Println("Failed to create config:", err)
		os.Exit(1)
	}

	err = gui.NewMainWindow(ctx, cancel, cfg, starteqSplash)
	if err != nil {
		fmt.Println("Failed to create main window:", err)
		os.Exit(1)
	}
	PatcherUrl = strings.TrimSuffix(PatcherUrl, "/")
	if Version == "" {
		Version = "dev"
	}
	c, err := client.New(ctx, cancel, cfg, Version, PatcherUrl)
	if err != nil {
		fmt.Println("Failed client new:", err)
		os.Exit(1)
	}

	go func() {

	}()

	c.AutoPlay()
	errCode := gui.Run()
	if errCode != 0 {
		fmt.Println("Failed to run:", errCode)
		os.Exit(1)
	}
	fmt.Println("Gui exited")
	cancel()

	<-ctx.Done()
}
