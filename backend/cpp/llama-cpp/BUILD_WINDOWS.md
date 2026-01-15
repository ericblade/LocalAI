# Building grpc-server.exe on Windows

## Quick Start (TL;DR)

```powershell
# 1. Install dependencies via vcpkg
Z:\src\vcpkg\vcpkg install grpc:x64-windows protobuf:x64-windows getopt-win32:x64-windows

# 2. Clone and checkout pinned llama.cpp
cd G:\LocalAI\backend\cpp\llama-cpp
git clone https://github.com/ggml-org/llama.cpp.git
cd llama.cpp
git checkout e4832e3ae4d58ac0ecbdbf4ae055424d6e628c9f

# 3. Run prepare script (copies files and applies fixes)
cd ..
bash ./prepare.sh  # Or manual steps if no bash (see Step 2)

# 4. Configure and build
cd llama.cpp
mkdir build; cd build
& 'C:\Program Files\Microsoft Visual Studio\2022\Professional\Common7\Tools\Launch-VsDevShell.ps1'
cmake .. -G "Visual Studio 17 2022" -DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows `
  -DGGML_VULKAN=1 -DBUILD_SHARED_LIBS=OFF -DLLAMA_CURL=ON -DLLAMA_BUILD_SERVER=ON
cmake --build . --config Release --target grpc-server -j 24

# 5. Deploy
Copy-Item "bin\Release\grpc-server.exe" -Destination "..\..\grpc-server.exe"
Copy-Item "Z:\src\vcpkg\installed\x64-windows\bin\*.dll" -Destination "..\..\"
```

**Build Output:** `G:\LocalAI\backend\cpp\llama-cpp\grpc-server.exe`

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
git checkout e4832e3ae4d58ac0ecbdbf4ae055424d6e628c9f  # Pin to tested commit
```

**Important:** Use the pinned commit from `Makefile` to ensure compatibility. The grpc-server implementation may break with newer llama.cpp versions.

### 2. Prepare Build Environment

**Option A: Using Bash (Git Bash, WSL, or MSYS2)**

If you have bash available (Git Bash, WSL, or MSYS2):

```powershell
cd G:\LocalAI\backend\cpp\llama-cpp
bash ./prepare.sh
```

The `prepare.sh` script will:
1. Copy all files from `llama.cpp/tools/server/` to `llama.cpp/tools/grpc-server/`
2. Overwrite with LocalAI's custom `CMakeLists.txt`
3. Overwrite with LocalAI's custom `grpc-server.cpp`
4. Copy vendor files (json.hpp, httplib.h)
5. Add `grpc-server` subdirectory to `llama.cpp/tools/CMakeLists.txt`

**Option B: Manual Setup (Windows-only, no Bash)**

If bash is not available, perform these steps manually:

```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp

# 1. Copy all server files to grpc-server directory
Copy-Item -Recurse -Force tools\server\* tools\grpc-server\

# 2. Copy LocalAI's custom CMakeLists.txt
Copy-Item -Force ..\CMakeLists.txt tools\grpc-server\CMakeLists.txt

# 3. Copy LocalAI's custom grpc-server.cpp
Copy-Item -Force ..\grpc-server.cpp tools\grpc-server\grpc-server.cpp

# 4. Copy vendor files
Copy-Item -Force vendor\nlohmann\json.hpp tools\grpc-server\json.hpp
Copy-Item -Force vendor\cpp-httplib\httplib.h tools\grpc-server\httplib.h

# 5. Add grpc-server to tools CMakeLists.txt (if not already present)
$cmakeContent = Get-Content tools\CMakeLists.txt -Raw
if ($cmakeContent -notmatch "grpc-server") {
    Add-Content tools\CMakeLists.txt "`n    add_subdirectory(grpc-server)"
}
```

**Why this is necessary:**
- grpc-server needs all the server infrastructure files (server-context.cpp, server-task.cpp, etc.)
- LocalAI's CMakeLists.txt has Windows-specific fixes for protoc and getopt
- LocalAI's grpc-server.cpp implements the gRPC protocol instead of HTTP

### 3. Create Build Directory
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp
mkdir build -ErrorAction SilentlyContinue
cd build
```

### 4. Run CMake Configuration
```powershell
# Launch Visual Studio Developer PowerShell (or run from VS Dev Command Prompt)
& 'C:\Program Files\Microsoft Visual Studio\2022\Professional\Common7\Tools\Launch-VsDevShell.ps1'

# Configure with CMake
cmake .. `
  -G "Visual Studio 17 2022" `
  -DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows `
  -DGGML_VULKAN=1 `
  -DBUILD_SHARED_LIBS=OFF `
  -DLLAMA_CURL=ON `
  -DLLAMA_BUILD_SERVER=ON
```

