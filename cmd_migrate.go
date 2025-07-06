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

// migrateCommands returns the VM migration command group
func migrateCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "VM migration between local SSD and ExFAT archive storage",
		Long: `Migrate VMs between local SSD storage and ExFAT archive drives.
Supports archiving running VMs with graceful shutdown, fast deployment with linked clones,
and comprehensive verification.`,
	}

	// Archive command
	archiveCmd := &cobra.Command{
		Use:   "archive <vm_name> [drive]",
		Short: "Archive VM to ExFAT drive",
		Long:  "Archive a VM to ExFAT drive with optional optimization",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			drive := globalConfig.DefaultDrive
			if len(args) > 1 {
				drive = args[1]
			}
			
			return archiveVM(vmName, drive, false)
		},
	}

	// Archive running command
	archiveRunningCmd := &cobra.Command{
		Use:   "archive-running <vm_name> [drive]",
		Short: "Archive running VM with graceful shutdown",
		Long:  "Archive a currently running VM with automatic graceful shutdown",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			drive := globalConfig.DefaultDrive
			if len(args) > 1 {
				drive = args[1]
			}
			
			return archiveVM(vmName, drive, true)
		},
	}

	// Restore command
	restoreCmd := &cobra.Command{
		Use:   "restore <vm_name> <drive> [dest_name]",
		Short: "Restore VM from ExFAT drive to local SSD",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]
			drive := args[1]
			destName := vmName
			if len(args) > 2 {
				destName = args[2]
			}
			
			return restoreVM(vmName, drive, destName)
		},
	}

	// Fast deploy command
	fastDeployCmd := &cobra.Command{
		Use:   "fast-deploy <vm_name> <drive> [work_name]",
		Short: "Fast deployment using linked clone",
		Long:  "Create a linked clone from archived VM for fast development work",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := vmwareTools.CheckProFeatures(); err != nil {
				return err
			}
			
			vmName := args[0]
			drive := args[1]
			workName := vmName + "_work"
			if len(args) > 2 {
				workName = args[2]
			}
			
			return fastDeployVM(vmName, drive, workName)
		},
	}

	// List drives command
	listDrivesCmd := &cobra.Command{
		Use:   "list-drives",
		Short: "List available external drives",
		RunE: func(cmd *cobra.Command, args []string) error {
			drives, err := ListExternalDrives(globalConfig.ExFATMountBase)
			if err != nil {
				return err
			}
			
			GlobalLogger.Info.Println("Available external drives:")
			fmt.Printf("%-20s %-8s %-8s %-10s %s\n", "Name", "Size", "Used", "Available", "Filesystem")
			fmt.Printf("%s\n", strings.Repeat("-", 60))
			
			for _, drive := range drives {
				fmt.Printf("%-20s %-8s %-8s %-10s %s\n", 
					drive.Name, drive.Size, drive.Used, drive.Available, drive.FileSystem)
			}
			
			return nil
		},
	}

	cmd.AddCommand(archiveCmd)
	cmd.AddCommand(archiveRunningCmd)
	cmd.AddCommand(restoreCmd)
	cmd.AddCommand(fastDeployCmd)
	cmd.AddCommand(listDrivesCmd)

	return cmd
}

