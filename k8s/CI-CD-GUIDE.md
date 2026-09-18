# CI/CD Guide — GitHub Actions + Minikube

Panduan lengkap setup CI/CD pipeline untuk Task Management API.

---

## 🎯 Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Actions                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────┐ │
│  │  Lint &   │───▶│  Build   │───▶│  Push    │───▶│Deploy │ │
│  │  Test     │    │  Docker  │    │  ghcr.io │    │  K8s  │ │
│  └──────────┘    └──────────┘    └──────────┘    └───────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │  GitHub Container      │
              │  Registry (ghcr.io)    │
              └───────────┬────────────┘
                          │
                          ▼
              ┌────────────────────────┐
              │  Minikube / K8s Cluster│
              └────────────────────────┘
```

---

## 📋 Prerequisites

1. **GitHub Account** — untuk menyimpan code & menjalankan Actions
2. **GitHub Repository** — push project ke GitHub
3. **Minikube** — sudah terinstall & running di local

---

## 🚀 Step-by-Step Setup

### Step 1: Push ke GitHub

```bash
# Inisialisasi git (jika belum)
cd task-management-api
git init
git add .
git commit -m "Initial commit"

# Buat repo di GitHub, lalu:
git remote add origin https://github.com/YOUR_USERNAME/task-management-api.git
git push -u origin main
```

### Step 2: Enable GitHub Container Registry (ghcr.io)

1. Buka **GitHub Repository** → **Settings** → **Actions** → **General**
2. Di bagian **Workflow permissions**, pilih:
   - ✅ **Read and write permissions**
   - ✅ **Allow GitHub Actions to create and approve pull requests**
3. Klik **Save**

### Step 3: Konfigurasi Minikube untuk CI/CD

Ada 2 opsi deploy ke Minikube:

#### Opsi A: Self-Hosted Runner (Recommended untuk Development)

Install GitHub Actions Runner di Mac kamu:

```bash
# 1. Buat folder untuk runner
mkdir -p ~/actions-runner && cd ~/actions-runner

# 2. Download runner (cek versi terbaru di GitHub)
curl -o actions-runner-linux-arm64.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.319.0/actions-runner-linux-arm64-2.319.0.tar.gz

# 3. Extract
tar xzf actions-runner-linux-arm64.tar.gz

# 4. Configure (ganti YOUR_TOKEN dengan token dari GitHub)
./config.sh --url https://github.com/YOUR_USERNAME/task-management-api \
  --token YOUR_TOKEN \
  --name "minikube-runner" \
  --labels "minikube,self-hosted,arm64"

# 5. Install & jalankan
./svc.sh install
./svc.sh start
```

Setelah runner terinstall, update workflow untuk pakai runner ini:

```yaml
deploy-minikube:
  runs-on: [self-hosted, minikube]
```

#### Opsi B: SSH ke Minikube dari GitHub Actions

Jika Minikube di Mac lokal, kamu bisa SSH dari GitHub Actions:

```bash
# 1. Generate SSH key
ssh-keygen -t ed25519 -C "github-actions"

# 2. Copy public key ke authorized_keys
cat ~/.ssh/id_ed25519.pub >> ~/.ssh/authorized_keys

