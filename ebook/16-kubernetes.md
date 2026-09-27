# Bab 16: Kubernetes — Container Orchestration dari Nol Sampai Production

## Kenapa Bab Ini Penting?

Selamat datang di **Stage 3: Advanced**. Ini bab pertama setelah kamu menyelesaikan DevOps Core (Docker, CI/CD, AWS, Nginx, Monitoring). Dan bab ini mungkin **bab terpenting di seluruh ebook** untuk karier DevOps kamu.

Di Bab 11, kamu belajar Docker — containerize satu app. Di Bab 12, kamu setup CI/CD — auto build dan deploy. Tapi sekarang bayangin ini:

```
Skenario: Startup kamu mulai grow
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Bulan 1: 1 server, 1 container Go app. Aman.
Bulan 3: Traffic naik. 1 server ga cukup. Tambah server? Gimana cara distribute traffic?
Bulan 6: 5 microservices. Setiap service butuh di-deploy, di-scale, di-monitor.
         Satu service crash jam 3 pagi. Siapa yang restart?
Bulan 12: 20+ services. 10 server. Rolling update tanpa downtime.
          Service A perlu connect ke Service B — gimana cara discover IP-nya?

Pertanyaan yang muncul:
- Gimana auto-restart container yang crash?
- Gimana scale dari 1 instance ke 10 instance otomatis?
- Gimana deploy version baru TANPA downtime?
- Gimana distribute traffic ke multiple instances?
- Gimana manage config dan secrets di banyak services?
- Gimana schedule container ke server yang masih ada resource?

Jawaban: KUBERNETES.
```

```
Sebelum bab ini:
- Docker Compose bisa jalanin multi-container di 1 server
- Deploy manual ke 1 EC2 instance
- Scale = tambah server manual, setup manual
- Container crash = harus restart manual
- Update = downtime (stop old, start new)
- Service discovery = hardcode IP

Setelah bab ini:
- Kubernetes orchestrate container di multi-server cluster
- Deploy declarative — define desired state, K8s yang eksekusi
- Auto-scaling berdasarkan CPU/memory
- Self-healing — container crash, otomatis restart
- Rolling update — zero-downtime deployment
- Service discovery by DNS name
- Full Task Manager API deployed on Kubernetes
```

> **Analogi gampang:** Docker itu kayak **kontainer kapal** — kamu bisa pack barang ke dalam kontainer standar. Tapi Kubernetes itu **pelabuhan** yang atur SEMUA kontainer. Pelabuhan yang tau kontainer mana harus ditaruh di kapal mana (scheduling), berapa banyak kontainer yang harus disiapkan (scaling), kalau ada kontainer rusak otomatis ganti (self-healing), dan kalau ada kontainer baru datang, yang lama dikeluarin pelan-pelan supaya ga ganggu operasi (rolling update). Docker = kontainer. Kubernetes = pelabuhan.

---

## 16.1 Apa Itu Kubernetes?

### Kubernetes in a Nutshell

Kubernetes (sering disingkat **K8s** — K, 8 huruf, s) itu **container orchestration platform**. Dibuat oleh Google, open-sourced tahun 2014, sekarang dikelola oleh **CNCF** (Cloud Native Computing Foundation).

```
Kubernetes Facts:
━━━━━━━━━━━━━━━━
- Nama: Kubernetes (bahasa Yunani, artinya "juru mudi")
- Singkatan: K8s (K + 8 huruf + s)
- Dibuat: Google, berdasarkan 15 tahun experience running Borg
- Open-sourced: 2014
- CNCF graduated: 2018
- Bahasa: Go (tentu saja!)
- Dipakai: Google, Spotify, Airbnb, Uber, Pinterest, semua cloud provider
- GitHub stars: 100,000+
```

### Problem yang K8s Solve

Kalau kamu udah punya Docker Compose, kenapa perlu Kubernetes? Karena Docker Compose itu kayak **mengelola 1 toko** — simple, kamu bisa handle sendiri. Kubernetes itu kayak **mengelola 100 cabang toko** — butuh system yang otomatis.

```
Problem 1: SCALING
━━━━━━━━━━━━━━━━━
Docker Compose: Manual. "Traffic naik? Hmm, sewa server baru, setup Docker,
                copy compose file, jalanin. 30 menit."
Kubernetes:     Automatic. "Traffic naik? K8s nambah instance dari 3 ke 10
                dalam 30 DETIK."

Problem 2: SELF-HEALING
━━━━━━━━━━━━━━━━━━━━━━
Docker Compose: Container crash? Kamu harus tau dulu, terus restart manual.
                Kalau crash jam 3 pagi? Besok pagi baru tau.
Kubernetes:     Container crash? K8s detect dalam DETIK, restart otomatis.
                Kamu bahkan ga perlu tau. K8s yang handle.

Problem 3: ROLLING UPDATE
━━━━━━━━━━━━━━━━━━━━━━━━
Docker Compose: "docker compose down" → deploy baru → "docker compose up"
                = DOWNTIME. User kena error beberapa detik/menit.
Kubernetes:     Deploy versi baru satu-per-satu. Yang lama masih serve traffic
                sampai yang baru ready. ZERO DOWNTIME.

Problem 4: SERVICE DISCOVERY
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Docker Compose: Service name = container name. Single host. Simple.
Kubernetes:     Service name = DNS entry. Multi-host. Auto-update.
                Service A pindah server? DNS tetap sama. No manual config.

Problem 5: MULTI-HOST
━━━━━━━━━━━━━━━━━━━━
Docker Compose: 1 server. Kalau server mati, SEMUA mati.
Kubernetes:     Multi-server cluster. 1 server mati? Container otomatis
                dipindah ke server lain. High availability.
```

### Fun Fact

Google menjalankan **SEMUA** — Gmail, YouTube, Search, Maps, Drive — di container. Mereka punya internal system bernama **Borg** yang udah dipake sejak 2003. Kubernetes itu versi open-source-nya Borg. Jadi K8s itu bukan framework baru yang unproven — ini based on **15+ tahun** experience Google running billions of containers.

---

## 16.2 Kubernetes vs Docker Compose

### Comparison Table

| Aspek | Docker Compose | Kubernetes |
|-------|---------------|------------|
| **Scope** | Single host (1 server) | Multi-host cluster |
| **Scaling** | Manual (`scale: 3`) | Auto (HPA, berdasarkan CPU/memory) |
| **Self-healing** | `restart: always` (basic) | Liveness/readiness probes, auto-reschedule |
| **Rolling update** | Stop → start (downtime) | Zero-downtime rolling update |
| **Service discovery** | Container name | DNS-based, automatic |
| **Load balancing** | Ga ada built-in | Built-in Service load balancer |
| **Config management** | `.env` file | ConfigMap + Secret (encrypted) |
| **Storage** | Docker volumes | PersistentVolume, StorageClass (dynamic) |
| **Networking** | Bridge network | CNI plugins, NetworkPolicy, Ingress |
| **Monitoring** | Manual setup | Built-in metrics, ready for Prometheus |
| **Complexity** | Low | High (tapi worth it) |
| **Learning curve** | 1 hari | 1-2 minggu |
| **Use case** | Dev, small projects | Production, scaling, microservices |

### When to Use Which?

```
Pakai Docker Compose kalau:
━━━━━━━━━━━━━━━━━━━━━━━━━━
- Development environment (local laptop)
- Side project yang traffic-nya kecil
- Single server deployment
- Tim kecil (1-3 developer)
- Mau simple dan cepat
- Budget terbatas (1 server aja cukup)

Pakai Kubernetes kalau:
━━━━━━━━━━━━━━━━━━━━━━━
- Production yang butuh high availability
- Traffic besar atau unpredictable (perlu auto-scaling)
- Multiple microservices
- Tim mulai grow (5+ developer)
- Zero-downtime deployment wajib
- Multi-server atau multi-region
- Perusahaan yang serius soal infrastructure

Real-world:
━━━━━━━━━━━
- Startup awal → Docker Compose (cukup!)
- Startup growing → K8s atau managed K8s (EKS/GKE)
- Enterprise → Kubernetes (wajib)
- Side project → Docker Compose
- Belajar → Docker Compose dulu, terus K8s
```

### Analogi

```
Docker Compose = Masak di dapur rumah sendiri
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- 1 kompor, 1 oven, 1 kulkas
- Kamu yang masak semua sendiri
- Kalau tamu banyak? Kewalahan.
- Kalau kompor rusak? Ga bisa masak sama sekali.

Kubernetes = Restaurant kitchen (dapur restoran)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- 10 kompor, 5 oven, 3 kulkas
- Head chef (control plane) koordinasi semua
- Line cook (worker nodes) yang masak
- Tamu rame? Tambah line cook.
- 1 kompor rusak? Pindah ke kompor lain, tamu ga tau.
- Menu baru? Gradually replace dish, ga perlu tutup restoran.
```

---

## 16.3 Kubernetes Architecture

### Overview

Kubernetes cluster terdiri dari 2 komponen utama: **Control Plane** (otak) dan **Worker Nodes** (tangan).

```
┌─────────────────────────────────────────────────────────────────┐
│                    KUBERNETES CLUSTER                             │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   CONTROL PLANE (Master)                  │    │
│  │                                                           │    │
│  │  ┌──────────────┐  ┌──────────┐  ┌──────────────────┐  │    │
│  │  │  API Server   │  │   etcd    │  │    Scheduler     │  │    │
│  │  │  (Frontend)   │  │  (Store)  │  │  (Penempatan)    │  │    │
│  │  └──────┬───────┘  └──────────┘  └──────────────────┘  │    │
│  │         │                                                │    │
│  │  ┌──────┴───────────────────────────────────────────┐   │    │
│  │  │            Controller Manager                      │   │    │
│  │  │  (Pastikan desired state = actual state)          │   │    │
│  │  └──────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│         │                    │                    │               │
│         ▼                    ▼                    ▼               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │ WORKER NODE 1│    │ WORKER NODE 2│    │ WORKER NODE 3│         │
│  │              │    │              │    │              │         │
│  │ ┌──────────┐│    │ ┌──────────┐│    │ ┌──────────┐│         │
│  │ │ kubelet  ││    │ │ kubelet  ││    │ │ kubelet  ││         │
│  │ └──────────┘│    │ └──────────┘│    │ └──────────┘│         │
│  │ ┌──────────┐│    │ ┌──────────┐│    │ ┌──────────┐│         │
│  │ │kube-proxy││    │ │kube-proxy││    │ │kube-proxy││         │
│  │ └──────────┘│    │ └──────────┘│    │ └──────────┘│         │
│  │ ┌──────────┐│    │ ┌──────────┐│    │ ┌──────────┐│         │
│  │ │Container ││    │ │Container ││    │ │Container ││         │
│  │ │ Runtime  ││    │ │ Runtime  ││    │ │ Runtime  ││         │
│  │ └──────────┘│    │ └──────────┘│    │ └──────────┘│         │
│  │              │    │              │    │              │         │
│  │ ┌─┐┌─┐┌─┐  │    │ ┌─┐┌─┐     │    │ ┌─┐┌─┐┌─┐  │         │
│  │ │P││P││P│  │    │ │P││P│     │    │ │P││P││P│  │         │
│  │ └─┘└─┘└─┘  │    │ └─┘└─┘     │    │ └─┘└─┘└─┘  │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│   P = Pod (container)                                            │
└─────────────────────────────────────────────────────────────────┘
```

### Control Plane — "Otak" Cluster

Control Plane itu decision-maker. Dia yang putuskan pod jalan di mana, kapan scale, kapan restart.

#### 1. API Server (kube-apiserver)

**Frontend** dari Kubernetes. Semua komunikasi lewat API Server — kubectl, dashboard, internal components, semuanya ngomong ke API Server.

```
API Server = Resepsionis hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Semua tamu (kubectl, dashboard) harus lewat resepsionis
- Resepsionis validasi identitas (authentication)
- Resepsionis cek apakah boleh (authorization)
- Resepsionis forward request ke department yang tepat
```

#### 2. etcd

**Key-value store** yang simpan SEMUA data cluster. State, config, secrets — semua ada di etcd. Ini "database"-nya Kubernetes.

```
etcd = Buku catatan hotel
━━━━━━━━━━━━━━━━━━━━━━━
- Semua informasi tamu (pod, service, config) dicatat di sini
- Kalau buku ini hilang, hotel ga tau apa-apa
- Makanya harus di-backup! (critical component)
- Distributed — multiple copies untuk reliability
```

#### 3. Scheduler (kube-scheduler)

**Decision-maker** yang putuskan pod baru harus jalan di worker node mana. Dia lihat resource yang tersedia, constraint, affinity rules.

```
Scheduler = Manager hotel yang assign kamar
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- "Tamu baru butuh kamar 2 bed, ada AC" (pod butuh 2 CPU, 4GB RAM)
- "Kamar 301 available, cocok" (Node 3 punya resource cukup)
- "Assign tamu ke kamar 301" (schedule pod ke Node 3)
```

#### 4. Controller Manager (kube-controller-manager)

**Loop manager** yang terus-terusan cek: "Apakah actual state = desired state?" Kalau beda, dia yang bikin jadi sama.

```
Controller Manager = Housekeeping manager hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- "Harusnya ada 3 kamar bersih" (desired: 3 replicas)
- "Sekarang cuma 2 yang bersih" (actual: 2 running)
- "Bersihkan 1 kamar lagi!" (create 1 more pod)
- Loop terus setiap beberapa detik
```

### Worker Nodes — "Tangan" Cluster

Worker Node itu yang beneran jalanin container. Setiap node punya 3 komponen:

#### 1. kubelet

**Agent** yang jalan di setiap worker node. Dia yang terima instruksi dari Control Plane dan pastikan container jalan sesuai spec.

```
kubelet = Petugas kamar hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Terima instruksi: "Jalankan container X di node ini"
- Start container, monitor health
- Report status ke Control Plane: "Container X healthy" / "Container X crashed"
```

#### 2. kube-proxy

**Network proxy** di setiap node. Handle networking — routing traffic ke pod yang tepat, load balancing.

```
kube-proxy = Resepsionis lobby di setiap lantai
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- "Mau ketemu service A? Sini, lewat sini"
- Route traffic ke pod yang tepat
- Load balance kalau ada multiple pods
```

#### 3. Container Runtime

**Engine** yang beneran jalanin container. Biasanya **containerd** (Docker under the hood) atau **CRI-O**.

```
Container Runtime = Mesin AC, listrik, air di kamar
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Infrastructure dasar yang bikin container bisa jalan
- Kamu ga langsung interact, tapi tanpa ini ga jalan
```

### Master-Worker Relationship

```
User (kubectl)
     │
     ▼
API Server (Control Plane)
     │
     ├── "Deploy 3 replicas of task-manager app"
     │
     ▼
Scheduler: "Node 1 punya resource, Node 2 juga, Node 3 juga"
     │
     ▼
Controller Manager: "Create 3 pods — 1 di Node 1, 1 di Node 2, 1 di Node 3"
     │
     ├──────────────┬──────────────┐
     ▼              ▼              ▼
kubelet (Node 1)  kubelet (Node 2)  kubelet (Node 3)
     │              │              │
     ▼              ▼              ▼
Pod running!     Pod running!     Pod running!

etcd: "3 pods running: node1=healthy, node2=healthy, node3=healthy"
```

---

## 16.4 Core Concepts & Objects

### Kubernetes Object Hierarchy

Sebelum masuk ke detail, kamu harus paham hierarchy ini:

```
Deployment
  └── ReplicaSet
       └── Pod
            └── Container(s)

Penjelasan:
━━━━━━━━━━
Deployment = "Gue mau 3 replicas app versi v1.2.0"
  └── ReplicaSet = "Oke, gue maintain 3 pods dengan image v1.2.0"
       └── Pod = "Gue wrapper 1 atau lebih container"
            └── Container = "Gue container Docker yang beneran jalan"
```

