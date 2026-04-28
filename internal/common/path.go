package common

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const DefaultBinaryPath = "agent-notify"

func ResolveBinaryPath(input string) string {
	input = strings.TrimSpace(input)
	if input != "" {
		return toUnixStylePath(input)
	}

	executablePath, err := os.Executable()
	if err == nil {
		if resolved, resolveErr := filepath.EvalSymlinks(executablePath); resolveErr == nil {
			return toUnixStylePath(resolved)
		}
		return toUnixStylePath(executablePath)
	}

	return DefaultBinaryPath
}

// toUnixStylePath converts Windows path to Unix-style path (e.g., C:\Users\... -> /c/Users/...)
// This format is commonly used in Git Bash and similar environments on Windows.
func toUnixStylePath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	// Check if it's a Windows absolute path (e.g., C:\...)
	if len(path) >= 2 && path[1] == ':' {
		driveLetter := strings.ToLower(string(path[0]))
		rest := path[2:]
		// Convert backslashes to forward slashes
		rest = strings.ReplaceAll(rest, "\\", "/")
		return "/" + driveLetter + rest
	}

	// Already a Unix-style path or relative path
	return strings.ReplaceAll(path, "\\", "/")
}
