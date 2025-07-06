package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ISODiscovery handles finding ISO files
type ISODiscovery struct {
	logger       *Logger
	searchDirs   []string
	excludePaths []string
}

// NewISODiscovery creates a new ISO discovery instance
func NewISODiscovery(logger *Logger) *ISODiscovery {
	homeDir, _ := os.UserHomeDir()
	
	return &ISODiscovery{
		logger: logger,
		searchDirs: []string{
			filepath.Join(homeDir, "Downloads"),
			filepath.Join(homeDir, "Desktop"),
			filepath.Join(homeDir, "Documents"),
			filepath.Join(homeDir, "Library"),
			filepath.Join(homeDir, "Virtual Machines"),
			filepath.Join(homeDir, "Documents", "Virtual Machines"),
			"/Applications/VMware Fusion.app/Contents/Library/isoimages",
		},
		excludePaths: []string{
			"/ISO_Archive/",
		},
	}
}

// FindAllISOs discovers all ISO files on the system
func (d *ISODiscovery) FindAllISOs() ([]ISOFile, error) {
	d.logger.Info.Println("🔍 Searching for ISO files on your system...")
	
	var allFiles []ISOFile
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// Search in predefined directories
	for _, dir := range d.searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		
		wg.Add(1)
		go func(searchDir string) {
			defer wg.Done()
			files := d.findInDirectory(searchDir)
			
			mu.Lock()
			allFiles = append(allFiles, files...)
			mu.Unlock()
		}(dir)
	}
	
	wg.Wait()
	
	// Use Spotlight for additional search
	d.logger.Info.Println("🔍 Using Spotlight search...")
	spotlightFiles := d.findWithSpotlight()
	
	// Merge and deduplicate
	allFiles = d.deduplicateFiles(append(allFiles, spotlightFiles...))
	
	// Calculate total size
	var totalSize int64
	for _, file := range allFiles {
		totalSize += file.Size
	}
	
	d.logger.Success.Printf("Found %d ISO files totaling %s\n", len(allFiles), FormatSizeSimple(totalSize))
	
	return allFiles, nil
}

// findInDirectory searches for ISO files in a specific directory
func (d *ISODiscovery) findInDirectory(dir string) []ISOFile {
	var files []ISOFile
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		
		// Skip if excluded path
		for _, exclude := range d.excludePaths {
			if strings.Contains(path, exclude) {
				return nil
			}
		}
		
		if strings.ToLower(filepath.Ext(path)) == ".iso" && info.Mode().IsRegular() {
			file := ISOFile{
				Path:          path,
				Size:          info.Size(),
				FormattedSize: FormatSizeSimple(info.Size()),
				Category:      string(GetCategory(path)),
			}
			files = append(files, file)
		}
		
		return nil
	})
	
	if err != nil {
		d.logger.Warning.Printf("Error searching %s: %v\n", dir, err)
	}
	
	return files
}

// findWithSpotlight uses macOS Spotlight to find additional ISO files
func (d *ISODiscovery) findWithSpotlight() []ISOFile {
	var files []ISOFile
	
	cmd := exec.Command("mdfind", "kMDItemDisplayName == '*.iso'")
	output, err := cmd.Output()
	if err != nil {
		d.logger.Warning.Printf("Spotlight search failed: %v\n", err)
		return files
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Skip if excluded path
		excluded := false
		for _, exclude := range d.excludePaths {
			if strings.Contains(line, exclude) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		
		if info, err := os.Stat(line); err == nil && info.Mode().IsRegular() {
			file := ISOFile{
				Path:          line,
				Size:          info.Size(),
				FormattedSize: FormatSizeSimple(info.Size()),
				Category:      string(CategorySpotlight),
			}
			files = append(files, file)
		}
	}
	
	return files
}

// deduplicateFiles removes duplicate files based on path
func (d *ISODiscovery) deduplicateFiles(files []ISOFile) []ISOFile {
	seen := make(map[string]bool)
	var unique []ISOFile
	
	for _, file := range files {
		if !seen[file.Path] {
			seen[file.Path] = true
			unique = append(unique, file)
		}
	}
	
	return unique
}

// CalculateHash calculates SHA256 hash of a file
func (d *ISODiscovery) CalculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// DisplayFiles shows the found files in a formatted table
func (d *ISODiscovery) DisplayFiles(files []ISOFile) {
	fmt.Printf("\n%-70s %-10s %-15s\n", "File Path", "Size", "Category")
	fmt.Printf("%-70s %-10s %-15s\n", strings.Repeat("-", 70), strings.Repeat("-", 10), strings.Repeat("-", 15))
	
	for _, file := range files {
		displayPath := TruncatePath(file.Path, 70)
		fmt.Printf("%-70s %-10s %-15s\n", displayPath, file.FormattedSize, file.Category)
	}
	
	// Show breakdown by category
	fmt.Println()
	d.logger.Info.Println("📊 Breakdown by category:")
	
	categoryStats := make(map[string]int64)
	for _, file := range files {
		categoryStats[file.Category] += file.Size
	}
	
	for category, size := range categoryStats {
		fmt.Printf("  • %s: %s\n", category, FormatSizeSimple(size))
	}
	
	// Key findings
	fmt.Println()
	d.logger.Info.Println("💡 Key findings:")
	if vmwareSize, exists := categoryStats["VMware"]; exists {
		fmt.Printf("  • VMware built-in ISOs: %s\n", FormatSizeSimple(vmwareSize))
	}
	
	var exportableSize int64
	for category, size := range categoryStats {
		if category != "VMware" {
			exportableSize += size
		}
	}
	if exportableSize > 0 {
		fmt.Printf("  • %s can be archived to free space\n", FormatSizeSimple(exportableSize))
	}
}