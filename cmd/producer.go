package cmd

import (
	"net/http"

	"github.com/hamishforbes/kafka-sli-exporter/config"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var producerCmd = &cobra.Command{
	Use:     "producer",
	Aliases: []string{"prod"},
	Short:   "producer",
	Run:     startProducer,
}

func init() {
	rootCmd.AddCommand(producerCmd)
}

// startProducer create a new kafka producer and prometheus server with metrics
func startProducer(cmd *cobra.Command, args []string) {

	cfg := config.Instance
	logrus.Info("Starting producers.....")
	for _, cluster := range cfg.Kafka {

		kafkaClusterMonitoring, err := kafka.NewProducer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka producer: " + cluster.BootstrapServer)
		} else {
			go kafkaClusterMonitoring.Start()
		}

	}
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(cfg.Prometheus.PrometheusServer+":"+cfg.Prometheus.PrometheusPort, nil)
	if err != nil {
		logrus.Error("Error creating prometheus server: " + cfg.Prometheus.PrometheusServer + ":" + cfg.Prometheus.PrometheusPort + err.Error())
	}
}
