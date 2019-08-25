package model

import (
	"github.com/jinzhu/gorm"
)

// User stores the users of bot.
type User struct {
	gorm.Model
	TelegramID       int64
	TelegramUserName string
	TelegramFullName string
}
