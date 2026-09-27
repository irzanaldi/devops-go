# Bab 12: CI/CD — GitHub Actions dari Nol Sampai Auto Deploy

## Kenapa Bab Ini Penting?

Di bab sebelumnya, kamu udah bisa containerize Task Manager API pakai Docker. Image-nya kecil (~20MB), docker-compose-nya jalan, semuanya rapi. Tapi... **deploy-nya masih manual**.

Setiap kali kamu push code baru, prosesnya kayak gini:

```
1. git push
2. SSH ke server
3. git pull
4. docker compose down
5. docker compose up -d --build
6. Cek logs, pastiin ga error
7. Test endpoint, pastiin jalan
8. Kalau error, rollback manual
```

Itu **8 step manual**. Setiap. Kali. Deploy. Dan itu belum termasuk jalanin test, lint, security scan, dll. Kalau kamu deploy 5x seminggu — itu 40 step manual per minggu. Bakal ada saatnya kamu **lupa satu step**, dan production down.

**CI/CD menghilangkan semua itu.** Push ke GitHub → otomatis test → otomatis build Docker image → otomatis deploy. Kamu cuma perlu push code. Sisanya mesin yang handle.

> **Analogi gampang:** Bayangin kamu kerja di **pabrik mobil**. Tanpa CI/CD = setiap mobil dirakit manual satu-satu. Ada yang lupa pasang baut, ada yang cat-nya beda, quality control manual. Dengan CI/CD = **assembly line otomatis**. Setiap mobil lewat jalur yang sama — di-test, di-cat, di-inspeksi, di-approve — hasilnya **konsisten** setiap kali.

---

## 12.1 Apa Itu CI/CD?

### CI — Continuous Integration

**CI** itu proses otomatis yang jalan setiap kali developer push code ke repository. Tujuannya: **pastikan kode baru ga break apa-apa**.

```
Developer push code
        │
        ▼
┌─────────────────────────┐
│   Continuous Integration │
│                          │
│   1. Checkout code       │
│   2. Install deps        │
│   3. Run linter          │
│   4. Run unit tests      │
│   5. Run integration     │
│      tests               │
│   6. Code coverage       │
│   7. Build binary        │
│                          │
│   Semua PASS? ───→ ✓ OK  │
│   Ada FAIL?  ───→ ✗ STOP │
└─────────────────────────┘
```

**Prinsip CI:**
- **Merge sering** — jangan nunggu 2 minggu baru merge. Merge kecil-kecil, sering-sering
- **Test otomatis** — setiap push, test jalan otomatis
- **Fail fast** — kalau ada yang error, langsung tau. Ga nunggu deploy baru ketahuan

### CD — Continuous Delivery vs Continuous Deployment

Nah, CD itu ada **dua versi**. Banyak yang bingung ini:

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│  Continuous DELIVERY                                     │
│  ════════════════                                        │
│  Code → Build → Test → [Manual Approval] → Deploy       │
│                              ▲                           │
│                         Ada orang                        │
│                         yang approve                     │
│                                                          │
│  "Selalu SIAP deploy, tapi deploy-nya manual trigger"    │
│                                                          │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  Continuous DEPLOYMENT                                   │
│  ═══════════════════                                     │
│  Code → Build → Test → Deploy (OTOMATIS!)                │
│                                                          │
│  Ga ada approval manual. Merge ke main = langsung live.  │
│                                                          │
│  "Setiap kode yang lolos test langsung ke production"    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Perbandingan

| Aspek | CI | CD (Delivery) | CD (Deployment) |
|-------|------|--------------|----------------|
| **Apa yang otomatis** | Test & build | Test, build, prepare deploy | Test, build, deploy |
| **Manual step** | Ga ada | Approve deploy | Ga ada |
| **Risk** | Low | Low | Medium (harus punya test yang bagus) |
| **Cocok untuk** | Semua team | Kebanyakan team | Team mature dengan test coverage tinggi |
| **Contoh** | Setiap push → run tests | Merge ke main → siap deploy, klik tombol | Merge ke main → langsung live |

### Kenapa CI/CD Penting?

**Tanpa CI/CD (manual deploy):**

```
1. Developer A push code ─── ga ada test ─── deploy langsung
2. Production error ─── baru tau setelah user complain
3. Rollback manual ─── 30 menit downtime
4. Developer B push code ─── ternyata conflict sama A
5. Bug masuk production lagi
6. Weekend deploy? Siapa yang mau lembur?
```

**Dengan CI/CD:**

```
1. Developer A push code ─── test otomatis jalan
2. Test fail ─── PR ditolak, ga bisa merge
3. Developer A fix ─── test pass ─── merge ─── deploy otomatis
4. 3 menit dari push sampai live
5. Rollback? Revert commit, deploy otomatis lagi
6. Weekend deploy? Push aja, pipeline yang handle
```

**Real numbers dari industri:**
- Deploy manual: 1-4 jam per deploy, 1-2x per minggu
- Dengan CI/CD: 3-10 menit per deploy, 10-50x per hari
- MTTR (Mean Time to Recovery) turun **drastis**

### Analogi Lain — Assembly Line

```
Tanpa CI/CD = Tukang Roti Manual:
────────────────────────────────
1. Bikin adonan (code)
2. Cek adonannya bener ga (test) ← KADANG LUPA
3. Masukkin oven (build) ← KADANG SETTING-NYA BEDA
4. Kirim ke toko (deploy) ← KADANG TELAT

Dengan CI/CD = Pabrik Roti Otomatis:
────────────────────────────────────
1. Masukin bahan (push code)
2. Mesin otomatis:
   - Cek bahan ✓ (lint)
   - Bikin adonan ✓ (build)
   - Quality check ✓ (test)
   - Masukin oven ✓ (Docker build)
   - Kirim ke toko ✓ (deploy)
3. Kalau ada bahan busuk → STOP! Ga jadi kirim.
```

---

## 12.2 CI/CD Tools — Mana yang Dipilih?

### Landscape CI/CD Tools

Ada banyak banget tools CI/CD di pasaran. Ini yang paling populer:

| Tool | Hosting | Free Tier | Integrated With | Bahasa Config |
|------|---------|-----------|----------------|--------------|
| **GitHub Actions** | Cloud (GitHub) | 2000 min/bulan (free repo) | GitHub | YAML |
| **GitLab CI** | Cloud/Self-hosted | 400 min/bulan | GitLab | YAML |
| **Jenkins** | Self-hosted | Free (open source) | Anything | Groovy |
| **CircleCI** | Cloud | 6000 min/bulan | GitHub/Bitbucket | YAML |
| **Travis CI** | Cloud | Free (open source) | GitHub | YAML |
| **AWS CodePipeline** | Cloud (AWS) | 1 pipeline free | AWS | JSON/YAML |
| **Azure DevOps** | Cloud (Azure) | 1800 min/bulan | Azure Repos/GitHub | YAML |

### Kenapa GitHub Actions?

Kita pakai **GitHub Actions** karena:

1. **Free** — 2000 menit/bulan untuk repo public (unlimited), private juga dapet 2000 menit
2. **Integrated** — Udah built-in di GitHub. Ga perlu setup server terpisah
3. **Popular** — Paling banyak dipakai di open source dan industry
4. **Ecosystem** — Ribuan "Actions" siap pakai (checkout, setup-go, docker-build, dll)
5. **Simple** — YAML config, ga perlu belajar Groovy (Jenkins) atau tool baru
6. **Scalable** — Dari personal project sampai enterprise
7. **Community** — Banyak contoh, banyak tutorial, Stack Overflow full jawabannya

### GitHub Actions vs Jenkins (Perbandingan Detail)

```
GitHub Actions:                     Jenkins:
─────────────                       ─────────
+ Zero setup                        - Harus setup server sendiri
+ Free tier generous                 + Free (open source)
+ YAML (simple)                     - Groovy (complex)
+ Built-in di GitHub                - Perlu integrate manual
+ Auto-scaling runners              - Harus manage server capacity
+ Marketplace actions               + Plugin ecosystem besar
- Vendor lock-in (GitHub)           + Self-hosted, ga lock-in
- Limited customization             + Sangat customizable
```

**Rule of thumb:**
- **Startup / personal project / small team** → GitHub Actions
- **Enterprise dengan kebutuhan custom** → Jenkins atau GitLab CI
- **Belajar CI/CD pertama kali** → GitHub Actions (paling gampang)

Kita pakai **GitHub Actions**. Let's go.

---

## 12.3 GitHub Actions — Konsep Dasar

### 6 Konsep yang Harus Kamu Paham

Sebelum nulis workflow file, kamu harus paham 6 konsep ini:

