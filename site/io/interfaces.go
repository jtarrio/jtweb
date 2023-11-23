package io

import (
	"errors"
	"io"
	"time"
)

type Input interface {
	io.Reader
	Close() error
}

type Output interface {
	io.Writer
	Close() error
}

type File interface {
	PathName() string
	GoTo(name string) File
	Create() (Output, error)
	CreateBytes([]byte) error
	Read() (Input, error)
	ReadBytes() ([]byte, error)
	Stat() (Stat, error)
	Chtime(mtime time.Time) error
	ForAllFiles(fn ForAllFilesFunc) error
}

type Stat struct {
	ModTime time.Time
}

type ForAllFilesFunc func(file File, err error) error

//lint:ignore ST1012 not an actual error
var SkipRemaining = errors.New("skip remaining files")
