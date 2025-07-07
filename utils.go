package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// Logger provides colored output
type Logger struct {
	Info    *color.Color
	Success *color.Color
	Warning *color.Color
	Error   *color.Color
}

// NewLogger creates a new logger with colored output
func NewLogger() *Logger {
	return &Logger{
		Info:    color.New(color.FgCyan),
		Success: color.New(color.FgGreen),
		Warning: color.New(color.FgYellow),
		Error:   color.New(color.FgRed),
	}
}

// FormatSize formats bytes into human-readable size
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSizeSimple formats bytes into simple GB/MB format (matches shell script)
func FormatSizeSimple(bytes int64) string {
	if bytes > 1073741824 {
		return fmt.Sprintf("%dGB", bytes/1073741824)
	} else if bytes > 1048576 {
		return fmt.Sprintf("%dMB", bytes/1048576)
	} else {
		return fmt.Sprintf("%dKB", bytes/1024)
	}
}

// GetCategory determines the category of an ISO file based on its path
func GetCategory(path string) FileCategory {
	baseName := strings.ToLower(filepath.Base(path))
	
	if strings.Contains(path, "VMware Fusion") {
		return CategoryVMware
	}
	
	if strings.Contains(baseName, "windows") || strings.Contains(baseName, "win") {
		return CategoryWindows
	}
	
	return CategoryUser
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetVMwareArchSuffix returns architecture suffix for VMware ISOs
func GetVMwareArchSuffix(path string) string {
	if strings.Contains(path, "/arm64/") {
		return "_arm64"
	} else if strings.Contains(path, "/x86_x64/") {
		return "_x64"
	}
	return ""
}

// TruncatePath truncates a path for display purposes
func TruncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// ParseSize parses a size string like "2GB" into bytes
func ParseSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) < 2 {
		return 0
	}
	
	unit := strings.ToUpper(sizeStr[len(sizeStr)-2:])
	valueStr := sizeStr[:len(sizeStr)-2]
	
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}
	
	switch unit {
	case "GB":
		return int64(value * 1073741824)
	case "MB":
		return int64(value * 1048576)
	case "KB":
		return int64(value * 1024)
	default:
		return 0
	}
}