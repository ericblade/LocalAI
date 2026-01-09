# Building grpc-server.exe on Windows

## Prerequisites

### System Requirements
- Windows 10+
- Visual Studio 2022+ (Professional or Community with C++ tools)
- CMake 3.24+
- PowerShell 5.1+

### Environment Setup

#### 1. Install vcpkg and Required Packages
```powershell
# Clone vcpkg (if not already present)
git clone https://github.com/Microsoft/vcpkg.git Z:\src\vcpkg
cd Z:\src\vcpkg
.\bootstrap-vcpkg.bat

# Install required packages for x64-windows
.\vcpkg install grpc:x64-windows protobuf:x64-windows getopt-win32:x64-windows
```

**Packages Installed:**
- `grpc:x64-windows` v1.71.0 - gRPC C++ libraries and tools
- `protobuf:x64-windows` v29.5.0 - Protocol Buffers
- `getopt-win32:x64-windows` v2.42.0 - Command-line parsing for Windows (replaces Unix getopt)

#### 2. Add Tools to PATH
After vcpkg installation, add these to your system PATH or PowerShell session:

```powershell
$env:PATH += ";Z:\src\vcpkg\installed\x64-windows\tools\protobuf"
$env:PATH += ";Z:\src\vcpkg\installed\x64-windows\tools\grpc"
$env:PATH += ";Z:\src\vcpkg\installed\x64-windows\bin"
```

