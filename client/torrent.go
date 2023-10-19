package client

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/xackery/starteq/torrent"
)

//go:embed rof2.torrent
var torrentContent []byte

// Torrent downloads the torrent
func (c *Client) Torrent(ctx context.Context) error {
	start := time.Now()
	m := torrent.Torrent{}
	err := m.Download(ctx, torrentContent)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	err = c.CopyBackup("everquest_rof2")
	if err != nil {
		return fmt.Errorf("copyBackup: %w", err)
	}

	fmt.Printf("Finished in %0.2f seconds\n", time.Since(start).Seconds())

	return nil
}