```
┌─────────────────────────────────────────────────────────────┐
│                        WORKFLOW                              │
│  (1 file YAML = 1 workflow)                                  │
│                                                              │
│  Event/Trigger ─────→ Workflow starts                        │
│  (push, PR, cron)                                            │
│                                                              │
│  ┌─────────── Job 1: test ────────────┐                      │
│  │  Runner: ubuntu-latest              │                     │
│  │                                     │                     │
│  │  Step 1: Checkout code (Action)     │                     │
│  │  Step 2: Setup Go (Action)          │                     │
│  │  Step 3: Run tests (run command)    │                     │
│  │  Step 4: Upload coverage (Action)   │                     │
│  └─────────────────────────────────────┘                     │
│           │                                                  │
│           │ needs: test (dependency)                          │
│           ▼                                                  │
│  ┌─────────── Job 2: build ───────────┐                      │
│  │  Runner: ubuntu-latest              │                     │
│  │                                     │                     │
│  │  Step 1: Checkout code (Action)     │                     │
│  │  Step 2: Build Docker image (run)   │                     │
│  │  Step 3: Push to registry (Action)  │                     │
│  └─────────────────────────────────────┘                     │
│           │                                                  │
│           │ needs: build                                      │
│           ▼                                                  │
│  ┌─────────── Job 3: deploy ──────────┐                      │
│  │  Runner: ubuntu-latest              │                     │
│  │                                     │                     │
│  │  Step 1: SSH to server              │                     │
│  │  Step 2: Pull new image             │                     │
│  │  Step 3: Restart containers         │                     │
│  └─────────────────────────────────────┘                     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 1. Workflow

**Workflow** = keseluruhan pipeline CI/CD kamu. Didefinisikan dalam **1 file YAML** di folder `.github/workflows/`.

```
.github/
└── workflows/
    ├── ci.yml          ← Workflow 1: test & lint
    ├── cd.yml          ← Workflow 2: build & deploy
    └── release.yml     ← Workflow 3: create release
```

Kamu bisa punya **banyak workflow** dalam satu repo. Setiap workflow = 1 file.

### 2. Event / Trigger

**Event** = apa yang memicu workflow jalan. Contoh: push code, bikin PR, schedule, manual trigger.

```yaml
on:
  push:                    # Setiap push ke branch tertentu
    branches: [main]
  pull_request:            # Setiap PR dibuat/update
    branches: [main]
  schedule:                # Cron job (jadwal)
    - cron: '0 0 * * *'   # Setiap hari jam 00:00 UTC
  workflow_dispatch:       # Manual trigger dari GitHub UI
```

### 3. Job

**Job** = unit kerja yang jalan di **satu runner** (satu virtual machine). Setiap job punya environment sendiri.

```yaml
jobs:
  test:           # Job 1
    runs-on: ubuntu-latest
    steps: [...]

  build:          # Job 2
    runs-on: ubuntu-latest
    needs: test   # Tunggu job "test" selesai dulu
    steps: [...]
```

- Job secara default jalan **parallel** (bersamaan)
- Kalau mau sequential, pakai `needs` untuk define dependency

### 4. Step

**Step** = satu aksi dalam job. Bisa berupa:
- **Action** (`uses`) — pakai action yang udah ada (dari marketplace)
- **Command** (`run`) — jalanin shell command

```yaml
steps:
  - uses: actions/checkout@v4       # Action: checkout code
  - uses: actions/setup-go@v5      # Action: install Go
  - run: go test ./...              # Command: jalanin test
  - run: go build -o main ./cmd/api/  # Command: build binary
```

### 5. Runner

**Runner** = virtual machine yang jalanin job. GitHub provide runner gratis:

| Runner | OS | CPU | RAM | Disk |
|--------|----|-----|-----|------|
| `ubuntu-latest` | Ubuntu 24.04 | 2 cores | 7 GB | 14 GB SSD |
| `macos-latest` | macOS 14 | 3 cores | 14 GB | 14 GB SSD |
| `windows-latest` | Windows Server 2022 | 2 cores | 7 GB | 14 GB SSD |

Kebanyakan CI/CD pakai `ubuntu-latest` karena paling cepat dan gratis.

### 6. Action

**Action** = building block reusable yang bisa kamu pakai di step. Kayak npm package tapi untuk CI/CD.

```yaml
# Action populer:
- uses: actions/checkout@v4          # Checkout repo
- uses: actions/setup-go@v5         # Install Go
- uses: actions/cache@v4            # Cache dependencies
- uses: docker/build-push-action@v6 # Build & push Docker image
- uses: docker/login-action@v3      # Login ke Docker registry
```

Action ada di **GitHub Marketplace**: [github.com/marketplace?type=actions](https://github.com/marketplace?type=actions)

### Analogi — Perbandingan dengan NestJS

```
GitHub Actions Concept    ≈    NestJS Equivalent
───────────────────────        ──────────────────
Workflow                  ≈    package.json scripts
Job                       ≈    npm script group
Step                      ≈    Single npm command
Action (uses)             ≈    npm package
Runner                    ≈    Node.js runtime
Event (on: push)          ≈    npm lifecycle hook (postinstall)
Secrets                   ≈    .env file
```

---

## 12.4 Workflow File — Anatomi Lengkap

### Lokasi File

Workflow file **harus** ada di folder `.github/workflows/` dari root repo kamu. Nama file bebas, tapi harus `.yml` atau `.yaml`.

```
task-manager/                   ← root repo
├── .github/
│   └── workflows/
│       ├── ci.yml              ← CI workflow
│       ├── cd.yml              ← CD workflow
│       └── release.yml         ← Release workflow
├── cmd/api/main.go
├── internal/
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

### Anatomi Workflow File

```yaml
# ============================================
# File: .github/workflows/ci.yml
# Workflow CI — Test & Lint
# ============================================

# 1. NAMA — muncul di GitHub Actions tab
name: CI Pipeline

# 2. TRIGGER — kapan workflow ini jalan
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

# 3. ENVIRONMENT VARIABLES — global (bisa diakses semua job)
env:
  GO_VERSION: '1.25'
  APP_NAME: task-manager

# 4. JOBS — unit kerja
jobs:
  # Job 1: Test
  test:
    # Runner — VM yang jalanin job ini
    name: Run Tests
    runs-on: ubuntu-latest

    # Steps — aksi-aksi dalam job
    steps:
      # Step 1: Checkout code dari repo
      - name: Checkout code
        uses: actions/checkout@v4

      # Step 2: Install Go
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      # Step 3: Download dependencies
      - name: Download dependencies
        run: go mod download

      # Step 4: Run tests
      - name: Run tests
        run: go test ./... -v -coverprofile=coverage.out

      # Step 5: Upload coverage
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
```

### Penjelasan Setiap Bagian

**`name:`** — Nama workflow. Muncul di tab "Actions" di GitHub.

**`on:`** — Event/trigger. Kapan workflow ini dijalankan.

**`env:`** — Environment variables global. Bisa diakses di semua jobs dan steps.

**`jobs:`** — Kumpulan jobs. Setiap job punya nama (`test`), runner (`ubuntu-latest`), dan steps.

**`runs-on:`** — Runner mana yang dipake. Hampir selalu `ubuntu-latest`.

**`steps:`** — Array of steps. Setiap step punya:
- `name:` — deskripsi (opsional tapi recommended)
- `uses:` — pakai Action dari marketplace **ATAU**
- `run:` — jalanin shell command
- `with:` — parameter untuk Action

**`${{ env.GO_VERSION }}`** — Expression syntax. Bisa akses env vars, secrets, context, dll.

### YAML Syntax Basics (Buat yang Belum Familiar)

YAML itu format konfigurasi yang pakai indentation (kayak Python). Kalau kamu dari JSON, ini konversinya:

```yaml
# YAML — Key-value
name: CI Pipeline          # string
timeout: 30                # number
verbose: true              # boolean

# YAML — Object (nested)
database:
  host: localhost
  port: 5432
  name: task_manager

# YAML — Array (list)
branches:
  - main
  - develop
  - staging

# YAML — Array inline
branches: [main, develop, staging]

# YAML — Multi-line string
description: |
  Ini baris pertama
  Ini baris kedua
  Ini baris ketiga

# YAML — Comment
# Ini comment — diawali dengan #
```

**Rules penting:**
1. **Indentation HARUS pakai spaces** (bukan tab). 2 spaces standard
2. **Case sensitive** — `Name` dan `name` itu beda
3. **Colon harus diikuti space** — `key: value` (bukan `key:value`)

---

## 12.5 Event Triggers — Kapan Workflow Jalan

### push — Setiap Push ke Branch

```yaml
on:
  push:
    branches:
      - main              # Push ke main
      - develop           # Push ke develop
      - 'release/**'      # Push ke branch release/* (wildcard)
```

### pull_request — Setiap PR

```yaml
on:
  pull_request:
    branches:
      - main              # PR yang target-nya main
      - develop           # PR yang target-nya develop
    types:
      - opened            # PR baru dibuat
      - synchronize       # PR di-update (push baru ke branch PR)
      - reopened          # PR dibuka lagi
```

### schedule — Cron Job

```yaml
on:
  schedule:
    # Cron format: minute hour day-of-month month day-of-week
    - cron: '0 0 * * *'      # Setiap hari jam 00:00 UTC
    - cron: '0 6 * * 1-5'    # Weekday jam 06:00 UTC
    - cron: '0 */6 * * *'    # Setiap 6 jam

# Cron cheat sheet:
# ┌───── minute (0 - 59)
# │ ┌───── hour (0 - 23)
# │ │ ┌───── day of month (1 - 31)
# │ │ │ ┌───── month (1 - 12)
# │ │ │ │ ┌───── day of week (0 - 6, 0=Sunday)
# │ │ │ │ │
# * * * * *
```

### workflow_dispatch — Manual Trigger

```yaml
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deploy to environment'
        required: true
        default: 'staging'
        type: choice
        options:
          - staging
          - production
      debug:
        description: 'Enable debug mode'
        required: false
        type: boolean
        default: false
```