# 3. Set GitHub Secrets:
#    MINIKUBE_HOST    = IP address Mac kamu (atau hostname)
#    MINIKUBE_USER    = username Mac kamu
#    MINIKUBE_SSH_KEY = isi private key (id_ed25519)
#    MINIKUBE_WORKDIR = /Users/a2292/task-management-api
```

### Step 4: Set GitHub Secrets

Buka **GitHub Repository** → **Settings** → **Secrets and variables** → **Actions**

| Secret Name | Value | Keterangan |
|-------------|-------|------------|
| `MINIKUBE_HOST` | IP Mac kamu | Contoh: `192.168.1.100` |
| `MINIKUBE_USER` | Username Mac | Contoh: `a2292` |
| `MINIKUBE_SSH_KEY` | SSH private key | Isi lengkap file `id_ed25519` |
| `MINIKUBE_WORKDIR` | Path project | `/Users/a2292/task-management-api` |

### Step 5: Update Image Name

Edit `k8s/06-api-deployment.yaml`:

```yaml
# Ganti YOUR_USERNAME dengan username GitHub kamu
image: ghcr.io/YOUR_USERNAME/task-management-api:latest
```

---

## 🔄 Cara Kerja Pipeline

### Trigger Pipeline

| Event | Trigger | Action |
|-------|---------|--------|
| **Push ke main** | `git push origin main` | Test → Build → Push → Deploy |
| **Pull Request** | Buka PR ke main | Test → Build (tidak push) |
| **Tag Release** | `git tag v1.0.0` | Test → Build → Push (versioned) |

### Flow Pipeline

```
Push to main
    │
    ▼
┌─────────────┐
│  🧪 Test     │  → go test, go vet
└──────┬──────┘
       │ PASS
       ▼
┌─────────────┐
│  🐳 Build    │  → Build binary + Docker image
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  📤 Push     │  → Push ke ghcr.io
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  🚀 Deploy   │  → Pull dari ghcr.io → Deploy ke K8s
└─────────────┘
```

---

## 🧪 Testing the Pipeline

### 1. Push Code ke GitHub

```bash
git add .
git commit -m "feat: add CI/CD pipeline"
git push origin main
```

### 2. Cek GitHub Actions

1. Buka **GitHub Repository** → **Actions** tab
2. Lihat workflow sedang berjalan
3. Klik untuk detail setiap job

### 3. Verifikasi Deploy

```bash
# Cek pod di Minikube
kubectl get pods -n task-api

# Cek image di Minikube
minikube image ls | grep task-management

# Test API
curl http://$(minikube ip):30080/health
```

---

## 📝 Perintah berguna

```bash
# Lihat log GitHub Actions (jika pakai self-hosted runner)
cd ~/actions-runner
tail -f _diag/*.log

# Manual deploy dari local
./deploy.sh

# Manual reload (build + rolling update)
./reload.sh

# Lihat status deployment
kubectl rollout status -n task-api deployment/task-api

# Rollback ke versi sebelumnya
kubectl rollout undo -n task-api deployment/task-api
```

---

## 🔧 Troubleshooting

### GitHub Actions gagal di step Build

```
Error: open /Users/.../.docker/buildx/activity/.tmp-xxx: operation not permitted
```

**Solusi:** Build binary di local, lalu push. Atau gunakan self-hosted runner.

### Deploy ke Minikube gagal

```
Error: image pull back off
```

**Solusi:**
1. Pastikan image sudah di-push ke ghcr.io
2. Pastikan Minikube bisa akses internet
3. Cek imagePullPolicy di deployment

### Minikube tidak bisa pull dari ghcr.io

```bash
# Login ke ghcr.io di Minikube
minikube ssh
docker login ghcr.io -u YOUR_USERNAME -p YOUR_TOKEN
```

---

## 📊 Monitoring Pipeline

### GitHub Actions Dashboard

Buka: `https://github.com/YOUR_USERNAME/task-management-api/actions`

### Cek Status Deploy

```bash
# Quick status
kubectl get all -n task-api

# Watch pods real-time
kubectl get pods -n task-api -w

# Lihat events
kubectl get events -n task-api --sort-by='.lastTimestamp'
```

---

## 🎯 Best Practices

1. **Gunakan branch protection** — branch `main` harus diverifikasi sebelum merge
2. **Tag releases** — gunakan semantic versioning (`v1.0.0`, `v1.1.0`)
3. **Environment secrets** — gunakan GitHub Environments untuk staging/production
4. **Rollback strategy** — selalu siap rollback jika deploy gagal
5. **Monitor logs** — cek logs setelah deploy untuk pastikan tidak ada error
