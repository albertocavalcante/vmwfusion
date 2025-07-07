package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"vmwfusion/internal/types"
	"vmwfusion/internal/vmware"
)

var (
	vmClient = vmware.NewClient()
	
	// VM command flags
	vmName     string
	vmOS       string
	vmMemory   int
	vmCPUs     int
	vmDiskSize int64
	jsonOutput bool
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Virtual machine operations",
	Long:  `Manage VMware Fusion virtual machines including create, delete, start, stop, and configuration operations.`,
}

var vmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all virtual machines",
	Long:  `Display a list of all virtual machines with their current status and basic information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		vms, err := vmClient.ListVMs(ctx)
		if err != nil {
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		if jsonOutput {
			return outputJSON(vms)
		}

		if len(vms) == 0 {
			color.Yellow("No virtual machines found")
			return nil
		}

		color.Cyan("Virtual Machines:")
		fmt.Printf("%-20s %-15s %-10s %-10s %-10s\n", "NAME", "STATE", "MEMORY", "CPUs", "OS")
		fmt.Println("--------------------------------------------------------------------------------")
		
		for _, vm := range vms {
			stateColor := getStateColor(vm.State)
			fmt.Printf("%-20s %-15s %-10d %-10d %-10s\n", 
				vm.Name, 
				stateColor.Sprintf("%s", vm.State), 
				vm.Memory, 
				vm.CPUs, 
				vm.OS)
		}
		return nil
	},
}

var vmCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		config := &types.VMConfig{
			Name:     args[0],
			OS:       vmOS,
			Memory:   vmMemory,
			CPUs:     vmCPUs,
			DiskSize: vmDiskSize,
		}

		vm, err := vmClient.CreateVM(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to create VM: %w", err)
		}

		color.Green("Successfully created VM: %s", vm.Name)
		return nil
	},
}

var vmDeleteCmd = &cobra.Command{
	Use:   "delete [vm-id]",
	Short: "Delete a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.DeleteVM(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to delete VM: %w", err)
		}

		color.Green("Successfully deleted VM: %s", args[0])
		return nil
	},
}

var vmStartCmd = &cobra.Command{
	Use:   "start [vm-id]",
	Short: "Start a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.StartVM(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to start VM: %w", err)
		}

		color.Green("Successfully started VM: %s", args[0])
		return nil
	},
}

var vmStopCmd = &cobra.Command{
	Use:   "stop [vm-id]",
	Short: "Stop a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.StopVM(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to stop VM: %w", err)
		}

		color.Green("Successfully stopped VM: %s", args[0])
		return nil
	},
}

var vmSuspendCmd = &cobra.Command{
	Use:   "suspend [vm-id]",
	Short: "Suspend a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.SuspendVM(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to suspend VM: %w", err)
		}

		color.Green("Successfully suspended VM: %s", args[0])
		return nil
	},
}

var vmResumeCmd = &cobra.Command{
	Use:   "resume [vm-id]",
	Short: "Resume a suspended virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		if err := vmClient.ResumeVM(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to resume VM: %w", err)
		}

		color.Green("Successfully resumed VM: %s", args[0])
		return nil
	},
}

var vmInfoCmd = &cobra.Command{
	Use:   "info [vm-id]",
	Short: "Show detailed information about a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		vm, err := vmClient.GetVM(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to get VM info: %w", err)
		}

		if jsonOutput {
			return outputJSON(vm)
		}

		color.Cyan("VM Information:")
		fmt.Printf("ID: %s\n", vm.ID)
		fmt.Printf("Name: %s\n", vm.Name)
		fmt.Printf("State: %s\n", getStateColor(vm.State).Sprintf("%s", vm.State))
		fmt.Printf("OS: %s\n", vm.OS)
		fmt.Printf("Memory: %d MB\n", vm.Memory)
		fmt.Printf("CPUs: %d\n", vm.CPUs)
		fmt.Printf("Disk Size: %d GB\n", vm.DiskSize)
		fmt.Printf("Path: %s\n", vm.Path)
		fmt.Printf("Created: %s\n", vm.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Modified: %s\n", vm.ModifiedAt.Format("2006-01-02 15:04:05"))

		if len(vm.Tags) > 0 {
			fmt.Printf("Tags: %v\n", vm.Tags)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vmCmd)

	// Add subcommands
	vmCmd.AddCommand(vmListCmd, vmCreateCmd, vmDeleteCmd, vmStartCmd, vmStopCmd, vmSuspendCmd, vmResumeCmd, vmInfoCmd)

	// Global VM flags
	vmCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")

	// Create command flags
	vmCreateCmd.Flags().StringVarP(&vmOS, "os", "o", "linux", "operating system type")
	vmCreateCmd.Flags().IntVarP(&vmMemory, "memory", "m", 2048, "memory in MB")
	vmCreateCmd.Flags().IntVarP(&vmCPUs, "cpus", "c", 2, "number of CPUs")
	vmCreateCmd.Flags().Int64VarP(&vmDiskSize, "disk", "d", 20, "disk size in GB")
}

// Helper functions
func getStateColor(state types.VMState) *color.Color {
	switch state {
	case types.VMStateRunning:
		return color.New(color.FgGreen)
	case types.VMStateStopped:
		return color.New(color.FgRed)
	case types.VMStateSuspended:
		return color.New(color.FgYellow)
	case types.VMStatePaused:
		return color.New(color.FgMagenta)
	default:
		return color.New(color.FgWhite)
	}
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}