# Jacobo Tarrío's website generator

This generator can render GitHub-Flavored Markdown files and Go template
files, and copy any other file type.

It can also create chronological and thematic tables of contents.

# Usage

You can configure the site generator using a configuration file, a set
of command-line options, or a combination of them; command-line options
override the configuration file settings.

You can specify the name of the configuration file with the `--config_file`
command-line option.

The configuration file is YAML with the following format:

```yaml
# Source file configuration.
files:
  # Location of the site content files.
  content: "src/content"
  # Location of the templates for rendering Markdown files.
  templates: "src/templates"

# Site configuration.
site:
  # Base for all internal links.
  # May be overridden through the --webroot flag.
  webroot: "http://example.com/"
  # Name of the site.
  name: "My example site"
  # URI of the site. Optional. If omitted, defaults to the value of `webroot`
  uri: "https://example.com"
  # Language-dependent variants of the website. Optional. One per language.
  by_language:
    # Spanish.
    es:
      webroot: "http://example.com/es/"
      name: "Sitio de ejemplo"
      uri: "http://example.com/es"
    # Galician
    gl:
      name: "Sitio de exemplo"
      # Unspecified values are copied from the default (even `uri`, so beware).

# Author configuration.
author:
  # The site's author's name.
  name: "John Doe"
  # The site's author's webpage. Optional. If omitted, defaults to `site.uri`.
  uri: "https://example.com/johndoe"

# Website generator configuration. Only required to run the generator.
generator:
  # Path where the website will be written to.
  # May be overridden through the --output_path flag.
  output: "rendered"
  # If true, tables of contents in each language will only show content in
  # that language. If false, tables of contents will show content in other
  # languages if it is not available in the same language. Default: false.
  hide_untranslated: false
  # Only output content with publish-dates up to this date and time.
  # If unspecified, the current date and time will be used.
  # May be overridden through the --publish_until flag.
  publish_until: "2060-01-01T00:00"
```

# How different files are handled

In general, the site generator will copy any files found in `files.content`
under the corresponding path in the `generator.output`. The exceptions are
Markdown files, Go template files, and `.htaccess` files.

## Markdown files

Markdown files have an `.md` extension. Their syntax is GitHub-Flavored
Markdown, with a few extensions.

The name of a Markdown file is its path relative to `files.content`, minus
its `.md` extension. The output file name its formed by adding the `.html`
extension to the Markdown file's name.

Markdown files are rendered through the template files found in 
`files.templates`. There must be a set of `page-LANG.tmpl` and `toc-LANG.tmpl`
files for each language that you have pages in.

### Extensions

#### Header

There must be a comment block including the header data. This comment block
starts with the `<!--HEADER` pseudo-tag and contains a YAML block.

```yaml
<!--HEADER
# The page's title
title: "A nice article"
# A short summary of the page, shown in the table of contents.
summary: "A brief discussion of beautiful and calming things."
# (Optional) An episode number, for serial publications.
episode: "43"
# The language the page is written in. The default is `en` (English).
language: "es"
# The publication date for the page. Used to sort the table of contents.
publish_date: "2020-04-01 01:23"
# (Optional) If true, do not show the publication date. Default: false..
no_publish_date: true
# (Optional) The name of the page's author, to override the site-wide setting.
author_name: "John Doe"
# (Optional) The URI of the page's author's website, to override the site-wide setting.
author_uri: "http://example.com/johndoe"
# (Optional) If true, do not show the author's name in the page. Default: false.
hide_author: true
# (Optional) If true, do not add this page in tables of contents. Default: false.
no_index: true
# (Optional) The name of the page this is a translation of.
translation_of: "textos/bonito-articulo"
# (Optional) A list of tags for this page. This page will appear in those tag's tables of contents.
tags:
  - "Tag 1"
  - "Another tag"
  - "And yet another one"
# (Optional) A list of old URIs, relative to the site's webroot, where this article used to be.
# If provided, the old URIs will redirect to this page.
old_uris:
  - "texts/a-nice-article.html"
  - "a-beautiful-post.html"
  - "posts/id/1234"
# (Optional) If true, this page will not be indexed and a "draft" marker will appear. Default: false.
draft: true
-->
```

