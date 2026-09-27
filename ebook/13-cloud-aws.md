# Bab 13: Cloud Deployment — AWS dari Nol Sampai Production

## Kenapa Bab Ini Penting?

Ini dia momen yang kamu tunggu-tunggu. Setelah 12 bab — dari belajar Go, bikin REST API, Docker, CI/CD — sekarang saatnya **deploy beneran ke cloud**. Bukan localhost. Bukan Docker di laptop. Tapi server di **Amazon Web Services** yang bisa diakses siapa aja dari mana aja di dunia.

Kenapa ini game changer? Karena:

```
Sebelum bab ini:
- API cuma jalan di localhost:8080
- Database PostgreSQL di laptop
- Redis di laptop
- "Production-ready" tapi ga ada production server

Setelah bab ini:
- API jalan di EC2 instance (server di cloud)
- PostgreSQL di RDS (managed, auto-backup, auto-patching)
- Redis di ElastiCache (managed, auto-failover)
- Accessible dari https://api.yourdomain.com
- CI/CD push ke GitHub → auto deploy ke AWS
```

Di bab ini, kita bakal deploy **seluruh stack Task Manager API** ke AWS — dari bikin akun sampai API bisa diakses publik. Step by step. Ga ada yang di-skip.

> **Analogi gampang:** Selama ini kamu masak di **dapur rumah** (localhost). Sekarang kamu mau buka **restoran beneran** (cloud). Kamu butuh sewa tempat (EC2), kulkas industrial (RDS), dispenser es (ElastiCache), gudang bahan (S3), dan sistem keamanan (IAM + Security Group). AWS itu landlord-nya — mereka sediain semuanya, kamu tinggal bayar sesuai pemakaian.

---

## 13.1 Apa Itu Cloud Computing?

### On-Premise vs Cloud

Sebelum cloud computing, kalau kamu mau bikin server, kamu harus:

```
On-Premise (Traditional):
━━━━━━━━━━━━━━━━━━━━━━━
1. Beli server fisik (Dell PowerEdge, $5,000+)
2. Cari data center / server room
3. Pasang rack, kabel, UPS, AC
4. Install OS (Ubuntu Server)
5. Setup networking (IP, firewall, DNS)
6. Install software (PostgreSQL, Redis, Nginx)
7. Setup backup, monitoring, security
8. Maintain hardware (disk rusak, RAM error)
9. Scale up? Beli server baru. Tunggu 2-4 minggu delivery.

Total setup time: 2-6 minggu
Total cost: $10,000+ upfront
Ongoing: $500-2000/bulan (listrik, internet, maintenance)
```

```
Cloud Computing (AWS):
━━━━━━━━━━━━━━━━━━━━
1. Bikin akun AWS (5 menit)
2. Launch EC2 instance (2 menit)
3. Server ready, accessible via SSH
4. Scale up? Klik tombol, 30 detik server baru
5. Scale down? Matiin instance, stop bayar
6. Hardware rusak? AWS yang handle

Total setup time: 10 menit
Total cost: $0 upfront (free tier!)
Ongoing: Pay-as-you-go ($5-50/bulan untuk project kecil)
```

### Analogi: Sewa Apartemen vs Bangun Rumah

```
On-Premise = Bangun Rumah Sendiri:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Beli tanah (hardware)
- Bayar arsitek (system architect)
- Bangun dari nol (setup server)
- Maintenance sendiri (atap bocor, pipa pecah)
- Mau renovasi? Ribet dan mahal
- Tapi rumah itu milik kamu sepenuhnya

Cloud = Sewa Apartemen:
━━━━━━━━━━━━━━━━━━━━━
- Tinggal pilih unit (launch instance)
- Langsung masuk, udah siap huni (server ready)
- AC rusak? Manajemen yang benerin (managed services)
- Mau pindah ke unit lebih besar? Tinggal request
- Bayar bulanan sesuai ukuran unit
- Tapi unit itu bukan milik kamu
```

### IaaS vs PaaS vs SaaS

Cloud computing ada **3 model utama**. Penting banget kamu paham ini karena bakal muncul di interview:

```
┌─────────────────────────────────────────────────────────────┐
│                    Cloud Service Models                       │
│                                                              │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐            │
│  │          │     │          │     │          │             │
│  │   SaaS   │     │   PaaS   │     │   IaaS   │            │
│  │          │     │          │     │          │             │
│  │ Gmail    │     │ Heroku   │     │ AWS EC2  │            │
│  │ Slack    │     │ Railway  │     │ AWS VPC  │            │
│  │ Notion   │     │ Vercel   │     │ DigitalOcean│          │
│  │          │     │ GCP App  │     │          │             │
│  │ Kamu cuma│     │ Engine   │     │ Kamu manage│           │
│  │ pakai    │     │          │     │ dari OS   │            │
│  │          │     │ Kamu cuma│     │ ke atas   │            │
│  │          │     │ deploy   │     │          │             │
│  │          │     │ code     │     │          │             │
│  └──────────┘     └──────────┘     └──────────┘            │
│                                                              │
│  ════════════════════════════════════════════════            │
│                                                              │
│  Yang kamu manage:                                           │
│                                                              │
│  SaaS:    [Cuma data kamu]                                  │
│  PaaS:    [Code + Data]                                     │
│  IaaS:    [OS + Runtime + Code + Data]                      │
│  On-Prem: [Hardware + Network + OS + Runtime + Code + Data] │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

| Model | Kamu Manage | Provider Manage | Contoh | Analogy |
|-------|------------|----------------|--------|---------|
| **IaaS** | OS, runtime, app, data | Hardware, network, virtualization | AWS EC2, DigitalOcean, GCP Compute | Sewa **tanah kosong** — kamu bangun sendiri |
| **PaaS** | App, data | Hardware sampai runtime | Heroku, Railway, Vercel, GCP App Engine | Sewa **kamar kos** — tinggal bawa barang |
| **SaaS** | Cuma data | Semuanya | Gmail, Slack, Notion, Figma | **Hotel** — tinggal tidur |

**Yang kita pelajari di bab ini:** Kebanyakan **IaaS** (EC2, VPC) + beberapa **managed services** (RDS, ElastiCache) yang posisinya di antara IaaS dan PaaS.

### Kenapa AWS?

Ada banyak cloud provider — kenapa kita pilih AWS?

```
Market Share Cloud Computing (2025):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
AWS         ████████████████████████████████  31%
Azure       ██████████████████████████        25%
Google Cloud ████████████████                 11%
Alibaba     ██████                            5%
Others      ████████████████████████████      28%
```

1. **Market Leader** — AWS punya market share terbesar. Artinya **paling banyak demand di job market**
2. **Paling Mature** — Launch sejak 2006, service paling lengkap (200+ services)
3. **Industri Standard** — Netflix, Airbnb, NASA, bahkan pesaingnya (Slack) pakai AWS
4. **Free Tier** — 12 bulan gratis untuk banyak services. Perfect buat belajar
5. **Dokumentasi** — Paling lengkap dan paling banyak tutorial
6. **Sertifikasi** — AWS certifications paling di-recognize di industry (AWS SAA, AWS DevOps Pro)
7. **Transferable Knowledge** — Konsep di AWS (VPC, IAM, managed DB) sama di semua cloud. Belajar AWS = belajar cloud computing

> **Pro tip:** Di Indonesia, **hampir semua lowongan DevOps** minta pengalaman AWS. Azure biasanya untuk enterprise yang pakai Microsoft stack. GCP biasanya untuk startup yang heavy di data/ML.

---

## 13.2 AWS Overview — Infrastruktur Global

### Region

**Region** itu lokasi fisik di dunia tempat AWS punya data center. Setiap region itu **independent** — outage di satu region ga affect region lain.

```
AWS Regions (yang relevan untuk kita):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ap-southeast-1    Singapore       ← Paling dekat dari Indonesia!
ap-southeast-3    Jakarta         ← Ada di Indonesia!
ap-northeast-1    Tokyo
us-east-1         N. Virginia     ← Region paling lengkap service-nya
us-west-2         Oregon
eu-west-1         Ireland
```

**Pilih region yang paling dekat dengan user kamu.** Untuk project Indonesia → pakai `ap-southeast-1` (Singapore) atau `ap-southeast-3` (Jakarta).

> **Catatan:** `ap-southeast-3` (Jakarta) itu relatif baru, jadi beberapa service mungkin belum available. `ap-southeast-1` (Singapore) lebih lengkap dan latency dari Indonesia masih rendah (~20-30ms). Untuk belajar, pakai **Singapore**.

### Availability Zone (AZ)

Setiap region punya **minimal 2 Availability Zones** (biasanya 3). AZ itu data center yang **terpisah fisik** tapi terhubung lewat network super cepat.

```
Region: ap-southeast-1 (Singapore)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌──────────────────────────────────────────────┐
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │   AZ-1a   │  │   AZ-1b   │  │   AZ-1c   │ │
│  │           │  │           │  │           │  │
│  │  DC 1     │  │  DC 2     │  │  DC 3     │  │
│  │  DC 2     │  │  DC 3     │  │  DC 4     │  │
│  │           │  │           │  │           │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘ │
│        │              │              │        │
│        └──────────────┴──────────────┘        │
│            Low-latency network (<2ms)         │
│                                              │
└──────────────────────────────────────────────┘
```

**Kenapa penting?** High Availability. Kalau satu AZ mati (kebakaran, banjir), app kamu masih jalan di AZ lain.

### Edge Location

**Edge Location** itu mini data center yang dipake untuk **caching content** (CloudFront CDN). Ada 400+ edge locations di seluruh dunia — termasuk di Jakarta.

```
User di Jakarta → Request ke API di Singapore:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Tanpa CDN:  Jakarta ────────── Singapore (~30ms)
Dengan CDN: Jakarta ── Edge Location Jakarta (~5ms)
                       ↓ (cache miss? forward ke Singapore)
```

### AWS Free Tier

AWS punya **Free Tier** yang cukup generous buat belajar:

| Service | Free Tier | Duration |
|---------|----------|----------|
| **EC2** | 750 jam/bulan t2.micro (atau t3.micro di region tertentu) | 12 bulan |
| **RDS** | 750 jam/bulan db.t3.micro, 20GB storage | 12 bulan |
| **ElastiCache** | 750 jam/bulan cache.t3.micro | 12 bulan |
| **S3** | 5GB storage, 20,000 GET, 2,000 PUT | 12 bulan |
| **ECR** | 500MB storage | Always free |
| **CloudWatch** | 10 metrics, 10 alarms, 1M API requests | Always free |
| **Lambda** | 1M requests/bulan, 400,000 GB-seconds | Always free |
| **VPC** | Free (kamu cuma bayar NAT Gateway, VPN, dll) | Always free |
| **IAM** | Free | Always free |

> **PERINGATAN PENTING:** Free tier punya **batas**. Kalau kamu lupa matiin EC2 instance yang running 24/7, setelah 750 jam (= 31 hari) kamu mulai kena charge. **SELALU set billing alarm!** Kita bahas di section Cost Management.

### Bikin Akun AWS

**Step-by-step bikin akun:**

```
1. Buka https://aws.amazon.com/
2. Klik "Create an AWS Account"
3. Isi email, password, account name
4. Pilih "Personal" account
5. Isi billing info (butuh credit card — ga langsung di-charge)
6. Verifikasi phone number
7. Pilih "Basic Support" plan (free)
8. Selesai!
```

> **Tips:** Pakai email khusus untuk AWS, jangan email pribadi. Kalau kamu punya domain sendiri, bikin email kayak `aws@yourdomain.com`. Ini best practice untuk separation of concerns.

> **Soal credit card:** AWS butuh credit card untuk verifikasi, tapi ga langsung charge selama kamu masih dalam Free Tier. Pakai virtual credit card (Jenius, Bank Jago) kalau kamu nervous. Set billing alarm segera setelah bikin akun!

---

## 13.3 AWS Services yang Perlu Diketahui

AWS punya 200+ services. **Jangan overwhelmed.** Untuk deploy Task Manager API, kamu cuma butuh ~10 services. Sisanya bisa dipelajari nanti kalau dibutuhin.

### Map Services ke Project Kita

```
Task Manager API Architecture:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

┌─────────────────── AWS Cloud ───────────────────────────┐
│                                                          │
│  ┌────────────── VPC (10.0.0.0/16) ──────────────────┐  │
│  │                                                    │  │
│  │  ┌── Public Subnet ──────────────────────────┐    │  │
│  │  │                                            │    │  │
│  │  │   ┌──────────┐         ┌──────────────┐   │    │  │
│  │  │   │   EC2     │         │ NAT Gateway  │   │    │  │
│  │  │   │ (Go API)  │         │              │   │    │  │
│  │  │   │ t2.micro  │         └──────────────┘   │    │  │
│  │  │   └────┬──────┘                            │    │  │
│  │  │        │                                   │    │  │
│  │  └────────┼───────────────────────────────────┘    │  │
│  │           │                                        │  │
│  │  ┌── Private Subnet ─────────────────────────┐    │  │
│  │  │        │                                   │    │  │
│  │  │   ┌────▼──────┐     ┌──────────────┐      │    │  │
│  │  │   │   RDS      │     │ ElastiCache   │     │    │  │
│  │  │   │ PostgreSQL │     │ Redis         │     │    │  │
│  │  │   │ db.t3.micro│     │ cache.t3.micro│     │    │  │
│  │  │   └───────────┘     └──────────────┘      │    │  │
│  │  │                                            │    │  │
│  │  └────────────────────────────────────────────┘    │  │
│  │                                                    │  │
│  └────────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │   ECR    │  │ Route 53 │  │CloudWatch│              │
│  │ (images) │  │  (DNS)   │  │(monitor) │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                                                          │
│  ┌──────────┐  ┌──────────┐                             │
│  │   S3     │  │   IAM    │                             │
│  │(storage) │  │(security)│                             │
│  └──────────┘  └──────────┘                             │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Comparison: AWS Service vs Self-Hosted

