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

public class %s {
    Init constructor() {
        // Acceso directo y automático al entorno (.env) del proyecto
        $this->apiKey = env("%s_API_KEY", "")
    }

    public func ping() {
        return "pong desde %s"
    }

    public func process(mixed $data) {
        return {
            "ok": true,
            "plugin": "%s",
            "data": $data
        }
    }
}
`, pluginName, className, strings.ToUpper(pluginName), pluginName, pluginName)

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
    runs-on: ubuntu-latest
    steps:
      - name: Checkout plugin repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Download Joss CLI & update to canary
        shell: bash
        run: |
          ASSET_URL=$(python3 <<'PYEOF'
          import urllib.request, json, sys
          url = "https://api.github.com/repos/joss-language/Joss-Programming-Language/releases"
          req = urllib.request.Request(url, headers={"User-Agent": "Joss-Plugin-Builder"})
          releases = json.loads(urllib.request.urlopen(req).read())
          rel = next((r for r in releases if r["prerelease"] and not r["draft"]), None)
          if not rel:
              rel = next((r for r in releases if not r["draft"]), None)
          if not rel: sys.exit("No release found")
          asset = next((a for a in rel["assets"] if "jossecurity-linux.zip" in a["name"]), None)
          if not asset:
              asset = next((a for a in rel["assets"] if "linux" in a["name"] and "amd64" in a["name"]), None)
          if not asset:
              asset = next((a for a in rel["assets"] if "linux" in a["name"]), None)
          if asset: print(asset["browser_download_url"])
          PYEOF
          )

          echo "Descargando Joss CLI base desde: $ASSET_URL"
          curl -sL "$ASSET_URL" -o joss_pkg
          unzip -o joss_pkg -d joss_extract
          JOSS_BIN=$(find joss_extract -type f -name 'joss-linux-amd64' | head -1)
          if [ -z "$JOSS_BIN" ]; then
            JOSS_BIN=$(find joss_extract -type f -name 'joss' ! -name '*.exe' | head -1)
          fi
          if [ -z "$JOSS_BIN" ]; then
            JOSS_BIN=$(find joss_extract -type f ! -name '*arm*' ! -name '*.sh' ! -name 'LICENSE*' ! -name 'README*' ! -name '*.exe' | head -1)
          fi
          cp "$JOSS_BIN" ./joss
          rm -rf joss_extract joss_pkg
          chmod +x ./joss
          mv ./joss /usr/local/bin/joss 2>/dev/null || true

          echo "Actualizando Joss CLI a modo canary..."
          joss update --canary -f || ./joss update --canary -f || true
          joss version || ./joss version

      - name: Parse plugin metadata from joss.yaml
        id: meta
        shell: bash
        run: |
          python3 <<'PYEOF'
          import os, sys
          with open('joss.yaml', encoding='utf-8-sig') as f:
              lines = f.read().splitlines()
          vals = {}
          for line in lines:
              if ':' in line and not line.startswith(' '):
                  k, _, v = line.partition(':')
                  vals[k.strip()] = v.strip().strip('"').strip("'")
          name = vals.get('name', '')
          ver  = vals.get('version', '')
          if not name or not ver:
              print(f'ERROR: name={name!r} version={ver!r}', file=sys.stderr)
              sys.exit(1)
          ref_type = os.environ.get('GITHUB_REF_TYPE', '')
          ref_name = os.environ.get('GITHUB_REF_NAME', '')
          tag = ref_name if ref_type == 'tag' else f'v{ver}'
          with open(os.environ['GITHUB_OUTPUT'], 'a') as f:
              f.write(f'name={name}\n')
              f.write(f'version={ver}\n')
              f.write(f'tag={tag}\n')
          print(f'Plugin: {name} v{ver}  tag={tag}')
          PYEOF

      - name: Compile pure JPBC .jp package
        shell: bash
        run: |
          NAME="${{ steps.meta.outputs.name }}"
          joss plugin compile . || ./joss plugin compile .
          JP="${NAME}.jp"
          if [ ! -f "$JP" ]; then
            echo "ERROR: $JP no se genero"
            exit 1
          fi
          echo "OK: ${JP} generado ($(du -sh "$JP" | cut -f1))"

      - name: Generate SHA-256 checksum
        shell: bash
        run: |
          sha256sum "${{ steps.meta.outputs.name }}.jp" > SHA256SUMS.txt
          cat SHA256SUMS.txt

      - name: Publish or update GitHub Release
        shell: bash
        env:
          GH_TOKEN: ${{ github.token }}
          GH_REPO: ${{ github.repository }}
          TAG:          ${{ steps.meta.outputs.tag }}
          NAME:         ${{ steps.meta.outputs.name }}
          VER:          ${{ steps.meta.outputs.version }}
          RELEASE_TYPE: ${{ inputs.release_type || 'published' }}
        run: |
          JP="${NAME}.jp"
          ASSETS=("$JP" "SHA256SUMS.txt")

          # Buscar si ya existe una release con este tag (draft, prerelease o published)
          if gh release view "$TAG" --repo "$GH_REPO" &>/dev/null; then
            echo "La release '$TAG' ya existe. Actualizando archivos y estado..."

            # Reemplazar archivos existentes
            gh release upload "$TAG" "${ASSETS[@]}" --repo "$GH_REPO" --clobber

            # Actualizar tipo de release segun el input
            if [ "$RELEASE_TYPE" = "draft" ]; then
              gh release edit "$TAG" --repo "$GH_REPO" --draft=true  --prerelease=false
            elif [ "$RELEASE_TYPE" = "prerelease" ]; then
              gh release edit "$TAG" --repo "$GH_REPO" --draft=false --prerelease=true
            else
              gh release edit "$TAG" --repo "$GH_REPO" --draft=false --prerelease=false
            fi

            echo "Release '$TAG' actualizada ($RELEASE_TYPE)."
          else
            echo "Creando nueva release '$TAG' ($RELEASE_TYPE)..."
            FLAGS=("$TAG" --repo "$GH_REPO" --title "$NAME v$VER" --generate-notes)
            [ "$RELEASE_TYPE" = "draft" ]      && FLAGS+=(--draft)
            [ "$RELEASE_TYPE" = "prerelease" ] && FLAGS+=(--prerelease)
            gh release create "${FLAGS[@]}" "${ASSETS[@]}"
            echo "Release '$TAG' creada ($RELEASE_TYPE)."
          fi
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
