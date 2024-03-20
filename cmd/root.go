package cmd

import (
	"fmt"
	"os"

	"github.com/hamishforbes/kafka-sli-exporter/config"
	"github.com/hamishforbes/kafka-sli-exporter/pkg/metrics"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:     "kafka-sli-exporter",
	Version: version,
	Short:   "kafka-sli-exporter - Generate synthetic Kafka SLI metrics",
	Long:    `kafka-sli-exporter is a CLI to artifically measure kafka cluster latency using prometheus histograms. `,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
var cfgFile string
var userLicense string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "config file (default is config.yaml)")
	cobra.OnInitialize(initCommand)
}

func initCommand() {

	createConfig()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// LOG_LEVEL not set, let's default to info
	lvl := config.Instance.Log.Level
	if len(lvl) == 0 {
		lvl = "info"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	// set global log level
	logrus.SetLevel(ll)
	metrics.InitMetrics(config.Instance.Kafka)
}
func createConfig() {

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logrus.Info("Using config file: " + viper.ConfigFileUsed())
	}

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatal("fatal error config file: " + err.Error())
	}
	config.Instance = new(config.Config)
	err = viper.Unmarshal(&config.Instance)
	if err != nil {
		logrus.Fatal("fatal error unmashal config file: " + err.Error())
	}
}
