# Bab 11: Docker & Containerization — Dari Nol Sampai Production

## Kenapa Bab Ini Penting?

Ini dia. **Bab pertama DevOps**. Setelah 10 bab ngomongin backend development — Go, API, database, caching, testing — sekarang kita masuk ke dunia yang bikin kamu jadi **full-spectrum engineer**. Dan semuanya dimulai dari Docker.

Docker itu **fondasi DevOps modern**. Kubernetes? Butuh Docker image. CI/CD pipeline? Build Docker image. Cloud deployment? Deploy Docker container. Microservices? Setiap service = Docker container. Kalau kamu belum paham Docker, semua bab setelah ini ga bakal masuk.

Di bab ini, kita bakal containerize project Task Manager API kita dari nol. Bukan cuma teori — kamu bakal bikin Dockerfile yang beneran production-ready, docker-compose.yml yang langsung bisa dipake development, dan paham gimana semua ini nyambung.

> **Analogi gampang:** Bayangin kamu pindahan rumah. Tanpa Docker = kamu angkut barang satu-satu, setiap barang beda bentuk, beda cara angkutnya. Dengan Docker = semua barang dikemas dalam **container** standar yang ukurannya sama. Mau kirim ke mana aja — truk, kapal, kereta — tinggal angkut container-nya. Ga peduli isinya apa.

---

## 11.1 Apa Itu Docker?

### The Problem — "It Works on My Machine"

Kamu pasti pernah ngalamin ini:

```
Developer A: "API-nya jalan di laptop gue"
Developer B: "Di laptop gue error, Node.js version-nya beda"
DevOps:      "Di server production beda lagi, dependency-nya ga lengkap"
```

Ini masalah klasik yang Docker selesaikan. Docker **package** aplikasi kamu beserta **semua dependency-nya** ke dalam satu unit yang bisa jalan di mana aja — laptop, server staging, server production, cloud — **hasilnya sama persis**.

### Container vs Virtual Machine (VM)

Sebelum Docker, solusinya pakai **Virtual Machine** (VM). Tapi VM itu berat — setiap VM jalan **OS penuh**.

```
┌──────────────────────────────────────────────────┐
│                 Virtual Machine                   │
│                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │   App A      │  │   App B      │  │  App C    │ │
│  │   Libs/Deps  │  │   Libs/Deps  │  │  Libs/Deps│ │
│  │   Guest OS   │  │   Guest OS   │  │  Guest OS │ │
│  │   (Ubuntu)   │  │   (CentOS)   │  │  (Debian) │ │
│  │   ~2GB       │  │   ~2GB       │  │  ~2GB     │ │
│  └─────────────┘  └─────────────┘  └───────────┘ │
│                  Hypervisor (VMware/VirtualBox)    │
│                      Host OS                      │
│                     Hardware                      │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│                   Container                       │
│                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │   App A      │  │   App B      │  │  App C    │ │
│  │   Libs/Deps  │  │   Libs/Deps  │  │  Libs/Deps│ │
│  │   ~50MB      │  │   ~100MB     │  │  ~15MB    │ │
│  └─────────────┘  └─────────────┘  └───────────┘ │
│                   Docker Engine                   │
│                      Host OS                      │
│                     Hardware                      │
└──────────────────────────────────────────────────┘
```

### Perbandingan VM vs Container

| Aspek | Virtual Machine | Container (Docker) |
|-------|----------------|-------------------|
| **Startup time** | Menit | **Detik** |
| **Size** | GB (include full OS) | **MB** (share OS kernel) |
| **Isolation** | Full (hardware level) | Process level |
| **Performance** | Overhead besar | **Near-native** |
| **Resource usage** | Berat (RAM, CPU) | **Ringan** |
| **Portability** | Bisa tapi besar | **Sangat portable** |
| **OS** | Full OS per VM | Share host OS kernel |

### Analogi Shipping Container

Kenapa namanya "Docker" dan "Container"? Karena konsepnya **persis** kayak container pengiriman barang:

```
Sebelum container pengiriman (1950an):
- Setiap barang beda bentuk, beda cara angkut
- Loading/unloading manual, satu-satu
- Sering rusak, sering hilang
- Butuh banyak tenaga kerja

Setelah container pengiriman (1960an+):
- Semua barang dimasukkin ke box standar (20ft/40ft)
- Kapal, truk, kereta — semua bisa angkut container yang sama
- Loading/unloading pakai crane, cepet
- Ga peduli isinya apa — containernya standar

Docker = shipping container untuk software:
- App kamu + semua dependency = 1 container
- Laptop, server, cloud — semua bisa jalanin container yang sama
- Ga peduli OS-nya apa — containernya standar
```

### Kenapa Docker Jadi Standard DevOps?

1. **Consistency** — "Works on my machine" is dead. Container jalan sama di mana-mana
2. **Isolation** — Setiap app punya environment sendiri. App A pakai Go 1.25, App B pakai Go 1.22? No problem
3. **Portability** — Build sekali, jalanin di mana aja
4. **Scalability** — Mau scale? Tinggal jalanin lebih banyak container
5. **Infrastructure as Code** — Dockerfile = recipe yang bisa di-version control
6. **Ecosystem** — Docker Hub punya ribuan image siap pakai (PostgreSQL, Redis, Nginx, dll)

### Fun Fact

Docker, Kubernetes, Terraform, Prometheus — semua **ditulis pakai Go**. Jadi kamu belajar Go + Docker = belajar pakai tools yang ditulis dalam bahasa yang sama. Meta banget.

---

## 11.2 Konsep Dasar Docker

Sebelum mulai ngoding, kamu harus paham 6 konsep ini:

### Image, Container, Registry — Trinity Docker

```
┌─────────────┐     docker build      ┌─────────────┐
│  Dockerfile  │ ──────────────────→  │    Image     │
│  (recipe)    │                       │  (blueprint) │
└─────────────┘                       └──────┬──────┘
                                              │
                                    docker run │    docker push
                                              │         │
                                    ┌─────────▼───┐    ▼
                                    │  Container   │  ┌─────────────┐
                                    │  (instance)  │  │  Registry   │
                                    │  (running!)  │  │ (Docker Hub)│
                                    └─────────────┘  └─────────────┘
```

### 1. Dockerfile — The Recipe

Dockerfile itu **resep** untuk bikin image. Kayak resep masakan — ditulis step by step.

```
Dockerfile = resep nasi goreng
- Mulai dari nasi putih (FROM)
- Tambah bumbu (COPY)
- Masak (RUN)
- Sajikan (CMD)
```

### 2. Image — The Blueprint

Image itu **blueprint** atau **template** yang ga bisa diubah (immutable). Dibuat dari Dockerfile.

```
Image = cetakan kue
- Kamu bisa bikin banyak kue dari 1 cetakan
- Cetakannya sendiri ga berubah
- Setiap kue yang dibikin = container baru
```

### 3. Container — The Running Instance

Container itu **instance** dari image yang **sedang berjalan**. Ini proses yang hidup.

```
Image → Container
Cetakan kue → Kue yang udah jadi
Class → Object (kalau pakai analogi OOP)
```

Satu image bisa jadi banyak container:

```bash
# Dari 1 image, bikin 3 container
docker run myapp    # container 1
docker run myapp    # container 2
docker run myapp    # container 3
```

### 4. Registry — The Store

Registry itu tempat simpan dan distribusi image. **Docker Hub** = npm registry-nya Docker.

```
Docker Hub = npm registry
Image = npm package
docker pull = npm install
docker push = npm publish
```

Registry populer:
- **Docker Hub** — default, public. `docker pull postgres:16`
- **GitHub Container Registry** (ghcr.io) — buat project GitHub
- **AWS ECR** — Amazon's private registry
- **Google GCR** — Google's private registry

### 5. Volume — Persistent Data

Container itu **ephemeral** (sementara). Kalau container dihapus, data di dalamnya **hilang**. Volume itu solusinya — data disimpan di luar container.

```
Tanpa volume:
Container mati → Data ilang → PostgreSQL data hilang semua

Dengan volume:
Container mati → Data di volume tetap ada → Start container baru, data masih utuh
```

### 6. Network — Container Communication

Docker network memungkinkan container-container **ngobrol satu sama lain**. Go app perlu connect ke PostgreSQL? Mereka harus di network yang sama.

