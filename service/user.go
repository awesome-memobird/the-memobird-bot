package service

import (
	"github.com/awesome-memobird/the-memobird-bot/model"
	"github.com/jinzhu/gorm"
)

// User provides core functionalities of user.
type User struct {
	DB *gorm.DB
}

// IsExistsByTelegramID returns true if given telegramID already exists.
func (u *User) IsExistsByTelegramID(telegramID int) (bool, error) {
	var count int
	r := u.DB.Model(&model.User{}).Where("telegram_id = ?", telegramID).Count(&count)
	return count == 1, r.Error
}

// GetByTelegramID returns user of given telegram ID.
func (u *User) GetByTelegramID(telegramID int) (*model.User, error) {
	var user model.User
	r := u.DB.First(&user, "telegram_id = ?", telegramID)
	return &user, r.Error
}

// New creates a user.
func (u *User) New(user *model.User) error {
	return u.DB.Create(user).Error
}
