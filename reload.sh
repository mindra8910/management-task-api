#!/bin/bash

# =====================================================
# Reload: Rebuild binary + rolling update
# Gunakan setelah setiap code change
# =====================================================

set -e

NAMESPACE="task-api"
API_IMAGE_NAME="task-management-api"
API_IMAGE_TAG="latest"
BINARY_OUT="/Users/a2292/task-management-api/task-manager"
IMAGE_ARCHIVE="/tmp/task-api-image.tar"

info() { echo -e "\033[0;32m[✓]\033[0m $1"; }
warn() { echo -e "\033[1;33m[!]\033[0m $1"; }
error() { echo -e "\033[0;31m[✗]\033[0m $1"; exit 1; }

echo "🔄 Reloading & rolling update..."
echo ""

# Ensure Host Docker aktif
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_PATH DOCKER_BUILDKIT
if [ -n "$MINIKUBE_ACTIVE_DOCKERD" ]; then
    eval $(minikube docker-env --unset) >/dev/null 2>&1 || true
fi

# Build binary baru
echo "Step 1/5: Build Go binary..."
GOCACHE=/tmp/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o "${BINARY_OUT}" ./cmd/api/main.go 2>&1
if [ ! -f "${BINARY_OUT}" ]; then error "Build gagal!"; fi
info "Binary built"

# Rebuild image
echo "Step 2/5: Build Docker image..."
DOCKER_BUILDKIT=0 docker build -t ${API_IMAGE_NAME}:${API_IMAGE_TAG} . 2>&1
info "Image rebuilt"

# Load ke Minikube
echo "Step 3/5: Load image ke Minikube..."
rm -f "${IMAGE_ARCHIVE}"
docker save -o "${IMAGE_ARCHIVE}" ${API_IMAGE_NAME}:${API_IMAGE_TAG}
minikube image load --daemon=false "${IMAGE_ARCHIVE}"
rm -f "${IMAGE_ARCHIVE}"
info "Image loaded"

# Rolling update
echo "Step 4/5: Trigger rolling update..."
kubectl rollout restart deployment/task-api -n ${NAMESPACE}
info "Rolling update started"

# Wait
echo "Step 5/5: Waiting for pods..."
kubectl rollout status -n ${NAMESPACE} deployment/task-api --timeout=120s
info "Rolling update complete!"

echo ""
./k8s/dashboard.sh
