package postgres

import "gorm.io/gorm"

type DBEngine interface {
	GetDB() *gorm.DB
}
