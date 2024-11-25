package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func promptString(reader *bufio.Reader, prompt, defaultVal string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultVal)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func promptInt(reader *bufio.Reader, prompt string, min, max int) int {
	for {
		fmt.Printf("%s (%d-%d): ", prompt, min, max)
		input, _ := reader.ReadString('\n')
		num, err := strconv.Atoi(strings.TrimSpace(input))
		if err == nil && num >= min && num <= max {
			return num
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

func promptYesNo(reader *bufio.Reader, prompt string) bool {
	for {
		fmt.Printf("%s (y/n): ", prompt)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func displayConfig(config *EmulatorConfig) {
	fmt.Println("========== CONFIG ==========")
	fmt.Println("Name:", config.Name)
	fmt.Println("Device Type:", config.DeviceType)
	fmt.Println("API Level:", config.APILevel)
	fmt.Println("System Image:", config.SystemImage)
	fmt.Printf("RAM: %d MB\n", config.RAM)

	if config.DeviceType == "custom" {
		fmt.Println("Resolution:", config.Resolution)
		fmt.Printf("Density: %d dpi\n", config.Density)
	}

	fmt.Println("Has Keyboard:", config.HasKeyboard)
	fmt.Println("============================")
}

func isPackageInstalled(sdkManager string, packageName string) bool {
	cmd := exec.Command(sdkManager, "--list_installed")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), packageName)
}

func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s %s\n", name, strings.Join(args, " "))

	return cmd.Run()
}

func createEmulator(androidHome string, config *EmulatorConfig) error {
	displayConfig(config)
	fmt.Println("\nCreating emulator with selected configuration...")

	sdkmanager := filepath.Join(androidHome, "cmdline-tools", "latest", "bin", "sdkmanager")

	if !isPackageInstalled(sdkmanager, config.SystemImage) {
		cmd := exec.Command(sdkmanager, "--install", config.SystemImage)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install system image: %w", err)
		}
	} else {
		fmt.Printf("System image %s is already installed\n", config.SystemImage)
	}

	avdmanager := filepath.Join(androidHome, "cmdline-tools", "latest", "bin", "avdmanager")
	createArgs := []string{
		"echo", "'no'", "|",
		"create", "avd",
		"--name", config.Name,
		"--package", config.SystemImage,
		"--force",
	}

	if config.DeviceType != "custom" {
		createArgs = append(createArgs, "--device", config.DeviceType)
	}

	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("echo no | %s create avd --force --name %s --package '%s'",
			avdmanager,
			config.Name,
			config.SystemImage,
		),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create AVD: %v", err)
	}

	configPath := filepath.Join(os.Getenv("HOME"), ".android", "avd", config.Name+".avd", "config.ini")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer f.Close()

	configs := []string{
		fmt.Sprintf("hw.ramSize=%d", config.RAM),
		"hw.cpu.ncore=4",
		fmt.Sprintf("hw.keyboard=%v", config.HasKeyboard),
		"hw.gpu.enabled=yes",
		"hw.gpu.mode=auto",
	}

	if config.DeviceType == "custom" {
		width, height, _ := strings.Cut(config.Resolution, "x")
		configs = append(configs,
			fmt.Sprintf("hw.lcd.width=%s", width),
			fmt.Sprintf("hw.lcd.height=%s", height),
			fmt.Sprintf("hw.lcd.density=%d", config.Density),
		)
	}

	configs = append(configs, config.ExtraConfigs...)

	for _, cfg := range configs {
		if _, err := f.WriteString(cfg + "\n"); err != nil {
			return fmt.Errorf("failed to write config: %v", err)
		}
	}

	return nil
}

func getSystemArchitecture() string {
	if runtime.GOARCH == "arm64" {
		return "arm64-v8a"
	}
	return "x86_64"
}
