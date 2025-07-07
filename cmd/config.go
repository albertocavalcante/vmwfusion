package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"vmwfusion/internal/types"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management for virtual machines",
	Long:  `Manage configuration settings for virtual machines including memory, CPU, and network settings.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [vm-id]",
	Short: "Get configuration for a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		config, err := vmClient.GetVMConfig(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get VM config: %w", err)
		}

		if jsonOutput {
			return outputJSON(config)
		}

		color.Cyan("VM Configuration:")
		fmt.Printf("Name: %s\n", config.Name)
		fmt.Printf("OS: %s\n", config.OS)
		fmt.Printf("Memory: %d MB\n", config.Memory)
		fmt.Printf("CPUs: %d\n", config.CPUs)
		fmt.Printf("Disk Size: %d GB\n", config.DiskSize)
		fmt.Printf("ISO Path: %s\n", config.ISOPath)
		fmt.Printf("Start on Boot: %t\n", config.StartOnBoot)

		if len(config.Network) > 0 {
			color.Cyan("\nNetwork Configuration:")
			for i, net := range config.Network {
				fmt.Printf("  Interface %d:\n", i+1)
				fmt.Printf("    Type: %s\n", net.Type)
				fmt.Printf("    Name: %s\n", net.Name)
				fmt.Printf("    MAC: %s\n", net.MacAddress)
				fmt.Printf("    Connected: %t\n", net.Connected)
			}
		}

		if len(config.Metadata) > 0 {
			color.Cyan("\nMetadata:")
			for key, value := range config.Metadata {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [vm-id]",
	Short: "Set configuration for a virtual machine",
	Long:  `Update configuration settings for a virtual machine. Use flags to specify which settings to update.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Get current config
		config, err := vmClient.GetVMConfig(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get current VM config: %w", err)
		}

		// Update with provided flags
		if cmd.Flags().Changed("memory") {
			config.Memory = vmMemory
		}
		if cmd.Flags().Changed("cpus") {
			config.CPUs = vmCPUs
		}
		if cmd.Flags().Changed("os") {
			config.OS = vmOS
		}
		if cmd.Flags().Changed("disk") {
			config.DiskSize = vmDiskSize
		}

		if err := vmClient.UpdateVMConfig(ctx, args[0], config); err != nil {
			return fmt.Errorf("failed to update VM config: %w", err)
		}

		color.Green("Successfully updated configuration for VM: %s", args[0])
		return nil
	},
}

var configMemoryCmd = &cobra.Command{
	Use:   "memory [vm-id] [memory-mb]",
	Short: "Set memory configuration for a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		memory, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid memory value: %s", args[1])
		}

		config, err := vmClient.GetVMConfig(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get VM config: %w", err)
		}

		config.Memory = memory

		if err := vmClient.UpdateVMConfig(ctx, args[0], config); err != nil {
			return fmt.Errorf("failed to update memory: %w", err)
		}

		color.Green("Successfully updated memory to %d MB for VM: %s", memory, args[0])
		return nil
	},
}

var configCPUCmd = &cobra.Command{
	Use:   "cpu [vm-id] [cpu-count]",
	Short: "Set CPU configuration for a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		cpus, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid CPU count: %s", args[1])
		}

		config, err := vmClient.GetVMConfig(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get VM config: %w", err)
		}

		config.CPUs = cpus

		if err := vmClient.UpdateVMConfig(ctx, args[0], config); err != nil {
			return fmt.Errorf("failed to update CPU count: %w", err)
		}

		color.Green("Successfully updated CPU count to %d for VM: %s", cpus, args[0])
		return nil
	},
}

var configNetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network configuration management",
	Long:  `Manage network configuration for virtual machines including adding, removing, and modifying network interfaces.`,
}

var configNetworkAddCmd = &cobra.Command{
	Use:   "add [vm-id] [network-type]",
	Short: "Add a network interface to a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		config, err := vmClient.GetVMConfig(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get VM config: %w", err)
		}

		// Add new network interface
		newNetwork := types.NetworkConfig{
			Type:      args[1],
			Name:      fmt.Sprintf("Network Interface %d", len(config.Network)+1),
			Connected: true,
		}

		config.Network = append(config.Network, newNetwork)

		if err := vmClient.UpdateVMConfig(ctx, args[0], config); err != nil {
			return fmt.Errorf("failed to add network interface: %w", err)
		}

		color.Green("Successfully added %s network interface to VM: %s", args[1], args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configGetCmd, configSetCmd, configMemoryCmd, configCPUCmd, configNetworkCmd)
	configNetworkCmd.AddCommand(configNetworkAddCmd)

	// Set command flags
	configSetCmd.Flags().IntVarP(&vmMemory, "memory", "m", 0, "memory in MB")
	configSetCmd.Flags().IntVarP(&vmCPUs, "cpus", "c", 0, "number of CPUs")
	configSetCmd.Flags().StringVarP(&vmOS, "os", "o", "", "operating system type")
	configSetCmd.Flags().Int64VarP(&vmDiskSize, "disk", "d", 0, "disk size in GB")

	// Global config flags
	configCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
}