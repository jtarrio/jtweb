package languages

import (
	"fmt"
	"time"
)

// Language provides methods for language-specific strings and formatting.
type Language interface {
	// Name returns the language's name in the language itself.
	Name() string
	// Code provides the language's ISO code.
	Code() string
	// FormatDate formats the given time as a date, in "September 2, 2020" format.
	FormatDate(t time.Time) string
	// PreferredLanguage takes a list of languages and returns which one is often preferred by speakers of this language.
	PreferredLanguage(languages []string) string
}

type languageBase struct {
	// The language's name.
	name string
	// The language's ISO code.
	code string
	// List of preferred languages in order (best one first.)
	preferredLanguages []string
}

var languages map[string]Language = map[string]Language{
	"en": LanguageEn,
	"es": LanguageEs,
	"gl": LanguageGl,
}

// FindByCode returns the Language object corresponding to the language code.
func FindByCode(code string) (Language, error) {
	l, ok := languages[code]
	if !ok {
		return nil, fmt.Errorf("No language available for \"%s\"", code)
	}
	return l, nil
}

// FindByCodeWithFallback works like FindByCode, but it returns a fallback language if the code was not found.
func FindByCodeWithFallback(name string, fallback Language) Language {
	lang, err := FindByCode(name)
	if err != nil {
		return fallback
	}
	return lang
}

func (l languageBase) Name() string {
	return l.name
}

func (l languageBase) Code() string {
	return l.code
}

func (l languageBase) PreferredLanguage(languages []string) string {
	langSet := make(map[string]bool)
	for _, lang := range languages {
		langSet[lang] = true
	}
	if langSet[l.Code()] {
		return l.Code()
	}
	for _, lang := range l.preferredLanguages {
		if langSet[lang] {
			return lang
		}
	}
	return ""
}
