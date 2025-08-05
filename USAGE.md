# Go-Dig Usage Guide

## Command-Line Options Reference

### Basic Syntax
```
go-dig.exe [OPTIONS] DOMAIN
```

### Options

#### `-t, --type <RECORD_TYPE>`
Specifies the DNS record type to query.

**Supported Record Types:**
- `A` (default) - IPv4 address records
- `AAAA` - IPv6 address records
- `MX` - Mail exchange records
- `CNAME` - Canonical name records
- `TXT` - Text records

**Examples:**
```cmd
go-dig.exe -t A google.com
go-dig.exe -t AAAA google.com
go-dig.exe -t MX gmail.com
go-dig.exe -t CNAME www.github.com
go-dig.exe -t TXT google.com
```

#### `-s, --server <DNS_SERVER>`
Specifies the DNS server to use for the query.

**Format:** IPv4 address (e.g., 8.8.8.8)

**Common DNS Servers:**
- `8.8.8.8` - Google Public DNS
- `8.8.4.4` - Google Public DNS (secondary)
- `1.1.1.1` - Cloudflare DNS
- `1.0.0.1` - Cloudflare DNS (secondary)
- `208.67.222.222` - OpenDNS
- `208.67.220.220` - OpenDNS (secondary)

**Examples:**
```cmd
go-dig.exe -s 8.8.8.8 google.com
go-dig.exe -s 1.1.1.1 -t AAAA cloudflare.com
```

#### `-h, --help`
Displays help information and exits.

```cmd
go-dig.exe -h
```

### Domain Argument

The domain name to query. This is a required argument.

**Valid formats:**
- `example.com`
- `www.example.com`
- `subdomain.example.com`
- `_service._protocol.example.com` (for SRV-like queries)

## Record Type Details

### A Records (IPv4 Addresses)
Returns IPv4 addresses for the domain.

```cmd
go-dig.exe google.com
go-dig.exe -t A google.com
```

**Sample Output:**
```
; <<>> go-dig <<>> google.com
;; Query time: 23 msec
;; SERVER: 192.168.1.1

google.com.             300     IN      A       142.250.191.14
```

### AAAA Records (IPv6 Addresses)
Returns IPv6 addresses for the domain.

```cmd
go-dig.exe -t AAAA google.com
```

**Sample Output:**
```
; <<>> go-dig <<>> google.com
;; Query time: 45 msec
;; SERVER: 192.168.1.1

google.com.             300     IN      AAAA    2607:f8b0:4004:c1b::65
```

### MX Records (Mail Exchange)
Returns mail server information for the domain.

```cmd
go-dig.exe -t MX gmail.com
```

**Sample Output:**
```
; <<>> go-dig <<>> gmail.com
;; Query time: 67 msec
;; SERVER: 192.168.1.1

gmail.com.              3600    IN      MX      5 gmail-smtp-in.l.google.com.
gmail.com.              3600    IN      MX      10 alt1.gmail-smtp-in.l.google.com.
```

### CNAME Records (Canonical Name)
Returns canonical name mappings for the domain.

```cmd
go-dig.exe -t CNAME www.github.com
```

**Sample Output:**
```
; <<>> go-dig <<>> www.github.com
;; Query time: 34 msec
;; SERVER: 192.168.1.1

www.github.com.         3600    IN      CNAME   github.com.
```

### TXT Records (Text)
Returns text records associated with the domain.

```cmd
go-dig.exe -t TXT google.com
```

**Sample Output:**
```
; <<>> go-dig <<>> google.com
;; Query time: 56 msec
;; SERVER: 192.168.1.1

google.com.             300     IN      TXT     "v=spf1 include:_spf.google.com ~all"
google.com.             300     IN      TXT     "docusign=05958488-4752-4ef2-95eb-aa7ba8a3bd0e"
```

## Output Format

### Successful Query Output
```
; <<>> go-dig <<>> [domain] [record-type]
;; Query time: [time] msec
;; SERVER: [dns-server]

[domain].    [ttl]    IN    [type]    [value]
```

### Error Output
Errors are displayed to stderr with descriptive messages:

```
Error: domain not found (NXDOMAIN)
Error: DNS server 8.8.8.8 is unreachable
Error: invalid domain name format
Error: unsupported record type 'INVALID'
```

## Advanced Usage Patterns

### Testing DNS Propagation
```cmd
REM Test multiple DNS servers
go-dig.exe -s 8.8.8.8 example.com
go-dig.exe -s 1.1.1.1 example.com
go-dig.exe -s 208.67.222.222 example.com
```

### Email Server Verification
```cmd
REM Check MX records
go-dig.exe -t MX yourdomain.com

REM Check SPF records
go-dig.exe -t TXT yourdomain.com

REM Check DMARC policy
go-dig.exe -t TXT _dmarc.yourdomain.com
```

### Website Configuration Check
```cmd
REM Check A record
go-dig.exe yourdomain.com

REM Check AAAA record (IPv6)
go-dig.exe -t AAAA yourdomain.com

REM Check www CNAME
go-dig.exe -t CNAME www.yourdomain.com
```

## Troubleshooting

### Common Error Messages

**"invalid domain name format"**
- Check domain name spelling
- Ensure proper domain format (no spaces, valid characters)

**"DNS server X.X.X.X is unreachable"**
- Check internet connection
- Verify DNS server IP address
- Try a different DNS server

**"domain not found (NXDOMAIN)"**
- Domain doesn't exist
- Check domain spelling
- Domain may not be registered

**"no records found for type X"**
- Domain exists but doesn't have records of the requested type
- Try different record type

**"query timeout"**
- Network connectivity issues
- DNS server overloaded
- Try different DNS server or check connection