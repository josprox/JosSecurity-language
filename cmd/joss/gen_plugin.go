package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createNewPluginProject(targetPath string) {
	pluginName := strings.ToLower(filepath.Base(targetPath))
	pluginName = strings.ReplaceAll(pluginName, "-", "_")
	if pluginName == "" || pluginName == "." {
		fmt.Println("Error: Especifica una ruta o nombre de plugin válido (ej: joss new plugin joss_bg_remover)")
		return
	}

	fmt.Printf("📦 Creando nuevo plugin oficial '%s' en '%s'...\n", pluginName, targetPath)

	dirs := []string{
		targetPath,
		filepath.Join(targetPath, "src"),
		filepath.Join(targetPath, "cmd", "sidecar"),
		filepath.Join(targetPath, ".github", "workflows"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error al crear directorio %s: %v\n", dir, err)
			return
		}
	}

	className := packageClassName(pluginName)

	manifestContent := fmt.Sprintf(`name: %s
version: 1.0.0
type: joss
description: Plugin oficial con sidecar nativo Go para Joss Language
author: Developer <dev@example.com>
license: MIT
entry:
  main: src/plugin.joss
native:
  protocol: joss-rpc-v1
  # URLs apuntan a los assets del GitHub Release.
  # El runtime descarga el sidecar correcto para la plataforma actual la primera vez.
  # Publica los binarios en GitHub Releases con: joss pub release
  windows-amd64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-windows-amd64.exe
  windows-arm64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-windows-arm64.exe
  linux-amd64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-linux-amd64
  linux-arm64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-linux-arm64
  darwin-amd64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-darwin-amd64
  darwin-arm64: https://github.com/TU_USUARIO/%s/releases/latest/download/%s-darwin-arm64
dependencies:
`, pluginName,
		pluginName, pluginName,
		pluginName, pluginName,
		pluginName, pluginName,
		pluginName, pluginName,
		pluginName, pluginName,
		pluginName, pluginName,
	)

	pluginJossContent := fmt.Sprintf(`// src/plugin.joss
// Clase exportada para interactuar con el plugin %s

class %s {
    function ping() {
        return Plugin::call("%s", "ping", [])
    }

    function process($data) {
        return Plugin::call("%s", "process", [$data])
    }
}
`, pluginName, className, pluginName, pluginName)

	sidecarGoContent := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type request struct {
	Protocol string        ` + "`" + `json:"protocol"` + "`" + `
	ID       string        ` + "`" + `json:"id"` + "`" + `
	Method   string        ` + "`" + `json:"method"` + "`" + `
	Args     []interface{} ` + "`" + `json:"args"` + "`" + `
}

type response struct {
	ID     string      ` + "`" + `json:"id"` + "`" + `
	Result interface{} ` + "`" + `json:"result,omitempty"` + "`" + `
	Error  interface{} ` + "`" + `json:"error,omitempty"` + "`" + `
}

func main() {
	var req request
	if err := json.NewDecoder(io.LimitReader(os.Stdin, 16<<20)).Decode(&req); err != nil {
		write(response{Error: map[string]string{"code": "BAD_REQUEST", "message": err.Error()}})
		return
	}
	if req.Protocol != "joss-rpc-v1" {
		write(response{ID: req.ID, Error: map[string]string{"code": "BAD_PROTOCOL", "message": "se requiere joss-rpc-v1"}})
		return
	}

	result, err := dispatch(req.Method, req.Args)
	if err != nil {
		write(response{ID: req.ID, Error: map[string]string{"code": "PLUGIN_ERROR", "message": err.Error()}})
		return
	}
	write(response{ID: req.ID, Result: result})
}

func dispatch(method string, args []interface{}) (interface{}, error) {
	switch method {
	case "ping":
		return map[string]interface{}{"status": "pong", "plugin": "%s"}, nil
	case "process":
		return map[string]interface{}{"ok": true, "args": args}, nil
	default:
		return nil, fmt.Errorf("método no soportado: %%s", method)
	}
}

func write(val response) {
	_ = json.NewEncoder(os.Stdout).Encode(val)
}
`, pluginName)

	goModContent := fmt.Sprintf(`module %s

go 1.24
`, pluginName)

	workflowContent := `name: Release Plugin Package

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      joss_channel:
        description: Canal del compilador Joss CLI (stable / canary)
        required: true
        default: stable
        type: choice
        options:
          - stable
          - canary
      release_type:
        description: Tipo de release en GitHub (draft, prerelease, published)
        required: true
        default: published
        type: choice
        options:
          - draft
          - prerelease
          - published

permissions:
  contents: write

jobs:
  release:
    name: Build Native Sidecars & Package .jp
    runs-on: windows-latest
    steps:
      - name: Checkout plugin repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Download Joss CLI release compiler
        shell: pwsh
        env:
          JOSS_CHANNEL: ${{ inputs.joss_channel || 'stable' }}
        run: |
          $channel = $env:JOSS_CHANNEL.ToLower()
          Write-Host "Obteniendo compilador de Joss para el canal: $channel"

          $releasesUrl = "https://api.github.com/repos/joss-language/Joss-Programming-Language/releases"
          try {
            $releases = Invoke-RestMethod -Uri $releasesUrl -Headers @{ "User-Agent" = "Joss-Plugin-Builder" }
          } catch {
            $releasesUrl = "https://api.github.com/repos/josprox/Joss-language/releases"
            $releases = Invoke-RestMethod -Uri $releasesUrl -Headers @{ "User-Agent" = "Joss-Plugin-Builder" }
          }

          $targetRelease = $null
          if ($channel -eq 'canary') {
            $targetRelease = $releases | Where-Object { $_.prerelease -eq $true -and $_.draft -eq $false } | Select-Object -First 1
            if (-not $targetRelease) {
              $targetRelease = $releases | Where-Object { $_.draft -eq $false } | Select-Object -First 1
            }
          } else {
            $targetRelease = $releases | Where-Object { $_.prerelease -eq $false -and $_.draft -eq $false } | Select-Object -First 1
            if (-not $targetRelease) {
              $targetRelease = $releases | Where-Object { $_.draft -eq $false } | Select-Object -First 1
            }
          }

          if (-not $targetRelease) {
            throw "No se encontró ningún release disponible para el canal '$channel'"
          }

          Write-Host "Release seleccionado: $($targetRelease.name) ($($targetRelease.tag_name))"

          $asset = $targetRelease.assets | Where-Object { $_.name -like '*windows.zip*' -or $_.name -eq 'joss.exe' -or $_.name -like '*windows-amd64*' } | Select-Object -First 1
          if (-not $asset) {
            $asset = $targetRelease.assets | Where-Object { $_.name -like '*.zip' -or $_.name -like '*.exe' } | Select-Object -First 1
          }

          if (-not $asset) {
            throw "No se encontró asset de Windows en el release $($targetRelease.tag_name)"
          }

          $downloadUrl = $asset.browser_download_url
          Write-Host "Descargando compilador Joss desde: $downloadUrl ($($asset.name))"
          Invoke-WebRequest -Uri $downloadUrl -OutFile "downloaded_joss_pkg" -UserAgent "Mozilla/5.0"

          if ($asset.name -like '*.zip') {
            Expand-Archive -LiteralPath "downloaded_joss_pkg" -DestinationPath "temp_joss" -Force
            $exe = Get-ChildItem -Path "temp_joss" -Recurse -Filter "joss.exe" | Select-Object -First 1
            if (-not $exe) {
              $exe = Get-ChildItem -Path "temp_joss" -Recurse -Filter "joss*" | Where-Object { ! $_.PSIsContainer } | Select-Object -First 1
            }
            if (-not $exe) {
              throw "No se encontró el ejecutable joss.exe dentro del ZIP $($asset.name)"
            }
            Copy-Item $exe.FullName "joss.exe" -Force
          } else {
            Move-Item "downloaded_joss_pkg" "joss.exe" -Force
          }

          ./joss.exe version

      - name: Parse plugin name and version from joss.yaml
        id: plugin_meta
        shell: pwsh
        run: |
          $manifest = Get-Content 'joss.yaml' -Raw
          if ($manifest -match 'name:\s*([^\r\n]+)') {
            $name = $Matches[1].Trim()
          } else {
            throw 'joss.yaml no declara name'
          }
          if ($manifest -match 'version:\s*([^\r\n]+)') {
            $ver = $Matches[1].Trim()
          } else {
            throw 'joss.yaml no declara version'
          }
          Write-Host "Plugin: $name v$ver"
          "name=$name" >> $env:GITHUB_OUTPUT
          "version=$ver" >> $env:GITHUB_OUTPUT
          $tag = if ($env:GITHUB_REF_TYPE -eq 'tag') { $env:GITHUB_REF_NAME } else { "v$ver" }
          "tag=$tag" >> $env:GITHUB_OUTPUT

      - name: Cross-compile native sidecars
        shell: pwsh
        run: |
          $pluginName = '${{ steps.plugin_meta.outputs.name }}'
          $targets = @(
              @{ os = 'windows'; arch = 'amd64'; ext = '.exe' },
              @{ os = 'windows'; arch = 'arm64'; ext = '.exe' },
              @{ os = 'linux';   arch = 'amd64'; ext = '' },
              @{ os = 'linux';   arch = 'arm64'; ext = '' },
              @{ os = 'darwin';  arch = 'amd64'; ext = '' },
              @{ os = 'darwin';  arch = 'arm64'; ext = '' }
          )
          $oldCGO = $env:CGO_ENABLED
          $env:CGO_ENABLED = '0'
          try {
            foreach ($t in $targets) {
              $os = $t.os
              $arch = $t.arch
              $ext = $t.ext
              $outDir = Join-Path $PWD ("native\" + $os + "-" + $arch)
              New-Item -ItemType Directory -Force -Path $outDir | Out-Null
              $outFile = Join-Path $outDir ($pluginName + $ext)
              $env:GOOS = $os
              $env:GOARCH = $arch
              Write-Host "Compilando sidecar ${os}-${arch} => $outFile"
              go build -trimpath -o $outFile ./cmd/sidecar
              if ($LASTEXITCODE -ne 0) {
                throw "Fallo compilacion sidecar para ${os}-${arch}"
              }
            }
          } finally {
            $env:CGO_ENABLED = $oldCGO
            $env:GOOS = ''
            $env:GOARCH = ''
          }

      - name: Pack .jp package
        shell: pwsh
        run: |
          $pluginName = '${{ steps.plugin_meta.outputs.name }}'
          ./joss.exe plugin compile .
          $jpFile = "$pluginName.jp"
          if (-not (Test-Path -LiteralPath $jpFile)) {
            throw "No se genero $jpFile"
          }

      - name: Generate SHA-256 checksum
        shell: pwsh
        run: |
          $pluginName = '${{ steps.plugin_meta.outputs.name }}'
          $jpFile = "$pluginName.jp"
          $hash = (Get-FileHash -LiteralPath $jpFile -Algorithm SHA256).Hash.ToLowerInvariant()
          "$hash  $jpFile" | Out-File -FilePath "SHA256SUMS.txt" -Encoding utf8

      - name: Create or update GitHub Release
        shell: pwsh
        env:
          GH_TOKEN: ${{ github.token }}
          GH_REPO: ${{ github.repository }}
          RELEASE_TAG: ${{ steps.plugin_meta.outputs.tag }}
          RELEASE_TYPE: ${{ inputs.release_type || 'published' }}
          PLUGIN_NAME: ${{ steps.plugin_meta.outputs.name }}
          PLUGIN_VERSION: ${{ steps.plugin_meta.outputs.version }}
        run: |
          $jpFile = "$env:PLUGIN_NAME.jp"
          $assets = @($jpFile, "SHA256SUMS.txt")

          $existing = gh release view $env:RELEASE_TAG --repo $env:GH_REPO --json tagName 2>$null
          if ($LASTEXITCODE -eq 0) {
            Write-Host "Actualizando Release existente $env:RELEASE_TAG..."
            gh release upload $env:RELEASE_TAG @assets --repo $env:GH_REPO --clobber
          } else {
            Write-Host "Creando nueva release $env:RELEASE_TAG ($env:RELEASE_TYPE)..."
            $flags = @($env:RELEASE_TAG, '--repo', $env:GH_REPO, '--title', "$env:PLUGIN_NAME v$env:PLUGIN_VERSION", '--generate-notes')
            if ($env:RELEASE_TYPE -eq 'draft') {
              $flags += '--draft'
            } elseif ($env:RELEASE_TYPE -eq 'prerelease') {
              $flags += '--prerelease'
            }
            $flags += $assets
            gh release create @flags
          }
`

	gitignoreContent := `native/
*.jp
*.exe
`

	readmeContent := fmt.Sprintf(`# %s

Plugin oficial para Joss Language.

## Estructura
- ` + "`" + `joss.yaml` + "`" + `: Manifiesto del plugin.
- ` + "`" + `src/plugin.joss` + "`" + `: Clase exportada en Joss.
- ` + "`" + `cmd/sidecar/main.go` + "`" + `: Ejecutable nativo RPC (sidecar).
- ` + "`" + `.github/workflows/release.yml` + "`" + `: Compilación cruzada y GitHub Releases.
`, pluginName)

	_ = os.WriteFile(filepath.Join(targetPath, "joss.yaml"), []byte(manifestContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "src", "plugin.joss"), []byte(pluginJossContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "cmd", "sidecar", "main.go"), []byte(sidecarGoContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "go.mod"), []byte(goModContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, ".github", "workflows", "release.yml"), []byte(workflowContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, ".gitignore"), []byte(gitignoreContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "README.md"), []byte(readmeContent), 0644)

	fmt.Printf("✨ Plugin '%s' creado exitosamente en '%s'\n", pluginName, targetPath)
}
