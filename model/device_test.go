package model

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomIntFixedLengthRangeOK(t *testing.T) {
	for l := 1; l < 10; l++ {
		for i := 1; i < 100000; i++ {
			n := randomIntFixedLength(l)
			assert.GreaterOrEqual(t, n, int64(math.Pow10(l)))
			assert.Less(t, n, int64(math.Pow10(l+1)))
		}
	}
}

func TestDeviceIsVerifiedOK(t *testing.T) {
	for d, expected := range map[*Device]bool{
		// no code
		&Device{}:          false,
		&Device{UserID: 1}: false,
		// code presents
		(&Device{}).GenerateVerificationCode():          false,
		(&Device{UserID: 1}).GenerateVerificationCode(): false,
		(&Device{UserID: 1}).GenerateVerificationCode(): false,
		// code verified without userID
		(&Device{VerificationCode: DeviceVerified}): false,
		// code verified with userID
		(&Device{UserID: 1, VerificationCode: DeviceVerified}): true,
	} {
		if expected {
			assert.True(t, d.IsVerified(), d)
		} else {
			assert.False(t, d.IsVerified(), d)
		}
	}
}