**Key Flags:**
- `-G "Visual Studio 17 2022"` - Use VS 2022 generator (adjust version if different, e.g., "Visual Studio 18 2026")
- `-DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows` - Tell CMake where to find vcpkg packages
- `-DGGML_VULKAN=1` - Enable Vulkan GPU support (optional, omit for CPU-only)
- `-DBUILD_SHARED_LIBS=OFF` - Build static libraries (simpler deployment)
- `-DLLAMA_CURL=ON` - Enable curl for model downloads
- `-DLLAMA_BUILD_SERVER=ON` - **Required** to build server tools including grpc-server

**Note:** Adjust Visual Studio path if using Community Edition:
```powershell
& 'C:\Program Files\Microsoft Visual Studio\2022\Community\Common7\Tools\Launch-VsDevShell.ps1'
```

### 5. Build grpc-server Target
```powershell
cmake --build . --config Release --target grpc-server -j 24
```

**Parallel Build:** Adjust `-j 24` to match your CPU core count for faster builds.

**Build Output:**
- Binary: `llama.cpp\build\bin\Release\grpc-server.exe`
- Size: ~60-70 MB (Release build with Vulkan)
- Build time: ~5-15 minutes depending on hardware

**Common Build Warnings (safe to ignore):**
- C4244, C4267: Type conversion warnings (size_t → int)
- C4101: Unreferenced local variable 'e' in exception handlers
- C4251: DLL interface warnings for protobuf/gRPC classes

### 6. Deploy Binary and Dependencies

#### Copy grpc-server.exe
```powershell
Copy-Item "bin\Release\grpc-server.exe" `
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

The build incorporates custom implementations for Windows compatibility:

### 1. CMakeLists.txt Modifications
**Location:** `backend/cpp/llama-cpp/CMakeLists.txt` (copied to `llama.cpp/tools/grpc-server/`)

**Key Changes:**
- **Protoc Resolution:** Uses `$<TARGET_FILE:protobuf::protoc>` instead of `find_program(protoc)`
  - Fixes "protoc not found" errors on Windows
  - Ensures CMake uses the vcpkg-installed protoc at build time

- **gRPC Plugin Resolution:** Uses `$<TARGET_FILE:gRPC::grpc_cpp_plugin>` instead of `find_program(grpc_cpp_plugin)`
  - Ensures correct gRPC code generator is used

- **Windows getopt Support:**
  ```cmake
  find_package(getopt CONFIG)
  if(WIN32 AND TARGET getopt::getopt_shared)
    target_link_libraries(${TARGET} PRIVATE getopt::getopt_shared)
  endif()
  ```
  - Links getopt-win32 library for command-line parsing on Windows

### 2. grpc-server.cpp Implementation
**Location:** `backend/cpp/llama-cpp/grpc-server.cpp`

**Key Features:**
- **gRPC Protocol:** Implements `Backend` service from `backend.proto` instead of HTTP REST API
- **JSON Serialization Fix:** Uses `.dump()` to convert nlohmann::json to string for protobuf
  - Line 2236: `reply->set_message(arr.dump())` instead of `reply->set_message(arr)`
- **Auto-fit Context:** Initializes `tensor_buft_overrides` for automatic context size fitting
- **Windows Compatibility:** Includes getopt header and proper DLL handling

### 3. Why prepare.sh is Required
The `prepare.sh` script (or manual equivalent) is necessary because:
1. **Server Infrastructure:** grpc-server needs all the llama.cpp server files:
   - `server-context.cpp/h` - Core server state management
   - `server-task.cpp/h` - Task queue and execution
   - `server-common.cpp/h` - Shared utilities
   - `server-queue.cpp/h` - Request queue management

2. **File Overwrite Order:** Server files must be copied *before* LocalAI customizations:
   ```bash
   # 1. Copy all upstream server files
   cp -r llama.cpp/tools/server/* llama.cpp/tools/grpc-server/
   # 2. Overwrite with LocalAI customizations
   cp CMakeLists.txt llama.cpp/tools/grpc-server/
   cp grpc-server.cpp llama.cpp/tools/grpc-server/
   ```

3. **Vendor Dependencies:** Copies required header-only libraries:
   - `nlohmann/json.hpp` - JSON serialization
   - `cpp-httplib/httplib.h` - HTTP utilities (used by server-common.cpp)

## Troubleshooting

