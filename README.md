# Go-Dig

A simple DNS lookup tool for Windows, implementing the functionality of the Unix `dig` command. This tool provides essential DNS query capabilities for network troubleshooting and administration on Windows systems.

## Features

- Query multiple DNS record types (A, AAAA, MX, CNAME, TXT)
- Use custom DNS servers
- Clear, readable output with response times
- Comprehensive error handling
- Single executable with no dependencies

## Installation

### Option 1: Download Pre-built Binary

1. Download the latest `go-dig.exe` from the releases page
2. Place it in a directory in your PATH (optional)
3. Run from command prompt or PowerShell

### Option 2: Build from Source

#### Prerequisites
- Go 1.19 or later
- Git (optional)

#### Build Steps

**Using Make (recommended):**
```cmd
git clone <repository-url>
cd go-dig
make build
```

**Using build script:**
```cmd
git clone <repository-url>
cd go-dig
build.bat
```

**Manual build:**
```cmd
git clone <repository-url>
cd go-dig
go build -o go-dig.exe .
```

The executable will be created in the `build/` directory (Make) or current directory (manual build).

## Usage

### Basic Syntax
```
go-dig.exe [options] <domain>
```

### Command-Line Options

| Option | Description | Example |
|--------|-------------|---------|
| `-t <type>` | DNS record type to query | `-t AAAA` |
| `-s <server>` | DNS server to use | `-s 8.8.8.8` |
| `-h` | Show help message | `-h` |

### Supported Record Types

- **A** - IPv4 addresses (default)
- **AAAA** - IPv6 addresses  
- **MX** - Mail exchange records
- **CNAME** - Canonical name records
- **TXT** - Text records

## Examples

### Basic DNS Lookup
Query A records (IPv4 addresses) for a domain:
```cmd
go-dig.exe google.com
```

Output:
```
; <<>> go-dig <<>> google.com
;; Query time: 23 msec
;; SERVER: 192.168.1.1

google.com.             300     IN      A       142.250.191.14
```

### Query Specific Record Types

**IPv6 addresses (AAAA records):**
```cmd
go-dig.exe -t AAAA google.com
```

**Mail exchange records:**
```cmd
go-dig.exe -t MX google.com
```

**Text records:**
```cmd
go-dig.exe -t TXT google.com
```

**Canonical name records:**
```cmd
go-dig.exe -t CNAME www.github.com
```

### Using Custom DNS Servers

**Query using Google DNS:**
```cmd
go-dig.exe -s 8.8.8.8 google.com
```

**Query using Cloudflare DNS:**
```cmd
go-dig.exe -s 1.1.1.1 -t AAAA cloudflare.com
```

### Advanced Examples

**Query MX records using custom DNS server:**
```cmd
go-dig.exe -t MX -s 8.8.8.8 gmail.com
```

**Query TXT records for domain verification:**
```cmd
go-dig.exe -t TXT _dmarc.google.com
```

## Common Use Cases

### Network Troubleshooting
```cmd
REM Check if domain resolves
go-dig.exe example.com

REM Test different DNS servers
go-dig.exe -s 8.8.8.8 example.com
go-dig.exe -s 1.1.1.1 example.com
```

### Email Configuration
```cmd
REM Check mail servers
go-dig.exe -t MX yourdomain.com

REM Verify SPF records
go-dig.exe -t TXT yourdomain.com
```

### Website Configuration
```cmd
REM Check IPv6 support
go-dig.exe -t AAAA yourwebsite.com

REM Verify CNAME setup
go-dig.exe -t CNAME www.yourwebsite.com
```

## Error Handling

The tool provides clear error messages for common issues:

- **Invalid domain names**: Clear validation error messages
- **Network issues**: Timeout and connectivity error details  
- **DNS server problems**: Server unreachable or error responses
- **Record not found**: NXDOMAIN and no-record-found messages

### Exit Codes

- `0` - Success
- `1` - Invalid arguments or general error
- `2` - Network or DNS error  
- `3` - System error
- `130` - Interrupted by user (Ctrl+C)

## Development

### Running Tests
```cmd
make test
```

### Building for Development
```cmd
make build-dev
```

### Code Formatting
```cmd
make fmt
```

## Troubleshooting

### Common Issues

**"Command not found" error:**
- Ensure `go-dig.exe` is in your PATH or use the full path to the executable

**DNS timeout errors:**
- Check your internet connection
- Try using a different DNS server with `-s` option
- Some corporate networks may block DNS queries

**Permission errors:**
- Run command prompt as administrator if needed
- Check Windows Defender or antivirus software

### Getting Help

For usage help:
```cmd
go-dig.exe -h
```

For additional support, please check the project's issue tracker.

## License

This project is open source. See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.