//go:build !windows
// +build !windows

package gui

import (
	"context"

	"github.com/xackery/starteq/config"
)

var ()

func NewMainWindow(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, splash []byte) error {
	return nil
}

func Run() int {
	return 0
}

func SubscribePatchButton(fn func()) {
}

func SubscribePlayButton(fn func()) {
}

func SubscribeAutoPatch(fn func()) {
}

func SubscribeClose(fn func(cancelled *bool, reason byte)) {
}

func IsAutoPatch() bool {
	return true
}

func SubscribeAutoPlay(fn func()) {
}

func IsAutoPlay() bool {
	return true
}

func SetAutoMode(value bool) {
}

func LogClear() {

}

func SetProgress(value int) {

}

func SetPatchMode(value bool) {

}

func IsAutoMode() bool {
	return true
}

func SetPatchText(value string) {

}

func MessageBoxYesNo(title, message string) bool {
	return false
}

func MessageBox(title, message string, isError bool) {
}

func SetTitle(title string) {

}

func Close() error {
	return nil
}