### 1. Pod — Unit Terkecil

Pod itu **unit terkecil** di Kubernetes. Satu pod = satu atau lebih container yang share network dan storage.

```
Pod = Kamar hotel
━━━━━━━━━━━━━━━━
- 1 kamar bisa punya 1 bed (single container) atau 2 bed (multi-container)
- Semua orang di 1 kamar share kamar mandi (network namespace)
- Setiap kamar punya nomor sendiri (IP address)
- Kamar itu temporary — tamu checkout, kamar dibersihkan (pod dihapus)
```

**Penting:** Kamu JARANG bikin Pod langsung. Biasanya kamu bikin **Deployment** yang otomatis bikin Pod.

### 2. ReplicaSet — Maintain Jumlah Pod

ReplicaSet pastikan **jumlah pod yang jalan sesuai desired count**. Kalau 1 pod mati, ReplicaSet bikin yang baru.

```
ReplicaSet = "Selalu sediakan 3 kamar bersih"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- desired: 3 pods
- actual: 2 pods (1 crash)
- action: create 1 pod baru → actual: 3 pods
```

**Penting:** Kamu JARANG bikin ReplicaSet langsung. Deployment yang manage ReplicaSet.

### 3. Deployment — Declarative Management

Deployment itu **cara utama deploy app di K8s**. Kamu declare: "Gue mau 3 replicas, pakai image v1.2.0, port 8080" — dan K8s yang bikin jadi kenyataan.

```
Deployment = Standing order di hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"Selalu sediakan 3 kamar tipe deluxe, semua punya AC, WiFi, dan handuk bersih"
- Hotel (K8s) yang atur gimana achieve itu
- Kalau 1 kamar rusak, otomatis ganti
- Mau upgrade ke tipe suite? Gradually replace, ga perlu tutup hotel
```

### 4. Service — Network Access ke Pod

Service itu **stable network endpoint** untuk akses pod. Pod bisa mati dan respawn (IP berubah), tapi Service IP/DNS tetap sama.

```
Service = Nomor telepon hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━
- Kamu telpon 021-123456 → selalu nyambung ke hotel
- Staff yang angkat bisa beda-beda (pod beda-beda)
- Tapi nomor teleponnya tetap sama
- Kalau staff lagi ga available, dialihkan ke yang lain (load balancing)
```

### 5. Namespace — Virtual Cluster

Namespace itu **virtual isolation** dalam satu cluster. Kayak folder di komputer — organizing resources.

```
Namespace = Lantai di hotel
━━━━━━━━━━━━━━━━━━━━━━━━━
- Lantai 1: Restaurant (namespace: development)
- Lantai 2: Kantor (namespace: staging)
- Lantai 3: Kamar VIP (namespace: production)
- Setiap lantai punya rule sendiri
```

### 6. ConfigMap — Konfigurasi Non-Sensitive

ConfigMap simpan konfigurasi yang **tidak rahasia** sebagai key-value pairs.

```
ConfigMap = Papan informasi hotel (public)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Jam breakfast: 07:00-10:00 (APP_ENV=production)
- WiFi name: HotelWiFi (PORT=8080)
- Check-out time: 12:00 (LOG_LEVEL=info)
- Semua orang bisa baca — ga rahasia
```

### 7. Secret — Konfigurasi Sensitive

Secret simpan data **rahasia** — password, API keys, certificates. Di-encode Base64 (bukan encrypted by default, tapi bisa di-encrypt).

```
Secret = Brankas hotel
━━━━━━━━━━━━━━━━━━━━
- Master key kamar (DB_PASSWORD)
- Kode safe (JWT_SECRET)
- Cuma authorized staff yang boleh akses
```

### 8. Ingress — HTTP Routing

Ingress itu **HTTP router** — routing external traffic ke Service berdasarkan hostname/path. Kayak Nginx reverse proxy, tapi K8s-native.

```
Ingress = Lobby hotel dengan banyak pintu masuk
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Pintu "api.hotel.com" → ke restaurant (API service)
- Pintu "admin.hotel.com" → ke kantor (admin service)
- Pintu "hotel.com/docs" → ke perpustakaan (docs service)
- Semua lewat 1 lobby (Ingress Controller)
```

### 9. PersistentVolume (PV) — Storage

PersistentVolume itu **storage yang persist** meskipun pod dihapus. Penting untuk database.

```
PV = Safe deposit box hotel
━━━━━━━━━━━━━━━━━━━━━━━━━━
- Tamu checkout (pod dihapus), barang di safe deposit tetap ada
- Tamu baru bisa akses safe deposit yang sama (pod baru mount PV yang sama)
```

### 10. PersistentVolumeClaim (PVC) — Request Storage

PVC itu **request** untuk storage. Pod ga langsung minta PV — dia bikin PVC, dan K8s yang carikan PV yang cocok.

```
PVC = Formulir request safe deposit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- "Saya butuh safe deposit ukuran medium" (PVC: 10Gi)
- Hotel carikan yang available (K8s bind PV ke PVC)
```

### Rangkuman Objects

| Object | Fungsi | Analogi |
|--------|--------|---------|
| **Pod** | Unit terkecil, wrapper container | Kamar hotel |
| **ReplicaSet** | Maintain jumlah pod | "Selalu 3 kamar bersih" |
| **Deployment** | Declarative pod management | Standing order |
| **Service** | Stable network endpoint | Nomor telepon hotel |
| **Namespace** | Virtual isolation | Lantai hotel |
| **ConfigMap** | Config non-sensitive | Papan informasi |
| **Secret** | Config sensitive | Brankas |
| **Ingress** | HTTP routing | Lobby dengan banyak pintu |
| **PV** | Persistent storage | Safe deposit box |
| **PVC** | Request storage | Formulir request |

---

## 16.5 Install Tools

### Tools yang Kamu Butuhkan

```
Tool yang perlu di-install:
━━━━━━━━━━━━━━━━━━━━━━━━━
1. kubectl     — CLI untuk interact dengan K8s cluster
2. minikube    — Local K8s cluster di laptop (RECOMMENDED untuk belajar)
3. (optional)  — k3s, kind, Docker Desktop K8s

Kita pakai MINIKUBE untuk belajar. Kenapa?
- Gratis, local di laptop
- Full K8s feature
- Easy setup
- Community besar, docs bagus
```

### Install kubectl (macOS)

kubectl itu **CLI tool** untuk ngomong ke Kubernetes cluster. Semua interaksi kamu sama K8s lewat kubectl.

```bash
# Install via Homebrew (recommended macOS)
brew install kubectl

# Verify installation
kubectl version --client

# Output:
# Client Version: v1.31.x
# Kustomize Version: v5.x.x

# Atau kalau mau specific version
brew install kubectl@1.31
```

### Install minikube (macOS)

Minikube itu **local Kubernetes cluster** yang jalan di laptop kamu. Perfect untuk belajar.

```bash
# Install minikube via Homebrew
brew install minikube

# Verify
minikube version

# Output:
# minikube version: v1.34.x
```

### Start Minikube Cluster

```bash
# Start minikube (default: 2 CPU, 2GB RAM)
minikube start

# Start dengan resource lebih (recommended kalau laptop kuat)
minikube start --cpus=4 --memory=4096 --driver=docker

# Output:
# * minikube v1.34.x on Darwin (arm64)
# * Using the docker driver based on existing profile
# * Starting control plane node minikube in cluster minikube
# * Pulling base image ...
# * Creating docker container (CPUs=4, Memory=4096MB) ...
# * Preparing Kubernetes v1.31.x on Docker ...
# * Verifying Kubernetes components...
# * Enabled addons: storage-provisioner, default-storageclass
# * Done! kubectl is now configured to use "minikube" cluster

# Verify cluster running
kubectl cluster-info

# Output:
# Kubernetes control plane is running at https://127.0.0.1:xxxxx
# CoreDNS is running at https://127.0.0.1:xxxxx/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

# Check node
kubectl get nodes

# Output:
# NAME       STATUS   ROLES           AGE   VERSION
# minikube   Ready    control-plane   1m    v1.31.x
```

### Minikube Commands yang Sering Dipakai

```bash
# Start cluster
minikube start

# Stop cluster (pause, ga hapus data)
minikube stop

# Delete cluster (hapus semua)
minikube delete

# Status cluster
minikube status

# Dashboard (web UI)
minikube dashboard

# SSH ke minikube node
minikube ssh

# Enable addon
minikube addons enable ingress          # Nginx Ingress Controller
minikube addons enable metrics-server   # Metrics (untuk HPA)
minikube addons enable dashboard        # Web dashboard
minikube addons list                    # List semua addons

# Tunnel (untuk akses LoadBalancer service dari host)
minikube tunnel
```

### Alternatif: k3s, kind, Docker Desktop

```
Minikube (RECOMMENDED untuk belajar):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Full Kubernetes experience
- Addons (Ingress, Dashboard, Metrics)
- Pakai VM atau Docker
- Resource: 2+ CPU, 2+ GB RAM

kind (Kubernetes in Docker):
━━━━━━━━━━━━━━━━━━━━━━━━━━
- K8s cluster di Docker containers
- Lightweight, cepat start
- Bagus untuk CI/CD testing
- Install: brew install kind

k3s (Lightweight Kubernetes):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
- K8s yang diringankan (50MB binary)
- Production-ready untuk edge/IoT
- Install: curl -sfL https://get.k3s.io | sh -
- Kita bahas detail di 16.23

Docker Desktop Kubernetes:
━━━━━━━━━━━━━━━━━━━━━━━━━━
- Built-in di Docker Desktop
- Settings → Kubernetes → Enable
- Simple tapi limited addons
```

---

## 16.6 kubectl Basics

### kubectl = Swiss Army Knife

kubectl itu tool utama kamu untuk interact dengan K8s. Hampir SEMUA yang kamu lakuin di K8s lewat kubectl.

```
Format command:
kubectl <verb> <resource> [name] [flags]

Contoh:
kubectl get pods                    # verb=get, resource=pods
kubectl describe service my-api     # verb=describe, resource=service, name=my-api
kubectl delete pod my-pod-xyz       # verb=delete, resource=pod, name=my-pod-xyz
kubectl apply -f deployment.yaml    # verb=apply, file=deployment.yaml
```

### Essential Commands

#### get — Lihat Resource

```bash
# List semua pods
kubectl get pods

# List pods di semua namespace
kubectl get pods --all-namespaces
kubectl get pods -A

# List pods dengan info lebih detail
kubectl get pods -o wide

# List semua resource types
kubectl get all

# List specific resources
kubectl get deployments
kubectl get services
kubectl get configmaps
kubectl get secrets
kubectl get ingress
kubectl get nodes
kubectl get namespaces

# Output format
kubectl get pods -o yaml    # YAML output
kubectl get pods -o json    # JSON output
kubectl get pods -o wide    # Extra columns (NODE, IP, dll)
```

#### describe — Detail Info

```bash
# Detail info pod (events, conditions, containers)
kubectl describe pod my-pod-xyz

# Detail info deployment
kubectl describe deployment task-manager

# Detail info service
kubectl describe service task-manager-svc

# Detail info node
kubectl describe node minikube
```

#### apply — Create/Update dari File

```bash
# Apply (create atau update) dari file YAML
kubectl apply -f deployment.yaml

# Apply semua file di folder
kubectl apply -f k8s/

# Apply dari URL
kubectl apply -f https://raw.githubusercontent.com/.../deployment.yaml
```

#### delete — Hapus Resource

```bash
# Delete specific resource
kubectl delete pod my-pod-xyz
kubectl delete deployment task-manager
kubectl delete service task-manager-svc

# Delete dari file (hapus semua yang didefinisikan di file)
kubectl delete -f deployment.yaml

# Delete semua pods di namespace
kubectl delete pods --all

# Delete namespace (hapus SEMUA di namespace itu!)
kubectl delete namespace my-namespace
```

#### logs — Lihat Log Container

```bash
# Lihat logs pod
kubectl logs my-pod-xyz

# Follow logs (real-time)
kubectl logs -f my-pod-xyz

# Last 100 lines
kubectl logs --tail=100 my-pod-xyz

# Logs dari container specific (kalau pod punya multiple containers)
kubectl logs my-pod-xyz -c container-name

# Logs dari semua pods dengan label tertentu
kubectl logs -l app=task-manager
```

#### exec — Masuk ke Container

```bash
# Shell ke container (kayak docker exec -it)
kubectl exec -it my-pod-xyz -- /bin/sh

# Jalankan command di container
kubectl exec my-pod-xyz -- ls -la /app
kubectl exec my-pod-xyz -- env
kubectl exec my-pod-xyz -- cat /etc/hosts
```

#### port-forward — Akses dari Localhost

```bash
# Forward port dari pod ke localhost
kubectl port-forward pod/my-pod-xyz 8080:8080

# Forward port dari service
kubectl port-forward service/task-manager-svc 8080:8080

# Akses di browser: http://localhost:8080
# Ctrl+C untuk stop
```

### Context dan Namespace

```bash
# Lihat current context (cluster mana yang active)
kubectl config current-context

# List semua contexts
kubectl config get-contexts

# Switch context
kubectl config use-context minikube

# Set default namespace (biar ga perlu -n terus)
kubectl config set-context --current --namespace=production

# Jalankan command di namespace specific
kubectl get pods -n kube-system
kubectl get pods -n production
kubectl get pods -n development
```

### Cheat Sheet — kubectl Quick Reference

```bash
# ========== GET (Lihat) ==========
kubectl get pods                     # List pods
kubectl get pods -o wide             # List pods + detail
kubectl get pods -A                  # Semua namespace
kubectl get all                      # Semua resource
kubectl get deploy,svc,pods          # Multiple resource types

# ========== DESCRIBE (Detail) ==========
kubectl describe pod <name>          # Detail pod
kubectl describe deploy <name>       # Detail deployment
kubectl describe svc <name>          # Detail service

# ========== CREATE/UPDATE ==========
kubectl apply -f file.yaml           # Create/update dari file
kubectl apply -f dir/                # Apply semua file di folder

# ========== DELETE ==========
kubectl delete -f file.yaml          # Delete dari file
kubectl delete pod <name>            # Delete pod
kubectl delete deploy <name>         # Delete deployment

# ========== DEBUG ==========
kubectl logs <pod>                   # View logs
kubectl logs -f <pod>                # Follow logs
kubectl exec -it <pod> -- /bin/sh    # Shell ke pod
kubectl port-forward svc/<name> 8080:8080  # Port forward

# ========== SCALE ==========
kubectl scale deploy <name> --replicas=5   # Scale manual

# ========== ROLLOUT ==========
kubectl rollout status deploy <name>       # Status rollout
kubectl rollout history deploy <name>      # History rollout
kubectl rollout undo deploy <name>         # Rollback
```

---

## 16.7 Pod — Unit Terkecil K8s

### Apa Itu Pod?

Pod itu unit terkecil yang bisa kamu deploy di Kubernetes. Satu pod berisi **satu atau lebih container** yang share:
- **Network namespace** — semua container di pod punya IP yang sama
- **Storage volumes** — bisa share volume
- **Lifecycle** — start dan stop bareng

```
Pod ≠ Container
━━━━━━━━━━━━━━
Pod itu WRAPPER untuk container.
Biasanya 1 pod = 1 container.
Tapi bisa juga 1 pod = 2+ containers (sidecar pattern).

Docker: docker run myapp
K8s:    kubectl run myapp --image=myapp
        → K8s bikin POD yang berisi container myapp
```

### Pod YAML Manifest — Single Container

