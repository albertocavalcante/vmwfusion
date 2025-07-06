package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// isoCommands returns the ISO management command group
func isoCommands() *cobra.Command {
	var (
		targetDrive    string
		includeVMware  bool
		forceOverwrite bool
	)

	cmd := &cobra.Command{
		Use:   "iso",
		Short: "ISO file management operations",
		Long: `Discover, export, and manage ISO files on your system.
Supports hash-based duplicate detection, manifest tracking, and intelligent workflows.`,
	}

	// Add flags
	cmd.PersistentFlags().StringVarP(&targetDrive, "drive", "d", "", "Target drive name (uses config default if not specified)")
	cmd.PersistentFlags().BoolVar(&includeVMware, "include-vmware", true, "Include VMware built-in ISOs")
	cmd.PersistentFlags().BoolVar(&forceOverwrite, "force", false, "Force overwrite existing files")

	// Find command
	findCmd := &cobra.Command{
		Use:   "find",
		Short: "Find all ISO files on the system",
		RunE: func(cmd *cobra.Command, args []string) error {
			discovery := NewISODiscovery(GlobalLogger)
			
			files, err := discovery.FindAllISOs()
			if err != nil {
				return err
			}
			
			discovery.DisplayFiles(files)
			return nil
		},
	}

	// Export command
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export found ISO files to external drive",
		Long:  "Export ISO files that were previously found with the 'find' command",
		RunE: func(cmd *cobra.Command, args []string) error {
			discovery := NewISODiscovery(GlobalLogger)
			
			// First find ISOs
			files, err := discovery.FindAllISOs()
			if err != nil {
				return err
			}
			
			if len(files) == 0 {
				GlobalLogger.Warning.Println("No ISO files found")
				return nil
			}
			
			drive := targetDrive
			if drive == "" {
				drive = globalConfig.DefaultDrive
			}
			
			config := ExportConfig{
				TargetDrive:    drive,
				ArchiveDir:     isoArchiveDir,
				ManifestFile:   ".iso_manifest.json",
				IncludeVMware:  includeVMware,
				ForceOverwrite: forceOverwrite,
			}
			
			exporter := NewISOExporter(GlobalLogger, discovery, config)
			_, err = exporter.ExportISOs(files)
			return err
		},
	}

	// Quick export command
	quickExportCmd := &cobra.Command{
		Use:   "quick-export",
		Short: "Find and export ISO files in one step",
		RunE: func(cmd *cobra.Command, args []string) error {
			discovery := NewISODiscovery(GlobalLogger)
			
			drive := targetDrive
			if drive == "" {
				drive = globalConfig.DefaultDrive
			}
			
			GlobalLogger.Info.Println("🚀 Quick ISO Export Workflow")
			GlobalLogger.Info.Printf("Target drive: %s\n", drive)
			fmt.Println()
			
			// Find ISOs
			files, err := discovery.FindAllISOs()
			if err != nil {
				return err
			}
			
			if len(files) == 0 {
				GlobalLogger.Warning.Println("No ISO files found")
				return nil
			}
			
			discovery.DisplayFiles(files)
			
			// Ask for confirmation
			fmt.Println()
			fmt.Printf("Export ISOs to %s? (y/N): ", drive)
			
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			
			if strings.TrimSpace(strings.ToLower(input)) != "y" {
				GlobalLogger.Info.Println("Export cancelled")
				return nil
			}
			
			config := ExportConfig{
				TargetDrive:    drive,
				ArchiveDir:     "ISO_Archive",
				ManifestFile:   ".iso_manifest.json",
				IncludeVMware:  includeVMware,
				ForceOverwrite: forceOverwrite,
			}
			
			exporter := NewISOExporter(GlobalLogger, discovery, config)
			result, err := exporter.ExportISOs(files)
			if err != nil {
				return err
			}
			
			// Offer deletion prompts
			if result.FilesExported > 0 {
				fmt.Println()
				GlobalLogger.Info.Println("💡 Next steps:")
				fmt.Printf("  • Verify exports: vmwfusion iso show-archived -d %s\n", drive)
				fmt.Printf("  • Delete originals: vmwfusion iso verify-delete -d %s\n", drive)
			}
			
			return nil
		},
	}

	// Verify delete command
	verifyDeleteCmd := &cobra.Command{
		Use:   "verify-delete",
		Short: "Fast verify using cached paths and prompt for deletion",
		RunE: func(cmd *cobra.Command, args []string) error {
			discovery := NewISODiscovery(GlobalLogger)
			
			drive := targetDrive
			if drive == "" {
				drive = globalConfig.DefaultDrive
			}
			
			config := ExportConfig{
				TargetDrive:    drive,
				ArchiveDir:     "ISO_Archive",
				ManifestFile:   ".iso_manifest.json",
				IncludeVMware:  includeVMware,
				ForceOverwrite: forceOverwrite,
			}
			
			verifier := NewISOVerifier(GlobalLogger, discovery, config)
			return verifier.VerifyAndDelete()
		},
	}

	// Show archived command
	showArchivedCmd := &cobra.Command{
		Use:   "show-archived",
		Short: "Show what's currently archived on the target drive",
		RunE: func(cmd *cobra.Command, args []string) error {
			discovery := NewISODiscovery(GlobalLogger)
			
			drive := targetDrive
			if drive == "" {
				drive = globalConfig.DefaultDrive
			}
			
			config := ExportConfig{
				TargetDrive:    drive,
				ArchiveDir:     "ISO_Archive",
				ManifestFile:   ".iso_manifest.json",
				IncludeVMware:  includeVMware,
				ForceOverwrite: forceOverwrite,
			}
			
			verifier := NewISOVerifier(GlobalLogger, discovery, config)
			return verifier.ShowArchived()
		},
	}

	cmd.AddCommand(findCmd)
	cmd.AddCommand(exportCmd)
	cmd.AddCommand(quickExportCmd)
	cmd.AddCommand(verifyDeleteCmd)
	cmd.AddCommand(showArchivedCmd)

	return cmd
}