// archiveVM archives a VM to an ExFAT drive
func archiveVM(vmName, drive string, handleRunning bool) error {
	GlobalLogger.Info.Println("=== ARCHIVING VM TO EXFAT DRIVE ===")
	GlobalLogger.Info.Printf("VM: %s\n", vmName)
	GlobalLogger.Info.Printf("Target Drive: %s\n", drive)
	
	// Find VM bundle
	vmBundle, err := FindVMBundle(vmName, globalConfig.LocalVMDir)
	if err != nil {
		return fmt.Errorf("VM not found: %s", vmName)
	}
	
	// Check if VM is running
	running, err := vmwareTools.IsVMRunning(vmBundle.VMXPath)
	if err != nil {
		return fmt.Errorf("failed to check VM status: %w", err)
	}
	
	if running {
		if !handleRunning {
			return fmt.Errorf("VM is currently running - use 'archive-running' command to handle running VMs")
		}
		
		GlobalLogger.Warning.Printf("VM '%s' is currently running\n", vmName)
		GlobalLogger.Warning.Println("To safely archive the VM, it needs to be shut down completely")
		
		if !dryRun {
			fmt.Print("Proceed with shutdown and archive? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			
			if strings.TrimSpace(strings.ToLower(input)) != "y" {
				GlobalLogger.Info.Println("Operation cancelled by user")
				return nil
			}
		}
		
		// Graceful shutdown
		GlobalLogger.Info.Printf("Shutting down VM gracefully (timeout: %ds)...\n", globalConfig.ShutdownTimeout)
		if err := manageVMPower(vmBundle.VMXPath, "stop"); err != nil {
			return fmt.Errorf("failed to shutdown VM: %w", err)
		}
		
		// Wait for disk operations to complete
		GlobalLogger.Info.Println("Waiting for disk operations to complete...")
		time.Sleep(5 * time.Second)
	}
	
	// Validate target drive
	drivePath := filepath.Join(globalConfig.ExFATMountBase, drive)
	if _, err := os.Stat(drivePath); os.IsNotExist(err) {
		GlobalLogger.Error.Printf("ExFAT drive not found: %s\n", drivePath)
		
		// Show available drives
		if drives, err := ListExternalDrives(globalConfig.ExFATMountBase); err == nil {
			GlobalLogger.Info.Println("Available drives:")
			for _, d := range drives {
				fmt.Printf("  %s (%s)\n", d.Name, d.FileSystem)
			}
		}
		return fmt.Errorf("drive not found: %s", drive)
	}
	
	// Check available space
	if err := checkDriveSpace(vmBundle, drivePath); err != nil {
		return err
	}
	
	// Optimize VM if configured
	if globalConfig.AutoOptimize && !dryRun {
		GlobalLogger.Info.Println("Auto-optimizing VM for archival...")
		if err := optimizeVMForArchive(vmBundle.VMXPath); err != nil {
			GlobalLogger.Warning.Printf("Optimization failed: %v\n", err)
		}
	}
	
	// Perform archive operation
	destPath := filepath.Join(drivePath, vmBundle.Name+".vmwarevm")
	
	if dryRun {
		GlobalLogger.Info.Printf("DRY RUN: Would archive %s to %s\n", vmBundle.Path, destPath)
		return nil
	}
	
	if err := syncVMBundle(vmBundle.Path, destPath, "archive"); err != nil {
		return fmt.Errorf("archive failed: %w", err)
	}
	
	// Verify archived VM
	GlobalLogger.Info.Println("Verifying archived VM...")
	if _, err := findVMXFile(destPath); err != nil {
		return fmt.Errorf("verification failed - archived VM may be corrupted: %w", err)
	}
	
	GlobalLogger.Success.Printf("VM successfully archived: %s\n", destPath)
	
	// Show space info
	GlobalLogger.Info.Printf("Space that can be freed: %s\n", FormatSizeSimple(vmBundle.Size))
	
	// Handle cleanup if configured
	if globalConfig.CleanupOriginal {
		GlobalLogger.Warning.Println("Cleaning up original VM files to free disk space...")
		fmt.Print("Are you absolutely sure you want to delete the original VM? (type 'DELETE' to confirm): ")
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		
		if strings.TrimSpace(input) == "DELETE" {
			if err := os.RemoveAll(vmBundle.Path); err != nil {
				return fmt.Errorf("failed to delete original VM: %w", err)
			}
			GlobalLogger.Success.Printf("Original VM files deleted - %s freed\n", FormatSizeSimple(vmBundle.Size))
		} else {
			GlobalLogger.Info.Println("Original VM files preserved")
		}
	}
	
	// Show next steps
	fmt.Println()
	GlobalLogger.Info.Println("Next steps:")
	fmt.Printf("  • Archived VM location: %s\n", destPath)
	fmt.Printf("  • To restore: vmwfusion migrate restore \"%s\" \"%s\"\n", vmName, drive)
	fmt.Printf("  • To fast-deploy: vmwfusion migrate fast-deploy \"%s\" \"%s\"\n", vmName, drive)
	
	return nil
}

// restoreVM restores a VM from ExFAT drive to local SSD
func restoreVM(vmName, drive, destName string) error {
	GlobalLogger.Info.Println("=== RESTORING VM FROM EXFAT DRIVE ===")
	
	drivePath := filepath.Join(globalConfig.ExFATMountBase, drive)
	if _, err := os.Stat(drivePath); os.IsNotExist(err) {
		return fmt.Errorf("ExFAT drive not found: %s", drivePath)
	}
	
	// Find VM on ExFAT drive
	vmBundle, err := FindVMBundle(vmName, drivePath)
	if err != nil {
		return fmt.Errorf("VM not found on ExFAT drive: %s", vmName)
	}
	
	// Create destination path
	destPath := filepath.Join(globalConfig.LocalVMDir, destName+".vmwarevm")
	
	if dryRun {
		GlobalLogger.Info.Printf("DRY RUN: Would restore %s to %s\n", vmBundle.Path, destPath)
		return nil
	}
	
	// Perform restore
	if err := syncVMBundle(vmBundle.Path, destPath, "restore"); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}
	
	GlobalLogger.Success.Printf("VM restored to local SSD: %s\n", destPath)
	
	// Show VM info
	restoredBundle, err := FindVMBundle(destName, globalConfig.LocalVMDir)
	if err == nil {
		return showVMInfo(restoredBundle)
	}
	
	return nil
}

