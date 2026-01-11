--- This is a FIRST PASS at building for Windows with support for LocalAI .. this will be cleaned up more ---

# Building Stable Diffusion GGML with Vulkan on Windows

## Overview
This guide documents building the Stable Diffusion GGML backend for Windows with Vulkan support using MSVC 14.50 and CMake.

## Prerequisites
- Visual Studio 2022/2026 Professional with MSVC 14.50 (v145 toolset)
- CMake 3.24+
- Go 1.20+
- Vulkan SDK
- Git with submodule support

## Build Environment
- **OS**: Windows 10/11 x64
- **Compiler**: MSVC 14.50.35717 (v145 toolset)
- **Build Path**: `g:\b\sdggml` (short path to avoid Windows 260-char limit)
- **Output**:
  - C++ DLL: `g:\b\sdggml\Release\gosd.dll`
  - Go Wrapper: `g:\LocalAI\backend\go\stablediffusion-ggml\stablediffusion-ggml.exe`

## Step 1: Clone Repository and Initialize Submodules

```bash
cd g:\LocalAI\backend\go\stablediffusion-ggml
git clone https://github.com/leejet/stable-diffusion.cpp.git sources/stablediffusion-ggml.cpp
cd sources/stablediffusion-ggml.cpp
git submodule update --init --recursive
# Or manually clone ggml:
git clone https://github.com/ggerganov/ggml.git ggml
cd ggml
git checkout 3e9f2ba3  # Correct commit for compatibility
```

## Step 2: Apply Critical Windows Fixes

Before configuring CMake, apply these fixes to avoid C runtime linking errors:

### Fix 1: Zip Library C Runtime (CRITICAL)
Edit `sources/stablediffusion-ggml.cpp/thirdparty/CMakeLists.txt`:

```cmake
set(Z_TARGET zip)
add_library(${Z_TARGET} OBJECT zip.c zip.h miniz.h)
target_include_directories(${Z_TARGET} PUBLIC .)

# Fix Windows C runtime linking to use static CRT
if(MSVC)
    target_compile_options(${Z_TARGET} PRIVATE /MT$<$<CONFIG:Debug>:d>)
    target_compile_definitions(${Z_TARGET} PRIVATE _CRT_SECURE_NO_WARNINGS)
endif()
```

**Why**: Without this, zip.obj will have unresolved `__imp__wstat64`, `__imp__localtime64_s`, etc. symbols causing link failures.

### Fix 2: GOSD Wrapper CMakeLists
Edit `CMakeLists.txt` in the wrapper directory:

```cmake
add_library(gosd MODULE gosd.cpp)
target_link_libraries(gosd PRIVATE stable-diffusion ggml)

if(CMAKE_CXX_COMPILER_ID MATCHES "GNU" AND CMAKE_CXX_COMPILER_VERSION VERSION_LESS 9.0)
    target_link_libraries(gosd PRIVATE stdc++fs)
endif()

# Windows security warnings suppression
if(MSVC)
    target_compile_definitions(gosd PRIVATE _CRT_SECURE_NO_WARNINGS)
endif()
```

## Step 3: Configure CMake with Vulkan and MSVC

**IMPORTANT**: Configure source (-S) from the **wrapper directory** (stablediffusion-ggml), not the upstream sources:

```bash
cd g:\b\sdggml

cmake -G "Visual Studio 17 2022" `
  -T v145 `
  -DCMAKE_BUILD_TYPE=Release `
  -DSD_VULKAN=ON `
  -DGGML_VULKAN=ON `
  -DGGML_BLAS=OFF `
  -DBUILD_SHARED_LIBS=OFF `
  -DCMAKE_CXX_FLAGS_RELEASE="/O2 /DNDEBUG" `
  -S "g:\LocalAI\backend\go\stablediffusion-ggml" `
  -B g:\b\sdggml
```

**Key Flags**:
- `-T v145`: Use MSVC v145 toolset (compatible with VS 2022+)
- `-DSD_VULKAN=ON`: Enable Vulkan backend in stable-diffusion.cpp
- `-DGGML_VULKAN=ON`: Enable Vulkan acceleration in GGML
- `-DGGML_BLAS=OFF`: **Disable BLAS** (not available on Windows, causes build failures)
- `-DBUILD_SHARED_LIBS=OFF`: Build as static libraries for easier linking
- `-S` and `-B`: Specify source and build directories explicitly

**Note**: After applying the fixes above, a **clean rebuild is required** to recompile zip.obj with `/MT`.

## Step 4: Fix DLL Exports (Windows-Specific)

When exporting C functions from a C++ DLL, ensure proper syntax in both header and implementation:

### In `gosd.h`:
```cpp
extern __declspec(dllexport) int load_model(...);
extern __declspec(dllexport) int gen_image(...);
extern __declspec(dllexport) void sd_img_gen_params_set_prompts(...);
// ... etc for all exported functions
```

### In `gosd.cpp`:
```cpp
extern "C" __declspec(dllexport) int load_model(...) {
    // implementation
}

