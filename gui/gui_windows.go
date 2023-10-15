//go:build windows
// +build windows

package gui

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"sync"

	"github.com/xackery/launcheq/config"
	"github.com/xackery/wlk/walk"
)

type Gui struct {
	ctx         context.Context
	cancel      context.CancelFunc
	mw          *walk.MainWindow
	splash      *walk.ImageView
	isAutoPatch *walk.CheckBox
	isAutoPlay  *walk.CheckBox
	patchButton *walk.PushButton
	playButton  *walk.PushButton
	progress    *walk.ProgressBar
	log         *walk.TextEdit
}

var (
	gui        *Gui
	isAutoMode bool
	mu         sync.RWMutex
)

// NewMainWindow creates a new main window
func NewMainWindow(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, splash []byte) error {
	gui = &Gui{
		ctx:    ctx,
		cancel: cancel,
	}
	isAutoMode = true

	var err error
	gui.mw, err = walk.NewMainWindowWithName("starteq")
	if err != nil {
		return fmt.Errorf("new main window: %w", err)
	}
	gui.mw.SetTitle("Start EQ (Client: Rain of Fear 2)")
	gui.mw.SetMinMaxSize(walk.Size{Width: 305, Height: 371}, walk.Size{Width: 305, Height: 371})
	gui.mw.SetLayout(walk.NewVBoxLayout())
	gui.mw.SetVisible(true)

	// convert splash from byte slice to png
	splashImg, err := png.Decode(bytes.NewReader(splash))
	if err != nil {
		return fmt.Errorf("decode splash: %w", err)
	}

	bmp, err := walk.NewBitmapFromImageForDPI(splashImg, 96)
	if err != nil {
		return fmt.Errorf("new bitmap from image for dpi: %w", err)
	}

	gui.splash, err = walk.NewImageView(gui.mw)
	if err != nil {
		return fmt.Errorf("new image view: %w", err)
	}
	gui.splash.SetImage(bmp)
	gui.splash.SetMinMaxSize(walk.Size{Width: 400, Height: 400}, walk.Size{Width: 400, Height: 400})
	gui.mw.Children().Add(gui.splash)
	gui.splash.SetVisible(true)

	gui.log, err = walk.NewTextEdit(gui.mw)
	if err != nil {
		return fmt.Errorf("new text edit: %w", err)
	}
	gui.log.SetReadOnly(true)
	gui.log.SetVisible(false)
	gui.log.SetMinMaxSize(walk.Size{Width: 400, Height: 400}, walk.Size{Width: 400, Height: 400})
	gui.mw.Children().Add(gui.log)

	gui.patchButton, err = walk.NewPushButton(gui.mw)
	if err != nil {
		return fmt.Errorf("new push button: %w", err)
	}
	gui.patchButton.SetMinMaxSize(walk.Size{Width: 95, Height: 52}, walk.Size{Width: 95, Height: 52})
	gui.patchButton.SetText("Patch")
	gui.patchButton.SetVisible(true)

	gui.isAutoPatch, err = walk.NewCheckBox(gui.mw)
	if err != nil {
		return fmt.Errorf("new check box: %w", err)
	}
	gui.isAutoPatch.SetText("Auto Patch")
	gui.isAutoPatch.SetChecked(cfg.IsAutoPatch)

	gui.isAutoPlay, err = walk.NewCheckBox(gui.mw)
	if err != nil {
		return fmt.Errorf("new check box: %w", err)
	}
	gui.isAutoPlay.SetText("Auto Play")
	gui.isAutoPlay.SetChecked(cfg.IsAutoPlay)

	gui.playButton, err = walk.NewPushButton(gui.mw)
	if err != nil {
		return fmt.Errorf("new push button: %w", err)
	}

	gui.playButton.SetText("Play")
	gui.playButton.SetMinMaxSize(walk.Size{Width: 95, Height: 52}, walk.Size{Width: 95, Height: 52})
	gui.playButton.SetVisible(true)
	gui.playButton.SetAlwaysConsumeSpace(true)

	comp, err := walk.NewComposite(gui.mw)
	if err != nil {
		return fmt.Errorf("new composite: %w", err)
	}
	comp.SetLayout(walk.NewHBoxLayout())
	comp.Children().Add(gui.patchButton)
	comp.Children().Add(gui.isAutoPatch)
	comp.Children().Add(gui.isAutoPlay)
	comp.Children().Add(gui.playButton)

	gui.progress, err = walk.NewProgressBar(gui.mw)
	if err != nil {
		return fmt.Errorf("new progress bar: %w", err)
	}

	gui.progress.SetMinMaxSize(walk.Size{Width: 400, Height: 39}, walk.Size{Width: 400, Height: 39})
	gui.progress.SetValue(50)
	gui.progress.SetMinMaxSize(walk.Size{Width: 400, Height: 39}, walk.Size{Width: 400, Height: 39})

	gui.mw.Children().Add(gui.progress)
	gui.mw.SetSize(walk.Size{Width: 305, Height: 371})

	return nil
}

func Run() int {
	if gui == nil {
		return 1
	}
	return gui.mw.Run()
}

// SubscribePatchButton subscribes to the patch button
func SubscribePatchButton(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.patchButton.Clicked().Attach(fn)
}

// SubscribePlayButton subscribes to the play button
func SubscribePlayButton(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.playButton.Clicked().Attach(fn)
}

func SubscribeAutoPatch(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.isAutoPatch.CheckedChanged().Attach(fn)
}

func SubscribeAutoPlay(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.isAutoPlay.CheckedChanged().Attach(fn)
}

// Logf logs a message to the gui
func Logf(format string, a ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}

	if !gui.log.Visible() && !isAutoMode {
		gui.log.SetVisible(true)
		gui.splash.SetVisible(false)
	}
	if !isAutoMode {
		//convert \n to \r\n
		format = strings.ReplaceAll(format, "\n", "\r\n")
		gui.log.AppendText(fmt.Sprintf(format, a...))
	}
}

func LogClear() {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.log.SetText("")
}

// SetAutoMode sets the gui to auto mode, this is
func SetAutoMode(value bool) {
	mu.Lock()
	defer mu.Unlock()
	isAutoMode = value
}

func SetPatchMode(value bool) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.patchButton.SetEnabled(!value)
}

func IsAutoPatch() bool {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return false
	}
	return gui.isAutoPatch.Checked()
}

func IsAutoPlay() bool {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return false
	}
	return gui.isAutoPlay.Checked()
}

func SetMaxProgress(value int) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.progress.SetRange(0, value)
}

func SetProgress(value int) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	gui.progress.SetValue(value)
}

func MessageBox(title string, message string, isError bool) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	// convert style to msgboxstyle
	icon := walk.MsgBoxIconInformation
	if isError {
		icon = walk.MsgBoxIconError
	}
	walk.MsgBox(gui.mw, title, message, icon)
}

func MessageBoxf(title string, format string, a ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if gui == nil {
		return
	}
	// convert style to msgboxstyle
	icon := walk.MsgBoxIconInformation
	walk.MsgBox(gui.mw, title, fmt.Sprintf(format, a...), icon)
}
