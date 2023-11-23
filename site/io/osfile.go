package io

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
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
