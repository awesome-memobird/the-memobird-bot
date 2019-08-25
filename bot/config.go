package bot

import (
	"time"
)

// Config contains configurations to create a bot.
type Config struct {
	Token         string
	PollerTimeout time.Duration

	UserService   UserService
	DeviceService DeviceService
	BirdService   BirdService
}
