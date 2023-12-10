package fromflags

import (
	"flag"
	"fmt"
	"io/ioutil"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/config/yamlconfig"
	"jacobo.tarrio.org/jtweb/secrets/secretsdir"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagPublishUntil = TimeFlag("publish_until", "Publish all posts older than the given date/time.")
var flagSecretsDir = flag.String("secrets_dir", "", "The name of a directory containing secrets files.")
var flagDryRun = flag.Bool("dry_run", false, "Do not perform the operations.")

func GetConfig() (config.Config, error) {
	if *flagConfigFile == "" {
		return nil, fmt.Errorf("the --config_file flag has not been specified")
	}
	file, err := ioutil.ReadFile(*flagConfigFile)
	if err != nil {
		return nil, err
	}
	reader := yamlconfig.NewConfigParser(file)

	if *flagSecretsDir != "" {
		reader.WithSecretSupplier(secretsdir.Create(*flagSecretsDir))
	}

	if *flagWebroot != "" {
		reader = reader.WithOptions(yamlconfig.OverrideWebroot(*flagWebroot))
	}
	if *flagOutputPath != "" {
		reader = reader.WithOptions(yamlconfig.OverrideOutput(*flagOutputPath))
	}
	if !flagPublishUntil.IsZero() {
		reader = reader.WithOptions(yamlconfig.OverridePublishUntil(*flagPublishUntil))
	}
	if *flagDryRun {
		reader = reader.WithOptions(yamlconfig.OverrideDryRun(true))
	}
	return reader.Parse()
}
