package main

import (
	"flag"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ClFlags is an struc that holds cl flags info
type ClFlags struct {
	collection string
}

// Config is a parent object for all the different configuration parts
type Config struct {
	mongo mongoConfig
}

// NewConfig initializes and parses the config file and/or environment using
// the viper library.
func NewConfig() *Config {
	parseConfig()

	c := &Config{}
	c.readConfig()

	return c
}

// getCLflags returns the CL args for the collection name
func getCLflags() ClFlags {

	flag.String("collection", "accessionId", "metadata object to retrieve")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("Could not bind process flags for commandline: %v", err)
	}

	collection := viper.GetString("collection")

	return ClFlags{collection: collection}

}

// configmongo populates a mongoConfig
func configMongo() mongoConfig {
	mongo := mongoConfig{}
	mongo.authMechanism = viper.GetString("mongo.authMechanism")
	mongo.database = viper.GetString("mongo.database")
	mongo.host = viper.GetString("mongo.host")
	mongo.port = viper.GetInt("mongo.port")
	mongo.user = viper.GetString("mongo.user")
	mongo.password = viper.GetString("mongo.password")

	if viper.IsSet("mongo.cacert") {
		mongo.caCert = viper.GetString("mongo.cacert")
	}

	return mongo
}

func (c *Config) readConfig() {

	c.mongo = configMongo()

	if viper.IsSet("loglevel") {
		stringLevel := viper.GetString("loglevel")
		intLevel, err := log.ParseLevel(stringLevel)
		if err != nil {
			log.Printf("Log level '%s' not supported, setting to 'trace'", stringLevel)
			intLevel = log.TraceLevel
		}
		log.SetLevel(intLevel)
		log.Printf("Setting log level to '%s'", stringLevel)
	}
}

func parseConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yaml")
	if viper.IsSet("configPath") {
		cp := viper.GetString("conifgPath")
		ss := strings.Split(strings.TrimLeft(cp, "/"), "/")
		viper.AddConfigPath(path.Join(ss...))
	}
	if viper.IsSet("configFile") {
		viper.SetConfigFile(viper.GetString("configFile"))
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Infoln("No config file found, using ENVs only")
		} else {
			log.Fatalf("Error when reading config file: '%s'", err)
		}
	}
}