| Kebutuhan | AWS Service | Self-Hosted Equivalent | Kenapa Pakai AWS |
|-----------|------------|----------------------|-----------------|
| **Server** | EC2 | VPS (DigitalOcean, Linode) | Auto-scaling, flexible instance types |
| **Database** | RDS PostgreSQL | Install PostgreSQL di server | Auto-backup, auto-patching, replicas |
| **Cache** | ElastiCache Redis | Install Redis di server | Auto-failover, managed upgrades |
| **Storage** | S3 | NFS, MinIO | 99.999999999% durability, cheap |
| **Network** | VPC | Physical network, pfSense | Software-defined, flexible |
| **Security** | IAM | Linux users + SSH keys | Fine-grained permissions, auditable |
| **DNS** | Route 53 | BIND, Cloudflare | Integrated dengan AWS services |
| **Monitoring** | CloudWatch | Prometheus + Grafana | Built-in, auto-integrated |
| **Container Registry** | ECR | Docker Registry, Harbor | Integrated dengan ECS/EKS |
| **Container Orchestration** | ECS / EKS | Docker Swarm, K8s manual | Managed, auto-scaling |

### Service Summary

| Service | Kategori | Fungsi | Free Tier? |
|---------|---------|--------|-----------|
| **EC2** | Compute | Virtual server | 750 jam/bulan t2.micro |
| **RDS** | Database | Managed database | 750 jam/bulan db.t3.micro |
| **ElastiCache** | Cache | Managed Redis/Memcached | 750 jam/bulan cache.t3.micro |
| **S3** | Storage | Object storage | 5GB |
| **VPC** | Network | Virtual network | Free |
| **IAM** | Security | User/role management | Free |
| **Route 53** | DNS | Domain name service | $0.50/hosted zone |
| **CloudWatch** | Monitoring | Logs, metrics, alarms | 10 metrics free |
| **ECR** | Container | Docker image registry | 500MB free |
| **ECS** | Container | Container orchestration | Free (bayar EC2-nya) |
| **EKS** | Container | Managed Kubernetes | $0.10/jam (~$72/bulan) |

> **ECS vs EKS:** ECS itu AWS-native container orchestration (simple, integrated). EKS itu managed Kubernetes (standard industry, complex). Untuk project kita, kita deploy manual dulu ke EC2. ECS/EKS kita bahas di bab Kubernetes.

---

## 13.4 IAM — Identity & Access Management

### Konsep Dasar

IAM itu **security system** di AWS. Dia ngatur **siapa** boleh ngapain di akun AWS kamu. Ini yang **pertama kali** harus kamu setup setelah bikin akun.

```
IAM Hierarchy:
━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│              AWS Account                    │
│              (Root User)                    │
│                                            │
│   ┌──── IAM Group: Developers ──────┐     │
│   │                                  │     │
│   │  ┌─────────┐  ┌─────────┐      │     │
│   │  │ IAM User │  │ IAM User │     │     │
│   │  │  (Aldi)  │  │  (Budi)  │     │     │
│   │  └─────────┘  └─────────┘      │     │
│   │                                  │     │
│   │  Policy: EC2 Full Access         │     │
│   │  Policy: S3 Read Only            │     │
│   └──────────────────────────────────┘     │
│                                            │
│   ┌──── IAM Role: EC2-S3-Access ────┐     │
│   │                                  │     │
│   │  Bisa di-assume oleh EC2         │     │
│   │  Policy: S3 Read/Write           │     │
│   └──────────────────────────────────┘     │
│                                            │
└────────────────────────────────────────────┘
```

### Root Account vs IAM User

| Aspek | Root Account | IAM User |
|-------|-------------|----------|
| **Apa** | Akun utama yang bikin AWS account | User tambahan yang kamu bikin |
| **Power** | Bisa SEMUA hal (termasuk delete akun) | Terbatas sesuai policy |
| **Kapan pakai** | Setup awal, billing, JARANG | **Sehari-hari** |
| **Best practice** | Lock away, MFA, JANGAN pakai daily | Pakai ini untuk kerja |

> **RULE #1 AWS SECURITY: JANGAN PERNAH pakai root account untuk daily use.** Root account itu kayak kunci brankas — disimpan di tempat aman, cuma dikeluarin kalau emergency. Bikin IAM user dengan permission yang dibutuhin, pakai itu untuk kerja sehari-hari.

### Konsep IAM

**1. User** — Representasi orang/service yang akses AWS

**2. Group** — Kumpulan user. Assign policy ke group, semua user di group itu dapet permission-nya

**3. Policy** — JSON document yang define **apa yang boleh dan ga boleh dilakukan**

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:StartInstances",
                "ec2:StopInstances"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Deny",
            "Action": "ec2:TerminateInstances",
            "Resource": "*"
        }
    ]
}
```

Ini artinya: boleh lihat, start, stop EC2 instances. Tapi **ga boleh terminate** (delete).

**4. Role** — Kayak user tapi bukan untuk orang. Untuk **services** (EC2 instance, Lambda function, dll). Role bisa di-"assume" oleh service.

```
Analogi:
- User = KTP kamu (identitas personal)
- Group = Divisi di kantor (semua orang di divisi sama punya akses yang sama)
- Policy = Job description (apa yang boleh kamu lakukan)
- Role = Topi yang bisa dipake bergantian (EC2 "memakai" role untuk akses S3)
```

### Bikin IAM User untuk Daily Use

**Step 1: Login sebagai Root, buka IAM**

```
1. Login ke https://console.aws.amazon.com/ dengan root account
2. Search "IAM" di search bar → klik IAM
3. Di sidebar, klik "Users"
4. Klik "Create user"
```

**Step 2: Create IAM User**

```
1. User name: "aldi-admin" (nama kamu)
2. Centang "Provide user access to the AWS Management Console"
3. Pilih "I want to create an IAM user"
4. Custom password → isi password yang kuat
5. Uncheck "Users must create a new password at next sign-in"
6. Klik "Next"
```

**Step 3: Set Permissions**

```
1. Pilih "Attach policies directly"
2. Search dan centang "AdministratorAccess"
   (ini temporary — nanti kita restrict)
3. Klik "Next"
4. Klik "Create user"
```

**Step 4: Save credentials**

```
1. Catat:
   - Account ID (12 digit angka di pojok kanan atas)
   - IAM user name
   - Password
   - Console sign-in URL: https://<account-id>.signin.aws.amazon.com/console
2. Simpan di password manager (1Password, Bitwarden, dll)
```

**Step 5: Create Access Key (untuk CLI)**

```
1. Klik user yang baru dibuat
2. Tab "Security credentials"
3. Scroll ke "Access keys" → klik "Create access key"
4. Pilih "Command Line Interface (CLI)"
5. Centang acknowledgment
6. Klik "Next" → "Create access key"
7. SIMPAN Access Key ID dan Secret Access Key
   (Secret cuma muncul SEKALI, ga bisa dilihat lagi!)
```

### Setup MFA (Multi-Factor Authentication)

MFA itu **WAJIB** untuk root account dan recommended untuk IAM user.

```
1. IAM Dashboard → klik user kamu
2. Tab "Security credentials"
3. "Multi-factor authentication (MFA)" → "Assign MFA device"
4. Device name: "aldi-phone"
5. Pilih "Authenticator app"
6. Scan QR code pakai Google Authenticator / Authy
7. Masukkin 2 consecutive codes
8. Done!
```

### IAM Best Practices

```
1. JANGAN pakai root account untuk daily use          ← PALING PENTING
2. Enable MFA di root account                         ← WAJIB
3. Bikin IAM user per orang (jangan share credentials)
4. Gunakan groups untuk assign permissions
5. Principle of Least Privilege — kasih permission secukupnya
6. Rotate access keys secara berkala
7. Gunakan IAM roles untuk services (bukan access keys)
8. Review permissions secara berkala
```

### Pakai AWS CLI dengan IAM User

Setelah bikin access key, kamu bisa configure AWS CLI:

```bash
# Configure AWS CLI (kita install nanti di section 13.16)
aws configure

# AWS Access Key ID: AKIA...EXAMPLE
# AWS Secret Access Key: wJalrXUtnFEMI...EXAMPLE
# Default region name: ap-southeast-1
# Default output format: json

# Test
aws sts get-caller-identity
# {
#     "UserId": "AIDA...",
#     "Account": "123456789012",
#     "Arn": "arn:aws:iam::123456789012:user/aldi-admin"
# }
```

> **Dari sini, SELALU login pakai IAM user**. Jangan pakai root lagi kecuali untuk billing dan account-level settings.

---

## 13.5 VPC — Virtual Private Cloud

### Konsep Networking di AWS

VPC itu **network virtual** kamu di AWS. Bayangin kamu punya **gedung kantor sendiri** di cloud — VPC itu gedung-nya. Di dalem gedung itu kamu bisa bikin ruangan-ruangan (subnets), pasang pintu masuk (Internet Gateway), dan atur siapa yang boleh masuk ke ruangan mana (Security Groups).

```
Analogi Gedung Kantor:
━━━━━━━━━━━━━━━━━━━━
VPC              = Gedung kantor
Subnet           = Ruangan di gedung
Internet Gateway = Pintu utama ke jalan raya
NAT Gateway      = Pintu service (buat ruangan dalam bisa akses keluar,
                   tapi orang luar ga bisa masuk)
Security Group   = Satpam di setiap ruangan
NACL             = Satpam di setiap lantai
Route Table      = Papan petunjuk arah
```

### CIDR Block — IP Address Range

VPC butuh range IP address. Ini ditentukan pakai **CIDR notation**:

```
10.0.0.0/16    = 65,536 IP addresses (10.0.0.0 - 10.0.255.255)
10.0.1.0/24    = 256 IP addresses (10.0.1.0 - 10.0.1.255)
10.0.2.0/24    = 256 IP addresses (10.0.2.0 - 10.0.2.255)

/16 → 2^16 = 65,536 addresses (VPC level)
/24 → 2^8  = 256 addresses (subnet level)
/32 → 2^0  = 1 address (single IP)
```

> **Gampangnya:** Angka setelah `/` itu berapa bit yang "fixed". Makin kecil angkanya = makin banyak IP available.

### Architecture VPC untuk Project Kita

```
┌──────────────── VPC: 10.0.0.0/16 ────────────────────────┐
│                                                           │
│  ┌─ Public Subnet 1: 10.0.1.0/24 (AZ-1a) ─────────────┐│
│  │                                                       ││
│  │  EC2 Instance (Go API)                                ││
│  │  IP: 10.0.1.x + Public IP                             ││
│  │                                                       ││
│  └───────────────────────────────────────────────────────┘│
│                                                           │
│  ┌─ Public Subnet 2: 10.0.2.0/24 (AZ-1b) ─────────────┐│
│  │                                                       ││
│  │  (Untuk future: Load Balancer, EC2 cadangan)          ││
│  │                                                       ││
│  └───────────────────────────────────────────────────────┘│
│                                                           │
│  ┌─ Private Subnet 1: 10.0.3.0/24 (AZ-1a) ────────────┐│
│  │                                                       ││
│  │  RDS PostgreSQL                                       ││
│  │  ElastiCache Redis                                    ││
│  │  IP: 10.0.3.x (NO public access)                     ││
│  │                                                       ││
│  └───────────────────────────────────────────────────────┘│
│                                                           │
│  ┌─ Private Subnet 2: 10.0.4.0/24 (AZ-1b) ────────────┐│
│  │                                                       ││
│  │  RDS replica (multi-AZ backup)                        ││
│  │  IP: 10.0.4.x (NO public access)                     ││
│  │                                                       ││
│  └───────────────────────────────────────────────────────┘│
│                                                           │
│  ┌──────────────┐    ┌───────────────┐                   │
│  │   Internet    │    │ NAT Gateway   │                   │
│  │   Gateway     │    │ (optional,    │                   │
│  │  (IGW)        │    │  NOT free!)   │                   │
│  └──────────────┘    └───────────────┘                   │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

### Public Subnet vs Private Subnet

| Aspek | Public Subnet | Private Subnet |
|-------|--------------|----------------|
| **Internet access** | Ya (via Internet Gateway) | Tidak langsung (butuh NAT) |
| **Public IP** | Bisa punya | Ga punya |
| **Accessible dari luar** | Ya | **Tidak** |
| **Diisi apa** | Web server, load balancer | Database, cache, internal services |
| **Kenapa** | Harus diakses user | **Ga boleh** diakses langsung dari internet |

> **Best Practice:** Database dan cache **SELALU** di private subnet. Jangan pernah taruh database di public subnet — itu undangan buat hacker.

### Internet Gateway (IGW)

Internet Gateway itu "pintu" yang menghubungkan VPC kamu ke internet. Tanpa IGW, ga ada resource di VPC yang bisa akses atau diakses dari internet.

### NAT Gateway

NAT Gateway memungkinkan resource di **private subnet** untuk akses internet (misalnya download update), tapi **internet ga bisa akses masuk** ke private subnet.

```
EC2 di public subnet:
Internet ←→ IGW ←→ EC2     (two-way)

RDS di private subnet:
Internet ← NAT ← RDS       (one-way keluar saja)
Internet → X → RDS          (ga bisa masuk!)
```