Ini bikin tombol **"Run workflow"** muncul di GitHub UI. Kamu bisa trigger manual dan input parameter.

### release — Saat Release Baru

```yaml
on:
  release:
    types:
      - published          # Release dipublish
      - created            # Release dibuat
```

### Kombinasi Triggers

```yaml
# Workflow ini jalan saat:
# - Push ke main atau develop
# - PR ke main
# - Manual trigger
# - Setiap hari jam 00:00 UTC
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  workflow_dispatch:
  schedule:
    - cron: '0 0 * * *'
```

### Path Filtering — Hanya Trigger Kalau File Tertentu Berubah

```yaml
on:
  push:
    branches: [main]
    paths:
      - '**.go'            # Hanya kalau file .go berubah
      - 'go.mod'           # Atau go.mod berubah
      - 'go.sum'           # Atau go.sum berubah
      - 'Dockerfile'       # Atau Dockerfile berubah
    paths-ignore:
      - '**.md'            # SKIP kalau cuma .md yang berubah
      - 'docs/**'          # SKIP kalau cuma docs/ yang berubah
      - 'ebook/**'         # SKIP kalau cuma ebook/ yang berubah
```

**Ini penting banget!** Kalau kamu cuma update README, ga perlu jalanin test dan build. Hemat waktu dan GitHub Actions minutes.

### Rangkuman Triggers

| Trigger | Kapan | Use Case |
|---------|-------|----------|
| `push` | Setiap push ke branch | CI — run tests |
| `pull_request` | PR dibuat/update | CI — check sebelum merge |
| `schedule` | Cron schedule | Nightly builds, security scan |
| `workflow_dispatch` | Manual trigger | Emergency deploy, ad-hoc tasks |
| `release` | Release dipublish | Build release artifacts |

---

## 12.6 CI Pipeline — Test & Lint

Sekarang kita mulai bikin **workflow beneran** yang bisa kamu copy-paste ke project Task Manager API.

### Workflow File: CI — Test & Lint

```yaml
# ============================================
# File: .github/workflows/ci.yml
# CI Pipeline — Lint, Test, Coverage
# ============================================

name: CI

on:
  push:
    branches: [main, develop]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'ebook/**'
  pull_request:
    branches: [main, develop]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'ebook/**'

env:
  GO_VERSION: '1.25'

jobs:
  # ─────────────────────────────────────────
  # Job 1: Lint — Check code quality
  # ─────────────────────────────────────────
  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  # ─────────────────────────────────────────
  # Job 2: Test — Unit tests
  # ─────────────────────────────────────────
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Download dependencies
        run: go mod download

      - name: Run tests with coverage
        run: go test ./... -v -race -coverprofile=coverage.out -covermode=atomic

      - name: Show coverage summary
        run: go tool cover -func=coverage.out

      - name: Upload coverage artifact
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.out
          retention-days: 7

  # ─────────────────────────────────────────
  # Job 3: Integration Test — Dengan database
  # ─────────────────────────────────────────
  integration-test:
    name: Integration Test
    runs-on: ubuntu-latest

    # Service containers — PostgreSQL & Redis untuk test
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: task_manager_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U postgres"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Download dependencies
        run: go mod download

      - name: Run integration tests
        env:
          APP_ENV: test
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: testpassword
          DB_NAME: task_manager_test
          REDIS_HOST: localhost
          REDIS_PORT: 6379
          JWT_SECRET: test-secret-key
        run: go test ./... -v -tags=integration -coverprofile=coverage-integration.out

      - name: Upload integration coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-integration
          path: coverage-integration.out
          retention-days: 7
```

### Penjelasan Detail

#### golangci-lint Action

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    args: --timeout=5m
```

`golangci-lint` itu **meta-linter** untuk Go. Dia jalanin banyak linter sekaligus:
- `govet` — cari bug umum
- `errcheck` — cek error yang ga di-handle
- `staticcheck` — static analysis
- `unused` — cari kode yang ga dipake
- `ineffassign` — cari assignment yang sia-sia
- Dan puluhan lainnya

`--timeout=5m` penting karena di CI kadang lint butuh waktu lebih lama.

#### Service Containers

```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: testpassword
      POSTGRES_DB: task_manager_test
    ports:
      - 5432:5432
    options: >-
      --health-cmd "pg_isready -U postgres"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

Service containers itu **Docker container yang jalan di samping job**. Sama kayak docker-compose, tapi di CI. GitHub Actions otomatis:
1. Pull image (`postgres:16-alpine`)
2. Start container
3. Tunggu health check pass
4. Expose port ke runner
5. Step kamu bisa connect ke `localhost:5432`
6. Setelah job selesai, container dihapus

**Ini persis kayak docker-compose.yml kita!** Tapi dalam format GitHub Actions.

#### Test Flags

```bash
go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
```

| Flag | Fungsi |
|------|--------|
| `./...` | Test semua package |
| `-v` | Verbose output |
| `-race` | Detect race condition |
| `-coverprofile=coverage.out` | Generate coverage file |
| `-covermode=atomic` | Accurate coverage untuk concurrent code |

#### Hasil di GitHub

Setelah push, kamu bisa lihat hasilnya di tab **Actions** di repo GitHub:

```
✓ CI Pipeline                          All checks have passed
  ├── ✓ Lint                           15s
  ├── ✓ Test                           32s
  └── ✓ Integration Test               45s
```

Kalau ada yang fail:

```
✗ CI Pipeline                          Some checks were not successful
  ├── ✓ Lint                           15s
  ├── ✗ Test                           28s  ← FAIL!
  └── ✓ Integration Test               45s
```

Klik job yang fail → kamu bisa lihat log detail step mana yang error.

---

## 12.7 CI Pipeline — Build & Push Docker Image

Setelah test pass, step berikutnya: **build Docker image dan push ke registry**.

### Docker Hub vs GitHub Container Registry (GHCR)

| Aspek | Docker Hub | GHCR (ghcr.io) |
|-------|-----------|----------------|
| **Harga** | 1 private repo gratis | Unlimited private (free) |
| **URL** | `docker.io/username/repo` | `ghcr.io/username/repo` |
| **Integration** | Manual login | Built-in GitHub token |
| **Visibility** | Public/Private | Tied to GitHub repo visibility |
| **Rate limit** | 100 pulls/6h (anonymous) | Generous |

**Recommendation:** Pakai **GHCR** kalau repo di GitHub. Lebih simple karena udah integrated.

### Workflow: Build & Push Docker Image

```yaml
# ============================================
# File: .github/workflows/build.yml
# Build Docker image & push to GHCR
# ============================================

name: Build & Push

on:
  push:
    branches: [main]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'ebook/**'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  # github.repository = "irzan-aldi/task-manager"
  # Jadi IMAGE_NAME = "irzan-aldi/task-manager"

jobs:
  build-and-push:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest

    # Permission untuk push ke GHCR
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # Docker Buildx — builder canggih yang support caching
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      # Login ke GitHub Container Registry
      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          # GITHUB_TOKEN otomatis tersedia — ga perlu setup manual!

      # Generate metadata (tags, labels)
      - name: Docker metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            # Tag: latest (selalu update)
            type=raw,value=latest
            # Tag: commit SHA (unique per commit)
            type=sha,prefix=
            # Tag: branch name
            type=ref,event=branch
            # Tag: semver (kalau ada tag v1.2.3)
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}

      # Build & push Docker image
      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Penjelasan

#### GITHUB_TOKEN

```yaml
password: ${{ secrets.GITHUB_TOKEN }}
```

`GITHUB_TOKEN` itu **token otomatis** yang GitHub bikin untuk setiap workflow run. Ga perlu bikin manual, ga perlu simpan di secrets. Dia punya permission berdasarkan `permissions:` yang kamu define.

#### Docker Metadata Action

```yaml
- name: Docker metadata
  id: meta
  uses: docker/metadata-action@v5
```

Action ini otomatis generate **tags** berdasarkan Git context:

```
Push ke branch main, commit abc1234:
→ ghcr.io/irzan-aldi/task-manager:latest
→ ghcr.io/irzan-aldi/task-manager:abc1234
→ ghcr.io/irzan-aldi/task-manager:main

Git tag v1.2.3:
→ ghcr.io/irzan-aldi/task-manager:1.2.3
→ ghcr.io/irzan-aldi/task-manager:1.2
```

#### Build Push Action + GHA Cache

```yaml
- name: Build and push
  uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    cache-from: type=gha        # Baca cache dari GitHub Actions cache
    cache-to: type=gha,mode=max # Simpan cache ke GitHub Actions cache
```

`cache-from` dan `cache-to` itu **game changer**. Build pertama mungkin 2-3 menit. Build kedua (kalau Dockerfile ga berubah banyak) bisa **30 detik** karena layer di-cache.

### Push ke Docker Hub (Alternatif)

Kalau kamu mau push ke **Docker Hub** instead of GHCR:

```yaml
env:
  REGISTRY: docker.io
  IMAGE_NAME: yourusername/task-manager

# ...

- name: Login to Docker Hub
  uses: docker/login-action@v3
  with:
    registry: docker.io
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}
    # Ini perlu setup di GitHub Secrets (bukan otomatis kayak GITHUB_TOKEN)