```
┌─────── Docker Network ───────┐
│                               │
│  ┌─────────┐   ┌───────────┐ │
│  │  Go App  │──→│ PostgreSQL │ │
│  │ :8080    │   │  :5432     │ │
│  └─────────┘   └───────────┘ │
│       │                       │
│       │         ┌───────────┐ │
│       └────────→│   Redis    │ │
│                 │  :6379     │ │
│                 └───────────┘ │
└───────────────────────────────┘
```

### Rangkuman Konsep

| Konsep | Analogi | Fungsi |
|--------|---------|--------|
| **Dockerfile** | Resep masakan | Instruksi untuk bikin image |
| **Image** | Cetakan kue | Template immutable |
| **Container** | Kue yang jadi | Instance yang running |
| **Registry** | Toko kue online | Tempat simpan/share image |
| **Volume** | Wadah penyimpanan | Data persistent |
| **Network** | Jalan penghubung | Komunikasi antar container |

---

## 11.3 Install Docker

### Docker Desktop (macOS)

Docker Desktop itu all-in-one package yang include Docker Engine, Docker CLI, Docker Compose, dan GUI.

**Cara install:**

```bash
# Cara 1: Download langsung
# Buka https://www.docker.com/products/docker-desktop/
# Download Docker Desktop for Mac (Apple Silicon / Intel)
# Double click .dmg → drag ke Applications

# Cara 2: Homebrew
brew install --cask docker
```

Setelah install, **buka Docker Desktop** dari Applications. Tunggu sampai statusnya **"Docker Desktop is running"** (icon Docker di menu bar warnanya stabil, bukan animasi).

### Verify Installation

```bash
# Check Docker version
docker version

# Output:
# Client:
#  Version:           27.x.x
#  API version:       1.47
#  Go version:        go1.22.x
#  OS/Arch:           darwin/arm64
#
# Server:
#  Version:           27.x.x
#  API version:       1.47
#  Go version:        go1.22.x
#  OS/Arch:           linux/arm64

# Check Docker info (lebih detail)
docker info

# Output (ringkasan):
# Containers: 0
# Images: 0
# Server Version: 27.x.x
# Storage Driver: overlay2
# ...

# Check Docker Compose
docker compose version
# Docker Compose version v2.x.x

# Quick test — pull dan run image hello-world
docker run hello-world
# Output: "Hello from Docker!"
# Kalau muncul ini, Docker udah siap!
```

### Troubleshooting Install

| Problem | Solution |
|---------|----------|
| `docker: command not found` | Buka Docker Desktop dulu, atau restart terminal |
| `Cannot connect to Docker daemon` | Docker Desktop belum running — buka dari Applications |
| `permission denied` | Jangan pakai `sudo` di macOS. Pastikan user kamu di Docker group |
| Slow performance | Docker Desktop → Settings → Resources → allocate lebih banyak RAM/CPU |

---

## 11.4 Dockerfile Basics — Instruksi Dasar

Dockerfile itu file teks biasa (tanpa extension) yang berisi instruksi step-by-step untuk build image. Setiap instruksi = 1 **layer** di image.

### Instruksi-Instruksi Penting

#### FROM — Base Image

```dockerfile
# Mulai dari image apa?
FROM golang:1.25-alpine

# Ini kayak bilang: "Mulai dari environment yang udah ada Go 1.25"
# Alpine = Linux distro yang super kecil (~5MB)
```

`FROM` itu **selalu instruksi pertama**. Kamu ga bisa bikin image dari nol (kecuali `FROM scratch`). Setiap image dibangun di atas image lain.

```
FROM node:20         → Udah ada Node.js 20
FROM golang:1.25     → Udah ada Go 1.25
FROM python:3.12     → Udah ada Python 3.12
FROM ubuntu:22.04    → Linux Ubuntu kosong
FROM alpine:3.20     → Linux Alpine (super kecil)
FROM scratch         → Kosong total (literally nothing)
```

#### WORKDIR — Set Working Directory

```dockerfile
WORKDIR /app

# Semua instruksi setelah ini dijalankan di /app
# Kalau folder /app belum ada, otomatis dibikin
# Sama kayak: mkdir -p /app && cd /app
```

#### COPY — Copy File dari Host ke Image

```dockerfile
# Copy file spesifik
COPY go.mod go.sum ./

# Copy semua file
COPY . .

# Format: COPY <source di laptop> <destination di image>
```

#### RUN — Jalankan Command saat Build

```dockerfile
# Install dependency
RUN go mod download

# Install system package
RUN apk add --no-cache git

# Compile binary
RUN go build -o main ./cmd/api/
```

`RUN` dijalankan **saat build image** (docker build), bukan saat container jalan.

#### EXPOSE — Dokumentasi Port

```dockerfile
EXPOSE 8080

# PENTING: EXPOSE itu cuma DOKUMENTASI!
# Dia ga beneran publish port.
# Kamu tetap perlu -p flag waktu docker run.
# Tapi ini penting buat dokumentasi dan tooling.
```

#### CMD — Default Command saat Container Start

```dockerfile
CMD ["./main"]

# Ini command yang jalan waktu container start (docker run)
# Bisa di-override waktu docker run:
# docker run myapp ./other-command
```

#### ENTRYPOINT — Fixed Command

```dockerfile
ENTRYPOINT ["./main"]

# Mirip CMD, tapi ga bisa di-override (secara default)
# CMD jadi "arguments" ke ENTRYPOINT
```

**CMD vs ENTRYPOINT:**

```dockerfile
# Pakai CMD — bisa di-override
CMD ["./main"]
# docker run myapp           → jalanin ./main
# docker run myapp /bin/sh   → jalanin /bin/sh (override!)

# Pakai ENTRYPOINT — fixed
ENTRYPOINT ["./main"]
# docker run myapp           → jalanin ./main
# docker run myapp --port=9090  → jalanin ./main --port=9090

# Kombinasi — ENTRYPOINT = command, CMD = default args
ENTRYPOINT ["./main"]
CMD ["--port=8080"]
# docker run myapp              → ./main --port=8080
# docker run myapp --port=9090  → ./main --port=9090
```

**Rule of thumb:** Untuk API server, pakai `CMD` aja. Lebih fleksibel.

#### ENV — Set Environment Variable

```dockerfile
ENV APP_ENV=production
ENV PORT=8080
```

#### ARG — Build-time Variable

```dockerfile
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine

# ARG = variabel yang cuma ada waktu build
# ENV = variabel yang ada di container saat runtime
```

### Rangkuman Instruksi

| Instruksi | Fungsi | Kapan Jalan |
|-----------|--------|-------------|
| `FROM` | Base image | - |
| `WORKDIR` | Set working directory | Build time |
| `COPY` | Copy file host → image | Build time |
| `RUN` | Jalankan command | Build time |
| `ENV` | Set environment variable | Build & Runtime |
| `ARG` | Build-time variable | Build time only |
| `EXPOSE` | Dokumentasi port | - (docs only) |
| `CMD` | Default command saat start | Runtime |
| `ENTRYPOINT` | Fixed command saat start | Runtime |

---

## 11.5 Dockerfile untuk Go — Single Stage

### Kenapa Go Cocok Banget Buat Docker

Go punya keunggulan **unik** yang bikin dia jadi bahasa paling Docker-friendly:

```
Go:
- Compile jadi SINGLE BINARY
- Ga butuh runtime (no JVM, no Node.js, no Python interpreter)
- Binary bisa jalan di container KOSONG (scratch)
- Image size: 10-20MB

Node.js:
- Butuh Node.js runtime (~150MB)
- Butuh node_modules (~200MB-1GB)
- Butuh package.json, package-lock.json
- Image size: 200MB-1GB

Java:
- Butuh JVM (~200MB)
- Butuh .jar + dependencies
- Image size: 200MB-500MB

Python:
- Butuh Python interpreter (~100MB)
- Butuh pip packages
- Image size: 200MB-800MB
```

**Kesimpulan:** Go binary itu self-contained. Ga butuh apa-apa. Tinggal copy binary, jalanin. Done.

### Single Stage Dockerfile (Simple)

Mari kita mulai dari yang paling sederhana dulu:

