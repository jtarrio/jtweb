package renderer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	r := strings.NewReader(`<p>Paragraph.</p>
<div class="foo bar" onclick="javascript:0">Div<span class="uno dos">Span</span></div>
<p style="border: 1px">Style.</p>
<p><a href="http://example.com" title="Title">Link</a></p>
<script>window.alert()</script>
<iframe src="http://example.com" width=56 height=98 class="f" allowed="xxx"></iframe>`)
	w := strings.Builder{}

	err := SanitizePost(&w, r, "http://site/base", "path/index.html")
	if err != nil {
		panic(err)
	}

	expected := `<p>Paragraph.</p>
<div class="foo bar">Div<span class="uno dos">Span</span></div>
<p style="border: 1px">Style.</p>
<p><a href="http://example.com" title="Title">Link</a></p>

<iframe src="http://example.com" width="56" height="98" class="f"></iframe>`
	assert.Equal(t, expected, w.String())
}

func TestMakeUrisAbsolute(t *testing.T) {
	r := strings.NewReader(`<p><a href="relative.html">Relative</a></p>
<p><a href="/site-absolute.html">Site absolute</a></p>
<p><a href="https://absolute/path.html">Absolute</a></p>
<img src="relative.jpg"/>
<img src="/site-absolute.jpg"/>
<img src="https://absolute/image.jpg"/>`)

	w := strings.Builder{}

	err := SanitizePost(&w, r, "http://site/base/", "path/index.html")
	if err != nil {
		panic(err)
	}

	expected := `<p><a href="http://site/base/path/relative.html">Relative</a></p>
<p><a href="http://site/base/site-absolute.html">Site absolute</a></p>
<p><a href="https://absolute/path.html">Absolute</a></p>
<img src="http://site/base/path/relative.jpg"/>
<img src="http://site/base/site-absolute.jpg"/>
<img src="https://absolute/image.jpg"/>`
	assert.Equal(t, expected, w.String())
}