> **PERINGATAN: NAT Gateway itu MAHAL** — sekitar $32/bulan + data transfer. Untuk project belajar, kamu bisa skip NAT Gateway. RDS dan ElastiCache ga butuh akses internet — mereka cuma perlu bisa diakses dari EC2 di VPC yang sama.

### Security Group vs NACL

| Aspek | Security Group | NACL |
|-------|---------------|------|
| **Level** | Instance (per EC2, per RDS) | Subnet |
| **Type** | Stateful | Stateless |
| **Default** | Deny all inbound, allow all outbound | Allow all |
| **Rules** | Allow only | Allow + Deny |
| **Analogi** | Satpam di depan ruangan | Satpam di lobby lantai |

**Stateful vs Stateless:**
- **Stateful** (Security Group): Kalau request masuk dibolehkan, response otomatis dibolehkan keluar
- **Stateless** (NACL): Harus define rules untuk inbound DAN outbound secara terpisah

Untuk project kita, **Security Group udah cukup**. NACL biasanya dipakai di enterprise environments.

### Bikin VPC lewat Console

**Step 1: Buka VPC Dashboard**

```
1. Login AWS Console
2. Search "VPC" → klik VPC
3. Klik "Create VPC"
```

**Step 2: Pakai VPC Wizard (recommended)**

```
1. Pilih "VPC and more" (wizard yang bikin VPC + subnets + route tables)
2. Name tag: "task-manager"
3. IPv4 CIDR: 10.0.0.0/16
4. Number of Availability Zones: 2
5. Number of public subnets: 2
6. Number of private subnets: 2
7. NAT gateways: None (hemat biaya — kita ga butuh untuk belajar)
8. VPC endpoints: None
9. Klik "Create VPC"
```

Wizard ini otomatis bikin:
- 1 VPC
- 2 public subnets (di 2 AZ berbeda)
- 2 private subnets (di 2 AZ berbeda)
- 1 Internet Gateway
- Route tables

**Step 3: Verify**

```
VPC created:
- VPC: task-manager-vpc (10.0.0.0/16)
- Public Subnet 1: task-manager-subnet-public1-ap-southeast-1a (10.0.0.0/20)
- Public Subnet 2: task-manager-subnet-public2-ap-southeast-1b (10.0.16.0/20)
- Private Subnet 1: task-manager-subnet-private1-ap-southeast-1a (10.0.128.0/20)
- Private Subnet 2: task-manager-subnet-private2-ap-southeast-1b (10.0.144.0/20)
- Internet Gateway: task-manager-igw
- Route Tables: public-rt, private-rt-1, private-rt-2
```

### Bikin VPC lewat AWS CLI

```bash
# Bikin VPC
aws ec2 create-vpc \
    --cidr-block 10.0.0.0/16 \
    --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=task-manager-vpc}]'

# Output: VpcId = vpc-0abc123def456

# Bikin Internet Gateway
aws ec2 create-internet-gateway \
    --tag-specifications 'ResourceType=internet-gateway,Tags=[{Key=Name,Value=task-manager-igw}]'

# Attach IGW ke VPC
aws ec2 attach-internet-gateway \
    --internet-gateway-id igw-0abc123 \
    --vpc-id vpc-0abc123def456

# Bikin Public Subnet 1
aws ec2 create-subnet \
    --vpc-id vpc-0abc123def456 \
    --cidr-block 10.0.1.0/24 \
    --availability-zone ap-southeast-1a \
    --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=task-manager-public-1a}]'

# Bikin Public Subnet 2
aws ec2 create-subnet \
    --vpc-id vpc-0abc123def456 \
    --cidr-block 10.0.2.0/24 \
    --availability-zone ap-southeast-1b \
    --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=task-manager-public-1b}]'

# Bikin Private Subnet 1
aws ec2 create-subnet \
    --vpc-id vpc-0abc123def456 \
    --cidr-block 10.0.3.0/24 \
    --availability-zone ap-southeast-1a \
    --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=task-manager-private-1a}]'

# Bikin Private Subnet 2
aws ec2 create-subnet \
    --vpc-id vpc-0abc123def456 \
    --cidr-block 10.0.4.0/24 \
    --availability-zone ap-southeast-1b \
    --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=task-manager-private-1b}]'

# Bikin Route Table untuk public subnets
aws ec2 create-route-table \
    --vpc-id vpc-0abc123def456 \
    --tag-specifications 'ResourceType=route-table,Tags=[{Key=Name,Value=task-manager-public-rt}]'

# Tambah route ke Internet Gateway
aws ec2 create-route \
    --route-table-id rtb-0abc123 \
    --destination-cidr-block 0.0.0.0/0 \
    --gateway-id igw-0abc123

# Associate public subnets dengan route table
aws ec2 associate-route-table \
    --route-table-id rtb-0abc123 \
    --subnet-id subnet-public-1a

aws ec2 associate-route-table \
    --route-table-id rtb-0abc123 \
    --subnet-id subnet-public-1b

# Enable auto-assign public IP di public subnets
aws ec2 modify-subnet-attribute \
    --subnet-id subnet-public-1a \
    --map-public-ip-on-launch
```

---

## 13.6 EC2 — Elastic Compute Cloud

### Apa Itu EC2?

EC2 itu **virtual server** di AWS. Kayak kamu sewa komputer di cloud — bisa install OS, software, deploy app. Bedanya, kamu ga perlu beli hardware. Tinggal pilih spek, klik launch, 30 detik server ready.

```
EC2 = Virtual Private Server (VPS)

DigitalOcean Droplet  = EC2 Instance
Linode VPS            = EC2 Instance
Vultr Cloud Compute   = EC2 Instance

Bedanya: EC2 punya lebih banyak options dan terintegrasi
         dengan 200+ AWS services lainnya
```

### Instance Types

AWS punya **ratusan** instance types. Yang perlu kamu tau untuk awal:

```
Format: [family][generation].[size]
Contoh: t2.micro

t  = Family (General purpose, burstable)
2  = Generation
micro = Size
```

| Instance Type | vCPU | Memory | Use Case | Harga (Singapore) |
|--------------|------|--------|----------|-------------------|
| **t2.micro** | 1 | 1 GB | **FREE TIER** — dev, small apps | Free / $0.0116/jam |
| t2.small | 1 | 2 GB | Small apps, staging | ~$0.023/jam |
| t2.medium | 2 | 4 GB | Medium apps | ~$0.046/jam |
| t3.micro | 2 | 1 GB | Better burstable | $0.0104/jam |
| t3.small | 2 | 2 GB | Small production | $0.0208/jam |
| m5.large | 2 | 8 GB | General production | ~$0.096/jam |
| c5.large | 2 | 4 GB | Compute-intensive | ~$0.085/jam |

**Untuk project kita:** `t2.micro` — **gratis** 750 jam/bulan selama 12 bulan. Cukup buat Task Manager API kita.

### Instance Family Cheat Sheet

```
t = Turbo/Burstable  → General purpose, bisa burst CPU sementara
                        Cocok untuk: dev, staging, small apps
m = Memory/General   → Balanced CPU + memory
                        Cocok untuk: production apps, database
c = Compute          → CPU-optimized
                        Cocok untuk: compute-heavy workloads
r = RAM              → Memory-optimized
                        Cocok untuk: in-memory database, caching
```

### AMI — Amazon Machine Image

AMI itu **template OS** untuk EC2. Kayak ISO file yang kamu pakai install Windows/Linux.

```
AMI Populer:
- Amazon Linux 2023      ← Recommended (optimized untuk AWS)
- Ubuntu Server 24.04    ← Familiar kalau udah biasa Linux
- Debian 12
- Red Hat Enterprise Linux
- Windows Server 2022
```

Kita pakai **Amazon Linux 2023** karena:
1. Optimized untuk AWS (performance lebih bagus)
2. Gratis (included di EC2 pricing)
3. Package manager: `dnf` (mirip `yum`)
4. Udah include AWS CLI

### Key Pair — SSH Access

Key pair itu **SSH key** untuk login ke EC2 instance. AWS generate public+private key pair — public key disimpan di EC2, private key kamu download.

```
Key Pair = Kunci rumah
- Private key (.pem file) = Kunci yang kamu pegang
- Public key = Gembok yang dipasang di pintu EC2
- Kalau private key hilang = ga bisa masuk ke EC2!
```

### Launch EC2 Instance — Step by Step

**Via AWS Console:**

```
1. AWS Console → EC2 → "Launch Instance"

2. Name: "task-manager-api"

3. AMI: Amazon Linux 2023 (Free tier eligible)

4. Instance type: t2.micro (Free tier eligible)

5. Key pair:
   - Klik "Create new key pair"
   - Name: "task-manager-key"
   - Type: RSA
   - Format: .pem (untuk macOS/Linux)
   - Download file task-manager-key.pem
   - SIMPAN BAIK-BAIK! Ga bisa download ulang!

6. Network settings:
   - VPC: task-manager-vpc (yang kita bikin tadi)
   - Subnet: task-manager-public-1a (PUBLIC subnet!)
   - Auto-assign public IP: Enable
   - Security group: Create new
     - Name: "task-manager-api-sg"
     - Rules:
       - SSH (22) from My IP
       - HTTP (80) from Anywhere
       - HTTPS (443) from Anywhere
       - Custom TCP (8080) from Anywhere

7. Storage: 8 GB gp3 (free tier: up to 30GB)

8. Advanced details: Skip untuk sekarang

9. Klik "Launch instance"
```

**Via AWS CLI:**

```bash
# Create key pair
aws ec2 create-key-pair \
    --key-name task-manager-key \
    --query 'KeyMaterial' \
    --output text > task-manager-key.pem

# Set permission (PENTING — SSH ga mau kalau permission terlalu open)
chmod 400 task-manager-key.pem

# Create security group
aws ec2 create-security-group \
    --group-name task-manager-api-sg \
    --description "Security group for Task Manager API" \
    --vpc-id vpc-0abc123def456

# Add inbound rules
# SSH
aws ec2 authorize-security-group-ingress \
    --group-id sg-0abc123 \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0

# HTTP
aws ec2 authorize-security-group-ingress \
    --group-id sg-0abc123 \
    --protocol tcp \
    --port 80 \
    --cidr 0.0.0.0/0

# HTTPS
aws ec2 authorize-security-group-ingress \
    --group-id sg-0abc123 \
    --protocol tcp \
    --port 443 \
    --cidr 0.0.0.0/0

# App port
aws ec2 authorize-security-group-ingress \
    --group-id sg-0abc123 \
    --protocol tcp \
    --port 8080 \
    --cidr 0.0.0.0/0

# Launch instance
aws ec2 run-instances \
    --image-id ami-0abcdef1234567890 \
    --instance-type t2.micro \
    --key-name task-manager-key \
    --security-group-ids sg-0abc123 \
    --subnet-id subnet-public-1a \
    --associate-public-ip-address \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=task-manager-api}]'
```

### SSH ke EC2 Instance

```bash
# Pastikan .pem file permission-nya bener
chmod 400 task-manager-key.pem

# SSH ke instance
# Amazon Linux 2023 default user: ec2-user
ssh -i task-manager-key.pem ec2-user@<public-ip>

# Contoh:
ssh -i task-manager-key.pem ec2-user@54.169.123.456

# First time connect akan muncul:
# "The authenticity of host... Are you sure you want to continue? (yes/no)"
# Ketik: yes

# Sekarang kamu di dalam EC2 instance!
[ec2-user@ip-10-0-1-123 ~]$
```

### Install Docker di EC2

Setelah SSH ke EC2, install Docker:

```bash
# Update system packages
sudo dnf update -y

# Install Docker
sudo dnf install docker -y

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker  # auto-start saat boot

# Add ec2-user ke docker group (biar ga perlu sudo tiap command)
sudo usermod -aG docker ec2-user

# PENTING: Logout dan login lagi biar group change take effect
exit
# SSH lagi
ssh -i task-manager-key.pem ec2-user@54.169.123.456

# Verify Docker
docker --version
# Docker version 25.x.x

docker run hello-world
# Hello from Docker!

# Install Docker Compose (plugin)
sudo mkdir -p /usr/local/lib/docker/cli-plugins
sudo curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-$(uname -m)" \
    -o /usr/local/lib/docker/cli-plugins/docker-compose
sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

# Verify
docker compose version
# Docker Compose version v2.x.x
```

---

## 13.7 Security Group — Deep Dive

### Konsep

Security Group itu **virtual firewall** untuk EC2 instance (dan services lain). Dia ngatur traffic masuk (inbound) dan keluar (outbound).

```
Internet → [Security Group] → EC2 Instance
              ↑
              │
        "Boleh masuk ga?"
        "Port berapa?"
        "Dari IP mana?"
```

### Rules untuk Task Manager API

Kita butuh beberapa Security Groups:

**1. Security Group untuk EC2 (API Server)**

| Type | Protocol | Port | Source | Kenapa |
|------|----------|------|--------|--------|
| SSH | TCP | 22 | My IP (atau VPN IP) | Remote access |
| HTTP | TCP | 80 | 0.0.0.0/0 (Anywhere) | Web traffic |
| HTTPS | TCP | 443 | 0.0.0.0/0 (Anywhere) | Secure web traffic |
| Custom TCP | TCP | 8080 | 0.0.0.0/0 (Anywhere) | Go API (temporary) |

**2. Security Group untuk RDS (Database)**

| Type | Protocol | Port | Source | Kenapa |
|------|----------|------|--------|--------|
| PostgreSQL | TCP | 5432 | sg-ec2 (Security Group EC2) | **Cuma EC2 yang bisa akses** |

