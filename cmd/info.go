package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "System and VMware Fusion information",
	Long:  `Display information about the system, VMware Fusion installation, and virtual machines.`,
}

var infoSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Display system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			systemInfo := map[string]interface{}{
				"os":           runtime.GOOS,
				"architecture": runtime.GOARCH,
				"go_version":   runtime.Version(),
				"cpus":         runtime.NumCPU(),
			}
			return outputJSON(systemInfo)
		}

		color.Cyan("System Information:")
		fmt.Printf("Operating System: %s\n", runtime.GOOS)
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
		
		return nil
	},
}

var infoVMwareCmd = &cobra.Command{
	Use:   "vmware",
	Short: "Display VMware Fusion installation information",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.VerifyInstallation(ctx); err != nil {
			color.Red("✗ VMware Fusion installation check failed: %v", err)
			return nil
		}

		color.Green("✓ VMware Fusion installation verified")
		
		// TODO: Add more detailed VMware info like version, paths, etc.
		color.Cyan("VMware Fusion Information:")
		fmt.Println("Status: Installed and operational")
		
		return nil
	},
}

var infoStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display VM statistics and summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		vms, err := vmClient.ListVMs(ctx)
		if err != nil {
			return fmt.Errorf("failed to get VM statistics: %w", err)
		}

		// Calculate statistics
		stats := map[string]interface{}{
			"total_vms": len(vms),
			"running":   0,
			"stopped":   0,
			"suspended": 0,
			"paused":    0,
			"unknown":   0,
		}

		var totalMemory, totalCPUs int
		var totalDiskSize int64

		for _, vm := range vms {
			switch vm.State {
			case "running":
				stats["running"] = stats["running"].(int) + 1
			case "stopped":
				stats["stopped"] = stats["stopped"].(int) + 1
			case "suspended":
				stats["suspended"] = stats["suspended"].(int) + 1
			case "paused":
				stats["paused"] = stats["paused"].(int) + 1
			default:
				stats["unknown"] = stats["unknown"].(int) + 1
			}
			
			totalMemory += vm.Memory
			totalCPUs += vm.CPUs
			totalDiskSize += vm.DiskSize
		}

		stats["total_memory_mb"] = totalMemory
		stats["total_cpus"] = totalCPUs
		stats["total_disk_gb"] = totalDiskSize

		if jsonOutput {
			return outputJSON(stats)
		}

		color.Cyan("VM Statistics:")
		fmt.Printf("Total VMs: %d\n", stats["total_vms"])
		fmt.Printf("  Running: %d\n", stats["running"])
		fmt.Printf("  Stopped: %d\n", stats["stopped"])
		fmt.Printf("  Suspended: %d\n", stats["suspended"])
		fmt.Printf("  Paused: %d\n", stats["paused"])
		fmt.Printf("  Unknown: %d\n", stats["unknown"])
		
		fmt.Printf("\nResource Allocation:\n")
		fmt.Printf("  Total Memory: %d MB\n", totalMemory)
		fmt.Printf("  Total CPUs: %d\n", totalCPUs)
		fmt.Printf("  Total Disk: %d GB\n", totalDiskSize)

		return nil
	},
}

var infoVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		version := map[string]string{
			"version":    "1.0.0",
			"build_date": "2024-01-01",
			"go_version": runtime.Version(),
			"git_commit": "dev",
		}

		if jsonOutput {
			return outputJSON(version)
		}

		color.Cyan("VMware Fusion CLI")
		fmt.Printf("Version: %s\n", version["version"])
		fmt.Printf("Build Date: %s\n", version["build_date"])
		fmt.Printf("Go Version: %s\n", version["go_version"])
		fmt.Printf("Git Commit: %s\n", version["git_commit"])

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	// Add subcommands
	infoCmd.AddCommand(infoSystemCmd, infoVMwareCmd, infoStatsCmd, infoVersionCmd)

	// Global info flags
	infoCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
}