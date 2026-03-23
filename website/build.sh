#!/bin/bash
set -e
rm -rf dist
mkdir -p dist
cp index.html dist/
cp ../shared/theme.css dist/
echo "website: built to dist/"