**3. Security Group untuk ElastiCache (Redis)**

| Type | Protocol | Port | Source | Kenapa |
|------|----------|------|--------|--------|
| Custom TCP | TCP | 6379 | sg-ec2 (Security Group EC2) | **Cuma EC2 yang bisa akses** |

### Best Practice: Least Privilege

```
JANGAN:
━━━━━━
Port 5432 (PostgreSQL) → Source: 0.0.0.0/0 (Anywhere)
→ Siapa aja di internet bisa coba connect ke database kamu!
→ Brute force attack, data breach, disaster.

YANG BENER:
━━━━━━━━━━
Port 5432 (PostgreSQL) → Source: sg-ec2-api
→ CUMA EC2 instance yang pakai security group "sg-ec2-api" yang bisa connect.
→ Orang dari internet ga bisa akses langsung.
```

```
JANGAN:
━━━━━━
Port 22 (SSH) → Source: 0.0.0.0/0 (Anywhere)
→ Siapa aja bisa coba SSH ke server kamu.
→ Brute force SSH attack.

YANG BENER:
━━━━━━━━━━
Port 22 (SSH) → Source: 103.123.xxx.xxx/32 (IP kamu doang)
→ CUMA dari IP kamu yang bisa SSH.
→ Kalau IP kamu berubah (ISP dynamic IP), update rule-nya.
```

### Bikin Security Groups via CLI

```bash
# Security Group untuk EC2
aws ec2 create-security-group \
    --group-name task-manager-api-sg \
    --description "SG for Task Manager API EC2" \
    --vpc-id vpc-0abc123

# Security Group untuk RDS
aws ec2 create-security-group \
    --group-name task-manager-rds-sg \
    --description "SG for Task Manager RDS PostgreSQL" \
    --vpc-id vpc-0abc123

# Security Group untuk ElastiCache
aws ec2 create-security-group \
    --group-name task-manager-redis-sg \
    --description "SG for Task Manager ElastiCache Redis" \
    --vpc-id vpc-0abc123

# RDS SG: Allow PostgreSQL from EC2 SG only
aws ec2 authorize-security-group-ingress \
    --group-id sg-rds-123 \
    --protocol tcp \
    --port 5432 \
    --source-group sg-ec2-123

# ElastiCache SG: Allow Redis from EC2 SG only
aws ec2 authorize-security-group-ingress \
    --group-id sg-redis-123 \
    --protocol tcp \
    --port 6379 \
    --source-group sg-ec2-123
```

> **Key insight:** Perhatiin `--source-group` — kita ga pakai IP address, tapi **security group ID**. Ini artinya "semua instance yang pakai security group ini boleh akses". Ini lebih flexible karena ga terikat IP spesifik.

---

## 13.8 RDS — Relational Database Service

### Kenapa RDS, Bukan Install PostgreSQL Sendiri?

Kamu bisa aja install PostgreSQL langsung di EC2. Tapi...

```
Install PostgreSQL di EC2 (self-managed):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Install sendiri ✓
- Configure sendiri ✓
- Backup sendiri (cron job pg_dump) ← bisa lupa
- Patching sendiri (security updates) ← bisa lupa
- Monitoring sendiri ← bisa lupa
- Replication sendiri ← complex
- Failover sendiri ← SANGAT complex
- Disk full? Handle sendiri ← bisa panic
- Upgrade version? Downtime! ← bisa error

RDS (managed):
━━━━━━━━━━━━━
- Launch instance, PostgreSQL ready ✓
- Automated backup (daily, configurable) ✓
- Automated patching ✓
- Monitoring built-in (CloudWatch) ✓
- Multi-AZ deployment (auto-failover) ✓
- Read replicas (scale reads) ✓
- Storage auto-scaling ✓
- Point-in-time recovery (restore ke menit tertentu!) ✓
- Encryption at rest dan in transit ✓
```

**Rule of thumb:** Kalau ada managed service yang affordable, **SELALU pakai managed service**. Kamu bayar sedikit lebih mahal tapi hemat **puluhan jam maintenance** per bulan. Waktu kamu lebih berharga dari $10-20/bulan.

### Launch RDS PostgreSQL

**Via AWS Console:**

```
1. AWS Console → RDS → "Create database"

2. Engine: PostgreSQL

3. Template: "Free tier"
   (Ini otomatis set beberapa settings untuk free tier)

4. Settings:
   - DB instance identifier: "task-manager-db"
   - Master username: "postgres"
   - Master password: [password kuat, CATAT!]
   - Confirm password

5. Instance configuration:
   - DB instance class: db.t3.micro (Free tier eligible)

6. Storage:
   - Type: gp3
   - Allocated storage: 20 GB (Free tier: up to 20GB)
   - Storage autoscaling: Disable (untuk hemat biaya)

7. Connectivity:
   - VPC: task-manager-vpc
   - Subnet group: Bikin baru (atau pakai default)
     → Pilih private subnets kamu
   - Public access: NO (PENTING — JANGAN set Yes!)
   - Security group: task-manager-rds-sg (yang kita bikin tadi)
   - Port: 5432

8. Database authentication: Password authentication

9. Additional configuration:
   - Initial database name: "task_manager"
   - Backup retention: 7 days
   - Monitoring: Enable (basic, free)
   - Maintenance window: Pick off-peak time

10. Klik "Create database"
    (Butuh 5-10 menit sampai status "Available")
```

> **PENTING: Public access = NO.** Database **tidak boleh** bisa diakses dari internet. Cuma EC2 di VPC yang sama yang bisa connect.

**Via AWS CLI:**

```bash
# Create DB subnet group (group of private subnets untuk RDS)
aws rds create-db-subnet-group \
    --db-subnet-group-name task-manager-db-subnet \
    --db-subnet-group-description "Subnets for Task Manager RDS" \
    --subnet-ids subnet-private-1a subnet-private-1b

# Launch RDS instance
aws rds create-db-instance \
    --db-instance-identifier task-manager-db \
    --db-instance-class db.t3.micro \
    --engine postgres \
    --engine-version 16 \
    --master-username postgres \
    --master-user-password "YourStrongPassword123!" \
    --allocated-storage 20 \
    --storage-type gp3 \
    --db-name task_manager \
    --vpc-security-group-ids sg-rds-123 \
    --db-subnet-group-name task-manager-db-subnet \
    --backup-retention-period 7 \
    --no-publicly-accessible \
    --no-multi-az \
    --tags Key=Project,Value=task-manager

# Check status (tunggu sampai "available")
aws rds describe-db-instances \
    --db-instance-identifier task-manager-db \
    --query 'DBInstances[0].DBInstanceStatus'
```

### Connect ke RDS dari EC2

Setelah RDS status "Available", kamu bisa connect dari EC2:

```bash
# SSH ke EC2 dulu
ssh -i task-manager-key.pem ec2-user@54.169.123.456

# Install PostgreSQL client
sudo dnf install postgresql16 -y

# Cari RDS endpoint di AWS Console:
# RDS → task-manager-db → Connectivity & security → Endpoint
# Contoh: task-manager-db.abcdef123456.ap-southeast-1.rds.amazonaws.com

# Connect ke RDS
psql -h task-manager-db.abcdef123456.ap-southeast-1.rds.amazonaws.com \
     -U postgres \
     -d task_manager

# Masukkan password yang kamu set tadi

# Kalau berhasil connect:
task_manager=> SELECT version();
# PostgreSQL 16.x on aarch64-unknown-linux-gnu...

task_manager=> \dt
# (belum ada table — nanti Go API yang bikin via migration)

task_manager=> \q
# Keluar
```

> **Troubleshooting kalau ga bisa connect:**
> 1. Check Security Group RDS — apakah port 5432 allow dari EC2 security group?
> 2. Check apakah RDS di subnet yang sama VPC dengan EC2
> 3. Check apakah RDS udah status "Available"
> 4. Check password — typo sering terjadi

### Free Tier RDS Settings

```
Agar tetap di Free Tier:
━━━━━━━━━━━━━━━━━━━━━━━
- Instance class: db.t3.micro ONLY
- Storage: Max 20 GB
- Multi-AZ: DISABLED (multi-AZ = 2x biaya)
- Read replicas: JANGAN bikin
- Backup retention: 7 hari (included)
- Storage autoscaling: DISABLED
- Monitoring: Basic (free)

Kalau salah satu di atas di-enable:
→ Kamu KELUAR dari free tier
→ Mulai kena charge ~$15-30/bulan
```

---

## 13.9 ElastiCache — Managed Redis

### Kenapa ElastiCache?

Sama kayak RDS vs install PostgreSQL sendiri — ElastiCache itu **managed Redis**. AWS yang handle maintenance, patching, backup, failover.

```
Task Manager API pakai Redis untuk:
1. Session caching
2. Rate limiting
3. Response caching

Tanpa ElastiCache → Install Redis di EC2 → maintenance sendiri
Dengan ElastiCache → Launch cluster → Redis ready, managed
```

### Launch Redis Cluster

**Via AWS Console:**

```
1. AWS Console → ElastiCache → "Create cluster"

2. Cluster engine: Redis

3. Cluster mode: Disabled (untuk single node, free tier)

4. Cluster info:
   - Name: "task-manager-redis"
   - Description: "Redis for Task Manager API"
   - Engine version: 7.x (latest stable)
   - Port: 6379

5. Node type: cache.t3.micro (Free tier eligible)

6. Number of replicas: 0 (untuk free tier, single node)

7. Subnet group:
   - Create new → pilih private subnets
   - Name: "task-manager-redis-subnet"

8. Security:
   - Security group: task-manager-redis-sg

9. Backup: Disable (free tier)

10. Maintenance: Pick off-peak time

11. Klik "Create"
    (Butuh 5-10 menit)
```

**Via AWS CLI:**

```bash
# Create cache subnet group
aws elasticache create-cache-subnet-group \
    --cache-subnet-group-name task-manager-redis-subnet \
    --cache-subnet-group-description "Subnets for Task Manager Redis" \
    --subnet-ids subnet-private-1a subnet-private-1b

# Create Redis cluster
aws elasticache create-cache-cluster \
    --cache-cluster-id task-manager-redis \
    --cache-node-type cache.t3.micro \
    --engine redis \
    --engine-version 7.1 \
    --num-cache-nodes 1 \
    --cache-subnet-group-name task-manager-redis-subnet \
    --security-group-ids sg-redis-123 \
    --tags Key=Project,Value=task-manager

# Check status
aws elasticache describe-cache-clusters \
    --cache-cluster-id task-manager-redis \
    --query 'CacheClusters[0].CacheClusterStatus'
```

### Connect ke ElastiCache dari EC2

```bash
# SSH ke EC2
ssh -i task-manager-key.pem ec2-user@54.169.123.456

# Install redis-cli
sudo dnf install redis6 -y

# Cari ElastiCache endpoint:
# ElastiCache → task-manager-redis → Primary Endpoint
# Contoh: task-manager-redis.abcdef.0001.apse1.cache.amazonaws.com

# Connect
redis-cli -h task-manager-redis.abcdef.0001.apse1.cache.amazonaws.com

# Test
127.0.0.1:6379> PING
PONG

127.0.0.1:6379> SET test "Hello from ElastiCache!"
OK

127.0.0.1:6379> GET test
"Hello from ElastiCache!"

127.0.0.1:6379> exit
```

### Free Tier ElastiCache Settings

```
Agar tetap di Free Tier:
━━━━━━━━━━━━━━━━━━━━━━━
- Node type: cache.t3.micro ONLY
- Number of nodes: 1 (single node)
- Cluster mode: Disabled
- Replicas: 0
- Backup: Disabled

Total free: 750 jam/bulan untuk 12 bulan
= Cukup untuk 1 node running 24/7 selama sebulan
```

> **PENTING:** ElastiCache **tidak bisa diakses dari luar VPC** — ini by design. Kamu ga bisa connect dari laptop langsung. Harus lewat EC2 yang ada di VPC yang sama.

---

## 13.10 S3 — Simple Storage Service

### Apa Itu S3?

S3 itu **object storage** — tempat simpan file (gambar, video, backup, static assets, dll). Beda sama EBS (disk yang di-attach ke EC2) — S3 itu **standalone**, bisa diakses dari mana aja, dan **sangat murah**.

```
S3 Concepts:
━━━━━━━━━━━
Bucket  = Folder utama (globally unique name)
Object  = File yang disimpan di bucket
Key     = Path/nama file di dalam bucket

Contoh:
Bucket: task-manager-assets
Object: uploads/avatar/user-123.jpg
Key:    uploads/avatar/user-123.jpg
URL:    https://task-manager-assets.s3.ap-southeast-1.amazonaws.com/uploads/avatar/user-123.jpg
```

### Use Cases untuk Project Kita

```
1. User avatar uploads
2. Task file attachments
3. Static assets (kalau ada frontend)
4. Database backups (pg_dump)
5. Application logs (archive)
```

### Bikin S3 Bucket

```bash
# Create bucket
# PENTING: Bucket name harus globally unique di seluruh AWS!
aws s3 mb s3://task-manager-assets-aldi --region ap-southeast-1

# List buckets
aws s3 ls

# Upload file
aws s3 cp ./test-file.txt s3://task-manager-assets-aldi/uploads/

# List files di bucket
aws s3 ls s3://task-manager-assets-aldi/

# Download file
aws s3 cp s3://task-manager-assets-aldi/uploads/test-file.txt ./downloaded.txt

# Sync folder
aws s3 sync ./local-folder s3://task-manager-assets-aldi/backup/

# Remove file
aws s3 rm s3://task-manager-assets-aldi/uploads/test-file.txt
```

