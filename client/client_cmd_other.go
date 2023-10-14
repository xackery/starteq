//go:build !windows
// +build !windows

package client

import (
	"os/exec"
)

func (c *Client) createCommand(isHidden bool, name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	return cmd
}
