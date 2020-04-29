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

These configuration options are available:

* `input_path` --- The full pathname where the input files are located.
* `output_path` --- The full pathname where the generated site will be written.
* `template_path` --- The full pathname for the templates used to convert Markdown files.
* `webroot` --- The base URI for the generated site. I recommend not including a trailing slash.
* `site_name` --- The site's name.
* `site_uri` --- The site's URI, if different from the webroot.
* `author_name` --- The default author's name.
* `author_uri` --- The default author's website URI, if different from the site URI.

The configuration file will look like this:

```yaml
template_path: "src/templates"
input_path: "src/content"
webroot: "https://jacobo.tarrio.org"
site_name: "Jacobo Tarrío"
site_uri: "https://jacobo.tarrio.org/"
author_name: "Jacobo Tarrío"
```

The command-line options will look like this:

```
--output_path="/var/www/jacobo.tarrio.org/website"
```

# How different files are handled

In general, the site generator will copy any files found in the `input_path`
under the corresponding path in the `output_path`. The exceptions are Markdown
files, Go template files, and `.htaccess` files.

## Markdown files

Markdown files have an `.md` extension. Their syntax is GitHub-Flavored
Markdown, with a few extensions.

The name of a Markdown file is its path relative to the `input_path`, minus
its `.md` extension. The output file name its formed by adding the `.html`
extension to the Markdown file's name.

Markdown files are rendered through the template files found in the
`template_path`. There must be a set of `page-LANG.tmpl`, `toc-LANG.tmpl`,
and `index-toc-LANG.tmpl` files for each language that you have pages in.

### Extensions

#### Header

There must be a comment block including the header data. This comment block
starts with the `<!--HEADER` pseudo-tag and contains a YAML definition for
all the header data.

```html
<!--HEADER
title: The page's title
publish_date: 2020-04-01
language: en
tags:
- website
- humor
-->
```

These are the fields that are available for the header:

* `title` --- the page's title. This field is mandatory.
* `summary` --- a one-line summary for use in indices.
* `language` --- the ISO code for the page's language. "en" is assumed if omitted.
* `publish_date` --- the date (and, optionally, time) when the page was posted.
* `hide_publish_date` --- boolean; if true, the publish date is not shown.
* `draft` --- boolean; if true, the page is not indexed and it's marked as a draft.
* `author_name` --- the author's name. If omitted, the default author name is used.
* `author_uri` --- the author's website URI. If omitted, the default author URI is used.
* `hide_author` --- boolean; if true, the author's name and URI are not shown.
* `tags` --- a list of strings containing tags to index the page under.
* `no_index` --- boolean; if true, the page is not indexed.
* `old_uri` --- a list of strings containing old URIs for this page that should be redirected to its new location. Those URIs must be relative to the webroot.
* `translation_of` --- the name of another page this page is a translation of.

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
have the names `page-LANG.tmpl`, `toc-LANG.tmpl`, and `index-toc-LANG.tmpl`,
where `LANG` is the ISO code specified in the page header's `language` field.

#### `page-LANG.tmpl`

This template is used to render the page itself. It is rendered from a 
`templates.PageData` structure that contains the following fields:

* `Title` --- the page's title.
* `Permalink` --- the page's permalink.
* `Author` --- a `templates.LinkData` structure containing the author's name and website URI.
* `Summary` --- the page's one-line summary.
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

* `BaseURI` --- the URI of the current page, minus the year and the `.html` extension. Used to render links to other indices for other years.
* `Tag` --- the current tag, if any.
* `Year` --- the current year. It's 0 if it's an index for undated stories.
* `YearCount` --- the number of stories in the current year.
* `TotalCount` --- the total number of indexed stories.
* `Stories` --- an array of `templates.PageData` structures with the story data for the current page.
* `StoryYears` --- an array containing other years that have stories.
* `UndatedStories` --- a boolean indicating if there are undated stories in the index.

#### `index-toc-LANG.tmpl`

This template is used to render a list of years with stories, either
general or for a particular tag. It is rendered from a `templates.IndexTocData`
structure that contains the following fields:

* `BaseURI` --- the URI of the current page, minus the `.html` extension. Used to render links to indices for the different years.
* `Tag` --- the current tag, if any.
* `TotalCount` --- the total number of indexed stories.
* `Years` --- an array of `templates.YearData` structures that have per-year information.

The `templates.YearData` structure has the following fields:

* `Year` --- the current year, or 0 for undated stories.
* `Count` --- the number of stories in this year.
* `Tags` --- an array of tag names the stories in this year belong to.

#### Template functions

Several functions are available in these three templates:

* `formatDate` --- takes a `time.Time` structure and formats it according to the current language.
* `getTagTocURI` --- takes a tag name and returns the URI of its master table of contents file.
* `getTagURIWithTime` --- takes a tag name and `time.Time` structure and returns the URI of its table of contents file for the year.
* `getTagURIWithYear` --- takes a tag name and a year and returns the URI of its table of contents file for the year.
* `getTocURI` --- returns the URI of the master table of contents file.
* `getTocURIWithTime` --- takes a `time.Time` structure and returns the URI of the table of contents for the year.
* `getTocURIWithYear` --- takes a year and returns the URI of the table of contents for the year.
* `getURI` --- takes a relative URI and makes it absolute to the webroot.
* `language` --- returns the current language's ISO code.
* `plural` --- takes a number, a singular form and a plural form, and returns either the singular or plural form depending on whether the number is 1 or not.
* `site` --- returns a `template.LinkData` structure containing the current site's name and URI.
* `webRoot` --- returns the webroot URI.
* `year` --- takes a `time.Time` structure and returns the year, or 0 for a Zero time.

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