### Bucket Policy — Access Control

By default, S3 bucket itu **private** (ga bisa diakses dari internet). Kalau mau serve static files publicly:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadForAssets",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::task-manager-assets-aldi/public/*"
        }
    ]
}
```

Ini artinya: siapa aja boleh **baca** (GET) files di folder `public/`. Files di luar folder `public/` tetap private.

> **Best Practice:** Jangan bikin seluruh bucket public. Cuma folder tertentu aja yang perlu public (misalnya `public/`). Sisanya private.

### Presigned URL

Presigned URL itu URL temporary yang punya permission bawaan. Berguna kalau kamu mau:
- User bisa upload file langsung ke S3 (tanpa lewat API server)
- User bisa download private file (tanpa bikin bucket public)

```
Normal flow:
User → API Server → S3 (server jadi middleman, waste bandwidth)

Presigned URL flow:
User → API Server (minta presigned URL)
     ← Presigned URL
User → S3 (upload/download langsung!)
```

### Go SDK — S3 Operations

Install AWS SDK for Go:

```bash
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
```

**Upload file ke S3:**

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
    client     *s3.Client
    bucketName string
}

func NewS3Client(bucketName string) (*S3Client, error) {
    // Load AWS config (dari environment variables atau ~/.aws/credentials)
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("ap-southeast-1"),
    )
    if err != nil {
        return nil, fmt.Errorf("unable to load AWS config: %w", err)
    }

    client := s3.NewFromConfig(cfg)
    return &S3Client{
        client:     client,
        bucketName: bucketName,
    }, nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucketName),
        Key:         aws.String(key),
        Body:        body,
        ContentType: aws.String(contentType),
    })
    if err != nil {
        return fmt.Errorf("failed to upload file: %w", err)
    }
    return nil
}

// GetFileURL returns the public URL of a file
func (s *S3Client) GetFileURL(key string) string {
    return fmt.Sprintf("https://%s.s3.ap-southeast-1.amazonaws.com/%s",
        s.bucketName, key)
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(ctx context.Context, key string) error {
    _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(key),
    })
    if err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    return nil
}
```

**Generate Presigned URL:**

```go
package storage

import (
    "context"
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// GeneratePresignedUploadURL generates a presigned URL for uploading
func (s *S3Client) GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(s.client)

    presignedURL, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(key),
    }, s3.WithPresignExpires(expiry))
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }

    return presignedURL.URL, nil
}

// GeneratePresignedDownloadURL generates a presigned URL for downloading
func (s *S3Client) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(s.client)

    presignedURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(key),
    }, s3.WithPresignExpires(expiry))
    if err != nil {
        return "", fmt.Errorf("failed to generate presigned URL: %w", err)
    }

    return presignedURL.URL, nil
}
```

**Contoh pakai di handler:**

```go
// Handler: Upload avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
    userID := c.GetString("user_id")

    file, header, err := c.Request.FormFile("avatar")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()

    // Generate unique key
    key := fmt.Sprintf("avatars/%s/%d-%s", userID, time.Now().Unix(), header.Filename)

    // Upload ke S3
    err = h.s3Client.UploadFile(c.Request.Context(), key, file, header.Header.Get("Content-Type"))
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to upload avatar"})
        return
    }

    // Return URL
    url := h.s3Client.GetFileURL(key)
    c.JSON(200, gin.H{"avatar_url": url})
}

// Handler: Get presigned upload URL (client-side upload)
func (h *UserHandler) GetUploadURL(c *gin.Context) {
    userID := c.GetString("user_id")
    filename := c.Query("filename")

    key := fmt.Sprintf("avatars/%s/%d-%s", userID, time.Now().Unix(), filename)

    // Generate presigned URL (valid 15 menit)
    url, err := h.s3Client.GeneratePresignedUploadURL(
        c.Request.Context(), key, 15*time.Minute,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate upload URL"})
        return
    }

    c.JSON(200, gin.H{
        "upload_url": url,
        "file_key":   key,
        "expires_in": "15 minutes",
    })
}
```

---

## 13.11 ECR — Elastic Container Registry

### Apa Itu ECR?

ECR itu **Docker image registry** milik AWS. Sama kayak Docker Hub, tapi private dan terintegrasi dengan AWS services (ECS, EKS, Lambda).

```
Docker Hub   = npm registry (public)
GitHub GHCR  = npm registry (terikat GitHub)
AWS ECR      = npm registry (terikat AWS, private by default)
```

### Kenapa ECR?

1. **Private by default** — image kamu ga bisa diakses orang lain
2. **Integrated** — tinggal pull dari ECR ke EC2/ECS/EKS
3. **IAM-based auth** — ga perlu manage credentials terpisah
4. **Free tier** — 500MB storage gratis (always free)
5. **Image scanning** — scan vulnerabilities otomatis

### Create ECR Repository

```bash
# Create repository
aws ecr create-repository \
    --repository-name task-manager \
    --image-scanning-configuration scanOnPush=true \
    --region ap-southeast-1

# Output:
# {
#     "repository": {
#         "repositoryUri": "123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager",
#         ...
#     }
# }

# Simpan repository URI — kamu butuh ini untuk push image
```

### Push Image ke ECR

```bash
# Step 1: Login ke ECR
aws ecr get-login-password --region ap-southeast-1 | \
    docker login --username AWS --password-stdin \
    123456789012.dkr.ecr.ap-southeast-1.amazonaws.com

# Output: "Login Succeeded"

# Step 2: Build image
docker build -t task-manager:latest .

# Step 3: Tag image untuk ECR
docker tag task-manager:latest \
    123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest

# Step 4: Push!
docker push 123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest

# Step 5: Verify
aws ecr describe-images \
    --repository-name task-manager \
    --query 'imageDetails[*].[imageTags,imageSizeInBytes]'
```

### Integrate ECR dengan GitHub Actions

Di bab CI/CD kita udah bikin workflow push ke GHCR. Sekarang kita ubah supaya push ke ECR:

```yaml
# ============================================
# File: .github/workflows/build-push-ecr.yml
# Build & Push Docker Image ke AWS ECR
# ============================================

name: Build & Push to ECR

on:
  push:
    branches: [main]

env:
  AWS_REGION: ap-southeast-1
  ECR_REPOSITORY: task-manager

jobs:
  build-and-push:
    name: Build & Push to ECR
    runs-on: ubuntu-latest

    # Permission untuk OIDC (recommended vs access keys)
    permissions:
      id-token: write
      contents: read

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # Login ke AWS pakai OIDC (lebih secure dari access keys)
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ${{ env.AWS_REGION }}
          # Atau pakai access keys (kurang recommended):
          # aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          # aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      # Login ke ECR
      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2

      # Build, tag, push
      - name: Build, tag, and push image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          # Build
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG .
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:latest .

          # Push both tags
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:latest

          echo "Image pushed: $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG"
```

**Setup Secrets di GitHub:**

```
GitHub repo → Settings → Secrets and variables → Actions → New repository secret

Tambahkan:
- AWS_ACCESS_KEY_ID: (dari IAM user)
- AWS_SECRET_ACCESS_KEY: (dari IAM user)
- AWS_ROLE_ARN: (kalau pakai OIDC — lebih recommended)
```

### ECR Lifecycle Policy (Bersihkan Image Lama)

Image lama numpuk = storage cost naik. Set lifecycle policy untuk auto-delete:

```bash
aws ecr put-lifecycle-policy \
    --repository-name task-manager \
    --lifecycle-policy-text '{
        "rules": [
            {
                "rulePriority": 1,
                "description": "Keep only last 10 images",
                "selection": {
                    "tagStatus": "any",
                    "countType": "imageCountMoreThan",
                    "countNumber": 10
                },
                "action": {
                    "type": "expire"
                }
            }
        ]
    }'
```

---

## 13.12 Deploy Task Manager ke AWS — Full Step-by-Step

Ini dia bagian yang paling penting. Kita deploy **seluruh stack** Task Manager API ke AWS. Step by step, dari awal sampai API bisa diakses dari internet.

### Overview Langkah-Langkah

```
┌─────────────────────────────────────────────────┐
│           Full Deployment Checklist               │
│                                                   │
│  1. [✓] Bikin VPC + subnets                      │
│  2. [✓] Bikin Security Groups                     │
│  3. [✓] Launch RDS PostgreSQL                     │
│  4. [✓] Launch ElastiCache Redis                  │
│  5. [ ] Launch EC2 instance                       │
│  6. [ ] SSH ke EC2, install Docker                │
│  7. [ ] Push image ke ECR (atau Docker Hub)       │
│  8. [ ] Pull image di EC2                         │
│  9. [ ] Run container dengan env vars             │
│  10.[ ] Test API dari public IP                   │
│  11.[ ] Setup domain (optional)                   │
│                                                   │
└─────────────────────────────────────────────────┘
```

Kita udah cover step 1-4 di sections sebelumnya. Sekarang kita lanjut dari step 5.

### Step 5: Launch EC2 Instance

Kalau kamu belum launch EC2, ikutin step di section 13.6.

```bash
# Verify EC2 running
aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=task-manager-api" \
    --query 'Reservations[0].Instances[0].[InstanceId,PublicIpAddress,State.Name]' \
    --output table
```

### Step 6: SSH ke EC2 + Install Docker

```bash
# SSH ke EC2
ssh -i task-manager-key.pem ec2-user@<PUBLIC_IP>

# Install Docker (kalau belum)
sudo dnf update -y
sudo dnf install docker -y
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker ec2-user

# Logout & login lagi
exit
ssh -i task-manager-key.pem ec2-user@<PUBLIC_IP>

# Verify
docker --version
```

### Step 7: Pull Image

**Option A: Dari ECR**

```bash
# Di EC2 instance, login ke ECR
# Note: EC2 butuh AWS CLI dan credentials. Amazon Linux udah include AWS CLI.

# Configure AWS CLI di EC2
aws configure
# Masukkan Access Key ID, Secret Access Key, region

# Login ke ECR
aws ecr get-login-password --region ap-southeast-1 | \
    docker login --username AWS --password-stdin \
    123456789012.dkr.ecr.ap-southeast-1.amazonaws.com

# Pull image
docker pull 123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest
```

**Option B: Dari Docker Hub (lebih simple)**

```bash
# Kalau kamu push ke Docker Hub instead
docker pull yourusername/task-manager:latest
```

**Option C: Dari GHCR (GitHub Container Registry)**

```bash
# Login ke GHCR
echo $GHCR_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull
docker pull ghcr.io/irzan-aldi/task-manager:latest
```

### Step 8-9: Run Container dengan Environment Variables

Ini bagian kritis — kamu harus set environment variables yang pointing ke RDS dan ElastiCache.

```bash
# Cari endpoints:
# RDS endpoint: AWS Console → RDS → task-manager-db → Endpoint
# ElastiCache endpoint: AWS Console → ElastiCache → task-manager-redis → Primary Endpoint

# Run container!
docker run -d \
    --name task-manager-api \
    -p 80:8080 \
    -e APP_ENV=production \
    -e PORT=8080 \
    -e DB_HOST=task-manager-db.abcdef123456.ap-southeast-1.rds.amazonaws.com \
    -e DB_PORT=5432 \
    -e DB_USER=postgres \
    -e DB_PASSWORD="YourStrongPassword123!" \
    -e DB_NAME=task_manager \
    -e REDIS_HOST=task-manager-redis.abcdef.0001.apse1.cache.amazonaws.com \
    -e REDIS_PORT=6379 \
    -e REDIS_PASSWORD="" \
    -e JWT_SECRET="your-super-secret-jwt-key-change-this" \
    --restart unless-stopped \
    123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest

# Check container running
docker ps

# Check logs
docker logs -f task-manager-api

# Output yang expected:
# [GIN-debug] Listening and serving HTTP on :8080
# Server starting on port 8080 (env: production)
# Database migrated successfully
```

> **Catatan: `-p 80:8080`** — ini map port 80 (HTTP standard) di EC2 ke port 8080 di container. Jadi user akses `http://PUBLIC_IP` (port 80) dan di-forward ke app di port 8080.

### Alternative: Pakai docker-compose di EC2

Untuk approach yang lebih organized, bikin `docker-compose.yml` di EC2:

```bash
# Bikin directory
mkdir -p /home/ec2-user/task-manager
cd /home/ec2-user/task-manager

# Bikin .env file
cat > .env << 'EOF'
APP_ENV=production
PORT=8080
DB_HOST=task-manager-db.abcdef123456.ap-southeast-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=YourStrongPassword123!
DB_NAME=task_manager
REDIS_HOST=task-manager-redis.abcdef.0001.apse1.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=your-super-secret-jwt-key-change-this
EOF

# Bikin docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  api:
    image: 123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest
    container_name: task-manager-api
    ports:
      - "80:8080"
    env_file:
      - .env
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
EOF

# Run
docker compose up -d

# Check
docker compose ps
docker compose logs -f
```

### Step 10: Test API dari Public IP

```bash
# Dari laptop kamu (bukan dari EC2)

# Health check
curl http://54.169.123.456/health
# {"status": "ok"}

# Register user
curl -X POST http://54.169.123.456/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "password123",
        "name": "Test User"
    }'

# Login
curl -X POST http://54.169.123.456/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "password123"
    }'
# Response: {"token": "eyJhbGciOi..."}

# Create board (pakai token dari login)
curl -X POST http://54.169.123.456/api/v1/boards \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer eyJhbGciOi..." \
    -d '{"name": "My First Cloud Board"}'

# SELAMAT! API kamu udah live di cloud!
```