```

---

## 12.8 CD Pipeline — Auto Deploy

### Deploy Strategy Options

Ada beberapa cara deploy dari CI/CD:

| Strategy | Complexity | Cocok untuk |
|----------|-----------|-------------|
| **SSH to server** | Simple | VPS, single server |
| **Docker Compose pull** | Simple | VPS dengan docker-compose |
| **AWS ECS** | Medium | AWS infrastructure |
| **Kubernetes** | Complex | Kubernetes cluster |
| **Serverless** | Medium | AWS Lambda, Cloud Run |

Kita mulai dari yang **paling simple** dulu: SSH ke server.

### Workflow: SSH Deploy (Simple)

```yaml
# ============================================
# File: .github/workflows/deploy.yml
# CD Pipeline — Deploy ke server via SSH
# ============================================

name: Deploy

on:
  # Trigger setelah workflow CI selesai dan sukses
  workflow_run:
    workflows: ["CI"]
    types:
      - completed
    branches: [main]

  # Atau manual trigger
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        default: 'production'
        type: choice
        options:
          - staging
          - production

jobs:
  deploy:
    name: Deploy to Server
    runs-on: ubuntu-latest
    # Hanya deploy kalau CI workflow SUKSES
    if: ${{ github.event.workflow_run.conclusion == 'success' || github.event_name == 'workflow_dispatch' }}

    # Environment protection — bisa tambah required reviewers
    environment:
      name: production
      url: https://api.yourdomain.com

    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_SSH_KEY }}
          port: ${{ secrets.SERVER_PORT }}
          script: |
            # Navigate ke project directory
            cd /home/deploy/task-manager

            # Pull image terbaru
            docker compose pull

            # Restart dengan image baru (zero-downtime)
            docker compose up -d --force-recreate --no-deps api

            # Cleanup image lama
            docker image prune -f

            # Health check
            sleep 5
            curl -f http://localhost:8080/health || exit 1

            echo "Deploy successful!"

      - name: Verify deployment
        run: |
          # Test dari luar
          sleep 10
          curl -f https://api.yourdomain.com/health || echo "Warning: External health check failed"
```

### Penjelasan Deploy Script

```bash
# Pull image terbaru dari registry
docker compose pull

# Restart HANYA service api (ga restart postgres/redis)
docker compose up -d --force-recreate --no-deps api
# --force-recreate: bikin container baru (bukan reuse yang lama)
# --no-deps: jangan restart dependency (postgres, redis)

# Cleanup image yang ga dipake
docker image prune -f

# Health check — pastiin app beneran jalan
sleep 5
curl -f http://localhost:8080/health || exit 1
# exit 1 = kalau health check fail, deploy dianggap gagal
```

### Setup di Server

Sebelum deploy workflow bisa jalan, kamu perlu setup server dulu:

```bash
# Di server production:

# 1. Bikin directory untuk project
mkdir -p /home/deploy/task-manager
cd /home/deploy/task-manager

# 2. Bikin docker-compose.yml (production)
cat > docker-compose.yml << 'EOF'
services:
  api:
    image: ghcr.io/irzan-aldi/task-manager:latest
    container_name: task-manager-api
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=task_manager
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - app-network
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    container_name: task-manager-db
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=task_manager
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: task-manager-redis
    volumes:
      - redis_data:/data
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    command: redis-server --appendonly yes

volumes:
  postgres_data:
  redis_data:

networks:
  app-network:
    driver: bridge
EOF

# 3. Bikin .env file di server
cat > .env << 'EOF'
DB_PASSWORD=your-super-secret-password
JWT_SECRET=your-jwt-secret-key-production
EOF

# 4. Login ke GHCR (supaya bisa pull image)
echo $GHCR_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 5. Start services
docker compose up -d
```

### Deploy ke AWS ECS (Mention)

Kalau kamu pakai AWS ECS, workflow-nya sedikit beda:

```yaml
# Simplified — AWS ECS deploy
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: ap-southeast-1

- name: Login to Amazon ECR
  uses: aws-actions/amazon-ecr-login@v2

- name: Deploy to ECS
  uses: aws-actions/amazon-ecs-deploy-task-definition@v2
  with:
    task-definition: task-definition.json
    service: task-manager-service
    cluster: task-manager-cluster
    wait-for-service-stability: true
```

### Deploy ke Kubernetes (Mention)

```yaml
# Simplified — Kubernetes deploy
- name: Set up kubectl
  uses: azure/setup-kubectl@v4

- name: Set Kubernetes context
  uses: azure/k8s-set-context@v4
  with:
    kubeconfig: ${{ secrets.KUBE_CONFIG }}

- name: Deploy to Kubernetes
  run: |
    kubectl set image deployment/task-manager \
      api=ghcr.io/irzan-aldi/task-manager:${{ github.sha }} \
      --namespace production
    kubectl rollout status deployment/task-manager --namespace production
```

Kita bakal cover Kubernetes dan AWS lebih detail di bab-bab selanjutnya. Untuk sekarang, SSH deploy udah lebih dari cukup.

---

## 12.9 Secrets & Environment Variables

### Apa Itu GitHub Secrets?

Secrets itu **variabel terenkripsi** yang bisa kamu simpan di GitHub dan akses di workflow. Cocok untuk:
- Password database
- API keys
- SSH keys
- Docker Hub tokens
- JWT secrets

**Penting:** Secrets itu **ga pernah muncul di logs**. Kalau workflow print secret, GitHub otomatis mask jadi `***`.

### Cara Menambah Secrets di GitHub

```
1. Buka repo di GitHub
2. Klik "Settings" tab
3. Di sidebar kiri, klik "Secrets and variables" → "Actions"
4. Klik "New repository secret"
5. Isi nama dan value
6. Klik "Add secret"
```

### Secrets yang Perlu Kamu Setup

```
Repository Secrets:
├── SERVER_HOST          → IP server production (misal: 103.xxx.xxx.xxx)
├── SERVER_USER          → SSH username (misal: deploy)
├── SERVER_SSH_KEY       → SSH private key (isi dari ~/.ssh/id_ed25519)
├── SERVER_PORT          → SSH port (misal: 22)
├── DOCKERHUB_USERNAME   → Docker Hub username (kalau pakai Docker Hub)
├── DOCKERHUB_TOKEN      → Docker Hub access token (kalau pakai Docker Hub)
└── PRODUCTION_DB_PASSWORD → Password DB production (kalau perlu di CI)
```

### Cara Akses Secrets di Workflow

```yaml
# Akses via ${{ secrets.NAMA_SECRET }}
steps:
  - name: Deploy via SSH
    uses: appleboy/ssh-action@v1
    with:
      host: ${{ secrets.SERVER_HOST }}
      username: ${{ secrets.SERVER_USER }}
      key: ${{ secrets.SERVER_SSH_KEY }}

  - name: Login to Docker Hub
    run: |
      echo "${{ secrets.DOCKERHUB_TOKEN }}" | docker login -u "${{ secrets.DOCKERHUB_USERNAME }}" --password-stdin
```

### GITHUB_TOKEN — Built-in Token

`GITHUB_TOKEN` itu **token otomatis** yang GitHub generate untuk setiap workflow run. Ga perlu bikin manual.

```yaml
# Ga perlu setup secret — otomatis tersedia
- name: Login to GHCR
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}   # Otomatis ada!
```

**Permissions GITHUB_TOKEN:**

```yaml
# Default permissions (restrictive)
permissions:
  contents: read

# Kalau mau push Docker image
permissions:
  contents: read
  packages: write

# Kalau mau write PR comments
permissions:
  contents: read
  pull-requests: write
```

### Environment Variables vs Secrets

```yaml
# Environment variables — ga sensitif, visible di logs
env:
  GO_VERSION: '1.25'
  APP_NAME: task-manager

# Secrets — sensitif, di-mask di logs
steps:
  - run: echo "DB password is ${{ secrets.DB_PASSWORD }}"
    # Output: "DB password is ***"
```

| Aspek | env | secrets |
|-------|-----|---------|
| **Visibility** | Visible di logs | Masked (`***`) |
| **Set di mana** | Workflow file | GitHub UI |
| **Cocok untuk** | Go version, app name | Password, API key, SSH key |
| **Bisa lihat value** | Ya | Tidak (sekali diset, ga bisa dilihat lagi) |

### Environment Protection Rules

GitHub punya fitur **Environments** yang bisa kamu pakai untuk:
- **Required reviewers** — deploy ke production butuh approval
- **Wait timer** — delay sebelum deploy (misal: 30 menit)
- **Branch restrictions** — hanya branch `main` yang bisa deploy ke production

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: production              # Environment name
      url: https://api.yourdomain.com
    # Kalau environment "production" punya required reviewer,
    # workflow bakal PAUSE dan nunggu approval sebelum lanjut
```

Setup di GitHub:
```
Settings → Environments → New environment
Name: production
→ Add "Required reviewers" (nama reviewer)
→ Add "Wait timer" (misal: 5 minutes)
→ Add "Deployment branches" → only "main"
```

---

## 12.10 Branch Strategy

### Strategy: main + develop + feature branches

```
main (production)
│
├── develop (staging/integration)
│   │
│   ├── feature/add-task-priority
│   ├── feature/user-avatar
│   ├── fix/login-bug
│   └── chore/update-deps
│
└── hotfix/critical-security-fix (langsung ke main)
```

### Flow

