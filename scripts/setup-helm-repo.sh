#!/bin/bash
set -e

echo "Setting up Helm repository..."

# Create charts directory
mkdir -p charts

# Package the Helm chart
helm package helm/circuit-breaker-controller --destination ./charts

# Create index.yaml
helm repo index ./charts --url https://piuschungath.github.io/circuit-breaker

echo "Helm repository files created in ./charts/"
echo "To publish:"
echo "1. Enable GitHub Pages in repository settings"
echo "2. Set source to 'gh-pages' branch"
echo "3. Push charts to gh-pages branch"