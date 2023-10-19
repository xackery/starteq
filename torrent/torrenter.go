package torrent

import "context"

type Torrenter interface {
	Download(ctx context.Context, torrentData []byte) error
}
