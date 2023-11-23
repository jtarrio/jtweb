package io

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

type osFile struct {
	rel  string
	base string
}

func cleanRel(name string) string {
	return path.Join("/", name)[1:]
}

func OsFile(basePath string) File {
	return &osFile{rel: "", base: cleanRel(basePath)}
}

func (i *osFile) PathName() string {
	return i.rel
}

func (i *osFile) path() string {
	return path.Join(i.base, i.rel)
}

func (i *osFile) GoTo(name string) File {
	return &osFile{rel: cleanRel(path.Join(i.rel, name)), base: i.base}
}

func (i *osFile) Create() (Output, error) {
	path := i.path()
	out, err := os.Create(path)
	if err == nil {
		return out, nil
	}
	err = os.MkdirAll(filepath.Dir(path), 0o775)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

func (i *osFile) Read() (Input, error) {
	return os.Open(i.path())
}

func (i *osFile) ReadBytes() ([]byte, error) {
	input, err := i.Read()
	if err != nil {
		return nil, err
	}
	defer input.Close()
	buffer := bytes.Buffer{}
	_, err = buffer.ReadFrom(input)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (i *osFile) Stat() (Stat, error) {
	stat, err := os.Stat(i.path())
	if err != nil {
		return Stat{}, err
	}
	return Stat{ModTime: stat.ModTime()}, nil
}

func (i *osFile) Chtime(mtime time.Time) error {
	return os.Chtimes(i.path(), mtime, mtime)
}

func (i *osFile) ForAllFiles(fn ForAllFilesFunc) error {
	err := filepath.WalkDir(i.path(), func(name string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		suffix, nerr := filepath.Rel(i.base, name)
		if nerr != nil {
			return nerr
		}
		return fn(i.GoTo(suffix), err)
	})
	if err == SkipRemaining {
		return nil
	}
	return err
}
