package io

import (
	"log"
	"time"
)

type dryRunFile struct {
	file File
}

func DryRunFile(file File) File {
	return &dryRunFile{file: file}
}

func (f *dryRunFile) Name() string {
	return f.file.Name()
}

func (f *dryRunFile) BaseName() string {
	return f.file.BaseName()
}

func (f *dryRunFile) FullPath() string {
	return f.file.FullPath()
}

func (f *dryRunFile) GoTo(name string) File {
	return DryRunFile(f.file.GoTo(name))
}

func (f *dryRunFile) Create() (Output, error) {
	return &nullOutput{file: f, bytes: 0}, nil
}

type nullOutput struct {
	file  *dryRunFile
	bytes int
}

func (o *nullOutput) Close() error {
	log.Printf("[Dry run] Creating file %s with %d bytes", o.file.FullPath(), o.bytes)
	return nil
}

func (f *dryRunFile) CreateBytes(content []byte) error {
	log.Printf("[Dry run] Creating file %s with %d bytes", f.FullPath(), len(content))
	return nil
}

func (o *nullOutput) Write(p []byte) (int, error) {
	o.bytes += len(p)
	return len(p), nil
}

func (f *dryRunFile) Read() (Input, error) {
	return f.file.Read()
}

func (f *dryRunFile) ReadBytes() ([]byte, error) {
	return f.file.ReadBytes()
}

func (f *dryRunFile) Stat() (Stat, error) {
	return f.file.Stat()
}

func (f *dryRunFile) Chtime(mtime time.Time) error {
	return nil
}

func (f *dryRunFile) ForAllFiles(fn ForAllFilesFunc) error {
	return f.file.ForAllFiles(func(file File, err error) error {
		return fn(DryRunFile(file), err)
	})
}
