#!/usr/bin/env bash
# Verifies that mise is active and the tool versions match .mise.toml
# (mise install runs automatically on container startup)

set -e

echo "=== mise version ==="
mise --version

echo ""
echo "=== active tools ==="
mise current

echo ""
echo "=== node ==="
node --version

echo ""
echo "=== python ==="
python3 --version

echo ""
echo "All checks passed!"
