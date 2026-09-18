#!/bin/bash

# =====================================================
# Deploy Task Management API ke Minikube
# =====================================================

set -e

NAMESPACE="task-api"
API_IMAGE_NAME="task-management-api"
API_IMAGE_TAG="latest"
IMAGE_ARCHIVE="/tmp/task-api-image.tar"
BINARY_OUT="/Users/a2292/task-management-api/task-manager"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

echo "=============================================="
echo "  Deploy Task Management API → Minikube"
echo "=============================================="
echo ""

# ---- Step 1: Cek Minikube ----
echo ""
echo -e "${YELLOW}Step 1/7: Memeriksa Minikube...${NC}"

if ! command -v minikube &> /dev/null; then
    error "Minikube tidak terinstall. Install: brew install minikube"
fi

if kubectl cluster-info &> /dev/null; then
    info "Minikube sudah running"
else
    warn "Minikube belum ready..."
    minikube start --driver=docker
    info "Minikube siap!"
fi

# ---- Step 2: Setup Host Docker ----
echo ""
echo -e "${YELLOW}Step 2/7: Setup Docker environment...${NC}"
unset DOCKER_HOST DOCKER_TLS_VERIFY DOCKER_CERT_PATH DOCKER_BUILDKIT
if [ -n "$MINIKUBE_ACTIVE_DOCKERD" ]; then
    eval $(minikube docker-env --unset) >/dev/null 2>&1 || true
fi
info "Host Docker aktif"

# ---- Step 3: Build Binary + Image ----
echo ""
echo -e "${YELLOW}Step 3/7: Build Go binary...${NC}"

GOCACHE=/tmp/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o "${BINARY_OUT}" ./cmd/api/main.go 2>&1

if [ ! -f "${BINARY_OUT}" ]; then
    error "Binary build gagal!"
fi
info "Binary berhasil: $(ls -lh ${BINARY_OUT} | awk '{print $5}')"

echo ""
echo -e "${YELLOW}Step 4/7: Build Docker image...${NC}"

DOCKER_BUILDKIT=0 docker build -t ${API_IMAGE_NAME}:${API_IMAGE_TAG} . 2>&1
info "Image built: ${API_IMAGE_NAME}:${API_IMAGE_TAG}"

# ---- Step 5: Load ke Minikube ----
echo ""
echo -e "${YELLOW}Step 5/7: Load image ke Minikube...${NC}"

rm -f "${IMAGE_ARCHIVE}"
docker save -o "${IMAGE_ARCHIVE}" ${API_IMAGE_NAME}:${API_IMAGE_TAG}
minikube image load --daemon=false "${IMAGE_ARCHIVE}"
rm -f "${IMAGE_ARCHIVE}"
info "Image loaded ke Minikube"

# ---- Step 6: Deploy manifests ----
echo ""
echo -e "${YELLOW}Step 6/7: Deploy Kubernetes manifests...${NC}"

# Hapus service lama jika ada (untuk menghindari nodePort conflict)
kubectl delete svc task-api -n ${NAMESPACE} --ignore-not-found >/dev/null 2>&1 || true

kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

manifests=(
    "01-secret.yaml"
    "03-postgres-pvc.yaml"
    "04-postgres.yaml"
    "05-init-script-cm.yaml"
    "06-api-deployment.yaml"
    "07-api-service.yaml"
)

for manifest in "${manifests[@]}"; do
    filepath="k8s/${manifest}"
    if [ ! -f "$filepath" ]; then
        error "Manifest not found: $filepath"
    fi
    echo "  → ${manifest}"
    kubectl apply -f "$filepath" --namespace=${NAMESPACE}
done
info "Semua manifests deployed!"

# ---- Step 7: Tunggu ready ----
echo ""
echo -e "${YELLOW}Step 7/7: Menunggu pods ready...${NC}"

echo "  ⏳ PostgreSQL..."
kubectl wait --namespace=${NAMESPACE} \
    --for=condition=ready pod \
    --selector=app=postgres \
    --timeout=120s && info "  PostgreSQL ✓" || warn "  PostgreSQL timeout"

echo "  ⏳ API server..."
kubectl rollout status --namespace=${NAMESPACE} \
    deployment/task-api \
    --timeout=120s && info "  API ✓" || warn "  API timeout"

echo ""
echo "─ Status Pods ─"
kubectl get pods -n ${NAMESPACE}

echo ""
echo "─ Services ─"
kubectl get svc -n ${NAMESPACE}

echo ""
echo "─ PVC ─"
kubectl get pvc -n ${NAMESPACE}

# ---- Final Output ----
API_PORT=$(kubectl get svc task-api --namespace=${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}')
MINIKUBE_IP=$(minikube ip)

echo ""
echo "=============================================="
echo -e "${GREEN}${YELLOW}  ✅ DEPLOYMENT SELESAI!${NC}"
echo "=============================================="
echo ""
echo "  📍 API URL:     http://${MINIKUBE_IP}:${API_PORT}"
echo "  📍 Health:      http://${MINIKUBE_IP}:${API_PORT}/health"
echo ""
echo "  Quick commands:"
echo "  kubectl logs -n ${NAMESPACE} -l app=task-api -f"
echo "  ./k8s/dashboard.sh"
echo ""