extern "C" __declspec(dllexport) int gen_image(...) {
    // implementation
}
// ... etc for all exported functions
```

**CRITICAL**: The correct order is `extern "C"` **first**, then `__declspec(dllexport)`. Reversed order causes C4502 warnings and symbol visibility failures.

## Step 5: Build the C++ Library

```bash
cmake --build g:\b\sdggml --config Release --target gosd -j 16
```

**Expected output**:
```
gosd.vcxproj -> G:\b\sdggml\Release\gosd.dll
```

**Result**:
- `g:\b\sdggml\Release\gosd.dll` (typically ~75 MB with full Vulkan support)
- Required dependencies: `stable-diffusion.lib`, `ggml.lib`, `ggml-vulkan.lib`

**Note**: First build may take 15-20 minutes. Expect harmless POSIX deprecation warnings for `strdup`.

## Step 6: Build Go Wrapper

Copy the DLL and compile the Go wrapper:

```bash
cd g:\LocalAI\backend\go\stablediffusion-ggml

# Copy the compiled DLL
copy g:\b\sdggml\Release\gosd.dll .\gosd.dll

# Build the Go executable
go build -o stablediffusion-ggml.exe .
```

**Note**: Ensure the Go loader uses `syscall.LoadLibrary` (see Windows-Specific Considerations) before building.

## Step 7: Verify Exports

Use `dumpbin` to verify all C functions are exported:

```powershell
dumpbin /exports g:\b\sdggml\Release\gosd.dll | findstr "load_model gen_image sd_"
```

Expected output (sample):
```
          1    0 00039DC0 gen_image
          2    1 0003AC30 load_model
          3    2 0003C130 sd_img_gen_params_get_vae_tiling_params
          4    3 0003C140 sd_img_gen_params_new
          ...
```

## Step 8: Test the Executable

```bash
cd g:\LocalAI\backend\go\stablediffusion-ggml
.\stablediffusion-ggml.exe
```

Should show gRPC server starting (not "procedure could not be found" errors).

## Windows-Specific Considerations

### 1. Zip Library Linker Errors

**Symptoms**: 7 unresolved external symbols during link:
```
error LNK2019: unresolved external symbol __imp__wstat64
error LNK2019: unresolved external symbol __imp__localtime64_s
error LNK2019: unresolved external symbol __imp__mktime64
error LNK2019: unresolved external symbol __imp__wmkdir
error LNK2019: unresolved external symbol __imp__wfreopen_s
error LNK2019: unresolved external symbol __imp_remove
error LNK2019: unresolved external symbol __imp__utime64
```

**Root Cause**: zip.obj compiled with DLL import mode (default) creates `__imp_` prefixed symbols that don't exist in static CRT libraries.

**Solution**: Already applied in Step 2 (Fix 1). Forces zip to use `/MT` (static CRT):
```cmake
if(MSVC)
    target_compile_options(${Z_TARGET} PRIVATE /MT$<$<CONFIG:Debug>:d>)
    target_compile_definitions(${Z_TARGET} PRIVATE _CRT_SECURE_NO_WARNINGS)
endif()
```

**Important**: After applying this fix, you **must** clean rebuild:
```powershell
Remove-Item g:\b\sdggml\* -Recurse -Force
cmake -G "Visual Studio 17 2022" ... # Reconfigure
cmake --build g:\b\sdggml --config Release --target gosd -j 16
```

### 2. Long Path Issues
Windows has a 260-character path limit. Use a short build directory:
```bash
# ✓ Good
cmake ... g:\b\sdggml

# ✗ Bad (may exceed limit)
cmake ... "C:\Users\Username\Projects\LocalAI\backend\go\stablediffusion-ggml\..."
```

### 3. DLL Loading with Go
Use `syscall.LoadLibrary` for Windows (not `purego.Dlopen`):
```go
import "syscall"

dll, err := syscall.LoadLibrary("gosd.dll")
if err != nil {
    panic(fmt.Sprintf("Failed to load DLL: %v", err))
}
```

### 3. Symbol Export Syntax
Windows C++ requires explicit export decorators. Missing `__declspec(dllexport)` on C function definitions causes "procedure could not be found" runtime errors.

### 4. Vulkan SDK Setup
Ensure Vulkan SDK is in PATH:
```powershell
# Verify VulkanSDK environment variable is set
echo $env:VULKAN_SDK

# Example (adjust version):
# C:\VulkanSDK\1.3.xxx\
```

## Troubleshooting

### Error: "The specified procedure could not be found"
**Cause**: C functions not exported from DLL.
**Solution**: Add `extern "C" __declspec(dllexport)` to all exported function definitions and declarations.

### Error: "CMake Error at glsl.cmake:45 (message): sh is not installed"
**Cause**: Shader generation requires Unix tools.
**Solution**: Use short build path or WSL for this step.

### Error: "visual studio not found" or toolset version mismatch
**Solution**: Specify correct toolset:
```bash
cmake -G "Visual Studio 17 2022" -T v145 ...
```

## Output Artifacts

After successful build:
- **DLL**: `g:\b\sdggml\Release\gosd.dll` (~75 MB with Vulkan support)
- **Executable**: `g:\LocalAI\backend\go\stablediffusion-ggml\stablediffusion-ggml.exe` (~16 MB)
- **Dependencies**: All compiled into the DLL (self-contained)

## Performance Notes

- **Vulkan Acceleration**: Significantly faster image generation on compatible GPUs
- **BLAS Support**: Optimizes CPU operations when GPU unavailable
- **Build Configuration**: Release mode with optimizations (`/O2`)

## References

- Stable Diffusion GGML: https://github.com/leejet/stable-diffusion.cpp
- GGML: https://github.com/ggerganov/ggml
- Vulkan SDK: https://vulkan.lunarg.com/
