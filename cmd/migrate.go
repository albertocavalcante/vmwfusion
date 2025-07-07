package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"vmwfusion/internal/types"
)

var (
	// Migration flags
	sourcePath      string
	destinationPath string
	convertFormat   string
	compress        bool
	keepOriginal    bool
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Virtual machine migration operations",
	Long:  `Migrate virtual machines between different locations, formats, or configurations.`,
}

var migrateVMCmd = &cobra.Command{
	Use:   "vm [vm-id]",
	Short: "Migrate a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		options := &types.MigrationOptions{
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			ConvertFormat:   convertFormat,
			Compress:        compress,
			KeepOriginal:    keepOriginal,
		}

		color.Cyan("Starting migration for VM: %s", args[0])
		if destinationPath != "" {
			fmt.Printf("Destination: %s\n", destinationPath)
		}
		if convertFormat != "" {
			fmt.Printf("Converting to format: %s\n", convertFormat)
		}

		if err := vmClient.MigrateVM(ctx, args[0], options); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		color.Green("✓ Successfully migrated VM: %s", args[0])
		return nil
	},
}

var migratePathCmd = &cobra.Command{
	Use:   "path [vm-id] [new-path]",
	Short: "Move a virtual machine to a new path",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		options := &types.MigrationOptions{
			DestinationPath: args[1],
			KeepOriginal:    keepOriginal,
		}

		color.Cyan("Moving VM %s to: %s", args[0], args[1])

		if err := vmClient.MigrateVM(ctx, args[0], options); err != nil {
			return fmt.Errorf("failed to move VM: %w", err)
		}

		color.Green("✓ Successfully moved VM to: %s", args[1])
		return nil
	},
}

var migrateConvertCmd = &cobra.Command{
	Use:   "convert [vm-id] [format]",
	Short: "Convert a virtual machine to a different format",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		options := &types.MigrationOptions{
			ConvertFormat: args[1],
			Compress:      compress,
			KeepOriginal:  keepOriginal,
		}

		color.Cyan("Converting VM %s to format: %s", args[0], args[1])

		if err := vmClient.MigrateVM(ctx, args[0], options); err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		color.Green("✓ Successfully converted VM to %s format", args[1])
		return nil
	},
}

var migrateExportCmd = &cobra.Command{
	Use:   "export [vm-id]",
	Short: "Export a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		options := &types.ExportOptions{
			Format:      convertFormat,
			Destination: destinationPath,
			Compress:    compress,
		}

		color.Cyan("Exporting VM: %s", args[0])
		if destinationPath != "" {
			fmt.Printf("Destination: %s\n", destinationPath)
		}

		if err := vmClient.ExportVM(ctx, args[0], options); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		color.Green("✓ Successfully exported VM: %s", args[0])
		return nil
	},
}

var migrateImportCmd = &cobra.Command{
	Use:   "import [vm-path]",
	Short: "Import a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		color.Cyan("Importing VM from: %s", args[0])

		vm, err := vmClient.ImportVM(ctx, args[0])
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		if jsonOutput {
			return outputJSON(vm)
		}

		color.Green("✓ Successfully imported VM: %s", vm.Name)
		fmt.Printf("VM ID: %s\n", vm.ID)
		fmt.Printf("Path: %s\n", vm.Path)
		return nil
	},
}

var migrateBatchCmd = &cobra.Command{
	Use:   "batch [source-dir] [dest-dir]",
	Short: "Batch migrate multiple VMs",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// Discover VMs in source directory
		result, err := vmClient.DiscoverVMs(ctx, []string{args[0]})
		if err != nil {
			return fmt.Errorf("failed to discover VMs: %w", err)
		}

		if len(result.VMs) == 0 {
			color.Yellow("No VMs found in source directory: %s", args[0])
			return nil
		}

		color.Cyan("Found %d VMs for batch migration", len(result.VMs))

		successCount := 0
		for _, vm := range result.VMs {
			options := &types.MigrationOptions{
				DestinationPath: args[1],
				KeepOriginal:    keepOriginal,
				Compress:        compress,
			}

			fmt.Printf("Migrating %s... ", vm.Name)
			if err := vmClient.MigrateVM(ctx, vm.ID, options); err != nil {
				color.Red("✗ Failed: %v", err)
				continue
			}
			color.Green("✓ Success")
			successCount++
		}

		color.Cyan("Batch migration completed: %d/%d successful", successCount, len(result.VMs))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Add subcommands
	migrateCmd.AddCommand(migrateVMCmd, migratePathCmd, migrateConvertCmd, migrateExportCmd, migrateImportCmd, migrateBatchCmd)

	// Global migration flags
	migrateCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	migrateCmd.PersistentFlags().BoolVarP(&keepOriginal, "keep-original", "k", false, "keep original VM after migration")
	migrateCmd.PersistentFlags().BoolVarP(&compress, "compress", "z", false, "compress during migration")

	// VM migrate flags
	migrateVMCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "source path for migration")
	migrateVMCmd.Flags().StringVarP(&destinationPath, "dest", "d", "", "destination path for migration")
	migrateVMCmd.Flags().StringVarP(&convertFormat, "format", "f", "", "target format for conversion")

	// Export flags
	migrateExportCmd.Flags().StringVarP(&destinationPath, "dest", "d", "", "export destination path")
	migrateExportCmd.Flags().StringVarP(&convertFormat, "format", "f", "ova", "export format (ova, ovf, vmx)")

	// Batch migration flags
	migrateBatchCmd.Flags().BoolVarP(&keepOriginal, "keep-original", "k", true, "keep original VMs after batch migration")
}