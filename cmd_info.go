package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// infoCommands returns the system information command group  
func infoCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "System and VMware information",
		Long:  "Display system information, VMware status, and drive information",
	}

	// System info command
	systemCmd := &cobra.Command{
		Use:   "system",
		Short: "Show system information",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Println("System Information:")
			
			// macOS version
			if output, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
				fmt.Printf("  macOS Version: %s\n", strings.TrimSpace(string(output)))
			} else {
				GlobalLogger.Warning.Printf("Could not get macOS version: %v\n", err)
			}
			
			// Hardware info
			if output, err := exec.Command("sysctl", "-n", "hw.model").Output(); err == nil {
				fmt.Printf("  Hardware Model: %s\n", strings.TrimSpace(string(output)))
			} else {
				GlobalLogger.Warning.Printf("Could not get hardware model: %v\n", err)
			}
			
			if output, err := exec.Command("sysctl", "-n", "hw.ncpu").Output(); err == nil {
				fmt.Printf("  CPU Cores: %s\n", strings.TrimSpace(string(output)))
			} else {
				GlobalLogger.Warning.Printf("Could not get CPU cores: %v\n", err)
			}
			
			if output, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
				memBytes := strings.TrimSpace(string(output))
				if len(memBytes) > 0 {
					if bytes, err := strconv.ParseInt(memBytes, 10, 64); err == nil {
						memGB := float64(bytes) / (1024 * 1024 * 1024)
						fmt.Printf("  Memory: %.1f GB\n", memGB)
					} else {
						fmt.Printf("  Memory: %s bytes\n", memBytes)
					}
				}
			} else {
				GlobalLogger.Warning.Printf("Could not get memory size: %v\n", err)
			}
			
			// Disk space on main drive
			fmt.Println("\nDisk Usage:")
			if output, err := exec.Command("df", "-h", "/").Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				if len(lines) > 1 {
					fmt.Printf("  Root: %s\n", lines[1])
				}
			} else {
				GlobalLogger.Warning.Printf("Could not get disk usage: %v\n", err)
			}
			
			return nil
		},
	}

	// VMware info command
	vmwareCmd := &cobra.Command{
		Use:   "vmware",
		Short: "Show VMware Fusion information",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Println("VMware Fusion Information:")
			
			// Check if VMware Fusion is installed
			// vmrun is in …/VMware Fusion.app/Contents/Library/vmrun
			// so we need to go up 3 levels to get to the .app bundle
			vmwareFusionApp := filepath.Clean(filepath.Join(globalConfig.VMwarePath, "..", "..", ".."))
			fmt.Printf("  Installation Path: %s\n", vmwareFusionApp)
			
			// VMware version
			plistPath := filepath.Join(vmwareFusionApp, "Contents", "Info.plist")
			if output, err := exec.Command("defaults", "read", plistPath, "CFBundleShortVersionString").Output(); err == nil {
				fmt.Printf("  Version: %s\n", strings.TrimSpace(string(output)))
			} else {
				GlobalLogger.Warning.Printf("Could not get VMware version: %v\n", err)
			}
			
			// Tool paths
			fmt.Printf("  vmrun Path: %s\n", vmwareTools.VMRunPath)
			fmt.Printf("  vmcli Path: %s\n", vmwareTools.VMCliPath)
			
			// Tool availability
			vmrunExists := FileExists(vmwareTools.VMRunPath)
			vmcliExists := FileExists(vmwareTools.VMCliPath)
			
			fmt.Printf("  vmrun Available: %t\n", vmrunExists)
			fmt.Printf("  vmcli Available: %t\n", vmcliExists)
			
			// Pro features
			if vmrunExists {
				if err := vmwareTools.CheckProFeatures(); err != nil {
					fmt.Printf("  Pro Features: No (%v)\n", err)
				} else {
					fmt.Printf("  Pro Features: Yes\n")
				}
			}
			
			// Running VMs
			if vmrunExists {
				runningVMs, err := vmwareTools.ListRunningVMs()
				if err != nil {
					fmt.Printf("  Running VMs: Error (%v)\n", err)
				} else {
					fmt.Printf("  Running VMs: %d\n", len(runningVMs))
					for i, vm := range runningVMs {
						if i < 3 { // Show first 3
							fmt.Printf("    • %s\n", vm)
						} else if i == 3 {
							fmt.Printf("    • ... and %d more\n", len(runningVMs)-3)
							break
						}
					}
				}
			}
			
			return nil
		},
	}

	// Drives info command
	drivesCmd := &cobra.Command{
		Use:   "drives",
		Short: "Show drive information",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Println("Drive Information:")
			
			drives, err := ListExternalDrives(globalConfig.ExFATMountBase)
			if err != nil {
				return fmt.Errorf("failed to list drives: %w", err)
			}
			
			if len(drives) == 0 {
				fmt.Println("  No external drives found")
				return nil
			}
			
			fmt.Printf("\n  %-20s %-10s %-10s %-12s %-10s\n", "Name", "Size", "Used", "Available", "Filesystem")
			fmt.Printf("  %s\n", strings.Repeat("-", 70))
			
			for _, drive := range drives {
				fmt.Printf("  %-20s %-10s %-10s %-12s %-10s\n", 
					drive.Name, drive.Size, drive.Used, drive.Available, drive.FileSystem)
			}
			
			// Show default drive status
			fmt.Println()
			defaultDriveFound := false
			for _, drive := range drives {
				if drive.Name == globalConfig.DefaultDrive {
					GlobalLogger.Success.Printf("Default archive drive '%s' is available\n", globalConfig.DefaultDrive)
					defaultDriveFound = true
					break
				}
			}
			
			if !defaultDriveFound {
				GlobalLogger.Warning.Printf("Default archive drive '%s' is not currently mounted\n", globalConfig.DefaultDrive)
			}
			
			return nil
		},
	}

	// Status command (overview)
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show overall system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			GlobalLogger.Info.Println("VMware Fusion CLI Status:")
			
			// Configuration status
			fmt.Printf("\n📋 Configuration:\n")
			fmt.Printf("  • Config file: %s\n", configPath)
			fmt.Printf("  • Local VM directory: %s\n", globalConfig.LocalVMDir)
			fmt.Printf("  • Default archive drive: %s\n", globalConfig.DefaultDrive)
			
			// VMware status
			fmt.Printf("\n🖥️  VMware Fusion:\n")
			fmt.Printf("  • Installation: Detected\n")
			
			if err := vmwareTools.CheckProFeatures(); err != nil {
				fmt.Printf("  • Edition: Standard (Pro features unavailable)\n")
			} else {
				fmt.Printf("  • Edition: Pro (All features available)\n")
			}
			
			// Running VMs
			if runningVMs, err := vmwareTools.ListRunningVMs(); err == nil {
				fmt.Printf("  • Running VMs: %d\n", len(runningVMs))
			}
			
			// Storage status
			fmt.Printf("\n💾 Storage:\n")
			
			// Check local VM directory space
			if output, err := exec.Command("df", "-h", globalConfig.LocalVMDir).Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				if len(lines) > 1 {
					fields := strings.Fields(lines[1])
					if len(fields) >= 4 {
						fmt.Printf("  • Local SSD available: %s\n", fields[3])
					}
				}
			} else {
				GlobalLogger.Warning.Printf("Could not get local VM directory space: %v\n", err)
			}
			
			// Check default drive
			drives, err := ListExternalDrives(globalConfig.ExFATMountBase)
			if err == nil {
				defaultDriveFound := false
				for _, drive := range drives {
					if drive.Name == globalConfig.DefaultDrive {
						fmt.Printf("  • Archive drive (%s): %s available\n", drive.Name, drive.Available)
						defaultDriveFound = true
						break
					}
				}
				if !defaultDriveFound {
					fmt.Printf("  • Archive drive (%s): Not mounted\n", globalConfig.DefaultDrive)
				}
			}
			
			// Show available commands
			fmt.Printf("\n🚀 Quick Commands:\n")
			fmt.Printf("  • List VMs: vmwfusion vm list\n")
			fmt.Printf("  • Find ISOs: vmwfusion iso find\n")
			fmt.Printf("  • Archive VM: vmwfusion migrate archive <vm_name>\n")
			fmt.Printf("  • Show config: vmwfusion config show\n")
			
			return nil
		},
	}

	cmd.AddCommand(systemCmd)
	cmd.AddCommand(vmwareCmd)
	cmd.AddCommand(drivesCmd)
	cmd.AddCommand(statusCmd)

	return cmd
}