package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ISOExporter handles exporting ISO files
type ISOExporter struct {
	logger    *Logger
	discovery *ISODiscovery
	config    ExportConfig
}

// NewISOExporter creates a new ISO exporter
func NewISOExporter(logger *Logger, discovery *ISODiscovery, config ExportConfig) *ISOExporter {
	return &ISOExporter{
		logger:    logger,
		discovery: discovery,
		config:    config,
	}
}

// ExportISOs exports ISO files to the target drive
func (e *ISOExporter) ExportISOs(files []ISOFile) (*ExportResult, error) {
	targetPath := filepath.Join("/Volumes", e.config.TargetDrive)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("drive not found: %s", targetPath)
	}
	
	archiveDir := filepath.Join(targetPath, e.config.ArchiveDir)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create archive directory: %w", err)
	}
	
	e.logger.Info.Printf("📦 Exporting ISOs to %s\n", e.config.TargetDrive)
	e.logger.Info.Printf("Target: %s\n", archiveDir)
	fmt.Println()
	
	// Load existing manifest
	manifest, err := e.loadManifest(archiveDir)
	if err != nil {
		e.logger.Warning.Printf("Could not load manifest: %v, creating new one\n", err)
		manifest = &Manifest{
			CreatedAt: time.Now(),
			Version:   "1.0",
			Exports:   make(map[string]ISOFile),
		}
	}
	
	result := &ExportResult{}
	
	for _, file := range files {
		// Skip VMware files unless explicitly included
		if file.Category == string(CategoryVMware) && !e.config.IncludeVMware {
			e.logger.Info.Printf("⏭️  Skipping VMware: %s\n", filepath.Base(file.Path))
			result.FilesSkipped++
			continue
		}
		
		exported, err := e.exportSingleFile(file, archiveDir, manifest)
		if err != nil {
			e.logger.Error.Printf("❌ Failed to export %s: %v\n", filepath.Base(file.Path), err)
			result.Errors = append(result.Errors, err)
			continue
		}
		
		if exported {
			result.FilesExported++
			result.TotalSize += file.Size
		} else {
			result.FilesSkipped++
		}
	}
	
	// Save updated manifest
	manifest.UpdatedAt = time.Now()
	if err := e.saveManifest(manifest, archiveDir); err != nil {
		e.logger.Warning.Printf("Failed to save manifest: %v\n", err)
	}
	
	// Display results
	fmt.Println()
	e.logger.Success.Println("📊 Export completed!")
	fmt.Printf("  • Files exported: %d\n", result.FilesExported)
	fmt.Printf("  • Files skipped: %d\n", result.FilesSkipped)
	fmt.Printf("  • Total exported: %s\n", FormatSizeSimple(result.TotalSize))
	fmt.Printf("  • Location: %s\n", archiveDir)
	
	return result, nil
}

