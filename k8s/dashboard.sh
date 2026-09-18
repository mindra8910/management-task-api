#!/bin/bash

# =====================================================
# Helper: Akses & Debugging Minikube
# =====================================================

set -e

NAMESPACE="task-api"
API_PORT=$(kubectl get svc task-api --namespace=${NAMESPACE} -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "30080")
MINIKUBE_IP=$(minikube ip)

info() { echo -e "\033[0;32m[✓]\033[0m $1"; }
warn() { echo -e "\033[1;33m[!]\033[0m $1"; }

echo "=============================================="
echo "  Task Management API - Dashboard"
echo "=============================================="
echo ""

# Check pod status
echo "📊 Pods:"
kubectl get pods -n ${NAMESPACE}
echo ""

# Check services
echo "🔗 Services:"
kubectl get svc -n ${NAMESPACE}
echo ""

# Check deployments
echo "📦 Deployments:"
kubectl get deployments -n ${NAMESPACE}
echo ""

# Health check
echo "💚 Health Check:"
curl -s http://${MINIKUBE_IP}:${API_PORT}/health | jq . 2>/dev/null || curl -s http://${MINIKUBE_IP}:${API_PORT}/health
echo ""

# Recent events in namespace
echo "📋 Recent Events:"
kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' 2>/dev/null | tail -10
echo ""

echo "=============================================="
echo "  Quick Commands"
echo "=============================================="
echo ""
echo "  # Lihat log API (follow):"
echo "  kubectl logs -n ${NAMESPACE} -l app=task-api -f"
echo ""
echo "  # Lihat log PostgreSQL (follow):"
echo "  kubectl logs -n ${NAMESPACE} -l app=postgres -f"
echo ""
echo "  # Exec ke dalam pod API:"
echo "  kubectl exec -n ${NAMESPACE} -l app=task-api -it -- /bin/sh"
echo ""
echo "  # Exec ke dalam pod PostgreSQL:"
echo "  kubectl exec -n ${NAMESPACE} -l app=postgres -it -- psql -U root -d taskdb"
echo ""
echo "  # Scale API ke N replicas:"
echo "  kubectl scale -n ${NAMESPACE} deployment/task-api --replicas=N"
echo ""
echo "  # Restart semua pods:"
echo "  kubectl rollout restart -n ${NAMESPACE} deployment/task-api"
echo ""
echo "  # Delete & redeploy:"
echo "  kubectl delete -n ${NAMESPACE} all --all && ./deploy.sh"
echo ""
echo "  🔗 API Base URL: http://${MINIKUBE_IP}:${API_PORT}"
echo ""
