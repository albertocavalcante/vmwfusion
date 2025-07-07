package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// ISO command flags
	isoSearchPaths []string
	recursive      bool
)

// isoCmd represents the iso command
var isoCmd = &cobra.Command{
	Use:   "iso",
	Short: "ISO file management and operations",
	Long:  `Manage ISO files including listing, validation, mounting, and unmounting operations for virtual machines.`,
}

var isoListCmd = &cobra.Command{
	Use:   "list [paths...]",
	Short: "List available ISO files",
	Long:  `Scan specified paths (or default locations) for ISO files and display information about them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		searchPaths := args
		if len(searchPaths) == 0 {
			searchPaths = getDefaultISOPaths()
		}

		isos, err := vmClient.ListISOs(ctx, searchPaths)
		if err != nil {
			return fmt.Errorf("failed to list ISOs: %w", err)
		}

		if jsonOutput {
			return outputJSON(isos)
		}

		if len(isos) == 0 {
			color.Yellow("No ISO files found in specified paths")
			return nil
		}

		color.Cyan("Available ISO Files:")
		fmt.Printf("%-40s %-15s %-15s %-20s\n", "PATH", "SIZE", "OS", "ARCHITECTURE")
		fmt.Println("--------------------------------------------------------------------------------")
		
		for _, iso := range isos {
			sizeStr := formatFileSize(iso.Size)
			fmt.Printf("%-40s %-15s %-15s %-20s\n", 
				truncatePath(iso.Path, 40), 
				sizeStr, 
				iso.OS, 
				iso.Architecture)
		}
		return nil
	},
}

var isoValidateCmd = &cobra.Command{
	Use:   "validate [iso-path]",
	Short: "Validate an ISO file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		info, err := vmClient.ValidateISO(ctx, args[0])
		if err != nil {
			return fmt.Errorf("ISO validation failed: %w", err)
		}

		if jsonOutput {
			return outputJSON(info)
		}

		color.Green("✓ ISO file is valid")
		fmt.Printf("Path: %s\n", info.Path)
		fmt.Printf("Name: %s\n", info.Name)
		fmt.Printf("Size: %s\n", formatFileSize(info.Size))
		fmt.Printf("OS: %s\n", info.OS)
		fmt.Printf("Architecture: %s\n", info.Architecture)
		fmt.Printf("Checksum: %s\n", info.Checksum)
		return nil
	},
}

var isoMountCmd = &cobra.Command{
	Use:   "mount [vm-id] [iso-path]",
	Short: "Mount an ISO file to a virtual machine",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Validate ISO first
		if _, err := vmClient.ValidateISO(ctx, args[1]); err != nil {
			return fmt.Errorf("ISO validation failed: %w", err)
		}

		if err := vmClient.MountISO(ctx, args[0], args[1]); err != nil {
			return fmt.Errorf("failed to mount ISO: %w", err)
		}

		color.Green("Successfully mounted ISO %s to VM: %s", args[1], args[0])
		return nil
	},
}

var isoUnmountCmd = &cobra.Command{
	Use:   "unmount [vm-id]",
	Short: "Unmount ISO file from a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.UnmountISO(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to unmount ISO: %w", err)
		}

		color.Green("Successfully unmounted ISO from VM: %s", args[0])
		return nil
	},
}

var isoInfoCmd = &cobra.Command{
	Use:   "info [iso-path]",
	Short: "Show detailed information about an ISO file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		info, err := vmClient.ValidateISO(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get ISO info: %w", err)
		}

		if jsonOutput {
			return outputJSON(info)
		}

		color.Cyan("ISO Information:")
		fmt.Printf("Path: %s\n", info.Path)
		fmt.Printf("Name: %s\n", info.Name)
		fmt.Printf("Size: %s (%d bytes)\n", formatFileSize(info.Size), info.Size)
		fmt.Printf("OS: %s\n", info.OS)
		fmt.Printf("Architecture: %s\n", info.Architecture)
		fmt.Printf("Created: %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Checksum: %s\n", info.Checksum)
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(isoCmd)

	// Add subcommands
	isoCmd.AddCommand(isoListCmd, isoValidateCmd, isoMountCmd, isoUnmountCmd, isoInfoCmd)

	// Global ISO flags
	isoCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	
	// List command flags
	isoListCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "search recursively in subdirectories")
}

// Helper functions
func getDefaultISOPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "Documents", "Virtual Machines"),
		"/Applications/VMware Fusion.app/Contents/Library/isoimages",
		"/usr/local/share/vmware-fusion/isoimages",
	}
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}