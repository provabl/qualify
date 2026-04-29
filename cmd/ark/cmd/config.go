package cmd

import (
	"fmt"

	"github.com/provabl/qualify/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Ark configuration",
	Long:  `View and modify Ark configuration settings stored in ~/.ark/config.yml`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value by key path.

Examples:
  ark config get current_profile
  ark config get agent.port
  ark config get backend.url`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// Load config
		path, err := config.GetConfigPath()
		if err != nil {
			ExitWithError(fmt.Errorf("get config path: %w", err))
		}

		cfg, err := config.Load(path)
		if err != nil {
			ExitWithError(fmt.Errorf("load config: %w", err))
		}

		// Get value
		value, err := cfg.Get(key)
		if err != nil {
			ExitWithError(err)
		}

		fmt.Println(value)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value by key path.

Examples:
  ark config set current_profile research
  ark config set agent.port 8737
  ark config set backend.url https://ark.example.edu`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Load config
		path, err := config.GetConfigPath()
		if err != nil {
			ExitWithError(fmt.Errorf("get config path: %w", err))
		}

		cfg, err := config.Load(path)
		if err != nil {
			ExitWithError(fmt.Errorf("load config: %w", err))
		}

		// Set value
		if err := cfg.Set(key, value); err != nil {
			ExitWithError(err)
		}

		// Save config
		if err := cfg.Save(path); err != nil {
			ExitWithError(fmt.Errorf("save config: %w", err))
		}

		fmt.Printf("✓ Set %s = %s\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Long:  `Display all configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		path, err := config.GetConfigPath()
		if err != nil {
			ExitWithError(fmt.Errorf("get config path: %w", err))
		}

		cfg, err := config.Load(path)
		if err != nil {
			ExitWithError(fmt.Errorf("load config: %w", err))
		}

		// Display config
		fmt.Printf("Configuration file: %s\n\n", path)

		// Marshal to YAML for display
		data, err := yaml.Marshal(cfg)
		if err != nil {
			ExitWithError(fmt.Errorf("marshal config: %w", err))
		}

		fmt.Print(string(data))
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration with defaults",
	Long:  `Create a new configuration file with default values.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.GetConfigPath()
		if err != nil {
			ExitWithError(fmt.Errorf("get config path: %w", err))
		}

		// Create default config
		cfg := config.DefaultConfig()

		// Save config
		if err := cfg.Save(path); err != nil {
			ExitWithError(fmt.Errorf("save config: %w", err))
		}

		fmt.Printf("✓ Configuration initialized at %s\n", path)
	},
}