```
1. Developer bikin branch dari develop:
   git checkout develop
   git checkout -b feature/add-task-priority

2. Developer kerja di feature branch, push:
   git push origin feature/add-task-priority

3. Developer bikin PR: feature/add-task-priority → develop
   → CI jalan otomatis (lint, test)
   → Status checks harus pass
   → Code review dari teammate
   → Merge ke develop

4. Periodically, develop di-merge ke main:
   PR: develop → main
   → CI jalan
   → CD jalan → deploy ke production

5. Hotfix? Bikin branch dari main, fix, PR ke main langsung
```

### Visual Flow

```
main:       ────●──────────────●──────────●────── (production)
                │              ▲          ▲
                │              │          │
develop:    ────●──●──●──●─────●──●──●────●────── (staging)
                   │  ▲  │           ▲
                   │  │  │           │
feature/a:         ●──●  │           │
                         │           │
feature/b:               ●──●──●─────●

● = commit
→ = merge
```

### Branch Protection Rules

Di GitHub, kamu bisa set **branch protection** supaya ga sembarang orang push ke `main`:

```
Settings → Branches → Add branch protection rule

Branch name pattern: main

✓ Require a pull request before merging
  ✓ Require approvals: 1
  ✓ Dismiss stale pull request approvals when new commits are pushed

✓ Require status checks to pass before merging
  ✓ Require branches to be up to date before merging
  Status checks: "CI" (nama workflow)

✓ Require conversation resolution before merging

✓ Do not allow bypassing the above settings
```

**Artinya:**
- Ga bisa push langsung ke `main` — harus lewat PR
- PR harus di-approve minimal 1 orang
- CI harus pass sebelum bisa merge
- Semua comment/conversation harus resolved

### Protection untuk develop

```
Branch name pattern: develop

✓ Require a pull request before merging
  ✓ Require approvals: 1

✓ Require status checks to pass before merging
  Status checks: "CI"
```

---

## 12.11 Pull Request Workflow

### PR Triggers CI Otomatis

Setiap kali PR dibuat atau di-update, CI workflow jalan otomatis:

```yaml
on:
  pull_request:
    branches: [main, develop]
```

### PR-Specific Workflow

```yaml
# ============================================
# File: .github/workflows/pr.yml
# Pull Request checks
# ============================================

name: PR Check

on:
  pull_request:
    branches: [main, develop]
    types: [opened, synchronize, reopened]

jobs:
  # ─────────────────────────────────────────
  # Lint + Test (sama kayak CI)
  # ─────────────────────────────────────────
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go test ./... -v -race -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out

  # ─────────────────────────────────────────
  # Build check — pastiin Docker image bisa di-build
  # ─────────────────────────────────────────
  build-check:
    name: Build Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image (tanpa push)
        uses: docker/build-push-action@v6
        with:
          context: .
          push: false    # JANGAN push — cuma test build aja
          tags: task-manager:pr-${{ github.event.pull_request.number }}

  # ─────────────────────────────────────────
  # PR Comment — report coverage di PR
  # ─────────────────────────────────────────
  coverage-report:
    name: Coverage Report
    runs-on: ubuntu-latest
    needs: test
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Run tests with coverage
        run: go test ./... -coverprofile=coverage.out
      - name: Generate coverage report
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
          echo "COVERAGE=$COVERAGE" >> $GITHUB_ENV
      - name: Comment PR with coverage
        uses: marocchino/sticky-pull-request-comment@v2
        with:
          message: |
            ## Test Coverage Report

            **Coverage:** ${{ env.COVERAGE }}

            <details>
            <summary>Coverage details</summary>

            ```
            $(go tool cover -func=coverage.out)
            ```

            </details>
```

### Alur PR — Dari Bikin Sampai Merge

```
1. Developer bikin PR
   │
   ▼
2. PR Check workflow OTOMATIS jalan
   ├── Lint ──→ ✓ Pass
   ├── Test ──→ ✓ Pass
   ├── Build Check ──→ ✓ Pass
   └── Coverage Report ──→ Comment di PR
   │
   ▼
3. Status checks muncul di PR page
   "All checks have passed" (hijau) atau
   "Some checks were not successful" (merah)
   │
   ▼
4. Code review oleh reviewer
   ├── Review comments
   ├── Request changes (kalau ada yang perlu diubah)
   └── Approve
   │
   ▼
5. Merge
   ├── Squash and merge (1 commit, clean history)
   ├── Create a merge commit
   └── Rebase and merge
   │
   ▼
6. CI/CD pipeline jalan di target branch
   └── Kalau merge ke main → deploy otomatis
```

### GitHub UI — Apa yang Kamu Lihat

```
┌─────────────────────────────────────────────────┐
│ Pull Request #42: Add task priority feature      │
│                                                  │
│ feature/add-task-priority → develop              │
│                                                  │
│ ┌──────────────────────────────────────────────┐ │
│ │ Checks                                       │ │
│ │                                              │ │
│ │ ✓ Lint          — 15s                        │ │
│ │ ✓ Test          — 32s                        │ │
│ │ ✓ Build Check   — 1m 12s                    │ │
│ │ ✓ Coverage      — Coverage: 78.5%           │ │
│ │                                              │ │
│ │ All checks have passed                       │ │
│ └──────────────────────────────────────────────┘ │
│                                                  │
│ Review: ✓ Approved by @teammate                  │
│                                                  │
│ [Squash and merge]                               │
└─────────────────────────────────────────────────┘
```

---

## 12.12 Matrix Strategy — Test Multi Versi

### Apa Itu Matrix?

Matrix memungkinkan kamu jalanin **satu job di banyak kombinasi** sekaligus. Misalnya: test di Go 1.24 DAN Go 1.25, di Ubuntu DAN macOS.

```yaml
jobs:
  test:
    name: Test (Go ${{ matrix.go-version }}, ${{ matrix.os }})
    runs-on: ${{ matrix.os }}

    strategy:
      matrix:
        go-version: ['1.24', '1.25']
        os: [ubuntu-latest, macos-latest]
      fail-fast: false    # Jangan stop kalau satu kombinasi fail

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go test ./... -v
```

### Apa yang Terjadi

Dari config di atas, GitHub Actions bikin **4 job** (2 Go versions x 2 OS):

```
✓ Test (Go 1.24, ubuntu-latest)    32s
✓ Test (Go 1.25, ubuntu-latest)    28s
✓ Test (Go 1.24, macos-latest)     45s
✓ Test (Go 1.25, macos-latest)     42s
```

Semua jalan **parallel**!

### Kapan Matrix Berguna?

| Scenario | Matrix Config |
|----------|--------------|
| **Library/package** — harus support multiple Go versions | `go-version: ['1.23', '1.24', '1.25']` |
| **CLI tool** — harus jalan di multiple OS | `os: [ubuntu, macos, windows]` |
| **Database compatibility** — test dengan multiple DB versions | `postgres: ['14', '15', '16']` |

### Kapan TIDAK Perlu Matrix?

- **Application/API** (kayak project kita) — biasanya cuma perlu test di 1 Go version dan 1 OS
- Matrix makan GitHub Actions minutes — kalau ga perlu, jangan pakai
- **Rule of thumb:** Pakai matrix untuk **libraries**, jangan untuk **applications**

### Matrix dengan include/exclude

```yaml
strategy:
  matrix:
    go-version: ['1.24', '1.25']
    os: [ubuntu-latest, macos-latest]
    exclude:
      # Skip Go 1.24 di macOS (ga perlu)
      - go-version: '1.24'
        os: macos-latest
    include:
      # Tambah kombinasi khusus
      - go-version: '1.25'
        os: ubuntu-latest
        experimental: true
```

---

## 12.13 Caching — Speed Up CI

### Kenapa Caching Penting?

Tanpa caching:

```
Step: Download dependencies
go mod download
→ Download 50+ packages dari internet
→ 30-60 detik SETIAP kali CI jalan
→ Padahal go.mod ga berubah!
```

Dengan caching:

```
Run pertama:
go mod download → 30 detik (download dari internet)
→ Cache disimpan

Run kedua:
go mod download → 2 detik (dari cache!)
→ 28 detik hemat PER RUN
```

### Cache Go Modules

#### Cara 1: actions/setup-go Built-in Cache (Recommended)

```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.25'
    cache: true    # ← Ini aja! Otomatis cache go modules
```

`actions/setup-go@v5` otomatis cache:
- `~/go/pkg/mod` (downloaded modules)
- `~/.cache/go-build` (build cache)

**Ini cara paling simple.** Cukup tambah `cache: true`.

#### Cara 2: actions/cache Manual

Kalau mau kontrol lebih:

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

**Penjelasan:**
- `path:` — folder apa yang di-cache
- `key:` — cache key. Berdasarkan OS + hash dari `go.sum`
- `restore-keys:` — fallback key kalau exact match ga ketemu
- Kalau `go.sum` berubah (dependency baru) → cache key baru → download ulang
- Kalau `go.sum` ga berubah → cache hit → instant

### Cache Docker Layers

Docker build juga bisa di-cache:

```yaml
- name: Build and push
  uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    # Cache Docker layers di GitHub Actions cache
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

Atau pakai **registry cache** (cache disimpan di registry, bisa share antar branch):

```yaml
- name: Build and push
  uses: docker/build-push-action@v6
  with:
    context: .
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    cache-from: type=registry,ref=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:buildcache
    cache-to: type=registry,ref=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:buildcache,mode=max
