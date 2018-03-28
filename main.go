package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"

	"github.com/davecgh/go-spew/spew"
)

// KafkaConfig Kafka configuration struct
type KafkaConfig struct {
	HostArray         []string
	Topic             string
	SecureKafkaEnable bool
}

// KafkaHandle Kafka Client
type KafkaHandle struct {
	config        *KafkaConfig
	kafkaClient   sarama.Client
	kafkaConsumer sarama.Consumer
	kafkaConfig   *sarama.Config
	controlChan   chan bool
}

// NewKafkaHandle method to initialize the ClientHandle with sane defaults
func NewKafkaHandle(config *KafkaConfig) *KafkaHandle {
	return &KafkaHandle{config: config}
}

//EnableSecureKafka Enable Secure Kafka config
func (k *KafkaHandle) EnableSecureKafka(ClientCertificateFile string, ClientPrivateKeyFile string, KafkaCAFile string) error {
	k.kafkaConfig.Net.TLS.Enable = true
	k.kafkaConfig.Net.TLS.Config = &tls.Config{}
	// TLS version needs to be 1.2
	k.kafkaConfig.Net.TLS.Config.MinVersion = tls.VersionTLS12
	k.kafkaConfig.Net.TLS.Config.MaxVersion = tls.VersionTLS12
	k.kafkaConfig.Net.TLS.Config.PreferServerCipherSuites = true
	k.kafkaConfig.Net.TLS.Config.InsecureSkipVerify = true

	// Handle client cert and also client key
	//TODO add support for password protected key?
	if ClientCertificateFile != "" && ClientPrivateKeyFile != "" {
		glog.V(1).Infof("Client Cert file: %s, Client Cert Key: %s\n", ClientCertificateFile, ClientPrivateKeyFile)
		cert, err := tls.LoadX509KeyPair(ClientCertificateFile, ClientPrivateKeyFile)

		if err != nil {
			glog.Warningf("Error loading certificats: %s", err)
			return err
		}
		glog.V(1).Infoln("Added root Cert in tls config")
		k.kafkaConfig.Net.TLS.Config.Certificates = []tls.Certificate{cert}
	}
	tlsCertPool := x509.NewCertPool()

	// CA file
	if KafkaCAFile != "" {
		glog.V(1).Infof("Kafka CA: %s\n", KafkaCAFile)
		kafkaCaCertFile, err := ioutil.ReadFile(KafkaCAFile)
		if err != nil {
			glog.Warningf("Failed to read kafka Certificate Authority file %s\n", err)
			return err
		}
		if !tlsCertPool.AppendCertsFromPEM(kafkaCaCertFile) {
			glog.Warningln("Failed to append certificates from Kafka Certificate Authority file")
			return nil
		}
		glog.V(1).Infoln("Added root Cert in tls config")
	}
	k.kafkaConfig.Net.TLS.Config.RootCAs = tlsCertPool
	return nil
}

// Initialize the Kafka client, connecting to Kafka and running some sanity checks
func (k *KafkaHandle) Initialize(ClientCertificateFile string, ClientPrivateKeyFile string, KafkaCAFile string, brokersIP []string) error {
	var err error

	if len(brokersIP) == 0 {
		glog.Warningf("Invalid Broker IP")
		return fmt.Errorf("Invalid Broker IP")
	}
	k.kafkaConfig = sarama.NewConfig()

	k.kafkaConfig.ClientID = "quarantine-client"
	if k.config.SecureKafkaEnable {
		err = k.EnableSecureKafka(ClientCertificateFile, ClientPrivateKeyFile, KafkaCAFile)
		if err != nil {
			glog.Warningf("Failed to enable Secure Kafka err %s\n", err.Error())
			return fmt.Errorf("Failed to enable secure kafka err %s", err.Error())
		}
	}

	// Create a new Kafka client - failure to connect to Kafka is a fatal error.
	k.kafkaClient, err = sarama.NewClient(brokersIP, k.kafkaConfig)
	if err != nil {
		glog.Warningf("Failed to connect to kafka (%s)", err)
		return fmt.Errorf("Failed to connect to kafka (%s)", err)
	}

	k.kafkaConsumer, err = sarama.NewConsumerFromClient(k.kafkaClient)
	if err != nil {
		glog.Errorf("Failed to start consumer (%s)", err)
	}
	defer k.kafkaConsumer.Close()

	return nil
}

