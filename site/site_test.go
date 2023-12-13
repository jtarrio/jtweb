package site

import (
	"testing"

	configtesting "jacobo.tarrio.org/jtweb/config/testing"
	iotesting "jacobo.tarrio.org/jtweb/io/testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptySite(t *testing.T) {
	config := configtesting.NewFakeConfig()
	rawContent, err := Read(config)
	if err != nil {
		panic(err)
	}
	content, err := rawContent.Index(nil, nil)
	if err != nil {
		panic(err)
	}
	err = content.Write()
	if err != nil {
		panic(err)
	}
	assert.Empty(t, iotesting.GetFileNames(config.OutputBase))
}

func TestCopiesFiles(t *testing.T) {
	config := configtesting.NewFakeConfig()
	err := config.InputBase.GoTo("file.txt").
		CreateBytes([]byte("This is file.txt"))
	if err != nil {
		panic(err)
	}

	rawContent, err := Read(config)
	if err != nil {
		panic(err)
	}
	content, err := rawContent.Index(nil, nil)
	if err != nil {
		panic(err)
	}

	err = content.Write()
	if err != nil {
		panic(err)
	}
	actual, err := config.OutputBase.GoTo("file.txt").ReadBytes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "This is file.txt", string(actual))
}

func TestAppliesTemplates(t *testing.T) {
	config := configtesting.NewFakeConfig()
	err := config.InputBase.GoTo("template.html.tmpl").
		CreateBytes([]byte("<html><body>Webroot: {{webRoot \"en\"}}</body></html>\n"))
	if err != nil {
		panic(err)
	}

	rawContent, err := Read(config)
	if err != nil {
		panic(err)
	}
	content, err := rawContent.Index(nil, nil)
	if err != nil {
		panic(err)
	}

	err = content.Write()
	if err != nil {
		panic(err)
	}
	actual, err := config.OutputBase.GoTo("template.html").ReadBytes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "<html><body>Webroot: http://webroot</body></html>\n", string(actual))
}

func TestParsesMarkdown(t *testing.T) {
	config := configtesting.NewFakeConfig()
	err := config.TemplateBase.GoTo("page-en.tmpl").
		CreateBytes([]byte("<html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>"))
	if err != nil {
		panic(err)
	}
	err = config.TemplateBase.GoTo("toc-en.tmpl").
		CreateBytes([]byte("<html><body>{{range .Stories}}<p>{{.Title}}</p>{{end}}</body></html>"))
	if err != nil {
		panic(err)
	}
	err = config.InputBase.GoTo("markdown.md").CreateBytes([]byte("<!--HEADER\n" +
		"title: The Title\n" +
		"-->\n" +
		"The content\n"))
	if err != nil {
		panic(err)
	}

	rawContent, err := Read(config)
	if err != nil {
		panic(err)
	}
	content, err := rawContent.Index(nil, nil)
	if err != nil {
		panic(err)
	}

	err = content.Write()
	if err != nil {
		panic(err)
	}
	actual, err := config.OutputBase.GoTo("markdown.html").ReadBytes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "<html><head><title>The Title</title></head><body><p>The content</p>\n</body></html>", string(actual))
	actual, err = config.OutputBase.GoTo("toc/toc-en.html").ReadBytes()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "<html><body><p>The Title</p></body></html>", string(actual))
}
