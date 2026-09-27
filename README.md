# Devops Go — Task Manager API

Task Manager API sederhana yang dibangun dengan Go. Ini adalah project belajar (learning project) yang mengikuti sebuah ebook DevOps/Backend, memakai Gin (HTTP), GORM + PostgreSQL (database), JWT (autentikasi), dan Redis (caching). Materi ebook ada di folder `ebook/`.

## Cara Setup

### Prasyarat

- **Go** 1.25.0 atau lebih baru (lihat `go.mod`)
- **PostgreSQL** — wajib. Dipakai sebagai database utama via GORM. Tabel dibuat otomatis lewat `AutoMigrate` saat server start, jadi cukup sediakan database kosong (default nama: `task_manager`).
- **Redis** — opsional. Dipakai untuk caching. Kalau Redis tidak tersedia, server tetap jalan tapi caching dinonaktifkan (lihat log peringatan).
- **make** — opsional, untuk memakai target di `Makefile`.

### 1. Clone repository

```bash
git clone <url-repo>
cd "Devops Go"   # nama folder mengandung spasi, jadi pakai tanda kutip
```

### 2. Siapkan environment variable

Konfigurasi dibaca dari environment variable (lihat `internal/config/config.go`). Buat file `.env` di root project. File `.env` sudah di-ignore git dan tidak akan ikut ter-commit.

Contoh isi `.env`:

```env
APP_ENV=development
PORT=8080
JWT_SECRET=ganti-dengan-secret-yang-aman

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=task_manager

# Redis (opsional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

> Semua variable punya nilai default di kode, jadi aplikasi tetap bisa jalan tanpa `.env` selama PostgreSQL bisa diakses di `localhost:5432` (user `postgres`, database `task_manager`).

### 3. Install dependency

```bash
go mod download
# atau
make deps        # menjalankan `go mod tidy`
```

### 4. Menjalankan server

```bash
make run
# atau tanpa make:
go run ./cmd/api
```

Server berjalan di `http://localhost:8080` (atau sesuai `PORT`).

### 5. Build binary

```bash
make build       # menghasilkan binary di ./bin/api
```

### 6. Menjalankan test

```bash
make test
# atau
go test ./...
```
