//go:build windows

package q

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func isDebuggerEnv() bool {
	for _, e := range os.Environ() {
		if len(e) > 5 && e[:5] == "DELVE" {
			return true
		}
	}
	return false
}

func CheckEnv() {
	if isDebuggerEnv() {
		fmt.Println("Debugger detected!")
		os.Exit(1)
	} else {
		fmt.Println("Running normally.")
	}
}

func modifyBinary() {
	file, _ := os.OpenFile(os.Args[0], os.O_WRONLY, 0755)
	defer file.Close()
	file.WriteAt([]byte{0x90, 0x90, 0x90}, 0x1234) // Write NOPs into binary
}

func ModifyMe() {
	modifyBinary()
	fmt.Println("Code modified at runtime!")
}

func detectBreakpoints() bool {
	start := time.Now()
	for i := 0; i < 1000000; i++ {
	} // Loop to detect pause
	elapsed := time.Since(start)
	return elapsed > 100*time.Millisecond // If execution is delayed, assume breakpoint
}

func TimingCheck() {
	if detectBreakpoints() {
		fmt.Println("Debugger detected!")
		os.Exit(1)
	} else {
		// fmt.Println("Running normally.")
	}
}

// getParentProcessNameWindows gets the parent process name using WMIC for Windows
func getParentProcessNameWindows() string {
	ppid := os.Getppid()
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessID=%d", ppid), "get", "Name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1]) // Skip header "Name"
	}
	return ""
}

// isDebuggerParent detects if the parent process is a known debugger
func isDebuggerParent() bool {
	parent := getParentProcessNameWindows()
	fmt.Printf("Parent process name: %s\n", parent)

	// If parent name is blank, it could be an error or process-hiding attempt
	if parent == "" {
		fmt.Println("Blank parent process name — suspicious activity detected!")
		return true
	}

	// List of known debugger process names
	debuggerNames := []string{"dlv", "x64dbg", "lldb", "gdb", "windbg", "ollydbg", "immunitydebugger"}

	// List of known "safe" parent processes to ignore
	safeNames := []string{"cmd.exe", "powershell.exe", "pwsh.exe", "explorer.exe", "conhost.exe"}

	// Check for known debugger names
	for _, dbg := range debuggerNames {
		if strings.Contains(strings.ToLower(parent), dbg) {
			fmt.Printf("Debugger process detected: %s\n", parent)
			return true
		}
	}

	// Check if the parent is in the "safe" list
	for _, safe := range safeNames {
		if strings.Contains(strings.ToLower(parent), safe) {
			fmt.Printf("Safe parent process detected: %s\n", parent)
			return false
		}
	}

	// If parent is not in safe names and it's not blank, treat it as suspicious
	fmt.Println("Parent process not in safe list — suspicious activity detected!")
	return true
}

// KillTheChild exits if a debugger parent process is detected
func KillTheChild() {
	fmt.Println("Detecting process")
	if isDebuggerParent() {
		fmt.Println("Debugger detected! Exiting.")
		os.Exit(1)
	} else {
		fmt.Println("Running normally.")
	}
}
