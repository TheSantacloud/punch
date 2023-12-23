package database

import (
	"fmt"

	"github.com/dormunis/punch/pkg/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(engine string, path string) (*gorm.DB, error) {
	if engine != "sqlite3" {
		return nil, fmt.Errorf("unsupported database engine: %s", engine)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	// TODO: use default currency for RepoCompany
	db.AutoMigrate(&repositories.RepoCompany{}, &repositories.RepoSession{})

	return db, nil
}
