package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate VM to a different host",
	RunE:  runMigrateCmd,
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {
	// FIXED: Using cmd.Context() for proper cancellation and timeout handling
	ctx := cmd.Context()
	
	fmt.Println("Starting VM migration...")
	
	// Migration is a very long-running operation that should be cancellable
	if err := migrateVM(ctx, args[0], args[1]); err != nil {
		return fmt.Errorf("failed to migrate VM: %w", err)
	}
	
	fmt.Println("VM migration completed successfully")
	return nil
}

func migrateVM(ctx context.Context, vmName, targetHost string) error {
	// Simulate long migration process
	select {
	case <-time.After(30 * time.Second): // Very long operation
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}