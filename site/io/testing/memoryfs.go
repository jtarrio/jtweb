package testing

import (
	"bytes"
	"fmt"
	"path"

	"jacobo.tarrio.org/jtweb/site/io"
)

type memoryFs struct {
	files map[string][]byte
}

type memoryFile struct {
	fs  *memoryFs
	rel string
}

func cleanRel(name string) string {
	return path.Join("/", name)[1:]
}

func NewMemoryFs(basePath string) io.File {
	return &memoryFile{fs: &memoryFs{}, rel: ""}
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
	o.file.fs.files[o.file.rel] = o.Bytes()
	return nil
}

func (i *memoryFile) Read() (io.Input, error) {
	b, ok := i.fs.files[i.rel]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", i.rel)
	}
	return &memoryInput{Reader: *bytes.NewReader(b), file: i}, nil
}

type memoryInput struct {
	bytes.Reader
	file *memoryFile
}

func (o *memoryInput) Close() error {
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
