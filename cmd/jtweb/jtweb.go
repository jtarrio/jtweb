package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

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
		"Use 'list' to view all available operations.")

type operation struct {
	name        string
	description string
	operate     func(content *site.Contents) error
}

func getAvailableOperations(cfg config.Config) []operation {
	ops := []operation{}
	if cfg.Generator() != nil {
		ops = append(ops, operation{
			name:        "generate",
			description: "Generate the website",
			operate: func(content *site.Contents) error {
				return content.Write()
			},
		})
	}
	for _, mailer := range cfg.Mailers() {
		ops = append(ops, operation{
			name:        fmt.Sprintf("email=%s", mailer.Name()),
			description: fmt.Sprintf("Send emails for '%s' with language '%s' and engine '%s'", mailer.Name(), mailer.Language().Code(), mailer.Engine().Name()),
			operate: func(content *site.Contents) error {
				return generator.NewEmailGenerator(content, mailer.Language(), mailer.Engine()).
					WithOptions(
						generator.NotBefore(mailer.SendAfter()),
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
	for i, op := range operations {
		names[i] = op.name
	}
	wanted_names := map[string]bool{}
	for _, f := range strings.Split(filter, ",") {
		add := true
		if len(f) > 0 && f[0] == '-' {
			add = false
			f = f[1:]
		}
		for _, name := range names {
			if match(f, name) {
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
		println("Available operations:")
		for _, op := range operations {
			fmt.Printf(" - %s\n   %s\n", op.name, op.description)
		}
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
