package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	fmt.Println("Generating platform-specific icons...")
	
	// Create icons directory if it doesn't exist
	iconsDir := "assets/icons"
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		log.Fatalf("Failed to create icons directory: %v", err)
	}
	
	// Source SVG file
	sourceSVG := "assets/icon.svg"
	if _, err := os.Stat(sourceSVG); os.IsNotExist(err) {
		log.Fatalf("Source SVG file not found: %s", sourceSVG)
	}
	
	// Generate different sizes and formats
	generateIcons(sourceSVG, iconsDir)
	
	fmt.Println("Icon generation completed!")
}

func generateIcons(sourceSVG, outputDir string) {
	// Define the icon sizes and formats needed
	iconConfigs := []struct {
		size   int
		format string
		name   string
	}{
		// Windows ICO sizes
		{16, "png", "icon-16.png"},
		{32, "png", "icon-32.png"},
		{48, "png", "icon-48.png"},
		{64, "png", "icon-64.png"},
		{128, "png", "icon-128.png"},
		{256, "png", "icon-256.png"},
		
		// macOS ICNS sizes
		{512, "png", "icon-512.png"},
		{1024, "png", "icon-1024.png"},
		
		// Linux standard sizes
		{22, "png", "icon-22.png"},
		{24, "png", "icon-24.png"},
		{36, "png", "icon-36.png"},
		{72, "png", "icon-72.png"},
		{96, "png", "icon-96.png"},
		{144, "png", "icon-144.png"},
		{192, "png", "icon-192.png"},
	}
	
	// Try to use different tools based on availability
	for _, config := range iconConfigs {
		outputPath := filepath.Join(outputDir, config.name)
		
		// Try different conversion methods
		if err := convertWithInkscape(sourceSVG, outputPath, config.size); err != nil {
			if err := convertWithImageMagick(sourceSVG, outputPath, config.size); err != nil {
				if err := convertWithRSVG(sourceSVG, outputPath, config.size); err != nil {
					fmt.Printf("Warning: Could not generate %s (size %d): %v\n", config.name, config.size, err)
					continue
				}
			}
		}
		
		fmt.Printf("Generated: %s (%dx%d)\n", config.name, config.size, config.size)
	}
	
	// Generate Windows ICO file if possible
	generateWindowsICO(outputDir)
	
	// Generate macOS ICNS file if possible
	generateMacOSICNS(outputDir)
}

func convertWithInkscape(input, output string, size int) error {
	cmd := exec.Command("inkscape", 
		"--export-type=png",
		fmt.Sprintf("--export-width=%d", size),
		fmt.Sprintf("--export-height=%d", size),
		fmt.Sprintf("--export-filename=%s", output),
		input)
	return cmd.Run()
}

func convertWithImageMagick(input, output string, size int) error {
	cmd := exec.Command("magick", "convert",
		"-background", "transparent",
		"-size", fmt.Sprintf("%dx%d", size, size),
		input, output)
	return cmd.Run()
}

func convertWithRSVG(input, output string, size int) error {
	cmd := exec.Command("rsvg-convert",
		"-w", fmt.Sprintf("%d", size),
		"-h", fmt.Sprintf("%d", size),
		"-o", output,
		input)
	return cmd.Run()
}

func generateWindowsICO(iconsDir string) {
	if runtime.GOOS != "windows" {
		return
	}
	
	// Try to create ICO file using ImageMagick
	icoPath := filepath.Join(iconsDir, "app.ico")
	pngFiles := []string{
		filepath.Join(iconsDir, "icon-16.png"),
		filepath.Join(iconsDir, "icon-32.png"),
		filepath.Join(iconsDir, "icon-48.png"),
		filepath.Join(iconsDir, "icon-64.png"),
		filepath.Join(iconsDir, "icon-128.png"),
		filepath.Join(iconsDir, "icon-256.png"),
	}
	
	// Check if all PNG files exist
	var existingFiles []string
	for _, file := range pngFiles {
		if _, err := os.Stat(file); err == nil {
			existingFiles = append(existingFiles, file)
		}
	}
	
	if len(existingFiles) > 0 {
		args := append([]string{"convert"}, existingFiles...)
		args = append(args, icoPath)
		
		cmd := exec.Command("magick", args...)
		if err := cmd.Run(); err == nil {
			fmt.Printf("Generated: app.ico\n")
		}
	}
}

func generateMacOSICNS(iconsDir string) {
	if runtime.GOOS != "darwin" {
		return
	}
	
	// Try to create ICNS file using iconutil
	icnsPath := filepath.Join(iconsDir, "app.icns")
	iconsetDir := filepath.Join(iconsDir, "app.iconset")
	
	// Create iconset directory
	if err := os.MkdirAll(iconsetDir, 0755); err != nil {
		return
	}
	
	// Copy PNG files to iconset with proper naming
	iconsetFiles := map[string]string{
		"icon-16.png":   "icon_16x16.png",
		"icon-32.png":   "icon_16x16@2x.png",
		"icon-32.png":   "icon_32x32.png",
		"icon-64.png":   "icon_32x32@2x.png",
		"icon-128.png":  "icon_128x128.png",
		"icon-256.png":  "icon_128x128@2x.png",
		"icon-256.png":  "icon_256x256.png",
		"icon-512.png":  "icon_256x256@2x.png",
		"icon-512.png":  "icon_512x512.png",
		"icon-1024.png": "icon_512x512@2x.png",
	}
	
	for src, dst := range iconsetFiles {
		srcPath := filepath.Join(iconsDir, src)
		dstPath := filepath.Join(iconsetDir, dst)
		
		if _, err := os.Stat(srcPath); err == nil {
			if data, err := os.ReadFile(srcPath); err == nil {
				os.WriteFile(dstPath, data, 0644)
			}
		}
	}
	
	// Generate ICNS file
	cmd := exec.Command("iconutil", "-c", "icns", iconsetDir, "-o", icnsPath)
	if err := cmd.Run(); err == nil {
		fmt.Printf("Generated: app.icns\n")
		// Clean up iconset directory
		os.RemoveAll(iconsetDir)
	}
}
