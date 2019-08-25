package model

import (
	"math"
	"math/rand"

	"github.com/jinzhu/gorm"
)

// Device stores the memobird devices.
type Device struct {
	gorm.Model
	MemobirdID       string
	UserID           uint
	VerificationCode int64
}

// DeviceVerified indicates the device was verified.
const DeviceVerified int64 = -1

func randomIntFixedLength(len int) int64 {
	if len == 0 {
		return -1
	}
	var lo = int64(math.Pow10(len))
	var hi = lo * 10
	return lo + int64(rand.Int63n(hi-lo))
}

// GenerateVerificationCode sets the VerificationCode to a random number.
func (d *Device) GenerateVerificationCode() *Device {
	// TODO: avoid duplicate code when user has multiple devices waiting to be verified.
	d.VerificationCode = randomIntFixedLength(6)
	return d
}

// IsVerified returns true if the device was verified.
func (d Device) IsVerified() bool {
	return d.VerificationCode == DeviceVerified && d.UserID > 0
}