```

### Impact Caching

| Step | Tanpa Cache | Dengan Cache | Hemat |
|------|------------|-------------|-------|
| Go mod download | 30s | 2s | 28s |
| Go build | 45s | 10s | 35s |
| Docker build | 120s | 30s | 90s |
| **Total** | **195s** | **42s** | **153s (~78%)** |

**Caching bisa hemat 70-80% waktu CI.** Ini penting banget kalau kamu sering push.

---

## 12.14 Notifications — Slack/Discord Alert

### Kenapa Perlu Notification?

Kamu ga mau harus terus-terusan buka GitHub buat check apakah CI pass atau fail. Lebih baik **dapet alert** di Slack atau Discord kalau ada yang fail.

### Discord Notification

```yaml
# Tambah step ini di akhir workflow
- name: Notify Discord on failure
  if: failure()    # Hanya jalan kalau ada step yang FAIL
  uses: sarisia/actions-status-discord@v1
  with:
    webhook: ${{ secrets.DISCORD_WEBHOOK }}
    status: ${{ job.status }}
    title: "CI Pipeline Failed"
    description: |
      **Repository:** ${{ github.repository }}
      **Branch:** ${{ github.ref_name }}
      **Commit:** ${{ github.sha }}
      **Author:** ${{ github.actor }}
    color: 0xff0000    # Red
```

### Slack Notification

```yaml
- name: Notify Slack on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    fields: repo,message,commit,author,action,eventName,ref,workflow
    mention: 'here'
    if_mention: failure
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Simple Webhook (Universal — Slack/Discord/Telegram)

Kalau mau pakai webhook langsung tanpa action:

```yaml
- name: Send notification
  if: failure()
  run: |
    curl -X POST "${{ secrets.WEBHOOK_URL }}" \
      -H "Content-Type: application/json" \
      -d '{
        "content": "CI FAILED! Repo: ${{ github.repository }}, Branch: ${{ github.ref_name }}, Commit: ${{ github.sha }}, Author: ${{ github.actor }}"
      }'
```

### Setup Discord Webhook

```
1. Buka Discord server
2. Klik channel yang mau dipake → Edit Channel → Integrations
3. Webhooks → New Webhook
4. Copy Webhook URL
5. Simpan di GitHub Secrets: DISCORD_WEBHOOK
```

### Notification Strategy

| Event | Notification? | Channel |
|-------|--------------|---------|
| CI pass | No (too noisy) | - |
| CI fail | Yes | #ci-alerts |
| Deploy success | Yes | #deployments |
| Deploy fail | Yes (urgent!) | #ci-alerts + mention @here |

**Rule:** Jangan notify untuk setiap CI pass. Terlalu banyak notif = orang ignore. **Hanya notify untuk failures**.

---

## 12.15 Full CI/CD Pipeline — Production Ready

Sekarang kita gabungkan semuanya jadi **satu pipeline lengkap**. Ini workflow production-ready yang bisa kamu langsung pakai.

### Complete Workflow

```yaml
# ============================================
# File: .github/workflows/pipeline.yml
# Full CI/CD Pipeline — Task Manager API
# Lint → Test → Build → Push → Deploy
# ============================================

name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'ebook/**'
      - 'LICENSE'
  pull_request:
    branches: [main, develop]
    paths-ignore:
      - '**.md'
      - 'docs/**'
      - 'ebook/**'

env:
  GO_VERSION: '1.25'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # ═══════════════════════════════════════════
  # Stage 1: LINT
  # ═══════════════════════════════════════════
  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m

  # ═══════════════════════════════════════════
  # Stage 2: TEST (parallel with lint)
  # ═══════════════════════════════════════════
  test:
    name: Test
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: task_manager_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U postgres"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        env:
          APP_ENV: test
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: testpassword
          DB_NAME: task_manager_test
          REDIS_HOST: localhost
          REDIS_PORT: 6379
          JWT_SECRET: test-secret-key
        run: |
          go test ./... -v -race \
            -coverprofile=coverage.out \
            -covermode=atomic

      - name: Coverage summary
        run: |
          echo "## Test Coverage" >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY
          go tool cover -func=coverage.out >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out
          retention-days: 7

  # ═══════════════════════════════════════════
  # Stage 3: BUILD & PUSH (setelah lint + test pass)
  # ═══════════════════════════════════════════
  build:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    needs: [lint, test]    # Tunggu lint DAN test selesai

    # HANYA push ke branch main (bukan PR)
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'

    permissions:
      contents: read
      packages: write

    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Docker metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=raw,value=latest
            type=sha,prefix=

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Image info
        run: |
          echo "## Docker Image" >> $GITHUB_STEP_SUMMARY
          echo "Image pushed to:" >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY
          echo "${{ steps.meta.outputs.tags }}" >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY

  # ═══════════════════════════════════════════
  # Stage 4: DEPLOY (setelah build selesai)
  # ═══════════════════════════════════════════
  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: build    # Tunggu build selesai

    # HANYA deploy dari branch main
    if: github.ref == 'refs/heads/main'

    environment:
      name: production
      url: https://api.yourdomain.com

    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_SSH_KEY }}
          port: ${{ secrets.SERVER_PORT }}
          script: |
            cd /home/deploy/task-manager
            docker compose pull
            docker compose up -d --force-recreate --no-deps api
            docker image prune -f
            sleep 5
            curl -f http://localhost:8080/health || exit 1
            echo "Deploy successful at $(date)"

      - name: Deploy summary
        run: |
          echo "## Deployment" >> $GITHUB_STEP_SUMMARY
          echo "- **Environment:** production" >> $GITHUB_STEP_SUMMARY
          echo "- **Commit:** ${{ github.sha }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Author:** ${{ github.actor }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Status:** Deployed" >> $GITHUB_STEP_SUMMARY

  # ═══════════════════════════════════════════
  # NOTIFICATION (kalau ada yang fail)
  # ═══════════════════════════════════════════
  notify:
    name: Notify on Failure
    runs-on: ubuntu-latest
    needs: [lint, test, build, deploy]
    if: failure()    # Hanya jalan kalau ada job yang FAIL

    steps:
      - name: Send Discord notification
        uses: sarisia/actions-status-discord@v1
        with:
          webhook: ${{ secrets.DISCORD_WEBHOOK }}
          status: failure
          title: "Pipeline Failed!"
          description: |
            **Repo:** ${{ github.repository }}
            **Branch:** ${{ github.ref_name }}
            **Commit:** `${{ github.sha }}`
            **Author:** ${{ github.actor }}
            **Workflow:** ${{ github.workflow }}
          color: 0xff0000
```

### Visualisasi Pipeline

```
Push ke main:

┌────────┐      ┌────────┐      ┌────────┐      ┌────────┐
│  Lint  │      │  Test  │      │ Build  │      │ Deploy │
│  15s   │      │  45s   │      │  2m    │      │  30s   │
│        │      │        │      │        │      │        │
│ ✓ Pass ├──┐   │ ✓ Pass ├──┐   │ ✓ Pass ├──────│ ✓ Pass │
└────────┘  │   └────────┘  │   └────────┘      └────────┘
            │               │        ▲
            │  needs:[lint,  │────────┘
            └──test]─────────┘

Total: ~3 menit dari push sampai deploy!

Kalau PR (bukan push ke main):

┌────────┐      ┌────────┐
│  Lint  │      │  Test  │      Build & Deploy SKIP
│  15s   │      │  45s   │      (karena if: github.ref == 'refs/heads/main')
│ ✓ Pass │      │ ✓ Pass │
└────────┘      └────────┘
```

### Job Dependencies (needs)

```yaml
jobs:
  lint:        # Jalan duluan (ga ada needs)
  test:        # Jalan duluan (ga ada needs)
               # lint dan test jalan PARALLEL!

  build:
    needs: [lint, test]    # Tunggu lint DAN test selesai
                           # Kalau salah satu FAIL → build SKIP

  deploy:
    needs: build           # Tunggu build selesai

  notify:
    needs: [lint, test, build, deploy]
    if: failure()          # Jalan kalau salah satu job FAIL
```

### $GITHUB_STEP_SUMMARY

```yaml
echo "## Test Coverage" >> $GITHUB_STEP_SUMMARY
```

Ini fitur keren — kamu bisa nulis **Markdown** ke summary yang muncul di halaman workflow run. Jadi ga perlu scroll logs panjang-panjang.

---

## 12.16 Makefile Integration — Consistency Local dan CI

### Problem: Command Beda di Local vs CI

```
Local:
  golangci-lint run
  go test ./... -v -race -coverprofile=coverage.out
  go build -o bin/api cmd/api/main.go

CI:
  golangci-lint run --timeout=5m
  go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/api cmd/api/main.go
```

Beda command di local dan CI = **resep disaster**. "Kok di local pass tapi di CI fail?"

### Solution: Makefile sebagai Single Source of Truth

Update Makefile kita:

