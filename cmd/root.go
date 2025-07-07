package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vmctl",
	Short: "VM management CLI tool",
	Long:  "A command-line tool for managing virtual machines with proper cancellation support",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Set up context with signal handling for proper cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	
	// Set the context on the root command so child commands can access it
	return rootCmd.ExecuteContext(ctx)
}