```yaml
# pod-basic.yaml
# Pod sederhana — 1 container
apiVersion: v1
kind: Pod
metadata:
  name: task-manager-pod
  labels:
    app: task-manager
    version: v1.0.0
spec:
  containers:
    - name: task-manager
      image: task-manager:latest
      ports:
        - containerPort: 8080
      env:
        - name: APP_ENV
          value: "production"
        - name: PORT
          value: "8080"
      resources:
        requests:
          memory: "64Mi"
          cpu: "100m"
        limits:
          memory: "128Mi"
          cpu: "250m"
```

Penjelasan:

```
apiVersion: v1        → Versi API (Pod pakai v1)
kind: Pod             → Tipe resource
metadata:
  name:               → Nama pod (unique di namespace)
  labels:             → Key-value pairs untuk organize/select pods
spec:
  containers:         → List container di pod ini
    - name:           → Nama container
      image:          → Docker image
      ports:          → Port yang di-expose
      env:            → Environment variables
      resources:
        requests:     → Minimum resource yang dibutuhkan (scheduling)
        limits:       → Maximum resource yang boleh dipakai (enforcement)

CPU units:
  "100m" = 100 milli-CPU = 0.1 CPU = 10% dari 1 CPU core
  "250m" = 250 milli-CPU = 0.25 CPU = 25% dari 1 CPU core
  "1"    = 1 full CPU core

Memory units:
  "64Mi"  = 64 MiB (mebibytes)
  "128Mi" = 128 MiB
  "1Gi"   = 1 GiB (gibibyte)
```

```bash
# Create pod
kubectl apply -f pod-basic.yaml

# Check status
kubectl get pods

# Output:
# NAME                 READY   STATUS    RESTARTS   AGE
# task-manager-pod     1/1     Running   0          30s

# Detail info
kubectl describe pod task-manager-pod

# Logs
kubectl logs task-manager-pod

# Delete pod
kubectl delete pod task-manager-pod
```

### Multi-Container Pod — Sidecar Pattern

Kadang kamu butuh **lebih dari 1 container** di pod. Pattern yang paling umum: **sidecar**.

```
Sidecar Pattern:
━━━━━━━━━━━━━━━
Container utama + container pembantu yang share resources.

Contoh:
- App container + log shipper container (kirim logs ke Loki)
- App container + proxy container (Envoy sidecar)
- App container + config watcher (reload config otomatis)
```

```yaml
# pod-sidecar.yaml
# Pod dengan sidecar pattern — app + log forwarder
apiVersion: v1
kind: Pod
metadata:
  name: task-manager-with-sidecar
  labels:
    app: task-manager
spec:
  containers:
    # Container utama — Go app
    - name: task-manager
      image: task-manager:latest
      ports:
        - containerPort: 8080
      volumeMounts:
        - name: shared-logs
          mountPath: /var/log/app

    # Sidecar — log forwarder
    - name: log-forwarder
      image: busybox:1.36
      command: ["sh", "-c", "tail -f /var/log/app/*.log"]
      volumeMounts:
        - name: shared-logs
          mountPath: /var/log/app

  volumes:
    - name: shared-logs
      emptyDir: {}
```

### Pod Lifecycle

```
Pod Lifecycle:
━━━━━━━━━━━━━

Pending ──→ Running ──→ Succeeded
                   │
                   └──→ Failed
                   │
                   └──→ Unknown

Pending:    Pod diterima K8s, tapi belum running
            (image lagi di-pull, menunggu scheduling)

Running:    Minimal 1 container running

Succeeded:  Semua container selesai dengan exit code 0
            (biasanya untuk Job, bukan long-running app)

Failed:     Minimal 1 container exit dengan error

Unknown:    K8s ga bisa determine status
            (biasanya node communication issue)
```

**Penting:** Kamu JARANG bikin Pod langsung. Di production, kamu selalu pakai **Deployment** yang manage pods. Pod langsung itu untuk testing/debugging aja.

---

## 16.8 Deployment — Declarative Pod Management

### Kenapa Deployment, Bukan Pod Langsung?

```
Pod langsung:
━━━━━━━━━━━━
- Pod crash → ga ada yang restart
- Mau 3 replicas → bikin 3 Pod YAML manual
- Update image → delete pod manual, create baru manual
- Ga ada rolling update, ga ada rollback

Deployment:
━━━━━━━━━━
- Pod crash → Deployment bikin pod baru otomatis
- Mau 3 replicas → set replicas: 3, done
- Update image → rolling update otomatis, zero downtime
- Bisa rollback ke versi sebelumnya
```

### Full Deployment YAML — Task Manager Go App

```yaml
# k8s/deployment.yaml
# Deployment untuk Task Manager Go API
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-manager
  labels:
    app: task-manager
    tier: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: task-manager
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1           # Boleh ada 1 pod LEBIH dari replicas saat update
      maxUnavailable: 0      # 0 pod boleh down saat update (zero downtime!)
  template:
    metadata:
      labels:
        app: task-manager
        version: v1.0.0
    spec:
      containers:
        - name: task-manager
          image: task-manager:v1.0.0
          ports:
            - containerPort: 8080
              protocol: TCP
          env:
            - name: APP_ENV
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: APP_ENV
            - name: PORT
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: PORT
            - name: DB_HOST
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: DB_HOST
            - name: DB_PORT
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: DB_PORT
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: task-manager-secret
                  key: DB_USER
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: task-manager-secret
                  key: DB_PASSWORD
            - name: DB_NAME
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: DB_NAME
            - name: REDIS_HOST
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: REDIS_HOST
            - name: REDIS_PORT
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: REDIS_PORT
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: task-manager-secret
                  key: JWT_SECRET
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
```

### Penjelasan Detail

```yaml
# replicas: 3
# "Gue mau 3 instances dari app ini running setiap saat"

# selector.matchLabels
# Deployment manage pods yang punya label "app: task-manager"
# Ini HARUS match dengan template.metadata.labels

# strategy.type: RollingUpdate
# Saat update, ganti pod satu-per-satu (bukan semua sekaligus)

# maxSurge: 1
# Saat rolling update, boleh ada 1 pod EXTRA
# Jadi kalau replicas=3, bisa ada 4 pod sebentar saat transition
# Artinya: 3 pods lama masih serve + 1 pod baru lagi starting

# maxUnavailable: 0
# Ga boleh ada pod yang down saat update
# Ini guarantee zero-downtime update
# Pod baru HARUS ready dulu sebelum pod lama di-kill

# template
# Ini template untuk Pod yang akan dibuat
# Setiap pod yang dibuat Deployment akan pakai spec ini
```

### Deploy dan Verify

```bash
# Apply deployment
kubectl apply -f k8s/deployment.yaml

# Check deployment status
kubectl get deployments

# Output:
# NAME           READY   UP-TO-DATE   AVAILABLE   AGE
# task-manager   3/3     3            3           45s

# Check pods yang dibuat
kubectl get pods

# Output:
# NAME                            READY   STATUS    RESTARTS   AGE
# task-manager-7d6f8c5b4f-abc12   1/1     Running   0          45s
# task-manager-7d6f8c5b4f-def34   1/1     Running   0          45s
# task-manager-7d6f8c5b4f-ghi56   1/1     Running   0          45s

# Detail deployment
kubectl describe deployment task-manager

# Watch pods real-time
kubectl get pods -w
```

### Rolling Update Strategy

```
Rolling Update dengan maxSurge=1, maxUnavailable=0:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Step 1 (initial): 3 pods v1.0.0 running
[v1.0.0] [v1.0.0] [v1.0.0]

Step 2: Create 1 pod v1.1.0 (maxSurge=1, jadi total bisa 4)
[v1.0.0] [v1.0.0] [v1.0.0] [v1.1.0 starting...]

Step 3: v1.1.0 ready. Kill 1 v1.0.0
[v1.0.0] [v1.0.0] [v1.1.0]

Step 4: Create 1 more v1.1.0
[v1.0.0] [v1.0.0] [v1.1.0] [v1.1.0 starting...]

Step 5: v1.1.0 ready. Kill 1 v1.0.0
[v1.0.0] [v1.1.0] [v1.1.0]

Step 6: Create 1 more v1.1.0
[v1.0.0] [v1.1.0] [v1.1.0] [v1.1.0 starting...]

Step 7: v1.1.0 ready. Kill last v1.0.0
[v1.1.0] [v1.1.0] [v1.1.0]

DONE! Zero downtime. User ga ngerasa apa-apa.
```

### Rollback

```bash
# Update image ke versi baru
kubectl set image deployment/task-manager task-manager=task-manager:v1.1.0

# Monitor rollout progress
kubectl rollout status deployment/task-manager

# Output:
# Waiting for deployment "task-manager" rollout to finish:
#   1 out of 3 new replicas have been updated...
#   2 out of 3 new replicas have been updated...
#   3 out of 3 new replicas have been updated...
# deployment "task-manager" successfully rolled out

# Lihat rollout history
kubectl rollout history deployment/task-manager

# Output:
# REVISION  CHANGE-CAUSE
# 1         <none>
# 2         <none>

# OH NO! v1.1.0 ada bug! Rollback ke versi sebelumnya
kubectl rollout undo deployment/task-manager

# Rollback ke revision specific
kubectl rollout undo deployment/task-manager --to-revision=1

# Verify
kubectl rollout status deployment/task-manager
kubectl get pods
```

---

## 16.9 Service — Network Access ke Pod

### Kenapa Butuh Service?

```
Masalah:
━━━━━━━━
Pod punya IP. Tapi pod bisa mati dan respawn — IP berubah!
Kalau app lain hardcode IP pod, waktu pod restart, koneksi putus.

Solusi:
━━━━━━━
Service = stable endpoint.
Service punya IP dan DNS name yang TETAP.
Service route traffic ke pods di belakangnya.
Pod mati? Service otomatis route ke pod yang hidup.
```

### 3 Tipe Service

```
┌─────────────────────────────────────────────────────────────┐
│                    Service Types                              │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐   │
│  │  ClusterIP   │  │  NodePort    │  │  LoadBalancer      │   │
│  │              │  │              │  │                     │   │
│  │  Internal    │  │  External    │  │  External           │   │
│  │  only        │  │  (dev/test)  │  │  (production)       │   │
│  │              │  │              │  │                     │   │
│  │  Akses dari  │  │  Akses dari  │  │  Akses dari         │   │
│  │  dalam       │  │  <NodeIP>:   │  │  cloud LB           │   │
│  │  cluster     │  │  <NodePort>  │  │  (AWS ELB, etc)     │   │
│  │              │  │              │  │                     │   │
│  │  Default     │  │  Port range: │  │  Auto provision     │   │
│  │              │  │  30000-32767 │  │  external LB        │   │
│  └─────────────┘  └─────────────┘  └───────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

#### ClusterIP (Default) — Internal Only

Cuma bisa diakses dari DALAM cluster. Service-to-service communication.

```yaml
# Contoh: PostgreSQL service (cuma diakses internal oleh Go app)
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  type: ClusterIP    # Default, bisa dihilangkan
  selector:
    app: postgres
  ports:
    - port: 5432           # Port di service
      targetPort: 5432     # Port di container
```

#### NodePort — External Access (Dev/Testing)

Bisa diakses dari luar cluster via `<NodeIP>:<NodePort>`.

```yaml
# Contoh: Go app accessible dari luar (development)
apiVersion: v1
kind: Service
metadata:
  name: task-manager-nodeport
spec:
  type: NodePort
  selector:
    app: task-manager
  ports:
    - port: 8080           # Port di service (internal)
      targetPort: 8080     # Port di container
      nodePort: 30080      # Port di node (external, 30000-32767)
```

#### LoadBalancer — External Access (Production)

Provision cloud load balancer (AWS ELB, GCP LB). Production-grade.

```yaml
# Contoh: Go app di production (AWS EKS)
apiVersion: v1
kind: Service
metadata:
  name: task-manager-lb
spec:
  type: LoadBalancer
  selector:
    app: task-manager
  ports:
    - port: 80             # External port
      targetPort: 8080     # Container port
```

### Full Service YAML — Task Manager

```yaml
# k8s/service.yaml
# Service untuk Task Manager Go API
apiVersion: v1
kind: Service
metadata:
  name: task-manager-svc
  labels:
    app: task-manager
spec:
  type: ClusterIP
  selector:
    app: task-manager      # Route ke pods dengan label app=task-manager
  ports:
    - name: http
      port: 80             # Port yang dipake service lain untuk connect
      targetPort: 8080     # Port di container Go app
      protocol: TCP
```

### Service Discovery by DNS

Ini salah satu fitur paling powerful K8s. Setiap Service otomatis punya **DNS name**.

```
DNS Format:
<service-name>.<namespace>.svc.cluster.local

Contoh:
task-manager-svc.default.svc.cluster.local
postgres.default.svc.cluster.local
redis.default.svc.cluster.local

Shortcut (di namespace yang sama):
task-manager-svc   ← cukup nama service aja!
postgres
redis
```

Artinya, di config Go app kamu, DB_HOST bukan IP address — tapi **nama service**:

```yaml
env:
  - name: DB_HOST
    value: "postgres"       # Bukan 10.0.0.5, tapi nama service!
  - name: REDIS_HOST
    value: "redis"          # K8s DNS resolve ini ke IP pod Redis
```

Ini persis kayak Docker Compose! Di docker-compose, kamu pakai service name. Di K8s juga sama.

```bash
# Apply service
kubectl apply -f k8s/service.yaml

# Check service
kubectl get services

# Output:
# NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
# task-manager-svc   ClusterIP   10.96.123.45    <none>        80/TCP    30s
# kubernetes         ClusterIP   10.96.0.1       <none>        443/TCP   1h

# Test dari dalam cluster (port-forward ke localhost)
kubectl port-forward service/task-manager-svc 8080:80

# Sekarang bisa akses: http://localhost:8080/health
```

---

## 16.10 ConfigMap & Secret

### ConfigMap — Non-Sensitive Configuration

ConfigMap itu tempat simpan config yang **tidak rahasia** — APP_ENV, PORT, DB_HOST, dll.

```yaml
# k8s/configmap.yaml
# ConfigMap untuk Task Manager
apiVersion: v1
kind: ConfigMap
metadata:
  name: task-manager-config
  labels:
    app: task-manager
data:
  APP_ENV: "production"
  PORT: "8080"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "task_manager"
  REDIS_HOST: "redis"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
```

```bash
# Apply configmap
kubectl apply -f k8s/configmap.yaml

# Check configmap
kubectl get configmaps

# Detail
kubectl describe configmap task-manager-config

# Output data
kubectl get configmap task-manager-config -o yaml
```

### Secret — Sensitive Configuration

Secret itu tempat simpan data **rahasia** — password, API keys, certificates.

```yaml
# k8s/secret.yaml
# Secret untuk Task Manager
apiVersion: v1
kind: Secret
metadata:
  name: task-manager-secret
  labels:
    app: task-manager
type: Opaque
data:
  # Nilai di-encode base64
  # echo -n "postgres" | base64 → cG9zdGdyZXM=
  # echo -n "supersecretpassword" | base64 → c3VwZXJzZWNyZXRwYXNzd29yZA==
  # echo -n "my-jwt-secret-key-2026" | base64 → bXktand0LXNlY3JldC1rZXktMjAyNg==
  DB_USER: cG9zdGdyZXM=
  DB_PASSWORD: c3VwZXJzZWNyZXRwYXNzd29yZA==
  JWT_SECRET: bXktand0LXNlY3JldC1rZXktMjAyNg==
  REDIS_PASSWORD: ""
```

#### Base64 Encoding

```bash
# Encode ke base64
echo -n "postgres" | base64
# Output: cG9zdGdyZXM=

echo -n "supersecretpassword" | base64
# Output: c3VwZXJzZWNyZXRwYXNzd29yZA==

echo -n "my-jwt-secret-key-2026" | base64
# Output: bXktand0LXNlY3JldC1rZXktMjAyNg==