#### Image captions and grouping

When several images are specified one right after the other in the same
paragraph, they are grouped in a `<span class="multipleImgs">` element.

When there is content in the same paragraph after an image or after a group
of images, this content is wrapped in a `<span class="imageCaption">` element.

#### YouTube videos

You can add YouTube videos using the `!youtube(URI)` syntax.

### Rendering templates

Markdown files are rendered to HTML and written using the template files for
the page's language. The template files are found in the `template_path` and
have the names `page-LANG.tmpl` and `toc-LANG.tmpl`, where `LANG` is the ISO
code specified in the page header's `language` field.

#### `page-LANG.tmpl`

This template is used to render the page itself. It is rendered from a 
`templates.PageData` structure that contains the following fields:

* `Title` --- the page's title.
* `Permalink` --- the page's permalink.
* `Author` --- a `templates.LinkData` structure containing the author's name and website URI.
* `Summary` --- the page's one-line summary.
* `Episode` --- the episode number or name.
* `PublishDate` --- the page's publish date, or zero if it's been unspecified or hidden.
* `Tags` --- an array of tag names.
* `Content` --- the page's rendered content in HTML.
* `NewerPage` --- a `templates.LinkData` structure that points to the next newer page by publish date.
* `OlderPage` --- a `templates.LinkData` structure that points to the next older page by publish date.
* `Translations` --- an array of `templates.TranslationData` structures pointing to other translations of this page.
* `Draft` --- a boolean indicating whether this page is a draft.

`templates.LinkData` structures contain a `Name` field and a `URI` field.

`templates.TranslationData` structures contain a `Language` field,
a `Name` field and a `URI` field.

#### `toc-LANG.tmpl`

This template is used to render a table of contents for a particular year.
This table of contents could be general or for a particular tag. It is
rendered from a `templates.TocData` structure that contains the following
fields:

* `Tag` --- the current tag, if any.
* `TotalCount` --- the total number of indexed stories.
* `Stories` --- an array of `templates.PageData` structures with the story data for the current page.

#### Template functions

Several functions are available in these three templates:

* `formatDate` --- takes a `time.Time` structure and formats it according to the current language.
* `getTagURI` --- takes a tag name and returns the URI of its table of contents file.
* `getTocURI` --- returns the URI of the general table of contents file.
* `getURI` --- takes a relative URI and makes it absolute to the webroot.
* `language` --- returns the current language's ISO code.
* `plural` --- takes a number, a singular form and a plural form, and returns either the singular or plural form depending on whether the number is 1 or not.
* `site` --- returns a `template.LinkData` structure containing the current site's name and URI.
* `webRoot` --- returns the webroot URI.

## Go template files

Template files have a `.tmpl` extension that will be stripped off from
the output file's name. They are treated as Go HTML template files, as
defined in the Go `html/template` package.

Template files receive a `site.Contents` structure. The exact contents
of the structure are subject to change, so you should use the source
code to know what's in it and how to use it.

The templates can also use the following functions:

* `latestPage` --- receives a language's ISO code and returns a `page.Page` structure for the latest Markdown page in that language.
* `webRoot` --- returns the webroot URI.

## `.htaccess` files

These files are mostly copied as is, but you can insert `RewriteRule`
directives by inserting the following line:

```
### REDIRECTS ###
```

Where this line appears, it will be replaced with a series of
`RewriteRule` statements that redirect each page's `old_uri` URIs to
its current location.

Don't forget to enable `RewriteEngine on` in your `.htaccess` file or
your virtual host configuration!

---

Copyright 2020 Jacobo Tarrío
