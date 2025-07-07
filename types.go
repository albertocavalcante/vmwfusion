package main

import (
	"time"
)

// ISOFile represents an ISO file found on the system
type ISOFile struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	FormattedSize string   `json:"formatted_size"`
	Hash         string    `json:"hash,omitempty"`
	Category     string    `json:"category"`
	Exported     bool      `json:"exported"`
	ExportedAt   time.Time `json:"exported_at,omitempty"`
	DestFileName string    `json:"dest_filename,omitempty"`
}

// Manifest represents the export manifest file
type Manifest struct {
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Version   string             `json:"version"`
	Exports   map[string]ISOFile `json:"exports"` // key is source path
}

// ExportConfig holds configuration for export operations
type ExportConfig struct {
	TargetDrive    string
	ArchiveDir     string
	ManifestFile   string
	IncludeVMware  bool
	ForceOverwrite bool
}

// FileCategory represents the type of ISO file
type FileCategory string

const (
	CategoryVMware    FileCategory = "VMware"
	CategoryWindows   FileCategory = "Windows"
	CategoryUser      FileCategory = "User"
	CategorySpotlight FileCategory = "Spotlight"
)

// ExportResult represents the result of an export operation
type ExportResult struct {
	FilesExported int
	FilesSkipped  int
	TotalSize     int64
	Errors        []error
}