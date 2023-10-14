//go:build windows
// +build windows

package client

import (
	"os/exec"
	"syscall"
)

func (c *Client) createCommand(isHidden bool, name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: isHidden}
	return cmd
}