### Build Fails: "CMake Error: Could not find CMAKE_ROOT"
- Ensure you're running from Visual Studio Developer PowerShell
- Run `Launch-VsDevShell.ps1` before CMake commands

### Build Fails: "_PROTOBUF_PROTOC-NOTFOUND"
**Symptom:** `'_PROTOBUF_PROTOC-NOTFOUND' is not recognized as an internal or external command`

**Cause:** CMakeLists.txt is using `find_program` instead of target-based protoc resolution

**Fix:** Ensure you ran `prepare.sh` or manually copied the corrected `CMakeLists.txt`:
```powershell
Copy-Item "G:\LocalAI\backend\cpp\llama-cpp\CMakeLists.txt" `
  -Destination "G:\LocalAI\backend\cpp\llama-cpp\llama.cpp\tools\grpc-server\CMakeLists.txt" `
  -Force
```

Then reconfigure:
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp\build
cmake .. -G "Visual Studio 17 2022" -DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows `
  -DGGML_VULKAN=1 -DBUILD_SHARED_LIBS=OFF -DLLAMA_CURL=ON -DLLAMA_BUILD_SERVER=ON
```

### Build Fails: "grpc_cpp_plugin not found"
Same as above - ensure the corrected CMakeLists.txt is in place.

### Build Fails: Getopt Linking Errors
**Symptom:** `unresolved external symbol getopt` or `LNK2019` errors

**Fix:** Ensure getopt-win32 is installed and CMakeLists.txt has:
```cmake
find_package(getopt CONFIG)
if(WIN32 AND TARGET getopt::getopt_shared)
  target_link_libraries(${TARGET} PRIVATE getopt::getopt_shared)
endif()
```

### Build Fails: "common_remote_get_content not found"
**Symptom:** `error C3861: 'common_remote_get_content': identifier not found`

**Cause:** Version mismatch - llama.cpp has been updated beyond the pinned commit

**Fix:** Checkout the pinned commit from Makefile:
```powershell
cd llama.cpp
git fetch origin
git checkout e4832e3ae4d58ac0ecbdbf4ae055424d6e628c9f
cd build
# Clean and rebuild
cmake --build . --config Release --target grpc-server --clean-first
```

### grpc-server.exe Not Found After Build
**Symptom:** Executable not in `tools\grpc-server\Release\grpc-server.exe`

**Location:** Check `bin\Release\grpc-server.exe` instead - the build output path changed in recent CMake versions.

### Build Fails: "protoc not found" (despite vcpkg installation)
- Ensure PATH includes `Z:\src\vcpkg\installed\x64-windows\tools\protobuf`
- Verify `protoc.exe` exists in that directory
- **Better:** Use the corrected CMakeLists.txt which doesn't rely on PATH

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

For subsequent builds (after upstream updates or code changes):

### Quick Rebuild (No Source Changes)
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp\build
cmake --build . --config Release --target grpc-server -j 24
Copy-Item "bin\Release\grpc-server.exe" -Destination "..\..\grpc-server.exe" -Force
```

### Full Rebuild (After llama.cpp Update)
```powershell
cd G:\LocalAI\backend\cpp\llama-cpp\llama.cpp
git fetch origin
git checkout e4832e3ae4d58ac0ecbdbf4ae055424d6e628c9f  # Pin to tested commit

# Re-run prepare script
cd ..
bash ./prepare.sh  # Or manual steps if no bash

# Clean build
cd llama.cpp\build
Remove-Item -Recurse -Force * -ErrorAction SilentlyContinue
cmake .. -G "Visual Studio 17 2022" -DCMAKE_PREFIX_PATH=Z:\src\vcpkg\installed\x64-windows `
  -DGGML_VULKAN=1 -DBUILD_SHARED_LIBS=OFF -DLLAMA_CURL=ON -DLLAMA_BUILD_SERVER=ON
cmake --build . --config Release --target grpc-server -j 24
Copy-Item "bin\Release\grpc-server.exe" -Destination "..\..\grpc-server.exe" -Force
```

### After Modifying grpc-server.cpp or CMakeLists.txt
```powershell
# Re-run prepare to copy your changes
cd G:\LocalAI\backend\cpp\llama-cpp
bash ./prepare.sh

# Rebuild
cd llama.cpp\build
cmake --build . --config Release --target grpc-server -j 24 --clean-first
Copy-Item "bin\Release\grpc-server.exe" -Destination "..\..\grpc-server.exe" -Force
```

**Note:** Always use the pinned commit from `Makefile`. Newer llama.cpp commits may have API changes incompatible with LocalAI's grpc-server.cpp.

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
