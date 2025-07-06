package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// vmCommands returns the VM management command group
func vmCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Virtual machine lifecycle management",
		Long: `Manage VMware Fusion VMs including power operations, cloning, 
optimization, and information display.`,
	}

	// List VMs command
	listCmd := &cobra.Command{
		Use:   "list [directory]",
		Short: "List virtual machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			searchDir := globalConfig.LocalVMDir
			if len(args) > 0 {
				searchDir = args[0]
			}
			
			GlobalLogger.Info.Printf("VMs in %s:\n", searchDir)
			
			err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Continue on errors
				}
				
				if info.IsDir() && strings.HasSuffix(info.Name(), ".vmwarevm") {
					vmBundle, err := FindVMBundle(strings.TrimSuffix(info.Name(), ".vmwarevm"), searchDir)
					if err != nil {
						fmt.Printf("  %s (error: %v)\n", info.Name(), err)
						return nil
					}
					
					// Check if running
					running, _ := vmwareTools.IsVMRunning(vmBundle.VMXPath)
					status := "stopped"
					if running {
						status = "running"
					}
					
					fmt.Printf("  %-50s %s (%s)\n", vmBundle.Name, FormatSizeSimple(vmBundle.Size), status)
				}
				
				return nil
			})
			
			return err
		},
	}

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info <vm_name>",
		Short: "Show detailed VM information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			
			vmBundle, err := FindVMBundle(vmName, globalConfig.LocalVMDir)
			if err != nil {
				return fmt.Errorf("VM not found: %s", vmName)
			}
			
			return showVMInfo(vmBundle)
		},
	}

	// Power management command
	powerCmd := &cobra.Command{
		Use:   "power <vm_name> <action>",
		Short: "Manage VM power state",
		Long:  "Control VM power: start, stop, reset, suspend",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			action := args[1]
			
			vmBundle, err := FindVMBundle(vmName, globalConfig.LocalVMDir)
			if err != nil {
				return fmt.Errorf("VM not found: %s", vmName)
			}
			
			return manageVMPower(vmBundle.VMXPath, action)
		},
	}

	// Clone command
	cloneCmd := &cobra.Command{
		Use:   "clone <source_vm> <dest_name> [type]",
		Short: "Create a VM clone",
		Long:  "Create a VM clone (full or linked). Requires VMware Fusion Pro.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := vmwareTools.CheckProFeatures(); err != nil {
				return err
			}
			
			sourceVM := args[0]
			destName := args[1]
			cloneType := globalConfig.PreferredCloneType
			if len(args) > 2 {
				cloneType = args[2]
			}
			
			return createVMClone(sourceVM, destName, cloneType)
		},
	}

	// Optimize command
	optimizeCmd := &cobra.Command{
		Use:   "optimize <vm_name>",
		Short: "Optimize VM for archival storage",
		Long:  "Remove snapshots and compact virtual disks to reduce VM size",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			
			vmBundle, err := FindVMBundle(vmName, globalConfig.LocalVMDir)
			if err != nil {
				return fmt.Errorf("VM not found: %s", vmName)
			}
			
			return optimizeVMForArchive(vmBundle.VMXPath)
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(infoCmd)
	cmd.AddCommand(powerCmd)
	cmd.AddCommand(cloneCmd)
	cmd.AddCommand(optimizeCmd)

	return cmd
}

