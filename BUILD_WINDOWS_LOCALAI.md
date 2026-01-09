# Building LocalAI.exe on Windows

## Prerequisites

### System Requirements
- Windows 10+
- Git (for cloning and version info)
- CMake 3.24+ (for some backends)
- OpenSSL libraries (for HTTPS/TLS)

### Go Installation

LocalAI is written in Go. You need Go 1.24.4+ installed.

#### 1. Install Go
Download and install from [golang.org](https://golang.org/dl/)

**Recommended:** Go 1.25.x (as of January 2026)

```powershell
# Verify Go installation
go version
# Should output: go version go1.25.x windows/amd64
```

#### 2. Verify Go Path
```powershell
go env GOPATH
# Typically: C:\Users\<username>\go
```

Ensure `$GOPATH\bin` is in your PATH for installed tools.

#### 3. (Optional) Additional Build Tools
For Windows builds with all features, optionally install:
- **FFmpeg** - For audio/video processing
- **Build tools** - For C/C++ compilation (if building backends)
- **CMake** - For backend compilation

## Building LocalAI

### 1. Clone the Repository
```powershell
cd G:\LocalAI  # or your preferred location
git clone https://github.com/ericblade/LocalAI.git .
cd G:\LocalAI
```

**Note:** If already cloned, ensure you're on the correct branch:
```powershell
git checkout dev      # or master for stable
git pull origin dev
git submodule update --init --recursive
```

### 2. Install Protocol Buffer Compiler

LocalAI uses gRPC protocol buffers. Install the compiler and Go plugin generators:

```powershell
# Download protoc (v26.1 matches the CI)
$ProtocVersion = "26.1"
$ProtocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v${ProtocVersion}/protoc-${ProtocVersion}-win64.zip"
Invoke-WebRequest -Uri $ProtocUrl -OutFile protoc.zip
Expand-Archive protoc.zip -DestinationPath C:\protoc
Remove-Item protoc.zip

# Add protoc to PATH
$env:PATH += ";C:\protoc\bin"

# Install Go protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1958fcbe2ca8bd93af633f11e97d44e567e945af

# Verify PATH includes go/bin
$env:PATH += ";$((go env GOPATH))\bin"
```

**Verify:**
```powershell
protoc --version
# Should output: libprotoc 26.1
```

### 3. Generate Protocol Buffer Code

```powershell
cd G:\LocalAI
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1958fcbe2ca8bd93af633f11e97d44e567e945af

# Generate Go protobuf files
make protogen-go
```

**What this does:**
- Compiles `.proto` files in `pkg/grpc/proto/`
- Generates Go bindings in `pkg/grpc/` (auto-generated, don't edit)

### 4. Build the Main Binary

Simple build:
```powershell
cd G:\LocalAI
go build -o local-ai.exe ./cmd/local-ai
```

**Build Output:**
- Binary: `G:\LocalAI\local-ai.exe`
- Size: ~50-80 MB (varies with features)

### 5. (Optional) Build the Launcher

LocalAI includes a GUI launcher application. **Note:** The launcher requires CGO and OpenGL development headers. Building it on Windows requires a GCC for Windows, such as MinGW.

If you want to skip the launcher and just use the server:
- Use `local-ai.exe` directly
- The launcher is optional and not required to run LocalAI

Building the launcher is not something I've attempted to do, and is beyond the scope of htis guide. If you figure it out, please document it and update this file.

## Troubleshooting

### Build Fails: "protoc not found"
```powershell
# Check if protoc is in PATH
where protoc
# Should output: C:\protoc\bin\protoc.exe

# If not found, add to PATH
$env:PATH += ";C:\protoc\bin;$((go env GOPATH))\bin"
```

### Build Fails: "cannot find package"
```powershell
# Go may need to download dependencies
go mod download
go mod tidy

# Then retry build
go build -o local-ai.exe ./cmd/local-ai
```

### Build Fails: "missing protoc-gen-go"
```powershell
# Ensure PATH includes go/bin
$env:PATH += ";$((go env GOPATH))\bin"

# Reinstall plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1958fcbe2ca8bd93af633f11e97d44e567e945af
```

### "binary is too large" / Slow startup
- This is normal for Go binaries on Windows
- Use `-s -w` ldflags to reduce size
- Consider UPX compression (if not interferes with Windows Defender)

## Running LocalAI

### Basic Startup
```powershell
cd G:\LocalAI
.\local-ai.exe
```

**Default behavior:**
- Listens on `http://localhost:8080`
- Loads models from `./models` directory

### With Custom Configuration
```powershell
.\local-ai.exe --models-path "C:\MyModels" --address 0.0.0.0:9000
```

## Dependencies

### Runtime Dependencies
None required if built as static binary. All Go dependencies are compiled in.

### Build Dependencies
- **Go 1.24.4+** - Language toolchain
- **protoc 26.1+** - Protocol buffer compiler
- **git** - Version information
- **CMake** (optional) - For building C++ backends
- **C++ compiler** (optional) - For native compilation support

### Go Modules
Key dependencies (from go.mod):
- `github.com/labstack/echo/v4` - HTTP framework
- `github.com/google/grpc` - gRPC framework
- `github.com/mudler/LocalAI/internal` - Internal packages
- Various ML/AI libraries (loaded dynamically via plugins)

## Continuous Builds / Dev Mode

For development with live reload on Windows:

```powershell
# Install air (live reload tool)
go install github.com/air-verse/air@latest

# Run in watch mode with Windows config
cd G:\LocalAI
air -c air.windows.toml
```

This watches source files and rebuilds on changes.

**Note:** Use `air.windows.toml` on Windows (which outputs `local-ai.exe`). On other platforms, use `air -c .air.toml` (which outputs `local-ai`).

## Testing the Build

```powershell
# Start LocalAI
\.\local-ai.exe --address 127.0.0.1:8080

# Test in another PowerShell window
$response = Invoke-RestMethod -Uri "http://localhost:8080/healthz"
$response

# Should output health status JSON
```

## Build Artifacts

After building:

```powershell
ls -la G:\LocalAI\local-ai.exe
```

- local-ai.exe — main server (50-80 MB). Self-contained; no extra runtime needed.

## Environment Details (as of January 2026)

- **Go Version:** 1.24.4+ (up to 1.25.x recommended)
- **Protoc Version:** 26.1
- **protoc-gen-go:** v1.34.2
- **protoc-gen-go-grpc:** 1958fcbe2ca8bd93af633f11e97d44e567e945af
- **Repository:** https://github.com/ericblade/LocalAI (fork)
- **Branch:** dev (or master for stable)
- **Build Date:** January 8, 2026

## References

- [Go Documentation](https://golang.org/doc/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [LocalAI Repository](https://github.com/mudler/LocalAI)
- [LocalAI Documentation](https://localai.io)
