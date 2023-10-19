package torrent

import "context"

type Mock struct {
}

func (m *Mock) Download(ctx context.Context, torrentData []byte) error {
	return nil
}
