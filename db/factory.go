package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySqlDialector(opts *Options) gorm.Dialector {
	return mysql.New(mysql.Config{
		DSN: opts.Dsn,
	})
}

func NewSqliteDialector(opts *Options) gorm.Dialector {
	exec, _ := filepath.Abs("./")
	return sqlite.Open(filepath.Join(exec, "sms.db"))
}

func NewDialector(opts *Options) (gorm.Dialector, error) {
	switch opts.DriverType {
	case "mysql":
		return NewMySqlDialector(opts), nil
	case "sqlite":
		return NewSqliteDialector(opts), nil
	default:
		return nil, fmt.Errorf("unsupported driver type: %v", opts.DriverType)
	}
}

func OpenConnection(opts *Options) (*gorm.DB, error) {
	gormConf := &gorm.Config{
		SkipDefaultTransaction: opts.SkipDefaultTransaction,
	}

	dialector, err := NewDialector(opts)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dialector, gormConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: dsn: %s error:%w", opts.Dsn, err)
	}

	rawdb, err := db.DB()
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	rawdb.SetConnMaxIdleTime(opts.MaxIdleTime)
	rawdb.SetMaxIdleConns(opts.MaxIdleConns)
	rawdb.SetMaxOpenConns(opts.MaxOpenConns)

	return db, nil
}
