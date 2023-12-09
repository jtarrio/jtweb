package fromflags

import (
	"flag"
	"fmt"
	"io/ioutil"

	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/config/yamlconfig"
	"jacobo.tarrio.org/jtweb/secrets/yamlsecrets"
)

var flagConfigFile = flag.String("config_file", "", "The name of the file containing the site's configuration.")
var flagOutputPath = flag.String("output_path", "", "The full pathname where the rendered HTML files will be output.")
var flagWebroot = flag.String("webroot", "", "The URI where the generated content will live.")
var flagPublishUntil = TimeFlag("publish_until", "Publish all posts older than the given date/time.")
var flagSecretsYaml = flag.String("secrets_yaml", "", "The name of a YAML file containing secrets.")

func GetConfig() (config.Config, error) {
	if *flagConfigFile == "" {
		return nil, fmt.Errorf("the --config_file flag has not been specified")
	}
	file, err := ioutil.ReadFile(*flagConfigFile)
	if err != nil {
		return nil, err
	}
	reader := yamlconfig.NewConfigParser(file)

	if *flagSecretsYaml != "" {
		file, err := ioutil.ReadFile(*flagSecretsYaml)
		if err != nil {
			return nil, err
		}
		supplier, err := yamlsecrets.Parse(file)
		if err != nil {
			return nil, err
		}
		reader.WithSecretSupplier(supplier)
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
	return reader.Parse()
}
