package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var targets = []struct {
	OS   string
	Arch string
	Name string
	Arm  string // For GOARM
}{
	{"windows", "amd64", "joss.exe", ""},
	{"linux", "amd64", "joss-linux-amd64", ""},
	{"linux", "arm64", "joss-linux-arm64", ""},
	{"linux", "arm", "joss-linux-armv7", "7"},
	{"darwin", "amd64", "joss-macos-amd64", ""},
	{"darwin", "arm64", "joss-macos-arm64", ""},
}

func main() {
	installerDir := "installer"

	// 1. Clean/Create installer directory
	fmt.Println("[Dist] Cleaning installer directory...")
	os.RemoveAll(installerDir)
	if err := os.MkdirAll(installerDir, 0755); err != nil {
		fatal(err)
	}

	// 2. Build Binaries
	fmt.Println("[Dist] Building binaries...")
	for _, t := range targets {
		fmt.Printf("  -> Building for %s/%s...\n", t.OS, t.Arch)
		cmd := exec.Command("go", "build", "-o", filepath.Join(installerDir, t.Name), "./cmd/joss")
		cmd.Env = append(os.Environ(), "GOOS="+t.OS, "GOARCH="+t.Arch)
		if t.Arm != "" {
			cmd.Env = append(cmd.Env, "GOARM="+t.Arm)
		}
		// Enable CGO? Usually disabled for cross-compilation unless we have cross-compilers.
		// JosSecurity CLI (cmd/joss) shouldn't need CGO for basic functionality (unlike the runner).
		// So we disable CGO for portability.
		cmd.Env = append(cmd.Env, "CGO_ENABLED=0")

		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("Error building %s: %v\n%s\n", t.Name, err, string(out))
			os.Exit(1)
		}
	}

	// 3. Copy VSIX
	fmt.Println("[Dist] Copying VSIX extension...")
	vsixPath := findVSIX()
	if vsixPath == "" {
		fmt.Println("[Dist] WARNING: No .vsix file found in vscode-joss/ or installer/ (source). Skipping.")
	} else {
		dest := filepath.Join(installerDir, filepath.Base(vsixPath))
		if err := copyFile(vsixPath, dest); err != nil {
			fatal(err)
		}
		fmt.Printf("  -> Copied %s\n", filepath.Base(vsixPath))
	}

	// 4. Create ZIP
	fmt.Println("[Dist] Creating jossecurity-binaries.zip...")
	if err := zipDirectory(installerDir, filepath.Join(installerDir, "jossecurity-binaries.zip")); err != nil {
		fatal(err)
	}

	fmt.Println("[Dist] Done! Artifacts in 'installer/'")
}

func findVSIX() string {
	// Check vscode-joss folder first
	matches, _ := filepath.Glob("vscode-joss/*.vsix")
	if len(matches) > 0 {
		return matches[0] // Return first match
	}
	// Check current dir or other locations if needed
	return ""
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func zipDirectory(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the zip file itself if it's inside the source directory
		if path == target {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Make paths relative to source
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func fatal(err error) {
	fmt.Printf("[Dist] Error: %v\n", err)
	os.Exit(1)
}