```makefile
# ============================================
# Makefile — Task Manager API
# Single source of truth untuk semua commands
# ============================================

.PHONY: run build test test-race test-coverage lint fmt clean deps docker-build docker-push help

# ──── Variables ────
APP_NAME := task-manager
BINARY := bin/api
MAIN := cmd/api/main.go
COVERAGE_FILE := coverage.out
DOCKER_REGISTRY := ghcr.io
DOCKER_IMAGE := $(DOCKER_REGISTRY)/irzan-aldi/$(APP_NAME)
GIT_SHA := $(shell git rev-parse --short HEAD)

# ──── Development ────

## Run server (development)
run:
	go run $(MAIN)

## Build binary
build:
	go build -o $(BINARY) $(MAIN)

## Build binary (production — static, stripped)
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $(BINARY) $(MAIN)

## Download dependencies
deps:
	go mod tidy
	go mod download

## Format code
fmt:
	go fmt ./...

## Clean build artifacts
clean:
	rm -rf bin/ tmp/ $(COVERAGE_FILE)

# ──── Quality ────

## Run linter
lint:
	golangci-lint run --timeout=5m

## Run tests
test:
	go test ./... -v

## Run tests with race detector
test-race:
	go test ./... -v -race

## Run tests with coverage report
test-coverage:
	go test ./... -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	go tool cover -func=$(COVERAGE_FILE)

## Show coverage in browser
coverage-html:
	go test ./... -coverprofile=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE)

# ──── Docker ────

## Build Docker image
docker-build:
	docker build -t $(APP_NAME):$(GIT_SHA) -t $(APP_NAME):latest .

## Push Docker image to registry
docker-push:
	docker tag $(APP_NAME):latest $(DOCKER_IMAGE):latest
	docker tag $(APP_NAME):$(GIT_SHA) $(DOCKER_IMAGE):$(GIT_SHA)
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(GIT_SHA)

## Start all services (docker compose)
up:
	docker compose up -d --build

## Stop all services
down:
	docker compose down

## Start development environment
dev:
	docker compose -f docker-compose.dev.yml up

# ──── Help ────

## Show this help
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""
	@echo "Usage: make <target>"
```

### Pakai Makefile di CI

Sekarang di workflow, ganti command langsung jadi `make`:

```yaml
steps:
  - name: Checkout code
    uses: actions/checkout@v4

  - name: Setup Go
    uses: actions/setup-go@v5
    with:
      go-version: '1.25'
      cache: true

  - name: Download dependencies
    run: make deps

  - name: Lint
    run: make lint

  - name: Test with coverage
    run: make test-coverage

  - name: Build (production)
    run: make build-prod
```

### Keuntungan

```
Sebelum (tanpa Makefile):
─────────────────────────
Local:  go test ./... -v
CI:     go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
Server: go test ./...
→ 3 tempat beda command. Bug waiting to happen.

Sesudah (dengan Makefile):
──────────────────────────
Local:  make test-coverage
CI:     make test-coverage
Server: make test-coverage
→ 1 command. Hasilnya sama persis di mana-mana.
```

| Benefit | Penjelasan |
|---------|-----------|
| **Consistency** | Command sama di local, CI, server |
| **Single source of truth** | Update Makefile = update semua |
| **Documentation** | `make help` = daftar semua command |
| **Abstraction** | Developer ga perlu ingat flags panjang |
| **DRY** | Ga repeat command di banyak tempat |

---

## 12.17 Common Issues & Debugging

### Problem 1: Workflow Ga Jalan

```
Symptom: Push code tapi workflow ga trigger
```

**Checklist:**
```bash
# 1. File ada di tempat yang benar?
.github/workflows/ci.yml    # HARUS di .github/workflows/

# 2. YAML syntax valid?
# Install yamllint atau pakai online validator
yamllint .github/workflows/ci.yml

# 3. Branch match?
on:
  push:
    branches: [main]    # Kamu push ke "main" atau "master"?

# 4. Path filter exclude file kamu?
paths-ignore:
  - '**.md'             # Kalau kamu cuma update .md, workflow skip

# 5. Workflow disabled?
# Check di GitHub: Actions tab → workflow → ... → Enable workflow
```

### Problem 2: Step Fail — Permission Denied

```
Error: permission denied
```

**Solution:**

```yaml
# Tambah permissions di job
jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write    # Perlu ini untuk push ke GHCR
```

### Problem 3: Docker Build Fail di CI

```
Error: failed to fetch anonymous token: unexpected status: 403 Forbidden
```

**Solution:**

```yaml
# Pastiin login dulu sebelum build
- name: Login to GHCR
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### Problem 4: Service Container Ga Ready

```
Error: connection refused (postgres)
```

**Solution:**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    # HARUS punya health check options
    options: >-
      --health-cmd "pg_isready -U postgres"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

### Problem 5: Secrets Ga Ke-set

```
Error: secret 'SERVER_HOST' not found
```

**Checklist:**
1. Udah tambah secret di GitHub? (Settings → Secrets → Actions)
2. Nama persis sama? (case sensitive!)
3. Kalau di fork, secrets dari repo original **ga bisa diakses**

### Problem 6: Cache Ga Kerja

```
Cache not found for keys: Linux-go-abc123...
```

**Solution:**

```yaml
# Pastiin key-nya consistent
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    # hashFiles harus match path yang benar
```

### Tool: act — Run GitHub Actions Locally

`act` itu tool yang memungkinkan kamu **jalanin GitHub Actions di laptop** tanpa push ke GitHub. Super useful buat debugging.

```bash
# Install act
brew install act

# Run semua workflows
act

# Run specific event
act push

# Run specific job
act -j test

# Run dengan secrets
act -s MY_SECRET=value

# List workflows
act -l

# Dry run (lihat apa yang bakal dijalanin tanpa benar-benar jalanin)
act -n
```

**Limitation act:**
- Service containers ga fully supported
- Beberapa GitHub context ga tersedia
- Performance beda sama GitHub runners
- Tapi tetap **sangat berguna** untuk debug YAML syntax dan step logic

### Debugging Tips

```yaml
# 1. Enable debug logging
# Di GitHub: Re-run jobs → "Enable debug logging"
# Atau set secret: ACTIONS_RUNNER_DEBUG = true

# 2. Print environment info
- name: Debug info
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Ref: ${{ github.ref }}"
    echo "SHA: ${{ github.sha }}"
    echo "Actor: ${{ github.actor }}"
    echo "Runner OS: ${{ runner.os }}"
    go version
    go env

# 3. Print file structure
- name: Show project files
  run: |
    ls -la
    find . -name "*.go" | head -20

# 4. Conditional step for debugging
- name: Debug step
  if: runner.debug == '1'    # Hanya jalan kalau debug mode ON
  run: |
    env | sort
    cat go.mod
```

### Re-run Failed Workflows

```
Di GitHub:
1. Buka Actions tab
2. Klik workflow run yang fail
3. Klik "Re-run all jobs" atau "Re-run failed jobs"
4. Centang "Enable debug logging" kalau mau lebih detail
```

---

## 12.18 Best Practices

### 1. Keep CI Fast (< 5 Menit)

```
IDEAL:
Lint:    15s
Test:    30s
Build:   90s
Deploy:  30s
Total:   ~3 menit

TERLALU LAMA:
Total:   > 10 menit
→ Developer bosen nunggu
→ Feedback loop terlalu lambat
→ PR queue menumpuk
```

**Cara bikin CI cepat:**
- Cache dependencies dan build artifacts
- Run lint dan test **parallel** (separate jobs tanpa `needs`)
- Skip workflow kalau cuma `.md` yang berubah (`paths-ignore`)
- Pakai Go test cache (`-count=1` hanya kalau perlu)
- Jangan matrix kalau ga perlu

### 2. Fail Fast

```yaml
strategy:
  fail-fast: true    # Default. Stop semua job kalau 1 fail.
```

Jangan buang waktu jalanin 10 job lagi kalau job pertama udah fail.

### 3. Cache Everything

```yaml
# Go modules
- uses: actions/setup-go@v5
  with:
    go-version: '1.25'
    cache: true

# Docker layers
- uses: docker/build-push-action@v6
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### 4. Separate CI dan CD

```
Option A: 1 file (simple, cocok buat project kecil)
.github/workflows/pipeline.yml → semua di satu file

Option B: Separate files (recommended untuk production)
.github/workflows/ci.yml      → lint + test (setiap push/PR)
.github/workflows/cd.yml      → build + deploy (cuma main)
```

Separate lebih bagus karena:
- CI jalan di semua branch/PR, CD cuma di main
- Bisa re-run CD tanpa re-run CI
- Permissions lebih granular

### 5. Pin Action Versions

```yaml
# JANGAN — bisa break kapan aja
- uses: actions/checkout@main
- uses: actions/setup-go@latest

# LAKUKAN — predictable
- uses: actions/checkout@v4
- uses: actions/setup-go@v5

# PALING AMAN — pin ke commit SHA
- uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
```

**Kenapa pin?**
- `@main` bisa berubah kapan aja → workflow bisa break
- `@v4` lebih stable, tapi major version bisa ada breaking change
- `@sha` paling aman — immutable, ga bisa berubah

### 6. Minimal Permissions

```yaml
# JANGAN — terlalu banyak permission
permissions: write-all

# LAKUKAN — hanya yang dibutuhkan
permissions:
  contents: read
  packages: write
```

Principle of least privilege. Kalau token bocor, damage-nya minimal.

### 7. Jangan Hardcode Secrets

```yaml
# JANGAN
- run: echo "password123" | docker login --password-stdin

# LAKUKAN
- run: echo "${{ secrets.DOCKER_TOKEN }}" | docker login --password-stdin
```

### 8. Use Timeouts

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    timeout-minutes: 10    # Kill job kalau > 10 menit
    steps:
      - name: Run tests
        timeout-minutes: 5  # Kill step kalau > 5 menit
        run: go test ./...
