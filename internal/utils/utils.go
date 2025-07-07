package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if !DirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration into human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-d.Minutes()*60)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - hours*60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// TruncateString truncates a string to maxLen with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ValidateVMName validates a VM name according to common rules
func ValidateVMName(name string) error {
	if name == "" {
		return fmt.Errorf("VM name cannot be empty")
	}
	
	if len(name) > 255 {
		return fmt.Errorf("VM name too long (max 255 characters)")
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("VM name contains invalid character: %s", char)
		}
	}
	
	return nil
}

// SanitizeFilename removes invalid characters from filename
func SanitizeFilename(filename string) string {
	// Replace invalid characters with underscore
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := reg.ReplaceAllString(filename, "_")
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	return sanitized
}

// GetVMXFiles finds all .vmx files in a directory
func GetVMXFiles(dir string) ([]string, error) {
	var vmxFiles []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".vmx") {
			vmxFiles = append(vmxFiles, path)
		}
		
		return nil
	})
	
	return vmxFiles, err
}

// ParseMemoryString parses memory strings like "2GB", "1024MB", "512"
func ParseMemoryString(memStr string) (int, error) {
	memStr = strings.ToUpper(strings.TrimSpace(memStr))
	
	if memStr == "" {
		return 0, fmt.Errorf("empty memory string")
	}
	
	// If it's just a number, assume MB
	if matched, _ := regexp.MatchString(`^\d+$`, memStr); matched {
		var mem int
		_, err := fmt.Sscanf(memStr, "%d", &mem)
		return mem, err
	}
	
	// Parse with units
	var mem int
	var unit string
	n, err := fmt.Sscanf(memStr, "%d%s", &mem, &unit)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid memory format: %s", memStr)
	}
	
	switch unit {
	case "B", "BYTES":
		return mem / (1024 * 1024), nil
	case "KB", "K":
		return mem / 1024, nil
	case "MB", "M":
		return mem, nil
	case "GB", "G":
		return mem * 1024, nil
	case "TB", "T":
		return mem * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown memory unit: %s", unit)
	}
}

// GenerateVMID generates a unique VM identifier
func GenerateVMID() string {
	return fmt.Sprintf("vm-%d", time.Now().UnixNano())
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// ExpandPath expands ~ to home directory in paths
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(GetHomeDir(), path[2:])
	}
	return path
}

// FindExecutable searches for an executable in PATH
func FindExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("executable %s not found in PATH", name)
	}
	return path, nil
}

// GetAvailablePort finds an available port starting from the given port
func GetAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+1000; port++ {
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from %d", startPort)
}

func isPortAvailable(port int) bool {
	// Simple port availability check - in a real implementation,
	// you'd want to actually try to bind to the port
	return true
}