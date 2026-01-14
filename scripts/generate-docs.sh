#!/bin/bash

# Generate HTML documentation from OpenAPI spec

# Create docs directory if it doesn't exist
mkdir -p ./docs/html

# Check if swagger-ui-dist is installed
if [ ! -d "node_modules/swagger-ui-dist" ]; then
    echo "Installing swagger-ui-dist..."
    npm install swagger-ui-dist
fi

# Copy swagger-ui files to docs/html
cp -r node_modules/swagger-ui-dist/* ./docs/html

# Create custom index.html that uses our OpenAPI spec
cat > ./docs/html/index.html << EOF
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
EOF

echo "HTML documentation generated successfully in ./docs/html/"
echo "You can open ./docs/html/index.html in your browser to view the documentation."
