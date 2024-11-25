package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	config := &EmulatorConfig{}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Android Emulator Creation Wizard ===")

	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		fmt.Println("ANDROID_HOME environment variable not set")
		os.Exit(1)
	}

	config.Name = promptString(reader, "Enter emulator name", "Tablet_Emulator")

	fmt.Println("\nSelect device type:")
	devices := []string{
		"pixel_c (Pixel C Tablet)",
		"pixel_tablet (Pixel Tablet)",
		"nexus_9 (Nexus 9)",
		"custom",
	}

	for i, device := range devices {
		fmt.Printf("%d. %s\n", i+1, device)
	}
	deviceIndex := promptInt(reader, "Enter number", 1, len(devices)) - 1

	if deviceIndex < len(devices)-1 {
		config.DeviceType = strings.Split(devices[deviceIndex], " ")[0]
	} else {
		config.DeviceType = "custom"
	}

	fmt.Println("\nSelect API Level:")
	apiLevels := []string{
		"34 (Android 14)",
		"33 (Android 13)",
		"32 (Android 12L)",
		"31 (Android 12)",
		"30 (Android 11)",
	}

	for i, api := range apiLevels {
		fmt.Printf("%d. %s\n", i+1, api)
	}
	apiIndex := promptInt(reader, "Enter number", 1, len(apiLevels)) - 1
	config.APILevel = strings.Split(apiLevels[apiIndex], " ")[0]

	fmt.Println("\nSelect system image:")
	images := []string{
		"Google APIs",
		"Google Play",
		"AOSP",
	}
	for i, img := range images {
		fmt.Printf("%d. %s\n", i+1, img)
	}

	imageIndex := promptInt(reader, "Enter number", 1, len(images)) - 1

	arch := getSystemArchitecture()

	switch imageIndex {
	case 0:
		config.SystemImage = fmt.Sprintf("system-images;android-%s;google_apis;%s", config.APILevel, arch)
	case 1:
		if arch == "arm64-v8a" {
			fmt.Println("\nNote: Google Play system images might not be available for ARM64. Falling back to Google APIs...")
			config.SystemImage = fmt.Sprintf("system-images;android-%s;google_apis;%s", config.APILevel, arch)
		} else {
			config.SystemImage = fmt.Sprintf("system-images;android-%s;google_apis_playstore;%s", config.APILevel, arch)
		}
	case 2:
		config.SystemImage = fmt.Sprintf("system-images;android-%s;default;%s", config.APILevel, arch)
	}

	fmt.Printf("\nSelected system image: %s\n", config.SystemImage)

	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		fmt.Println("\nConfiguring hardware acceleration for Apple Silicon...")
		config.ExtraConfigs = []string{
			"hw.cpu.ncore=4",
			"hw.ramSize=4096",
			"hw.lcd.density=320",
			"hw.gpu.enabled=yes",
			"hw.gpu.mode=auto",
			"hw.keyboard=yes",
			// Apple Silicon specific settings
			"hw.cpu.arch=arm64",
			"hw.cpu.model=cortex-a57",
			"hw.hypervisorDriver=hvf",
			"hw.useextension=on",
		}
	}

	config.RAM = promptInt(reader, "\nEnter RAM size in MB (recommended: 4096)", 1024, 8192)

	if config.DeviceType == "custom" {
		fmt.Println("\nSelect screen resolution:")
		resolutions := []string{
			"1800x2560 (Pixel C)",
			"2560x1600 (Nexus 10)",
			"2048x1536 (iPad-like)",
			"1280x800 (Typical tablet)",
		}
		for i, res := range resolutions {
			fmt.Printf("%d. %s\n", i+1, res)
		}
		resIndex := promptInt(reader, "Enter number", 1, len(resolutions)) - 1
		config.Resolution = strings.Split(resolutions[resIndex], " ")[0]
	}

	if config.DeviceType == "custom" {
		fmt.Println("\nSelect screen density:")
		densities := []string{
			"320 (xhdpi)",
			"240 (hdpi)",
			"160 (mdpi)",
		}
		for i, den := range densities {
			fmt.Printf("%d. %s\n", i+1, den)
		}
		denIndex := promptInt(reader, "Enter number", 1, len(densities)) - 1
		config.Density, _ = strconv.Atoi(strings.Split(densities[denIndex], " ")[0])
	}

	config.HasKeyboard = promptYesNo(reader, "\nEmulate hardware keyboard?")

	fmt.Println("\nCreating emulator with selected configuration...")
	if err := createEmulator(androidHome, config); err != nil {
		fmt.Printf("Error creating emulator: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nEmulator created successfully!")
	fmt.Printf("To start the emulator, run: %s/emulator/emulator -avd %s\n", androidHome, config.Name)
}