### Troubleshooting Deployment

| Problem | Kemungkinan Penyebab | Solusi |
|---------|---------------------|--------|
| `curl: connection refused` | Container ga jalan / port salah | `docker ps`, check port mapping |
| `curl: connection timed out` | Security Group ga allow port | Check inbound rules EC2 SG |
| Container exit immediately | App crash (DB connection fail) | `docker logs task-manager-api` |
| `dial tcp: connection refused` (DB) | RDS SG ga allow dari EC2 | Check RDS SG source = EC2 SG |
| `FATAL: password authentication failed` | Password salah | Double check RDS password |
| `no such host` (Redis) | ElastiCache endpoint salah | Copy endpoint dari AWS Console |

---

## 13.13 Route 53 + Domain

### Kenapa Butuh Domain?

Sekarang API kamu bisa diakses di `http://54.169.123.456`. Tapi:
- IP address susah diingat
- IP bisa berubah kalau EC2 di-stop/start
- Ga professional (siapa yang mau kasih IP address ke client?)

**Solusinya:** Domain name. `api.taskmanager.com` lebih keren dari `54.169.123.456`.

### Option 1: Route 53 (AWS DNS)

Route 53 itu DNS service dari AWS. Fully integrated dengan AWS services.

```
Pricing:
- Hosted zone: $0.50/bulan per domain
- DNS queries: $0.40 per juta queries (murah banget)
- Domain registration: $12-15/tahun untuk .com
```

**Register domain via Route 53:**

```
1. AWS Console → Route 53 → "Register domain"
2. Cari domain yang mau kamu beli (misalnya: taskmanager-api.com)
3. Add to cart → checkout
4. Isi contact info
5. Done! Domain active dalam hitungan menit - jam
```

**Create A Record (pointing ke EC2):**

```
1. Route 53 → Hosted zones → pilih domain kamu
2. Klik "Create record"
3. Record name: "api" (ini jadi api.yourdomain.com)
4. Record type: A
5. Value: 54.169.123.456 (Public IP EC2 kamu)
6. TTL: 300 (5 menit)
7. Klik "Create records"
```

```bash
# Via CLI
aws route53 change-resource-record-sets \
    --hosted-zone-id Z0123456789 \
    --change-batch '{
        "Changes": [
            {
                "Action": "UPSERT",
                "ResourceRecordSet": {
                    "Name": "api.yourdomain.com",
                    "Type": "A",
                    "TTL": 300,
                    "ResourceRecords": [
                        {
                            "Value": "54.169.123.456"
                        }
                    ]
                }
            }
        ]
    }'
```

### Option 2: Elastic IP (Biar IP Ga Berubah)

Masalah: Kalau EC2 di-stop lalu start lagi, public IP-nya **berubah**. Ini bikin DNS record kamu jadi salah.

Solusinya: **Elastic IP** — static IP yang kamu "own" dan bisa attach ke EC2 mana aja.

```bash
# Allocate Elastic IP
aws ec2 allocate-address --domain vpc
# Output: AllocationId = eipalloc-0abc123

# Associate ke EC2
aws ec2 associate-address \
    --instance-id i-0abc123 \
    --allocation-id eipalloc-0abc123

# Sekarang IP-nya fixed, ga berubah meskipun EC2 di-stop/start
```

> **PERINGATAN:** Elastic IP itu **GRATIS** selama di-attach ke running instance. Tapi kalau Elastic IP **ga di-attach** (idle), AWS charge ~$3.60/bulan. Jangan allocate Elastic IP kalau ga dipake!

### Option 3: Cloudflare (Lebih Murah dan Lebih Powerful)

Kalau kamu udah punya domain dari provider lain (Namecheap, GoDaddy, Niagahoster), bisa pakai **Cloudflare** sebagai DNS — **gratis**.

```
Cloudflare advantages:
- DNS management: FREE
- DDoS protection: FREE
- SSL/HTTPS: FREE (auto)
- CDN: FREE
- Analytics: FREE

Route 53 advantages:
- Integrated dengan AWS services
- SLA 100% uptime
- Latency-based routing
```

**Setup Cloudflare:**

```
1. Sign up di https://cloudflare.com
2. Add site → masukkan domain kamu
3. Cloudflare kasih nameserver (ns1.cloudflare.com, ns2.cloudflare.com)
4. Update nameserver di domain registrar kamu
5. Di Cloudflare, tambah A record:
   - Name: api
   - Content: 54.169.123.456 (EC2 public IP)
   - Proxy status: Proxied (orange cloud) → auto HTTPS!
6. Done!
```

> **Rekomendasi:** Untuk belajar, **Cloudflare** lebih recommended karena gratis dan lebih gampang setup. Route 53 dipake kalau kamu butuh deep integration dengan AWS (misalnya latency-based routing, health check DNS).

---

## 13.14 SSL/HTTPS

### Kenapa HTTPS?

```
HTTP:  Data dikirim plain text → siapa aja bisa baca (man-in-the-middle)
HTTPS: Data di-encrypt → cuma sender dan receiver yang bisa baca

HTTP:  http://api.yourdomain.com  → Browser kasih warning "Not Secure"
HTTPS: https://api.yourdomain.com → Browser kasih gembok hijau
```

API yang handle authentication (JWT, password) **WAJIB** pakai HTTPS di production. Kalau ga, password user bisa di-sniff.

### Option 1: Cloudflare (Paling Gampang)

Kalau kamu pakai Cloudflare DNS (section sebelumnya), **HTTPS otomatis aktif**. Ga perlu setup apa-apa. Cloudflare yang handle SSL certificate.

```
Cloudflare SSL Modes:
- Flexible: HTTPS antara user ↔ Cloudflare, HTTP antara Cloudflare ↔ EC2
- Full: HTTPS semua jalur (butuh cert di server)
- Full (Strict): HTTPS semua jalur + valid cert di server

Untuk mulai: "Flexible" udah cukup
Production serius: "Full (Strict)"
```

### Option 2: Let's Encrypt + Certbot (Free SSL)

Kalau kamu ga pakai Cloudflare, bisa pakai **Let's Encrypt** — free SSL certificate.

```bash
# SSH ke EC2
ssh -i task-manager-key.pem ec2-user@54.169.123.456

# Install Nginx (sebagai reverse proxy)
sudo dnf install nginx -y
sudo systemctl start nginx
sudo systemctl enable nginx

# Install Certbot
sudo dnf install certbot python3-certbot-nginx -y

# Bikin Nginx config untuk domain kamu
sudo tee /etc/nginx/conf.d/task-manager.conf << 'EOF'
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Test Nginx config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Get SSL certificate
sudo certbot --nginx -d api.yourdomain.com

# Follow prompts:
# - Email: your@email.com
# - Agree to ToS: Y
# - Redirect HTTP to HTTPS: 2 (yes, redirect)

# Certificate auto-renewal test
sudo certbot renew --dry-run

# Certificate auto-renews via cron/systemd timer
# Check timer:
sudo systemctl list-timers | grep certbot
```

Setelah Certbot selesai, Nginx config otomatis di-update dengan SSL:

```nginx
server {
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    listen 443 ssl;
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
}

server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

### Option 3: AWS Certificate Manager (ACM) + Load Balancer

ACM itu free SSL certificate dari AWS. Tapi **cuma bisa dipake di AWS services** (Load Balancer, CloudFront, API Gateway) — ga bisa di-install di EC2 langsung.

```
Kalau pakai ACM, architecture-nya jadi:
User → HTTPS → Load Balancer (ACM cert) → HTTP → EC2 (internal)

Keuntungan:
- Certificate gratis dan auto-renew
- SSL termination di Load Balancer (EC2 ga perlu handle SSL)

Kerugian:
- Butuh Load Balancer ($16-18/bulan) ← ga free tier
- Overkill untuk single server
```

> **Rekomendasi untuk belajar:** Pakai **Cloudflare** (gratis, paling gampang). Untuk production serius, pakai **ACM + Load Balancer**.

---

## 13.15 CloudWatch — Monitoring

### Apa Itu CloudWatch?

CloudWatch itu **monitoring service** dari AWS. Dia collect metrics, logs, dan bisa bikin alarms.

```
CloudWatch = Prometheus + Grafana versi AWS

Metrics  = Data angka (CPU usage, memory, disk, network)
Logs     = Text output dari app (stdout, stderr)
Alarms   = Alert kalau metric melebihi threshold
Dashboard = Visualisasi metrics
```

### Basic Monitoring (Free)

EC2 instance otomatis kirim metrics ke CloudWatch setiap **5 menit** (free). Metrics yang tersedia:

```
EC2 Metrics (Free):
- CPUUtilization        → Berapa % CPU yang dipake
- NetworkIn             → Bytes masuk
- NetworkOut            → Bytes keluar
- DiskReadOps           → Disk read operations
- DiskWriteOps          → Disk write operations
- StatusCheckFailed     → Instance health

RDS Metrics (Free):
- CPUUtilization
- FreeableMemory
- DatabaseConnections
- ReadIOPS / WriteIOPS
- FreeStorageSpace
```

### Setup CloudWatch Alarm — CPU > 80%

Kita mau dapet notifikasi kalau CPU EC2 di atas 80% selama 5 menit. Ini bisa berarti app kamu overloaded.

**Via Console:**

```
1. AWS Console → CloudWatch → Alarms → "Create alarm"
2. Select metric:
   - EC2 → Per-Instance Metrics
   - Pilih instance task-manager-api
   - Metric: CPUUtilization
3. Conditions:
   - Greater/Equal than: 80
   - Period: 5 minutes
   - Datapoints: 1 out of 1
4. Actions:
   - Create new SNS topic: "task-manager-alerts"
   - Email: your@email.com
5. Name: "task-manager-cpu-high"
6. Create alarm
7. CONFIRM email subscription (cek inbox!)
```

**Via AWS CLI:**

```bash
# Create SNS topic untuk notifications
aws sns create-topic --name task-manager-alerts
# Output: TopicArn = arn:aws:sns:ap-southeast-1:123456789012:task-manager-alerts

# Subscribe email
aws sns subscribe \
    --topic-arn arn:aws:sns:ap-southeast-1:123456789012:task-manager-alerts \
    --protocol email \
    --notification-endpoint your@email.com

# CEHCK EMAIL DAN CONFIRM SUBSCRIPTION!

# Create CPU alarm
aws cloudwatch put-metric-alarm \
    --alarm-name task-manager-cpu-high \
    --alarm-description "CPU utilization > 80% for 5 minutes" \
    --metric-name CPUUtilization \
    --namespace AWS/EC2 \
    --statistic Average \
    --period 300 \
    --threshold 80 \
    --comparison-operator GreaterThanOrEqualToThreshold \
    --evaluation-periods 1 \
    --dimensions "Name=InstanceId,Value=i-0abc123" \
    --alarm-actions arn:aws:sns:ap-southeast-1:123456789012:task-manager-alerts
```

### CloudWatch Logs — Application Logs

Kamu bisa kirim logs dari Docker container ke CloudWatch:

```bash
# Install CloudWatch agent di EC2
sudo dnf install amazon-cloudwatch-agent -y

# Atau, lebih simple: Run Docker with awslogs driver
docker run -d \
    --name task-manager-api \
    -p 80:8080 \
    --log-driver=awslogs \
    --log-opt awslogs-region=ap-southeast-1 \
    --log-opt awslogs-group=/ecs/task-manager \
    --log-opt awslogs-stream-prefix=api \
    -e APP_ENV=production \
    ... (env vars) \
    123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/task-manager:latest
```

> **Catatan:** Untuk pakai awslogs driver, EC2 instance butuh IAM role/policy yang allow write ke CloudWatch Logs.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
```

### Setup Billing Alarm (PENTING!)

Ini **wajib** — alarm kalau billing kamu melebihi threshold.

```bash
# PENTING: Billing alarms harus di-setup di region us-east-1
aws cloudwatch put-metric-alarm \
    --region us-east-1 \
    --alarm-name billing-alarm-10usd \
    --alarm-description "Billing exceeds $10" \
    --metric-name EstimatedCharges \
    --namespace AWS/Billing \
    --statistic Maximum \
    --period 21600 \
    --threshold 10 \
    --comparison-operator GreaterThanOrEqualToThreshold \
    --evaluation-periods 1 \
    --dimensions "Name=Currency,Value=USD" \
    --alarm-actions arn:aws:sns:us-east-1:123456789012:billing-alerts
```

> **Catatan:** Sebelum bikin billing alarm, kamu harus enable "Receive Billing Alerts" di AWS Console → Billing → Billing Preferences.

---

## 13.16 AWS CLI

### Install AWS CLI

**macOS:**

```bash
# Via Homebrew (recommended)
brew install awscli

# Verify
aws --version
# aws-cli/2.x.x Python/3.x.x Darwin/24.x.x
```

**Linux (EC2 Amazon Linux udah include):**

```bash
# Amazon Linux 2023 — AWS CLI udah pre-installed
aws --version

# Ubuntu/Debian
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Verify
aws --version
```

### Configure

```bash
aws configure

# Prompt:
# AWS Access Key ID [None]: AKIA...YOUR_KEY
# AWS Secret Access Key [None]: wJalr...YOUR_SECRET
# Default region name [None]: ap-southeast-1
# Default output format [None]: json

# Verify
aws sts get-caller-identity
```

**Multiple profiles:**

