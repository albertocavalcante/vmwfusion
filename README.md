# VMware Fusion CLI

A powerful, modern command-line interface for managing VMware Fusion virtual machines. This CLI provides comprehensive functionality for VM lifecycle management, configuration, migration, and monitoring.

## Features

- **VM Management**: Create, start, stop, suspend, resume, and delete virtual machines
- **Configuration Management**: Modify VM settings including memory, CPU, disk, and network
- **ISO Operations**: Mount, unmount, and validate ISO files
- **Migration & Export**: Move VMs between locations and export to different formats
- **Discovery**: Automatically discover existing VMs in your system
- **Verification**: Validate VM integrity and system configuration
- **System Information**: Display detailed system and VMware installation info

## Installation

### Prerequisites

- VMware Fusion installed on macOS
- Go 1.21 or later (for building from source)

### Build from Source

```bash
git clone <repository-url>
cd vmwfusion
go mod download
go build -o vmwfusion
```

### Install Dependencies

```bash
go mod download
```

## Usage

### Basic Commands

```bash
# List all virtual machines
vmwfusion vm list

# Create a new VM
vmwfusion vm create myvm --os linux --memory 4096 --cpus 4 --disk 50

# Start a VM
vmwfusion vm start myvm

# Stop a VM
vmwfusion vm stop myvm

# Get VM information
vmwfusion vm info myvm
```

### Configuration Management

```bash
# Get VM configuration
vmwfusion config get myvm

# Set VM memory
vmwfusion config memory myvm 8192

# Set VM CPU count
vmwfusion config cpu myvm 8

# Add network interface
vmwfusion config network add myvm nat
```

### ISO Operations

```bash
# List available ISOs
vmwfusion iso list

# Validate an ISO file
vmwfusion iso validate /path/to/ubuntu.iso

# Mount ISO to VM
vmwfusion iso mount myvm /path/to/ubuntu.iso

# Unmount ISO from VM
vmwfusion iso unmount myvm
```

### Migration & Export

```bash
# Move VM to new location
vmwfusion migrate path myvm /new/path

# Convert VM format
vmwfusion migrate convert myvm ova

# Export VM
vmwfusion migrate export myvm --dest /export/path --format ova

# Import VM
vmwfusion migrate import /path/to/vm.ova
```

### Discovery & Verification

```bash
# Discover VMs in default paths
vmwfusion discover vms

# Scan specific path
vmwfusion discover scan /path/to/vms

# Verify VM integrity
vmwfusion verify vm myvm

# Verify all VMs
vmwfusion verify all

# Verify VMware installation
vmwfusion verify installation
```

### System Information

```bash
# Display system information
vmwfusion info system

# Show VMware Fusion info
vmwfusion info vmware

# Display VM statistics
vmwfusion info stats

# Show version information
vmwfusion info version
```

## Command Structure

The CLI is organized into logical command groups:

### VM Commands (`vmwfusion vm`)
- `list` - List all virtual machines
- `create` - Create a new virtual machine
- `delete` - Delete a virtual machine
- `start` - Start a virtual machine
- `stop` - Stop a virtual machine
- `suspend` - Suspend a virtual machine
- `resume` - Resume a suspended virtual machine
- `info` - Show detailed VM information

### Configuration Commands (`vmwfusion config`)
- `get` - Get VM configuration
- `set` - Update VM configuration
- `memory` - Set memory allocation
- `cpu` - Set CPU count
- `network` - Network interface management

### ISO Commands (`vmwfusion iso`)
- `list` - List available ISO files
- `validate` - Validate ISO file integrity
- `mount` - Mount ISO to VM
- `unmount` - Unmount ISO from VM
- `info` - Show ISO file information

### Migration Commands (`vmwfusion migrate`)
- `vm` - Migrate a virtual machine
- `path` - Move VM to new location
- `convert` - Convert VM format
- `export` - Export virtual machine
- `import` - Import virtual machine
- `batch` - Batch migrate multiple VMs

### Discovery Commands (`vmwfusion discover`)
- `vms` - Discover VMs in paths
- `paths` - Show default discovery paths
- `scan` - Deep scan specific path

### Verification Commands (`vmwfusion verify`)
- `vm` - Verify specific VM
- `all` - Verify all VMs
- `installation` - Verify VMware installation
- `config` - Verify VM configuration

### Information Commands (`vmwfusion info`)
- `system` - Display system information
- `vmware` - Show VMware installation info
- `stats` - Display VM statistics
- `version` - Show version information

## Global Flags

- `--config` - Specify configuration file
- `--verbose, -v` - Enable verbose output
- `--json, -j` - Output in JSON format (where applicable)

## Configuration

The CLI can be configured using a YAML configuration file. By default, it looks for `.vmwfusion.yaml` in your home directory.

Example configuration:

```yaml
# Default VM settings
defaults:
  memory: 2048
  cpus: 2
  disk_size: 20
  os: "linux"

# Discovery paths
discovery:
  paths:
    - "~/Virtual Machines"
    - "~/Documents/Virtual Machines"
  include_hidden: false
  max_depth: 5

# VMware settings
vmware:
  fusion_path: "/Applications/VMware Fusion.app"
  tools_path: "/usr/local/bin"
```

## Architecture

The CLI is built with a clean, modular architecture:

- **cmd/**: Cobra command definitions organized by functionality
- **internal/types/**: Type definitions and data structures
- **internal/vmware/**: VMware Fusion client interface and implementation
- **internal/utils/**: Utility functions and helpers

### Key Design Principles

1. **Separation of Concerns**: Commands, business logic, and VMware integration are clearly separated
2. **Interface-Driven**: VMware operations are abstracted behind a clean interface
3. **Extensible**: Easy to add new commands and functionality
4. **Testable**: Modular design enables comprehensive testing
5. **User-Friendly**: Intuitive command structure with helpful output and error messages

## Development

### Project Structure

```
vmwfusion/
├── cmd/                    # Command definitions
│   ├── root.go            # Root command and global configuration
│   ├── vm.go              # VM management commands
│   ├── config.go          # Configuration commands
│   ├── iso.go             # ISO operations
│   ├── migrate.go         # Migration commands
│   ├── discover.go        # Discovery commands
│   ├── verify.go          # Verification commands
│   └── info.go            # Information commands
├── internal/
│   ├── types/             # Type definitions
│   │   └── types.go
│   ├── vmware/            # VMware client
│   │   └── client.go
│   └── utils/             # Utilities
│       └── utils.go
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── README.md             # This file
```

### Adding New Commands

1. Create a new file in `cmd/` directory
2. Define your command using Cobra conventions
3. Add any new types to `internal/types/`
4. Implement VMware operations in the client interface
5. Update the README documentation

### Building

```bash
# Build for current platform
go build -o vmwfusion

# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o vmwfusion-darwin-amd64
GOOS=linux GOARCH=amd64 go build -o vmwfusion-linux-amd64
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific tests
go test ./cmd -v
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue in the GitHub repository
- Check the documentation for common use cases
- Review the command help text: `vmwfusion [command] --help`