package testing

import "jacobo.tarrio.org/jtweb/io"

func GetFileNames(base io.File) []string {
	out := []string{}
	base.ForAllFiles(func(file io.File, err error) error {
		out = append(out, file.Name())
		return nil
	})
	return out
}
