#!/bin/bash
# Build script for simple-redirects Docker image
# Reads Go version from .tool-versions as source of truth

set -e

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Read Go version from .tool-versions
if [ -f "${PROJECT_ROOT}/.tool-versions" ]; then
  GO_VERSION_FROM_FILE=$(grep "^golang" "${PROJECT_ROOT}/.tool-versions" | awk '{print $2}')
  GO_VERSION="${GO_VERSION:-${GO_VERSION_FROM_FILE}}"
else
  echo "Error: .tool-versions file not found"
  exit 1
fi

# Default nginx version (can be overridden via environment variable)
NGINX_VERSION="${NGINX_VERSION:-1.27.3}"
IMAGE_NAME="${IMAGE_NAME:-simple-redirects}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Building Docker image: ${IMAGE_NAME}:${IMAGE_TAG}${NC}"
echo -e "${BLUE}Go version: ${GO_VERSION} ${YELLOW}(from .tool-versions)${NC}"
echo -e "${BLUE}Nginx version: ${NGINX_VERSION}${NC}"
echo ""

# Build the image from project root
cd "${PROJECT_ROOT}"
docker build \
  --build-arg GO_VERSION="${GO_VERSION}" \
  --build-arg NGINX_VERSION="${NGINX_VERSION}" \
  -t "${IMAGE_NAME}:${IMAGE_TAG}" \
  .

echo ""
echo -e "${GREEN}✓ Build completed successfully${NC}"
echo -e "${GREEN}Image: ${IMAGE_NAME}:${IMAGE_TAG}${NC}"

# Show image size
docker images "${IMAGE_NAME}:${IMAGE_TAG}" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