# Decode dari base64
echo "cG9zdGdyZXM=" | base64 -d
# Output: postgres
```

**PENTING:** Base64 itu **bukan encryption**! Base64 cuma encoding — siapapun bisa decode. Untuk encryption beneran, pakai:
- **Sealed Secrets** (Bitnami)
- **SOPS** (Mozilla)
- **Vault** (HashiCorp)
- **AWS Secrets Manager + External Secrets Operator**

#### Alternatif: stringData (Tanpa Base64)

```yaml
# Secret dengan stringData — ga perlu encode base64 manual
apiVersion: v1
kind: Secret
metadata:
  name: task-manager-secret
type: Opaque
stringData:
  DB_USER: "postgres"
  DB_PASSWORD: "supersecretpassword"
  JWT_SECRET: "my-jwt-secret-key-2026"
  REDIS_PASSWORD: ""
```

K8s otomatis encode ke base64 saat disimpan. `stringData` lebih enak buat development, tapi JANGAN commit ke git!

### Pakai ConfigMap & Secret di Deployment

Ada 2 cara: **environment variable** atau **mount sebagai file**.

#### Cara 1: Environment Variable (yang kita pakai)

```yaml
# Di Deployment spec
containers:
  - name: task-manager
    env:
      # Dari ConfigMap
      - name: APP_ENV
        valueFrom:
          configMapKeyRef:
            name: task-manager-config
            key: APP_ENV

      # Dari Secret
      - name: DB_PASSWORD
        valueFrom:
          secretKeyRef:
            name: task-manager-secret
            key: DB_PASSWORD

      # Atau load SEMUA keys dari ConfigMap sekaligus
    envFrom:
      - configMapRef:
          name: task-manager-config
      - secretRef:
          name: task-manager-secret
```

#### Cara 2: Mount sebagai File

```yaml
# Mount ConfigMap sebagai file
containers:
  - name: task-manager
    volumeMounts:
      - name: config-volume
        mountPath: /etc/config
volumes:
  - name: config-volume
    configMap:
      name: task-manager-config

# File yang muncul di container:
# /etc/config/APP_ENV → berisi "production"
# /etc/config/PORT → berisi "8080"
# /etc/config/DB_HOST → berisi "postgres"
```

```bash
# Apply secret
kubectl apply -f k8s/secret.yaml

# Check secrets (nilai ga ditampilkan)
kubectl get secrets

# Detail (nilai di-encode base64)
kubectl get secret task-manager-secret -o yaml

# Decode specific value
kubectl get secret task-manager-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
# Output: supersecretpassword
```

---

## 16.11 Ingress — HTTP Routing

### Apa Itu Ingress?

Ingress itu **HTTP(S) router** di K8s — kayak Nginx reverse proxy, tapi K8s-native. Dia route traffic berdasarkan hostname dan path.

```
Tanpa Ingress:
━━━━━━━━━━━━━
Setiap service butuh LoadBalancer sendiri.
5 services = 5 load balancers = 5x biaya = 5 IP publik.

Dengan Ingress:
━━━━━━━━━━━━━━
1 Ingress Controller (1 LoadBalancer) route ke semua services.
5 services = 1 load balancer = 1 IP publik.
Routing berdasarkan domain/path.

                    ┌───────────────────────────┐
                    │     Ingress Controller     │
                    │    (Nginx / Traefik)       │
Internet ─────────→│                            │
                    │  api.taskmanager.com/      │──→ task-manager-svc
                    │  api.taskmanager.com/docs  │──→ docs-svc
                    │  admin.taskmanager.com/    │──→ admin-svc
                    └───────────────────────────┘
```

### Ingress Controller

Ingress **resource** (YAML) cuma define rules. Kamu butuh **Ingress Controller** yang beneran execute rules itu. Yang paling populer: **nginx-ingress**.

```bash
# Enable Ingress Controller di minikube
minikube addons enable ingress

# Verify
kubectl get pods -n ingress-nginx

# Output:
# NAME                                        READY   STATUS    RESTARTS   AGE
# ingress-nginx-controller-xxxxxxx-xxxxx      1/1     Running   0          1m
```

### Path-Based Routing

```yaml
# k8s/ingress.yaml
# Ingress — route berdasarkan path
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-manager-ingress
  labels:
    app: task-manager
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-manager-svc
                port:
                  number: 80
```

### Host-Based Routing (Multiple Domains)

```yaml
# Ingress dengan multiple hosts
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: multi-host-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
    # API
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-manager-svc
                port:
                  number: 80
    # Admin panel
    - host: admin.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: admin-svc
                port:
                  number: 80
```

### TLS/SSL

```yaml
# Ingress dengan TLS
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-manager-ingress-tls
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.taskmanager.com
      secretName: taskmanager-tls-secret    # K8s Secret dengan TLS cert
  rules:
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-manager-svc
                port:
                  number: 80
```

### Full Ingress YAML untuk Task Manager

```yaml
# k8s/ingress.yaml
# Production-ready Ingress untuk Task Manager
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-manager-ingress
  labels:
    app: task-manager
  annotations:
    # Nginx specific
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    # CORS
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "*"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "Authorization, Content-Type"
spec:
  ingressClassName: nginx
  rules:
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-manager-svc
                port:
                  number: 80
```

### Test Ingress di Minikube

```bash
# Apply ingress
kubectl apply -f k8s/ingress.yaml

# Check ingress
kubectl get ingress

# Output:
# NAME                     CLASS   HOSTS                  ADDRESS        PORTS   AGE
# task-manager-ingress     nginx   api.taskmanager.com    192.168.49.2   80      30s

# Dapatkan minikube IP
minikube ip
# Output: 192.168.49.2

# Tambahkan ke /etc/hosts (supaya domain resolve ke minikube)
echo "$(minikube ip) api.taskmanager.com" | sudo tee -a /etc/hosts

# Test
curl http://api.taskmanager.com/health
# Output: {"status":"ok"}
```

---

## 16.12 Persistent Storage

### Masalah: Container = Ephemeral

Container itu sementara. Pod dihapus = data hilang. Ini masalah untuk **database** — kamu ga mau data PostgreSQL hilang kalau pod restart.

```
Tanpa PersistentVolume:
━━━━━━━━━━━━━━━━━━━━━
Pod PostgreSQL running → simpan data
Pod PostgreSQL crash → restart → DATA HILANG! Semua table kosong.

Dengan PersistentVolume:
━━━━━━━━━━━━━━━━━━━━━━
Pod PostgreSQL running → simpan data ke PV (disk external)
Pod PostgreSQL crash → restart → mount PV yang sama → DATA MASIH ADA!
```

### PersistentVolume (PV) dan PersistentVolumeClaim (PVC)

```
PV  = "Disk 10GB tersedia di storage" (provisioned oleh admin)
PVC = "Gue butuh disk 10GB" (requested oleh pod)
K8s match PVC ke PV yang cocok.

       Admin creates PV          Pod creates PVC
       ┌──────────┐              ┌──────────┐
       │  PV: 10Gi │←── bind ──→│ PVC: 10Gi │←── mount ── Pod
       └──────────┘              └──────────┘
```

### StorageClass — Dynamic Provisioning

Di production (EKS, GKE), kamu ga perlu bikin PV manual. **StorageClass** otomatis provision storage.

```
Static Provisioning (manual):
Admin bikin PV → Pod bikin PVC → K8s bind

Dynamic Provisioning (otomatis):
Pod bikin PVC → StorageClass otomatis bikin PV → K8s bind
Lebih modern, lebih enak.
```

### Full PV dan PVC YAML

```yaml
# k8s/storage.yaml
# PersistentVolume untuk PostgreSQL data

# PV — disk yang tersedia (di minikube pakai hostPath)
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres-pv
  labels:
    type: local
    app: postgres
spec:
  storageClassName: manual
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce       # Hanya 1 node yang bisa mount
  hostPath:
    path: "/mnt/data/postgres"

---
# PVC — request dari pod untuk storage
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  labels:
    app: postgres
spec:
  storageClassName: manual
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

### Access Modes

```
ReadWriteOnce (RWO):
- 1 node bisa mount read-write
- Cocok untuk database (PostgreSQL, MySQL)
- Paling umum

ReadOnlyMany (ROX):
- Multiple nodes bisa mount read-only
- Cocok untuk shared config/assets

ReadWriteMany (RWX):
- Multiple nodes bisa mount read-write
- Butuh network storage (NFS, EFS)
- Cocok untuk shared storage
```

```bash
# Apply storage
kubectl apply -f k8s/storage.yaml

# Check PV dan PVC
kubectl get pv
# NAME          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   STORAGECLASS
# postgres-pv   5Gi        RWO            Retain           Bound    manual

kubectl get pvc
# NAME           STATUS   VOLUME        CAPACITY   ACCESS MODES   STORAGECLASS
# postgres-pvc   Bound    postgres-pv   5Gi        RWO            manual
```

---

## 16.13 Deploy PostgreSQL on K8s

### StatefulSet, Bukan Deployment — Kenapa?

```
Deployment vs StatefulSet:
━━━━━━━━━━━━━━━━━━━━━━━━

Deployment:
- Pod names = random (task-manager-7d6f8c5b4f-abc12)
- Pod bisa di-schedule ke node mana aja
- Cocok untuk STATELESS app (Go API, Node.js)
- Pod interchangeable — ga peduli yang mana

StatefulSet:
- Pod names = ordered (postgres-0, postgres-1, postgres-2)
- Stable network identity
- Stable persistent storage (PVC per pod)
- Ordered deployment dan scaling
- Cocok untuk STATEFUL app (PostgreSQL, Redis cluster, Kafka)
- Pod identity matters — postgres-0 itu master, postgres-1 itu replica
```

### Full PostgreSQL StatefulSet Manifests

```yaml
# k8s/postgres-statefulset.yaml
# PostgreSQL StatefulSet + Headless Service

# Headless Service (ClusterIP: None)
# Untuk stable DNS per pod: postgres-0.postgres.default.svc.cluster.local
apiVersion: v1
kind: Service
metadata:
  name: postgres
  labels:
    app: postgres
spec:
  type: ClusterIP
  clusterIP: None         # Headless! DNS resolve langsung ke pod IP
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432

---
# ConfigMap untuk PostgreSQL config
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
data:
  POSTGRES_DB: "task_manager"

---
# Secret untuk PostgreSQL credentials
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
type: Opaque
stringData:
  POSTGRES_USER: "postgres"
  POSTGRES_PASSWORD: "supersecretpassword"

---
# StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  labels:
    app: postgres
spec:
  serviceName: postgres    # Harus match dengan Headless Service name
  replicas: 1             # Single instance (untuk production, pakai 2+ dengan replication)
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
              name: postgres
          env:
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: postgres-config
                  key: POSTGRES_DB
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: POSTGRES_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: POSTGRES_PASSWORD
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: postgres-data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3

  # VolumeClaimTemplate — setiap pod dapat PVC sendiri
  volumeClaimTemplates:
    - metadata:
        name: postgres-data
      spec:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
```

### Kenapa Headless Service?

```
Regular Service (ClusterIP):
━━━━━━━━━━━━━━━━━━━━━━━━━━
DNS: postgres → resolve ke CLUSTER IP → load balance ke pods
Kamu ga tau traffic pergi ke pod mana.

Headless Service (ClusterIP: None):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DNS: postgres → resolve ke SEMUA pod IPs
DNS: postgres-0.postgres → resolve ke pod 0 SPECIFICALLY

Untuk database, kamu butuh connect ke specific pod
(master vs replica). Headless Service enables that.
```

```bash
# Deploy PostgreSQL
kubectl apply -f k8s/postgres-statefulset.yaml

# Check StatefulSet
kubectl get statefulsets

# Output:
# NAME       READY   AGE
# postgres   1/1     60s

# Check pod (nama ordered!)
kubectl get pods

# Output:
# NAME         READY   STATUS    RESTARTS   AGE
# postgres-0   1/1     Running   0          60s
# (kalau replicas: 2, ada postgres-1)

# Check PVC (otomatis dibuat oleh volumeClaimTemplate)
kubectl get pvc

# Output:
# NAME                    STATUS   VOLUME     CAPACITY   ACCESS MODES
# postgres-data-postgres-0   Bound    pvc-xxx   5Gi        RWO

# Test connect ke PostgreSQL
kubectl exec -it postgres-0 -- psql -U postgres -d task_manager

# Inside psql:
# task_manager=# \dt
# task_manager=# SELECT 1;
# task_manager=# \q
```

---

## 16.14 Deploy Redis on K8s

Redis untuk caching biasanya **single-node** (ga perlu StatefulSet). Deployment cukup.

```yaml
# k8s/redis.yaml
# Redis Deployment + Service

apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          command:
            - redis-server
            - "--appendonly"
            - "yes"
            - "--maxmemory"
            - "128mb"
            - "--maxmemory-policy"
            - "allkeys-lru"
          resources:
            requests:
              memory: "64Mi"
              cpu: "50m"
            limits:
              memory: "192Mi"
              cpu: "200m"
          livenessProbe:
            exec:
              command:
                - redis-cli
                - ping
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            exec:
              command:
                - redis-cli
                - ping
            initialDelaySeconds: 3
            periodSeconds: 5
          volumeMounts:
            - name: redis-data
              mountPath: /data
      volumes:
        - name: redis-data
          emptyDir: {}    # Data persist selama pod hidup.
                          # Untuk persistent, pakai PVC.

---
# Service untuk Redis
apiVersion: v1
kind: Service
metadata:
  name: redis
  labels:
    app: redis
spec:
  type: ClusterIP
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
```

```bash
# Deploy Redis
kubectl apply -f k8s/redis.yaml

# Check
kubectl get pods -l app=redis
kubectl get svc redis

# Test connect
kubectl exec -it $(kubectl get pod -l app=redis -o jsonpath='{.items[0].metadata.name}') -- redis-cli ping
# Output: PONG
```

---

## 16.15 Deploy Task Manager — Full Stack

### Struktur File K8s

Ini full structure K8s manifests untuk Task Manager:

```
k8s/
├── namespace.yaml              # Namespace
├── configmap.yaml              # ConfigMap
├── secret.yaml                 # Secret
├── postgres-statefulset.yaml   # PostgreSQL
├── redis.yaml                  # Redis
├── deployment.yaml             # Go API
├── service.yaml                # Go API Service
├── ingress.yaml                # Ingress
├── hpa.yaml                    # HorizontalPodAutoscaler
└── kustomization.yaml          # (optional) Kustomize
```

### Step 1: Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: task-manager
  labels:
    app: task-manager
    environment: production
```

### Step 2: ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: task-manager-config
  namespace: task-manager
  labels:
    app: task-manager
data:
  APP_ENV: "production"
  PORT: "8080"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "task_manager"
  REDIS_HOST: "redis"
  REDIS_PORT: "6379"
  LOG_LEVEL: "info"
```

### Step 3: Secret

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: task-manager-secret
  namespace: task-manager
  labels:
    app: task-manager
type: Opaque
stringData:
  DB_USER: "postgres"
  DB_PASSWORD: "supersecretpassword"
  JWT_SECRET: "my-jwt-secret-key-2026-change-in-production"
  REDIS_PASSWORD: ""
```

### Step 4: PostgreSQL StatefulSet + Service + PVC

```yaml
# k8s/postgres-statefulset.yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: task-manager
  labels:
    app: postgres
spec:
  type: ClusterIP
  clusterIP: None
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: task-manager
  labels:
    app: postgres
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: task-manager-config
                  key: DB_NAME
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: task-manager-secret
                  key: DB_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: task-manager-secret
                  key: DB_PASSWORD
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: postgres-data
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 5
            periodSeconds: 5
  volumeClaimTemplates:
    - metadata:
        name: postgres-data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 5Gi
