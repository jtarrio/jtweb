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
	return &osFile{rel: "", base: cleanRel(filepath.ToSlash(basePath))}
}

func (f *osFile) Name() string {
	return f.rel
}

func (f *osFile) BaseName() string {
	return filepath.Base(f.rel)
}

func (f *osFile) FullPath() string {
	return filepath.FromSlash(path.Join(f.base, f.rel))
}

func (f *osFile) GoTo(name string) File {
	return &osFile{rel: cleanRel(path.Join(f.rel, name)), base: f.base}
}

func (f *osFile) Create() (Output, error) {
	path := f.FullPath()
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

func (f *osFile) CreateBytes(content []byte) error {
	output, err := f.Create()
	if err != nil {
		return err
	}
	_, err = output.Write(content)
	output.Close()
	return err
}

func (f *osFile) Read() (Input, error) {
	return os.Open(f.FullPath())
}

func (f *osFile) ReadBytes() ([]byte, error) {
	input, err := f.Read()
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

func (f *osFile) Stat() (Stat, error) {
	stat, err := os.Stat(f.FullPath())
	if err != nil {
		return Stat{}, err
	}
	return Stat{ModTime: stat.ModTime()}, nil
}

func (f *osFile) Chtime(mtime time.Time) error {
	return os.Chtimes(f.FullPath(), mtime, mtime)
}

func (f *osFile) ForAllFiles(fn ForAllFilesFunc) error {
	err := filepath.WalkDir(f.FullPath(), func(name string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		suffix, nerr := filepath.Rel(f.base, filepath.ToSlash(name))
		if nerr != nil {
			return nerr
		}
		return fn(f.GoTo(suffix), err)
	})
	if err == SkipRemaining {
		return nil
	}
	return err
}
