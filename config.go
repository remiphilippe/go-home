package main

import (
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// Config go-hone configuration structure
type Config struct {
	APIEnabled  bool
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

	TwilioEnabled bool
	TwilioSID     string
	TwilioToken   string
	TwilioFrom    string
	TwilioTo      string
	TwilioLimit   int
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
		glog.Fatalf("Fatal error config file: %s", err.Error())
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
	c.APIEnabled = o.GetBool("enabled")

	glog.V(2).Infof("Loading Kafka Config...")
	k := viper.Sub("kafka")
	c.KafkaBroker = k.GetStringSlice("brokers")
	c.KafkaRootCA = k.GetString("rootca")
	c.KafkaCert = k.GetString("cert")
	c.KafkaKey = k.GetString("key")
	c.KafkaTopic = k.GetString("topic")
	c.KafkaSSL = k.GetBool("ssl")

	glog.V(2).Infof("Loading Twilio Config...")
	t := viper.Sub("twilio")
	c.TwilioEnabled = t.GetBool("enabled")
	c.TwilioToken = t.GetString("token")
	c.TwilioSID = t.GetString("sid")
	c.TwilioFrom = t.GetString("from")
	c.TwilioTo = t.GetString("to")
	c.TwilioLimit = t.GetInt("limit")
}