```

### Step 5: Redis Deployment + Service

```yaml
# k8s/redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: task-manager
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          command: ["redis-server", "--appendonly", "yes", "--maxmemory", "128mb", "--maxmemory-policy", "allkeys-lru"]
          resources:
            requests:
              memory: "64Mi"
              cpu: "50m"
            limits:
              memory: "192Mi"
              cpu: "200m"
          livenessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 3
            periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: task-manager
  labels:
    app: redis
spec:
  type: ClusterIP
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
```

### Step 6: Go App Deployment + Service

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-manager
  namespace: task-manager
  labels:
    app: task-manager
    tier: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: task-manager
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: task-manager
        version: v1.0.0
    spec:
      containers:
        - name: task-manager
          image: task-manager:v1.0.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              protocol: TCP
          envFrom:
            - configMapRef:
                name: task-manager-config
            - secretRef:
                name: task-manager-secret
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 3
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 2
            periodSeconds: 5
            failureThreshold: 10
```

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: task-manager-svc
  namespace: task-manager
  labels:
    app: task-manager
spec:
  type: ClusterIP
  selector:
    app: task-manager
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
```

### Step 7: Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: task-manager-ingress
  namespace: task-manager
  labels:
    app: task-manager
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
spec:
  ingressClassName: nginx
  rules:
    - host: api.taskmanager.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: task-manager-svc
                port:
                  number: 80
```

### Step 8: HorizontalPodAutoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: task-manager-hpa
  namespace: task-manager
  labels:
    app: task-manager
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: task-manager
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70    # Scale up kalau avg CPU > 70%
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80    # Scale up kalau avg memory > 80%
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60    # Tunggu 60s sebelum scale up lagi
      policies:
        - type: Pods
          value: 2                       # Max tambah 2 pods per step
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300    # Tunggu 5 menit sebelum scale down
      policies:
        - type: Pods
          value: 1                       # Max kurangi 1 pod per step
          periodSeconds: 60
```

### Deploy Semua — Step by Step

```bash
# ====================================
# DEPLOY FULL STACK TASK MANAGER ON K8S
# ====================================

# 0. Pastikan minikube running
minikube start --cpus=4 --memory=4096
minikube addons enable ingress
minikube addons enable metrics-server

# 1. Build Docker image di minikube's Docker daemon
# (supaya minikube bisa akses image tanpa registry)
eval $(minikube docker-env)
docker build -t task-manager:v1.0.0 .

# 2. Create namespace
kubectl apply -f k8s/namespace.yaml

# 3. Apply ConfigMap dan Secret
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# 4. Deploy PostgreSQL
kubectl apply -f k8s/postgres-statefulset.yaml

# Tunggu PostgreSQL ready
kubectl wait --for=condition=ready pod/postgres-0 -n task-manager --timeout=120s

# 5. Deploy Redis
kubectl apply -f k8s/redis.yaml

# Tunggu Redis ready
kubectl wait --for=condition=ready -l app=redis pod -n task-manager --timeout=60s

# 6. Deploy Go App
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 7. Deploy Ingress
kubectl apply -f k8s/ingress.yaml

# 8. Deploy HPA
kubectl apply -f k8s/hpa.yaml

# ====================================
# VERIFY EVERYTHING
# ====================================

# Check semua resources di namespace
kubectl get all -n task-manager

# Expected output:
# NAME                                READY   STATUS    RESTARTS   AGE
# pod/postgres-0                      1/1     Running   0          2m
# pod/redis-xxxxxxx-xxxxx             1/1     Running   0          1m
# pod/task-manager-xxxxxxx-xxxxx      1/1     Running   0          45s
# pod/task-manager-xxxxxxx-yyyyy      1/1     Running   0          45s
# pod/task-manager-xxxxxxx-zzzzz      1/1     Running   0          45s
#
# NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)
# service/postgres           ClusterIP   None           <none>        5432/TCP
# service/redis              ClusterIP   10.96.x.x      <none>        6379/TCP
# service/task-manager-svc   ClusterIP   10.96.x.x      <none>        80/TCP
#
# NAME                           READY   UP-TO-DATE   AVAILABLE
# deployment.apps/redis          1/1     1            1
# deployment.apps/task-manager   3/3     3            3
#
# NAME                                      DESIRED   CURRENT   READY
# replicaset.apps/redis-xxxxxxx             1         1         1
# replicaset.apps/task-manager-xxxxxxx      3         3         3
#
# NAME                        READY
# statefulset.apps/postgres   1/1

# Check ingress
kubectl get ingress -n task-manager

# Check HPA
kubectl get hpa -n task-manager

# ====================================
# TEST
# ====================================

# Port-forward untuk test
kubectl port-forward service/task-manager-svc 8080:80 -n task-manager

# Test endpoint
curl http://localhost:8080/health
# Output: {"status":"ok"}

curl http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'

# Atau pakai Ingress (kalau udah setup /etc/hosts)
echo "$(minikube ip) api.taskmanager.com" | sudo tee -a /etc/hosts
curl http://api.taskmanager.com/health
```

### Shortcut: Apply Semua Sekaligus

```bash
# Kalau semua file di folder k8s/
kubectl apply -f k8s/ -n task-manager

# Atau apply in order (recommended untuk pertama kali)
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml -f k8s/secret.yaml
kubectl apply -f k8s/postgres-statefulset.yaml
sleep 30  # tunggu PostgreSQL ready
kubectl apply -f k8s/redis.yaml
sleep 10  # tunggu Redis ready
kubectl apply -f k8s/deployment.yaml -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
```

---

## 16.16 Scaling

### Manual Scaling

```bash
# Scale deployment ke 5 replicas
kubectl scale deployment task-manager --replicas=5 -n task-manager

# Check
kubectl get pods -n task-manager -l app=task-manager

# Output:
# NAME                            READY   STATUS    RESTARTS   AGE
# task-manager-xxxxxxx-11111      1/1     Running   0          5m
# task-manager-xxxxxxx-22222      1/1     Running   0          5m
# task-manager-xxxxxxx-33333      1/1     Running   0          5m
# task-manager-xxxxxxx-44444      1/1     Running   0          10s   ← baru!
# task-manager-xxxxxxx-55555      1/1     Running   0          10s   ← baru!

# Scale down ke 2
kubectl scale deployment task-manager --replicas=2 -n task-manager
```

### HorizontalPodAutoscaler (HPA) — Auto Scaling

HPA otomatis adjust jumlah replicas berdasarkan **CPU** atau **memory** usage.

```
HPA Logic:
━━━━━━━━━
desired_replicas = ceil(current_replicas * (current_metric / target_metric))

Contoh:
- current_replicas: 3
- current CPU usage: 90%
- target CPU: 70%
- desired = ceil(3 * (90/70)) = ceil(3.86) = 4

→ HPA scale up ke 4 replicas!

Contoh scale down:
- current_replicas: 4
- current CPU usage: 30%
- target CPU: 70%
- desired = ceil(4 * (30/70)) = ceil(1.71) = 2

→ HPA scale down ke 2 replicas!
```

```yaml
# k8s/hpa.yaml (Full version, sama dengan di atas)
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: task-manager-hpa
  namespace: task-manager
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: task-manager
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

```bash
# Prerequisite: metrics-server harus enabled
minikube addons enable metrics-server

# Apply HPA
kubectl apply -f k8s/hpa.yaml

# Check HPA status
kubectl get hpa -n task-manager

# Output:
# NAME               REFERENCE                 TARGETS           MINPODS   MAXPODS   REPLICAS
# task-manager-hpa   Deployment/task-manager   15%/70%, 20%/80%   2         10        3

# Watch HPA (real-time)
kubectl get hpa -n task-manager -w

# Detail HPA
kubectl describe hpa task-manager-hpa -n task-manager
```

### VPA (Vertical Pod Autoscaler) — Concept

```
HPA = Horizontal scaling → tambah JUMLAH pod
VPA = Vertical scaling → tambah RESOURCE per pod

HPA: 3 pods × 100m CPU → 5 pods × 100m CPU (lebih banyak pod)
VPA: 3 pods × 100m CPU → 3 pods × 250m CPU (pod lebih kuat)

Kapan pakai VPA:
- Ga bisa horizontal scale (database, single-instance)
- Pod butuh resource lebih, bukan pod lebih banyak
- Fine-tuning resource requests

Catatan: VPA ga built-in — harus install terpisah.
Untuk sekarang, HPA aja dulu.
```

---

## 16.17 Health Checks in K8s

### 3 Tipe Probe

K8s punya 3 tipe health check (probe):

```
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes Probes                          │
│                                                              │
│  ┌───────────────┐  ┌───────────────┐  ┌─────────────────┐ │
│  │ livenessProbe  │  │readinessProbe │  │  startupProbe   │ │
│  │                │  │               │  │                 │ │
│  │ "Masih hidup?" │  │ "Siap terima  │  │ "Udah start?"   │ │
│  │                │  │  traffic?"    │  │                 │ │
│  │ Gagal → K8s    │  │ Gagal → K8s   │  │ Gagal → K8s     │ │
│  │ RESTART pod    │  │ STOP kirim    │  │ RESTART pod     │ │
│  │                │  │ traffic       │  │ (tapi tunggu    │ │
│  │                │  │ (tapi pod     │  │  lebih lama)    │ │
│  │                │  │  tetap hidup) │  │                 │ │
│  └───────────────┘  └───────────────┘  └─────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### livenessProbe — "Masih Hidup?"

Cek apakah container masih **alive**. Kalau gagal, K8s **kill dan restart** container.

```
Kapan pakai:
- Detect deadlock (app masih running tapi ga respond)
- Detect hang (goroutine stuck)
- App yang ga bisa recover sendiri

Endpoint: /health — simple check, return 200 OK
```

#### readinessProbe — "Siap Terima Traffic?"

Cek apakah container **siap terima request**. Kalau gagal, K8s **stop kirim traffic** (tapi ga restart).

```
Kapan pakai:
- App masih loading config/data
- App lagi warm up cache
- Database belum connected
- Dependency belum ready

Endpoint: /health — bisa cek DB + Redis connectivity
```

#### startupProbe — "Udah Start?"

Cek apakah container udah **selesai startup**. Selama startup probe belum sukses, liveness dan readiness probe ga dijalankan.

```
Kapan pakai:
- App yang startup-nya LAMA (loading model ML, migration DB)
- Biar liveness probe ga kill pod yang masih starting
- Timeout lebih panjang dari liveness
```

### 3 Metode Probe

```
1. httpGet — HTTP request ke endpoint
   Sukses kalau response status 200-399

2. exec — Jalankan command di container
   Sukses kalau exit code 0

3. tcpSocket — TCP connection ke port
   Sukses kalau port reachable
```

### Full Probe Configuration untuk Task Manager

```yaml
# Di Deployment spec → containers → probes
containers:
  - name: task-manager
    image: task-manager:v1.0.0
    ports:
      - containerPort: 8080

    # Startup Probe — cek sekali saat startup
    # Timeout total: failureThreshold × periodSeconds = 10 × 5 = 50 detik
    startupProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 2     # Tunggu 2 detik setelah container start
      periodSeconds: 5           # Cek setiap 5 detik
      failureThreshold: 10      # Boleh gagal 10 kali sebelum restart

    # Liveness Probe — cek terus selama pod running
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5     # Mulai cek setelah startup probe done
      periodSeconds: 10          # Cek setiap 10 detik
      timeoutSeconds: 3          # Timeout per check
      failureThreshold: 3       # 3 kali gagal berturut-turut → restart
      successThreshold: 1       # 1 kali sukses → dianggap healthy

    # Readiness Probe — cek apakah siap terima traffic
    readinessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 3
      periodSeconds: 5
      timeoutSeconds: 3
      failureThreshold: 3       # 3 kali gagal → stop traffic
      successThreshold: 1       # 1 kali sukses → mulai terima traffic
```

### Connect ke Health Endpoint dari Bab 15

Di Bab 15 (Monitoring), kita udah bikin health check endpoint:

```go
// internal/handler/health.go (sudah ada di project)
func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}
```

Endpoint `/health` ini yang dipake oleh semua probe. Kalau kamu mau readinessProbe yang lebih thorough (cek DB + Redis), bisa bikin endpoint `/ready` terpisah:

```go
// internal/handler/health.go — enhanced version
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/redis/go-redis/v9"
    "context"
    "time"
)

// HealthCheck — liveness probe (simple, fast)
func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
    })
}

// ReadinessCheck — readiness probe (cek dependencies)
func ReadinessCheck(db *gorm.DB, redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
        defer cancel()

        // Check database
        sqlDB, err := db.DB()
        if err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "status": "not ready",
                "reason": "database connection failed",
            })
            return
        }
        if err := sqlDB.PingContext(ctx); err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "status": "not ready",
                "reason": "database ping failed",
            })
            return
        }

        // Check Redis
        if err := redisClient.Ping(ctx).Err(); err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "status": "not ready",
                "reason": "redis ping failed",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "status": "ready",
            "checks": gin.H{
                "database": "connected",
                "redis":    "connected",
            },
        })
    }
}
```

Kalau udah punya `/ready`, readinessProbe bisa pakai path `/ready` yang lebih thorough:

```yaml
readinessProbe:
  httpGet:
    path: /ready       # Cek DB + Redis
    port: 8080
livenessProbe:
  httpGet:
    path: /health      # Simple check aja
    port: 8080
```

---

## 16.18 Rolling Update & Rollback

### Zero-Downtime Deployment

Ini salah satu alasan utama perusahaan pakai Kubernetes — deploy versi baru **tanpa downtime**.

```
Tanpa K8s (traditional deploy):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Stop old version         ← DOWNTIME STARTS
2. Deploy new version
3. Start new version
4. Wait for health check    ← DOWNTIME ENDS
Total downtime: 30 detik - 5 menit

Dengan K8s (rolling update):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Start new pod (v2)
2. Wait for readiness probe → new pod ready
3. Route traffic ke new pod
4. Kill old pod (v1)
5. Repeat for all pods
Total downtime: 0
```

### Update Image Version

```bash
# Cara 1: Edit YAML, ubah image tag, apply
# (recommended — declarative, version controlled)

# Edit k8s/deployment.yaml:
# image: task-manager:v1.0.0  →  image: task-manager:v1.1.0
kubectl apply -f k8s/deployment.yaml

# Cara 2: kubectl set image (imperative — quick update)
kubectl set image deployment/task-manager \
  task-manager=task-manager:v1.1.0 \
  -n task-manager

# Cara 3: kubectl edit (edit live — NOT recommended)
kubectl edit deployment task-manager -n task-manager
```

### Monitor Rollout

```bash
# Watch rollout status (real-time)
kubectl rollout status deployment/task-manager -n task-manager

# Output (live):
# Waiting for deployment "task-manager" rollout to finish:
#   1 out of 3 new replicas have been updated...
#   1 out of 3 new replicas have been updated...
#   2 out of 3 new replicas have been updated...
#   3 out of 3 new replicas have been updated...
#   Waiting for 1 old replicas to be terminated...
# deployment "task-manager" successfully rolled out

# Watch pods (lihat rolling update happen)
kubectl get pods -n task-manager -w

# Output (live):
# NAME                            READY   STATUS              RESTARTS   AGE
# task-manager-old-11111          1/1     Running             0          10m
# task-manager-old-22222          1/1     Running             0          10m
# task-manager-old-33333          1/1     Running             0          10m
# task-manager-new-aaaaa          0/1     ContainerCreating   0          2s
# task-manager-new-aaaaa          1/1     Running             0          5s
# task-manager-old-11111          1/1     Terminating         0          10m
# task-manager-new-bbbbb          0/1     ContainerCreating   0          1s
# ...
```

### Rollback

```bash
# Rollout history
kubectl rollout history deployment/task-manager -n task-manager

# Output:
# REVISION  CHANGE-CAUSE
# 1         Initial deployment
# 2         kubectl set image deployment/task-manager task-manager=task-manager:v1.1.0

