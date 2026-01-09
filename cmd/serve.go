package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hb-chen/opskills/internal/agent"
	"github.com/hb-chen/opskills/internal/config"
	"github.com/hb-chen/opskills/internal/llm"
	"github.com/hb-chen/opskills/internal/server"
	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/skill/direct"
	"github.com/hb-chen/opskills/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addrHTTP, addrGrpc string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Ops Agent server",
	Long:  `Start the HTTP and gRPC servers for the Ops Agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override server addresses from flags if provided
		if addrHTTP != "" {
			cfg.Server.HTTP.Addr = addrHTTP
		}
		if addrGrpc != "" {
			cfg.Server.GRPC.Addr = addrGrpc
		}

		// Create context
		ctx, cancel := context.WithCancel(cmd.Context())

		// Setup signal handling
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		// Initialize Pipeline
		pipeline, err := initPipeline(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize pipeline: %w", err)
		}

		// Start servers (gRPC and HTTP with Web UI)
		go func() {
			if err := server.Serve(ctx, cfg, pipeline); err != nil {
				logger.Errorf("Server error: %v", err)
				cancel()
			}
		}()

		// Wait for interrupt signal
		sig := <-quit
		logger.Infof("Received signal %s, shutting down...", sig.String())
		cancel()

		return nil
	},
}

func init() {
	// Add flags to serve command
	serveCmd.Flags().StringVar(&addrHTTP, "addr-http", "", "HTTP server address (overrides config file)")
	serveCmd.Flags().StringVar(&addrGrpc, "addr-grpc", "", "gRPC server address (overrides config file)")

	// Bind flags to viper
	_ = viper.BindPFlag("server.http.addr", serveCmd.Flags().Lookup("addr-http"))
	_ = viper.BindPFlag("server.grpc.addr", serveCmd.Flags().Lookup("addr-grpc"))

	rootCmd.AddCommand(serveCmd)
}

// initPipeline initializes the agent pipeline
func initPipeline(cfg *config.Config) (*agent.Pipeline, error) {
	// Load skills
	skillsDir := cfg.Skills.Dir
	if skillsDir == "" {
		skillsDir = "./skills"
	}

	loader := skill.NewLoader(skillsDir)
	skills, err := loader.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	// Create registry
	registry := skill.NewRegistry()
	for _, s := range skills {
		registry.Register(s)
	}

	// Create LLM client
	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("LLM API key not configured")
	}

	llmClient, err := llm.NewClient(cfg.LLM.Provider, apiKey, cfg.LLM.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create agents
	planner := agent.NewPlanningAgent(llmClient, registry)
	executor := direct.NewDirectExecutor(30 * time.Minute) // 30 minutes timeout
	executorAgent := agent.NewExecutorAgent(executor, registry)

	// Create pipeline
	pipeline := agent.NewPipeline(planner, executorAgent)

	return pipeline, nil
}
