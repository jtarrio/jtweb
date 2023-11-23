package testing

import "jacobo.tarrio.org/jtweb/site/io"

func GetFileNames(base io.File) []string {
	out := []string{}
	base.ForAllFiles(func(file io.File, err error) error {
		out = append(out, file.PathName())
		return nil
	})
	return out
}