package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage VM configuration",
	RunE:  runConfigCmd,
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	// FIXED: Using cmd.Context() for proper cancellation and timeout handling
	// This allows user cancellation (Ctrl+C) and respects command timeouts
	ctx := cmd.Context()
	
	// Simulate a long-running operation
	fmt.Println("Starting VM configuration update...")
	
	// This operation can now be cancelled by the user (Ctrl+C)
	if err := updateVMConfig(ctx); err != nil {
		return fmt.Errorf("failed to update VM config: %w", err)
	}
	
	fmt.Println("VM configuration updated successfully")
	return nil
}

func updateVMConfig(ctx context.Context) error {
	// Simulate long-running operation
	select {
	case <-time.After(10 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
}