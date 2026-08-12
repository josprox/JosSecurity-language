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

	fmt.Printf("📦 Creando nuevo plugin oficial '%s' en '%s' (Bytecode JPBC puro)...\n", pluginName, targetPath)

	dirs := []string{
		targetPath,
		filepath.Join(targetPath, "src"),
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
description: Plugin oficial en Bytecode JPBC autoejecutable para Joss Language
author: Developer <dev@example.com>
license: MIT
entry:
  main: src/plugin.joss
dependencies:
`, pluginName)

	pluginJossContent := fmt.Sprintf(`// src/plugin.joss
// Clase exportada para interactuar con el plugin %s

class %s {
    func ping() {
        return "pong desde %s"
    }

    func process($data) {
        return {
            "ok": true,
            "plugin": "%s",
            "data": $data
        }
    }
}
`, pluginName, className, pluginName, pluginName)

	workflowContent := `name: Release Plugin Package (.jp)

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
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
    name: Build & Package .jp (Bytecode JPBC)
    runs-on: windows-latest
    steps:
      - name: Checkout plugin repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Download Joss CLI compiler
        shell: pwsh
        run: |
          $releasesUrl = "https://api.github.com/repos/joss-language/Joss-Programming-Language/releases"
          try {
            $releases = Invoke-RestMethod -Uri $releasesUrl -Headers @{ "User-Agent" = "Joss-Plugin-Builder" }
          } catch {
            $releasesUrl = "https://api.github.com/repos/josprox/Joss-language/releases"
            $releases = Invoke-RestMethod -Uri $releasesUrl -Headers @{ "User-Agent" = "Joss-Plugin-Builder" }
          }

          $targetRelease = $releases | Where-Object { $_.draft -eq $false } | Select-Object -First 1
          if (-not $targetRelease) {
            throw "No se encontró ningún release de Joss CLI disponible"
          }

          $asset = $targetRelease.assets | Where-Object { $_.name -like '*windows.zip*' -or $_.name -eq 'joss.exe' -or $_.name -like '*windows-amd64*' } | Select-Object -First 1
          if (-not $asset) {
            $asset = $targetRelease.assets | Where-Object { $_.name -like '*.zip' -or $_.name -like '*.exe' } | Select-Object -First 1
          }

          $downloadUrl = $asset.browser_download_url
          Write-Host "Descargando compilador Joss desde: $downloadUrl"
          Invoke-WebRequest -Uri $downloadUrl -OutFile "downloaded_joss_pkg" -UserAgent "Mozilla/5.0"

          if ($asset.name -like '*.zip') {
            Expand-Archive -LiteralPath "downloaded_joss_pkg" -DestinationPath "temp_joss" -Force
            $exe = Get-ChildItem -Path "temp_joss" -Recurse -Filter "joss.exe" | Select-Object -First 1
            Copy-Item $exe.FullName "joss.exe" -Force
          } else {
            Move-Item "downloaded_joss_pkg" "joss.exe" -Force
          }

          ./joss.exe version

      - name: Parse plugin name and version
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

      - name: Compile pure JPBC .jp package
        shell: pwsh
        run: |
          $pluginName = '${{ steps.plugin_meta.outputs.name }}'
          ./joss.exe plugin compile .
          $jpFile = "$pluginName.jp"
          if (-not (Test-Path -LiteralPath $jpFile)) {
            throw "No se generó el paquete $jpFile"
          }

      - name: Generate SHA-256 checksum
        shell: pwsh
        run: |
          $pluginName = '${{ steps.plugin_meta.outputs.name }}'
          $jpFile = "$pluginName.jp"
          $hash = (Get-FileHash -LiteralPath $jpFile -Algorithm SHA256).Hash.ToLowerInvariant()
          "$hash  $jpFile" | Out-File -FilePath "SHA256SUMS.txt" -Encoding utf8

      - name: Publish GitHub Release
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
            gh release upload $env:RELEASE_TAG @assets --repo $env:GH_REPO --clobber
          } else {
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

	gitignoreContent := `*.jp
*.exe
`

	readmeContent := fmt.Sprintf("# %s\n\nPlugin oficial en Bytecode JPBC puro para Joss Language.\n\n## Estructura\n- `joss.yaml`: Manifiesto del paquete.\n- `src/plugin.joss`: Implementación y clase exportada en Joss.\n- `.github/workflows/release.yml`: Compilación a .jp y publicación automática en GitHub Releases.\n\n## Compilar a paquete .jp\n\n```bash\njoss plugin compile .\n```\n", pluginName)

	_ = os.WriteFile(filepath.Join(targetPath, "joss.yaml"), []byte(manifestContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "src", "plugin.joss"), []byte(pluginJossContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, ".github", "workflows", "release.yml"), []byte(workflowContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, ".gitignore"), []byte(gitignoreContent), 0644)
	_ = os.WriteFile(filepath.Join(targetPath, "README.md"), []byte(readmeContent), 0644)

	fmt.Printf("✨ Plugin '%s' creado exitosamente en '%s' (únicamente Bytecode .jp autoejecutable)\n", pluginName, targetPath)
}
