//go:build !windows
// +build !windows

package gui

import "context"

func NewMainWindow(ctx context.Context, cancel context.CancelFunc, splash []byte) error {
	return nil
}

func Run() int {
	return 0
}