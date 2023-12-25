package lib

import "jacobo.tarrio.org/jtweb/site"

func OpGenerate() OpFn {
	return func(rawContent *site.RawContents) error {
		notAfter := getTimeOrDefault(rawContent.Config.DateFilters().Generate().NotAfter(), rawContent.Config.DateFilters().Now())
		content, err := rawContent.Index(nil, notAfter)
		if err != nil {
			return err
		}
		return content.Write()
	}
}
