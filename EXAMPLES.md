# Go-Dig Examples

This document provides practical examples for common DNS lookup scenarios using go-dig.

## Basic DNS Queries

### Simple Domain Lookup
```cmd
go-dig.exe google.com
```
Returns the IPv4 address(es) for google.com using your system's default DNS server.

### Query Specific Record Types
```cmd
REM IPv6 addresses
go-dig.exe -t AAAA google.com

REM Mail servers
go-dig.exe -t MX gmail.com

REM Text records
go-dig.exe -t TXT google.com

REM Canonical names
go-dig.exe -t CNAME www.github.com
```

## Network Troubleshooting

### Test DNS Resolution
```cmd
REM Check if a domain resolves at all
go-dig.exe example.com

REM Test with different DNS servers to compare results
go-dig.exe -s 8.8.8.8 example.com
go-dig.exe -s 1.1.1.1 example.com
go-dig.exe example.com
```

### Diagnose Connectivity Issues
```cmd
REM Test well-known domains
go-dig.exe google.com
go-dig.exe cloudflare.com
go-dig.exe microsoft.com

REM Test with external DNS servers
go-dig.exe -s 8.8.8.8 google.com
go-dig.exe -s 1.1.1.1 google.com
```

### Check DNS Propagation
```cmd
REM Query multiple DNS servers for the same domain
go-dig.exe -s 8.8.8.8 newdomain.com
go-dig.exe -s 1.1.1.1 newdomain.com
go-dig.exe -s 208.67.222.222 newdomain.com
go-dig.exe newdomain.com
```

## Email Server Configuration

### Verify Mail Server Setup
```cmd
REM Check MX records for your domain
go-dig.exe -t MX yourdomain.com

REM Verify mail servers resolve to IP addresses
go-dig.exe mail.yourdomain.com
go-dig.exe smtp.yourdomain.com
```

### Email Security Records
```cmd
REM Check SPF record
go-dig.exe -t TXT yourdomain.com

REM Check DMARC policy
go-dig.exe -t TXT _dmarc.yourdomain.com

REM Check DKIM selector (replace 'selector' with actual selector)
go-dig.exe -t TXT selector._domainkey.yourdomain.com
```

### Popular Email Providers
```cmd
REM Gmail MX records
go-dig.exe -t MX gmail.com

REM Outlook/Hotmail MX records
go-dig.exe -t MX outlook.com

REM Yahoo Mail MX records
go-dig.exe -t MX yahoo.com
```

## Website and CDN Configuration

### Basic Website Checks
```cmd
REM Check main domain
go-dig.exe yourdomain.com

REM Check www subdomain
go-dig.exe www.yourdomain.com

REM Check if www is a CNAME
go-dig.exe -t CNAME www.yourdomain.com
```

### IPv6 Support Verification
```cmd
REM Check if website supports IPv6
go-dig.exe -t AAAA yourdomain.com
go-dig.exe -t AAAA www.yourdomain.com

REM Test popular sites with IPv6
go-dig.exe -t AAAA google.com
go-dig.exe -t AAAA facebook.com
go-dig.exe -t AAAA cloudflare.com
```

### CDN and Load Balancer Testing
```cmd
REM Check if domain uses CDN (multiple A records)
go-dig.exe cdn-enabled-site.com

REM Test from different DNS servers to see geographic differences
go-dig.exe -s 8.8.8.8 global-site.com
go-dig.exe -s 1.1.1.1 global-site.com
```

## Security and Verification

### Domain Verification Records
```cmd
REM Google site verification
go-dig.exe -t TXT yourdomain.com

REM Microsoft domain verification
go-dig.exe -t TXT yourdomain.com

REM General TXT record inspection
go-dig.exe -t TXT _verification.yourdomain.com
```

### Certificate Authority Authorization (CAA)
```cmd
REM Note: CAA records require special handling, but TXT records may contain related info
go-dig.exe -t TXT yourdomain.com
```

### Subdomain Discovery
```cmd
REM Common subdomains to check
go-dig.exe mail.yourdomain.com
go-dig.exe ftp.yourdomain.com
go-dig.exe api.yourdomain.com
go-dig.exe admin.yourdomain.com
go-dig.exe test.yourdomain.com
go-dig.exe staging.yourdomain.com
```

## Development and Testing

### Local Development
```cmd
REM Test local development domains (if configured in hosts file)
go-dig.exe localhost.dev
go-dig.exe myapp.local

REM Compare with external DNS
go-dig.exe -s 8.8.8.8 myapp.com
```

### Staging Environment Verification
```cmd
REM Check staging subdomains
go-dig.exe staging.yourdomain.com
go-dig.exe test.yourdomain.com
go-dig.exe dev.yourdomain.com

REM Verify they point to different IPs than production
go-dig.exe yourdomain.com
```

## Batch Operations

### Windows Batch Script Example
Create a file called `dns-check.bat`:
```batch
@echo off
echo Checking DNS for %1...
echo.

echo A Records:
go-dig.exe %1
echo.

echo AAAA Records:
go-dig.exe -t AAAA %1
echo.

echo MX Records:
go-dig.exe -t MX %1
echo.

echo TXT Records:
go-dig.exe -t TXT %1
echo.

echo CNAME for www:
go-dig.exe -t CNAME www.%1
```

Usage: `dns-check.bat example.com`

### PowerShell Script Example
Create a file called `DNS-Check.ps1`:
```powershell
param(
    [Parameter(Mandatory=$true)]
    [string]$Domain
)

Write-Host "DNS Check for $Domain" -ForegroundColor Green
Write-Host "=" * 40

Write-Host "`nA Records:" -ForegroundColor Yellow
& go-dig.exe $Domain

Write-Host "`nAAAA Records:" -ForegroundColor Yellow
& go-dig.exe -t AAAA $Domain

Write-Host "`nMX Records:" -ForegroundColor Yellow
& go-dig.exe -t MX $Domain

Write-Host "`nTXT Records:" -ForegroundColor Yellow
& go-dig.exe -t TXT $Domain

Write-Host "`nCNAME for www:" -ForegroundColor Yellow
& go-dig.exe -t CNAME "www.$Domain"
```

Usage: `PowerShell -ExecutionPolicy Bypass -File DNS-Check.ps1 example.com`

## Performance Testing

### Response Time Comparison
```cmd
REM Compare response times from different DNS servers
echo Testing Google DNS:
go-dig.exe -s 8.8.8.8 google.com

echo Testing Cloudflare DNS:
go-dig.exe -s 1.1.1.1 google.com

echo Testing OpenDNS:
go-dig.exe -s 208.67.222.222 google.com

echo Testing System Default:
go-dig.exe google.com
```

### Load Testing Simulation
```cmd
REM Simple loop to test DNS server performance (Windows batch)
for /L %%i in (1,1,10) do (
    echo Query %%i:
    go-dig.exe -s 8.8.8.8 google.com
)
```

## Error Handling Examples

### Testing Error Conditions
```cmd
REM Test non-existent domain
go-dig.exe nonexistentdomain12345.com

REM Test invalid DNS server
go-dig.exe -s 999.999.999.999 google.com

REM Test unsupported record type
go-dig.exe -t INVALID google.com

REM Test malformed domain
go-dig.exe "invalid domain name"
```

### Handling Network Issues
```cmd
REM Test with unreachable DNS server
go-dig.exe -s 192.0.2.1 google.com

REM Test timeout scenarios (disconnect network and run)
go-dig.exe google.com
```

These examples cover the most common use cases for DNS lookups and troubleshooting scenarios you'll encounter in network administration and development work.