```

Tanpa timeout, job yang hang bisa **jalan 6 jam** (default GitHub limit) dan makan semua minutes kamu.

### 9. Gunakan Job Summaries

```yaml
- name: Summary
  run: |
    echo "## Build Summary" >> $GITHUB_STEP_SUMMARY
    echo "- Image: task-manager:${{ github.sha }}" >> $GITHUB_STEP_SUMMARY
    echo "- Size: $(docker images task-manager --format '{{.Size}}')" >> $GITHUB_STEP_SUMMARY
```

### 10. Branch Protection + Required Checks

```
Settings → Branches → Branch protection rules:

✓ Require status checks: "CI"
✓ Require PR before merge
✓ Require approvals: 1

Ini "enforce" CI. Ga ada bypass.
```

### Rangkuman Best Practices

| Practice | Kenapa |
|----------|--------|
| CI < 5 menit | Fast feedback loop |
| Cache everything | Hemat waktu dan money |
| Fail fast | Jangan buang resource |
| Pin action versions | Predictable builds |
| Minimal permissions | Security |
| Separate CI/CD | Flexibility |
| Timeouts | Prevent runaway jobs |
| Path filtering | Skip unnecessary runs |
| Makefile integration | Consistency |
| Branch protection | Enforce quality gate |

---

## 12.19 Cheat Sheet — GitHub Actions

### Workflow Structure

```yaml
name: Workflow Name                    # Nama workflow

on:                                    # Triggers
  push:
    branches: [main]
    paths-ignore: ['**.md']
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * *'
  workflow_dispatch:                   # Manual trigger
    inputs:
      param:
        description: 'Parameter'
        required: true
        type: string

env:                                   # Global env vars
  MY_VAR: 'value'

permissions:                           # Token permissions
  contents: read
  packages: write

jobs:
  job-name:                            # Job
    name: Display Name
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [other-job]                 # Dependencies
    if: github.ref == 'refs/heads/main'  # Conditional

    services:                          # Service containers
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432

    env:                               # Job-level env vars
      JOB_VAR: 'value'

    steps:
      - name: Step name
        uses: actions/checkout@v4      # Use an Action
        with:
          fetch-depth: 0               # Action parameters

      - name: Run command
        run: echo "Hello"              # Run shell command
        env:
          STEP_VAR: 'value'            # Step-level env vars

      - name: Multi-line command
        run: |
          echo "Line 1"
          echo "Line 2"

      - name: Conditional step
        if: success()                  # Conditions
        run: echo "Previous step passed"

    outputs:                           # Job outputs
      result: ${{ steps.step-id.outputs.value }}
```

### Expressions & Contexts

```yaml
# ──── Contexts ────
${{ github.repository }}        # "owner/repo"
${{ github.ref }}               # "refs/heads/main"
${{ github.ref_name }}          # "main"
${{ github.sha }}               # Full commit SHA
${{ github.actor }}             # Username yang trigger
${{ github.event_name }}        # "push", "pull_request", etc
${{ github.workspace }}         # Workspace path
${{ github.run_id }}            # Unique run ID
${{ github.run_number }}        # Run number (incrementing)

# ──── Secrets ────
${{ secrets.MY_SECRET }}        # Repository secret
${{ secrets.GITHUB_TOKEN }}     # Auto-generated token

# ──── Environment ────
${{ env.MY_VAR }}               # Environment variable
${{ runner.os }}                # "Linux", "macOS", "Windows"

# ──── Job outputs ────
${{ needs.job-name.outputs.result }}   # Output dari job lain

# ──── Step outputs ────
${{ steps.step-id.outputs.value }}     # Output dari step

# ──── Conditions ────
${{ success() }}                # Previous steps succeeded
${{ failure() }}                # Any previous step failed
${{ always() }}                 # Always run
${{ cancelled() }}              # Workflow cancelled
```

### Common Actions

```yaml
# ──── Core ────
actions/checkout@v4                    # Checkout repo
actions/setup-go@v5                    # Install Go
actions/cache@v4                       # Cache files
actions/upload-artifact@v4             # Upload artifact
actions/download-artifact@v4           # Download artifact

# ──── Docker ────
docker/setup-buildx-action@v3         # Docker Buildx
docker/login-action@v3                # Login to registry
docker/metadata-action@v5             # Generate tags
docker/build-push-action@v6           # Build & push

# ──── Quality ────
golangci/golangci-lint-action@v6      # Go linter

# ──── Deploy ────
appleboy/ssh-action@v1                # SSH to server

# ──── Notification ────
sarisia/actions-status-discord@v1     # Discord notification
8398a7/action-slack@v3                # Slack notification
```

### Conditional Execution

```yaml
# Hanya di branch main
if: github.ref == 'refs/heads/main'

# Hanya untuk push (bukan PR)
if: github.event_name == 'push'

# Hanya kalau step sebelumnya fail
if: failure()

# Selalu jalan (cleanup, notification)
if: always()

# Kombinasi
if: github.event_name == 'push' && github.ref == 'refs/heads/main'

# Negation
if: github.event_name != 'pull_request'

# Contains
if: contains(github.event.head_commit.message, '[skip ci]')
```

### Service Containers

```yaml
services:
  postgres:
    image: postgres:16-alpine
    env:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: testpassword
      POSTGRES_DB: testdb
    ports:
      - 5432:5432
    options: >-
      --health-cmd "pg_isready -U postgres"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

  redis:
    image: redis:7-alpine
    ports:
      - 6379:6379
    options: >-
      --health-cmd "redis-cli ping"
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

### Caching Patterns

```yaml
# Go — Simple (recommended)
- uses: actions/setup-go@v5
  with:
    go-version: '1.25'
    cache: true

# Go — Manual
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: ${{ runner.os }}-go-

# Docker — GHA cache
- uses: docker/build-push-action@v6
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Job Outputs

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.version.outputs.version }}
    steps:
      - name: Get version
        id: version
        run: echo "version=1.2.3" >> $GITHUB_OUTPUT

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying version ${{ needs.build.outputs.version }}"
```

### Quick Reference: File Structure

```
.github/
└── workflows/
    ├── ci.yml              # CI — lint + test (setiap push/PR)
    ├── cd.yml              # CD — build + deploy (main only)
    ├── pr.yml              # PR checks
    └── release.yml         # Release workflow
```

---

## 12.20 Checklist Bab Ini

### Konsep CI/CD

- [ ] Paham apa itu CI (Continuous Integration)
- [ ] Paham perbedaan CD Delivery vs CD Deployment
- [ ] Paham kenapa CI/CD penting (vs manual deploy)
- [ ] Paham analogi assembly line

### GitHub Actions Basics

- [ ] Paham 6 konsep: Workflow, Job, Step, Runner, Event, Action
- [ ] Paham YAML syntax dasar
- [ ] Tau lokasi workflow file: `.github/workflows/`
- [ ] Paham event triggers (push, pull_request, schedule, workflow_dispatch)
- [ ] Paham path filtering (paths-ignore)

### CI Pipeline

- [ ] Bisa bikin workflow lint (golangci-lint)
- [ ] Bisa bikin workflow test (go test dengan coverage)
- [ ] Bisa setup service containers (PostgreSQL + Redis)
- [ ] Paham perbedaan unit test vs integration test di CI

### Docker Build & Push

- [ ] Bisa build Docker image di CI
- [ ] Bisa push ke GHCR (GitHub Container Registry)
- [ ] Paham Docker metadata (tags: latest, SHA, semver)
- [ ] Paham Docker layer caching di CI

### Deploy

- [ ] Paham deploy options (SSH, ECS, Kubernetes)
- [ ] Bisa bikin SSH deploy workflow
- [ ] Paham environment protection rules
- [ ] Paham conditional deployment (hanya branch main)

### Secrets & Security

- [ ] Bisa setup GitHub Secrets
- [ ] Paham GITHUB_TOKEN (otomatis)
- [ ] Paham perbedaan env vars vs secrets
- [ ] Paham minimal permissions (principle of least privilege)

### Branch Strategy

- [ ] Paham main + develop + feature branch strategy
- [ ] Bisa setup branch protection rules
- [ ] Paham PR workflow (create → CI → review → merge)

### Advanced

- [ ] Paham matrix strategy (kapan pakai, kapan tidak)
- [ ] Bisa setup caching (Go modules, Docker layers)
- [ ] Bisa setup notifications (Discord/Slack)
- [ ] Paham Makefile integration (consistency local dan CI)

### Debugging

- [ ] Tau cara debug workflow yang fail
- [ ] Tau tentang `act` (run locally)
- [ ] Bisa re-run failed workflows
- [ ] Bisa enable debug logging

### Full Pipeline

- [ ] Punya workflow file lengkap: lint → test → build → push → deploy
- [ ] Paham job dependencies (`needs`)
- [ ] Paham conditional execution (`if`)
- [ ] Workflow jalan di bawah 5 menit

### Best Practices

- [ ] CI < 5 menit
- [ ] Cache everything
- [ ] Pin action versions
- [ ] Minimal permissions
- [ ] Path filtering
- [ ] Timeout di semua jobs
- [ ] Separate CI dan CD (atau tau kapan gabung)

---

> **Next:** [Bab 13: Cloud Deployment — AWS](13-cloud-aws.md) — kita bakal deploy Task Manager API ke AWS. EC2, RDS, ElastiCache, load balancer — infrastruktur production beneran. Setelah CI/CD pipeline udah jalan, sekarang saatnya punya server yang beneran di cloud.