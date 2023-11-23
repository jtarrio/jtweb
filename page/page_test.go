package page

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestParseEmpty(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("")
	_, err := Parse("test", &buf)

	assert.Errorf(t, err, "missing title")
}

func TestParseMinimalHeader(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("<!--HEADER\n" +
		"title: \"The title\"\n" +
		"-->")
	p, err := Parse("test", &buf)

	root := ast.NewDocument()
	root.SetLines(text.NewSegments())
	assert.Nil(t, err)
	assert.Equal(t, HeaderData{
		Title:    "The title",
		Language: "en",
	}, p.Header)
}

func TestParseFullHeader(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("<!--HEADER\n" +
		"title: \"The title\"\n" +
		"language: \"gl\"\n" +
		"summary: \"The summary\"\n" +
		"episode: \"The episode\"\n" +
		"publish_date: \"2021-08-15 01:23 +0100\"\n" +
		"no_publish_date: true\n" +
		"author_name: \"The author\"\n" +
		"author_uri: \"The author uri\"\n" +
		"hide_author: true\n" +
		"tags: [\"tag1\", \"tag2\"]\n" +
		"no_index: true\n" +
		"old_uris: [\"old_uri1\", \"old_uri2\"]\n" +
		"translation_of: \"other_article\"\n" +
		"draft: true\n" +
		"-->")
	p, err := Parse("test", &buf)

	root := ast.NewDocument()
	root.SetLines(text.NewSegments())
	assert.Nil(t, err)
	assert.Equal(t, HeaderData{
		Title:           "The title",
		Language:        "gl",
		Summary:         "The summary",
		Episode:         "The episode",
		PublishDate:     time.Date(2021, 8, 15, 1, 23, 0, 0, time.FixedZone("", 3600)),
		HidePublishDate: true,
		AuthorName:      "The author",
		AuthorURI:       "The author uri",
		HideAuthor:      true,
		Tags:            []string{"tag1", "tag2"},
		NoIndex:         true,
		OldURI:          []string{"old_uri1", "old_uri2"},
		TranslationOf:   "other_article",
		Draft:           true,
	}, p.Header)
}

func TestParseHeaderPublishDate(t *testing.T) {
	type testCase struct {
		date     string
		expected time.Time
	}
	testCases := []testCase{
		{"20210815", time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC)},
		{"202108150123", time.Date(2021, 8, 15, 1, 23, 0, 0, time.UTC)},
		{"20210815012345", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"202108150123Z", time.Date(2021, 8, 15, 1, 23, 0, 0, time.UTC)},
		{"20210815012345Z", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"202108150123+0100", time.Date(2021, 8, 15, 1, 23, 0, 0, time.FixedZone("", 3600))},
		{"20210815012345+0100", time.Date(2021, 8, 15, 1, 23, 45, 0, time.FixedZone("", 3600))},
		{"2021-08-15", time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC)},
		{"2021-08-15T01:23", time.Date(2021, 8, 15, 1, 23, 0, 0, time.UTC)},
		{"2021-08-15T01:23:45", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"2021-08-15T01:23Z", time.Date(2021, 8, 15, 1, 23, 0, 0, time.UTC)},
		{"2021-08-15T01:23:45Z", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"2021-08-15T01:23+0100", time.Date(2021, 8, 15, 1, 23, 0, 0, time.FixedZone("", 3600))},
		{"2021-08-15T01:23:45+0100", time.Date(2021, 8, 15, 1, 23, 45, 0, time.FixedZone("", 3600))},
		{"2021-08-15 01:23", time.Date(2021, 8, 15, 1, 23, 0, 0, time.UTC)},
		{"2021-08-15 01:23:45", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"2021-08-15 01:23 +01:00", time.Date(2021, 8, 15, 1, 23, 0, 0, time.FixedZone("", 3600))},
		{"2021-08-15 01:23:45 +01:00", time.Date(2021, 8, 15, 1, 23, 45, 0, time.FixedZone("", 3600))},
		{"08/15/2021", time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC)},
		{"08/15/2021 01:23:45", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"15-08-2021", time.Date(2021, 8, 15, 0, 0, 0, 0, time.UTC)},
		{"15-08-2021 01:23:45", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"15-08-2021 1:23:45AM", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"15-08-2021 1:23:45PM", time.Date(2021, 8, 15, 13, 23, 45, 0, time.UTC)},
		{"15-08-2021 1:23:45am", time.Date(2021, 8, 15, 1, 23, 45, 0, time.UTC)},
		{"15-08-2021 1:23:45pm", time.Date(2021, 8, 15, 13, 23, 45, 0, time.UTC)},
		{"15-08-2021 1:23:45am -07:00", time.Date(2021, 8, 15, 1, 23, 45, 0, time.FixedZone("", -25200))},
		{"15-08-2021 1:23:45am -07:00 PDT", time.Date(2021, 8, 15, 1, 23, 45, 0, time.FixedZone("PDT", -25200))},
	}
	for _, testCase := range testCases {
		var buf bytes.Buffer
		buf.WriteString("<!--HEADER\n" +
			"title: \"The title\"\n" +
			"publish_date: \"" + testCase.date + "\"\n" +
			"-->")
		p, err := Parse("test", &buf)
		assert.Nil(t, err)
		assert.Equal(t, testCase.expected, p.Header.PublishDate, "for %s", testCase.date)
	}
}

func TestParseHeaderErrorOnBadDate(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("<!--HEADER\n" +
		"title: \"The title\"\n" +
		"publish_date: \"foo bar\"\n" +
		"-->")
	_, err := Parse("test", &buf)

	assert.Errorf(t, err, "invalid date format")
}

func TestRenderSimple(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("<!--HEADER\n" +
		"title: \"The title\"\n" +
		"-->\n" +
		"Testing.")
	p, err := Parse("test", &buf)
	assert.Nil(t, err)
	var html strings.Builder
	err = p.Render(&html)
	assert.Nil(t, err)
	assert.Equal(t, "<p>Testing.</p>\n", html.String())
}

func DisabledTestRenderDangerous(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("<!--HEADER\n" +
		"title: \"The title\"\n" +
		"-->\n" +
		"Testing.\n\n" +
		"<script>window.alert('');</script>")
	p, err := Parse("test", &buf)
	assert.Nil(t, err)
	var html strings.Builder
	err = p.Render(&html)
	assert.Nil(t, err)
	assert.Equal(t, "<p>Testing.</p>\n", html.String())
}