```bash
# Configure profile untuk project tertentu
aws configure --profile task-manager

# Pakai profile
aws ec2 describe-instances --profile task-manager

# Atau set environment variable
export AWS_PROFILE=task-manager
aws ec2 describe-instances  # otomatis pakai profile task-manager
```

### Basic Commands Cheat Sheet

**EC2:**

```bash
# List all instances
aws ec2 describe-instances \
    --query 'Reservations[*].Instances[*].[InstanceId,InstanceType,State.Name,PublicIpAddress,Tags[?Key==`Name`].Value|[0]]' \
    --output table

# Start instance
aws ec2 start-instances --instance-ids i-0abc123

# Stop instance
aws ec2 stop-instances --instance-ids i-0abc123

# Reboot instance
aws ec2 reboot-instances --instance-ids i-0abc123

# Terminate (DELETE) instance
aws ec2 terminate-instances --instance-ids i-0abc123

# Describe security groups
aws ec2 describe-security-groups --group-ids sg-0abc123
```

**S3:**

```bash
# List buckets
aws s3 ls

# List files in bucket
aws s3 ls s3://bucket-name/

# Copy file
aws s3 cp file.txt s3://bucket-name/

# Sync directory
aws s3 sync ./local-dir s3://bucket-name/remote-dir

# Remove file
aws s3 rm s3://bucket-name/file.txt

# Remove all files (DANGER!)
aws s3 rm s3://bucket-name/ --recursive
```

**RDS:**

```bash
# List instances
aws rds describe-db-instances \
    --query 'DBInstances[*].[DBInstanceIdentifier,DBInstanceClass,Engine,DBInstanceStatus]' \
    --output table

# Get endpoint
aws rds describe-db-instances \
    --db-instance-identifier task-manager-db \
    --query 'DBInstances[0].Endpoint.Address' \
    --output text

# Create snapshot (manual backup)
aws rds create-db-snapshot \
    --db-instance-identifier task-manager-db \
    --db-snapshot-identifier task-manager-backup-$(date +%Y%m%d)

# Stop instance (save cost — tapi auto-restart after 7 days!)
aws rds stop-db-instance --db-instance-identifier task-manager-db

# Start instance
aws rds start-db-instance --db-instance-identifier task-manager-db
```

**ElastiCache:**

```bash
# List clusters
aws elasticache describe-cache-clusters \
    --query 'CacheClusters[*].[CacheClusterId,CacheNodeType,Engine,CacheClusterStatus]' \
    --output table

# Get endpoint
aws elasticache describe-cache-clusters \
    --cache-cluster-id task-manager-redis \
    --show-cache-node-info \
    --query 'CacheClusters[0].CacheNodes[0].Endpoint' \
    --output json
```

**CloudWatch:**

```bash
# Get CPU metrics for EC2
aws cloudwatch get-metric-statistics \
    --namespace AWS/EC2 \
    --metric-name CPUUtilization \
    --dimensions Name=InstanceId,Value=i-0abc123 \
    --start-time $(date -u -v-1H +%Y-%m-%dT%H:%M:%S) \
    --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
    --period 300 \
    --statistics Average

# List alarms
aws cloudwatch describe-alarms \
    --query 'MetricAlarms[*].[AlarmName,StateValue,MetricName]' \
    --output table
```

### Automate dengan Script

Bikin script untuk common tasks:

```bash
#!/bin/bash
# File: deploy.sh
# Deploy Task Manager API ke EC2

set -e

# Config
EC2_HOST="54.169.123.456"
KEY_FILE="task-manager-key.pem"
ECR_REGISTRY="123456789012.dkr.ecr.ap-southeast-1.amazonaws.com"
IMAGE_NAME="task-manager"
IMAGE_TAG="${1:-latest}"  # Default: latest, atau pass tag sebagai argument

echo "=== Deploying $IMAGE_NAME:$IMAGE_TAG ==="

# Step 1: Push image ke ECR (kalau belum)
echo "Step 1: Login to ECR..."
aws ecr get-login-password --region ap-southeast-1 | \
    docker login --username AWS --password-stdin $ECR_REGISTRY

echo "Step 2: Build and push image..."
docker build -t $ECR_REGISTRY/$IMAGE_NAME:$IMAGE_TAG .
docker push $ECR_REGISTRY/$IMAGE_NAME:$IMAGE_TAG

# Step 3: SSH ke EC2 dan deploy
echo "Step 3: Deploy to EC2..."
ssh -i $KEY_FILE ec2-user@$EC2_HOST << REMOTE_COMMANDS
    # Login ke ECR di EC2
    aws ecr get-login-password --region ap-southeast-1 | \
        docker login --username AWS --password-stdin $ECR_REGISTRY

    # Pull image terbaru
    docker pull $ECR_REGISTRY/$IMAGE_NAME:$IMAGE_TAG

    # Stop old container
    docker stop task-manager-api 2>/dev/null || true
    docker rm task-manager-api 2>/dev/null || true

    # Start new container
    cd /home/ec2-user/task-manager
    docker compose up -d

    # Verify
    sleep 5
    docker ps
    curl -s http://localhost/health
REMOTE_COMMANDS

echo "=== Deploy complete! ==="
echo "API: http://$EC2_HOST/health"
```

```bash
# Pakai:
chmod +x deploy.sh
./deploy.sh          # Deploy latest
./deploy.sh v1.2.3   # Deploy specific version
```

---

## 13.17 Cost Management — Jangan Sampai Kena Tagihan

### AWS Free Tier Limits

**BACA INI BAIK-BAIK.** Banyak orang belajar AWS tiba-tiba kena tagihan $50-100 karena ga matiin resources.

```
Free Tier Limits (12 bulan pertama):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EC2:         750 jam/bulan t2.micro     = ~1 instance 24/7
RDS:         750 jam/bulan db.t3.micro  = ~1 instance 24/7
ElastiCache: 750 jam/bulan cache.t3.micro = ~1 node 24/7
S3:          5 GB storage
ECR:         500 MB (always free)

JEBAKAN UMUM:
━━━━━━━━━━━━
1. Lupa matiin EC2 yang ga dipake → $8-50/bulan
2. Launch 2 EC2 t2.micro (750 jam / 2 = masing-masing 375 jam) → kena charge!
3. RDS multi-AZ enabled → 2x biaya → keluar free tier
4. NAT Gateway → $32/bulan (BUKAN free tier!)
5. Elastic IP ga di-attach → $3.60/bulan
6. Load Balancer → $16-18/bulan (BUKAN free tier!)
7. EBS volumes unused → kena charge
8. Snapshots numpuk → kena charge
```

### Setup Billing Alert (HARUS!)

Ini **langkah pertama** setelah bikin akun AWS. Ga negotiable.

```
1. AWS Console → klik username pojok kanan atas
2. "Billing and Cost Management"
3. Left sidebar → "Budgets"
4. "Create a budget"
5. Pilih "Zero spend budget"
   atau "Monthly cost budget"
6. Budget amount: $5 (atau $10)
7. Email: your@email.com
8. Create budget
```

```bash
# Via CLI — bikin budget $10/bulan
aws budgets create-budget \
    --account-id 123456789012 \
    --budget '{
        "BudgetName": "Monthly-10-USD",
        "BudgetLimit": {
            "Amount": "10",
            "Unit": "USD"
        },
        "TimeUnit": "MONTHLY",
        "BudgetType": "COST"
    }' \
    --notifications-with-subscribers '[
        {
            "Notification": {
                "NotificationType": "ACTUAL",
                "ComparisonOperator": "GREATER_THAN",
                "Threshold": 80,
                "ThresholdType": "PERCENTAGE"
            },
            "Subscribers": [
                {
                    "SubscriptionType": "EMAIL",
                    "Address": "your@email.com"
                }
            ]
        }
    ]'
```

### Estimasi Biaya Per Bulan

**Skenario 1: Full Free Tier (Tahun Pertama)**

| Service | Cost |
|---------|------|
| EC2 t2.micro (1x, 24/7) | **$0** (free tier) |
| RDS db.t3.micro (1x, 24/7) | **$0** (free tier) |
| ElastiCache cache.t3.micro (1x, 24/7) | **$0** (free tier) |
| S3 (< 5GB) | **$0** (free tier) |
| ECR (< 500MB) | **$0** (always free) |
| **Total** | **$0/bulan** |

**Skenario 2: Setelah Free Tier Habis (13+ bulan)**

| Service | Cost |
|---------|------|
| EC2 t2.micro (1x, 24/7) | ~$8.50/bulan |
| RDS db.t3.micro (1x, 24/7) | ~$12-15/bulan |
| ElastiCache cache.t3.micro (1x, 24/7) | ~$12/bulan |
| S3 (1GB) | ~$0.02/bulan |
| Data transfer (10GB) | ~$0.90/bulan |
| **Total** | **~$33-37/bulan** |

**Skenario 3: Production Kecil (setelah free tier)**

| Service | Cost |
|---------|------|
| EC2 t3.small (1x, 24/7) | ~$15/bulan |
| RDS db.t3.small (1x, multi-AZ) | ~$50/bulan |
| ElastiCache cache.t3.micro (1x) | ~$12/bulan |
| ALB (Application Load Balancer) | ~$18/bulan |
| NAT Gateway | ~$32/bulan |
| S3 (10GB) | ~$0.25/bulan |
| Route 53 | ~$0.50/bulan |
| CloudWatch | ~$3/bulan |
| **Total** | **~$130/bulan** |

### Tips Hemat

```
1. MATIIN resources yang ga dipake
   - Stop EC2 kalau ga dipake (stop = ga bayar compute, tapi disk tetap bayar)
   - Stop RDS kalau ga dipake (auto-restart after 7 hari!)
   - Delete resources yang beneran ga dipake

2. Pakai Reserved Instances / Savings Plans
   - Kalau kamu tau butuh 1 tahun, bisa hemat 30-40%
   - Hanya untuk production yang pasti long-term

3. Right-sizing
   - Jangan pakai instance yang terlalu besar
   - Monitor CPU/memory, kalau rendah → downgrade

4. Spot Instances (untuk non-critical workloads)
   - 60-90% lebih murah dari on-demand
   - Tapi bisa di-terminate kapan aja oleh AWS

5. Review Cost Explorer bulanan
   - AWS Console → Billing → Cost Explorer
   - Lihat service mana yang paling mahal
   - Identify resources yang waste

6. Tag everything
   - Setiap resource kasih tag: Project=task-manager
   - Bisa filter cost per project

7. Alternatif lebih murah
   - Untuk project kecil, DigitalOcean/Hetzner bisa lebih murah
   - Railway/Fly.io buat yang ga mau manage infra
```

### Cleanup Script — Matiin Semua Resources

Kalau kamu udah selesai belajar dan mau **stop semua** biar ga kena charge:

```bash
#!/bin/bash
# File: aws-cleanup.sh
# HATI-HATI: Script ini menghapus resources!

echo "=== AWS Cleanup: Task Manager ==="
echo "WARNING: This will DELETE all resources!"
read -p "Are you sure? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Cancelled."
    exit 0
fi

# 1. Terminate EC2
echo "Terminating EC2..."
INSTANCE_ID=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=task-manager-api" \
    --query 'Reservations[0].Instances[0].InstanceId' \
    --output text)
aws ec2 terminate-instances --instance-ids $INSTANCE_ID

# 2. Delete RDS (skip final snapshot to save cost)
echo "Deleting RDS..."
aws rds delete-db-instance \
    --db-instance-identifier task-manager-db \
    --skip-final-snapshot

# 3. Delete ElastiCache
echo "Deleting ElastiCache..."
aws elasticache delete-cache-cluster \
    --cache-cluster-id task-manager-redis

# 4. Release Elastic IP (kalau ada)
echo "Releasing Elastic IPs..."
ALLOC_ID=$(aws ec2 describe-addresses \
    --filters "Name=tag:Name,Values=task-manager" \
    --query 'Addresses[0].AllocationId' \
    --output text)
if [ "$ALLOC_ID" != "None" ]; then
    aws ec2 release-address --allocation-id $ALLOC_ID
fi

# 5. Delete ECR images
echo "Deleting ECR repository..."
aws ecr delete-repository \
    --repository-name task-manager \
    --force

# 6. Empty dan delete S3 bucket
echo "Deleting S3 bucket..."
aws s3 rm s3://task-manager-assets-aldi --recursive
aws s3 rb s3://task-manager-assets-aldi

echo "=== Cleanup complete! ==="
echo "Note: VPC, Security Groups, dan Subnets masih ada (free)."
echo "Note: RDS dan ElastiCache butuh beberapa menit untuk fully deleted."
echo "VERIFY: Check AWS Console besok untuk pastikan ga ada charge unexpected."
```

---

## 13.18 AWS vs Alternatives

AWS bukan satu-satunya pilihan. Tergantung kebutuhan, ada alternatives yang mungkin lebih cocok:

### Comparison Table

| Provider | Tipe | Best For | Min Cost/bulan | Learning Curve |
|----------|------|----------|---------------|---------------|
| **AWS** | IaaS++ | Enterprise, career growth, full control | $0 (free tier) → $30+ | Tinggi |
| **GCP** | IaaS++ | Data/ML, Kubernetes (GKE), startup credits | $0 (free tier) → $25+ | Tinggi |
| **Azure** | IaaS++ | Microsoft stack, enterprise, .NET | $0 (free tier) → $30+ | Tinggi |
| **DigitalOcean** | IaaS | Simple VPS, straightforward pricing | $4-6 | Rendah |
| **Hetzner** | IaaS | Murah banget (EU), dedicated servers | $3-5 (EU region) | Rendah |
| **Railway** | PaaS | Deploy cepat, zero config | $0 (hobby) → $5+ | Sangat Rendah |
| **Fly.io** | PaaS | Edge deployment, Docker-native | $0 (free) → $5+ | Rendah |
| **Render** | PaaS | Simple alternative to Heroku | $0 (free) → $7+ | Sangat Rendah |
| **Vercel** | PaaS | Frontend/serverless | $0 (hobby) → $20+ | Sangat Rendah |

