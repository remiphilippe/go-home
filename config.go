package main

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// Config go-hone configuration structure
type Config struct {
	APISecret   string
	APIKey      string
	APIEndpoint string
	APIVerify   bool

	KafkaBroker []string
	KafkaRootCA string
	KafkaCert   string
	KafkaKey    string
	KafkaTopic  string
	KafkaSSL    bool
}

// NewConfig Creates a new configuration struct, return a *Config
func NewConfig() *Config {
	c := new(Config)

	// Configuration file management
	// name of config file (without extension)
	viper.SetConfigName("config")
	viper.AddConfigPath("conf/")
	viper.SetConfigType("toml")
	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil {
		// Handle errors reading the config file
		glog.Fatal("Fatal error config file: %s \n")
	}

	c.loadConfig()

	return c
}

func (c *Config) loadConfig() {
	//TODO need some error verification here
	glog.V(2).Infof("Loading OpenAPI Config...")
	o := viper.Sub("openapi")
	c.APISecret = o.GetString("secret")
	c.APIKey = o.GetString("key")
	c.APIEndpoint = o.GetString("endpoint")
	c.APIVerify = o.GetBool("verify")

	glog.V(2).Infof("Loading Kafka Config...")
	k := viper.Sub("kafka")
	c.KafkaBroker = k.GetStringSlice("brokers")
	c.KafkaRootCA = k.GetString("rootca")
	c.KafkaCert = k.GetString("cert")
	c.KafkaKey = k.GetString("key")
	c.KafkaTopic = k.GetString("topic")
	c.KafkaSSL = k.GetBool("ssl")

}
