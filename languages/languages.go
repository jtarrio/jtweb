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
	// If none of them is preferred, the first language is returned.
	PreferredLanguage(languages []Language) Language
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

// AllLanguages returns a list with all known languages.
func AllLanguages() []Language {
	out := []Language{}
	for _, lang := range languages {
		out = append(out, lang)
	}
	return out
}

// FindByCode returns the Language object corresponding to the language code.
func FindByCode(code string) (Language, error) {
	l, ok := languages[code]
	if !ok {
		return nil, fmt.Errorf("no language available for \"%s\"", code)
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

func (l *languageBase) Name() string {
	return l.name
}

func (l *languageBase) Code() string {
	return l.code
}

func (l *languageBase) PreferredLanguage(languages []Language) Language {
	langSet := make(map[string]Language)
	for _, lang := range languages {
		langSet[lang.Code()] = lang
	}
	lang, ok := langSet[l.Code()]
	if ok {
		return lang
	}
	for _, code := range l.preferredLanguages {
		lang, ok := langSet[code]
		if ok {
			return lang
		}
	}
	return languages[0]
}

type LanguageSlice []Language

func (s LanguageSlice) Len() int {
	return len(s)
}

func (s LanguageSlice) Less(i, j int) bool {
	return s[i].Code() < s[j].Code()
}

func (s LanguageSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
