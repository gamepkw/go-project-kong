package database

import (
	"gorm.io/gorm"
)

type database struct {
	db *gorm.DB
}

func New(db *gorm.DB) Database {
	return &database{
		db: db,
	}
}

type Database interface {
}
