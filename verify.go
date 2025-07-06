package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ISOVerifier handles verification and deletion of original files
type ISOVerifier struct {
	logger    *Logger
	discovery *ISODiscovery
	config    ExportConfig
}

// NewISOVerifier creates a new ISO verifier
func NewISOVerifier(logger *Logger, discovery *ISODiscovery, config ExportConfig) *ISOVerifier {
	return &ISOVerifier{
		logger:    logger,
		discovery: discovery,
		config:    config,
	}
}

// VerifyAndDelete verifies exports using cached manifest and prompts for deletion
func (v *ISOVerifier) VerifyAndDelete() error {
	v.logger.Info.Printf("🔍 Quick Verify & Delete Workflow (using cached paths)\n")
	v.logger.Info.Printf("Target drive: %s\n", v.config.TargetDrive)
	fmt.Println()
	
	targetPath := filepath.Join("/Volumes", v.config.TargetDrive)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("drive not found: %s", targetPath)
	}
	
	archiveDir := filepath.Join(targetPath, v.config.ArchiveDir)
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		v.logger.Warning.Println("No archive found. Run export first.")
		return nil
	}
	
	// Load manifest
	exporter := NewISOExporter(v.logger, v.discovery, v.config)
	manifest, err := exporter.loadManifest(archiveDir)
	if err != nil || len(manifest.Exports) == 0 {
		v.logger.Warning.Println("No manifest entries found. Run quick-export first.")
		return nil
	}
	
	v.logger.Info.Println("🔍 Verifying cached exports and prompting for deletion...")
	fmt.Println()
	
	// Display verification results
	verifiedFiles, err := v.displayVerificationResults(manifest)
	if err != nil {
		return err
	}
	
	if len(verifiedFiles) == 0 {
		v.logger.Warning.Println("No original files found from manifest")
		return nil
	}
	
	fmt.Println()
	v.logger.Info.Println("🗑️  Ready to prompt for deletion of originals...")
	fmt.Println()
	fmt.Print("Proceed with deletion prompts? (y/N): ")
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	
	if input := strings.TrimSpace(strings.ToLower(input)); input != "y" && input != "yes" {
		v.logger.Info.Println("Deletion cancelled")
		return nil
	}
	
	return v.promptDeleteOriginals(verifiedFiles)
}

// displayVerificationResults shows verification status of cached files
func (v *ISOVerifier) displayVerificationResults(manifest *Manifest) ([]ISOFile, error) {
	fmt.Printf("%-70s %-10s %-15s\n", "File Path", "Size", "Status")
	fmt.Printf("%-70s %-10s %-15s\n", strings.Repeat("-", 70), strings.Repeat("-", 10), strings.Repeat("-", 15))
	
	var verifiedFiles []ISOFile
	var totalSize int64
	var foundCount, missingCount int
	
	for _, exportedFile := range manifest.Exports {
		if _, err := os.Stat(exportedFile.Path); os.IsNotExist(err) {
			missingCount++
			displayPath := TruncatePath(exportedFile.Path, 70)
			fmt.Printf("%-70s %-10s %-15s\n", displayPath, "--", "Missing")
			continue
		}
		
		// File still exists, verify it
		currentHash, err := v.discovery.CalculateHash(exportedFile.Path)
		if err != nil {
			v.logger.Warning.Printf("Could not hash %s: %v\n", exportedFile.Path, err)
			continue
		}
		
		status := "Verified"
		if currentHash != exportedFile.Hash {
			status = "Changed"
		}
		
		displayPath := TruncatePath(exportedFile.Path, 70)
		fmt.Printf("%-70s %-10s %-15s\n", displayPath, exportedFile.FormattedSize, status)
		
		verifiedFiles = append(verifiedFiles, exportedFile)
		totalSize += exportedFile.Size
		foundCount++
	}
	
	fmt.Println()
	v.logger.Success.Printf("Found %d files from manifest (%s)\n", foundCount, FormatSizeSimple(totalSize))
	if missingCount > 0 {
		v.logger.Info.Printf("%d files no longer exist\n", missingCount)
	}
	
	return verifiedFiles, nil
}

// promptDeleteOriginals prompts user to delete each original file
func (v *ISOVerifier) promptDeleteOriginals(files []ISOFile) error {
	v.logger.Info.Println("🗑️  Delete original files to free disk space?")
	fmt.Println()
	
	deletedCount := 0
	keptCount := 0
	var deletedSize int64
	
	for _, file := range files {
		// Skip VMware files - they should never be deleted
		if file.Category == string(CategoryVMware) {
			v.logger.Info.Printf("⏭️  Skipping VMware file: %s\n", filepath.Base(file.Path))
			continue
		}
		
		fileName := filepath.Base(file.Path)
		fmt.Printf("Delete %s (%s)? (y/N): ", fileName, file.FormattedSize)
		
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			v.logger.Warning.Printf("Failed to read input: %v\n", err)
			continue
		}
		choice := strings.TrimSpace(strings.ToLower(input))
		
		if choice == "y" || choice == "yes" {
			if err := os.Remove(file.Path); err != nil {
				v.logger.Error.Printf("❌ Failed to delete: %s - %v\n", fileName, err)
			} else {
				v.logger.Success.Printf("✅ Deleted: %s\n", fileName)
				deletedCount++
				deletedSize += file.Size
			}
		} else {
			v.logger.Info.Printf("⏭️  Kept: %s\n", fileName)
			keptCount++
		}
	}
	
	fmt.Println()
	v.logger.Info.Println("🧹 Cleanup completed!")
	fmt.Printf("  • Files deleted: %d (%s freed)\n", deletedCount, FormatSizeSimple(deletedSize))
	fmt.Printf("  • Files kept: %d\n", keptCount)
	
	return nil
}

// ShowArchived displays what's currently archived
func (v *ISOVerifier) ShowArchived() error {
	targetPath := filepath.Join("/Volumes", v.config.TargetDrive)
	archiveDir := filepath.Join(targetPath, v.config.ArchiveDir)
	
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		v.logger.Warning.Printf("No archive found at: %s\n", archiveDir)
		return nil
	}
	
	v.logger.Info.Printf("📁 Archived ISOs on %s:\n", v.config.TargetDrive)
	fmt.Println()
	
	var totalSize int64
	var count int
	
	fmt.Printf("%-50s %-10s\n", "Filename", "Size")
	fmt.Printf("%-50s %-10s\n", strings.Repeat("-", 50), strings.Repeat("-", 10))
	
	err := filepath.Walk(archiveDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if strings.ToLower(filepath.Ext(path)) == ".iso" && info.Mode().IsRegular() {
			fileName := filepath.Base(path)
			formattedSize := FormatSizeSimple(info.Size())
			
			// Truncate filename if too long
			if len(fileName) > 50 {
				fileName = fileName[:47] + "..."
			}
			
			fmt.Printf("%-50s %-10s\n", fileName, formattedSize)
			totalSize += info.Size()
			count++
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("error reading archive: %w", err)
	}
	
	fmt.Println()
	v.logger.Success.Printf("Total: %d files, %s\n", count, FormatSizeSimple(totalSize))
	v.logger.Info.Printf("Location: %s\n", archiveDir)
	
	return nil
}