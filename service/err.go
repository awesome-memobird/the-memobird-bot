package service

import "github.com/jinzhu/gorm"

// IsRecordNotFoundError returns true if the given error is caused by a missing record.
var IsRecordNotFoundError = gorm.IsRecordNotFoundError
