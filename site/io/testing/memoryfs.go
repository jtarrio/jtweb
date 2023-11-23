package testing

import (
	"bytes"
	"fmt"
	"path"
	"time"

	"jacobo.tarrio.org/jtweb/site/io"
)

var defaultMtime = time.Date(2023, 1, 1, 12, 34, 56, 0, time.UTC)

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

func (i *memoryFile) PathName() string {
	return i.rel
}

func (i *memoryFile) GoTo(name string) io.File {
	return &memoryFile{fs: i.fs, rel: cleanRel(path.Join(i.rel, name))}
}

func (i *memoryFile) Create() (io.Output, error) {
	return &memoryOutput{file: i}, nil
}

type memoryOutput struct {
	bytes.Buffer
	file *memoryFile
}

func (o *memoryOutput) Close() error {
	o.file.fs.files[o.file.rel] = memoryFsEntry{mtime: defaultMtime, content: o.Bytes()}
	return nil
}

func (i *memoryFile) CreateBytes(content []byte) error {
	i.fs.files[i.rel] = memoryFsEntry{mtime: defaultMtime, content: content}
	return nil
}

func (i *memoryFile) Read() (io.Input, error) {
	b, ok := i.fs.files[i.rel]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", i.rel)
	}
	return &memoryInput{Reader: *bytes.NewReader(b.content), file: i}, nil
}

type memoryInput struct {
	bytes.Reader
	file *memoryFile
}

func (*memoryInput) Close() error {
	return nil
}

func (i *memoryFile) ReadBytes() ([]byte, error) {
	b, ok := i.fs.files[i.rel]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", i.rel)
	}
	return b.content, nil
}

func (i *memoryFile) Stat() (io.Stat, error) {
	b, ok := i.fs.files[i.rel]
	if !ok {
		return io.Stat{}, fmt.Errorf("file does not exist: %s", i.rel)
	}
	return io.Stat{ModTime: b.mtime}, nil
}

func (i *memoryFile) Chtime(mtime time.Time) error {
	b, ok := i.fs.files[i.rel]
	if !ok {
		return fmt.Errorf("file does not exist: %s", i.rel)
	}
	b.mtime = mtime
	i.fs.files[i.rel] = b
	return nil
}

func (i *memoryFile) ForAllFiles(fn io.ForAllFilesFunc) error {
	for k := range i.fs.files {
		err := fn(i.GoTo(k), nil)
		if err == io.SkipRemaining {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}
