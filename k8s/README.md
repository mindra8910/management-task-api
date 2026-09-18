# Kubernetes Deployment Guide - Task Management API

Panduan lengkap deploy ke Minikube.

## 📋 Prerequisites

- **Minikube** — install: `brew install minikube`
- **Docker** — install: `brew install --cask docker`
- **kubectl** — install: `brew install kubectl`
- **jq** — install: `brew install jq`

## 🚀 Quick Start

### 1. Start Minikube

```bash
minikube start --driver=docker --memory=4096 --cpus=2
```

### 2. Deploy Semua Service (API + PostgreSQL)

```bash
./deploy.sh
```

Script ini akan otomatis:
- Mengecek & start Minikube jika belum running
- Set Docker context ke Minikube
- Build Docker image untuk Minikube
- Apply semua Kubernetes manifests
- Menunggu pods ready
- Tampilkan URL API

### 3. Akses Dashboard

```bash
./k8s/dashboard.sh
```

Tampilkan status pods, services, health check, dan quick commands.

## 🗂️ Struktur File

```
task-management-api/
├── k8s/
│   ├── 00-namespace.yaml       # Namespace task-api
│   ├── 01-secret.yaml          # Secret (DB password, JWT secret, DB_DSN)
│   ├── 03-postgres-pvc.yaml    # PersistentVolumeClaim untuk database
│   ├── 04-postgres.yaml        # PostgreSQL Deployment + Service
│   ├── 05-init-script-cm.yaml  # ConfigMap inisialisasi schema DB
│   ├── 06-api-deployment.yaml  # API Deployment (2 replicas)
│   ├── 07-api-service.yaml     # API Service (NodePort 30080)
│   └── dashboard.sh            # Dashboard command
├── deploy.sh                   # Deploy script utama
├── reload.sh                   # Reload image setelah code change
├── Dockerfile                  # Production Docker build
└── .dockerignore               # Exclude files dari Docker build
```

## 🔧 Customisasi

### Ganti Image Name

Edit `k8s/06-api-deployment.yaml`:

```yaml
spec:
  template:
    spec:
      containers:
        - name: api
          image: ghcr.io/YOUR_USERNAME/task-management-api:latest
```

Atau edit di `deploy.sh`:
```bash
API_IMAGE_NAME="your-custom-image"
```

### Scale Replicas

```bash
kubectl scale -n task-api deployment/task-api --replicas=3
```

### Upgrade Image Setelah Code Change

```bash
./reload.sh
```

Ini akan rebuild image dan trigger rolling update otomatis.

## 🐛 Troubleshooting

### Pod tidak ready / CrashLoopBackOff

```bash
# Cek log pod API
kubectl logs -n task-api -l app=task-api

# Describe pod untuk detail error
kubectl describe pod -n task-api -l app=task-api

# Cek event namespace
kubectl get events -n task-api --sort-by='.lastTimestamp'
```

### Database belum terinisialisasi

PostgreSQL akan otomatis menjalankan `init.sql` pada first boot. Jika sudah ada data:
```bash
# Masuk ke psql
kubectl exec -n task-api -l app=postgres -it -- psql -U root -d taskdb
```

### Image pull failed (jika pakai remote registry)

Jika menggunakan private registry:
```bash
kubectl create secret docker-registry regsecret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_USER \
  --docker-password=YOUR_TOKEN \
  -n task-api

# Tambahkan imagePullSecrets ke deployment
```

### Minikube Ingress (untuk production-like setup)

```bash
minikube addons enable ingress
# Buat Ingress resource
```

## 📌 Ports

| Service | Port | Access |
|---------|------|--------|
| API | 30080 (NodePort) | `http://$(minikube ip):30080` |
| PostgreSQL | 5432 (ClusterIP) | Internal only (host: `postgres`) |

## 🎯 Environment Variables

| Variable | Value | Source |
|----------|-------|--------|
| `DB_DSN` | `postgres://root:root@postgres:5432/taskdb?sslmode=disable` | Secret |
| `JWT_SECRET` | `supersecretkey` | Secret |
| `TZ` | `Asia/Jakarta` | Env |

## 🏗️ Architecture Diagram

```
                    ┌─────────────────────┐
                    │    NodePort 30080   │
                    │   (k8s Service)     │
                    └──────┬──────────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
     ┌────────▼──────┐      ┌──────────▼───────┐
     │  API Pod #1   │      │  API Pod #2      │
     │  (ReplicaSet) │      │  (ReplicaSet)    │
     └───────┬───────┘      └────────┬─────────┘
             │                       │
             └──────────┬────────────┘
                        │ ClusterIP
                        │ port: 5432
                        │ host: postgres
                 ┌──────▼───────┐
                 │ PostgreSQL   │
                 │ Deployment   │
                 └──────────────┘
                 📁 PVC: 1Gi
```
