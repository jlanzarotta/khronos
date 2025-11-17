package util

import (
	"fmt"
	"sync"
	"time"
)

// Spinner represents a simple loading spinner
type Spinner struct {
	message string
	delay   time.Duration
	chars   []rune
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewSpinner creates a new spinner with a custom message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		delay:   100 * time.Millisecond,
		chars:   []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'},
		stopCh:  make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stopCh:
				fmt.Print("\r\033[K") // Clear the line
				return
			default:
				fmt.Printf("\r%c %s", s.chars[i], s.message)
				i = (i + 1) % len(s.chars)
				time.Sleep(s.delay)
			}
		}
	}()
}

// Stop stops the spinner and optionally prints a completion message
func (s *Spinner) Stop(completionMsg string) {
	close(s.stopCh)
	s.wg.Wait()
	if completionMsg != "" {
		fmt.Println(completionMsg)
	}
}

// RunWithSpinner executes a task while displaying a spinner
func RunWithSpinner(message string, task func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := task()

	if err != nil {
		spinner.Stop(fmt.Sprintf("✗ %s - Failed", message))
		return err
	}

	spinner.Stop(fmt.Sprintf("✓ %s - Complete", message))
	return nil
}