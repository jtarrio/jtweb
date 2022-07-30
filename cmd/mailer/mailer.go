package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/page"
	"jacobo.tarrio.org/jtweb/renderer/templates"
	"jacobo.tarrio.org/jtweb/site"
)

var flagLanguage = flag.String("language", "", "The language to send mail for.")
var flagSendAfter = flag.String("send_after", "", "Schedule posts newer than the given date/time.")

func main() {
	flag.Parse()

	var language string
	if *flagLanguage != "" {
		language = *flagLanguage
	} else {
		panic(fmt.Errorf("--language was not specified"))
	}

	sendAfter := time.Now()
	if *flagSendAfter != "" {
		parsed, err := time.Parse(time.RFC3339, *flagSendAfter)
		if err != nil {
			panic(err)
		}
		sendAfter = parsed
	}

	cfg, err := site.FromFlags()
	if err != nil {
		panic(err)
	}
	content, err := cfg.Read()
	if err != nil {
		panic(err)
	}

	toc, ok := content.Toc[*flagLanguage]
	if !ok {
		panic(fmt.Errorf("no table of contents for language: %s", *flagLanguage))
	}

	var pages []*page.Page
	for _, name := range toc.All {
		page := content.Pages[name]
		if page.Header.PublishDate.After(sendAfter) {
			println("Page: ", page.Name, " Date: ", page.Header.PublishDate.String())
			pages = append(pages, page)
		}
	}

	t := &templates.Templates{
		TemplatePath: content.TemplatePath,
		WebRoot:      content.GetWebRoot(language),
		Site:         templates.LinkData{Name: content.GetSiteName(language), URI: content.GetSiteURI(language)},
	}

	sb := strings.Builder{}
	for _, page := range pages {
		content.OutputPlainEmail(&sb, t, page)
	}

	println(sb.String())

}
