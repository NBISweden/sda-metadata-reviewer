package main

import (
	"errors"
	"flag"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ClFlags is an struc that holds cl flags info
type ClFlags struct {
	action string
}

// Config is a parent object for all the different configuration parts
type Config struct {
	mongo    mongoConfig
	s3       S3Config
	postgres DBConfig
}

// NewConfig initializes and parses the config file and/or environment using
// the viper library.
func NewConfig() *Config {
	parseConfig()

	c := &Config{}
	c.readConfig()

	return c
}

// getCLflags returns the given CL options
func getCLflags() ClFlags {

	flag.String("action", "", "action to perform")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatalf("Could not bind process flags for commandline: %v", err)
	}

	action := viper.GetString("action")

	return ClFlags{action: action}

}

// configmongo populates a mongoConfig
func configMongo() mongoConfig {
	mongo := mongoConfig{}
	mongo.authMechanism = viper.GetString("mongo.authMechanism")
	mongo.host = viper.GetString("mongo.host")
	mongo.port = viper.GetInt("mongo.port")
	mongo.user = viper.GetString("mongo.user")
	mongo.password = viper.GetString("mongo.password")

	if viper.IsSet("mongo.cacert") {
		mongo.caCert = viper.GetString("mongo.cacert")
	}

	return mongo
}

// configmongo populates a mongoConfig
func configS3() S3Config {
	s3 := S3Config{}

	s3.URL = viper.GetString("s3.url")
	s3.AccessKey = viper.GetString("s3.accesskey")
	s3.SecretKey = viper.GetString("s3.secretkey")
	s3.Bucket = viper.GetString("s3.bucket")

	// Defaults (move to viper?)

	s3.Port = 443
	s3.Region = "us-east-1"
	s3.NonExistRetryTime = 2 * time.Minute

	if viper.IsSet("s3.port") {
		s3.Port = viper.GetInt("s3.port")
	}

	if viper.IsSet("s3.region") {
		s3.Region = viper.GetString("s3.region")
	}

	if viper.IsSet("s3.chunksize") {
		s3.Chunksize = viper.GetInt("s3.chunksize") * 1024 * 1024
	}

	if viper.IsSet("s3.cacert") {
		s3.Cacert = viper.GetString("s3.cacert")
	}

	return s3
}

// configDatabase provides configuration for the database
func configDatabase() (DBConfig, error) {
	db := DBConfig{}

	// All these are required
	db.Host = viper.GetString("db.host")
	db.Port = viper.GetInt("db.port")
	db.User = viper.GetString("db.user")
	db.Password = viper.GetString("db.password")
	db.Database = viper.GetString("db.database")
	db.SslMode = viper.GetString("db.sslmode")

	// Optional settings
	if db.SslMode == "verify-full" {
		// Since verify-full is specified, these are required.
		if !(viper.IsSet("db.clientCert") && viper.IsSet("db.clientKey")) {
			return db, errors.New("when db.sslMode is set to verify-full both db.clientCert and db.clientKey are needed")
		}
	}
	if viper.IsSet("db.clientKey") {
		db.ClientKey = viper.GetString("db.clientKey")
	}
	if viper.IsSet("db.clientCert") {
		db.ClientCert = viper.GetString("db.clientCert")
	}
	if viper.IsSet("db.cacert") {
		db.CACert = viper.GetString("db.cacert")
	}

	return db, nil
}

func (c *Config) readConfig() {

	c.mongo = configMongo()
	c.s3 = configS3()
	c.postgres, _ = configDatabase()

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