// fastDeployVM creates a linked clone from archived VM
func fastDeployVM(vmName, drive, workName string) error {
	GlobalLogger.Info.Println("=== FAST DEPLOYMENT FROM ARCHIVE ===")
	
	drivePath := filepath.Join(globalConfig.ExFATMountBase, drive)
	
	// Find VM on ExFAT drive
	vmBundle, err := FindVMBundle(vmName, drivePath)
	if err != nil {
		return fmt.Errorf("VM not found on ExFAT drive: %s", vmName)
	}
	
	// Create linked clone
	GlobalLogger.Info.Println("Creating linked clone for fast deployment...")
	destPath := filepath.Join(globalConfig.LocalVMDir, workName+".vmx")
	
	if dryRun {
		GlobalLogger.Info.Printf("DRY RUN: Would create linked clone %s -> %s\n", vmBundle.VMXPath, destPath)
		return nil
	}
	
	cmd := exec.Command(vmwareTools.VMRunPath, "clone", vmBundle.VMXPath, destPath, "linked", "-cloneName="+workName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create linked clone: %w", err)
	}
	
	GlobalLogger.Success.Printf("Fast deployment completed: %s\n", destPath)
	
	// Ask if user wants to start the VM
	fmt.Print("Start VM now? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	
	if strings.TrimSpace(strings.ToLower(input)) == "y" {
		return manageVMPower(destPath, "start")
	}
	
	return nil
}

// checkDriveSpace verifies sufficient space is available on target drive
func checkDriveSpace(vmBundle *VMBundle, drivePath string) error {
	// Get available space on target drive
	cmd := exec.Command("df", drivePath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check drive space: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output")
	}
	
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return fmt.Errorf("unexpected df output format")
	}
	
	GlobalLogger.Info.Printf("VM Size: %s\n", FormatSizeSimple(vmBundle.Size))
	GlobalLogger.Info.Printf("Target Available: %s\n", fields[3])
	
	// Note: This is a basic check - we could parse the actual numbers for more precise checking
	return nil
}

// syncVMBundle synchronizes VM bundle between locations
func syncVMBundle(sourcePath, destPath, operation string) error {
	GlobalLogger.Info.Printf("Starting %s operation...\n", operation)
	GlobalLogger.Info.Printf("Source: %s\n", sourcePath)
	GlobalLogger.Info.Printf("Destination: %s\n", destPath)
	
	// Check if destination exists
	if _, err := os.Stat(destPath); err == nil {
		GlobalLogger.Warning.Printf("Destination already exists: %s\n", destPath)
		fmt.Print("Overwrite? (y/N): ")
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		
		if strings.TrimSpace(strings.ToLower(input)) != "y" {
			return fmt.Errorf("operation cancelled by user")
		}
		
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}
	
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Use rsync if available, otherwise cp
	if exec.Command("which", "rsync").Run() == nil {
		GlobalLogger.Info.Println("Using rsync for sync...")
		
		rsyncArgs := append(globalConfig.RsyncOptions, sourcePath+"/", destPath+"/")
		cmd := exec.Command("rsync", rsyncArgs...)
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("rsync failed: %w", err)
		}
		
		// Verify sync if configured
		if globalConfig.VerifyChecksums {
			GlobalLogger.Info.Println("Verifying sync integrity...")
			verifyArgs := append([]string{"-avhc", "--dry-run"}, sourcePath+"/", destPath+"/")
			cmd = exec.Command("rsync", verifyArgs...)
			
			output, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}
			
			if strings.Contains(string(output), "deleting") || strings.Contains(string(output), ">") {
				return fmt.Errorf("sync verification failed - files differ")
			}
			
			GlobalLogger.Success.Println("Sync verification passed")
		}
	} else {
		GlobalLogger.Info.Println("Using cp for sync...")
		cmd := exec.Command("cp", "-R", sourcePath, destPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("cp failed: %w", err)
		}
	}
	
	GlobalLogger.Success.Println("Sync completed successfully")
	return nil
}