```dockerfile
# Dockerfile (single stage — simple, cocok buat belajar)

# 1. Mulai dari image Go
FROM golang:1.25-alpine

# 2. Set working directory
WORKDIR /app

# 3. Copy dependency files dulu (supaya di-cache)
COPY go.mod go.sum ./

# 4. Download dependencies
RUN go mod download

# 5. Copy semua source code
COPY . .

# 6. Build binary
RUN go build -o main ./cmd/api/

# 7. Expose port
EXPOSE 8080

# 8. Jalanin binary
CMD ["./main"]
```

**Test build:**

```bash
# Build image
docker build -t task-manager:simple .

# Check size
docker images task-manager:simple
# REPOSITORY     TAG      SIZE
# task-manager   simple   ~500MB  ← GEDE BANGET!
```

500MB?! Itu karena image-nya include **seluruh Go toolchain** (compiler, tools, dll). Kamu cuma butuh binary-nya aja, ga butuh compiler di production. Ini yang bakal kita solve pakai multi-stage build.

### Perbandingan: Node.js Dockerfile

Buat context, ini Dockerfile Node.js (NestJS):

```dockerfile
# Node.js Dockerfile — perlu lebih banyak step
FROM node:20-alpine

WORKDIR /app

# Copy dependency files
COPY package.json package-lock.json ./

# Install dependencies (ini bisa ~200MB+)
RUN npm ci --only=production

# Copy source code
COPY . .

# Build (kalau TypeScript)
RUN npm run build

EXPOSE 3000

# Butuh Node.js runtime untuk jalanin app
CMD ["node", "dist/main.js"]
```

Perhatikan: Node.js **butuh runtime** (node) untuk jalanin app. Go? Ga butuh apa-apa.

---

## 11.6 Multi-Stage Build — Production-Ready

### Konsep Multi-Stage

Multi-stage build itu **2+ tahap build** dalam satu Dockerfile:
- **Stage 1 (builder):** Install dependencies, compile code → hasilnya binary
- **Stage 2 (final):** Copy binary aja ke image minimal → image kecil banget

```
Stage 1: Build                    Stage 2: Production
┌─────────────────────┐          ┌──────────────────┐
│ golang:1.25-alpine  │          │ alpine:3.20       │
│ (~300MB)            │          │ (~5MB)            │
│                     │          │                   │
│ go.mod, go.sum      │          │                   │
│ Source code          │  COPY   │                   │
│ Go compiler          │ ─────→ │ main (binary)     │
│ Dependencies         │         │ (~10MB)           │
│ Binary (main)       │          │                   │
│                     │          │ Total: ~15MB      │
│ Total: ~500MB       │          │                   │
└─────────────────────┘          └──────────────────┘
     DIBUANG                          INI YANG DIPAKE
```

### Multi-Stage Dockerfile — Task Manager API

```dockerfile
# ============================================
# Dockerfile — Task Manager API (Production)
# Multi-stage build: 500MB → ~20MB
# ============================================

# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

# Install git (diperlukan beberapa Go module)
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files dulu (leverage Docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
# CGO_ENABLED=0 → static binary (ga depend ke C library)
# -ldflags="-s -w" → strip debug info (binary lebih kecil)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api/

# ---- Stage 2: Production ----
FROM alpine:3.20

# Install ca-certificates (untuk HTTPS calls) dan tzdata (timezone)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy HANYA binary dari stage builder
COPY --from=builder /app/main .

# Non-root user (security best practice)
RUN adduser -D -g '' appuser
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# Start app
CMD ["./main"]
```

### Penjelasan Setiap Bagian

**`CGO_ENABLED=0`** — Bikin binary yang pure Go, ga depend ke C library. Penting supaya binary bisa jalan di Alpine/scratch yang ga punya glibc.

**`-ldflags="-s -w"`** — Strip symbol table (`-s`) dan debug info (`-w`). Binary jadi lebih kecil, ga ada efek ke performance.

**`FROM alpine:3.20`** (stage 2) — Image Linux yang super kecil (~5MB). Cukup buat jalanin binary Go.

**`ca-certificates`** — Diperlukan kalau app kamu bikin HTTPS request (misal ke external API).

**`tzdata`** — Timezone data. Diperlukan kalau app kamu pake `time.LoadLocation()`.

**`adduser -D -g '' appuser`** — Bikin non-root user. **Jangan jalanin app sebagai root** di production.

**`HEALTHCHECK`** — Docker bisa tau apakah app kamu healthy atau engga. Endpoint `/health` kita udah ada dari project.

### Perbandingan Size

```bash
# Build kedua versi
docker build -t task-manager:simple -f Dockerfile.simple .
docker build -t task-manager:prod .

# Bandingkan size
docker images | grep task-manager
# REPOSITORY     TAG      SIZE
# task-manager   simple   ~500MB   ← Single stage
# task-manager   prod     ~20MB    ← Multi-stage (25x lebih kecil!)
```

| Versi | Size | Isi |
|-------|------|-----|
| Single stage | ~500MB | Go compiler + tools + source code + binary |
| Multi-stage (alpine) | ~20MB | Alpine Linux + binary |
| Multi-stage (scratch) | ~10MB | Binary only (ga ada shell, ga ada apa-apa) |

### Variant: scratch (Ultra Minimal)

Kalau kamu mau image sekecil-kecilnya:

```dockerfile
# Stage 2 pakai scratch — literally kosong
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/main /main

ENTRYPOINT ["/main"]
```

**Kelebihan scratch:** Ultra kecil (~10MB), ultra secure (ga ada shell, ga ada tool yang bisa dieksploit).

**Kekurangan scratch:** Ga bisa `docker exec -it bash` buat debugging. Ga ada shell sama sekali. Cocok buat production yang udah mature. Buat sekarang, kita pakai **alpine** aja.

---

## 11.7 .dockerignore — Jaga Image Tetap Bersih

`.dockerignore` itu kayak `.gitignore` tapi untuk Docker. File yang di-list ga akan di-COPY ke image.

### Buat File .dockerignore

```
# .dockerignore

# Version control
.git
.gitignore

# IDE
.vscode
.idea
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Docker files (ga perlu di-copy ke dalam image)
Dockerfile
Dockerfile.*
docker-compose*.yml
.dockerignore

# Documentation
*.md
docs/
ebook/
LICENSE
README.md

# Test files (ga perlu di production)
*_test.go
testdata/

# Environment files (JANGAN masuk ke image — security risk!)
.env
.env.*

# Binary yang udah ada (kita build ulang di dalam Docker)
main
tmp/

# Air config (development tool)
.air.toml
```

### Kenapa .dockerignore Penting?

1. **Kecepatan build** — Kurangi file yang di-copy = build lebih cepat
2. **Image size** — Ga perlu file yang ga dibutuhin
3. **Security** — `.env`, credentials, dll ga boleh masuk ke image
4. **Cache** — File yang ga relevan berubah ga bikin cache invalidation

**Tanpa .dockerignore:**

```
COPY . .
# Copy SEMUA file: .git (bisa ratusan MB), node_modules, .env, docs, dll
# Build context: 500MB+
# Build time: lama
```

**Dengan .dockerignore:**

```
COPY . .
# Cuma copy source code yang dibutuhkan
# Build context: 5MB
# Build time: cepet
```

---

## 11.8 Docker Commands — Cheat Sheet

### Build Image

```bash
# Build image dari Dockerfile di current directory
docker build -t task-manager:latest .

# Build dengan tag spesifik
docker build -t task-manager:v1.0.0 .

# Build dari Dockerfile yang beda nama
docker build -t task-manager:dev -f Dockerfile.dev .

# Build tanpa cache (force rebuild semua layer)
docker build --no-cache -t task-manager:latest .

# Build dengan build argument
docker build --build-arg APP_ENV=production -t task-manager:prod .
```

### Run Container

```bash
# Run container (foreground — output di terminal)
docker run task-manager:latest

# Run di background (detached mode)
docker run -d task-manager:latest

# Run dengan port mapping
# Format: -p <host_port>:<container_port>
docker run -d -p 8080:8080 task-manager:latest

# Run dengan nama container
docker run -d --name my-api -p 8080:8080 task-manager:latest

# Run dengan environment variables
docker run -d -p 8080:8080 \
    -e APP_ENV=production \
    -e DB_HOST=postgres \
    -e DB_PASSWORD=secret \
    task-manager:latest

# Run dengan env file
docker run -d -p 8080:8080 --env-file .env task-manager:latest

# Run dengan volume mount
docker run -d -p 8080:8080 -v $(pwd):/app task-manager:latest

# Run interaktif (masuk ke shell)
docker run -it alpine:3.20 /bin/sh
```

