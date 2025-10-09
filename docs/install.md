## Installation

### 1. Download
* Download the appropriate archive for your platform from the @RELEASES@ page

### 2. Verify Checksum

```bash
# Download the checksums file
# Verify the archive
sha256sum -c go-xbuild-go-vX.X.X-darwin-arm64.d.tar.gz
```
Repeat the step for other archives

### 3. Extract
macOS/Linux:

```bash
tar -xzf go-xbuild-go-vX.X.X-darwin-arm64.d.tar.gz
cd go-xbuild-go-vX.X.X-darwin-arm64.d
```

Repeat the step for other archives

Windows:

The tar command is available in Windows 10 (1803) and later, or you can
use the GUI (right-click → Extract All). After extracting, copy/rename the
binary somewhere in your PATH.

### 4. Install

```bash
# macOS/Linux
sudo cp go-xbuild-go-vX.X.X-darwin-arm64 /usr/local/bin/go-xbuild-go
sudo chmod +x /usr/local/bin/go-xbuild-go
```

```bash
# Windows
copy go-xbuild-go-vX.X.X-windows-amd64.exe C:\Windows\System32\go-xbuild-go.exe
```

### Building from source

Install @GO@ first

```bash
git clone https://github.com/muquit/go-xbuild-go
cd go-xbuild-go
go build .
or 
make build
```

Please look at @MAKEFILE@ for more info

