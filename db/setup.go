package db

import (
	"sync"
	"time"
)

var options *Options

func DefaultOptions() *Options {
	return options
}

type Options struct {
	DriverType string
	Dsn        string

	SkipDefaultTransaction bool

	MaxIdleConns int
	MaxOpenConns int
	MaxIdleTime  time.Duration

	connMap sync.Map
}

func newOptions() *Options {
	return &Options{
		SkipDefaultTransaction: true,
	}
}

func Setup(fn func(options *Options)) {
	if options == nil {
		options = newOptions()
	}
	if fn == nil {
		return
	}
	fn(options)
}
