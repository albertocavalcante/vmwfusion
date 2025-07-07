package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VMwareTools holds paths to VMware utilities
type VMwareTools struct {
	VMRunPath string
	VMCliPath string
	BasePath  string
}

// Config holds global configuration
type Config struct {
	LocalVMDir        string   `json:"local_vm_dir"`
	VMwarePath        string   `json:"vmware_path"`
	ExFATMountBase    string   `json:"exfat_mount_base"`
	DefaultDrive      string   `json:"default_archive_drive"`
	RsyncOptions      []string `json:"rsync_options"`
	ShutdownTimeout   int      `json:"shutdown_timeout"`
	CleanupOriginal   bool     `json:"cleanup_original_after_archive"`
	AutoOptimize      bool     `json:"auto_optimize_for_archive"`
	VerifyChecksums   bool     `json:"verify_checksums"`
	PreferredCloneType string  `json:"preferred_clone_type"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		LocalVMDir:         filepath.Join(homeDir, "Virtual Machines"),
		VMwarePath:         "/Applications/VMware Fusion.app/Contents/Library",
		ExFATMountBase:     "/Volumes",
		DefaultDrive:       "T9",
		RsyncOptions:       []string{"-avhc", "--progress", "--delete"},
		ShutdownTimeout:    120,
		CleanupOriginal:    false,
		AutoOptimize:       true,
		VerifyChecksums:    true,
		PreferredCloneType: "linked",
	}
}

// LoadConfig loads configuration from file with fallback to defaults
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()
	
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".vmwfusion-config.json")
	}
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil // Return defaults if no config file
	}
	
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return config, err
	}
	
	return config, nil
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(configPath string) error {
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".vmwfusion-config.json")
	}
	
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}

// DetectVMwareTools detects and validates VMware Fusion installation
func DetectVMwareTools(vmwarePath string) (*VMwareTools, error) {
	if vmwarePath == "" {
		vmwarePath = "/Applications/VMware Fusion.app/Contents/Library"
	}
	
	tools := &VMwareTools{
		BasePath:  vmwarePath,
		VMRunPath: filepath.Join(vmwarePath, "vmrun"),
		VMCliPath: filepath.Join(vmwarePath, "vmcli"),
	}
	
	// Check if VMware Fusion is installed by validating the configured path
	// The vmwarePath typically points to /Applications/VMware Fusion.app/Contents/Library
	// so we need to go up two levels to get to the .app bundle
	vmwareFusionApp := filepath.Join(vmwarePath, "..", "..")
	if _, err := os.Stat(vmwareFusionApp); os.IsNotExist(err) {
		return nil, fmt.Errorf("VMware Fusion not found at %s - please check your VMware installation", vmwareFusionApp)
	}
	
	// Check if vmrun exists (required)
	if _, err := os.Stat(tools.VMRunPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vmrun not found at %s", tools.VMRunPath)
	}
	
	// vmcli is optional but preferred for advanced operations
	if _, err := os.Stat(tools.VMCliPath); os.IsNotExist(err) {
		// This is a warning, not an error
	}
	
	return tools, nil
}

// CheckProFeatures checks if VMware Fusion Pro features are available
func (vt *VMwareTools) CheckProFeatures() error {
	cmd := exec.Command(vt.VMRunPath, "clone")
	output, err := cmd.CombinedOutput()
	
	if err != nil || !strings.Contains(string(output), "Create a copy") {
		return fmt.Errorf("VMware Fusion Pro required for cloning features - standard Fusion does not support vmrun clone")
	}
	
	return nil
}

// IsVMRunning checks if a VM is currently running
func (vt *VMwareTools) IsVMRunning(vmxPath string) (bool, error) {
	cmd := exec.Command(vt.VMRunPath, "list")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	
	return strings.Contains(string(output), vmxPath), nil
}

// ListRunningVMs returns a list of currently running VMs
func (vt *VMwareTools) ListRunningVMs() ([]string, error) {
	cmd := exec.Command(vt.VMRunPath, "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var runningVMs []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Total running VMs:") {
			runningVMs = append(runningVMs, line)
		}
	}
	
	return runningVMs, nil
}

// GlobalLogger provides a shared logger instance
var GlobalLogger *Logger

// InitializeLogger sets up the global logger
func InitializeLogger() {
	GlobalLogger = NewLogger()
}

// VMBundle represents a VMware VM bundle
type VMBundle struct {
	Path    string
	Name    string
	VMXPath string
	Size    int64
}

// FindVMBundle searches for a VM bundle by name
func FindVMBundle(vmName, searchDir string) (*VMBundle, error) {
	if searchDir == "" {
		homeDir, _ := os.UserHomeDir()
		searchDir = filepath.Join(homeDir, "Virtual Machines")
	}
	
	var foundPath string
	
	// Strategy 1: Direct pattern matching
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		
		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(vmName)) && strings.HasSuffix(info.Name(), ".vmwarevm") {
			foundPath = path
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if foundPath == "" {
		// Strategy 2: Use Spotlight search
		cmd := exec.Command("mdfind", fmt.Sprintf("kMDItemFSName == '*%s*.vmwarevm'", vmName))
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" && strings.Contains(strings.ToLower(line), strings.ToLower(vmName)) {
					foundPath = line
					break
				}
			}
		}
	}
	
	if foundPath == "" {
		return nil, fmt.Errorf("VM bundle not found: %s", vmName)
	}
	
	// Find VMX file
	vmxPath, err := findVMXFile(foundPath)
	if err != nil {
		return nil, fmt.Errorf("no VMX file found in bundle %s: %w", foundPath, err)
	}
	
	// Get bundle size
	size, err := getDirSize(foundPath)
	if err != nil {
		size = 0 // Non-fatal
	}
	
	return &VMBundle{
		Path:    foundPath,
		Name:    strings.TrimSuffix(filepath.Base(foundPath), ".vmwarevm"),
		VMXPath: vmxPath,
		Size:    size,
	}, nil
}

// findVMXFile finds the VMX file within a VM bundle
func findVMXFile(bundlePath string) (string, error) {
	var vmxPath string
	
	err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".vmx") {
			vmxPath = path
			return filepath.SkipDir
		}
		
		return nil
	})
	
	if err != nil {
		return "", err
	}
	
	if vmxPath == "" {
		return "", fmt.Errorf("no VMX file found")
	}
	
	return vmxPath, nil
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

// ListExternalDrives lists available external drives
func ListExternalDrives(mountBase string) ([]ExternalDrive, error) {
	if mountBase == "" {
		mountBase = "/Volumes"
	}
	
	var drives []ExternalDrive
	
	entries, err := os.ReadDir(mountBase)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "Macintosh HD" {
			drivePath := filepath.Join(mountBase, entry.Name())
			
			// Get filesystem info
			cmd := exec.Command("df", "-h", drivePath)
			output, err := cmd.Output()
			if err != nil {
				continue
			}
			
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 4 {
					drive := ExternalDrive{
						Name:      entry.Name(),
						Path:      drivePath,
						Size:      fields[1],
						Used:      fields[2],
						Available: fields[3],
					}
					
					// Get filesystem type
					cmd = exec.Command("diskutil", "info", drivePath)
					if output, err := cmd.Output(); err == nil {
						if strings.Contains(string(output), "ExFAT") {
							drive.FileSystem = "ExFAT"
						} else if strings.Contains(string(output), "APFS") {
							drive.FileSystem = "APFS"
						} else if strings.Contains(string(output), "HFS") {
							drive.FileSystem = "HFS+"
						}
					}
					
					drives = append(drives, drive)
				}
			}
		}
	}
	
	return drives, nil
}

// ExternalDrive represents an external drive
type ExternalDrive struct {
	Name       string
	Path       string
	Size       string
	Used       string
	Available  string
	FileSystem string
}