// showVMInfo displays detailed information about a VM
func showVMInfo(vmBundle *VMBundle) error {
	GlobalLogger.Info.Println("VM Information:")
	fmt.Printf("Name: %s\n", vmBundle.Name)
	fmt.Printf("Bundle: %s\n", vmBundle.Path)
	fmt.Printf("VMX File: %s\n", vmBundle.VMXPath)
	fmt.Printf("Bundle Size: %s\n", FormatSizeSimple(vmBundle.Size))
	
	// Power state
	running, err := vmwareTools.IsVMRunning(vmBundle.VMXPath)
	if err != nil {
		fmt.Printf("Power State: Unknown (error: %v)\n", err)
	} else {
		status := "Stopped"
		if running {
			status = "Running"
		}
		fmt.Printf("Power State: %s\n", status)
	}
	
	// VM configuration using vmcli if available
	if FileExists(vmwareTools.VMCliPath) {
		fmt.Println("\nVM Configuration:")
		
		// Get VM settings
		configs := map[string]string{
			"Display Name": "displayName",
			"Memory":       "memsize",
			"CPUs":         "numvcpus",
			"Guest OS":     "guestOS",
		}
		
		for label, param := range configs {
			cmd := exec.Command(vmwareTools.VMCliPath, "ConfigParams", "GetEntry", param, vmBundle.VMXPath)
			if output, err := cmd.Output(); err == nil {
				value := strings.TrimSpace(string(output))
				if value != "" {
					fmt.Printf("  %s: %s\n", label, value)
				}
			}
		}
	}
	
	// Disk information
	fmt.Println("\nDisk Information:")
	err = filepath.Walk(vmBundle.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vmdk") && 
		   !strings.Contains(info.Name(), "-s") { // Skip snapshot files
			fmt.Printf("  %s: %s\n", info.Name(), FormatSizeSimple(info.Size()))
		}
		
		return nil
	})
	
	if err != nil {
		GlobalLogger.Warning.Printf("Could not read disk information: %v\n", err)
	}
	
	// Snapshot information
	fmt.Println("\nSnapshots:")
	snapshotCount := 0
	err = filepath.Walk(vmBundle.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vmsd") {
			// Count snapshots in the file
			if data, err := os.ReadFile(path); err == nil {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					if strings.Contains(line, "snapshot") {
						snapshotCount++
					}
				}
			}
		}
		
		return nil
	})
	
	fmt.Printf("  Count: %d\n", snapshotCount)
	
	return nil
}

// manageVMPower handles VM power state changes
func manageVMPower(vmxPath, action string) error {
	GlobalLogger.Info.Printf("Managing VM power: %s\n", action)
	
	switch action {
	case "start":
		running, err := vmwareTools.IsVMRunning(vmxPath)
		if err != nil {
			return err
		}
		if running {
			GlobalLogger.Info.Println("VM is already running")
			return nil
		}
		
		cmd := exec.Command(vmwareTools.VMRunPath, "start", vmxPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start VM: %w", err)
		}
		GlobalLogger.Success.Println("VM started")
		
	case "stop":
		running, err := vmwareTools.IsVMRunning(vmxPath)
		if err != nil {
			return err
		}
		if !running {
			GlobalLogger.Info.Println("VM is already stopped")
			return nil
		}
		
		GlobalLogger.Info.Println("Stopping VM (soft shutdown)...")
		cmd := exec.Command(vmwareTools.VMRunPath, "stop", vmxPath, "soft")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop VM: %w", err)
		}
		
		// Wait for VM to stop
		timeout := time.Duration(globalConfig.ShutdownTimeout) * time.Second
		start := time.Now()
		for time.Since(start) < timeout {
			running, err := vmwareTools.IsVMRunning(vmxPath)
			if err != nil || !running {
				break
			}
			time.Sleep(2 * time.Second)
		}
		
		// Check if still running and force stop if needed
		if running, _ := vmwareTools.IsVMRunning(vmxPath); running {
			GlobalLogger.Warning.Println("Soft stop timed out, forcing stop...")
			cmd = exec.Command(vmwareTools.VMRunPath, "stop", vmxPath, "hard")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to force stop VM: %w", err)
			}
		}
		
		GlobalLogger.Success.Println("VM stopped")
		
	case "reset":
		cmd := exec.Command(vmwareTools.VMRunPath, "reset", vmxPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reset VM: %w", err)
		}
		GlobalLogger.Success.Println("VM reset")
		
	case "suspend":
		cmd := exec.Command(vmwareTools.VMRunPath, "suspend", vmxPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to suspend VM: %w", err)
		}
		GlobalLogger.Success.Println("VM suspended")
		
	default:
		return fmt.Errorf("unknown power action: %s (valid: start, stop, reset, suspend)", action)
	}
	
	return nil
}