# Rollback ke versi sebelumnya (revision sebelumnya)
kubectl rollout undo deployment/task-manager -n task-manager

# Rollback ke revision specific
kubectl rollout undo deployment/task-manager -n task-manager --to-revision=1

# Verify
kubectl rollout status deployment/task-manager -n task-manager
kubectl describe deployment task-manager -n task-manager | grep Image
```

### Blue-Green Deployment (Concept)

```
Blue-Green Deployment:
━━━━━━━━━━━━━━━━━━━━━
2 environment: Blue (current) dan Green (new)
Deploy ke Green. Test. Switch traffic. Done.

     ┌─────────────┐
     │   Service     │
     │  (selector:   │
     │   version=v1) │ ← switch selector ke v2 saat ready
     └──────┬───────┘
            │
     ┌──────┴──────────┐
     ▼                 ▼
┌──────────┐    ┌──────────┐
│  Blue     │    │  Green    │
│  (v1)     │    │  (v2)     │
│  ACTIVE   │    │  STANDBY  │
└──────────┘    └──────────┘

Step 1: Deploy v2 ke Green
Step 2: Test Green internally
Step 3: Switch Service selector ke v2
Step 4: Green = ACTIVE, Blue = STANDBY
Step 5: Kalau ada masalah, switch back ke v1 (instant rollback!)
```

### Canary Deployment (Concept)

```
Canary Deployment:
━━━━━━━━━━━━━━━━━
Roll out ke sebagian KECIL users dulu. Monitor. Kalau oke, gradually roll ke semua.

Step 1: 90% traffic → v1, 10% traffic → v2 (canary)
Step 2: Monitor error rate, latency, dll
Step 3: Kalau oke → 70% v1, 30% v2
Step 4: Kalau oke → 50/50
Step 5: Kalau oke → 0% v1, 100% v2
Step 6: Remove v1

Tools untuk canary di K8s:
- Istio (service mesh)
- Flagger (automated canary)
- Argo Rollouts
- Nginx canary annotations
```

---

## 16.19 Namespace & Resource Management

### Namespace untuk Environment Isolation

```bash
# Buat namespace per environment
kubectl create namespace development
kubectl create namespace staging
kubectl create namespace production
```

```yaml
# k8s/namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: development
  labels:
    environment: development
---
apiVersion: v1
kind: Namespace
metadata:
  name: staging
  labels:
    environment: staging
---
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    environment: production
```

```
Namespace Strategy:
━━━━━━━━━━━━━━━━━
development  → Developer testing, playground
staging      → Pre-production, QA testing
production   → Live traffic, real users

Atau per-team:
team-backend    → Backend services
team-frontend   → Frontend services
team-data       → Data pipelines

Benefit:
- Resource isolation (tiap namespace bisa punya quota)
- Access control (RBAC per namespace)
- Network policies per namespace
- Billing tracking (tag resource per namespace)
```

### ResourceQuota — Limit per Namespace

```yaml
# k8s/resource-quota.yaml
# Limit resource yang boleh dipakai namespace development
apiVersion: v1
kind: ResourceQuota
metadata:
  name: dev-quota
  namespace: development
spec:
  hard:
    # Compute
    requests.cpu: "4"           # Total CPU requests: max 4 cores
    requests.memory: "8Gi"     # Total memory requests: max 8Gi
    limits.cpu: "8"            # Total CPU limits: max 8 cores
    limits.memory: "16Gi"     # Total memory limits: max 16Gi

    # Object count
    pods: "20"                 # Max 20 pods
    services: "10"             # Max 10 services
    configmaps: "20"           # Max 20 ConfigMaps
    secrets: "20"              # Max 20 Secrets
    persistentvolumeclaims: "5"  # Max 5 PVCs
```

### LimitRange — Default Resource per Container

```yaml
# k8s/limit-range.yaml
# Default resource limit per container di namespace
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: development
spec:
  limits:
    - type: Container
      default:
        cpu: "200m"
        memory: "256Mi"
      defaultRequest:
        cpu: "100m"
        memory: "128Mi"
      max:
        cpu: "1"
        memory: "1Gi"
      min:
        cpu: "50m"
        memory: "32Mi"
```

### Resource Requests vs Limits

```
requests: "Minimal resource yang DIJAMIN"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Dipakai Scheduler untuk decide pod jalan di node mana
- Node harus punya resource >= requests untuk schedule pod
- "Ini yang gue butuhkan MINIMAL"

limits: "Maximum resource yang BOLEH dipakai"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Kalau container pakai CPU > limit → throttled (diperlambat)
- Kalau container pakai memory > limit → OOMKilled (dibunuh!)
- "Ini batas maximum gue"

Best Practice:
resources:
  requests:
    cpu: "100m"        # "Biasanya gue pakai 10% CPU"
    memory: "128Mi"    # "Biasanya gue pakai 128MB RAM"
  limits:
    cpu: "500m"        # "Paling banyak gue butuh 50% CPU"
    memory: "256Mi"    # "Paling banyak 256MB RAM, lebih dari ini = bug"
```

```bash
# Apply resource management
kubectl apply -f k8s/resource-quota.yaml
kubectl apply -f k8s/limit-range.yaml

# Check quota usage
kubectl describe resourcequota dev-quota -n development

# Output:
# Name:                   dev-quota
# Namespace:              development
# Resource                Used    Hard
# --------                ----    ----
# limits.cpu              500m    8
# limits.memory           512Mi   16Gi
# pods                    3       20
# requests.cpu            200m    4
# requests.memory         256Mi   8Gi
```

---

## 16.20 Monitoring in K8s

### Kubernetes + Prometheus = Perfect Match

Di Bab 15, kita setup Prometheus + Grafana di Docker Compose. Di K8s, monitoring jadi LEBIH powerful karena K8s native support Prometheus.

```
K8s Monitoring Stack:
━━━━━━━━━━━━━━━━━━━━

┌────────────────────────────────────────────────────────┐
│                  Monitoring Stack                        │
│                                                          │
│  Prometheus Operator                                     │
│  ├── Prometheus       → Scrape & store metrics           │
│  ├── Alertmanager     → Handle alerts                    │
│  └── Grafana          → Visualize dashboards             │
│                                                          │
│  kube-state-metrics   → K8s object metrics               │
│  │                      (pod count, deployment status)    │
│  │                                                        │
│  node-exporter        → Node metrics                     │
│                         (CPU, memory, disk, network)      │
│                                                          │
│  App metrics          → Custom app metrics               │
│                         (/metrics endpoint dari Go app)   │
└────────────────────────────────────────────────────────┘
```

### Prometheus Operator

Prometheus Operator itu cara paling gampang deploy monitoring stack di K8s. Dia manage Prometheus, Alertmanager, Grafana sebagai K8s resources.

```bash
# Install kube-prometheus-stack via Helm (detail di section Helm)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin123

# Check pods
kubectl get pods -n monitoring

# Output:
# NAME                                                     READY   STATUS    AGE
# monitoring-grafana-xxxxxxx-xxxxx                         3/3     Running   2m
# monitoring-kube-prometheus-operator-xxxxxxx-xxxxx        1/1     Running   2m
# monitoring-kube-state-metrics-xxxxxxx-xxxxx              1/1     Running   2m
# monitoring-prometheus-node-exporter-xxxxx                1/1     Running   2m
# prometheus-monitoring-kube-prometheus-prometheus-0        2/2     Running   2m
# alertmanager-monitoring-kube-prometheus-alertmanager-0    2/2     Running   2m
```

### ServiceMonitor — Auto Discover App Metrics

Supaya Prometheus otomatis scrape metrics dari Go app kita, bikin **ServiceMonitor**:

```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: task-manager-monitor
  namespace: task-manager
  labels:
    app: task-manager
    release: monitoring     # Harus match dengan Helm release name
spec:
  selector:
    matchLabels:
      app: task-manager
  endpoints:
    - port: http            # Nama port di Service
      path: /metrics        # Path metrics endpoint
      interval: 15s         # Scrape setiap 15 detik
  namespaceSelector:
    matchNames:
      - task-manager
```

### Access Grafana

```bash
# Port-forward Grafana
kubectl port-forward service/monitoring-grafana 3000:80 -n monitoring

# Buka browser: http://localhost:3000
# Username: admin
# Password: admin123 (yang kita set di Helm)

# Built-in dashboards:
# - Kubernetes / Compute Resources / Cluster
# - Kubernetes / Compute Resources / Namespace (Pods)
# - Kubernetes / Compute Resources / Node
# - Node Exporter / Nodes
# - CoreDNS
# - etcd
```

### Grafana Dashboards untuk K8s

```
Grafana Dashboard IDs (import dari grafana.com):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Dashboard ID: 315   → Kubernetes Cluster Overview
Dashboard ID: 6417  → Kubernetes Pods
Dashboard ID: 3119  → Kubernetes Cluster Monitoring (via Prometheus)
Dashboard ID: 1860  → Node Exporter Full
Dashboard ID: 12740 → Kubernetes Pod Resources

Import: Grafana → + → Import → Enter ID → Load → Select Datasource → Import
```

---

## 16.21 Helm — Package Manager untuk K8s

### Apa Itu Helm?

Helm itu **package manager** untuk Kubernetes. Kalau npm buat Node.js, pip buat Python, Helm buat K8s.

```
Analogi:
━━━━━━━
npm install express        → Install Express.js package
helm install prometheus    → Install Prometheus di K8s cluster

Tanpa Helm:
- Bikin 10+ YAML files untuk Prometheus
- Manually configure setiap setting
- Update? Edit 10+ files manual

Dengan Helm:
- helm install prometheus
- Semua YAML otomatis di-generate dari template
- Configure via values.yaml
- Update? helm upgrade
```

### Konsep Helm

```
Chart   = Package (template YAML files + default values)
         Kayak npm package — kumpulan file yang bisa di-install

Values  = Configuration file (values.yaml)
         Kayak .env — customize behavior

Release = Instance dari chart yang udah di-install
         Kayak running process — chart yang udah deployed
```

### Install Helm

```bash
# Install via Homebrew
brew install helm

# Verify
helm version

# Output:
# version.BuildInfo{Version:"v3.x.x", ...}
```

### Helm Basics

```bash
# Add repository (tempat chart disimpan)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Search chart
helm search repo prometheus
helm search repo nginx
helm search repo postgresql

# Install chart
helm install my-release chart-name

# Install dengan custom values
helm install my-release chart-name -f values.yaml

# Install di namespace specific
helm install my-release chart-name -n namespace --create-namespace

# List installed releases
helm list
helm list -A    # semua namespace

# Upgrade release (update values atau chart version)
helm upgrade my-release chart-name -f values.yaml

# Rollback release
helm rollback my-release 1    # rollback ke revision 1

# Uninstall release
helm uninstall my-release

# Show default values (untuk tau apa aja yang bisa di-configure)
helm show values chart-name
```

### Install Apps dengan Helm

#### Nginx Ingress Controller

```bash
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace
```

#### Prometheus + Grafana

```bash
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin123 \
  --set prometheus.prometheusSpec.retention=7d
```

#### PostgreSQL

```bash
helm install postgres bitnami/postgresql \
  --namespace task-manager \
  --set auth.postgresPassword=supersecret \
  --set auth.database=task_manager \
  --set primary.persistence.size=10Gi
```

#### Redis

```bash
helm install redis bitnami/redis \
  --namespace task-manager \
  --set auth.enabled=false \
  --set master.persistence.size=1Gi \
  --set replica.replicaCount=0
```

### Create Custom Helm Chart — Task Manager

```bash
# Scaffold chart baru
helm create task-manager-chart

# Struktur yang dihasilkan:
# task-manager-chart/
# ├── Chart.yaml           # Metadata chart
# ├── values.yaml          # Default values
# ├── templates/
# │   ├── deployment.yaml  # Deployment template
# │   ├── service.yaml     # Service template
# │   ├── ingress.yaml     # Ingress template
# │   ├── hpa.yaml         # HPA template
# │   ├── serviceaccount.yaml
# │   ├── _helpers.tpl     # Template helpers
# │   ├── NOTES.txt        # Post-install notes
# │   └── tests/
# │       └── test-connection.yaml
# └── charts/              # Sub-chart dependencies
```

#### Chart.yaml

```yaml
# task-manager-chart/Chart.yaml
apiVersion: v2
name: task-manager
description: A Helm chart for Task Manager Go API
type: application
version: 1.0.0        # Chart version
appVersion: "v1.0.0"  # App version
```

#### values.yaml

```yaml
# task-manager-chart/values.yaml
# Default values — bisa di-override per environment

replicaCount: 3

image:
  repository: task-manager
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  host: api.taskmanager.com
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"

resources:
  requests:
    cpu: 100m
    memory: 64Mi
  limits:
    cpu: 500m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

config:
  APP_ENV: production
  PORT: "8080"
  DB_HOST: postgres
  DB_PORT: "5432"
  DB_NAME: task_manager
  REDIS_HOST: redis
  REDIS_PORT: "6379"

secrets:
  DB_USER: postgres
  DB_PASSWORD: supersecretpassword
  JWT_SECRET: my-jwt-secret-key-2026

probes:
  liveness:
    path: /health
    initialDelaySeconds: 5
    periodSeconds: 10
  readiness:
    path: /health
    initialDelaySeconds: 3
    periodSeconds: 5