// exportSingleFile exports a single ISO file
func (e *ISOExporter) exportSingleFile(file ISOFile, archiveDir string, manifest *Manifest) (bool, error) {
	// Calculate hash for verification
	hash, err := e.discovery.CalculateHash(file.Path)
	if err != nil {
		return false, fmt.Errorf("failed to calculate hash: %w", err)
	}
	
	// Check manifest first for fast verification
	if existingFile, exists := manifest.Exports[file.Path]; exists {
		if existingFile.Hash == hash {
			e.logger.Success.Printf("✅ Already exported (verified): %s → %s\n", 
				filepath.Base(file.Path), existingFile.DestFileName)
			return false, nil // Not newly exported, but verified
		} else {
			e.logger.Warning.Printf("⚠️  File changed since last export: %s\n", filepath.Base(file.Path))
		}
	}
	
	// Generate destination filename
	destFileName := e.generateDestFileName(file)
	destPath := filepath.Join(archiveDir, destFileName)
	
	// Check if destination file exists
	if FileExists(destPath) {
		e.logger.Info.Printf("📄 Found existing: %s\n", destFileName)
		
		// Compare hashes
		if destHash, err := e.discovery.CalculateHash(destPath); err == nil {
			if hash == destHash {
				e.logger.Success.Printf("✅ Already exported (identical): %s\n", destFileName)
				// Update manifest
				e.updateManifestEntry(manifest, file, destFileName, hash)
				return false, nil
			}
		}
		
		// File exists but differs
		e.logger.Warning.Printf("⚠️  File exists but differs: %s\n", destFileName)
		choice := e.promptOverwrite()
		
		switch choice {
		case "y":
			e.logger.Info.Printf("📄 Overwriting: %s (%s)\n", destFileName, file.FormattedSize)
		case "r":
			timestamp := time.Now().Format("20060102_150405")
			name := strings.TrimSuffix(destFileName, filepath.Ext(destFileName))
			ext := filepath.Ext(destFileName)
			destFileName = fmt.Sprintf("%s_%s%s", name, timestamp, ext)
			destPath = filepath.Join(archiveDir, destFileName)
			e.logger.Info.Printf("📄 Copying as: %s (%s)\n", destFileName, file.FormattedSize)
		default:
			e.logger.Info.Printf("⏭️  Skipping: %s\n", destFileName)
			return false, nil
		}
	} else {
		e.logger.Info.Printf("📄 Copying: %s (%s)\n", destFileName, file.FormattedSize)
	}
	
	// Copy the file
	if err := e.copyFile(file.Path, destPath); err != nil {
		return false, fmt.Errorf("copy failed: %w", err)
	}
	
	// Verify copy
	if err := e.verifyFile(file.Path, destPath); err != nil {
		os.Remove(destPath) // Clean up failed copy
		return false, fmt.Errorf("verification failed: %w", err)
	}
	
	e.logger.Success.Printf("✅ Exported: %s\n", destFileName)
	
	// Update manifest
	e.updateManifestEntry(manifest, file, destFileName, hash)
	
	return true, nil
}

// generateDestFileName generates the destination filename, handling VMware arch suffixes
func (e *ISOExporter) generateDestFileName(file ISOFile) string {
	fileName := filepath.Base(file.Path)
	
	if file.Category == string(CategoryVMware) {
		suffix := GetVMwareArchSuffix(file.Path)
		if suffix != "" {
			name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			ext := filepath.Ext(fileName)
			fileName = fmt.Sprintf("%s%s%s", name, suffix, ext)
		}
	}
	
	return fileName
}

// promptOverwrite prompts user for overwrite decision
func (e *ISOExporter) promptOverwrite() string {
	fmt.Print("Overwrite existing file? (y/N/r=rename): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(input))
}

// copyFile copies a file from source to destination
func (e *ISOExporter) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	return destFile.Sync()
}

// verifyFile verifies that the copy was successful by comparing sizes
func (e *ISOExporter) verifyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return err
	}
	
	if srcInfo.Size() != dstInfo.Size() {
		return fmt.Errorf("size mismatch: source=%d, destination=%d", srcInfo.Size(), dstInfo.Size())
	}
	
	return nil
}

// loadManifest loads the manifest file
func (e *ISOExporter) loadManifest(archiveDir string) (*Manifest, error) {
	manifestPath := filepath.Join(archiveDir, e.config.ManifestFile)
	
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var manifest Manifest
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, err
	}
	
	return &manifest, nil
}

// saveManifest saves the manifest file
func (e *ISOExporter) saveManifest(manifest *Manifest, archiveDir string) error {
	manifestPath := filepath.Join(archiveDir, e.config.ManifestFile)
	
	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

// updateManifestEntry updates an entry in the manifest
func (e *ISOExporter) updateManifestEntry(manifest *Manifest, file ISOFile, destFileName, hash string) {
	manifest.Exports[file.Path] = ISOFile{
		Path:          file.Path,
		Size:          file.Size,
		FormattedSize: file.FormattedSize,
		Hash:          hash,
		Category:      file.Category,
		Exported:      true,
		ExportedAt:    time.Now(),
		DestFileName:  destFileName,
	}
}