### Container Management

```bash
# List running containers
docker ps

# List ALL containers (termasuk yang stopped)
docker ps -a

# Stop container
docker stop my-api

# Start container yang udah stop
docker start my-api

# Restart container
docker restart my-api

# Remove container (harus stop dulu)
docker rm my-api

# Force remove (langsung stop + remove)
docker rm -f my-api

# Remove semua stopped containers
docker container prune
```

### Logs & Debugging

```bash
# Lihat logs container
docker logs my-api

# Follow logs (real-time, kayak tail -f)
docker logs -f my-api

# Lihat 100 baris terakhir
docker logs --tail 100 my-api

# Logs dengan timestamp
docker logs -t my-api

# Exec — masuk ke container yang running
docker exec -it my-api /bin/sh

# Exec — jalankan command di container
docker exec my-api ls -la /app

# Inspect — lihat detail container (IP, env, mount, dll)
docker inspect my-api

# Stats — monitor resource usage real-time
docker stats
```

### Image Management

```bash
# List semua images
docker images

# Remove image
docker rmi task-manager:latest

# Remove unused images
docker image prune

# Remove SEMUA unused data (images, containers, volumes, networks)
docker system prune -a

# Pull image dari registry
docker pull postgres:16-alpine

# Push image ke registry
docker push myrepo/task-manager:latest

# Tag image
docker tag task-manager:latest myrepo/task-manager:v1.0.0
```

### Quick Reference Table

| Command | Fungsi |
|---------|--------|
| `docker build -t name .` | Build image |
| `docker run -d -p 8080:8080 name` | Run container |
| `docker ps` | List running containers |
| `docker ps -a` | List all containers |
| `docker logs -f name` | Follow container logs |
| `docker exec -it name /bin/sh` | Shell ke container |
| `docker stop name` | Stop container |
| `docker rm name` | Remove container |
| `docker images` | List images |
| `docker rmi name` | Remove image |
| `docker system prune -a` | Cleanup everything |

---

## 11.9 Environment Variables di Docker

### 3 Cara Pass Environment Variables

#### Cara 1: -e Flag (Inline)

```bash
docker run -d -p 8080:8080 \
    -e APP_ENV=production \
    -e PORT=8080 \
    -e DB_HOST=postgres \
    -e DB_PORT=5432 \
    -e DB_USER=postgres \
    -e DB_PASSWORD=supersecret \
    -e DB_NAME=task_manager \
    -e REDIS_HOST=redis \
    -e REDIS_PORT=6379 \
    -e JWT_SECRET=my-jwt-secret-key \
    task-manager:latest
```

**Kapan pakai:** Quick test, sedikit variable. Ga cocok kalau banyak variable — terlalu panjang.

#### Cara 2: --env-file (Recommended)

Bikin file `.env.docker`:

```env
# .env.docker — environment variables untuk Docker
APP_ENV=production
PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=supersecret
DB_NAME=task_manager

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=my-jwt-secret-key-change-this
```

```bash
docker run -d -p 8080:8080 --env-file .env.docker task-manager:latest
```

**Kapan pakai:** Banyak variable. File ini ga boleh masuk ke git (tambahkan ke .gitignore).

#### Cara 3: ENV di Dockerfile (Default Values)

```dockerfile
# Di Dockerfile — set default values
ENV APP_ENV=production
ENV PORT=8080
```

**Kapan pakai:** Default values yang aman. Bisa di-override waktu `docker run -e`.

### Prioritas (Mana yang Menang?)

```
docker run -e APP_ENV=staging  →  MENANG (paling tinggi)
--env-file .env.docker         →  Override ENV di Dockerfile
ENV di Dockerfile               →  Default (paling rendah)
```

### Perbandingan dengan NestJS

```
NestJS:                          Docker/Go:
.env file                       .env.docker file
process.env.DB_HOST             os.Getenv("DB_HOST")
@nestjs/config ConfigService    config.Load() (manual)
```

Ingat config kita di `internal/config/config.go`:

```go
func Load() *Config {
    return &Config{
        Env:  getEnv("APP_ENV", "development"),
        Port: getEnv("PORT", "8080"),
        DB: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            Name:     getEnv("DB_NAME", "task_manager"),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
        },
        JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
    }
}
```

Perhatikan: semua config **udah baca dari environment variable** pakai `os.Getenv`. Ini artinya app kita **Docker-ready** tanpa perlu ubah apa-apa. Tinggal pass env variable lewat Docker.

---

## 11.10 Docker Compose — Multi-Container Orchestration

### Kenapa Perlu Docker Compose?

Sampai sekarang, kita cuma jalanin **1 container**. Tapi project kita butuh:
- **Go API** — port 8080
- **PostgreSQL** — port 5432
- **Redis** — port 6379

Tanpa Docker Compose:

```bash
# Bikin network dulu
docker network create task-net

# Run PostgreSQL
docker run -d --name postgres \
    --network task-net \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=secret \
    -e POSTGRES_DB=task_manager \
    -v pgdata:/var/lib/postgresql/data \
    -p 5432:5432 \
    postgres:16-alpine

# Run Redis
docker run -d --name redis \
    --network task-net \
    -p 6379:6379 \
    redis:7-alpine

# Run Go API
docker run -d --name api \
    --network task-net \
    -e DB_HOST=postgres \
    -e REDIS_HOST=redis \
    -p 8080:8080 \
    task-manager:latest

# Itu 3 command panjang + harus urut (DB dulu, baru API)
```

Dengan Docker Compose:

```bash
docker compose up -d
# DONE. 1 command. Semua jalan.
```

### Docker Compose Basics

Docker Compose pakai file `docker-compose.yml` (atau `compose.yml`) yang define semua services, networks, volumes dalam **satu file**.

#### Struktur Dasar docker-compose.yml

```yaml
# docker-compose.yml

services:
  # Service 1: nama_service
  nama_service:
    image: image_name:tag      # pakai image dari registry
    # ATAU
    build: ./path              # build dari Dockerfile
    ports:
      - "host:container"       # port mapping
    environment:
      - KEY=VALUE              # env variables
    volumes:
      - name:/path             # volume mount
    depends_on:
      - other_service          # dependency
    networks:
      - network_name           # network

volumes:
  nama_volume:                 # define named volume

networks:
  nama_network:                # define network
    driver: bridge
```

#### Keyword Penting

| Keyword | Fungsi | Equivalent docker run |
|---------|--------|-----------------------|
| `image` | Pakai image dari registry | `docker run image:tag` |
| `build` | Build dari Dockerfile | `docker build` + `docker run` |
| `ports` | Port mapping | `-p host:container` |
| `environment` | Env variables | `-e KEY=VALUE` |
| `env_file` | Load dari file | `--env-file` |
| `volumes` | Mount volume | `-v name:/path` |
| `depends_on` | Startup order | Manual ordering |
| `networks` | Network membership | `--network name` |
| `restart` | Restart policy | `--restart always` |
| `healthcheck` | Health check | `HEALTHCHECK` di Dockerfile |
| `command` | Override CMD | Command di akhir `docker run` |

---

## 11.11 Docker Compose untuk Project — Full Setup

Ini dia bagian yang paling hands-on. Kita bikin docker-compose.yml yang lengkap untuk project Task Manager API.

### docker-compose.yml — Production

```yaml
# docker-compose.yml — Production-ready setup
# Usage: docker compose up -d

services:
  # ============================================
  # Go API — Task Manager
  # ============================================
  api:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: task-manager-api
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-supersecret}
      - DB_NAME=task_manager
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=
      - JWT_SECRET=${JWT_SECRET:-change-this-in-production}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

  # ============================================
  # PostgreSQL Database
  # ============================================
  postgres:
    image: postgres:16-alpine
    container_name: task-manager-db
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD:-supersecret}
      - POSTGRES_DB=task_manager
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  # ============================================
  # Redis Cache
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: task-manager-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - task-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    command: redis-server --appendonly yes

# ============================================
# Volumes — Persistent Data
# ============================================
volumes:
  postgres_data:
    name: task-manager-postgres-data
  redis_data:
    name: task-manager-redis-data

# ============================================
# Networks
# ============================================
networks:
  task-network:
    name: task-manager-network
    driver: bridge
```

