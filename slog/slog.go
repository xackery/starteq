package slog

import (
	"fmt"
	"os"
	"sync"
)

var (
	mu       sync.RWMutex
	cacheLog string
	handlers []func(format string, a ...interface{})
	isDumped bool
)

func init() {
	AddHandler(func(format string, a ...interface{}) {
		fmt.Printf(format, a...)
	})
}

// AddHandler adds a log handler
func AddHandler(handler func(format string, a ...interface{})) {
	mu.Lock()
	defer mu.Unlock()
	handlers = append(handlers, handler)
}

// Dump writes the log to a file
func Dump(path string) error {
	mu.RLock()
	defer mu.RUnlock()
	if len(cacheLog) == 0 {
		return nil
	}
	if !isDumped {
		err := os.WriteFile(path, []byte(cacheLog), os.ModePerm)
		if err != nil {
			return fmt.Errorf("write log: %w", err)
		}
		cacheLog = ""
		isDumped = true
		return nil
	}
	//append to existing file instead
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(cacheLog)
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	cacheLog = ""
	return nil
}

// Printf writes to the log
func Printf(format string, a ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	for _, handler := range handlers {
		handler(format, a...)
	}
	cacheLog += fmt.Sprintf(format, a...)
}

// Println writes to the log
func Println(a ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	for _, handler := range handlers {
		handler("%s\n", fmt.Sprint(a...))
	}
	cacheLog += fmt.Sprintf("%s\n", fmt.Sprint(a...))
}

// Print is similar to printf, but adds a newline
func Print(format string, a ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	for _, handler := range handlers {
		handler(format+"\n", a...)
	}
	cacheLog += fmt.Sprintf(format+"\n", a...)
}
