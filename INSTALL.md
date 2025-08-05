# Installation Guide

## Quick Start

1. **Download or Build**: Get the `go-dig.exe` executable
2. **Place in PATH** (optional): Copy to a directory in your system PATH
3. **Run**: Open Command Prompt or PowerShell and run `go-dig.exe`

## Installation Methods

### Method 1: Download Pre-built Binary (Recommended)

1. Download `go-dig.exe` from the releases page
2. Save it to a convenient location (e.g., `C:\Tools\go-dig.exe`)
3. Test the installation:
   ```cmd
   C:\Tools\go-dig.exe google.com
   ```

### Method 2: Build from Source

#### Prerequisites
- Go 1.19 or later installed
- Git (optional, for cloning repository)

#### Steps

**Option A: Using build script (Windows)**
```cmd
git clone <repository-url>
cd go-dig
build.bat
```

**Option B: Using Make (if available)**
```cmd
git clone <repository-url>
cd go-dig
make build
```

**Option C: Manual build**
```cmd
git clone <repository-url>
cd go-dig
go build -o go-dig.exe .
```

The executable will be created in the `build/` directory or current directory.

## Adding to System PATH (Optional)

To use `go-dig` from anywhere in the command line:

### Windows 10/11 (GUI Method)
1. Copy `go-dig.exe` to `C:\Tools\` (create directory if needed)
2. Open Settings → System → About → Advanced system settings
3. Click "Environment Variables"
4. Under "System variables", find and select "Path", then click "Edit"
5. Click "New" and add `C:\Tools`
6. Click "OK" to close all dialogs
7. Open a new Command Prompt and test: `go-dig google.com`

### Windows (Command Line Method)
```cmd
REM Create tools directory
mkdir C:\Tools

REM Copy executable
copy go-dig.exe C:\Tools\

REM Add to PATH (requires admin privileges)
setx PATH "%PATH%;C:\Tools" /M
```

### PowerShell Method
```powershell
# Create tools directory
New-Item -ItemType Directory -Path "C:\Tools" -Force

# Copy executable
Copy-Item "go-dig.exe" "C:\Tools\"

# Add to PATH (requires admin privileges)
$env:PATH += ";C:\Tools"
[Environment]::SetEnvironmentVariable("PATH", $env:PATH, [EnvironmentVariableTarget]::Machine)
```

## Verification

After installation, verify everything works:

```cmd
REM Test basic functionality
go-dig google.com

REM Test help
go-dig -h

REM Test different record types
go-dig -t AAAA google.com
go-dig -t MX gmail.com

REM Test custom DNS server
go-dig -s 8.8.8.8 google.com
```

## Troubleshooting Installation

### "Command not found" or "'go-dig' is not recognized"

**Solution 1**: Use full path to executable
```cmd
C:\path\to\go-dig.exe google.com
```

**Solution 2**: Add directory to PATH (see above)

**Solution 3**: Copy to existing PATH directory
```cmd
REM Copy to Windows directory (not recommended but works)
copy go-dig.exe C:\Windows\System32\
```

### "Windows protected your PC" or SmartScreen warning

This may appear when running the executable for the first time:

1. Click "More info"
2. Click "Run anyway"
3. This is normal for unsigned executables

### Antivirus software blocking execution

Some antivirus software may flag the executable:

1. Add `go-dig.exe` to your antivirus whitelist
2. Temporarily disable real-time protection during installation
3. Build from source if you prefer

### Permission errors

If you get permission errors:

1. Run Command Prompt as Administrator
2. Ensure you have write permissions to the target directory
3. Try installing to your user directory instead of system directories

## Uninstallation

To remove go-dig:

1. Delete the `go-dig.exe` file
2. Remove the directory from your PATH if you added it
3. Delete any shortcuts or aliases you created

## System Requirements

- **Operating System**: Windows 7 or later
- **Architecture**: 64-bit (x64)
- **Dependencies**: None (self-contained executable)
- **Network**: Internet connection for DNS queries
- **Disk Space**: ~5MB for executable

## Security Considerations

- The executable makes network connections to DNS servers
- No data is stored or transmitted beyond DNS queries
- Source code is available for security review
- Consider running from a restricted user account if needed

## Getting Help

If you encounter issues:

1. Check this installation guide
2. Review the README.md for usage examples
3. Test with known working domains (google.com, cloudflare.com)
4. Check your network connectivity
5. Try different DNS servers with the `-s` option