### Penjelasan Detail

#### Service: api

```yaml
api:
    build:
      context: .           # Build context = current directory
      dockerfile: Dockerfile  # Pakai Dockerfile yang udah kita bikin

    depends_on:
      postgres:
        condition: service_healthy  # Tunggu PostgreSQL HEALTHY dulu
      redis:
        condition: service_healthy  # Tunggu Redis HEALTHY dulu
    # PENTING: depends_on tanpa condition cuma nunggu container START,
    # bukan nunggu service READY. Pakai condition: service_healthy!
```

**Kenapa `service_healthy` penting?**

```
Tanpa health check:
1. PostgreSQL container starts    ← belum ready terima koneksi
2. API container starts           ← langsung connect ke PostgreSQL
3. Connection refused!            ← PostgreSQL belum siap
4. API crash!

Dengan health check:
1. PostgreSQL container starts    ← belum ready
2. Health check: pg_isready...    ← belum ready
3. Health check: pg_isready...    ← READY!
4. API container starts           ← PostgreSQL udah siap
5. Connection success!
```

#### Service: postgres

```yaml
postgres:
    image: postgres:16-alpine    # Pakai image resmi dari Docker Hub
    volumes:
      - postgres_data:/var/lib/postgresql/data
      # /var/lib/postgresql/data = tempat PostgreSQL simpan data
      # postgres_data = named volume, data persist meski container dihapus
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      # pg_isready = tool bawaan PostgreSQL untuk check readiness
```

#### Service: redis

```yaml
redis:
    command: redis-server --appendonly yes
    # --appendonly yes = Redis persist data ke disk (AOF mode)
    # Kalau tanpa ini, data Redis hilang waktu container restart
```

#### Environment Variables dengan Default

```yaml
environment:
    - DB_PASSWORD=${DB_PASSWORD:-supersecret}
    # ${DB_PASSWORD:-supersecret} artinya:
    # - Kalau env var DB_PASSWORD ada di host → pakai itu
    # - Kalau ga ada → pakai "supersecret" sebagai default
    # Jadi kamu bisa override dari .env file atau shell
```

### docker-compose.dev.yml — Development (Hot Reload)

