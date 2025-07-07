# VMware Fusion CLI (`vmwfusion`)

A unified CLI tool for VMware Fusion that provides comprehensive VM and ISO management with modern Go architecture.

## Features

### 🎯 **Core Capabilities**
- **ISO Management**: Discovery, export, and cleanup with hash-based deduplication
- **VM Lifecycle**: Power management, cloning, optimization, and detailed information
- **VM Migration**: Archive/restore between local SSD and ExFAT storage with verification
- **Fast Deployment**: Linked clone creation for rapid development workflows
- **Configuration Management**: JSON-based configuration with validation

### 🚀 **Key Improvements Over Shell Scripts**
- **Performance**: Concurrent operations, parallel file processing
- **Reliability**: Type safety, comprehensive error handling, graceful recovery
- **User Experience**: Consistent CLI patterns, colored output, helpful status commands
- **Maintainability**: Clean modular architecture, single binary deployment

## Installation

### Prerequisites
- macOS with VMware Fusion installed
- Go 1.21+ (for building from source)

### Build from Source
```bash
git clone https://github.com/albertocavalcante/vmwfusion.git
cd vmwfusion
go build -o vmwfusion
sudo mv vmwfusion /usr/local/bin/
```

### Quick Start
```bash
# Create configuration file
vmwfusion config create

# Check system status
vmwfusion info status

# Find and manage ISO files
vmwfusion iso find
vmwfusion iso quick-export

# Manage VMs
vmwfusion vm list
vmwfusion vm info "My VM"
vmwfusion vm power "My VM" stop

# Archive/restore VMs
vmwfusion migrate archive "My VM"
vmwfusion migrate restore "My VM" T9
```

## Command Structure

```
vmwfusion
├── iso         # ISO file management
│   ├── find          # Discover ISO files
│   ├── export        # Export to external drive  
│   ├── quick-export  # Find and export workflow
│   ├── verify-delete # Fast verify using cached paths
│   └── show-archived # Show archived ISOs
├── vm          # VM lifecycle management
│   ├── list          # List VMs with status
│   ├── info          # Detailed VM information
│   ├── power         # Start/stop/reset/suspend
│   ├── clone         # Create VM clones (Pro required)
│   └── optimize      # Optimize for archival
├── migrate     # VM migration operations
│   ├── archive       # Archive VM to ExFAT
│   ├── restore       # Restore from archive
│   ├── fast-deploy   # Linked clone deployment
│   └── list-drives   # Show available drives
├── config      # Configuration management
│   ├── show          # Display current config
│   ├── create        # Create sample config
│   ├── validate      # Validate configuration
│   └── reset         # Reset to defaults
└── info        # System information
    ├── status        # Overall system status
    ├── system        # macOS system info
    ├── vmware        # VMware Fusion details
    └── drives        # External drive info
```

## Configuration

VMware Fusion CLI uses JSON configuration at `~/.vmwfusion-config.json`:

```json
{
  "local_vm_dir": "/Users/username/Virtual Machines",
  "vmware_path": "/Applications/VMware Fusion.app/Contents/Library",
  "exfat_mount_base": "/Volumes",
  "default_archive_drive": "T9",
  "rsync_options": ["-avhc", "--progress", "--delete"],
  "preferred_clone_type": "linked",
  "auto_optimize_for_archive": true,
  "verify_checksums": true,
  "shutdown_timeout": 120,
  "cleanup_original_after_archive": false
}
```

## Examples

### ISO Management
```bash
# Find all ISO files with categorization
vmwfusion iso find

# Export ISOs to T9 drive with hash verification
vmwfusion iso quick-export --drive T9

# Fast verify and delete originals (no re-export)
vmwfusion iso verify-delete
```

### VM Management
```bash
# List all VMs with power status
vmwfusion vm list

# Get detailed VM information
vmwfusion vm info "Windows 11"

# Clone VM (requires VMware Fusion Pro)
vmwfusion vm clone "Base VM" "Dev Environment" linked

# Optimize VM for archival (remove snapshots, compact disks)
vmwfusion vm optimize "My VM"
```

### VM Migration
```bash
# Archive running VM with graceful shutdown
vmwfusion migrate archive-running "Production VM" T9

# Restore VM from archive
vmwfusion migrate restore "Production VM" T9 "Local Copy"

# Fast deployment with linked clone
vmwfusion migrate fast-deploy "Base VM" T9 "Quick Dev"

# Show available external drives
vmwfusion migrate list-drives
```

## Advanced Features

### Hash-Based Deduplication
- SHA-256 verification prevents duplicate exports
- Manifest tracking on ExFAT drives for fast verification
- Smart conflict resolution (overwrite/rename/skip)

### VMware Pro Detection
- Automatic detection of VMware Fusion Pro features
- Graceful fallback for Standard edition limitations
- Clear error messages for unsupported operations

### Concurrent Operations
- Parallel ISO discovery and processing
- Concurrent file operations where safe
- Progress reporting for long-running tasks

### Configuration Validation
```bash
# Validate all configuration settings
vmwfusion config validate

# Show current configuration and tool status
vmwfusion config show
```

## Requirements

- **macOS**: Tested on macOS 12+
- **VMware Fusion**: Standard or Pro edition
- **Storage**: ExFAT formatted drives recommended for cross-platform compatibility
- **Optional**: `rsync` for enhanced sync operations (usually pre-installed)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Migration from Shell Scripts

If you were using the previous shell script versions, vmwfusion provides equivalent functionality with these command mappings:

- `iso-manager-simple.sh find` → `vmwfusion iso find`
- `iso-manager-simple.sh quick-export` → `vmwfusion iso quick-export`
- `vm-manager-optimized.sh archive` → `vmwfusion migrate archive`
- `vm-manager-optimized.sh restore` → `vmwfusion migrate restore`
- `vm-cli-manager.sh fast-deploy` → `vmwfusion migrate fast-deploy`

The unified CLI provides all previous functionality with enhanced reliability, performance, and user experience.