// createVMClone creates a VM clone
func createVMClone(sourceVM, destName, cloneType string) error {
	GlobalLogger.Info.Printf("Creating %s clone: %s\n", cloneType, destName)
	
	// Find source VM
	sourceBundle, err := FindVMBundle(sourceVM, globalConfig.LocalVMDir)
	if err != nil {
		return fmt.Errorf("source VM not found: %s", sourceVM)
	}
	
	// Stop source VM if running
	if running, _ := vmwareTools.IsVMRunning(sourceBundle.VMXPath); running {
		GlobalLogger.Warning.Println("Stopping source VM...")
		if err := manageVMPower(sourceBundle.VMXPath, "stop"); err != nil {
			return fmt.Errorf("failed to stop source VM: %w", err)
		}
		// Wait for VM to actually stop
		for i := 0; i < 30; i++ { // Max 30 seconds
			if running, _ := vmwareTools.IsVMRunning(sourceBundle.VMXPath); !running {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	
	// Create destination path
	destPath := filepath.Join(globalConfig.LocalVMDir, destName+".vmx")
	
	// Create clone
	GlobalLogger.Info.Println("Creating clone...")
	cmd := exec.Command(vmwareTools.VMRunPath, "clone", sourceBundle.VMXPath, destPath, cloneType, "-cloneName="+destName)
	
	if dryRun {
		GlobalLogger.Info.Printf("DRY RUN: Would execute: %s\n", strings.Join(cmd.Args, " "))
		return nil
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create clone: %w", err)
	}
	
	GlobalLogger.Success.Printf("Clone created successfully: %s\n", destPath)
	
	// Ask if user wants to start the clone
	fmt.Print("Start cloned VM now? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		GlobalLogger.Warning.Printf("Failed to read input: %v\n", err)
		return nil
	}
	
	if input := strings.TrimSpace(strings.ToLower(input)); input == "y" || input == "yes" {
		return manageVMPower(destPath, "start")
	}
	
	return nil
}

// optimizeVMForArchive optimizes a VM for archival storage
func optimizeVMForArchive(vmxPath string) error {
	GlobalLogger.Info.Println("Optimizing VM for archival...")
	
	// Ensure VM is stopped
	if running, _ := vmwareTools.IsVMRunning(vmxPath); running {
		GlobalLogger.Info.Println("Stopping VM for optimization...")
		if err := manageVMPower(vmxPath, "stop"); err != nil {
			return err
		}
	}
	
	// Delete snapshots
	GlobalLogger.Info.Println("Removing snapshots...")
	cmd := exec.Command(vmwareTools.VMRunPath, "deleteSnapshot", vmxPath, "*")
	if err := cmd.Run(); err != nil {
		GlobalLogger.Warning.Printf("Failed to delete snapshots (may not exist): %v\n", err)
	}
	
	// Compact disks
	vmBundle := filepath.Dir(vmxPath)
	GlobalLogger.Info.Println("Compacting virtual disks...")
	
	err := filepath.Walk(vmBundle, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vmdk") && 
		   !strings.Contains(info.Name(), "-s") { // Skip snapshot files
			
			GlobalLogger.Info.Printf("Compacting disk: %s\n", info.Name())
			
			// Try vmcli first
			if FileExists(vmwareTools.VMCliPath) {
				cmd := exec.Command(vmwareTools.VMCliPath, "Disk", "Compact", path, vmxPath)
				if err := cmd.Run(); err == nil {
					return nil // Success with vmcli
				}
			}
			
			// Fallback to vmware-vdiskmanager
			vdiskManager := filepath.Join(vmwareTools.BasePath, "vmware-vdiskmanager")
			if FileExists(vdiskManager) {
				cmd := exec.Command(vdiskManager, "-k", path)
				if err := cmd.Run(); err != nil {
					GlobalLogger.Warning.Printf("Failed to compact %s: %v\n", info.Name(), err)
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("error during disk compaction: %w", err)
	}
	
	GlobalLogger.Success.Println("VM optimization completed")
	return nil
}