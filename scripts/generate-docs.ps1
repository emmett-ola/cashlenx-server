#!/usr/bin/env pwsh

# Generate HTML documentation from OpenAPI spec

# Create docs directory if it doesn't exist
$docsHtmlPath = Join-Path -Path $PSScriptRoot -ChildPath "html"
if (-not (Test-Path -Path $docsHtmlPath -PathType Container)) {
    New-Item -Path $docsHtmlPath -ItemType Directory -Force | Out-Null
}

# Check if swagger-ui-dist is installed
$nodeModulesPath = Join-Path -Path $PSScriptRoot -ChildPath ".." | Join-Path -ChildPath "node_modules"
$swaggerUiPath = Join-Path -Path $nodeModulesPath -ChildPath "swagger-ui-dist"

if (-not (Test-Path -Path $swaggerUiPath -PathType Container)) {
    Write-Host "Installing swagger-ui-dist..."
    Set-Location -Path (Join-Path -Path $PSScriptRoot -ChildPath "..")
    npm install swagger-ui-dist
}

# Copy swagger-ui files to docs/html
Copy-Item -Path "$swaggerUiPath/*" -Destination $docsHtmlPath -Recurse -Force

# Create custom index.html that uses our OpenAPI spec
$indexHtmlContent = @"
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>CashLenX API Documentation</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
  <style>
    body {
      margin: 0;
      padding: 0;
    }
    .swagger-ui .topbar {
      background-color: #2c3e50;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="./swagger-ui-bundle.js"></script>
  <script src="./swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: '../openapi.yaml',
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>
"@

$indexHtmlPath = Join-Path -Path $docsHtmlPath -ChildPath "index.html"
Set-Content -Path $indexHtmlPath -Value $indexHtmlContent -Force

Write-Host "HTML documentation generated successfully in $docsHtmlPath"
Write-Host "You can open $indexHtmlPath in your browser to view the documentation."
