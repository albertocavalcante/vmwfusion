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
	// Discovery flags
	discoveryPaths  []string
	includeHidden   bool
	maxDepth        int
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover virtual machines in specified paths",
	Long:  `Scan file system paths to discover existing virtual machines and their configurations.`,
}

var discoverVMsCmd = &cobra.Command{
	Use:   "vms [paths...]",
	Short: "Discover virtual machines in specified paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		searchPaths := args
		if len(searchPaths) == 0 {
			searchPaths = getDefaultVMPaths()
		}

		color.Cyan("Discovering VMs in paths: %v", searchPaths)

		result, err := vmClient.DiscoverVMs(ctx, searchPaths)
		if err != nil {
			return fmt.Errorf("discovery failed: %w", err)
		}

		if jsonOutput {
			return outputJSON(result)
		}

		if result.Count == 0 {
			color.Yellow("No virtual machines found in specified paths")
			return nil
		}

		color.Green("Found %d virtual machine(s)", result.Count)
		fmt.Printf("%-30s %-15s %-15s %-40s\n", "NAME", "STATE", "OS", "PATH")
		fmt.Println("--------------------------------------------------------------------------------")

		for _, vm := range result.VMs {
			stateColor := getStateColor(vm.State)
			fmt.Printf("%-30s %-15s %-15s %-40s\n", 
				vm.Name, 
				stateColor.Sprintf("%s", vm.State), 
				vm.OS,
				truncatePath(vm.Path, 40))
		}

		return nil
	},
}

var discoverPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show default discovery paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultPaths := getDefaultVMPaths()

		if jsonOutput {
			return outputJSON(map[string][]string{"default_paths": defaultPaths})
		}

		color.Cyan("Default VM Discovery Paths:")
		for i, path := range defaultPaths {
			exists := "✗"
			if _, err := os.Stat(path); err == nil {
				exists = "✓"
			}
			fmt.Printf("%d. %s %s\n", i+1, exists, path)
		}

		return nil
	},
}

var discoverScanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Deep scan a specific path for VMs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		color.Cyan("Deep scanning path: %s", args[0])
		if maxDepth > 0 {
			fmt.Printf("Max depth: %d\n", maxDepth)
		}
		fmt.Printf("Include hidden: %t\n", includeHidden)

		result, err := vmClient.DiscoverVMs(ctx, []string{args[0]})
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if jsonOutput {
			return outputJSON(result)
		}

		color.Cyan("Scan Results:")
		fmt.Printf("Scanned paths: %v\n", result.Paths)
		fmt.Printf("VMs found: %d\n", result.Count)

		if result.Count > 0 {
			fmt.Println("\nDetailed Results:")
			fmt.Printf("%-25s %-15s %-10s %-10s %-30s\n", "NAME", "STATE", "MEMORY", "CPUs", "CREATED")
			fmt.Println("--------------------------------------------------------------------------------")

			for _, vm := range result.VMs {
				stateColor := getStateColor(vm.State)
				fmt.Printf("%-25s %-15s %-10d %-10d %-30s\n", 
					vm.Name, 
					stateColor.Sprintf("%s", vm.State), 
					vm.Memory,
					vm.CPUs,
					vm.CreatedAt.Format("2006-01-02 15:04"))
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	// Add subcommands
	discoverCmd.AddCommand(discoverVMsCmd, discoverPathsCmd, discoverScanCmd)

	// Global discovery flags
	discoverCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")

	// Scan command flags
	discoverScanCmd.Flags().BoolVarP(&includeHidden, "hidden", "a", false, "include hidden files and directories")
	discoverScanCmd.Flags().IntVarP(&maxDepth, "depth", "d", 0, "maximum directory depth (0 = unlimited)")
}

// Helper functions
func getDefaultVMPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, "Virtual Machines"),
		filepath.Join(home, "Documents", "Virtual Machines"),
		"/Users/Shared/Virtual Machines",
		"/Library/Application Support/VMware Fusion/Virtual Machines",
		filepath.Join(home, "Library", "Application Support", "VMware Fusion", "Virtual Machines"),
	}
}