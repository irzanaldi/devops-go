# Bab 2: Strategi Belajar — Project End-to-End

## Konsep

Bukan belajar teori satu-satu, tapi langsung **build → deploy → monitor** dalam 1 project nyata.

Project yang kita bangun: **Task Management API** (mirip Trello simplified).

Kenapa project ini?
- Cukup simpel buat fokus ke infra
- Cukup kompleks buat latihan backend advanced
- Cover semua skill yang dibutuhkan

---

## Stage-by-Stage

### Stage 1 — Backend Solid (Minggu 1-2)
**Stack: Go (Gin) + PostgreSQL + Redis**

- REST API: CRUD tasks, boards, users
- Auth: JWT + refresh token + RBAC (admin/member)
- Redis: caching + rate limiting
- Asynq (Redis-based queue): email notification async
- Unit test + integration test
- API documentation pakai Swagger

**Output:** API jalan di localhost, test coverage > 80%

### Stage 2 — Dockerize (Minggu 3)

- Dockerfile multi-stage build (dev vs prod)
- Docker Compose: app + postgres + redis
- Environment variable management
- Health check endpoint

**Output:** `docker compose up` → semua jalan

### Stage 3 — CI/CD Pipeline (Minggu 4)

- GitHub Actions:
  - PR → auto run test + lint
  - Merge main → build Docker image → push ke Docker Hub/ECR
  - Auto deploy ke server
- Branch strategy: main + develop + feature branches

**Output:** Push code → otomatis test → otomatis deploy

### Stage 4 — Deploy ke Cloud (Minggu 5-6)

- AWS Free Tier:
  - EC2 buat app
  - RDS buat PostgreSQL
  - ElastiCache buat Redis
  - S3 buat file storage
- Nginx reverse proxy + SSL (Let's Encrypt)
- Setup VPC, security group, IAM basics

**Output:** API live di production

### Stage 5 — Monitoring & Logging (Minggu 7)

- Prometheus: collect metrics
- Grafana: dashboard visualisasi
- Logger (zerolog) → ship ke CloudWatch atau ELK
- Alert ke Telegram/Discord kalau server down

**Output:** Dashboard monitoring live, alert jalan

### Stage 6 — Level Up (Minggu 8+)

- Terraform: tulis ulang semua infra AWS jadi code
- Kubernetes: migrate dari EC2 ke EKS/k3s
- Load testing pakai k6
- Add microservice (notification service terpisah)

---

## Cara Belajar Tiap Stage

```
1. Baca docs/tutorial 30 menit → pahami konsep
2. Langsung kerjain di project → stuck itu normal
3. Stuck > 30 menit → cari solusi (docs, stackoverflow, AI)
4. Selesai 1 stage → tulis catatan di e-book ini
5. Commit semua ke GitHub → ini jadi portfolio
```

---

## Timeline

| Kalau | Timeline |
|-------|----------|
| Full-time belajar | 3 bulan |
| Part-time (2-3 jam/hari) | 4-5 bulan |
| Weekend warrior | 6-7 bulan |
