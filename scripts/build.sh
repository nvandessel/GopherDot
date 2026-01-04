#!/bin/bash
set -e

# Configuration
APP_NAME="g4d"
BUILD_DIR="bin"
DIST_DIR="dist"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")

echo "📦 Packaging $APP_NAME version $VERSION..."

# Clean and prepare
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# Build all binaries using Makefile
make build-all

# Package binaries
echo "🗜️  Compressing binaries..."

for binary in "$BUILD_DIR"/*; do
    filename=$(basename "$binary")
    
    # Skip if directory
    if [ -d "$binary" ]; then continue; fi

    echo "   Processing $filename..."
    
    if [[ "$filename" == *".exe" ]]; then
        # Windows - zip
        zip -j "$DIST_DIR/${filename%.exe}.zip" "$binary"
    else
        # Linux/macOS - tar.gz
        tar -czf "$DIST_DIR/$filename.tar.gz" -C "$BUILD_DIR" "$filename"
    fi
done

# Generate checksums
echo "🔒 Generating checksums..."
cd "$DIST_DIR"
sha256sum * > checksums.txt
cd ..

echo "✅ Build and packaging complete! Artifacts in $DIST_DIR/"
ls -lh "$DIST_DIR"
