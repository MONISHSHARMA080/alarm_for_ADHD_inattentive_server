package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type ActivityInfo struct {
	WindowTitle string
	ProcessName string
	Timestamp   time.Time
}

// checkDependencies verifies all required tools are installed
func checkDependencies() error {
	// Check for xdotool
	if _, err := exec.LookPath("xdotool"); err != nil {
		return fmt.Errorf("xdotool is not installed. Please run: sudo apt-get install xdotool")
	}

	// Check for X11 display
	if _, err := exec.Command("xdotool", "getdisplaygeometry").Output(); err != nil {
		return fmt.Errorf("cannot connect to X display. Make sure you're running in X11 session: %v", err)
	}

	return nil
}

func getActiveWindowTitle() (string, error) {
	// First get the active window ID
	windowID, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get active window ID: %v", err)
	}

	// Then get the window name using the ID
	cmd := exec.Command("xdotool", "getwindowname", strings.TrimSpace(string(windowID)))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get window name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getActiveProcessName() (string, error) {
	// Get window ID
	windowID, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get active window ID: %v", err)
	}

	// Get PID using window ID
	cmd := exec.Command("xdotool", "getwindowpid", strings.TrimSpace(string(windowID)))
	pidBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get window PID: %v", err)
	}

	// Get process name using PID
	pid := strings.TrimSpace(string(pidBytes))
	cmd = exec.Command("ps", "-p", pid, "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get process name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func monitorActivity(interval time.Duration) {
	consecutiveErrors := 0
	maxConsecutiveErrors := 3

	for {
		windowTitle, err := getActiveWindowTitle()
		if err != nil {
			consecutiveErrors++
			log.Printf("Error getting window title (attempt %d/%d): %v",
				consecutiveErrors, maxConsecutiveErrors, err)

			if consecutiveErrors >= maxConsecutiveErrors {
				log.Println("Too many consecutive errors. Possible causes:")
				log.Println("1. No X11 session active")
				log.Println("2. Running in Wayland instead of X11")
				log.Println("3. No active window")
				log.Println("4. Insufficient permissions")
				log.Println("\nTrying again in 10 seconds...")
				time.Sleep(10 * time.Second)
				consecutiveErrors = 0
				continue
			}

			time.Sleep(interval)
			continue
		}
		consecutiveErrors = 0 // Reset error counter on success

		processName, err := getActiveProcessName()
		if err != nil {
			log.Printf("Error getting process name: %v", err)
			time.Sleep(interval)
			continue
		}

		activity := ActivityInfo{
			WindowTitle: windowTitle,
			ProcessName: processName,
			Timestamp:   time.Now(),
		}

		// Log the activity
		fmt.Printf("[%s] Process: %s | Window: %s\n",
			activity.Timestamp.Format("15:04:05"),
			activity.ProcessName,
			activity.WindowTitle)

		// Application specific logging
		switch {
		case strings.Contains(strings.ToLower(processName), "chrome"):
			fmt.Println("→ Browser activity detected")
		case strings.Contains(strings.ToLower(processName), "code"):
			fmt.Println("→ VS Code activity detected")
		case strings.Contains(strings.ToLower(processName), "nvim"):
			fmt.Println("→ Neovim activity detected")
		}

		time.Sleep(interval)
	}
}

func main() {
	log.Println("Starting activity monitor...")

	// Check dependencies first
	if err := checkDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	fmt.Println("All dependencies checked. Starting monitoring...")
	fmt.Println("Press Ctrl+C to stop")

	// Monitor every 5 seconds
	monitorActivity(5 * time.Second)
}
