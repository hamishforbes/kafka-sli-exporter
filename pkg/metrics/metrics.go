package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/service-level-indicator-exporter-for-kafka/config"
)

const Namespace = "kafka"
const Subsystem = "sli"

// TotalMessageSend Producer instance will increase counter with total messages send per kafka cluster
var TotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "message_sent_count",
		Help:      "Number of messages sucessfully sent to kafka",
	},
	[]string{"cluster", "topic"},
)

// ErrorTotalMessageSend Producer instance will increase counter if we are not able of send a message per kafka cluster
var ErrorTotalMessageSend = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "message_send_error_count",
		Help:      "Number of messages which failed to send to kafka",
	},
	[]string{"cluster", "topic"},
)

// ClusterUp Producer will set up Gauge values to 0 if cluster is unreacheable or 1 if we are able to connect to kafka cluster
var ClusterUp = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "cluster_up",
		Help:      "Kafka clusters with errors",
	},
	[]string{"cluster"},
)

// MessageSendDuration Producer summary with rate duration/reqs send
var MessageSendDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace:                       Namespace,
		Subsystem:                       Subsystem,
		Name:                            "message_send_duration_seconds",
		Help:                            "Message send latency",
		Buckets:                         prometheus.ExponentialBucketsRange(0.001, 2, 20),
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	},
	[]string{"cluster", "topic"},
)

// TotalMessageRead Consumer instance will increase counter with total messages read per kafka cluster
var TotalMessageRead = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "message_read_count",
		Help:      "Number of messages read by Kafka consumer",
	},
	[]string{"cluster", "topic"},
)

// TotalMessageRead Consumer instance will increase counter if it is unable of read from kafka cluster
var ErrorInRead = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "message_read_error_count",
		Help:      "Number of message read errors",
	},
	[]string{"cluster", "topic"},
)

// InitMetrics function call when app start for register and init the metrics
func InitMetrics(cfg []config.KafkaConfig) {
	for _, cluster := range cfg {
		prometheus.Register(TotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(TotalMessageRead.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(ClusterUp.WithLabelValues(cluster.BootstrapServer))
		prometheus.Register(ErrorTotalMessageSend.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(ErrorInRead.WithLabelValues(cluster.BootstrapServer, cluster.Topic))
		prometheus.Register(MessageSendDuration)
	}
}

func ResetMetrics(cfg []config.KafkaConfig) {
	for _, cluster := range cfg {
		TotalMessageSend.DeleteLabelValues(cluster.BootstrapServer, cluster.Topic)
		TotalMessageRead.DeleteLabelValues(cluster.BootstrapServer, cluster.Topic)
		ClusterUp.DeleteLabelValues(cluster.BootstrapServer)
		ErrorTotalMessageSend.DeleteLabelValues(cluster.BootstrapServer, cluster.Topic)
		ErrorInRead.DeleteLabelValues(cluster.BootstrapServer, cluster.Topic)
	}
}