### When to Use What?

```
Lagi belajar, mau boost CV, mau jadi DevOps?
→ AWS (atau GCP/Azure)

Project pribadi kecil, budget minim?
→ DigitalOcean, Hetzner, Railway

Startup, MVP cepat, ga mau manage infra?
→ Railway, Fly.io, Render

Enterprise, compliance ketat, Microsoft stack?
→ Azure

Data engineering, ML, Kubernetes?
→ GCP

Freelance project, client budget kecil?
→ DigitalOcean + Cloudflare
```

### Migrasi Antar Cloud

```
Good news: Karena kamu pakai Docker, migrasi antar cloud itu GAMPANG.

Docker image kamu bisa jalan di:
- AWS EC2 ✓
- GCP Compute Engine ✓
- Azure VM ✓
- DigitalOcean Droplet ✓
- Any VPS with Docker ✓

Yang berubah cuma:
- Environment variables (DB host, Redis host)
- Network configuration
- Provider-specific services (RDS → Cloud SQL, dll)

Yang TIDAK berubah:
- Application code
- Docker image
- CI/CD pipeline (mostly)
```

---

## 13.19 Best Practices

### Security

```
1. NEVER use root account for daily operations
   → Bikin IAM user, lock root account

2. Enable MFA everywhere
   → Root account, IAM users, console login

3. Principle of Least Privilege
   → Kasih permission yang dibutuhin aja, ga lebih
   → "Boleh baca S3" bukan "Admin S3"

4. Database di private subnet
   → RDS, ElastiCache JANGAN punya public access
   → Cuma accessible dari EC2 di VPC yang sama

5. Rotate credentials
   → Access keys, database passwords, JWT secrets
   → Minimal setiap 90 hari

6. Jangan hardcode credentials
   → Pakai environment variables
   → Atau AWS Secrets Manager / Parameter Store

7. Security Groups: least privilege
   → SSH cuma dari IP kamu (bukan 0.0.0.0/0)
   → Database cuma dari EC2 SG
```

### Infrastructure

```
1. Tag semua resources
   → Key: Project, Value: task-manager
   → Key: Environment, Value: production
   → Key: Owner, Value: aldi
   → Helps: cost tracking, automation, cleanup

2. Pakai IaC (Infrastructure as Code)
   → Terraform (bab 17) atau CloudFormation
   → Jangan bikin resources manual di console
   → Reproducible, version controlled, reviewable

3. Multi-AZ untuk production
   → RDS multi-AZ: auto-failover kalau 1 AZ down
   → EC2 di multiple AZ behind load balancer
   → (Tapi mahal — skip untuk belajar)

4. Backup strategy
   → RDS: automated backups enabled (7+ hari retention)
   → S3: versioning enabled
   → EC2: AMI snapshot periodik

5. Monitoring & alerting
   → CloudWatch alarms untuk CPU, memory, disk
   → Billing alarm (WAJIB)
   → Application-level monitoring (health check)
```

### Cost

```
1. Set billing alarm SEBELUM apapun
2. Review Cost Explorer mingguan
3. Stop/terminate unused resources
4. Right-size instances (jangan overkill)
5. Use Reserved Instances untuk long-term workloads
6. Clean up old snapshots, AMIs, unused EBS volumes
7. Consider Spot Instances untuk non-critical workloads
```

### Documentation

```
1. Catat semua resources yang kamu bikin
   → Instance IDs, endpoints, security group IDs
   → Simpan di README atau wiki

2. Catat alasan kenapa resource itu ada
   → "EC2 t2.micro: API server, free tier"
   → 6 bulan dari sekarang kamu ga ingat kenapa bikinnya

3. Catat cara deploy
   → Step by step, bisa diikutin orang lain
   → Atau lebih bagus: automate pakai script/CI/CD

4. Diagram architecture
   → Gambar sederhana: service apa, connect ke mana
   → Update kalau ada perubahan
```

---

## 13.20 Cheat Sheet

### AWS CLI Commands — Quick Reference

```bash
# ==================== IDENTITY ====================
aws sts get-caller-identity              # Siapa saya?
aws configure                           # Setup credentials
aws configure list                      # Lihat config aktif

# ==================== EC2 ====================
aws ec2 describe-instances --output table    # List instances
aws ec2 start-instances --instance-ids i-xxx  # Start
aws ec2 stop-instances --instance-ids i-xxx   # Stop
aws ec2 terminate-instances --instance-ids i-xxx  # DELETE!
aws ec2 describe-security-groups             # List SGs
aws ec2 describe-vpcs                        # List VPCs

# ==================== S3 ====================
aws s3 ls                                # List buckets
aws s3 ls s3://bucket/                   # List files
aws s3 cp file s3://bucket/              # Upload
aws s3 cp s3://bucket/file ./            # Download
aws s3 sync ./dir s3://bucket/dir        # Sync
aws s3 rm s3://bucket/file               # Delete
aws s3 mb s3://new-bucket               # Create bucket
aws s3 rb s3://bucket                    # Delete bucket

# ==================== RDS ====================
aws rds describe-db-instances --output table  # List
aws rds stop-db-instance --db-instance-identifier xxx    # Stop
aws rds start-db-instance --db-instance-identifier xxx   # Start
aws rds create-db-snapshot --db-instance-identifier xxx \
    --db-snapshot-identifier backup-name                  # Backup

# ==================== ElastiCache ====================
aws elasticache describe-cache-clusters --output table   # List
aws elasticache delete-cache-cluster --cache-cluster-id xxx  # Delete

# ==================== ECR ====================
# Login
aws ecr get-login-password --region ap-southeast-1 | \
    docker login --username AWS --password-stdin ACCOUNT.dkr.ecr.REGION.amazonaws.com

# Push
docker tag img:tag ACCOUNT.dkr.ecr.REGION.amazonaws.com/repo:tag
docker push ACCOUNT.dkr.ecr.REGION.amazonaws.com/repo:tag

# List images
aws ecr describe-images --repository-name repo-name

# ==================== CloudWatch ====================
aws cloudwatch describe-alarms --output table    # List alarms
aws cloudwatch get-metric-statistics \
    --namespace AWS/EC2 --metric-name CPUUtilization \
    --dimensions Name=InstanceId,Value=i-xxx \
    --start-time 2026-01-01T00:00:00 \
    --end-time 2026-01-02T00:00:00 \
    --period 3600 --statistics Average            # Get metrics

# ==================== IAM ====================
aws iam list-users                       # List users
aws iam list-groups                      # List groups
aws iam list-roles                       # List roles
aws iam get-user --user-name xxx         # User details

# ==================== COST ====================
aws ce get-cost-and-usage \
    --time-period Start=2026-04-01,End=2026-04-30 \
    --granularity MONTHLY \
    --metrics BlendedCost                # Monthly cost
```

### Service Comparison Table

```
┌──────────────┬────────────────┬──────────────────┬───────────────┐
│ Kebutuhan    │ AWS            │ GCP              │ Azure         │
├──────────────┼────────────────┼──────────────────┼───────────────┤
│ VM           │ EC2            │ Compute Engine   │ Virtual Machines│
│ Managed DB   │ RDS            │ Cloud SQL        │ Azure SQL     │
│ Cache        │ ElastiCache    │ Memorystore      │ Azure Cache   │
│ Storage      │ S3             │ Cloud Storage    │ Blob Storage  │
│ DNS          │ Route 53       │ Cloud DNS        │ Azure DNS     │
│ Container Reg│ ECR            │ Artifact Registry│ ACR           │
│ Container    │ ECS / EKS      │ Cloud Run / GKE  │ ACI / AKS    │
│ Serverless   │ Lambda         │ Cloud Functions  │ Azure Functions│
│ CDN          │ CloudFront     │ Cloud CDN        │ Azure CDN     │
│ Monitoring   │ CloudWatch     │ Cloud Monitoring │ Azure Monitor │
│ IAM          │ IAM            │ Cloud IAM        │ Azure AD      │
│ IaC          │ CloudFormation │ Deployment Mgr   │ ARM Templates │
│ Secrets      │ Secrets Manager│ Secret Manager   │ Key Vault     │
│ Load Balancer│ ALB/NLB        │ Cloud Load Bal   │ Azure LB      │
│ VPN          │ VPC            │ VPC              │ VNet          │
└──────────────┴────────────────┴──────────────────┴───────────────┘
```

### Architecture Patterns

```
Pattern 1: Simple (Belajar / Side Project)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
User → EC2 (API + Docker) → RDS + ElastiCache
Cost: $0 (free tier) atau $30-40/bulan

Pattern 2: Standard (Startup / Small Product)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
User → CloudFlare → ALB → EC2 (auto-scaling) → RDS (multi-AZ) + ElastiCache
Cost: $100-300/bulan

Pattern 3: Production (Mature Product)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
User → CloudFront → ALB → ECS/EKS (Fargate) → RDS + ElastiCache + S3
+ CloudWatch + WAF + Secrets Manager
Cost: $300-1000+/bulan
```

---

## 13.21 Checklist Bab Ini

### Cloud Computing Basics

- [ ] Paham apa itu cloud computing (on-premise vs cloud)
- [ ] Paham perbedaan IaaS vs PaaS vs SaaS
- [ ] Paham kenapa AWS (market leader, job demand)
- [ ] Paham analogi sewa apartemen vs bangun rumah

### AWS Account & Setup

- [ ] Punya akun AWS
- [ ] Udah bikin IAM user (ga pakai root untuk daily use!)
- [ ] MFA enabled di root account
- [ ] AWS CLI terinstall dan configured
- [ ] **Billing alarm udah di-set!** (ini paling penting)

### AWS Infrastructure Concepts

- [ ] Paham Region, Availability Zone, Edge Location
- [ ] Paham free tier limits dan durasinya (12 bulan)
- [ ] Tau 10 services utama yang dibutuhin (EC2, RDS, S3, dll)
- [ ] Paham perbedaan managed service vs self-hosted

### IAM

- [ ] Paham root account vs IAM user
- [ ] Paham IAM concepts: User, Group, Policy, Role
- [ ] Udah bikin IAM user dengan access keys
- [ ] Paham principle of least privilege

### VPC & Networking

- [ ] Paham apa itu VPC, subnet, CIDR
- [ ] Paham public subnet vs private subnet
- [ ] Paham Internet Gateway dan NAT Gateway
- [ ] Paham Security Group (dan kenapa NACL biasanya ga perlu)
- [ ] Udah bikin VPC dengan public + private subnets

### EC2

- [ ] Paham instance types (family, generation, size)
- [ ] Paham AMI dan key pair
- [ ] Bisa launch EC2 instance (t2.micro, free tier)
- [ ] Bisa SSH ke EC2
- [ ] Bisa install Docker di EC2

### Security Groups

- [ ] Paham inbound vs outbound rules
- [ ] Bisa bikin security group dengan least privilege
- [ ] Database SG cuma allow dari EC2 SG (bukan 0.0.0.0/0)
- [ ] SSH cuma dari IP kamu (bukan dari anywhere)

### RDS

- [ ] Paham kenapa managed database (bukan install sendiri)
- [ ] Bisa launch RDS PostgreSQL (db.t3.micro, free tier)
- [ ] Public access = NO
- [ ] Bisa connect ke RDS dari EC2

### ElastiCache

- [ ] Bisa launch ElastiCache Redis (cache.t3.micro, free tier)
- [ ] Bisa connect ke Redis dari EC2

### S3

- [ ] Paham bucket, object, key
- [ ] Bisa upload/download file via CLI
- [ ] Paham presigned URL
- [ ] Bisa pakai AWS SDK Go untuk S3 operations

### ECR

- [ ] Bisa bikin ECR repository
- [ ] Bisa push Docker image ke ECR
- [ ] Bisa integrate ECR dengan GitHub Actions

### Deployment

- [ ] Udah deploy Task Manager API ke AWS
- [ ] API bisa diakses dari public IP
- [ ] Go API connect ke RDS PostgreSQL
- [ ] Go API connect ke ElastiCache Redis
- [ ] Health check endpoint responsive

### Domain & SSL

- [ ] Paham cara setup domain (Route 53 atau Cloudflare)
- [ ] Paham cara setup HTTPS (Cloudflare, Let's Encrypt, atau ACM)

### Monitoring

- [ ] CloudWatch basic monitoring enabled
- [ ] CPU alarm setup (> 80%)
- [ ] Billing alarm setup (WAJIB!)

### Cost Management

- [ ] Paham free tier limits
- [ ] Billing alert configured
- [ ] Tau estimasi biaya per bulan
- [ ] Tau cara cleanup resources
- [ ] Punya cleanup script

### AWS CLI

- [ ] AWS CLI installed dan configured
- [ ] Bisa jalanin basic commands (ec2, s3, rds)
- [ ] Bisa automate dengan scripts

---

> **Next:** [Bab 14: Nginx & Reverse Proxy](14-nginx.md) — kita bakal deep dive ke Nginx sebagai reverse proxy, load balancer, SSL termination, rate limiting, dan static file serving. Nginx itu "pintu depan" yang nge-handle semua traffic sebelum masuk ke app kamu. Di AWS, Nginx jalan di EC2 di depan Docker container.