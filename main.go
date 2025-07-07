package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	configPath     string
	verbose        bool
	dryRun         bool
	
	// Global config
	globalConfig *Config
	vmwareTools  *VMwareTools
)

func main() {
	// Initialize logger
	InitializeLogger()
	
	var rootCmd = &cobra.Command{
		Use:   "vmwfusion",
		Short: "VMware Fusion CLI - Comprehensive VM and ISO management",
		Long: `A unified CLI tool for VMware Fusion that provides:
• ISO file discovery, export, and management with hash-based deduplication
• VM lifecycle management (power, cloning, optimization)  
• VM migration between local SSD and ExFAT archive storage
• Fast deployment with linked clones
• Comprehensive configuration management`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			config, err := LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			globalConfig = config
			
			// Detect VMware tools
			tools, err := DetectVMwareTools(config.VMwarePath)
			if err != nil {
				return fmt.Errorf("VMware Fusion detection failed: %w", err)
			}
			vmwareTools = tools
			
			if verbose {
				GlobalLogger.Info.Printf("Config loaded from: %s\n", configPath)
				GlobalLogger.Info.Printf("VMware tools detected at: %s\n", tools.BasePath)
			}
			
			return nil
		},
	}

	// Global flags
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".vmwfusion-config.json")
	
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath, "Configuration file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")

	// Add command groups
	rootCmd.AddCommand(isoCommands())
	rootCmd.AddCommand(vmCommands())
	rootCmd.AddCommand(migrateCommands())
	rootCmd.AddCommand(configCommands())
	rootCmd.AddCommand(infoCommands())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

