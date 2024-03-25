package kafka

import (
	"log"
	"strconv"
	"time"

	"github.com/hamishforbes/kafka-sli-exporter/pkg/metrics"

	"github.com/IBM/sarama"
	"github.com/hamishforbes/kafka-sli-exporter/config"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// KafkaProducer interface
type KafkaProducer interface {
	Start()
	Stop()
	SendMessage(*sarama.ProducerMessage) error
}

// producer struct contains the kafkaProducer client and expose the config
type producer struct {
	Topic           string
	BootstrapServer string
	KafkaConfig     *sarama.Config
	KafkaClient     sarama.SyncProducer
	stopchan        chan bool
}

var producerInstance *producer

// NewProducer return a new Synchronous KafkaProducer for one kafkaCluster. We will set up here the Producer req/s and kafka monitoring topics.
// This producer is the most restricted producer available, because it is synchronous and wait for all ACKs
func NewProducer(config config.KafkaConfig) (Producer KafkaProducer, err error) {

	producerInstance := new(producer)
	producerInstance.Topic = config.Topic
	producerInstance.BootstrapServer = config.BootstrapServer
	producerInstance.KafkaConfig = sarama.NewConfig()
	if config.ProducerConfig.MinVersion {
		producerInstance.KafkaConfig.Version = sarama.MinVersion
	}
	producerInstance.KafkaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	producerInstance.KafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerInstance.KafkaConfig.Producer.Return.Successes = true

	// Logic for configure req/s if MessagesSecond is set up. If not we will send as much as req/s that kafka cluster can handle.
	if config.ProducerConfig.MessagesSecond != 0 {
		flushFrequency := time.Duration(1000/config.ProducerConfig.MessagesSecond) * time.Millisecond
		producerInstance.KafkaConfig.Producer.Flush.Frequency = flushFrequency
		producerInstance.KafkaConfig.Producer.Flush.Messages = 1
		producerInstance.KafkaConfig.Producer.Flush.MaxMessages = 1
	}

	producerInstance.KafkaConfig.Producer.Retry.Max = 100
	producerInstance.KafkaConfig.Producer.Retry.Backoff = time.Duration(250 * time.Millisecond)
	producerInstance.KafkaConfig.Metadata.Retry.Max = 100
	producerInstance.KafkaConfig.Metadata.Retry.Backoff = time.Duration(2 * time.Second)

	// Authenticate against kafka broker using TLS if required
	if config.Tls.Enabled {
		tlsConfig, err := common.NewTLSConfig(config.Tls.ClientCert,
			config.Tls.ClientKey,
			config.Tls.ClusterCert)
		if err != nil {
			log.Fatal(err)
		}
		// This can be used on test server if domain does not match cert:
		tlsConfig.InsecureSkipVerify = true
		producerInstance.KafkaConfig.Net.TLS.Enable = true
		producerInstance.KafkaConfig.Net.TLS.Config = tlsConfig
	}

	producerInstance.KafkaClient, err = sarama.NewSyncProducer([]string{config.BootstrapServer}, producerInstance.KafkaConfig)
	if err != nil {
		logrus.Error("Error creating the kafka sync producer " + err.Error())
	}
	return producerInstance, err
}

// Start function for run the synchronous producer in a loop
func (k *producer) Start() {
	for {
		select {
		default:
			message := &sarama.ProducerMessage{
				Topic:     k.Topic,
				Partition: -1,
				Value:     sarama.StringEncoder("example message"),
			}
			if k.SendMessage(message) != nil {
				metrics.ClusterUp.WithLabelValues(k.BootstrapServer).Set(0)
			} else {
				metrics.ClusterUp.WithLabelValues(k.BootstrapServer).Set(1)
			}
		case <-k.stopchan:
			return
		}
	}
}

func (k *producer) Stop() {
	logrus.Info("Closing go producer......")
	k.stopchan <- true
	close(k.stopchan)

}

// sendMessage and increase prometheus metrics if success/fail
func (k *producer) SendMessage(msg *sarama.ProducerMessage) error {
	timer := prometheus.NewTimer(metrics.MessageSendDuration.WithLabelValues(k.BootstrapServer, k.Topic))
	partition, offset, err := k.KafkaClient.SendMessage(msg)
	duration := timer.ObserveDuration()
	sleep := time.Duration(0)
	if k.KafkaConfig.Producer.Flush.Messages == 1 {
		sleep = k.KafkaConfig.Producer.Flush.Frequency - duration
	}
	if err != nil {
		logrus.Error("Error sending message " + err.Error())
		metrics.ErrorTotalMessageSend.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
	} else {
		metrics.TotalMessageSend.WithLabelValues(k.BootstrapServer, k.Topic).Inc()
		logrus.Debug("Message was saved to partion: "+strconv.Itoa(int(partition))+
			". Message offset is: "+strconv.Itoa(int(offset))+
			". Send duration was: "+duration.String(),
			". Sleeping for: "+sleep.String(),
		)
	}
	time.Sleep(sleep)

	return err
}
