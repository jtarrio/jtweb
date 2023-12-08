package testing

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"jacobo.tarrio.org/jtweb/io"
)

var DefaultMtime = time.Date(2023, 1, 1, 12, 34, 56, 0, time.UTC)

type memoryFs struct {
	files map[string]memoryFsEntry
}

type memoryFsEntry struct {
	mtime   time.Time
	content []byte
}

type memoryFile struct {
	fs  *memoryFs
	rel string
}

func cleanRel(name string) string {
	return path.Join("/", name)[1:]
}

func NewMemoryFs() io.File {
	return &memoryFile{fs: &memoryFs{files: map[string]memoryFsEntry{}}, rel: ""}
}

func (f *memoryFile) Name() string {
	return f.rel
}

func (f *memoryFile) BaseName() string {
	return filepath.Base(f.rel)
}

func (f *memoryFile) FullPath() string {
	return "/memory/" + f.rel
}

func (f *memoryFile) GoTo(name string) io.File {
	return &memoryFile{fs: f.fs, rel: cleanRel(path.Join(f.rel, name))}
}

func (f *memoryFile) Create() (io.Output, error) {
	return &memoryOutput{file: f}, nil
}

type memoryOutput struct {
	bytes.Buffer
	file *memoryFile
}

func (o *memoryOutput) Close() error {
	o.file.fs.files[o.file.rel] = memoryFsEntry{mtime: DefaultMtime, content: o.Bytes()}
	return nil
}

func (f *memoryFile) CreateBytes(content []byte) error {
	f.fs.files[f.rel] = memoryFsEntry{mtime: DefaultMtime, content: content}
	return nil
}

func (f *memoryFile) Read() (io.Input, error) {
	b, ok := f.fs.files[f.rel]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", f.rel)
	}
	return &memoryInput{Reader: *bytes.NewReader(b.content), file: f}, nil
}

type memoryInput struct {
	bytes.Reader
	file *memoryFile
}

func (*memoryInput) Close() error {
	return nil
}

func (f *memoryFile) ReadBytes() ([]byte, error) {
	b, ok := f.fs.files[f.rel]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", f.rel)
	}
	return b.content, nil
}

func (f *memoryFile) Stat() (io.Stat, error) {
	b, ok := f.fs.files[f.rel]
	if !ok {
		return io.Stat{}, fmt.Errorf("file does not exist: %s", f.rel)
	}
	return io.Stat{ModTime: b.mtime}, nil
}

func (f *memoryFile) Chtime(mtime time.Time) error {
	b, ok := f.fs.files[f.rel]
	if !ok {
		return fmt.Errorf("file does not exist: %s", f.rel)
	}
	b.mtime = mtime
	f.fs.files[f.rel] = b
	return nil
}

func (f *memoryFile) ForAllFiles(fn io.ForAllFilesFunc) error {
	for k := range f.fs.files {
		err := fn(f.GoTo(k), nil)
		if err == io.SkipRemaining {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}
