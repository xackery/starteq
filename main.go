package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/xackery/launcheq/client"
)

var (
	Version    string
	PatcherUrl string
)

func main() {

	PatcherUrl = strings.TrimSuffix(PatcherUrl, "/")
	c, err := client.New(Version, PatcherUrl)
	if err != nil {
		fmt.Println("Failed client new:", err)
		os.Exit(1)
	}
	c.Patch()
}
