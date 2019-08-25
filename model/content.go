package model

import "github.com/jinzhu/gorm"

// Content stores the contents sent to print.
type Content struct {
	gorm.Model
	ContentID int64
	IsPrinted bool
}
