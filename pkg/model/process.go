package model

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/mudler/LocalAI/pkg/signals"
	process "github.com/mudler/go-processmanager"
	"github.com/mudler/xlog"
)

var forceBackendShutdown bool = os.Getenv("LOCALAI_FORCE_BACKEND_SHUTDOWN") == "true"

// findBashOnWindows locates a bash shell on Windows via PATH.
// Returns the resolved path to "bash" or "bash" if not found.
func findBashOnWindows() string {
	if p, err := exec.LookPath("bash"); err == nil {
		return p
	}
	// Fallback when PATH lookup fails
	return "bash"
}

func (ml *ModelLoader) deleteProcess(s string) error {
	model, ok := ml.models[s]
	if !ok {
		xlog.Debug("Model not found", "model", s)
		return fmt.Errorf("model %s not found", s)
	}

	defer delete(ml.models, s)

	retries := 1
	for model.GRPC(false, ml.wd).IsBusy() {
		xlog.Debug("Model busy. Waiting.", "model", s)
		dur := time.Duration(retries*2) * time.Second
		if dur > retryTimeout {
			dur = retryTimeout
		}
		time.Sleep(dur)
		retries++

		if retries > 10 && forceBackendShutdown {
			xlog.Warn("Model is still busy after retries. Forcing shutdown.", "model", s, "retries", retries)
			break
		}
	}

	xlog.Debug("Deleting process", "model", s)

	process := model.Process()
	if process == nil {
		xlog.Error("No process", "model", s)
		// Nothing to do as there is no process
		return nil
	}

	err := process.Stop()
	if err != nil {
		xlog.Error("(deleteProcess) error while deleting process", "error", err, "model", s)
	}

	return err
}

func (ml *ModelLoader) StopGRPC(filter GRPCProcessFilter) error {
	var err error = nil
	ml.mu.Lock()
	defer ml.mu.Unlock()

	for k, m := range ml.models {
		if filter(k, m.Process()) {
			e := ml.deleteProcess(k)
			err = errors.Join(err, e)
		}
	}
	return err
}

func (ml *ModelLoader) StopAllGRPC() error {
	return ml.StopGRPC(all)
}

func (ml *ModelLoader) GetGRPCPID(id string) (int, error) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	p, exists := ml.models[id]
	if !exists {
		return -1, fmt.Errorf("no grpc backend found for %s", id)
	}
	if p.Process() == nil {
		return -1, fmt.Errorf("no grpc backend found for %s", id)
	}
	return strconv.Atoi(p.Process().PID)
}

func (ml *ModelLoader) startProcess(grpcProcess, id string, serverAddress string, args ...string) (*process.Process, error) {
	// Make sure the process is executable (POSIX only)
	if runtime.GOOS != "windows" {
		if fi, err := os.Stat(grpcProcess); err == nil {
			if fi.Mode()&0111 == 0 {
				xlog.Debug("Process is not executable. Making it executable.", "process", grpcProcess)
				if err := os.Chmod(grpcProcess, 0700); err != nil {
					return nil, err
				}
			}
		}
	}

	xlog.Debug("Loading GRPC Process", "process", grpcProcess)

	xlog.Debug("GRPC Service will be running", "id", id, "address", serverAddress)

	workDir, err := filepath.Abs(filepath.Dir(grpcProcess))
	if err != nil {
		return nil, err
	}

	// If this is a shell script on Windows, wrap it with a shell
	processName := filepath.Base(grpcProcess)
	processArgs := append([]string{}, args...)
	bindAddr := serverAddress
	if runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(grpcProcess), ".sh") {
		processName = findBashOnWindows()
		raw := strings.ReplaceAll(grpcProcess, "\\", "/")
		var converted string
		shellKind := "bash"
		lowerName := strings.ToLower(processName)
		// Detect legacy WSL shim bash.exe in System32
		if strings.Contains(lowerName, "\\system32\\bash.exe") || strings.HasSuffix(lowerName, "system32/bash.exe") {
			shellKind = "wsl-bash"
			if len(raw) >= 2 && raw[1] == ':' {
				converted = "/mnt/" + strings.ToLower(string(raw[0])) + raw[2:]
			} else {
				converted = raw
			}
			// bash.exe (WSL shim) directly executes the script
			processArgs = append([]string{converted}, args...)
			// For WSL, bind to all interfaces to ensure Windows can dial
			bindAddr = strings.Replace(bindAddr, "127.0.0.1", "0.0.0.0", 1)
		} else {
			// Git Bash/MSYS path style: /<drive>/...
			if len(raw) >= 2 && raw[1] == ':' {
				converted = "/" + strings.ToLower(string(raw[0])) + raw[2:]
			} else {
				converted = raw
			}
			// bash.exe (Git/MSYS) directly executes the script
			processArgs = append([]string{converted}, args...)
		}
		xlog.Debug("Windows shell script launch", "shellKind", shellKind, "shellPath", processName, "scriptPath", grpcProcess, "convertedPath", converted, "bindAddr", bindAddr, "processArgs", processArgs)
	}

	grpcControlProcess := process.New(
		process.WithTemporaryStateDir(),
		process.WithName(processName),
		process.WithArgs(append(processArgs, []string{"--addr", bindAddr}...)...),
		process.WithEnvironment(os.Environ()...),
		process.WithWorkDir(workDir),
	)

	xlog.Debug("Process configuration", "name", processName, "workDir", workDir, "finalArgs", append(processArgs, []string{"--addr", bindAddr}...))

	if ml.wd != nil {
		ml.wd.Add(serverAddress, grpcControlProcess)
		ml.wd.AddAddressModelMap(serverAddress, id)
	}

	if err := grpcControlProcess.Run(); err != nil {
		return grpcControlProcess, err
	}

	xlog.Debug("GRPC Service state dir", "dir", grpcControlProcess.StateDir())

	signals.RegisterGracefulTerminationHandler(func() {
		err := grpcControlProcess.Stop()
		if err != nil {
			xlog.Error("error while shutting down grpc process", "error", err)
		}
	})

	go func() {
		t, err := tail.TailFile(grpcControlProcess.StderrPath(), tail.Config{Follow: true})
		if err != nil {
			xlog.Debug("Could not tail stderr")
		}
		for line := range t.Lines {
			xlog.Debug("GRPC stderr", "id", strings.Join([]string{id, serverAddress}, "-"), "line", line.Text)
		}
	}()
	go func() {
		t, err := tail.TailFile(grpcControlProcess.StdoutPath(), tail.Config{Follow: true})
		if err != nil {
			xlog.Debug("Could not tail stdout")
		}
		for line := range t.Lines {
			xlog.Debug("GRPC stdout", "id", strings.Join([]string{id, serverAddress}, "-"), "line", line.Text)
		}
	}()

	return grpcControlProcess, nil
}
