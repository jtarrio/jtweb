# Jacobo Tarrío's website generator

This generator was built for blogs or newsletter websites consisting of a
series of dated posts that follow a common style.

It will render posts written in GitHub-Flavored Markdown and Go template files,
and it will copy any other file type.

Posts may be translated, and the translations for each post will be interlinked.

It will use the information from the posts' headers to build time-based and
tag-based tables of contents. The tables of contents will appear in every
language that is present in the website and can be configured to display
only articles written (or translated) in that language or articles written
in any available language.

RSS feeds will also be generated.

This generator can also schedule emails to be sent. You will need an external
system that will take care of managing subscriptions and sending the emails.
So far, only Mailerlite is supported.

Note: I built this for my personal site,
[jacobo.tarrio.org](https://jacobo.tarrio.org) and my newsletters,
[A Folla](https://folla.gal) and [Coding Sheet](https://coding-sheet.org).
Therefore, it fits my needs even though it is quite limited in many areas.
If you are looking for a website generator for yourself, please consider
[Hugo](https://gohugo.io/) and [Lume](https://lume.land/).

# Usage

The site generator is configured through a file, though you can override certain
values from the command line.

You must specify the name of the configuration file with the `--config_file`
command-line option.

## Configuration file

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
  # If true, the generator does not run by default though it can be enabled
  # through the --operations flag. Default: false.
  disabled: false

# Mail configurations
mailers:
  # A name for this configuration.
  - name: "galego"
    # If true, this mail configuration will not run by default though it can
    # be enabled through the --operations flag. You can use this for your
    # test configurations. Default: false.
    disabled: false
    # Send posts available in this language.
    language: "gl"
    # The subject line will contain this prefix followed by
    # the episode number and the page's title. Optional.
    subject_prefix: "O meu boletín"
    # Mailerlite configuration.
    mailerlite:
      # The name of the secret file containing the API key.
      apikey_secret: "mailerlite-key"
      # The group to send email to.
      group: 12345

# Date filters
date_filters:
  # Sets a maximum date for the posts that will be generated.
  # By default, the maximum date is the current date and time.
  # You would normally not specify it in your configuration,
  # and only override it for testing.
  generate:
    # A hard cutoff date. Defaults to the current date.
    # May be overridden with the --generate_not_after flag.
    not_after: "2023-08-01T00:00:00Z"
    # A rolling cutoff date, in days after today. May be negative.
    # Only used if not_before is unspecified.
    not_after_days: 14
  # Sets a minimum and maximum date for the posts that will be emailed.
  # By default, the minimum date is the current date and time, and there
  # is no maximum date.
  mail:
    # Do not send posts before this date. Defaults to the current date.
    # May be overridden with the --mail_not_before flag.
    not_before: "2023-08-01T00:00:00Z"
    # Do not send posts after this date. Defaults to the far future.
    # May be overridden with the --mail_not_after flag.
    not_after: "2023-08-01T00:00:00Z"
    # Do not send posts this many days after today. Optional.
    # Useful to only schedule posts for a few days at a time even if you
    # write content many months in advance.
    not_after_days: 14
```

## Secrets

If you use secrets in your configuration (such as in the `apikey_secret` field),
you must specify a secrets directory using the `--secrets_dir` command-line
flag.

This directory must contain one file for each secret, with the same name that
is used in the configuration file, and the content of this file is the secret
itself.

Be careful with line endings when you create the secret files! For example,
this command creates a file named `mailerlite-key` with an API key:

```
$ echo -n 'THE_MAILERLITE_API_KEY' > secrets/mailerlite-key
```

The `-n` flag to `echo` prevents a newline at the end of the API key.

## Command-line flags

The following command-line flags are available.

* `--config_file` -- The name of the file that contains the configuration.
* `--secrets_dir` -- The name of the directory containing secrets files.
* `--operations` -- The names of the operations that must be performed,
  or the word "list".

  This is a list of items to be added, separated using commas. If an item is
  prepended with a minus sign (`-`), it is removed instead of added. You may
  use wildcards to match several operations at once for adding or for removing.

  Some operations appear as "disabled" in the list, and by default they are not
  added. You can add them by specifying them with their exact name (no
  wildcards). You can remove them by exact name or by wildcard, though.

  The default value is `*`, which adds all (non-disabled) operations.

  Examples:
  * `--operations=*,-mail=` -- adds all operations and then removes all the "mail" operations.
  * `--operations=*,mail=test` -- adds all operations and also the (possibly disabled) operation `mail=test`.
  * `--operations=` -- no operations (empty list).
  * `--operations=mail=*,-mail=foo` -- adds all "mail" operations except `mail=foo`.
* `--webroot` -- Override the `site.webroot` configuration for every language.
* `--generate_not_after` -- Override the `date_filters.generate.not_after` configuration.
* `--mail_not_before` -- Override the `date_filters.mail.not_before` configuration.
* `--mail_not_after` -- Override the `date_filters.mail.not_after` configuration.
* `--dry_run` -- Simulate the file generation and email scheduling operations, printing out what would happen.

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

Copyright 2020 Jacobo Tarrío.
Distributed under the terms of the Apache License version 2.0.