Untuk development, kita mau **hot reload** — kode berubah, app otomatis restart. Kita pakai [Air](https://github.com/air-verse/air), Go's hot reload tool.

Pertama, bikin `Dockerfile.dev`:

```dockerfile
# Dockerfile.dev — Development with hot reload (Air)

FROM golang:1.25-alpine

# Install Air untuk hot reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Source code di-mount via volume (bukan COPY)
# Jadi setiap perubahan file otomatis ke-sync

EXPOSE 8080

# Jalanin Air (bukan binary langsung)
CMD ["air", "-c", ".air.toml"]
```

Bikin `.air.toml` (config Air):

```toml
# .air.toml — Air configuration for hot reload

root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api/"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "ebook", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "toml"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true
```

Sekarang `docker-compose.dev.yml`:

```yaml
# docker-compose.dev.yml — Development setup with hot reload
# Usage: docker compose -f docker-compose.dev.yml up

services:
  # ============================================
  # Go API — Development (Hot Reload)
  # ============================================
  api:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: task-manager-api-dev
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=devpassword
      - DB_NAME=task_manager_dev
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=
      - JWT_SECRET=dev-secret-key
    volumes:
      # Bind mount — source code di laptop sync ke container
      - .:/app
      # Anonymous volume — jangan sync tmp/ (Air build output)
      - /app/tmp
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - task-network-dev

  # ============================================
  # PostgreSQL — Development
  # ============================================
  postgres:
    image: postgres:16-alpine
    container_name: task-manager-db-dev
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=devpassword
      - POSTGRES_DB=task_manager_dev
    volumes:
      - postgres_data_dev:/var/lib/postgresql/data
    networks:
      - task-network-dev
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ============================================
  # Redis — Development
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: task-manager-redis-dev
    ports:
      - "6379:6379"
    networks:
      - task-network-dev
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data_dev:
    name: task-manager-postgres-data-dev

networks:
  task-network-dev:
    name: task-manager-network-dev
    driver: bridge
```

### Perbedaan Dev vs Prod

| Aspek | Development | Production |
|-------|------------|------------|
| **Dockerfile** | `Dockerfile.dev` | `Dockerfile` (multi-stage) |
| **Binary** | Air (hot reload) | Compiled binary |
| **Source code** | Bind mount (live sync) | COPY (baked into image) |
| **Image size** | ~500MB (ada Go tools) | ~20MB (binary only) |
| **Env** | `APP_ENV=development` | `APP_ENV=production` |
| **DB password** | Simple (`devpassword`) | Strong (from env/secret) |
| **Redis persistence** | Off (ga perlu) | On (`--appendonly yes`) |
| **Restart policy** | Ga ada | `unless-stopped` |

### Jalankan Development

```bash
# Start semua services (development)
docker compose -f docker-compose.dev.yml up

# Start di background
docker compose -f docker-compose.dev.yml up -d

# Sekarang edit source code kamu — Air otomatis rebuild + restart!
# Try: ubah handler/health.go, save, liat log container

# Stop semua
docker compose -f docker-compose.dev.yml down

# Stop + hapus volumes (reset database)
docker compose -f docker-compose.dev.yml down -v
```

### Jalankan Production

```bash
# Build dan start semua services (production)
docker compose up -d --build

# Check status
docker compose ps

# Check logs
docker compose logs -f api

# Stop
docker compose down
```

---

## 11.12 Docker Networking

### Tipe Network di Docker

```bash
# List networks
docker network ls
# NETWORK ID     NAME      DRIVER    SCOPE
# xxxxxxxxxxxx   bridge    bridge    local    ← default
# xxxxxxxxxxxx   host      host      local
# xxxxxxxxxxxx   none      null      local
```

| Driver | Penjelasan | Use Case |
|--------|-----------|----------|
| **bridge** | Default. Container bisa komunikasi via nama | Multi-container app (ini yang kita pakai) |
| **host** | Container share network host | Performance, tapi kurang isolasi |
| **none** | Ga ada network | Isolated container |

### Container Communication by Service Name

Di Docker Compose, container bisa **ngobrol pakai nama service** — ga perlu IP address.

```yaml
services:
  api:
    environment:
      - DB_HOST=postgres    # ← pakai nama service, bukan IP!
      - REDIS_HOST=redis    # ← pakai nama service, bukan IP!

  postgres:
    # Docker Compose otomatis bikin DNS entry: "postgres" → IP container ini

  redis:
    # Docker Compose otomatis bikin DNS entry: "redis" → IP container ini
```

**Kenapa bisa?** Docker Compose otomatis bikin **DNS resolution** di network-nya. Nama service = hostname.

```
api → connect ke "postgres:5432"
Docker DNS → resolve "postgres" → 172.18.0.2
api → connect ke 172.18.0.2:5432
```

Ini kenapa di config kita, `DB_HOST` di-set ke `postgres` (nama service), bukan `localhost` atau IP address.

### expose vs ports

```yaml
services:
  api:
    ports:
      - "8080:8080"    # Accessible dari HOST dan container lain

  postgres:
    ports:
      - "5432:5432"    # Accessible dari HOST dan container lain
    # Kalau production, mungkin mau:
    expose:
      - "5432"         # HANYA accessible dari container lain (dalam network)
                       # Ga bisa diakses dari host!
```

| Keyword | Host Access | Container Access | Use Case |
|---------|------------|-----------------|----------|
| `ports: "5432:5432"` | Ya | Ya | Development (bisa connect via pgAdmin di host) |
| `expose: "5432"` | Tidak | Ya | Production (hanya API yang bisa akses DB) |

**Best practice production:** Database dan Redis pakai `expose` aja, jangan `ports`. Ga perlu accessible dari luar.

### Visualisasi Network

```
Host Machine (macOS)
├── localhost:8080  ──→  ┌─── Docker Network (bridge) ───┐
│                        │                                │
│                        │  ┌──────────────────────────┐  │
│                        │  │   api (task-manager-api)  │  │
│                        │  │   Port: 8080              │  │
│                        │  │                           │  │
│                        │  │   DB_HOST=postgres ───────┤──┤──→ postgres
│                        │  │   REDIS_HOST=redis ───────┤──┤──→ redis
│                        │  └──────────────────────────┘  │
│                        │                                │
├── localhost:5432  ──→  │  ┌──────────────────────────┐  │
│   (dev only)           │  │   postgres                │  │
│                        │  │   Port: 5432              │  │
│                        │  │   Volume: postgres_data   │  │
│                        │  └──────────────────────────┘  │
│                        │                                │
├── localhost:6379  ──→  │  ┌──────────────────────────┐  │
│   (dev only)           │  │   redis                   │  │
│                        │  │   Port: 6379              │  │
│                        │  │   Volume: redis_data      │  │
│                        │  └──────────────────────────┘  │
│                        │                                │
│                        └────────────────────────────────┘
```

---

## 11.13 Docker Volumes — Data Persistence

### Masalah: Container itu Ephemeral

```bash
# Jalanin PostgreSQL tanpa volume
docker run -d --name postgres-test postgres:16-alpine

# Insert data...
# docker exec postgres-test psql -U postgres -c "CREATE TABLE test(id INT);"

# Hapus container
docker rm -f postgres-test

# Jalanin lagi
docker run -d --name postgres-test postgres:16-alpine

# Data tadi? HILANG!
# Karena container baru = filesystem baru
```

### Solusi: Volumes

Volume = storage yang **hidup di luar container**. Container dihapus, data tetap ada.

### 2 Tipe Volume

#### 1. Named Volume — Managed by Docker

```yaml
# docker-compose.yml
services:
  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data
      #   ↑ named volume      ↑ path di dalam container

volumes:
  postgres_data:    # Docker yang manage lokasi penyimpanannya
```

```bash
# Named volume disimpan di Docker internal storage
docker volume ls
# DRIVER    VOLUME NAME
# local     task-manager-postgres-data

# Inspect volume — lihat di mana disimpan
docker volume inspect task-manager-postgres-data
# "Mountpoint": "/var/lib/docker/volumes/task-manager-postgres-data/_data"

# Hapus volume
docker volume rm task-manager-postgres-data
```

**Kapan pakai:** Production. Data database, data Redis, data yang perlu persist.

#### 2. Bind Mount — Map Host Directory

```yaml
# docker-compose.yml
services:
  api:
    volumes:
      - .:/app
      # ↑ path di host (current directory)
      #    ↑ path di container
      # File di host dan container SYNC real-time
```

```bash
# Bind mount = symlink antara host dan container
# Edit file di host → otomatis berubah di container
# Edit file di container → otomatis berubah di host
```

**Kapan pakai:** Development. Source code sync untuk hot reload.

### Perbandingan

| Aspek | Named Volume | Bind Mount |
|-------|-------------|------------|
| **Managed by** | Docker | You (host filesystem) |
| **Location** | Docker internal | Anywhere on host |
| **Portability** | Portable | Tied to host path |
| **Performance** | Optimal | Bisa slower (terutama macOS) |
| **Use case** | Database data, persistent storage | Source code (development) |
| **Syntax** | `volume_name:/path` | `./host/path:/path` |

### Volume di Project Kita

```yaml
# Development
volumes:
  - .:/app                            # Bind mount — source code sync
  - /app/tmp                          # Anonymous volume — exclude tmp/
  - postgres_data_dev:/var/lib/postgresql/data  # Named volume — DB data

# Production
volumes:
  - postgres_data:/var/lib/postgresql/data    # Named volume — DB data
  - redis_data:/data                          # Named volume — Redis data
  # Ga ada bind mount — source code udah di-COPY ke image
```

---

## 11.14 Development vs Production — Side by Side

### File Structure

```
task-manager/
├── cmd/api/main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── model/
│   ├── repository/
│   ├── router/
│   └── service/
├── Dockerfile              ← Production (multi-stage)
├── Dockerfile.dev          ← Development (hot reload)
├── docker-compose.yml      ← Production compose
├── docker-compose.dev.yml  ← Development compose
├── .dockerignore           ← Files to exclude
├── .air.toml               ← Air hot reload config
├── .env.docker             ← Environment variables (git-ignored)
├── go.mod
└── go.sum
```

### Workflow Comparison

**Development:**

```bash
# Start development environment
docker compose -f docker-compose.dev.yml up

# Edit code → Air auto-rebuilds → API restarts
# Connect ke DB via pgAdmin: localhost:5432
# Connect ke Redis via redis-cli: localhost:6379
# Test API: http://localhost:8080/health

# Stop
docker compose -f docker-compose.dev.yml down
```

**Production:**

```bash
# Build optimized image + start
docker compose up -d --build

# Check health
curl http://localhost:8080/health

# View logs
docker compose logs -f api

# Stop
docker compose down
```

### Decision Table

| Pertanyaan | Development | Production |
|-----------|-------------|------------|
| Mau edit code dan langsung lihat hasilnya? | `docker-compose.dev.yml` | - |
| Mau deploy ke server? | - | `docker-compose.yml` |
| Mau test build production locally? | - | `docker compose up --build` |
| Image size penting? | Ga | `Dockerfile` multi-stage |
| Perlu connect ke DB langsung? | `ports: 5432:5432` | `expose: 5432` |

---

## 11.15 Debugging Docker — Troubleshooting

### Tool #1: docker logs

```bash
# Lihat logs container
docker logs task-manager-api

# Follow real-time
docker logs -f task-manager-api

# Lihat 50 baris terakhir
docker logs --tail 50 task-manager-api

# Logs sejak 5 menit lalu
docker logs --since 5m task-manager-api

# Dengan timestamp
docker logs -t task-manager-api
```

### Tool #2: docker exec

```bash
# Masuk ke shell container
docker exec -it task-manager-api /bin/sh

# Sekarang kamu "di dalam" container
# Bisa jalanin command:
ls -la /app/
env                    # lihat semua env vars
cat /etc/os-release    # lihat OS info
wget -qO- http://localhost:8080/health  # test endpoint dari dalam

# Keluar
exit
```

### Tool #3: docker inspect

```bash
# Lihat semua detail container
docker inspect task-manager-api

# Ambil IP address aja
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' task-manager-api

# Ambil environment variables
docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' task-manager-api
```

### Tool #4: docker compose logs

```bash
# Logs semua services
docker compose logs

# Logs service spesifik
docker compose logs api
docker compose logs postgres

# Follow semua service
docker compose logs -f

# Follow satu service
docker compose logs -f api
```

### Common Errors dan Solutions

#### Error 1: "Cannot connect to database"

```
Failed to connect to database: dial tcp 172.18.0.2:5432: connect: connection refused
```

**Penyebab:** API start sebelum PostgreSQL ready.

**Solusi:** Pakai `depends_on` dengan `condition: service_healthy`:

```yaml
api:
  depends_on:
    postgres:
      condition: service_healthy  # ← ini yang penting!
```

#### Error 2: "Port already in use"

```
Error starting userland proxy: listen tcp4 0.0.0.0:5432: bind: address already in use
```

**Penyebab:** Port 5432 udah dipake service lain di host (mungkin PostgreSQL local).

**Solusi:**

```bash
# Cari siapa yang pake port
lsof -i :5432

# Option 1: Stop service yang pake port itu
brew services stop postgresql

# Option 2: Ganti port mapping di docker-compose.yml
ports:
  - "5433:5432"    # host port 5433 → container port 5432
```

#### Error 3: "No space left on device"

```
Error: no space left on device
```

**Penyebab:** Docker image/container/volume terlalu banyak.

**Solusi:**

```bash
# Lihat disk usage Docker
docker system df

# Cleanup everything yang ga dipake
docker system prune -a

# Cleanup volumes yang ga dipake
docker volume prune
```

#### Error 4: "exec format error"

```
exec /app/main: exec format error
```

**Penyebab:** Binary di-build untuk architecture yang beda (misal build di ARM, run di x86).

**Solusi:** Pastikan `GOOS` dan `GOARCH` benar di Dockerfile:

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/api/
```

Atau kalau kamu di Mac Apple Silicon, untuk deploy ke Linux server:

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/api/
```

#### Error 5: Container exit immediately

```bash
docker ps -a
# STATUS: Exited (1) 5 seconds ago
```

**Debug:**

```bash
# Lihat logs kenapa exit
docker logs task-manager-api

# Common causes:
# - Binary crash (panic)
# - Config error (missing env var)
# - Port conflict inside container
# - Database connection failed
```

#### Error 6: "COPY failed: file not found"

```
COPY failed: file not found in build context
```

**Penyebab:** File yang mau di-COPY ga ada, atau kena exclude dari `.dockerignore`.

**Solusi:** Check `.dockerignore` — pastikan file yang dibutuhin ga di-ignore.

### Debugging Checklist

```
Container ga mau start?
├── docker logs <container>          → cek error message
├── docker ps -a                     → cek exit code
├── docker inspect <container>       → cek config
└── docker compose config            → cek compose file valid

Container start tapi app error?
├── docker exec -it <container> sh   → masuk ke container
├── env                              → cek env vars benar
├── wget localhost:8080/health       → test dari dalam
└── docker compose logs -f api       → watch logs real-time

Network issue?
├── docker network ls                → list networks
├── docker network inspect <name>    → lihat containers di network
└── docker exec api ping postgres    → test connectivity
```

---

## 11.16 Docker Best Practices

### 1. Leverage Layer Caching

Docker build itu **layer by layer**. Kalau satu layer ga berubah, Docker pakai cache. **Urutkan instruksi dari yang jarang berubah ke yang sering berubah.**

```dockerfile
# SALAH — setiap kali source code berubah, go mod download ulang
COPY . .
RUN go mod download
RUN go build -o main ./cmd/api/

# BENAR — dependency di-cache, cuma rebuild yang berubah
COPY go.mod go.sum ./       # Layer 1: jarang berubah
RUN go mod download          # Layer 2: di-cache kalau go.mod ga berubah
COPY . .                     # Layer 3: sering berubah
RUN go build -o main ./cmd/api/  # Layer 4: rebuild kalau source berubah
```

```
Kalau cuma ubah source code:
Layer 1: go.mod/go.sum   → CACHED ✓ (ga berubah)
Layer 2: go mod download → CACHED ✓ (ga berubah)
Layer 3: COPY . .        → REBUILD (source code berubah)
Layer 4: go build        → REBUILD

Kalau ga pakai caching strategy:
SEMUA layer rebuild dari awal → LAMBAT
```

### 2. Minimize Layers

Gabungkan RUN commands yang related:

```dockerfile
# KURANG OPTIMAL — 3 layers
RUN apk update
RUN apk add git
RUN apk add ca-certificates

# OPTIMAL — 1 layer
RUN apk add --no-cache git ca-certificates
```

### 3. Non-Root User

**Jangan jalanin app sebagai root.** Kalau ada vulnerability, attacker dapet akses root.

```dockerfile
# Bikin user biasa
RUN adduser -D -g '' appuser

# Switch ke user biasa
USER appuser

# Sekarang app jalan sebagai appuser, bukan root
```

### 4. Health Check

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1
```

Ini bikin Docker tau apakah app kamu **beneran healthy**, bukan cuma container running.

### 5. .dockerignore

Selalu bikin `.dockerignore`. Udah kita bahas di section 11.7.

### 6. Pin Image Versions

```dockerfile
# JANGAN — bisa berubah kapan aja
FROM golang:latest
FROM postgres:latest

# LAKUKAN — predictable dan reproducible
FROM golang:1.25-alpine
FROM postgres:16-alpine
```

### 7. Security Scanning

```bash
# Scan image untuk vulnerabilities
docker scout cves task-manager:latest

# Atau pakai tool external
# Trivy (populer):
# brew install aquasecurity/trivy/trivy
# trivy image task-manager:latest
```

### 8. Clean up di RUN

```dockerfile
# Install build tools, compile, lalu hapus build tools
RUN apk add --no-cache --virtual .build-deps gcc musl-dev \
    && go build -o main ./cmd/api/ \
    && apk del .build-deps
```

### Rangkuman Best Practices

| Practice | Kenapa |
|----------|--------|
| Cache dependencies dulu | Build lebih cepat |
| Multi-stage build | Image lebih kecil |
| Non-root user | Security |
| Pin image versions | Reproducible builds |
| .dockerignore | Smaller build context |
| Health check | Proper monitoring |
| Minimize layers | Smaller image |
| Security scanning | Catch vulnerabilities |

---

## 11.17 Practical Exercise — Hands-On

Sekarang saatnya praktik. Ikuti step by step dari nol.

### Pre-requisite

```bash
# Pastikan Docker Desktop running
docker version

# Pastikan di project directory
cd "/Users/irzan-aldi/Downloads/PROJECT/Devops Go"
```

### Step 1: Bikin Dockerfile

Bikin file `Dockerfile` di root project:

```dockerfile
# Dockerfile — Task Manager API (Production)

# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api/

# ---- Stage 2: Production ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./main"]
```

### Step 2: Bikin .dockerignore

Bikin file `.dockerignore` di root project:

```
.git
.gitignore
.vscode
.idea
.DS_Store
*.md
docs/
ebook/
LICENSE
Dockerfile
Dockerfile.*
docker-compose*.yml
.dockerignore
.env
.env.*
*_test.go
testdata/
main
tmp/
.air.toml
```

### Step 3: Build Image

```bash
# Build production image
docker build -t task-manager:v1 .

# Output:
# [+] Building 45.2s (17/17) FINISHED
# ...
# => exporting to image
# => => naming to docker.io/library/task-manager:v1

# Cek image size
docker images task-manager:v1
# REPOSITORY     TAG   IMAGE ID       CREATED         SIZE
# task-manager   v1    abc123def456   5 seconds ago   ~20MB
```

### Step 4: Bikin docker-compose.yml

Bikin file `docker-compose.yml` di root project (copy dari section 11.11 di atas).

### Step 5: Start Semua Services

```bash
# Build dan start
docker compose up -d --build

# Output:
# [+] Running 4/4
#  ✔ Network task-manager-network        Created
#  ✔ Container task-manager-redis        Started
#  ✔ Container task-manager-db           Started
#  ✔ Container task-manager-api          Started

# Cek status
docker compose ps
# NAME                  STATUS              PORTS
# task-manager-api      Up (healthy)        0.0.0.0:8080->8080/tcp
# task-manager-db       Up (healthy)        0.0.0.0:5432->5432/tcp
# task-manager-redis    Up (healthy)        0.0.0.0:6379->6379/tcp
```

### Step 6: Test API

```bash
# Health check
curl http://localhost:8080/health
# {"status":"ok"}

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Docker Test",
        "email": "docker@test.com",
        "password": "password123"
    }'
