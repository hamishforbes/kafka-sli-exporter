package tests

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	msConfig "github.com/hamishforbes/kafka-sli-exporter/config"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/kafka"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

type testConfig struct {
	kafkaConfig                   msConfig.KafkaConfig
	expectErr                     bool
	expectedTotalMessageSend      float64
	expectedErrorTotalMessageSend float64
	expectedTotalMessageRead      float64
	expectedClusterUp             float64
}

func TestE2E(t *testing.T) {
	configs := []testConfig{
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9093",
				Topic:           "monitoring-topic",
			},
			expectErr:                     true,
			expectedTotalMessageSend:      0,
			expectedTotalMessageRead:      0,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             0,
		},
		{
			kafkaConfig: msConfig.KafkaConfig{
				BootstrapServer: "localhost:9092",
				Topic:           "monitoring-topic",
				ConsumerConfig: msConfig.ConsumerConfig{
					FromBeginning: true,
				},
			},
			expectErr:                     false,
			expectedTotalMessageSend:      1,
			expectedTotalMessageRead:      1,
			expectedErrorTotalMessageSend: 0,
			expectedClusterUp:             1,
		},
	}
	startEnviron(t)
	for _, config := range configs {
		testProducer(t, config)
		testConsumer(t, config)
	}
}

func startEnviron(t *testing.T) {
	compose, err := tc.NewDockerCompose("../../compose.yaml")
	require.NoError(t, err, "NewDockerComposeAPI()")
	t.Cleanup(func() {
		require.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	compose.Up(ctx, tc.Wait(true))
}

func testProducer(t *testing.T, config testConfig) {
	assert := assert.New(t)
	producer, err := kafka.NewProducer(config.kafkaConfig)
	if config.expectErr == true {
		assert.Error(err)
	} else {
		metrics.InitMetrics([]msConfig.KafkaConfig{config.kafkaConfig})
		message := &sarama.ProducerMessage{
			Topic:     config.kafkaConfig.Topic,
			Partition: -1,
			Value:     sarama.StringEncoder("example message"),
		}
		producer.SendMessage(message)
		assert.Equal(config.expectedTotalMessageSend, testutil.ToFloat64(metrics.TotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
		assert.Equal(config.expectedErrorTotalMessageSend, testutil.ToFloat64(metrics.ErrorTotalMessageSend.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
	}
}

func testConsumer(t *testing.T, config testConfig) {
	assert := assert.New(t)

	consumer, err := kafka.NewConsumer(config.kafkaConfig)
	if config.expectErr == true {
		assert.Error(err)
	} else {
		go consumer.Start()
		time.Sleep(10 * time.Second)
		consumer.Stop()
		assert.Equal(config.expectedTotalMessageRead, testutil.ToFloat64(metrics.TotalMessageRead.WithLabelValues(config.kafkaConfig.BootstrapServer, config.kafkaConfig.Topic)))
	}

}