// consumerLoop get messages of Kafka
func consumerLoop(cons sarama.Consumer, topic string, part int32, h *Tetration, t *Twilio) {
	glog.V(1).Infof("Consuming Topic %s Partition %d \n", topic, part)
	partitionConsumer, err := cons.ConsumePartition(topic, part, sarama.OffsetOldest)
	if err != nil {
		glog.Fatal(err.Error())
	}

	for {
		select {
		// Each time a message is send to the kafka bus this is called
		case msg := <-partitionConsumer.Messages():
			alert := Alert{}

			// Get our alert in a struct
			json.Unmarshal(msg.Value, &alert)

			// In this example we're interested in the SIDE_CHANNEL attacks only
			if strings.Contains(alert.KeyID, "SIDE_CHANNEL") {
				// Dump the alert to show progress
				spew.Dump(alert)

				// If Tetration is defined, we shall Isolate
				if h != nil {
					glog.V(2).Infof("Tetration is enabled, isolation started\n")
					// Alert will return a sensor name and not IP, we need to get the IPs from inventory
					// a host could have multiple IPs
					ips := h.getSensorIP(alert.AlertDetail.SensorID)
					for _, v := range ips {
						// Isolate VMs by assigning tags to IP (quarantine=isolate)
						h.Isolate(v["ip"], v["scope"])
					}
				} else {
					glog.V(2).Infof("Tetration is not enabled, no isolation\n")
				}

				// If Twilio is defined, we shall SMS
				if t != nil {
					glog.V(2).Infof("Twilio is enabled, SMS\n")
					//t.SendSMS(t.To, alert.AlertText)
					t.Message <- &Message{To: t.To, Text: alert.AlertText}
				} else {
					glog.V(2).Infof("Twilio is not enabled, no SMS\n")
				}

				glog.V(1).Infof("Consumed message offset %d on partition %d\n", msg.Offset, part)
			}
		}
	}
}

func main() {
	flag.Parse()
	glog.Infoln("Starting go-home...")
	config := NewConfig()

	kafkaConfig := new(KafkaConfig)
	// Topic is provided in the topics.txt file
	kafkaConfig.Topic = config.KafkaTopic
	kafkaConfig.SecureKafkaEnable = config.KafkaSSL
	kafkaHandle := NewKafkaHandle(kafkaConfig)

	var tetration *Tetration
	if config.APIEnabled {
		tetration = NewTetration(config)
	}

	var twilio *Twilio
	if config.TwilioEnabled {
		twilio = NewTwilio(config.TwilioToken, config.TwilioSID, config.TwilioFrom, config.TwilioLimit)
		// We set the To based on the configuration file, but we could also input it directly here (or get it from somewhere)
		twilio.To = config.TwilioTo
	}

	// KafkaCert / KafkaKey / KafkaRootCA and KafkaBroker are provided by Tetration
	// for more info, see the README
	glog.V(1).Infof("Initializing Kafka...")
	err := kafkaHandle.Initialize(config.KafkaCert, config.KafkaKey, config.KafkaRootCA, config.KafkaBroker)
	if err != nil {
		glog.Errorf("Kafka Initialization failed: %s", err.Error())
		return
	}

	glog.V(1).Infof("Getting Partitions...")
	partitions, err := kafkaHandle.kafkaConsumer.Partitions(kafkaConfig.Topic)
	if err != nil {
		glog.Errorf("Getting Kafka partition failed: %s", err.Error())
		return
	}
	glog.V(2).Infof("Topic %s has %d partitions\n", kafkaConfig.Topic, len(partitions))

	//TODO do I really need to use offset? Seems like queue is not draining...
	//var startOffset, endOffset int64

	// Message can arrive on any partition so we need to consume all partitions
	for _, part := range partitions {
		cons, err := sarama.NewConsumerFromClient(kafkaHandle.kafkaClient)
		if err != nil {
			panic(err)
		}
		go consumerLoop(cons, kafkaConfig.Topic, part, tetration, twilio)
	}

	// This is for a demo, so keep the program running until some hits enter
	// Note that we use go routine above, if you kill the program no more messages
	fmt.Scanln()
	fmt.Println("done")
}
