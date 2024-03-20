package cmd

import (
	"net/http"

	"github.com/hamishforbes/kafka-sli-exporter/config"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var consumerCmd = &cobra.Command{
	Use:     "consumer",
	Aliases: []string{"cons"},
	Short:   "consumer",
	Run:     startConsumer,
}

func init() {
	rootCmd.AddCommand(consumerCmd)
}

// startConsumer create a new consumer and prometheus server with metrics
func startConsumer(cmd *cobra.Command, args []string) {

	cfg := config.Instance
	logrus.Info("Starting consumer.....")
	for _, cluster := range cfg.Kafka {

		kafkaClusterMonitoring, err := kafka.NewConsumer(cluster)
		if err != nil {
			logrus.Error("Error creating kafka consumer: " + cluster.BootstrapServer)
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
