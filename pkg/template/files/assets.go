package files

import "path/filepath"

func GetAssetFiles(path string) map[string]string {
	return map[string]string{
		filepath.Join(path, "assets", "js", "app.js"): `console.log('JosSecurity Enterprise v3.0 - Inicializado');`,

		// SCSS Modules
		filepath.Join(path, "assets", "css", "_variables.scss"): `$primary: #2563eb;
$primary-dark: #1e40af;
$secondary: #64748b;
$background: #f8fafc;
$surface: #ffffff;
$text: #1e293b;
$text-light: #64748b;
$border: #e2e8f0;
$danger: #ef4444;
$success: #22c55e;
$info: #3b82f6;
`,
		filepath.Join(path, "assets", "css", "_layout.scss"): `body {
    font-family: 'Inter', sans-serif;
    background-color: $background;
    color: $text;
    margin: 0;
    line-height: 1.6;
}

.navbar {
    background-color: $surface;
    border-bottom: 1px solid $border;
    padding: 1rem 0;
    margin-bottom: 2rem;
}

.container-nav {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.brand {
    font-size: 1.5rem;
    font-weight: 700;
    color: $primary;
    text-decoration: none;
}

.nav-links a {
    margin-left: 1.5rem;
    text-decoration: none;
    color: $text;
    font-weight: 500;
    transition: color 0.2s;
}

.nav-links a:hover {
    color: $primary;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 1rem;
    min-height: 80vh;
}

.footer {
    text-align: center;
    padding: 2rem 0;
    color: $text-light;
    border-top: 1px solid $border;
    margin-top: auto;
}

.text-center { text-align: center; }
.d-flex { display: flex; }
.justify-content-center { justify-content: center; }
.justify-content-between { justify-content: space-between; }
.align-items-center { align-items: center; }
.gap-3 { gap: 1rem; }
.mt-4 { margin-top: 1.5rem; }
.mb-4 { margin-bottom: 1.5rem; }
.mb-5 { margin-bottom: 3rem; }
`,
		filepath.Join(path, "assets", "css", "_components.scss"): `.card {
    background: $surface;
    border-radius: 0.75rem;
    box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
    margin-bottom: 2rem;
    overflow: hidden;
}

.card-header {
    padding: 1.5rem;
    border-bottom: 1px solid $border;
    background: #f1f5f9;
}

.card-header h2 {
    margin: 0;
    font-size: 1.25rem;
}

.card-body {
    padding: 2rem;
}

.card-footer {
    padding: 1rem 2rem;
    background: #f8fafc;
    border-top: 1px solid $border;
}

.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    border-radius: 0.5rem;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
}

.btn-primary {
    background-color: $primary;
    color: white;
}

.btn-primary:hover {
    background-color: $primary-dark;
}

.btn-outline-light {
    border: 2px solid $primary;
    color: $primary;
    background: transparent;
}

.btn-outline-light:hover {
    background: $primary;
    color: white;
}

.btn-outline-danger {
    border: 1px solid $danger;
    color: $danger;
    background: transparent;
    padding: 0.5rem 1rem;
}

.btn-outline-danger:hover {
    background: $danger;
    color: white;
}

.btn-block {
    display: block;
    width: 100%;
    text-align: center;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
}

.form-control {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid $border;
    border-radius: 0.5rem;
    font-family: inherit;
    font-size: 1rem;
    box-sizing: border-box;
}

.form-control:focus {
    outline: none;
    border-color: $primary;
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.alert {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
}

.alert-danger {
    background-color: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
}

.alert-success {
    background-color: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
}

.alert-info {
    background-color: #eff6ff;
    color: #1e40af;
    border: 1px solid #dbeafe;
}

.badge {
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 600;
}

.badge-info {
    background-color: #e0f2fe;
    color: #0369a1;
}

.stat-card {
    background: #f8fafc;
    padding: 1.5rem;
    border-radius: 0.5rem;
    text-align: center;
    border: 1px solid $border;
}

.stat-number {
    font-size: 2.5rem;
    font-weight: 700;
    color: $primary;
    margin: 0.5rem 0 0;
}
`,
		filepath.Join(path, "assets", "css", "app.scss"): `// Main SCSS Entry Point
@import "variables";
@import "layout";
@import "components";
`,
	}
}
