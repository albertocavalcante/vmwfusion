package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VM instance",
	RunE:  runCreateCmd,
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	// FIXED: Using cmd.Context() for proper cancellation and timeout handling
	ctx := cmd.Context()
	
	fmt.Println("Creating VM instance...")
	
	// Long-running VM creation that cannot be cancelled
	if err := createVMInstance(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}
	
	fmt.Println("VM instance created successfully")
	return nil
}

func createVMInstance(ctx context.Context, name string) error {
	// Simulate VM creation process
	select {
	case <-time.After(15 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
}