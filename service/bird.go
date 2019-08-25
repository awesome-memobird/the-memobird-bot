package service

import (
	"fmt"

	"github.com/awesome-memobird/the-memobird-bot/memobird"
)

// BirdApp represents the ability to interact with a memobird.
type BirdApp interface {
	PrintText(text string, birdID string) (*memobird.PrintResult, error)
	BindDevice(deviceID string) (*memobird.BindResult, error)
}

// Bird provide core functionalities of memobird.
type Bird struct {
	BirdApp BirdApp
}

// PrintTextToBird sends given text to memobird of birdID for printing.
func (b *Bird) PrintTextToBird(birdID, text string) (*memobird.PrintResult, error) {
	return b.BirdApp.PrintText(text, birdID)
}

// BindBirdWithMessage binds birdID and sends given message to bird.
func (b *Bird) BindBirdWithMessage(birdID, msg string) (*memobird.PrintResult, error) {
	result, err := b.BirdApp.BindDevice(birdID)
	if err != nil {
		return nil, fmt.Errorf("binding device: %w", err)
	}
	if !result.IsSuccess {
		return nil, fmt.Errorf("binding device: %w", result.Err)
	}

	r, err := b.PrintTextToBird(birdID, msg)
	if err != nil {
		return nil, fmt.Errorf("printing verification code: %w", err)
	}
	return r, err
}