# {"data":{"id":1,"name":"Docker Test","email":"docker@test.com","role":"member"}}

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "email": "docker@test.com",
        "password": "password123"
    }'
# {"data":{"access_token":"eyJ...","refresh_token":"eyJ..."}}
```

### Step 7: Check Logs

```bash
# Lihat logs API
docker compose logs api
# task-manager-api  | Database connected successfully
# task-manager-api  | Database migrated successfully
# task-manager-api  | Redis connected successfully
# task-manager-api  | Server starting on port 8080 (env: production)

# Follow logs real-time
docker compose logs -f api

# Lihat logs semua services
docker compose logs -f
```

### Step 8: Exec Into Container

```bash
# Masuk ke API container
docker exec -it task-manager-api /bin/sh

# Di dalam container:
/app $ ls -la
# total 20
# drwxr-xr-x    1 root     root          4096 ...
# -rwxr-xr-x    1 root     root         15000 ... main

/app $ env | grep DB
# DB_HOST=postgres
# DB_PORT=5432
# DB_USER=postgres

/app $ wget -qO- http://localhost:8080/health
# {"status":"ok"}

/app $ exit

# Masuk ke PostgreSQL container
docker exec -it task-manager-db psql -U postgres -d task_manager

# Di dalam psql:
task_manager=# \dt
#           List of relations
#  Schema | Name   | Type  | Owner
# --------+--------+-------+----------
#  public | boards | table | postgres
#  public | tasks  | table | postgres
#  public | users  | table | postgres

