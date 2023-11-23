package io

import (
	"errors"
	"io"
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
	Read() (Input, error)
	ForAllFiles(fn ForAllFilesFunc) error
}

type ForAllFilesFunc func(file File, err error) error

//lint:ignore ST1012 not an actual error
var SkipRemaining = errors.New("skip remaining files")
