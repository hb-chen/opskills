package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/hb-chen/opskills/internal/config"
	"github.com/hb-chen/opskills/pkg/logger"
)

var (
	cfgFile, logLevel, logPath string
	stderr, debug              bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opskills-agent",
	Short: "Intelligent Ops Agent with LLM and Multi-Agent Collaboration",
	Long: `An intelligent Ops tool using Golang, langgraphgo, and langchaingo.
Supports LLM-driven planning, multi-agent collaboration, and skill capabilities.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger first
		if err := initLogger(logPath, logLevel, debug, stderr); err != nil {
			return err
		}

		// Initialize config
		initConfig()

		logger.Info("Starting opskills-agent...")
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// This will be implemented to start the server
		return fmt.Errorf("server start not yet implemented - use 'serve' command")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./configs/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&stderr, "stderr", "e", false, "log to stderr")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "log level: DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
	rootCmd.PersistentFlags().StringVar(&logPath, "log-path", "./log", "log file path")

	// Bind flags to viper
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log.path", rootCmd.PersistentFlags().Lookup("log-path"))
	_ = viper.BindPFlag("log.debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config.Init()

	if cfgFile != "" {
		// Use config file from the flag.
		config.Viper().SetConfigFile(cfgFile)
	} else {
		// Search config in current directory and configs directory
		config.Viper().AddConfigPath(".")
		config.Viper().AddConfigPath("./configs")
		config.Viper().SetConfigType("yaml")
		config.Viper().SetConfigName("config")
	}

	config.Viper().AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := config.Viper().ReadInConfig(); err != nil {
		logger.Warnf("Config file not found: %v", err)
	} else {
		logger.Infof("Using config file: %s", config.Viper().ConfigFileUsed())
	}
}

const logCallerSkip = 1

func initLogger(path, level string, debug, e bool) error {
	writer := getLogWriter(path)
	if e {
		stderrWriter, _, err := zap.Open("stderr")
		if err != nil {
			return err
		}
		writer = stderrWriter
	}

	// Parse log level
	logLevel := zapcore.InfoLevel
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	// Create encoder
	encoder := getLogEncoder(debug, e)

	// Create core
	core := zapcore.NewCore(encoder, writer, logLevel)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(logCallerSkip))

	// Replace global logger
	logger.ReplaceLogger(zapLogger)

	return nil
}

func getLogEncoder(debug, e bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if debug && e {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(path string) zapcore.WriteSyncer {
	path = strings.TrimRight(path, "/")
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path + "/opskills-agent.log",
		MaxSize:    10,   // megabytes
		MaxBackups: 10,   // number of backups
		MaxAge:     30,   // days
		Compress:   true, // compress old files
	}
	return zapcore.AddSync(lumberJackLogger)
}