task_manager=# SELECT id, name, email FROM users;
#  id |    name     |     email
# ----+-------------+----------------
#   1 | Docker Test | docker@test.com

task_manager=# \q

# Test Redis
docker exec -it task-manager-redis redis-cli
127.0.0.1:6379> PING
# PONG
127.0.0.1:6379> KEYS *
127.0.0.1:6379> exit
```

### Step 9: Stop dan Cleanup

```bash
# Stop semua containers (data preserved)
docker compose down

# Stop + hapus volumes (reset database!)
docker compose down -v

# Hapus image
docker rmi task-manager:v1
```

### Step 10: Development Mode

```bash
# Bikin Dockerfile.dev dan docker-compose.dev.yml (dari section 11.11)
# Bikin .air.toml (dari section 11.11)

# Start development
docker compose -f docker-compose.dev.yml up

# Sekarang edit file Go apa aja
# Air akan auto-rebuild dan restart API

# Contoh: edit internal/handler/health.go
# Ubah response, save, lihat logs — otomatis rebuild!
```

---

## 11.18 Cheat Sheet — Docker & Docker Compose

### Docker CLI

```bash
# ========== IMAGE ==========
docker build -t name:tag .            # Build image
docker build -t name:tag -f Dockerfile.dev .  # Build dari specific Dockerfile
docker build --no-cache -t name:tag . # Build tanpa cache
docker images                         # List images
docker rmi name:tag                   # Remove image
docker image prune                    # Remove unused images
docker tag name:tag newname:newtag    # Tag image
docker pull name:tag                  # Pull dari registry
docker push name:tag                  # Push ke registry

# ========== CONTAINER ==========
docker run -d name:tag                # Run di background
docker run -d -p 8080:8080 name:tag   # Run dengan port mapping
docker run -d --name myapp name:tag   # Run dengan nama
docker run -d -e KEY=VAL name:tag     # Run dengan env var
docker run -d --env-file .env name    # Run dengan env file
docker run -d -v vol:/path name:tag   # Run dengan volume
docker run -it name:tag /bin/sh       # Run interaktif (shell)
docker ps                             # List running containers
docker ps -a                          # List all containers
docker stop name                      # Stop container
docker start name                     # Start stopped container
docker restart name                   # Restart container
docker rm name                        # Remove container
docker rm -f name                     # Force remove

# ========== DEBUG ==========
docker logs name                      # View logs
docker logs -f name                   # Follow logs
docker logs --tail 100 name           # Last 100 lines
docker logs --since 5m name           # Logs sejak 5 menit lalu
docker exec -it name /bin/sh          # Shell ke container
docker exec name env                  # Lihat env vars
docker inspect name                   # Detail container
docker stats                          # Resource usage real-time

# ========== CLEANUP ==========
docker system df                      # Disk usage
docker system prune -a                # Remove semua unused
docker container prune                # Remove stopped containers
docker image prune -a                 # Remove unused images
docker volume prune                   # Remove unused volumes
docker network prune                  # Remove unused networks

# ========== VOLUME ==========
docker volume ls                      # List volumes
docker volume create name             # Create volume
docker volume rm name                 # Remove volume
docker volume inspect name            # Detail volume

# ========== NETWORK ==========
docker network ls                     # List networks
docker network create name            # Create network
docker network rm name                # Remove network
docker network inspect name           # Detail network
```

### Docker Compose CLI

```bash
# ========== LIFECYCLE ==========
docker compose up                     # Start (foreground)
docker compose up -d                  # Start (background)
docker compose up -d --build          # Build + start
docker compose up -d --force-recreate # Recreate containers
docker compose down                   # Stop + remove containers
docker compose down -v                # Stop + remove containers + volumes
docker compose stop                   # Stop (keep containers)
docker compose start                  # Start stopped containers
docker compose restart                # Restart all

# ========== STATUS ==========
docker compose ps                     # List containers
docker compose top                    # List processes
docker compose config                 # Validate compose file

# ========== LOGS ==========
docker compose logs                   # All logs
docker compose logs api               # Logs service spesifik
docker compose logs -f                # Follow all logs
docker compose logs -f api            # Follow service logs
docker compose logs --tail 50 api     # Last 50 lines

# ========== EXEC ==========
docker compose exec api /bin/sh       # Shell ke service
docker compose exec postgres psql -U postgres  # psql ke DB

# ========== BUILD ==========
docker compose build                  # Build all services
docker compose build api              # Build satu service
docker compose build --no-cache       # Build tanpa cache

# ========== SPECIFIC FILE ==========
docker compose -f docker-compose.dev.yml up     # Pakai file spesifik
docker compose -f docker-compose.dev.yml down   # Stop file spesifik
```

### Dockerfile Reference

```dockerfile
# ========== INSTRUCTIONS ==========
FROM image:tag                       # Base image
FROM image:tag AS name               # Named stage (multi-stage)
WORKDIR /path                        # Set working directory
COPY src dst                         # Copy file host → image
COPY --from=builder /src /dst        # Copy dari stage lain
RUN command                          # Run command saat build
ENV KEY=value                        # Set environment variable
ARG KEY=value                        # Build-time variable
EXPOSE port                          # Dokumentasi port
CMD ["executable", "arg"]            # Default command (overridable)
ENTRYPOINT ["executable"]            # Fixed command
USER username                        # Set user
HEALTHCHECK CMD command              # Health check
```

---

## 11.19 Checklist Bab Ini

### Konsep Dasar

- [ ] Paham apa itu Docker dan kenapa penting (container vs VM)
- [ ] Paham bedanya Image, Container, dan Registry
- [ ] Paham analogi shipping container
- [ ] Docker Desktop terinstall dan running

### Dockerfile

- [ ] Paham instruksi dasar: FROM, WORKDIR, COPY, RUN, EXPOSE, CMD
- [ ] Paham perbedaan CMD vs ENTRYPOINT
- [ ] Bisa nulis Dockerfile single-stage untuk Go
- [ ] Bisa nulis Dockerfile multi-stage (production-ready)
- [ ] Paham kenapa Go cocok banget buat Docker (single binary)
- [ ] Paham kenapa multi-stage build penting (500MB vs 20MB)
- [ ] Udah bikin .dockerignore

### Docker Commands

- [ ] Bisa build image (`docker build -t name .`)
- [ ] Bisa run container (`docker run -d -p 8080:8080 name`)
- [ ] Bisa lihat logs (`docker logs -f name`)
- [ ] Bisa exec ke container (`docker exec -it name /bin/sh`)
- [ ] Bisa stop/remove container dan image
- [ ] Bisa pass environment variables (inline, env-file)

### Docker Compose

- [ ] Paham kenapa Docker Compose dibutuhin (multi-container)
- [ ] Bisa nulis docker-compose.yml (services, volumes, networks)
- [ ] Bisa jalanin `docker compose up -d` dan `docker compose down`
- [ ] Paham depends_on dengan health check
- [ ] Compose file include Go API + PostgreSQL + Redis

### Networking & Volumes

- [ ] Paham bridge network dan container communication by service name
- [ ] Paham perbedaan `ports` vs `expose`
- [ ] Paham perbedaan named volume vs bind mount
- [ ] Paham kenapa volume penting untuk data persistence

### Development vs Production

- [ ] Punya Dockerfile.dev (hot reload) dan Dockerfile (production)
- [ ] Punya docker-compose.dev.yml dan docker-compose.yml
- [ ] Paham perbedaan dev vs prod setup
- [ ] Air hot reload bisa jalan di Docker

### Debugging & Best Practices

- [ ] Bisa debug container yang error (logs, exec, inspect)
- [ ] Paham common errors dan solusinya
- [ ] Paham layer caching strategy
- [ ] Paham kenapa non-root user penting
- [ ] Paham kenapa pin image versions penting

### Practical

- [ ] Udah build image Task Manager API
- [ ] Udah run full stack (API + PostgreSQL + Redis) pakai docker compose
- [ ] Udah test API endpoint dari container
- [ ] Udah exec ke container dan explore
- [ ] Udah coba development mode dengan hot reload

---

> **Next:** [Bab 12: CI/CD — GitHub Actions](12-cicd.md) — kita bakal automate semuanya. Push ke GitHub → otomatis test → otomatis build Docker image → otomatis deploy. Ini yang bikin kamu beneran jadi DevOps engineer.