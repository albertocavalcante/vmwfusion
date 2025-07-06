package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// configCommands returns the configuration management command group
func configCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage vmwfusion configuration settings",
	}

	// Show config command
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Println("Current configuration:")
			fmt.Printf("  Config file: %s\n", configPath)
			fmt.Printf("  Local VM dir: %s\n", globalConfig.LocalVMDir)
			fmt.Printf("  VMware path: %s\n", globalConfig.VMwarePath)
			fmt.Printf("  ExFAT mount base: %s\n", globalConfig.ExFATMountBase)
			fmt.Printf("  Default archive drive: %s\n", globalConfig.DefaultDrive)
			fmt.Printf("  Shutdown timeout: %ds\n", globalConfig.ShutdownTimeout)
			fmt.Printf("  Cleanup original: %t\n", globalConfig.CleanupOriginal)
			fmt.Printf("  Auto optimize: %t\n", globalConfig.AutoOptimize)
			fmt.Printf("  Verify checksums: %t\n", globalConfig.VerifyChecksums)
			fmt.Printf("  Preferred clone type: %s\n", globalConfig.PreferredCloneType)
			fmt.Printf("  Rsync options: %v\n", globalConfig.RsyncOptions)
			
			fmt.Println("\nVMware Tools:")
			fmt.Printf("  Base path: %s\n", vmwareTools.BasePath)
			fmt.Printf("  vmrun path: %s\n", vmwareTools.VMRunPath)
			fmt.Printf("  vmcli path: %s\n", vmwareTools.VMCliPath)
			
			// Check if vmcli exists
			if FileExists(vmwareTools.VMCliPath) {
				fmt.Printf("  vmcli available: Yes\n")
			} else {
				fmt.Printf("  vmcli available: No (some advanced features unavailable)\n")
			}
			
			// Check Pro features
			if err := vmwareTools.CheckProFeatures(); err != nil {
				fmt.Printf("  Pro features: No (%v)\n", err)
			} else {
				fmt.Printf("  Pro features: Yes\n")
			}
			
			return nil
		},
	}

	// Create config command
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create sample configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf("Configuration file already exists: %s\n", configPath)
				fmt.Print("Overwrite? (y/N): ")
				
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					GlobalLogger.Warning.Printf("Failed to read input: %v\n", err)
					return nil
				}
				
				if response != "y" && response != "Y" {
					GlobalLogger.Info.Println("Configuration creation cancelled")
					return nil
				}
			}
			
			// Create default config
			config := DefaultConfig()
			
			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			
			if err := config.SaveConfig(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			
			GlobalLogger.Success.Printf("Sample configuration created at: %s\n", configPath)
			GlobalLogger.Info.Println("Edit the file to customize settings for your environment")
			
			return nil
		},
	}

	// Edit config command (opens in default editor)
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				GlobalLogger.Warning.Printf("Configuration file does not exist: %s\n", configPath)
				fmt.Print("Create it now? (y/N): ")
				
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					GlobalLogger.Warning.Printf("Failed to read input: %v\n", err)
					return fmt.Errorf("failed to read user input")
				}
				
				if response == "y" || response == "Y" {
					config := DefaultConfig()
					if err := config.SaveConfig(configPath); err != nil {
						return fmt.Errorf("failed to create config file: %w", err)
					}
					GlobalLogger.Success.Printf("Configuration file created: %s\n", configPath)
				} else {
					return fmt.Errorf("configuration file does not exist")
				}
			}
			
			// Get editor from environment or use default
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "nano" // Default editor
			}
			
			GlobalLogger.Info.Printf("Opening configuration file with %s...\n", editor)
			
			// Note: In a real implementation, we'd use exec.Command to open the editor
			// For now, just show the path
			fmt.Printf("Please edit the configuration file at: %s\n", configPath)
			
			return nil
		},
	}

	// Validate config command
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Printf("Validating configuration: %s\n", configPath)
			
			// Try to load config
			config, err := LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			
			// Validate paths
			var errors []string
			
			if _, err := os.Stat(config.LocalVMDir); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("Local VM directory does not exist: %s", config.LocalVMDir))
			}
			
			if _, err := os.Stat(config.VMwarePath); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("VMware path does not exist: %s", config.VMwarePath))
			}
			
			if _, err := os.Stat(config.ExFATMountBase); os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("ExFAT mount base does not exist: %s", config.ExFATMountBase))
			}
			
			// Validate VMware tools
			tools, err := DetectVMwareTools(config.VMwarePath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("VMware tools validation failed: %v", err))
			} else {
				if _, err := os.Stat(tools.VMRunPath); os.IsNotExist(err) {
					errors = append(errors, fmt.Sprintf("vmrun not found: %s", tools.VMRunPath))
				}
			}
			
			// Validate numeric values
			if config.ShutdownTimeout < 0 || config.ShutdownTimeout > 600 {
				errors = append(errors, fmt.Sprintf("Invalid shutdown timeout: %d (should be 0-600)", config.ShutdownTimeout))
			}
			
			// Validate clone type
			validCloneTypes := []string{"full", "linked"}
			validType := false
			for _, t := range validCloneTypes {
				if config.PreferredCloneType == t {
					validType = true
					break
				}
			}
			if !validType {
				errors = append(errors, fmt.Sprintf("Invalid preferred clone type: %s (should be 'full' or 'linked')", config.PreferredCloneType))
			}
			
			// Show results
			if len(errors) > 0 {
				GlobalLogger.Error.Println("Configuration validation failed:")
				for _, err := range errors {
					fmt.Printf("  • %s\n", err)
				}
				return fmt.Errorf("configuration has %d error(s)", len(errors))
			}
			
			GlobalLogger.Success.Println("Configuration validation passed")
			return nil
		},
	}

	// Reset config command
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Warning.Printf("This will reset configuration to defaults: %s\n", configPath)
			fmt.Print("Are you sure? (y/N): ")
			
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				GlobalLogger.Warning.Printf("Failed to read input: %v\n", err)
				return nil
			}
			
			if response != "y" && response != "Y" {
				GlobalLogger.Info.Println("Configuration reset cancelled")
				return nil
			}
			
			config := DefaultConfig()
			if err := config.SaveConfig(configPath); err != nil {
				return fmt.Errorf("failed to reset config file: %w", err)
			}
			
			GlobalLogger.Success.Printf("Configuration reset to defaults: %s\n", configPath)
			return nil
		},
	}

	cmd.AddCommand(showCmd)
	cmd.AddCommand(createCmd)
	cmd.AddCommand(editCmd)
	cmd.AddCommand(validateCmd)
	cmd.AddCommand(resetCmd)

	return cmd
}