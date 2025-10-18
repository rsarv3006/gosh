# Installation

Get gosh running on your system with these simple installation methods.

## Homebrew (Recommended)

The easiest way to install gosh is via Homebrew:

```bash
# Install via homebrew tap
brew install rsarv3006/gosh/gosh

# Add to system shells (optional, to use as login shell)
echo '/opt/homebrew/bin/gosh' | sudo tee -a /etc/shells

# Set as default shell (optional)
chsh -s /opt/homebrew/bin/gosh

# Run
gosh
```

## Go Install

If you have Go installed, you can directly install the latest release:

```bash
# Install the latest release
go install github.com/rsarv3006/gosh@latest

# Make sure your Go bin directory is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Run
gosh
```

## Build from Source

For developers or those who want the latest development version:

```bash
# Clone the repository
git clone https://github.com/rsarv3006/gosh
cd gosh

# Build
go build

# Run
./gosh

# Or install for system-wide use
go install .
```

## System Requirements

- **Operating System**: macOS, Linux (Windows users can use PowerShell as an alternative)
- **Go**: Version 1.21 or higher (for building from source)
- **Architecture**: amd64, arm64

## Verify Installation

Once installed, verify gosh is working:

```bash
# Check version
gosh --version

# Or inside gosh
gosh> fmt.Println("main".GetVersion())
0.2.3

# Help command
gosh --help
```

## PATH Configuration

Make sure gosh is in your PATH:

### For Go Install Users

```bash
# Add this to your ~/.zshrc or ~/.bashrc
export PATH=$PATH:$(go env GOPATH)/bin

# Reload your shell
source ~/.zshrc  # or ~/.bashrc
```

### For Homebrew Users

Homebrew automatically configures the PATH, but you may need to start a new terminal session.

## Usage Modes

### Interactive Shell

The most common usage - start an interactive session:

```bash
gosh
```

### Single Command Execution

Execute a single command and exit:

```bash
gosh -c 'fmt.Println("Hello from gosh")'

# Or with shell commands
gosh -c 'ls -la'

# With Go code
gosh -c 'files := $(ls); fmt.Printf("Found %d files\n", len(strings.Split(files, "\n")))'
```

### As Login Shell

Make gosh your default shell:

```bash
# Add to available shells
echo $(which gosh) | sudo tee -a /etc/shells

# Set as default shell
chsh -s $(which gosh)
```

## Post-Installation Setup

### Create Configuration Directory

Set up your gosh configuration:

```bash
# Create config directory
mkdir -p ~/.config/gosh

# Initialize with example config
gosh
gosh> init
```

This creates an example `~/.config/gosh/config.go` file with useful functions.

### Verify Functions

Test some built-in functions:

```bash
gosh> gs  # Git status helper
gosh> build()  # Build helper
gosh> test()  # Test helper
```

## Troubleshooting

### Common Issues

**Command not found: gosh**

```bash
# Check if gosh is in PATH
which gosh

# If not found, add Go bin directory to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

**Permission denied**

```bash
# Make sure gosh is executable
chmod +x $(which gosh)
```

**Config loading errors**

```bash
# Check config directory permissions
ls -la ~/.config/gosh/

# If needed, fix permissions
chmod 755 ~/.config/gosh
```

**Go not installed**

```bash
# Install Go on macOS
brew install go

# Or download from https://go.dev/
```

### Getting Help

If you encounter issues:

1. Check this troubleshooting section
2. Review the [User Guide](guide.md) for configuration help
3. Open an issue on [GitHub](https://github.com/rsarv3006/gosh/issues)

## Upgrade Instructions

### Homebrew Users

```bash
brew upgrade rsarv3006/gosh/gosh
```

### Go Install Users

```bash
go install github.com/rsarv3006/gosh@latest
```

### Source Users

```bash
cd gosh
git pull origin main
go build
```

## What's Next?

Once gosh is installed and working:

- [Getting Started Guide](getting-started.md) - Learn basic usage
- [Configuration Examples](config.md) - Set up custom functions
- [User Guide](guide.md) - Advanced features
- [CLI Reference](reference.md) - Complete command documentation

---

Enjoy your enhanced shell experience! 🚀
