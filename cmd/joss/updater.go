package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jossecurity/joss/pkg/version"
)

const (
	githubReleasesURL = "https://api.github.com/repos/josprox/Joss-language/releases"
	defaultChannel    = "stable"
)

type UpdateConfig struct {
	Channel          string `json:"channel"`
	LastCheck        int64  `json:"last_check"`
	LatestVersion    string `json:"latest_version"`
	LatestAssetURL   string `json:"latest_asset_url"`
	NotificationDone bool   `json:"notification_done"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []GitHubAsset `json:"assets"`
}

func getUpdateConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".joss")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "update_config.json")
}

func loadUpdateConfig() UpdateConfig {
	cfg := UpdateConfig{Channel: defaultChannel}
	path := getUpdateConfigPath()
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	if cfg.Channel == "" {
		cfg.Channel = defaultChannel
	}
	return cfg
}

func saveUpdateConfig(cfg UpdateConfig) {
	path := getUpdateConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		os.WriteFile(path, data, 0644)
	}
}

// checkUpdateBackground performs a fast, non-blocking check for new versions on startup.
func checkUpdateBackground() {
	cfg := loadUpdateConfig()

	// Check at most once per 30 minutes
	now := time.Now().Unix()
	if now-cfg.LastCheck < 1800 && cfg.LatestVersion != "" {
		if isVersionNewer(cfg.LatestVersion, version.Version) {
			printUpdateNotification(cfg.LatestVersion, cfg.Channel)
		}
		return
	}

	// Non-blocking quick check
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Silent recover
			}
		}()

		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequest("GET", githubReleasesURL, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "Joss-CLI-Updater/"+version.Version)

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			return
		}
		defer resp.Body.Close()

		var releases []GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil || len(releases) == 0 {
			return
		}

		targetRelease := selectReleaseByChannel(releases, cfg.Channel)
		if targetRelease == nil {
			return
		}

		remoteVer := cleanVersionTag(targetRelease.TagName)
		assetURL := findMatchingAsset(targetRelease.Assets, runtime.GOOS, runtime.GOARCH)

		cfg.LastCheck = time.Now().Unix()
		cfg.LatestVersion = remoteVer
		cfg.LatestAssetURL = assetURL
		saveUpdateConfig(cfg)

		if isVersionNewer(remoteVer, version.Version) {
			printUpdateNotification(remoteVer, cfg.Channel)
		}
	}()
}

func printUpdateNotification(remoteVer, channel string) {
	channelTag := strings.ToUpper(channel)
	fmt.Printf("\n💡 \033[1;33m[JOSS UPDATE]\033[0m ¡Nueva versión disponible! \033[1;36mv%s\033[0m -> \033[1;32mv%s\033[0m (%s)\n", version.Version, remoteVer, channelTag)
	fmt.Printf("   Ejecuta \033[1;35mjoss update --%s\033[0m para actualizar automáticamente el runtime, SDK y extensión.\n\n", strings.ToLower(channel))
}

// handleUpdateCommand executes the 'joss update' CLI command.
func handleUpdateCommand(args []string) {
	cfg := loadUpdateConfig()
	force := false
	// Parse flags for channel override (--canary / --stable) and force flag (-f / --force)
	for _, arg := range args {
		argLower := strings.ToLower(arg)
		if argLower == "-f" || argLower == "--force" || argLower == "-force" {
			force = true
		} else if strings.Contains(argLower, "canary") {
			cfg.Channel = "canary"
		} else if strings.Contains(argLower, "stable") {
			cfg.Channel = "stable"
		}
	}

	saveUpdateConfig(cfg)

	fmt.Printf("\n=======================================================\n")
	fmt.Printf("🔄 ACTUALIZADOR DE JOSS (Joss Auto-Updater)\n")
	fmt.Printf(" Versión Actual : v%s\n", version.Version)
	fmt.Printf(" Canal Elegido  : %s\n", strings.ToUpper(cfg.Channel))
	if force {
		fmt.Printf(" Modo Forzado   : ACTIVO (-f)\n")
	}
	fmt.Printf("=======================================================\n\n")

	fmt.Println("🌐 Consultando la API de GitHub Releases...")

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", githubReleasesURL, nil)
	if err != nil {
		fmt.Printf("Error preparando solicitud: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", "Joss-CLI-Updater/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error de conexión a GitHub: %v (Comprueba tu conexión a Internet)\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Error obteniendo releases de GitHub (Status %d)\n", resp.StatusCode)
		return
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil || len(releases) == 0 {
		fmt.Println("No se encontraron releases públicas de Joss en GitHub.")
		return
	}

	targetRelease := selectReleaseByChannel(releases, cfg.Channel)
	if targetRelease == nil {
		fmt.Printf("No se encontró ningún release disponible para el canal '%s'.\n", cfg.Channel)
		return
	}

	remoteVer := cleanVersionTag(targetRelease.TagName)
	fmt.Printf("📦 Release seleccionado : %s (%s)\n", targetRelease.Name, targetRelease.TagName)

	if !force && compareVersions(remoteVer, version.Version) == 0 {
		fmt.Printf("✨ Joss ya se encuentra actualizado en la versión v%s para el canal %s.\n", version.Version, strings.ToUpper(cfg.Channel))
		fmt.Println("💡 Tip: Usa 'joss update -f' para forzar la re-descarga e instalación del binario.")
		fmt.Println()
		return
	}

	if force {
		fmt.Println("⚠️  Re-descargando y re-instalando la versión del release por modo forzado (-f)...")
	}

	assetURL := findMatchingAsset(targetRelease.Assets, runtime.GOOS, runtime.GOARCH)
	if assetURL == "" {
		fmt.Printf("⚠️ No se encontró paquete compilado específico para %s/%s en la release %s.\n", runtime.GOOS, runtime.GOARCH, targetRelease.TagName)
		fmt.Println("Descargando actualización del repositorio general...")
		if len(targetRelease.Assets) > 0 {
			assetURL = targetRelease.Assets[0].BrowserDownloadURL
		}
	}

	if assetURL == "" {
		fmt.Println("Error: No hay binarios disponibles para descargar en este release.")
		return
	}

	fmt.Printf("⬇️  Descargando actualización desde: %s\n", assetURL)

	tempDir, err := os.MkdirTemp("", "joss-update-*")
	if err != nil {
		fmt.Printf("Error creando directorio temporal: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	downloadPath := filepath.Join(tempDir, "update_download.bin")
	if err := downloadFile(downloadPath, assetURL); err != nil {
		fmt.Printf("Error descargando la actualización: %v\n", err)
		return
	}

	binaryToApply := downloadPath

	// Extract binary if download is a ZIP archive
	if data, err := os.ReadFile(downloadPath); err == nil && len(data) >= 4 && data[0] == 'P' && data[1] == 'K' {
		zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err == nil {
			extractedBin := filepath.Join(tempDir, "extracted_joss_binary")
			found := false
			expectedName := fmt.Sprintf("joss-%s-%s", runtime.GOOS, runtime.GOARCH)
			if runtime.GOOS == "windows" {
				expectedName += ".exe"
			}
			for _, file := range zipReader.File {
				name := strings.ToLower(filepath.Base(file.Name))
				if name == "joss.exe" || name == "joss" || name == expectedName {
					rc, err := file.Open()
					if err == nil {
						content, _ := io.ReadAll(rc)
						rc.Close()
						if len(content) > 0 {
							os.WriteFile(extractedBin, content, 0755)
							binaryToApply = extractedBin
							found = true
							break
						}
					}
				}
			}
			if !found {
				for _, file := range zipReader.File {
					if !file.FileInfo().IsDir() && strings.HasPrefix(strings.ToLower(filepath.Base(file.Name)), "joss") {
						rc, err := file.Open()
						if err == nil {
							content, _ := io.ReadAll(rc)
							rc.Close()
							if len(content) > 0 {
								os.WriteFile(extractedBin, content, 0755)
								binaryToApply = extractedBin
								break
							}
						}
					}
				}
			}
		}
	}

	fmt.Println("⚙️  Aplicando actualización de runtime, SDK y extensión...")

	currentExe, err := os.Executable()
	if err != nil {
		fmt.Printf("Error obteniendo ruta del ejecutable actual: %v\n", err)
		return
	}

	// Apply self-update binary replacement safely
	if err := replaceExecutable(currentExe, binaryToApply); err != nil {
		fmt.Printf("Error aplicando actualización del binario: %v\n", err)
		return
	}

	cfg.LastCheck = time.Now().Unix()
	cfg.LatestVersion = remoteVer
	cfg.LatestAssetURL = assetURL
	saveUpdateConfig(cfg)

	fmt.Printf("\n✨ ¡JOSS SE HA ACTUALIZADO EXITOSAMENTE!\n")
	fmt.Printf(" Nueva Versión : v%s (%s)\n", remoteVer, strings.ToUpper(cfg.Channel))
	fmt.Printf(" Ejecutable     : %s\n\n", currentExe)
}

func selectReleaseByChannel(releases []GitHubRelease, channel string) *GitHubRelease {
	if strings.ToLower(channel) == "canary" {
		// Prefer prerelease if available, otherwise latest release
		for i := range releases {
			if releases[i].Prerelease && !releases[i].Draft {
				return &releases[i]
			}
		}
		// Fallback to latest
		if len(releases) > 0 {
			return &releases[0]
		}
		return nil
	}

	// Stable channel: non-draft and non-prerelease
	for i := range releases {
		if !releases[i].Draft && !releases[i].Prerelease {
			return &releases[i]
		}
	}
	// Fallback if none found
	if len(releases) > 0 {
		return &releases[0]
	}
	return nil
}

func findMatchingAsset(assets []GitHubAsset, targetOS, targetArch string) string {
	targetOS = strings.ToLower(targetOS)
	targetArch = strings.ToLower(targetArch)

	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, targetOS) && strings.Contains(name, targetArch) {
			return a.BrowserDownloadURL
		}
	}
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, targetOS) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor devolvió status %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func replaceExecutable(currentExe, newExe string) error {
	if runtime.GOOS == "windows" {
		oldExe := currentExe + ".old"
		os.Remove(oldExe)

		// Rename running executable to .old
		if err := os.Rename(currentExe, oldExe); err != nil {
			return fmt.Errorf("no se pudo renombrar ejecutable actual: %w", err)
		}

		// Copy new executable to currentExe location
		input, err := os.ReadFile(newExe)
		if err != nil {
			return err
		}
		if err := os.WriteFile(currentExe, input, 0755); err != nil {
			// Rollback
			os.Rename(oldExe, currentExe)
			return fmt.Errorf("no se pudo escribir nuevo ejecutable: %w", err)
		}
		return nil
	}

	// Unix-like OS (Linux / macOS)
	input, err := os.ReadFile(newExe)
	if err != nil {
		return err
	}
	return os.WriteFile(currentExe, input, 0755)
}

func cleanVersionTag(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

func parseVersionParts(v string) []int {
	v = cleanVersionTag(v)
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		fmt.Sscanf(p, "%d", &n)
		nums = append(nums, n)
	}
	return nums
}

func compareVersions(v1, v2 string) int {
	nums1 := parseVersionParts(v1)
	nums2 := parseVersionParts(v2)

	maxLen := len(nums1)
	if len(nums2) > maxLen {
		maxLen = len(nums2)
	}

	for i := 0; i < maxLen; i++ {
		n1 := 0
		if i < len(nums1) {
			n1 = nums1[i]
		}
		n2 := 0
		if i < len(nums2) {
			n2 = nums2[i]
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

func isVersionNewer(remote, current string) bool {
	return compareVersions(remote, current) > 0
}