#### 3. (Optional) Vulkan Support
If building with Vulkan GPU acceleration:
- Install [Vulkan SDK](https://vulkan.lunarg.com/sdk/home)
- Set `GGML_VULKAN=1` during CMake configuration

## Build Steps

### 1. Clone/Update llama.cpp Repository
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp

# Clone the llama.cpp repository if it doesn't exist
git clone https://github.com/ggml-org/llama.cpp.git

cd llama.cpp
git fetch origin
git checkout master
```

### 2. Copy gRPC Server Implementation
Copy the gRPC server files from the LocalAI repo into the llama.cpp source tree:

```powershell
# Copy grpc-server.cpp to llama.cpp tools/server/
Copy-Item "..\grpc-server.cpp" -Destination ".\tools\server\grpc-server.cpp" -Force

# Ensure tools/server directory exists
mkdir -ErrorAction SilentlyContinue .\tools\server
```

### 3. Create Build Directory
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp
mkdir build -ErrorAction SilentlyContinue
cd build
```

### 4. Run CMake Configuration
```powershell
cmake .. `
  -G "Visual Studio 17 2022" `
  -DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows `
  -DGGML_VULKAN=1 `
  -DBUILD_SHARED_LIBS=OFF `
  -DLLAMA_CURL=ON
```

**Key Flags:**
- `-G "Visual Studio 17 2022"` - Use VS 2022 generator (adjust version if different)
- `-DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows` - Tell CMake where to find vcpkg packages
- `-DGGML_VULKAN=1` - Enable Vulkan GPU support (optional)
- `-DBUILD_SHARED_LIBS=OFF` - Build static libraries (simpler deployment)
- `-DLLAMA_CURL=ON` - Enable curl for model downloads

### 5. Build grpc-server Target
```powershell
cmake --build . --config Release --target grpc-server
```

**Build Output:**
- Binary: `llama.cpp\build\tools\grpc-server\Release\grpc-server.exe`
- Size: ~67 MB (Release build with Vulkan)

### 6. Deploy Binary and Dependencies

#### Copy grpc-server.exe
```powershell
Copy-Item "llama.cpp\build\tools\grpc-server\Release\grpc-server.exe" `
  -Destination "G:\LocalAI\backend\cpp\llama-cpp\grpc-server.exe" `
  -Force
```

#### Copy Required DLLs
```powershell
# Copy all DLLs from vcpkg to backend directory
Copy-Item "Z:\src\vcpkg\installed\x64-windows\bin\*.dll" `
  -Destination "G:\LocalAI\backend\cpp\llama-cpp\" `
  -Force -ErrorAction SilentlyContinue
```

**DLLs Typically Required:**
- libprotobuf.dll, libprotoc.dll (Protocol Buffers)
- grpc.dll, grpc++.dll (gRPC)
- getopt.dll (command-line parsing on Windows)
- abseil_dll.dll (dependency)
- libcurl.dll (HTTP/model downloads)
- zlib.dll (compression, if needed)

## Running grpc-server

### Start the Server
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp
.\grpc-server.exe -a localhost:50051
```

**Output:** `Server listening on localhost:50051`

### Background Execution (for LocalAI)
```powershell
Start-Process .\grpc-server.exe -ArgumentList "-a 0.0.0.0:50051" -WindowStyle Hidden
```

## Integration with LocalAI

### Configuration File
Update `G:\LocalAI\configuration\external_backends.json`:

```json
{
  "llama-cpp": "localhost:50051"
}
```

**Key Points:**
- Backend name must match model YAML `backend:` field
- URI format: `host:port` (not file path)
- grpc-server must be running before LocalAI starts

### Test
```powershell
cd G:\LocalAI
.\local-ai.exe --address 127.0.0.1:8080
```

Monitor logs for successful backend connection. Model inference should work via REST/Web UI.

## Code Modifications

The build incorporates minimal patches to upstream llama.cpp:

### 1. Auto-fit Context Support (common/common.cpp)
- Initializes `tensor_buft_overrides` for buffer management
- Supports `n_ctx <= 0` for automatic context size fitting

### 2. Windows Compatibility (tools/grpc-server/CMakeLists.txt)
- Removed unused absl dependency
- Added getopt-win32 configuration
- Conditional linking for Windows platforms

### 3. gRPC Server Windows Fixes (grpc-server.cpp)
- Added reflection guard for Windows DLL exports
- Fixed JSON object serialization with `.dump()`
- Proper initialization of tensor buffer overrides

## Troubleshooting

### Build Fails: "protoc not found"
- Ensure PATH includes `Z:\src\vcpkg\installed\x64-windows\tools\protobuf`
- Verify `protoc.exe` exists in that directory

### Build Fails: "grpc_cpp_plugin not found"
- Ensure PATH includes `Z:\src\vcpkg\installed\x64-windows\tools\grpc`
- Verify `grpc_cpp_plugin.exe` exists in that directory

### grpc-server.exe Won't Start: Missing DLLs
- Copy all DLLs from `Z:\src\vcpkg\installed\x64-windows\bin\` to backend directory
- Use Windows Dependency Walker to identify missing imports

### LocalAI: "Backend not found"
- Verify grpc-server.exe is running: `netstat -ano | findstr 50051`
- Check external_backends.json format (object, not array)
- Ensure URI is `localhost:50051` (not file path)

### Port Already in Use
```powershell
# Find process on port 50051
Get-NetTCPConnection -LocalPort 50051 | Select-Object @{Name="PID";Expression={$_.OwningProcess}} | ForEach-Object {Get-Process -Id $_.PID}

# Kill if needed
Stop-Process -Id <PID> -Force
```

## Rebuild Workflow

For subsequent builds (after upstream updates):

```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp
git fetch upstream
git rebase upstream/master  # or cherry-pick patches if using patch files
cd build
cmake --build . --config Release --target grpc-server -j24
Copy-Item "build\tools\grpc-server\Release\grpc-server.exe" -Destination "..\grpc-server.exe" -Force
```

## Testing

### Manual Inference Test
```powershell
# Start grpc-server
.\grpc-server.exe -a localhost:50051 &

# In another terminal, start LocalAI
cd G:\LocalAI
.\local-ai.exe --address 127.0.0.1:8080

# Test via REST API
Invoke-RestMethod -Uri "http://localhost:8080/v1/chat/completions" `
  -Method Post `
  -Body (@{model="llama3-vulkan"; messages=@(@{role="user"; content="Hello"})} | ConvertTo-Json) `
  -ContentType "application/json"
```

## Environment Details (as of January 2026)

- **Build Date:** January 8, 2026
- **CMake Version:** 3.24+
- **Visual Studio:** 2022 Professional
- **Vulkan SDK:** Z:\tools\VulkanSDK_full
- **vcpkg Location:** Z:\src\vcpkg
- **grpc Version:** 1.71.0
- **protobuf Version:** 29.5.0
- **getopt-win32 Version:** 2.42.0

## References

- [llama.cpp Repository](https://github.com/ggml-org/llama.cpp)
- [vcpkg Documentation](https://github.com/Microsoft/vcpkg)
- [gRPC C++ Documentation](https://grpc.io/docs/languages/cpp/)
- [LocalAI Documentation](https://localai.io)
