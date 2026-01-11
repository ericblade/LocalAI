# Building Whisper.cpp with Vulkan on Windows

This guide explains how to build the Whisper backend with Vulkan GPU acceleration on Windows.

## Prerequisites

- **OS**: Windows 10/11 x64
- **Compiler**: MSVC 19.50+ (Visual Studio 2022+)
- **CMake**: 3.12 or later
- **Vulkan SDK**: 1.4.0 or later - [Download](https://vulkan.lunarg.com/sdk/home)
- **Git**: For cloning repositories

### Vulkan SDK Installation

1. Download and install the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home)
2. The installer will set up the required environment variables (VULKAN_SDK)
3. Verify installation: The Vulkan SDK should be in `C:\VulkanSDK\1.x.x\` or similar

## Building Steps

### Step 1: Clone and Setup

Clone whisper.cpp with the pinned version:

```powershell
cd g:\LocalAI\backend\go\whisper
mkdir -p sources/whisper.cpp
cd sources/whisper.cpp
git init
git remote add origin https://github.com/ggml-org/whisper.cpp
git fetch origin 679bdb53dbcbfb3e42685f50c7ff367949fd4d48
git checkout 679bdb53dbcbfb3e42685f50c7ff367949fd4d48
git submodule update --init --recursive --depth 1 --single-branch
cd ..\..
```

**Important**: Use a short path for the build directory to avoid Windows path length issues (260 char limit). This guide uses `g:\b\whisper` instead of the full path.

### Step 2: Launch Visual Studio Developer Shell

This is critical - the build requires the VS development environment with proper compiler and linker setup:

```powershell
& "Z:\program files\microsoft visual studio\18\professional\Common7\tools\Launch-VsDevShell.ps1"
```

### Step 3: Configure CMake

From within the VS Dev Shell, configure with short path:

```powershell
mkdir g:\b\whisper
cd g:\b\whisper
cmake g:\LocalAI\backend\go\whisper -G "Visual Studio 18 2026" -DGGML_VULKAN=ON
```

**Key options**:
- `-DGGML_VULKAN=ON` - Enable Vulkan GPU acceleration
- Detector will find Vulkan SDK automatically
- Will verify all Vulkan extensions: cooperative matrix, coopmat2, integer dot product, bfloat16

### Step 4: Build the C++ Library

Still in the VS Dev Shell:

```powershell
cd g:\b\whisper
cmake --build . --config Release --parallel 4
```

This will:
- Compile GGML base and CPU backends
- Generate Vulkan shaders (this takes several minutes)
- Compile whisper.cpp with Vulkan support
- Output: `g:\b\whisper\Release\gowhisper.dll` (~40-50MB with Vulkan shaders embedded)

**Build time**: ~10-15 minutes (mostly shader compilation). Vulkan shader generation runs in parallel with GGML compilation.

### Step 5: Build the Go Wrapper

Still in the same VS Dev Shell (environment is already configured):

```powershell
cd g:\LocalAI\backend\go\whisper
$env:WHISPER_LIBRARY="g:\b\whisper\Release\gowhisper.dll"
& "z:\tools\go\bin\go.exe" build -o whisper.exe ./
```

This will:
- Compile `lib_windows.go` (Windows DLL loader using syscall)
- Compile `lib_unix.go` (Linux/macOS loader using purego - excluded on Windows)
- Generate platform-independent `whisper.exe` (17-20MB)

**Platform independence**: The Go wrapper uses build tags to select the correct library loader automatically based on the target platform. No conditional compilation needed in main.go.

## Configuration

When running the Whisper service, set the library path to the Vulkan DLL:

```bash
$env:WHISPER_LIBRARY="g:\b\whisper\Release\gowhisper.dll"
.\whisper.exe --addr 0.0.0.0:50052
```

Or in a batch file:

```batch
@echo off
set WHISPER_LIBRARY=g:\b\whisper\Release\gowhisper.dll
.\whisper.exe --addr 0.0.0.0:50052
```

## Platform Independence

The Go wrapper uses build tags for platform-specific library loading:

- **Windows** (`lib_windows.go`): Uses `syscall.LoadLibrary()` for DLL loading
- **Unix/Linux/macOS** (`lib_unix.go`): Uses `purego.Dlopen()` for .so loading

The build system automatically selects the correct file based on the target platform. Main.go calls the generic `loadWhisperLib()` function, which dispatches to the platform-specific implementation at compile time.

## Troubleshooting

### Path Length Issues (Windows 260 Character Limit)

**Problem**: Build fails with "path too long" errors during shader generation
- Error: `System.IO.DirectoryNotFoundException: Could not find a part of the path`

**Solution**: Use a short build path like `g:\b\whisper` instead of the full path.

**Why**: Vulkan shader generation creates nested directories for intermediate build artifacts. The full path `G:\LocalAI\backend\go\whisper\build-vulkan\sources\whisper.cpp\ggml\src\ggml-vulkan\vulkan-shaders-gen-prefix\src\vulkan-shaders-gen-build\CMakeFiles\CMakeScratch\...` quickly exceeds Windows' 260-character PATH_MAX limit.

### Missing VS Dev Environment

**Problem**: `cmake`, `cl.exe`, or `link.exe` not found
- Error: `The term 'cmake' is not recognized`

**Solution**: Always launch the VS Developer PowerShell first:
```powershell
& "Z:\program files\microsoft visual studio\18\professional\Common7\tools\Launch-VsDevShell.ps1"
```

**Why**: The dev shell sets up environment variables (INCLUDE, LIB, PATH) needed for compilation. Running cmake from a regular PowerShell won't find the MSVC toolchain.

### Go Executable Not Found

**Problem**: `go: The term 'go' is not recognized`
- The VS Dev Shell doesn't include Go in its PATH

**Solution**: Use full path to Go:
```powershell
& "z:\tools\go\bin\go.exe" build -o whisper.exe ./
```

### CMake Generator Mismatch

**Problem**: `CMake Error: Could not create named generator Visual Studio 18 2025`
- The correct generator for VS 2026 is `Visual Studio 18 2026`, not 2025

**Solution**: Use the correct year matching your VS installation:
```powershell
cmake ... -G "Visual Studio 18 2026" ...
```

### DLL Not Found at Runtime

**Problem**: `cannot open shared object file` or similar when running whisper.exe

**Solution**: Verify the DLL path and set WHISPER_LIBRARY:
```powershell
$env:WHISPER_LIBRARY="g:\b\whisper\Release\gowhisper.dll"
.\whisper.exe --addr 0.0.0.0:50052
```

### "The specified procedure could not be found" Error

**Problem**: 
```
panic: The specified procedure could not be found.
goroutine 1 [running]:
github.com/ebitengine/purego.RegisterLibFunc(...)
```

**Root Cause**: The C functions in `gowhisper.cpp` weren't exported from the DLL. On Windows, functions in a DLL must be explicitly marked with `__declspec(dllexport)`.

**Solution**: Ensure `gowhisper.h` and `gowhisper.cpp` include the GOWHISPER_EXPORT macro:

In `gowhisper.h`:
```cpp
#ifdef _WIN32
#define GOWHISPER_EXPORT extern "C" __declspec(dllexport)
#else
#define GOWHISPER_EXPORT
#endif
```

All function declarations must have GOWHISPER_EXPORT on a separate line:
```cpp
GOWHISPER_EXPORT
int load_model(const char *const model_path);
```

And all function definitions in `gowhisper.cpp` must have GOWHISPER_EXPORT on a separate line:
```cpp
GOWHISPER_EXPORT
int load_model(const char *const model_path) {
  // implementation
}
```

After adding/fixing the exports, rebuild the DLL with `cmake --build . --config Release --parallel 4` in the VS Dev Shell.

## Build Performance

**Shader Compilation**: The first build takes 10-15 minutes because Vulkan requires compiling ~100+ compute shaders for different operations. Subsequent builds are faster if only source code changes (shaders are cached).

**Parallel Build**: Use `cmake --build . --config Release --parallel 4` to compile multiple GGML kernels and shaders in parallel. Adjust the number based on available CPU cores.

**DLL Size**: With Vulkan shaders embedded, the DLL is 40-50MB. This is normal and includes all shader bytecode needed for runtime compilation.

## Performance Notes

- **GPU Memory**: Whisper models require 500MB - 3GB depending on model size
- **CPU Fallback**: If Vulkan initialization fails, models will fall back to CPU inference
- **First Run**: First run with Vulkan may be slow as shaders are compiled
- **Optimal Resolution**: Whisper processes 16kHz mono audio efficiently on GPU

## Model Files

Download Whisper models from [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp):

```bash
cd models
# Download base model (optimal performance/accuracy)
curl -L -o ggml-base.en.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin
```

## Testing

Test the build with a sample audio file:

```bash
.\whisper.exe -m models/ggml-base.en.bin -f samples/jfk.wav
```

Or via gRPC (when running as LocalAI backend):

```bash
# In another terminal
grpcurl -plaintext -d '{"model":"base.en","audio_path":"samples/jfk.wav"}' \
  localhost:50052 whisper.Whisper/Transcribe
```

## Next Steps

- Deploy to LocalAI by copying whisper.exe to the appropriate backend directory
- Configure external_backends.json to route audio transcription to this service
- Set WHISPER_LIBRARY environment variable in LocalAI startup scripts
