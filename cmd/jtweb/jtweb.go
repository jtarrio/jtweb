package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/email/generator"
	"jacobo.tarrio.org/jtweb/site"
)

var flagOperations = flag.String("operations", "*",
	"The names of the operations that must be performed. "+
		"This is a list of items separated with commas. "+
		"The asterisk ('*') is a wildcard. "+
		"Prepend with '-' to remove instead of adding. "+
		"Skipped operations don't match any wildcards for adding, but they "+
		"match normally for removing. To add a skipped operation, include its "+
		"full name without wildcards. "+
		"Use 'list' to view all available operations.")

type operation struct {
	name        string
	description string
	skipped     bool
	operate     func(content *site.RawContents) error
}

func getTimeOrDefault(when *time.Time, def time.Time) *time.Time {
	if when == nil {
		return &def
	}
	return when
}

func getAvailableOperations(cfg config.Config) []operation {
	ops := []operation{}
	if cfg.Generator() != nil {
		ops = append(ops, operation{
			name:        "generate",
			description: "Generate the website",
			skipped:     cfg.Generator().SkipOperation(),
			operate: func(rawContent *site.RawContents) error {
				notAfter := getTimeOrDefault(rawContent.Config.DateFilters().Generate().NotAfter(), rawContent.Config.DateFilters().Now())
				content, err := rawContent.Index(nil, notAfter)
				if err != nil {
					return err
				}
				return content.Write()
			},
		})
	}
	for _, mailer_iter := range cfg.Mailers() {
		// Make a copy of the mailer.
		mailer := mailer_iter
		ops = append(ops, operation{
			name:        fmt.Sprintf("email=%s", mailer.Name()),
			description: fmt.Sprintf("Send emails for '%s' with language '%s' and engine '%s'", mailer.Name(), mailer.Language().Code(), mailer.Engine().Name()),
			skipped:     mailer.SkipOperation(),
			operate: func(rawContent *site.RawContents) error {
				notBefore := getTimeOrDefault(rawContent.Config.DateFilters().Mail().NotBefore(), rawContent.Config.DateFilters().Now())
				notAfter := rawContent.Config.DateFilters().Mail().NotAfter()
				content, err := rawContent.Index(notBefore, notAfter)
				if err != nil {
					return err
				}
				return generator.NewEmailGenerator(content, mailer.Language(), mailer.Engine()).
					WithOptions(
						generator.NotScheduled(),
						generator.NamePrefix(mailer.SubjectPrefix()),
						generator.SubjectPrefix(mailer.SubjectPrefix()),
					).SendMails()
			},
		})
	}
	return ops
}

func selectOperations(filter string, operations []operation) []operation {
	names := make([]string, len(operations))
	skipped_names := map[string]bool{}
	for i, op := range operations {
		names[i] = op.name
		if op.skipped {
			skipped_names[op.name] = true
		}
	}
	wanted_names := map[string]bool{}
	for _, f := range strings.Split(filter, ",") {
		add := true
		if len(f) > 0 && f[0] == '-' {
			add = false
			f = f[1:]
		}
		for _, name := range names {
			if match(f, name) && (!add || f == name || !skipped_names[name]) {
				wanted_names[name] = add
			}
		}
	}
	ops := []operation{}
	for _, op := range operations {
		if wanted_names[op.name] {
			ops = append(ops, op)
		}
	}
	return ops
}

func listOperations(operations []operation) {
	println("Available operations:")
	for _, op := range operations {
		if op.skipped {
			fmt.Printf(" - %s (skipped)\n   %s\n", op.name, op.description)
		} else {
			fmt.Printf(" - %s\n   %s\n", op.name, op.description)
		}
	}
}

func match(pattern, name string) (matched bool) {
	if pattern == "" {
		return name == pattern
	}

	if pattern == "*" {
		return true
	}
	return deepMatchRune([]rune(name), []rune(pattern))
}

func deepMatchRune(str, pattern []rune) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		default:
			if len(str) == 0 || str[0] != pattern[0] {
				return false
			}
		case '*':
			return deepMatchRune(str, pattern[1:]) ||
				(len(str) > 0 && deepMatchRune(str[1:], pattern))
		}

		str = str[1:]
		pattern = pattern[1:]
	}

	return len(str) == 0 && len(pattern) == 0
}

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}

	operations := getAvailableOperations(cfg)
	if *flagOperations == "list" {
		listOperations(operations)
		return
	}
	operations = selectOperations(*flagOperations, operations)

	content, err := site.Read(cfg)
	if err != nil {
		panic(err)
	}

	for _, op := range operations {
		err := op.operate(content)
		if err != nil {
			log.Print(err)
		}
	}
}
