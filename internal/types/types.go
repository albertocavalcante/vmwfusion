package types

import "time"

// VM represents a virtual machine
type VM struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	State       VMState          `json:"state"`
	OS          string           `json:"os"`
	Memory      int              `json:"memory_mb"`
	CPUs        int              `json:"cpus"`
	DiskSize    int64            `json:"disk_size_gb"`
	Network     []NetworkConfig  `json:"network"`
	CreatedAt   time.Time        `json:"created_at"`
	ModifiedAt  time.Time        `json:"modified_at"`
	Tags        []string         `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// VMState represents the current state of a VM
type VMState string

const (
	VMStateRunning   VMState = "running"
	VMStateStopped   VMState = "stopped"
	VMStateSuspended VMState = "suspended"
	VMStatePaused    VMState = "paused"
	VMStateUnknown   VMState = "unknown"
)

// NetworkConfig represents network configuration for a VM
type NetworkConfig struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	Connected  bool   `json:"connected"`
}

// VMConfig holds configuration for VM operations
type VMConfig struct {
	Name        string            `json:"name"`
	OS          string           `json:"os"`
	Memory      int              `json:"memory_mb"`
	CPUs        int              `json:"cpus"`
	DiskSize    int64            `json:"disk_size_gb"`
	ISOPath     string           `json:"iso_path"`
	Network     []NetworkConfig  `json:"network"`
	StartOnBoot bool             `json:"start_on_boot"`
	Metadata    map[string]string `json:"metadata"`
}

// MigrationOptions holds options for VM migration
type MigrationOptions struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	ConvertFormat   string `json:"convert_format"`
	Compress        bool   `json:"compress"`
	KeepOriginal    bool   `json:"keep_original"`
}

// ISOInfo represents information about an ISO file
type ISOInfo struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Name         string    `json:"name"`
	OS           string    `json:"os"`
	Architecture string    `json:"architecture"`
	CreatedAt    time.Time `json:"created_at"`
	Checksum     string    `json:"checksum"`
}

// ExportOptions holds options for VM export
type ExportOptions struct {
	Format      string `json:"format"`
	Destination string `json:"destination"`
	Compress    bool   `json:"compress"`
	IncludeISO  bool   `json:"include_iso"`
}

// DiscoveryResult represents discovered VMs
type DiscoveryResult struct {
	VMs   []VM   `json:"vms"`
	Count int    `json:"count"`
	Paths []string `json:"search_paths"`
}