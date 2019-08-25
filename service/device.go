package service

import (
	"fmt"

	"github.com/awesome-memobird/the-memobird-bot/model"
	"github.com/jinzhu/gorm"
)

// Device provides core functionalities of device.
type Device struct {
	DB *gorm.DB
}

// IsFree returns true if given deviceID doesn't exist or not owned by other users.
func (d *Device) IsFree(memobirdID string) (bool, error) {
	var device model.Device
	r := d.DB.First(&device, "memobird_id = ?", memobirdID)
	if r.RecordNotFound() {
		return true, nil
	}
	return !device.IsVerified(), r.Error
}

// New creates a device with verification code generated.
func (d *Device) New(device *model.Device) (*model.Device, error) {
	device.GenerateVerificationCode()
	return device, d.DB.Create(device).Error
}

// VerifyCodeByUserID checks if given verification code matches to the user.
func (d *Device) VerifyCodeByUserID(code string, userID uint) (bool, error) {
	var err error
	r := d.DB.Model(&model.Device{}).
		Where("user_id = ? and verification_code = ?", userID, code).
		Update("verification_code", model.DeviceVerified)
	if r.Error != nil {
		err = fmt.Errorf("updating device: %w", r.Error)
	}
	return r.RowsAffected == 1, err
}

// GetByUserID returns Device with given userID.
func (d *Device) GetByUserID(userID uint) (*model.Device, error) {
	var device model.Device
	return &device, d.DB.First(&device, "user_id = ?", userID).Error
}