```

#### templates/deployment.yaml

```yaml
# task-manager-chart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "task-manager.fullname" . }}
  labels:
    {{- include "task-manager.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "task-manager.selectorLabels" . | nindent 6 }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        {{- include "task-manager.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: {{ .Values.service.targetPort }}
          envFrom:
            - configMapRef:
                name: {{ include "task-manager.fullname" . }}-config
            - secretRef:
                name: {{ include "task-manager.fullname" . }}-secret
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          livenessProbe:
            httpGet:
              path: {{ .Values.probes.liveness.path }}
              port: {{ .Values.service.targetPort }}
            initialDelaySeconds: {{ .Values.probes.liveness.initialDelaySeconds }}
            periodSeconds: {{ .Values.probes.liveness.periodSeconds }}
          readinessProbe:
            httpGet:
              path: {{ .Values.probes.readiness.path }}
              port: {{ .Values.service.targetPort }}
            initialDelaySeconds: {{ .Values.probes.readiness.initialDelaySeconds }}
            periodSeconds: {{ .Values.probes.readiness.periodSeconds }}
```

#### Install Custom Chart

```bash
# Install dengan default values
helm install task-manager ./task-manager-chart -n task-manager --create-namespace

# Install dengan custom values (staging)
helm install task-manager-staging ./task-manager-chart \
  -n staging \
  --create-namespace \
  -f values-staging.yaml

# Upgrade (setelah update chart/values)
helm upgrade task-manager ./task-manager-chart -n task-manager -f values.yaml

# Dry run (preview tanpa apply)
helm install task-manager ./task-manager-chart --dry-run --debug

# Template (render YAML tanpa install)
helm template task-manager ./task-manager-chart
```

---

## 16.22 Managed Kubernetes — EKS, GKE, AKS

### Apa Itu Managed Kubernetes?

```
Self-hosted K8s:
━━━━━━━━━━━━━━━
- Kamu setup Control Plane sendiri
- Kamu manage etcd, API Server, upgrades
- Kamu handle networking, storage
- Kamu yang patch security vulnerabilities
- Full control, full responsibility

Managed K8s (EKS, GKE, AKS):
━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Cloud provider manage Control Plane
- Kamu cuma manage Worker Nodes (dan kadang itu juga managed)
- Automatic upgrades, patching
- Integrated dengan cloud services (IAM, ELB, EBS, dll)
- Bayar untuk Control Plane + Worker Nodes
```

### Comparison

| Aspek | AWS EKS | GCP GKE | Azure AKS |
|-------|---------|---------|-----------|
| **Cloud** | Amazon Web Services | Google Cloud Platform | Microsoft Azure |
| **Control Plane Cost** | $0.10/hour ($73/month) | Free (1 cluster) | Free |
| **Worker Nodes** | EC2 instances | GCE instances | Azure VMs |
| **Auto Scaling** | Cluster Autoscaler / Karpenter | Cluster Autoscaler | Cluster Autoscaler |
| **Managed Nodes** | Fargate (serverless) | Autopilot | Virtual Nodes |
| **Storage** | EBS, EFS | Persistent Disk | Azure Disk |
| **Load Balancer** | ELB/ALB/NLB | GCP LB | Azure LB |
| **Registry** | ECR | GCR / Artifact Registry | ACR |
| **Monitoring** | CloudWatch | Cloud Monitoring | Azure Monitor |
| **Learning Curve** | Medium-High | Low-Medium | Medium |
| **Best For** | Enterprise, AWS shop | GCP shop, K8s purist | Microsoft shop |

### AWS EKS Setup Basics

```bash
# Prerequisites
# - AWS CLI configured
# - eksctl installed (brew install eksctl)

# Create EKS cluster
eksctl create cluster \
  --name task-manager-cluster \
  --region ap-southeast-1 \
  --version 1.31 \
  --nodegroup-name workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 5

# Ini butuh 15-20 menit. eksctl bikin:
# - VPC + Subnets
# - EKS Control Plane
# - Worker Node Group (EC2 instances)
# - IAM Roles
# - Security Groups

# Verify
kubectl get nodes

# Deploy app (sama kayak minikube!)
kubectl apply -f k8s/

# Delete cluster (PENTING! Biar ga kena charge)
eksctl delete cluster --name task-manager-cluster --region ap-southeast-1
```

### Cost Considerations

```
AWS EKS Estimated Cost (minimal setup):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EKS Control Plane:  $0.10/hour × 730 hours = $73/month
3x t3.medium:       $0.0416/hour × 3 × 730 = $91/month
EBS Storage (30GB): $0.10/GB × 30 × 3 = $9/month
Load Balancer:      ~$16/month + data transfer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$189/month (minimal)

Dengan auto-scaling bisa naik ke $300-500/month.

Bandingkan:
- Single EC2 t3.medium: ~$30/month
- K8s EKS: ~$189/month minimum

K8s MAHAL untuk project kecil!
Tapi WORTH IT untuk production yang butuh:
- High availability
- Auto-scaling
- Zero-downtime deployment
- Multi-service orchestration
```

### When to Self-host vs Managed

```
Self-host (k3s/kubeadm):
━━━━━━━━━━━━━━━━━━━━━━━
- Budget sangat terbatas
- 1-3 server aja cukup
- Tim punya expertise manage K8s
- Edge/IoT deployment
- Development/staging environment
- Cost: $5-30/month (VPS aja)

Managed (EKS/GKE/AKS):
━━━━━━━━━━━━━━━━━━━━━━
- Production yang serius
- Ga mau manage Control Plane
- Butuh integration cloud services
- Tim ga punya K8s expertise yang deep
- SLA requirements
- Cost: $100-500+/month
```

---

## 16.23 k3s — Lightweight Kubernetes

### Apa Itu k3s?

k3s itu **Kubernetes yang diringankan** — dibikin oleh Rancher (sekarang SUSE). Binary-nya cuma **~50MB** (vs K8s yang ratusan MB). Production-ready, CNCF certified.

```
k3s Facts:
━━━━━━━━━
- Binary size: ~50MB (vs K8s ~300MB+)
- Memory usage: ~512MB (vs K8s 2GB+)
- Single binary — install dalam 30 detik
- Includes: containerd, Flannel, CoreDNS, Traefik
- CNCF certified Kubernetes
- Perfect for: edge, IoT, CI/CD, small projects, budget setups
```

### Install k3s di EC2

```bash
# SSH ke EC2 instance (Ubuntu)
ssh -i key.pem ubuntu@your-ec2-ip

# Install k3s (literally 1 command!)
curl -sfL https://get.k3s.io | sh -

# That's it. K8s cluster running.

# Check
sudo kubectl get nodes
# NAME          STATUS   ROLES                  AGE   VERSION
# ip-10-0-1-5   Ready    control-plane,master   30s   v1.31.x+k3s1

# Check pods
sudo kubectl get pods -A
# NAMESPACE     NAME                                      READY   STATUS
# kube-system   coredns-xxxxxxx-xxxxx                     1/1     Running
# kube-system   local-path-provisioner-xxxxxxx-xxxxx      1/1     Running
# kube-system   metrics-server-xxxxxxx-xxxxx              1/1     Running
# kube-system   svclb-traefik-xxxxx                       2/2     Running
# kube-system   traefik-xxxxxxx-xxxxx                     1/1     Running

# k3s config kubectl
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# Atau biar ga perlu sudo
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $USER:$USER ~/.kube/config
```

### Deploy Task Manager di k3s

```bash
# Build image di server (atau push ke registry dulu)
docker build -t task-manager:v1.0.0 .

# Import image ke k3s (k3s pakai containerd, bukan Docker)
sudo k3s ctr images import task-manager-v1.0.0.tar

# Atau simpelnya, push ke Docker Hub / GitHub Container Registry
# lalu k3s pull dari situ

# Deploy (YAML yang sama!)
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# k3s udah include Traefik sebagai Ingress Controller
# Jadi Ingress langsung jalan tanpa install tambahan!

# Verify
kubectl get all -n task-manager
```

### k3s vs Minikube vs EKS

| Aspek | k3s | Minikube | EKS |
|-------|-----|----------|-----|
| **Use case** | Production (small) | Local dev/learning | Production (enterprise) |
| **Resources** | 512MB RAM | 2GB+ RAM | EC2 instances |
| **Install time** | 30 detik | 2-5 menit | 15-20 menit |
| **Cost** | $5/month (VPS) | Free (local) | $100+/month |
| **Multi-node** | Yes | No (single node) | Yes |
| **Production-ready** | Yes | No | Yes |
| **Managed** | Self-host | Local | AWS managed |
| **Good for** | Side projects, small teams | Learning, testing | Enterprise, scaling |

---

## 16.24 Common Issues & Debugging

### Pod Status Issues

```
┌─────────────────────────────────────────────────────────────┐
│              Troubleshooting Flow Chart                      │
│                                                              │
│  Pod Status?                                                 │
│  ├── Pending         → Resource issue / scheduling           │
│  ├── ContainerCreating → Image pulling / volume mounting     │
│  ├── CrashLoopBackOff → App crash, cek logs                 │
│  ├── ImagePullBackOff → Wrong image name / auth issue        │
│  ├── OOMKilled        → Memory limit terlalu kecil           │
│  ├── Evicted          → Node kehabisan disk/memory           │
│  └── Running tapi error → Cek app logs & events             │
└─────────────────────────────────────────────────────────────┘
```

#### 1. CrashLoopBackOff

Pod start, crash, restart, crash, restart... loop.

```bash
# Diagnose
kubectl describe pod <pod-name>    # Lihat events
kubectl logs <pod-name>            # Lihat app logs
kubectl logs <pod-name> --previous # Logs dari container yang CRASH sebelumnya

# Common causes:
# - App error (config salah, dependency ga connect)
# - Command salah di CMD/ENTRYPOINT
# - Port conflict
# - Missing environment variable

# Fix example:
# App crash karena DB_HOST ga di-set → tambahkan ConfigMap
# App crash karena port 8080 udah dipake → ubah PORT env var
```

#### 2. ImagePullBackOff

K8s ga bisa pull Docker image.

```bash
# Diagnose
kubectl describe pod <pod-name>

# Events:
# Warning  Failed     15s   kubelet  Failed to pull image "task-manager:v1.0.0":
#          rpc error: code = NotFound desc = failed to pull and unpack image

# Common causes:
# - Image name typo
# - Image ga ada di registry
# - Private registry tapi ga ada imagePullSecret
# - Network issue

# Fix:
# 1. Cek image name dan tag
# 2. Kalau pakai minikube: eval $(minikube docker-env) && docker build
# 3. Kalau private registry: bikin imagePullSecret
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=username \
  --docker-password=token
```

#### 3. Pending Pod

Pod ga bisa di-schedule ke node manapun.

```bash
# Diagnose
kubectl describe pod <pod-name>

# Events:
# Warning  FailedScheduling  default-scheduler  0/1 nodes are available:
#   1 Insufficient cpu.

# Common causes:
# - Resource requests terlalu besar (ga ada node yang cukup)
# - PVC ga bisa bound (storage ga tersedia)
# - Node selector/affinity ga match
# - Taints/tolerations

# Fix:
# 1. Kurangi resource requests
# 2. Tambah node / resource
# 3. Check PVC status: kubectl get pvc
```

#### 4. OOMKilled

Container kehabisan memory (melebihi limit).

```bash
# Diagnose
kubectl describe pod <pod-name>

# Last State: Terminated
# Reason: OOMKilled
# Exit Code: 137

# Fix:
# 1. Naikkan memory limit
resources:
  limits:
    memory: "512Mi"    # dari 256Mi ke 512Mi

# 2. Investigate memory leak di app
# - Go: pakai pprof untuk check memory usage
# - Cek goroutine leak
```

### Debugging Commands

```bash
# ========== QUICK DIAGNOSTICS ==========

# 1. Status semua resources
kubectl get all -n task-manager

# 2. Events (kronologi apa yang terjadi)
kubectl get events -n task-manager --sort-by='.lastTimestamp'

# 3. Pod detail (termasuk events)
kubectl describe pod <pod-name> -n task-manager

# 4. Logs
kubectl logs <pod-name> -n task-manager
kubectl logs <pod-name> -n task-manager --previous   # previous crash
kubectl logs -l app=task-manager -n task-manager      # semua pods

# 5. Shell ke pod
kubectl exec -it <pod-name> -n task-manager -- /bin/sh

# 6. DNS check (dari dalam pod)
kubectl exec -it <pod-name> -- nslookup postgres
kubectl exec -it <pod-name> -- nslookup redis

# 7. Network check (dari dalam pod)
kubectl exec -it <pod-name> -- wget -qO- http://postgres:5432 || echo "connection test"
kubectl exec -it <pod-name> -- wget -qO- http://task-manager-svc/health

# 8. Resource usage
kubectl top nodes                       # Node resource usage
kubectl top pods -n task-manager        # Pod resource usage

# 9. Check endpoints (service → pod mapping)
kubectl get endpoints -n task-manager

# 10. Debug pod (ephemeral debug container)
kubectl debug -it <pod-name> --image=busybox -n task-manager
```

### Debugging Checklist

```
Pod ga running:
━━━━━━━━━━━━━━
[ ] kubectl get pods — status apa?
[ ] kubectl describe pod — ada events error?
[ ] kubectl logs — ada error di app?
[ ] kubectl logs --previous — ada crash sebelumnya?
[ ] Image name dan tag benar?
[ ] Resource limits cukup?
[ ] ConfigMap/Secret exist?
[ ] PVC bound?

Service ga bisa diakses:
━━━━━━━━━━━━━━━━━━━━━━━
[ ] kubectl get svc — service exist?
[ ] kubectl get endpoints — ada endpoint? (kalau kosong = selector ga match)
[ ] kubectl describe svc — selector benar?
[ ] Labels di pod match dengan selector di service?
[ ] Port mapping benar? (port vs targetPort)
[ ] Pod running dan ready?

Ingress ga jalan:
━━━━━━━━━━━━━━━━
[ ] Ingress Controller installed? (kubectl get pods -n ingress-nginx)
[ ] kubectl get ingress — address ada?
[ ] Host di /etc/hosts (minikube)?
[ ] Service exist dan punya endpoints?
[ ] IngressClassName benar?
```

---

## 16.25 Best Practices

### Resource Management

```yaml
# SELALU set resource requests dan limits
# Tanpa ini, pod bisa pakai resource unlimited → ganggu pod lain

resources:
  requests:
    cpu: "100m"       # Scheduling: K8s pastikan node punya ini
    memory: "128Mi"
  limits:
    cpu: "500m"       # Enforcement: throttle kalau lebih
    memory: "256Mi"   # Enforcement: OOMKill kalau lebih
```

### Health Probes

```yaml
# SELALU set liveness dan readiness probe
# Tanpa ini, K8s ga tau app kamu healthy atau engga

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 3
  periodSeconds: 5
```

### PodDisruptionBudget (PDB)

PDB pastikan **minimum pods yang harus running** saat maintenance/upgrade.

```yaml
# k8s/pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: task-manager-pdb
  namespace: task-manager
spec:
  minAvailable: 2         # Minimal 2 pods harus running
  # ATAU
  # maxUnavailable: 1     # Maksimal 1 pod boleh down
  selector:
    matchLabels:
      app: task-manager
```

### Labels dan Annotations

```yaml
# Labels — untuk organize dan select resources
metadata:
  labels:
    app: task-manager          # Nama app
    tier: backend              # Layer (backend/frontend/database)
    version: v1.0.0            # Version
    environment: production    # Environment
    team: backend-team         # Team owner

# Annotations — untuk metadata/config tambahan
metadata:
  annotations:
    description: "Task Manager Go API"
    contact: "backend-team@company.com"
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
```

### Security Best Practices

```yaml
# 1. Non-root user
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
    - name: task-manager
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop:
            - ALL

# 2. NetworkPolicy — restrict traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: task-manager-netpol
  namespace: task-manager
spec:
  podSelector:
    matchLabels:
      app: task-manager
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx    # Cuma dari Ingress Controller
      ports:
        - port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres          # Cuma ke PostgreSQL
      ports:
        - port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis             # Dan Redis
      ports:
        - port: 6379
```

### GitOps (ArgoCD Concept)

```
GitOps = "Git sebagai single source of truth"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Traditional: Developer → kubectl apply → Cluster
GitOps:      Developer → git push → ArgoCD detects → Auto apply ke Cluster

Flow:
1. K8s manifests disimpan di Git repo
2. ArgoCD watch Git repo
3. Developer push perubahan ke Git
4. ArgoCD detect perubahan
5. ArgoCD apply ke cluster secara otomatis
6. Kalau ada drift (manual change), ArgoCD auto-revert

Benefit:
- Audit trail (git log = deployment history)
- Rollback = git revert
- Pull request = deployment approval
- Reproducible (clone repo, apply, done)

Tools:
- ArgoCD (paling populer)
- Flux (alternatif)
```

### RBAC (Role-Based Access Control)

```yaml
# Role — define permission di namespace
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer-role
  namespace: development
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps"]
    verbs: ["get", "list", "watch", "create", "update"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch", "create", "update"]
  # TIDAK ada permission untuk delete atau akses secrets

---
# RoleBinding — assign role ke user
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developer-binding
  namespace: development
subjects:
  - kind: User
    name: junior-dev@company.com
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: developer-role
  apiGroup: rbac.authorization.k8s.io
```

### Best Practices Summary

```
WAJIB:
━━━━━
[ ] Resource requests & limits di setiap container
[ ] Liveness + readiness probe di setiap container
[ ] Non-root user (security)
[ ] Labels yang konsisten (app, tier, version, environment)
[ ] Namespace per environment (dev/staging/prod)
[ ] Secrets di K8s Secret (bukan hardcode di YAML)
[ ] PDB untuk high-availability apps
[ ] Image tag specific (bukan :latest di production!)

RECOMMENDED:
━━━━━━━━━━━
[ ] NetworkPolicy untuk restrict traffic
[ ] RBAC untuk access control
[ ] GitOps (ArgoCD/Flux) untuk deployment
[ ] HPA untuk auto-scaling
[ ] Helm chart untuk package management
[ ] Pod anti-affinity (spread pods across nodes)
[ ] Monitoring dengan Prometheus + Grafana
```

---

## 16.26 Docker Compose → K8s Migration Cheat Sheet

Kalau kamu udah punya docker-compose.yml dan mau migrate ke K8s, ini mapping-nya:

```
┌──────────────────────────┬──────────────────────────────────┐
│   Docker Compose          │   Kubernetes                      │
├──────────────────────────┼──────────────────────────────────┤
│ services:                │ Deployment + Service              │
│   api:                   │                                    │
│     image: myapp:latest  │ Deployment.spec.template          │
│                          │   .spec.containers[].image        │
├──────────────────────────┼──────────────────────────────────┤
│ services:                │                                    │
│   api:                   │ Service                            │
│     ports:               │   spec.ports                      │
│       - "8080:8080"      │   type: NodePort/LoadBalancer     │
├──────────────────────────┼──────────────────────────────────┤
│ environment:             │ ConfigMap (non-sensitive)          │
│   - APP_ENV=prod         │ Secret (sensitive)                │
│   - DB_PASSWORD=xxx      │                                    │
├──────────────────────────┼──────────────────────────────────┤
│ env_file:                │ ConfigMap / Secret                 │
│   - .env                 │ (envFrom in Deployment)           │
├──────────────────────────┼──────────────────────────────────┤
│ volumes:                 │ PersistentVolumeClaim (PVC)       │
│   - pgdata:/var/lib/...  │ volumeClaimTemplates (StatefulSet)│
├──────────────────────────┼──────────────────────────────────┤
│ depends_on:              │ initContainers / readiness probe  │
│   postgres:              │ (K8s ga punya depends_on,         │
│     condition: healthy   │  tapi probe handle ini)           │
├──────────────────────────┼──────────────────────────────────┤
│ networks:                │ Service (DNS-based discovery)     │
│   - task-network         │ NetworkPolicy (isolation)         │
├──────────────────────────┼──────────────────────────────────┤
│ restart: always          │ Deployment restartPolicy: Always  │
│                          │ (default)                         │
├──────────────────────────┼──────────────────────────────────┤
│ healthcheck:             │ livenessProbe / readinessProbe    │
│   test: wget...          │ httpGet / exec / tcpSocket        │
├──────────────────────────┼──────────────────────────────────┤
│ build:                   │ Ga ada di K8s. Build image        │
│   context: .             │ di CI/CD pipeline, push ke        │
│   dockerfile: Dockerfile │ registry, K8s pull image.         │
├──────────────────────────┼──────────────────────────────────┤
│ deploy:                  │ Deployment.spec                   │
│   replicas: 3            │   replicas: 3                     │
├──────────────────────────┼──────────────────────────────────┤
│ (ga ada equivalent)      │ Ingress (HTTP routing)            │
│                          │ HPA (auto-scaling)                │
│                          │ Namespace (isolation)             │
│                          │ RBAC (access control)             │
│                          │ StatefulSet (for DBs)             │
└──────────────────────────┴──────────────────────────────────┘
```

### Docker Compose → K8s Conversion Example

Docker Compose:

```yaml
# docker-compose.yml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_PASSWORD=secret
    depends_on:
      postgres:
        condition: service_healthy
    restart: always

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_PASSWORD=secret
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: pg_isready -U postgres

volumes:
  pgdata:
```

Kubernetes equivalent (5 files):

```
k8s/
├── configmap.yaml      ← environment (non-sensitive)
├── secret.yaml         ← environment (sensitive: DB_PASSWORD)
├── deployment.yaml     ← api service + healthcheck
├── service.yaml        ← ports
└── postgres.yaml       ← postgres service + volume
```

---

## 16.27 Cheat Sheet

### kubectl Commands

```bash
# ============================================================
# KUBERNETES CHEAT SHEET
# ============================================================

# ========== CLUSTER INFO ==========
kubectl cluster-info                   # Cluster info
kubectl get nodes                      # List nodes
kubectl get nodes -o wide              # Nodes + detail
kubectl top nodes                      # Node resource usage
kubectl config current-context         # Current cluster
kubectl config get-contexts            # All clusters
kubectl config use-context <name>      # Switch cluster

# ========== NAMESPACE ==========
kubectl get namespaces                 # List namespaces
kubectl create namespace <name>        # Create namespace
kubectl config set-context --current --namespace=<name>  # Set default

# ========== PODS ==========
kubectl get pods                       # List pods
kubectl get pods -o wide               # Pods + node, IP
kubectl get pods -A                    # All namespaces
kubectl get pods -l app=task-manager   # Filter by label
kubectl get pods -w                    # Watch (real-time)
kubectl describe pod <name>            # Detail + events
kubectl logs <name>                    # View logs
kubectl logs -f <name>                 # Follow logs
kubectl logs --previous <name>         # Previous crash logs
kubectl logs -l app=task-manager       # Logs all pods with label
kubectl exec -it <name> -- /bin/sh     # Shell into pod
kubectl delete pod <name>              # Delete pod
kubectl top pods                       # Pod resource usage

# ========== DEPLOYMENTS ==========
kubectl get deployments                # List deployments
kubectl describe deployment <name>     # Detail
kubectl apply -f deployment.yaml       # Create/update
kubectl delete deployment <name>       # Delete
kubectl scale deploy <name> --replicas=5  # Scale
kubectl set image deploy/<name> <container>=<image>  # Update image
kubectl rollout status deploy/<name>   # Rollout status
kubectl rollout history deploy/<name>  # Rollout history
kubectl rollout undo deploy/<name>     # Rollback
kubectl rollout undo deploy/<name> --to-revision=1  # Rollback specific

# ========== SERVICES ==========
kubectl get services                   # List services
kubectl describe service <name>        # Detail
kubectl get endpoints                  # Service → Pod mapping
kubectl port-forward svc/<name> 8080:80  # Port forward

# ========== CONFIG ==========
kubectl get configmaps                 # List ConfigMaps
kubectl get secrets                    # List Secrets
kubectl describe configmap <name>      # Detail ConfigMap
kubectl get secret <name> -o yaml      # Secret values (base64)
kubectl create secret generic <name> --from-literal=key=value

# ========== STORAGE ==========
kubectl get pv                         # PersistentVolumes
kubectl get pvc                        # PersistentVolumeClaims
kubectl get storageclass               # StorageClasses

# ========== INGRESS ==========
kubectl get ingress                    # List Ingress
kubectl describe ingress <name>        # Detail Ingress

# ========== HPA ==========
kubectl get hpa                        # HorizontalPodAutoscaler
kubectl describe hpa <name>            # Detail HPA

# ========== DEBUGGING ==========
kubectl get events --sort-by='.lastTimestamp'  # Recent events
kubectl top pods                       # Resource usage
kubectl top nodes                      # Node usage
kubectl debug -it <pod> --image=busybox  # Debug pod

# ========== CLEANUP ==========
kubectl delete -f dir/                 # Delete from files
kubectl delete all --all -n <namespace>  # Delete everything
kubectl delete namespace <name>        # Delete namespace + all
```

### YAML Templates

```yaml
# ===== Deployment Template =====
apiVersion: apps/v1
kind: Deployment
metadata:
  name: MY-APP
  labels:
    app: MY-APP
spec:
  replicas: 3
  selector:
    matchLabels:
      app: MY-APP
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: MY-APP
    spec:
      containers:
        - name: MY-APP
          image: MY-IMAGE:TAG
          ports:
            - containerPort: 8080
          envFrom:
            - configMapRef:
                name: MY-APP-config
            - secretRef:
                name: MY-APP-secret
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "256Mi"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 5

---
# ===== Service Template =====
apiVersion: v1
kind: Service
metadata:
  name: MY-APP-svc
spec:
  type: ClusterIP
  selector:
    app: MY-APP
  ports:
    - port: 80
      targetPort: 8080

---
# ===== ConfigMap Template =====
apiVersion: v1
kind: ConfigMap
metadata:
  name: MY-APP-config
data:
  KEY: "value"

---
# ===== Secret Template =====
apiVersion: v1
kind: Secret
metadata:
  name: MY-APP-secret
type: Opaque
stringData:
  KEY: "sensitive-value"

---
# ===== Ingress Template =====
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: MY-APP-ingress
spec:
  ingressClassName: nginx
  rules:
    - host: MY-DOMAIN.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: MY-APP-svc
                port:
                  number: 80

---
# ===== HPA Template =====
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: MY-APP-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: MY-APP
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### Common Patterns

```bash
# Pattern 1: Deploy full app stack
kubectl create namespace myapp
kubectl apply -f k8s/ -n myapp

# Pattern 2: Quick test — run pod sementara
kubectl run test-pod --image=busybox --rm -it -- /bin/sh

# Pattern 3: DNS test dari dalam cluster
kubectl run dns-test --image=busybox --rm -it -- nslookup postgres

# Pattern 4: Port forward untuk local testing
kubectl port-forward svc/task-manager-svc 8080:80 -n task-manager

# Pattern 5: Restart semua pods (force new pods)
kubectl rollout restart deployment/task-manager -n task-manager

# Pattern 6: Check apa yang akan di-apply (dry run)
kubectl apply -f k8s/deployment.yaml --dry-run=client -o yaml

# Pattern 7: Generate YAML dari imperative command
kubectl create deployment nginx --image=nginx --replicas=3 --dry-run=client -o yaml > deployment.yaml

# Pattern 8: Copy file dari/ke pod
kubectl cp local-file.txt <pod>:/path/in/pod
kubectl cp <pod>:/path/in/pod local-file.txt
```

---

## 16.28 Checklist Bab Ini

### Konsep Dasar

- [ ] Paham apa itu Kubernetes dan kenapa penting
- [ ] Paham bedanya Docker Compose vs Kubernetes
- [ ] Paham K8s architecture (Control Plane vs Worker Nodes)
- [ ] Paham komponen Control Plane (API Server, etcd, Scheduler, Controller Manager)
- [ ] Paham komponen Worker Node (kubelet, kube-proxy, container runtime)

### Core Objects

- [ ] Paham hierarchy: Deployment → ReplicaSet → Pod → Container
- [ ] Paham apa itu Pod (unit terkecil, wrapper container)
- [ ] Paham apa itu Deployment (declarative pod management)
- [ ] Paham apa itu Service (stable network endpoint, 3 tipe)
- [ ] Paham apa itu ConfigMap dan Secret
- [ ] Paham apa itu Ingress (HTTP routing)
- [ ] Paham apa itu Namespace (isolation)
- [ ] Paham PersistentVolume, PVC, dan StorageClass

### Tools

- [ ] kubectl ter-install dan bisa dipakai
- [ ] minikube ter-install dan bisa start cluster
- [ ] Paham minikube addons (ingress, metrics-server)
- [ ] Helm ter-install

### kubectl Skills

- [ ] Bisa `kubectl get` (pods, deployments, services, dll)
- [ ] Bisa `kubectl describe` untuk lihat detail + events
- [ ] Bisa `kubectl apply -f` untuk create/update resources
- [ ] Bisa `kubectl delete` untuk hapus resources
- [ ] Bisa `kubectl logs` untuk lihat container logs
- [ ] Bisa `kubectl exec -it` untuk shell ke container
- [ ] Bisa `kubectl port-forward` untuk local testing
- [ ] Paham context dan namespace

### YAML Manifests

- [ ] Bisa nulis Pod YAML
- [ ] Bisa nulis Deployment YAML (replicas, strategy, probes)
- [ ] Bisa nulis Service YAML (ClusterIP, NodePort)
- [ ] Bisa nulis ConfigMap dan Secret YAML
- [ ] Bisa nulis Ingress YAML
- [ ] Bisa nulis PV/PVC YAML
- [ ] Bisa nulis HPA YAML

### Full Stack Deploy

- [ ] PostgreSQL deployed pakai StatefulSet
- [ ] Redis deployed pakai Deployment
- [ ] Task Manager Go app deployed pakai Deployment
- [ ] Service created untuk semua components
- [ ] ConfigMap dan Secret created
- [ ] Ingress configured
- [ ] HPA configured
- [ ] Semua pods running dan healthy
- [ ] Bisa akses API via port-forward atau Ingress

### Scaling & Updates

- [ ] Bisa manual scale (`kubectl scale`)
- [ ] HPA configured dan working
- [ ] Bisa rolling update (change image version)
- [ ] Bisa monitor rollout status
- [ ] Bisa rollback ke versi sebelumnya
- [ ] Paham konsep blue-green dan canary deployment

### Health Checks

- [ ] Paham bedanya liveness, readiness, startup probe
- [ ] Paham 3 metode probe (httpGet, exec, tcpSocket)
- [ ] Probes configured di Deployment
- [ ] /health endpoint connected ke probes

### Monitoring

- [ ] Paham monitoring stack di K8s (Prometheus Operator, kube-state-metrics)
- [ ] Bisa install monitoring stack via Helm
- [ ] Bisa akses Grafana dashboard
- [ ] ServiceMonitor configured untuk app

### Helm

- [ ] Paham konsep Helm (chart, values, release)
- [ ] Bisa install chart dari repository
- [ ] Bisa create custom Helm chart
- [ ] Bisa upgrade dan rollback release

### Security & Best Practices

- [ ] Resource requests dan limits di setiap container
- [ ] Health probes di setiap container
- [ ] Non-root user
- [ ] Consistent labels
- [ ] Namespace per environment
- [ ] Paham RBAC concept
- [ ] Paham NetworkPolicy concept
- [ ] Paham GitOps concept (ArgoCD)

### Debugging

- [ ] Bisa diagnose CrashLoopBackOff
- [ ] Bisa diagnose ImagePullBackOff
- [ ] Bisa diagnose Pending pod
- [ ] Bisa diagnose OOMKilled
- [ ] Paham debugging flow chart
- [ ] Bisa pakai `kubectl describe`, `logs`, `exec` untuk debug

### Production Readiness

- [ ] Paham managed K8s options (EKS, GKE, AKS)
- [ ] Paham k3s untuk lightweight deployment
- [ ] Paham cost considerations
- [ ] Paham when to use Docker Compose vs K8s vs managed K8s

---

## Stage 3 Dimulai: Advanced

Selamat! Kamu baru aja menyelesaikan **bab terpanjang dan terpenting** di Stage 3. Kubernetes itu skill yang **game-changer** di karier DevOps. Dengan K8s, kamu bisa:

```
Yang bisa kamu lakuin sekarang:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Deploy app di K8s cluster (minikube / k3s / EKS)
2. Scale otomatis berdasarkan traffic (HPA)
3. Zero-downtime deployment (rolling update)
4. Rollback kalau ada bug (kubectl rollout undo)
5. Service discovery antar services (DNS)
6. Manage config dan secrets secara proper
7. HTTP routing dengan Ingress
8. Persistent storage untuk database
9. Monitor K8s cluster (Prometheus + Grafana)
10. Package apps dengan Helm
11. Debug K8s issues

Journey so far:
━━━━━━━━━━━━━━
Bab 1-10:  Backend Developer (Go API)
Bab 11-15: DevOps Core (Docker, CI/CD, AWS, Nginx, Monitoring)
Bab 16:    Container Orchestration (Kubernetes) ← KAMU DI SINI

What's next:
━━━━━━━━━━━
Bab 17: Terraform — Infrastructure as Code
Bab 18: Microservices — Distributed System Patterns
Bab 19: Performance — Load Testing & Optimization
Bab 20: Security — Best Practices
```

Dari developer yang cuma bisa jalanin `docker compose up`, sekarang kamu bisa **orchestrate containers across multiple servers** dengan auto-scaling, self-healing, dan zero-downtime deployment. Itu skill yang dicari di **setiap** job posting DevOps/SRE/Platform Engineer.

> **Next up: Bab 17 — Terraform.** Kalau Kubernetes itu orchestrate containers, Terraform itu orchestrate **infrastructure** — bikin VPC, EC2, RDS, EKS, semuanya dari kode. Infrastructure as Code. No more clicking buttons in AWS Console.

---
