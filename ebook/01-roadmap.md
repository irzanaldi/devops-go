# Bab 1: Roadmap Fullstack → DevOps + Senior Backend

## Kenapa Dua-Duanya?

DevOps dan Senior Backend saling berkaitan. Senior Backend yang paham DevOps bisa:
- Deploy sendiri tanpa tunggu tim infra
- Design sistem yang scalable dari awal
- Debug production issue lebih cepat
- Jadi lebih valuable di job market

---

## Phase 1: Senior Backend Foundation (Bulan 1-3)

| Topic | Kenapa Penting |
|-------|---------------|
| **System Design** | Interview senior selalu tanya ini. Load balancer, caching, message queue, database scaling |
| **API Design** | REST best practices, GraphQL, gRPC, versioning, rate limiting |
| **Database Advanced** | Indexing strategy, query optimization, replication, sharding |
| **Caching** | Redis deep dive — session, cache-aside, pub/sub, rate limiter |
| **Message Queue** | RabbitMQ atau Kafka — async processing, event-driven architecture |
| **Authentication** | OAuth2, JWT deep dive, SSO, RBAC vs ABAC |
| **Testing** | Unit, integration, e2e — coverage strategy |

---

## Phase 2: DevOps Core (Bulan 3-6)

| Topic | Kenapa Penting |
|-------|---------------|
| **Linux & Networking** | Dasar semua DevOps. TCP/IP, DNS, firewall, SSH, systemd |
| **Docker** | Containerize app, multi-stage build, Docker Compose |
| **CI/CD** | GitHub Actions atau GitLab CI — automate test, build, deploy |
| **Cloud** | AWS (paling banyak demand). EC2, S3, RDS, VPC |
| **Nginx/Reverse Proxy** | Config, SSL termination, load balancing |
| **Monitoring** | Prometheus + Grafana, log management (ELK stack) |

---

## Phase 3: Advanced (Bulan 6-12)

| Topic | Kenapa Penting |
|-------|---------------|
| **Kubernetes** | Orchestration standard industri |
| **Infrastructure as Code** | Terraform — define infra as code, reproducible |
| **Security** | OWASP, secret management (Vault), network security |
| **Microservices** | Pattern: saga, circuit breaker, service mesh |
| **Performance** | Profiling, load testing (k6/Artillery), optimization |

---

## Tips Belajar

1. **Hands-on > teori** — setiap topic, langsung bikin project kecil
2. **Satu project end-to-end** — bikin backend app → dockerize → CI/CD → deploy ke AWS → monitoring
3. **Sertifikasi** — AWS SAA (Solutions Architect Associate) boost CV
4. **Open source** — contribute ke project ecosystem
