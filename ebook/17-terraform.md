# Bab 17: Infrastructure as Code — Terraform dari Nol Sampai Production

## Kenapa Bab Ini Penting?

Di Bab 13, kamu deploy Task Manager API ke AWS — bikin EC2, RDS, ElastiCache, S3, semuanya. Tapi gimana cara kamu bikin semua itu? **Klik-klik di AWS Console**. Login, pilih service, klik "Create", isi form, klik "Launch". Repeat untuk setiap resource.

Sekarang bayangin ini:

```
Skenario: Startup kamu mulai grow
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Bulan 1: 1 environment (production). 5 AWS resources. Setup manual? OK lah.

Bulan 3: Boss minta staging environment. "Bikin yang sama persis kayak production."
         Kamu buka AWS Console, bikin 1-1 lagi. Lupa setting security group?
         Bug muncul di staging tapi ga di production. "Kok beda?"

Bulan 6: Boss minta development environment JUGA.
         Sekarang 3 environment x 15 resources = 45 resources dikelola manual.
         Tim baru join. "Eh, gimana cara bikin environment baru?"
         "Hmm... gue udah lupa step-nya."

Bulan 12: Disaster. Production RDS dihapus sama intern yang lagi explore Console.
          "Siapa yang ngubah security group production kemarin?"
          "Ga tau, ga ada log-nya."
          "Gimana cara rebuild dari nol? Ada dokumentasinya?"
          "......"

Pertanyaan yang muncul:
- Gimana bikin infrastruktur REPRODUCIBLE?
- Gimana track SIAPA yang ubah APA dan KAPAN?
- Gimana bikin 3 environment yang IDENTIK?
- Gimana REVIEW infrastructure changes sebelum apply?
- Gimana ROLLBACK kalau ada yang salah?
- Gimana onboard engineer baru tanpa "coba klik yang ini"?

Jawaban: TERRAFORM.
```

```
Sebelum bab ini:
- Infrastructure dibuat manual via AWS Console (klik-klik)
- Ga ada version control untuk infra
- Bikin environment baru = "coba inget step-nya"
- Ga bisa review infra changes
- Ga ada audit trail

Setelah bab ini:
- Infrastructure defined as CODE (HCL files)
- Version controlled di Git — PR, review, merge
- Bikin environment baru = terraform apply (1 command)
- terraform plan = preview changes sebelum apply
- Full audit trail via Git history
- Full Task Manager AWS infrastructure in Terraform
```

> **Analogi gampang:** Bayangin kamu seorang chef. Selama ini kamu masak **tanpa resep** — "sedikit garam, kira-kira segini minyak, aduk sampai feel-nya pas." Hasilnya? Kadang enak, kadang ga. Dan kalau ada chef baru, ga bisa replicate rasanya. **Terraform itu resep masakan kamu.** Tertulis, terukur, reproducible. "200ml minyak, 2 sendok garam, masak 15 menit suhu 180C." Siapa aja yang punya resepnya bisa bikin **hasil yang sama persis**. Dan kalau mau ubah resep, kamu review dulu sebelum masak — bukan trial & error di dapur production.

---

## 17.1 Apa Itu Infrastructure as Code (IaC)?

### The Problem — Manual Infrastructure

Cara traditional bikin infrastructure:

```
Manual Infrastructure (Console/GUI):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Login AWS Console
2. Klik EC2 → Launch Instance
3. Pilih AMI → Ubuntu 22.04
4. Pilih instance type → t3.micro
5. Configure network → pilih VPC, subnet
6. Add storage → 20GB
7. Configure security group → port 22, 80, 443, 8080
8. Review → Launch
9. Repeat untuk RDS, ElastiCache, S3, IAM, dll...
10. Semua step ada di kepala kamu. Atau di Google Docs yang udah outdated.
```

### Masalah Tanpa IaC

```
Problem 1: SNOWFLAKE SERVERS
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Setiap server itu "unik" — di-setup manual dengan konfigurasi yang slightly beda.
Production punya Go 1.25, staging punya Go 1.24.
Security group production beda sama staging.
"It works on production" tapi crash di staging. Sound familiar?

Kayak bikin nasi goreng tanpa resep:
- Hari Senin: enak (pakai 2 sendok kecap)
- Hari Selasa: keasinan (pakai 3 sendok kecap)
- Hari Rabu: hambar (lupa kecap)
Setiap "masakan" beda. Ini yang namanya snowflake.

Problem 2: NO AUDIT TRAIL
━━━━━━━━━━━━━━━━━━━━━━━━
"Siapa yang ubah security group kemarin?"
"Kenapa port 3306 dibuka ke public?"
"Kapan EC2 ini di-resize dari t3.small ke t3.micro?"
Jawaban: ga tau. Ga ada log. Ga ada history.

Problem 3: "IT WORKS ON MY AWS ACCOUNT"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Developer A setup infra di account dia → jalan.
Developer B coba bikin yang sama di account dia → error.
"Hmm, mungkin ada step yang ke-skip..."

Problem 4: NO REVIEW PROCESS
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Code changes → Pull Request → Review → Merge.
Infra changes → Login Console → Klik → Done. Ga ada review.
Padahal infra changes bisa lebih destructive dari code changes.

Problem 5: DISASTER RECOVERY
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Region crash. Semua resource hilang.
"Gimana rebuild dari nol?"
"Hmm... butuh 3 hari. Kalau inget semua step-nya."
```

### Infrastructure as Code — The Solution

**Infrastructure as Code (IaC)** = mengelola infrastructure menggunakan **file konfigurasi** (code), bukan manual klik-klik.

```
IaC Principles:
━━━━━━━━━━━━━━
1. DECLARATIVE     → Kamu define "apa yang mau" (desired state), bukan "gimana caranya"
2. VERSION CONTROL → Infrastructure code disimpan di Git
3. IDEMPOTENT      → Jalanin 1x atau 10x, hasilnya sama
4. REPRODUCIBLE    → Bikin ulang environment = 1 command
5. REVIEWABLE      → PR untuk infrastructure changes
6. TESTABLE        → Validate sebelum apply
```

```
Manual vs IaC:
━━━━━━━━━━━━━━

Manual:
  "Eh tolong bikin EC2 t3.small, di ap-southeast-1,
   security group buka port 8080, attach ke VPC yang itu..."
  → Buka Console → Klik-klik → selesai (hopefully)

IaC (Terraform):
  resource "aws_instance" "app" {
    ami           = "ami-0abcdef1234567890"
    instance_type = "t3.small"
    subnet_id     = aws_subnet.public.id

    tags = {
      Name = "task-manager-app"
    }
  }
  → terraform apply → selesai (pasti)
```

### Analogi: Resep Masakan

```
Manual Infrastructure = Masak tanpa resep
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- "Kira-kira segini garamnya"
- Hasilnya beda-beda setiap kali
- Ga bisa dikasih ke orang lain
- Kalau gagal, ga tau step mana yang salah
- Ga bisa di-review

IaC = Masak PAKAI resep
━━━━━━━━━━━━━━━━━━━━━━
- "2 sendok makan garam" (exact specification)
- Hasilnya konsisten setiap kali (reproducible)
- Bisa dikasih ke chef lain (shareable)
- Kalau gagal, bisa trace step mana (debuggable)
- Bisa di-review dan di-improve (reviewable)
- Bisa di-version: "Resep v2.1 — kurangi garam jadi 1.5 sendok" (version controlled)
```

---

## 17.2 Kenapa Terraform?

Ada beberapa IaC tools di market. Kenapa kita pilih Terraform?

### IaC Tools Comparison

| Aspek | Terraform | CloudFormation | Pulumi | Ansible |
|-------|-----------|---------------|--------|---------|
| **Vendor** | HashiCorp | AWS | Pulumi Inc | Red Hat |
| **Cloud Support** | Multi-cloud (AWS, GCP, Azure, DO, dll) | **AWS only** | Multi-cloud | Multi-cloud |
| **Language** | HCL (HashiCorp Configuration Language) | JSON/YAML | Python, Go, TypeScript, dll | YAML |
| **Paradigm** | Declarative | Declarative | Imperative + Declarative | Imperative (procedural) |
| **State** | State file (local/remote) | Managed by AWS | State file | Stateless |
| **Learning Curve** | Medium | Medium | Low (kalau udah bisa programming) | Low |
| **Community** | Huge (open-source) | AWS docs only | Growing | Huge |
| **Job Market** | Paling banyak demand | AWS-specific roles | Niche | Config management roles |
| **Use Case** | Provisioning infra | Provisioning AWS infra | Provisioning infra | Config management, provisioning |
| **Pricing** | Free (open-source) + paid (Terraform Cloud) | Free (AWS) | Free tier + paid | Free (open-source) |

### Kenapa Terraform Menang?

```
1. CLOUD-AGNOSTIC
━━━━━━━━━━━━━━━━
CloudFormation = AWS only. Kalau pindah ke GCP, rewrite semua.
Terraform     = 1 tool untuk SEMUA cloud.

Pakai AWS sekarang? Terraform.
Pindah ke GCP nanti? Terraform (ganti provider aja).
Pakai multi-cloud? Terraform.

2. HCL — SIMPLE & READABLE
━━━━━━━━━━━━━━━━━━━━━━━━━━
CloudFormation JSON:
{
  "Resources": {
    "MyInstance": {
      "Type": "AWS::EC2::Instance",
      "Properties": {
        "InstanceType": "t3.micro",
        "ImageId": "ami-0abcdef1234567890"
      }
    }
  }
}
→ Verbose, susah dibaca, ga ada comment

Terraform HCL:
resource "aws_instance" "app" {
  instance_type = "t3.micro"
  ami           = "ami-0abcdef1234567890"
}
→ Clean, readable, ada comment, ada variable

3. PLAN BEFORE APPLY
━━━━━━━━━━━━━━━━━━━
Terraform punya `terraform plan` yang PREVIEW changes sebelum apply.
Kayak "dry run" — kamu bisa lihat apa yang bakal berubah.

"Ah, ini bakal DELETE database?! Cancel dulu."
Ini fitur penyelamat hidup. Literally.

4. HUGE ECOSYSTEM
━━━━━━━━━━━━━━━━
3,000+ providers (AWS, GCP, Azure, Cloudflare, GitHub, DataDog, dll)
10,000+ modules di Terraform Registry
Stackoverflow questions, tutorials, blog posts — semua ada.

5. JOB MARKET
━━━━━━━━━━━━
Survey DevOps Tools 2025:
Terraform     ████████████████████████████████  paling banyak dicari
Ansible       ████████████████████████
CloudFormation ██████████████████
Pulumi        ██████████
```

### Terraform vs Ansible — Penting Dipahami

Banyak yang bingung Terraform vs Ansible. Ini dua tools yang **complementary**, bukan competitor:

```
Terraform = PROVISIONING
━━━━━━━━━━━━━━━━━━━━━━━
"Bikin server, network, database"
- Create EC2 instance
- Create RDS PostgreSQL
- Create VPC, subnet
- Create S3 bucket

Ansible = CONFIGURATION MANAGEMENT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"Install software, configure server"
- Install Nginx di EC2
- Copy config files
- Start services
- Update packages

Real-world workflow:
1. Terraform → bikin EC2 instance
2. Ansible → install dan configure software di EC2 tersebut

Tapi di era container (Docker + K8s), Ansible less relevant:
- Docker image udah include semua software
- K8s handle configuration via ConfigMap/Secret
- Terraform provision infra, Docker/K8s handle app

Jadi buat kita: TERRAFORM AJA CUKUP.
```

### Fun Fact

Terraform ditulis pakai **Go**. Jadi kamu belajar Go, Docker (Go), Kubernetes (Go), dan Terraform (Go) — semua tools dalam satu bahasa. HashiCorp (pembuat Terraform) itu salah satu perusahaan yang paling heavy pakai Go.

---

## 17.3 Terraform Architecture

### Core Concepts

Terraform punya 5 konsep utama yang harus kamu paham:

```
┌────────────────────────────────────────────────────────────────────┐
│                      TERRAFORM ARCHITECTURE                         │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     CONFIGURATION FILES                       │  │
│  │                                                               │  │
│  │  main.tf          variables.tf       outputs.tf              │  │
│  │  ┌──────────┐    ┌──────────┐       ┌──────────┐            │  │
│  │  │ provider  │    │ variable  │       │ output   │            │  │
│  │  │ resource  │    │ variable  │       │ output   │            │  │
│  │  │ data      │    │ variable  │       │          │            │  │
│  │  │ module    │    │          │       │          │            │  │
│  │  └──────────┘    └──────────┘       └──────────┘            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                    terraform init                                   │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                        PROVIDERS                              │  │
│  │                                                               │  │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐         │  │
│  │  │ AWS  │  │ GCP  │  │Azure │  │GitHub│  │ K8s  │         │  │
│  │  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘         │  │
│  │                                                               │  │
│  │  Provider = plugin yang tau cara ngomong ke API cloud         │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                    terraform plan                                   │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                       EXECUTION PLAN                          │  │
│  │                                                               │  │
│  │  + aws_instance.app        (CREATE)                          │  │
│  │  ~ aws_security_group.web  (MODIFY)                          │  │
│  │  - aws_s3_bucket.old       (DESTROY)                         │  │
│  │                                                               │  │
│  │  "Preview apa yang bakal berubah"                            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                    terraform apply                                  │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                       STATE FILE                              │  │
│  │                    (terraform.tfstate)                         │  │
│  │                                                               │  │
│  │  {                                                            │  │
│  │    "resources": [                                             │  │
│  │      { "type": "aws_instance", "id": "i-abc123" },          │  │
│  │      { "type": "aws_security_group", "id": "sg-xyz789" }    │  │
│  │    ]                                                          │  │
│  │  }                                                            │  │
│  │                                                               │  │
│  │  "Database" Terraform — track apa yang udah dibuat            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                              │                                      │
│                              ▼                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     CLOUD RESOURCES                           │  │
│  │                                                               │  │
│  │  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐               │  │
│  │  │EC2 │ │RDS │ │S3  │ │VPC │ │IAM │ │EKS │               │  │
│  │  └────┘ └────┘ └────┘ └────┘ └────┘ └────┘               │  │
│  │                                                               │  │
│  │  Real resources di AWS                                       │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
```

### 1. Provider

**Provider** itu plugin yang bisa ngomong ke API cloud tertentu. Mau bikin resource di AWS? Pakai `aws` provider. Mau di GCP? Pakai `google` provider.

```
Provider = Translator/Penerjemah
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Kamu ngomong HCL → Provider translate ke AWS API calls
- aws provider → AWS API
- google provider → GCP API
- azurerm provider → Azure API
- kubernetes provider → K8s API
- github provider → GitHub API
```

### 2. Resource

**Resource** itu infrastruktur yang kamu mau bikin. EC2 instance, RDS database, S3 bucket — semua itu resource.

```hcl
# Resource = "bikin ini di cloud"
resource "aws_instance" "app" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.small"
}

# Format: resource "TIPE" "NAMA_LOKAL" { ... }
# TIPE       = aws_instance (dari provider AWS)
# NAMA_LOKAL = app (nama referensi di Terraform, bukan nama di AWS)
```

### 3. State

**State file** (`terraform.tfstate`) itu "database" Terraform. Dia track semua resource yang udah dibuat dan mapping-nya ke resource real di cloud.

```
State = "Catatan inventaris"
━━━━━━━━━━━━━━━━━━━━━━━━━━
"EC2 instance 'app' yang ada di file konfigurasi
 itu mapping-nya ke i-abc123 di AWS"

Tanpa state, Terraform ga tau resource mana yang dia manage.
Kalau state hilang → Terraform "lupa" semua resource-nya → chaos.
```

### 4. Plan

**Plan** itu preview perubahan. Terraform compare konfigurasi kamu vs state file vs real resources, lalu tunjukin apa yang bakal berubah.

```
Plan = "Preview sebelum apply"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
+ = CREATE (resource baru)
~ = UPDATE (modify resource existing)
- = DESTROY (hapus resource)

Kayak git diff tapi untuk infrastructure.
```

### 5. Apply

**Apply** itu eksekusi plan. Terraform beneran bikin/ubah/hapus resource di cloud.

### The Flow — Write, Plan, Apply

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│  WRITE  │ ──→ │  PLAN   │ ──→ │  APPLY   │
│         │     │         │     │          │
│ Edit    │     │ Preview │     │ Execute  │
│ .tf     │     │ changes │     │ changes  │
│ files   │     │         │     │          │
└─────────┘     └─────────┘     └──────────┘

1. WRITE  → Edit file .tf (tambah resource, ubah config)
2. PLAN   → terraform plan (lihat apa yang bakal berubah)
3. APPLY  → terraform apply (eksekusi perubahan)

Loop ini terus berulang setiap kali kamu mau ubah infrastructure.
```

---

## 17.4 Install Terraform

### macOS (Homebrew)

```bash
# Install Terraform via Homebrew
brew tap hashicorp/tap
brew install hashicorp/tap/terraform

# Verify installation
terraform version
# Terraform v1.12.x on darwin_arm64
```

### Linux

```bash
# Ubuntu/Debian
sudo apt-get update && sudo apt-get install -y gnupg software-properties-common

# Add HashiCorp GPG key
wget -O- https://apt.releases.hashicorp.com/gpg | \
  gpg --dearmor | \
  sudo tee /usr/share/keyrings/hashicorp-archive-keyring.gpg > /dev/null

# Add repository
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] \
  https://apt.releases.hashicorp.com $(lsb_release -cs) main" | \
  sudo tee /etc/apt/sources.list.d/hashicorp.list

# Install
sudo apt update && sudo apt install terraform

# Verify
terraform version
```

### Windows

```bash
# Via Chocolatey
choco install terraform

# Via Scoop
scoop install terraform

# Verify
terraform version
```

### IDE Extensions

```
VS Code Extensions untuk Terraform:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. HashiCorp Terraform (hashicorp.terraform)
   - Syntax highlighting
   - Auto-completion
   - Format on save
   - Go to definition
   - Terraform validate integration
   → WAJIB install

2. Terraform Doc Snippets
   - Quick snippets untuk resource blocks
   - Shortcut: tf-aws-instance → generate resource block

Setup VS Code settings.json:
{
  "[terraform]": {
    "editor.defaultFormatter": "hashicorp.terraform",
    "editor.formatOnSave": true,
    "editor.formatOnSaveMode": "file"
  },
  "[terraform-vars]": {
    "editor.defaultFormatter": "hashicorp.terraform",
    "editor.formatOnSave": true,
    "editor.formatOnSaveMode": "file"
  }
}
```

### Enable Autocomplete (Optional)

```bash
# Install autocomplete
terraform -install-autocomplete

# Restart shell
exec $SHELL
```

---

## 17.5 HCL Basics — HashiCorp Configuration Language

### Apa Itu HCL?

HCL (HashiCorp Configuration Language) itu bahasa konfigurasi yang didesain untuk **human-readable** dan **machine-friendly**. Bukan programming language — ini configuration language.

```
HCL vs JSON vs YAML:
━━━━━━━━━━━━━━━━━━━

JSON:
{
  "resource": {
    "aws_instance": {
      "app": {
        "ami": "ami-0abcdef1234567890",
        "instance_type": "t3.small",
        "tags": {
          "Name": "task-manager"
        }
      }
    }
  }
}
→ Verbose, ga bisa comment, susah dibaca

YAML:
resource:
  aws_instance:
    app:
      ami: ami-0abcdef1234567890
      instance_type: t3.small
      tags:
        Name: task-manager
→ Indentation-sensitive (error kalau salah spasi)

HCL:
resource "aws_instance" "app" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.small"

  tags = {
    Name = "task-manager"
  }
}
→ Clean, readable, bisa comment, type-safe
```

### Blocks

Block itu unit dasar HCL. Setiap block punya **type**, **labels**, dan **body**.

```hcl
# Format block:
# BLOCK_TYPE "LABEL_1" "LABEL_2" {
#   ARGUMENT = VALUE
# }

# Contoh: resource block
resource "aws_instance" "app" {       # block type: resource
                                       # label 1: aws_instance (resource type)
                                       # label 2: app (local name)
  ami           = "ami-0abcdef1234567890"  # argument
  instance_type = "t3.small"               # argument

  tags = {                                 # nested block
    Name = "task-manager"
  }
}

# Contoh: variable block
variable "instance_type" {            # block type: variable
                                       # label: instance_type (variable name)
  type        = string                # argument
  default     = "t3.small"            # argument
  description = "EC2 instance type"   # argument
}

# Contoh: output block
output "public_ip" {                  # block type: output
                                       # label: public_ip (output name)
  value       = aws_instance.app.public_ip  # argument (reference)
  description = "Public IP of EC2"          # argument
}
```

### Arguments & Expressions

```hcl
# ARGUMENTS
# Simple values
name     = "task-manager"           # string
count    = 3                        # number
enabled  = true                     # boolean

# Lists
ports    = [80, 443, 8080]          # list of numbers
names    = ["web", "api", "worker"] # list of strings

# Maps
tags = {                             # map
  Name        = "task-manager"
  Environment = "production"
  Project     = "task-manager-api"
}

# EXPRESSIONS & REFERENCES
# Reference resource attribute
subnet_id     = aws_subnet.public.id
# Format: RESOURCE_TYPE.LOCAL_NAME.ATTRIBUTE

# Reference variable
instance_type = var.instance_type

# Reference data source
ami           = data.aws_ami.ubuntu.id

# String interpolation
name = "task-manager-${var.environment}"

# Conditional
instance_type = var.environment == "production" ? "t3.medium" : "t3.micro"
```

### Comments

```hcl
# Single line comment (hash)

// Single line comment (double slash) — juga valid

/*
  Multi-line comment
  Bisa spanning beberapa baris
*/

resource "aws_instance" "app" {
  # Instance type depends on environment
  instance_type = var.instance_type  // inline comment juga bisa
}
```

### Variables

```hcl
# variables.tf — define variables

variable "project_name" {
  type        = string
  default     = "task-manager"
  description = "Name of the project"
}

variable "environment" {
  type        = string
  description = "Environment (dev, staging, production)"

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "instance_count" {
  type    = number
  default = 1
}

variable "enable_monitoring" {
  type    = bool
  default = true
}

variable "allowed_ports" {
  type    = list(number)
  default = [80, 443, 8080]
}

variable "common_tags" {
  type = map(string)
  default = {
    Project   = "task-manager"
    ManagedBy = "terraform"
  }
}

# Complex type
variable "app_config" {
  type = object({
    name     = string
    port     = number
    replicas = number
    env_vars = map(string)
  })
  default = {
    name     = "task-manager"
    port     = 8080
    replicas = 2
    env_vars = {
      APP_ENV = "production"
    }
  }
}
```

### Locals

```hcl
# locals = computed values, ga bisa di-override dari luar

locals {
  # String computation
  full_name = "${var.project_name}-${var.environment}"

  # Conditional
  instance_type = var.environment == "production" ? "t3.medium" : "t3.micro"

  # Merge tags
  common_tags = merge(var.common_tags, {
    Environment = var.environment
    Name        = local.full_name
  })
}

# Usage
resource "aws_instance" "app" {
  instance_type = local.instance_type
  tags          = local.common_tags
}
```

### Outputs

```hcl
# outputs.tf — expose values after apply

output "instance_id" {
  value       = aws_instance.app.id
  description = "ID of the EC2 instance"
}

output "public_ip" {
  value       = aws_instance.app.public_ip
  description = "Public IP of the EC2 instance"
}

output "database_endpoint" {
  value       = aws_db_instance.postgres.endpoint
  description = "RDS endpoint"
  sensitive   = true  # ga ditampilkan di console output
}
```

---

## 17.6 Provider Configuration

### AWS Provider Setup

```hcl
# providers.tf

terraform {
  required_version = ">= 1.12.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"  # versi 5.x (compatible updates)
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "task-manager"
      ManagedBy   = "terraform"
      Environment = var.environment
    }
  }
}
```

### Version Constraints

```hcl
# Version constraint syntax:
# = 5.0.0     → exactly 5.0.0
# >= 5.0.0    → 5.0.0 or newer
# ~> 5.0      → >= 5.0, < 6.0 (minor version updates OK)
# ~> 5.0.0    → >= 5.0.0, < 5.1.0 (patch updates only)
# >= 5.0, < 6.0 → between 5.0 and 6.0

# BEST PRACTICE: pakai ~> untuk lock major version
required_providers {
  aws = {
    source  = "hashicorp/aws"
    version = "~> 5.0"  # aman dari breaking changes
  }
}
```

### AWS Credentials — Cara Configure

Terraform butuh AWS credentials untuk manage resources. Ada beberapa cara:

```
Cara 1: AWS CLI Shared Credentials (RECOMMENDED untuk development)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Install AWS CLI
brew install awscli  # macOS

# Configure
aws configure
# AWS Access Key ID: AKIAIOSFODNN7EXAMPLE
# AWS Secret Access Key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
# Default region name: ap-southeast-1
# Default output format: json

# File yang dibuat: ~/.aws/credentials
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Terraform otomatis baca file ini. Ga perlu config apa-apa di .tf file.
```

```
Cara 2: Environment Variables (RECOMMENDED untuk CI/CD)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="ap-southeast-1"

# Terraform otomatis baca env vars ini
```

```
Cara 3: IAM Role (RECOMMENDED untuk production / EC2 / ECS)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Kalau Terraform dijalankan dari EC2 instance, attach IAM role ke instance.
# Terraform otomatis pakai role credentials. No access key needed.
# INI CARA PALING SECURE.
```

```
Cara 4: AWS Profile (kalau punya multiple AWS accounts)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# ~/.aws/credentials
[default]
aws_access_key_id = AKIADEFAULT...
aws_secret_access_key = ...

[staging]
aws_access_key_id = AKIASTAGING...
aws_secret_access_key = ...

[production]
aws_access_key_id = AKIAPRODUCTION...
aws_secret_access_key = ...

# Di Terraform:
provider "aws" {
  region  = "ap-southeast-1"
  profile = "staging"  # pakai credentials profile "staging"
}

# Atau via env var:
export AWS_PROFILE=staging
```

> **PENTING:** JANGAN PERNAH hardcode access keys di file .tf! Selalu pakai shared credentials, env vars, atau IAM roles.

---

## 17.7 First Resource — EC2 Instance

Sekarang kita praktek. Bikin EC2 instance pertama pakai Terraform.

### Project Setup

```bash
# Bikin directory untuk Terraform project
mkdir -p ~/terraform-learn/first-resource
cd ~/terraform-learn/first-resource
```

### File: main.tf

```hcl
# main.tf — First Terraform resource

terraform {
  required_version = ">= 1.12.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-southeast-1"  # Singapore
}

# Data source: cari AMI Ubuntu terbaru
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]  # Canonical (pembuat Ubuntu)

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Resource: EC2 instance
resource "aws_instance" "my_first_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name      = "my-first-terraform-server"
    ManagedBy = "terraform"
  }
}

# Output: tampilkan Public IP setelah instance dibuat
output "instance_id" {
  value       = aws_instance.my_first_server.id
  description = "Instance ID"
}

output "public_ip" {
  value       = aws_instance.my_first_server.public_ip
  description = "Public IP address"
}
```

### Step 1: terraform init

```bash
# Initialize Terraform — download provider plugins
terraform init

# Output:
# Initializing the backend...
# Initializing provider plugins...
# - Finding hashicorp/aws versions matching "~> 5.0"...
# - Installing hashicorp/aws v5.xx.x...
# - Installed hashicorp/aws v5.xx.x (signed by HashiCorp)
#
# Terraform has been successfully initialized!
```

```
Apa yang terjadi saat terraform init:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Baca required_providers block
2. Download provider plugins ke .terraform/ directory
3. Bikin .terraform.lock.hcl (lock file, kayak go.sum atau package-lock.json)

.terraform/          ← provider plugins (jangan commit ke git!)
.terraform.lock.hcl  ← lock file (COMMIT ini ke git!)
```

### Step 2: terraform plan

```bash
# Preview perubahan — DRY RUN
terraform plan

# Output:
# Terraform used the selected providers to generate the following execution plan.
#
# data.aws_ami.ubuntu: Reading...
# data.aws_ami.ubuntu: Read complete after 1s
#
# Terraform will perform the following actions:
#
#   # aws_instance.my_first_server will be created
#   + resource "aws_instance" "my_first_server" {
#       + ami                          = "ami-0abcdef1234567890"
#       + arn                          = (known after apply)
#       + instance_type                = "t3.micro"
#       + id                           = (known after apply)
#       + public_ip                    = (known after apply)
#       + tags                         = {
#           + "ManagedBy" = "terraform"
#           + "Name"      = "my-first-terraform-server"
#         }
#       + tags_all                     = {
#           + "ManagedBy" = "terraform"
#           + "Name"      = "my-first-terraform-server"
#         }
#       ...
#     }
#
# Plan: 1 to add, 0 to change, 0 to destroy.
#
# Changes to Outputs:
#   + instance_id = (known after apply)
#   + public_ip   = (known after apply)
```

```
Cara baca terraform plan output:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

+  = AKAN DIBUAT (create)
-  = AKAN DIHAPUS (destroy)
~  = AKAN DIUBAH (update in-place)
-/+ = AKAN DIHAPUS DAN DIBUAT ULANG (destroy then create / force replacement)

(known after apply) = nilainya baru tau setelah resource dibuat
                      (contoh: ID, IP address, ARN)

Plan: 1 to add, 0 to change, 0 to destroy.
       ↑ create    ↑ update      ↑ delete
```

### Step 3: terraform apply

```bash
# Apply — beneran bikin resource
terraform apply

# Terraform akan tampilkan plan lagi, lalu tanya konfirmasi:
# Do you want to perform these actions?
#   Terraform will perform the actions described above.
#   Only 'yes' will be accepted to approve.
#
# Enter a value: yes
#
# aws_instance.my_first_server: Creating...
# aws_instance.my_first_server: Still creating... [10s elapsed]
# aws_instance.my_first_server: Still creating... [20s elapsed]
# aws_instance.my_first_server: Creation complete after 32s [id=i-0abc123def456789]
#
# Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
#
# Outputs:
# instance_id = "i-0abc123def456789"
# public_ip   = "13.250.xxx.xxx"

# Atau skip konfirmasi (untuk automation):
terraform apply -auto-approve
```

Setelah apply, Terraform bikin file `terraform.tfstate` yang track resource yang udah dibuat.

### Step 4: terraform show

```bash
# Lihat state current
terraform show

# Lihat resource specific
terraform state list
# aws_instance.my_first_server

terraform state show aws_instance.my_first_server
```

### Step 5: terraform destroy

```bash
# Hapus SEMUA resource yang di-manage Terraform
terraform destroy

# Output:
# aws_instance.my_first_server: Destroying... [id=i-0abc123def456789]
# aws_instance.my_first_server: Still destroying... [10s elapsed]
# aws_instance.my_first_server: Destruction complete after 31s
#
# Destroy complete! Resources: 1 destroyed.

# Atau skip konfirmasi:
terraform destroy -auto-approve
```

### Files After terraform init/apply

```
my-terraform-project/
├── main.tf                  ← configuration kamu
├── .terraform/              ← provider plugins (JANGAN COMMIT)
│   └── providers/
│       └── registry.terraform.io/
│           └── hashicorp/
│               └── aws/
│                   └── 5.xx.x/
│                       └── darwin_arm64/
│                           └── terraform-provider-aws
├── .terraform.lock.hcl      ← lock file (COMMIT ke git)
├── terraform.tfstate        ← state file (JANGAN COMMIT, pindah ke remote)
└── terraform.tfstate.backup ← backup state (JANGAN COMMIT)
```

### .gitignore untuk Terraform

```gitignore
# .gitignore

# Local .terraform directories
**/.terraform/*

# .tfstate files — JANGAN COMMIT!
# State file berisi sensitive data (password, keys, dll)
*.tfstate
*.tfstate.*

# Crash log files
crash.log
crash.*.log

# Exclude override files
override.tf
override.tf.json
*_override.tf
*_override.tf.json

# Exclude tfvars files yang mungkin berisi secrets
*.auto.tfvars
*.auto.tfvars.json

# TAPI commit ini:
# - *.tf files (configuration)
# - .terraform.lock.hcl (lock file)
# - terraform.tfvars (kalau ga berisi secrets)
```

---

## 17.8 Variables — Parameterize Everything

### Variable Block

```hcl
# variables.tf

variable "aws_region" {
  type        = string
  default     = "ap-southeast-1"
  description = "AWS region to deploy resources"
}

variable "environment" {
  type        = string
  description = "Environment name (dev, staging, production)"

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "instance_type" {
  type        = string
  default     = "t3.micro"
  description = "EC2 instance type"
}

variable "db_password" {
  type        = string
  description = "Database password"
  sensitive   = true  # ga ditampilkan di plan/apply output
}
```

### Variable Types

```hcl
# STRING
variable "project_name" {
  type    = string
  default = "task-manager"
}

# NUMBER
variable "app_port" {
  type    = number
  default = 8080
}

# BOOLEAN
variable "enable_monitoring" {
  type    = bool
  default = true
}

# LIST
variable "availability_zones" {
  type    = list(string)
  default = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
}

# MAP
variable "instance_types" {
  type = map(string)
  default = {
    dev        = "t3.micro"
    staging    = "t3.small"
    production = "t3.medium"
  }
}

# OBJECT
variable "database_config" {
  type = object({
    engine         = string
    engine_version = string
    instance_class = string
    storage_gb     = number
    multi_az       = bool
  })
  default = {
    engine         = "postgres"
    engine_version = "16.6"
    instance_class = "db.t3.micro"
    storage_gb     = 20
    multi_az       = false
  }
}

# TUPLE
variable "cidr_blocks" {
  type    = tuple([string, string, string])
  default = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

# SET (like list tapi unique, unordered)
variable "allowed_ips" {
  type    = set(string)
  default = ["1.2.3.4/32", "5.6.7.8/32"]
}
```

### Cara Set Variable Values

```
4 cara set variable values (dari prioritas terendah ke tertinggi):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Default value (di variable block)
2. terraform.tfvars file (auto-loaded)
3. Environment variable (TF_VAR_xxx)
4. Command line flag (-var)
5. *.auto.tfvars file (auto-loaded)

Kalau ada conflict, yang prioritas lebih tinggi MENANG.
```

**Cara 1: Default value**

```hcl
variable "instance_type" {
  type    = string
  default = "t3.micro"  # default kalau ga di-set
}
```

**Cara 2: terraform.tfvars**

```hcl
# terraform.tfvars — auto-loaded oleh Terraform

aws_region    = "ap-southeast-1"
environment   = "production"
instance_type = "t3.small"
db_password   = "super-secret-password"
```

```bash
# Atau pakai file custom:
terraform apply -var-file="production.tfvars"
```

**Cara 3: Environment variables**

```bash
# Format: TF_VAR_<variable_name>
export TF_VAR_environment="production"
export TF_VAR_db_password="super-secret-password"
export TF_VAR_instance_type="t3.small"

terraform apply  # otomatis baca env vars
```

**Cara 4: Command line**

```bash
terraform apply -var="environment=production" -var="instance_type=t3.small"
```

### Sensitive Variables

```hcl
variable "db_password" {
  type        = string
  description = "Database master password"
  sensitive   = true  # PENTING!
}

# Saat terraform plan/apply, output-nya:
# + password = (sensitive value)
# Bukan:
# + password = "super-secret-password"
```

### Penggunaan di Resource

```hcl
# main.tf — pakai variables

provider "aws" {
  region = var.aws_region
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  tags = {
    Name        = "${var.project_name}-${var.environment}"
    Environment = var.environment
  }
}

# Pakai map variable
resource "aws_instance" "app_v2" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_types[var.environment]
  # dev → t3.micro, staging → t3.small, production → t3.medium
}
```

---

## 17.9 Outputs — Expose Values

### Output Block

```hcl
# outputs.tf

output "instance_id" {
  value       = aws_instance.app.id
  description = "ID of the EC2 instance"
}

output "public_ip" {
  value       = aws_instance.app.public_ip
  description = "Public IP of the EC2 instance"
}

output "public_dns" {
  value       = aws_instance.app.public_dns
  description = "Public DNS of the EC2 instance"
}

output "database_endpoint" {
  value       = aws_db_instance.postgres.endpoint
  description = "RDS PostgreSQL endpoint"
  sensitive   = true  # hide dari output
}

output "ssh_command" {
  value       = "ssh -i task-manager-key.pem ubuntu@${aws_instance.app.public_ip}"
  description = "SSH command to connect to the instance"
}
```

### Lihat Output Values

```bash
# Setelah terraform apply, output ditampilkan di akhir
# Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
#
# Outputs:
# instance_id = "i-0abc123def456789"
# public_ip = "13.250.xxx.xxx"
# ssh_command = "ssh -i task-manager-key.pem ubuntu@13.250.xxx.xxx"

# Lihat output kapan aja:
terraform output

# Lihat output specific:
terraform output public_ip
# "13.250.xxx.xxx"

# Output raw (tanpa quotes) — berguna untuk scripting:
terraform output -raw public_ip
# 13.250.xxx.xxx

# Output sebagai JSON:
terraform output -json
```

### Output untuk Module Reference

Output dari module bisa di-reference dari module lain. Ini penting untuk modular infrastructure:

```hcl
# Dari root module, akses output module networking:
module "networking" {
  source = "./modules/networking"
}

# Reference output networking di module compute:
module "compute" {
  source    = "./modules/compute"
  subnet_id = module.networking.public_subnet_id  # ← reference output
}
```

---

## 17.10 Data Sources — Query Existing Resources

### Apa Itu Data Source?

**Data source** itu cara Terraform **baca informasi** dari cloud tanpa bikin resource baru. Kayak `SELECT` di SQL — cuma baca, ga ubah.

```
Resource vs Data Source:
━━━━━━━━━━━━━━━━━━━━━━
resource = CREATE/UPDATE/DELETE (CRUD)
data     = READ only

resource "aws_instance" "app" { ... }     → BIKIN EC2 baru
data "aws_ami" "ubuntu" { ... }           → CARI AMI yang udah ada
data "aws_vpc" "default" { ... }          → AMBIL info VPC yang udah ada
data "aws_availability_zones" "az" { ... } → LIST AZ yang available
```

### Contoh Data Sources

```hcl
# 1. Cari AMI Ubuntu terbaru
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]  # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Usage:
resource "aws_instance" "app" {
  ami = data.aws_ami.ubuntu.id  # pakai AMI yang ditemukan
}

# 2. Ambil info VPC default
data "aws_vpc" "default" {
  default = true
}

# Usage:
resource "aws_security_group" "web" {
  vpc_id = data.aws_vpc.default.id
}

# 3. List availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Usage:
resource "aws_subnet" "public" {
  count             = length(data.aws_availability_zones.available.names)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

# 4. Ambil info AWS account
data "aws_caller_identity" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

# 5. Ambil region current
data "aws_region" "current" {}

output "region" {
  value = data.aws_region.current.name
}
```

### Kapan Pakai Data Source?

```
Pakai data source kalau:
━━━━━━━━━━━━━━━━━━━━━━
1. Kamu perlu info dari resource yang UDAH ADA (bukan kamu yang bikin)
2. Kamu mau cari AMI terbaru secara dinamis
3. Kamu mau reference resource dari Terraform state lain
4. Kamu mau ambil info account/region/AZ

Pakai resource kalau:
━━━━━━━━━━━━━━━━━━━━
1. Kamu mau BIKIN resource baru
2. Kamu mau Terraform MANAGE lifecycle resource tersebut
```

---

## 17.11 State Management

### Local State vs Remote State

Secara default, Terraform simpan state di file lokal (`terraform.tfstate`). Ini OK untuk belajar, tapi **BURUK** untuk team work.

```
Local State — PROBLEM:
━━━━━━━━━━━━━━━━━━━━━

Developer A: terraform apply (state di laptop A)
Developer B: terraform apply (state di laptop B)
→ State CONFLICT! Siapa yang punya "versi benar"?

State file berisi SEMUA info resource, termasuk:
- Database password (plaintext!)
- API keys
- Private IPs
Simpen di laptop = security risk.

Kalau laptop rusak / hilang:
→ State file hilang
→ Terraform "lupa" semua resource
→ Ga bisa manage infrastructure yang udah ada
→ Harus terraform import satu-satu (nightmare)
```

```
Remote State — SOLUTION:
━━━━━━━━━━━━━━━━━━━━━━━

State disimpan di remote backend (S3, GCS, Terraform Cloud).
Semua developer baca/tulis ke state yang SAMA.

Benefits:
1. TEAM COLLABORATION — semua orang pakai state yang sama
2. LOCKING — kalau A lagi apply, B ga bisa apply (prevent conflict)
3. ENCRYPTION — state encrypted at rest di S3
4. VERSIONING — S3 versioning = bisa rollback state
5. BACKUP — ga hilang kalau laptop rusak
```

### Setup Remote State — S3 + DynamoDB

**Step 1: Bikin S3 bucket dan DynamoDB table** (ini satu-satunya resource yang kamu bikin manual atau pakai Terraform terpisah)

```hcl
# backend-setup/main.tf — Run ini SATU KALI

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-southeast-1"
}

# S3 bucket untuk state file
resource "aws_s3_bucket" "terraform_state" {
  bucket = "task-manager-terraform-state"  # harus globally unique

  lifecycle {
    prevent_destroy = true  # jangan sampai ke-delete!
  }
}

# Enable versioning — bisa rollback state
resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Enable encryption — state berisi sensitive data
resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

# Block public access — JANGAN sampai public!
resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# DynamoDB table untuk state locking
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "task-manager-terraform-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}

output "s3_bucket" {
  value = aws_s3_bucket.terraform_state.id
}

output "dynamodb_table" {
  value = aws_dynamodb_table.terraform_locks.name
}
```

```bash
cd backend-setup
terraform init
terraform apply
```

**Step 2: Configure backend di project utama**

```hcl
# main.tf — project utama

terraform {
  required_version = ">= 1.12.0"

  # Remote backend
  backend "s3" {
    bucket         = "task-manager-terraform-state"
    key            = "task-manager/production/terraform.tfstate"
    region         = "ap-southeast-1"
    dynamodb_table = "task-manager-terraform-locks"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
```

```bash
# Re-init untuk switch ke remote backend
terraform init

# Terraform akan tanya:
# "Do you want to copy existing state to the new backend?"
# Enter a value: yes
```

### State Locking — Prevent Conflicts

```
Tanpa locking:
━━━━━━━━━━━━━

Developer A: terraform apply (mulai)
Developer B: terraform apply (mulai di waktu yang sama)
→ KEDUA-DUANYA modify state file → CORRUPT!

Dengan locking (DynamoDB):
━━━━━━━━━━━━━━━━━━━━━━━━

Developer A: terraform apply → LOCK acquired (DynamoDB entry created)
Developer B: terraform apply → "Error: Error acquiring the state lock"
→ B harus tunggu A selesai → SAFE!

Developer A: terraform apply complete → LOCK released
Developer B: terraform apply → LOCK acquired → proceed
```

### State Commands

```bash
# List semua resources di state
terraform state list

# Show detail resource tertentu
terraform state show aws_instance.app

# Pindahkan resource (rename)
terraform state mv aws_instance.app aws_instance.web

# Hapus resource dari state (Terraform "lupa" tapi resource di AWS tetap ada)
terraform state rm aws_instance.old_server

# Pull remote state ke lokal (untuk inspect)
terraform state pull > state.json
```

---

## 17.12 Modules — Reusable Infrastructure Components

### Apa Itu Module?

Module itu **package** yang group beberapa resource jadi satu unit reusable. Kayak function di programming — define sekali, pakai berkali-kali.

```
Tanpa module:
━━━━━━━━━━━━
main.tf → 500 baris
→ VPC, subnet, security group, EC2, RDS, S3, IAM... semua di 1 file
→ Susah dibaca, susah di-maintain, ga reusable

Dengan module:
━━━━━━━━━━━━━
modules/
  networking/  → VPC, subnets, NAT gateway, route tables
  compute/     → EC2, ASG, ALB
  database/    → RDS PostgreSQL
  cache/       → ElastiCache Redis
  storage/     → S3, ECR

main.tf → 50 baris
→ Panggil module, pass variables
→ Clean, reusable, testable
```

### Module Structure

```
modules/networking/
├── main.tf       ← resource definitions
├── variables.tf  ← input variables
├── outputs.tf    ← output values
└── README.md     ← documentation (optional)
```

### Create Custom Module — VPC Module

```hcl
# modules/networking/variables.tf

variable "project_name" {
  type        = string
  description = "Project name for resource naming"
}

variable "environment" {
  type        = string
  description = "Environment (dev, staging, production)"
}

variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "CIDR block for VPC"
}

variable "public_subnet_cidrs" {
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
  description = "CIDR blocks for public subnets"
}

variable "private_subnet_cidrs" {
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.20.0/24"]
  description = "CIDR blocks for private subnets"
}
```

```hcl
# modules/networking/main.tf

data "aws_availability_zones" "available" {
  state = "available"
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-${var.environment}-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-${var.environment}-igw"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-${var.environment}-public-${count.index + 1}"
    Type = "public"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${var.project_name}-${var.environment}-private-${count.index + 1}"
    Type = "private"
  }
}

# Elastic IP for NAT Gateway
resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name = "${var.project_name}-${var.environment}-nat-eip"
  }
}

# NAT Gateway (di public subnet, supaya private subnet bisa akses internet)
resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name = "${var.project_name}-${var.environment}-nat"
  }

  depends_on = [aws_internet_gateway.main]
}

# Route Table — Public
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-public-rt"
  }
}

# Route Table — Private
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-private-rt"
  }
}

# Route Table Association — Public
resource "aws_route_table_association" "public" {
  count = length(var.public_subnet_cidrs)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Route Table Association — Private
resource "aws_route_table_association" "private" {
  count = length(var.private_subnet_cidrs)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
```

```hcl
# modules/networking/outputs.tf

output "vpc_id" {
  value       = aws_vpc.main.id
  description = "VPC ID"
}

output "vpc_cidr" {
  value       = aws_vpc.main.cidr_block
  description = "VPC CIDR block"
}

output "public_subnet_ids" {
  value       = aws_subnet.public[*].id
  description = "List of public subnet IDs"
}

output "private_subnet_ids" {
  value       = aws_subnet.private[*].id
  description = "List of private subnet IDs"
}

output "nat_gateway_ip" {
  value       = aws_eip.nat.public_ip
  description = "NAT Gateway public IP"
}
```

### Use Module

```hcl
# main.tf — root module

module "networking" {
  source = "./modules/networking"

  project_name         = var.project_name
  environment          = var.environment
  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]
}

# Reference module output
resource "aws_instance" "app" {
  subnet_id = module.networking.public_subnet_ids[0]
  # ...
}
```

### Community Modules dari Terraform Registry

```hcl
# Pakai module dari Terraform Registry — ga perlu nulis sendiri!

# VPC module (community, paling populer)
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.project_name}-${var.environment}"
  cidr = "10.0.0.0/16"

  azs             = ["ap-southeast-1a", "ap-southeast-1b"]
  public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets = ["10.0.10.0/24", "10.0.20.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true  # hemat cost — 1 NAT aja

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}

# Security Group module
module "app_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name        = "${var.project_name}-app-sg"
  description = "Security group for Task Manager app"
  vpc_id      = module.vpc.vpc_id

  ingress_with_cidr_blocks = [
    {
      from_port   = 8080
      to_port     = 8080
      protocol    = "tcp"
      cidr_blocks = "0.0.0.0/0"
    },
    {
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = "YOUR_IP/32"
    },
  ]

  egress_rules = ["all-all"]
}
```

---

## 17.13 Task Manager Infrastructure — Full Terraform Code

Sekarang kita bikin **full Terraform code** untuk provision seluruh AWS infrastructure Task Manager API. Ini yang beneran kamu pake di production.

### Project Structure

```
terraform/
├── main.tf                 ← root module, call sub-modules
├── variables.tf            ← root variables
├── outputs.tf              ← root outputs
├── providers.tf            ← provider config + backend
├── terraform.tfvars        ← variable values (per environment)
├── locals.tf               ← computed locals
├── modules/
│   ├── networking/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── security/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── compute/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── database/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── cache/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── storage/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
└── environments/
    ├── dev.tfvars
    ├── staging.tfvars
    └── production.tfvars
```

### Root Files

```hcl
# providers.tf

terraform {
  required_version = ">= 1.12.0"

  backend "s3" {
    bucket         = "task-manager-terraform-state"
    key            = "task-manager/terraform.tfstate"
    region         = "ap-southeast-1"
    dynamodb_table = "task-manager-terraform-locks"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
```

```hcl
# variables.tf

variable "project_name" {
  type        = string
  default     = "task-manager"
  description = "Project name"
}

variable "environment" {
  type        = string
  description = "Environment (dev, staging, production)"

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "aws_region" {
  type        = string
  default     = "ap-southeast-1"
  description = "AWS region"
}

# Networking
variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "VPC CIDR block"
}

# Compute
variable "instance_type" {
  type        = string
  default     = "t3.micro"
  description = "EC2 instance type"
}

variable "key_pair_name" {
  type        = string
  description = "SSH key pair name"
}

# Database
variable "db_instance_class" {
  type        = string
  default     = "db.t3.micro"
  description = "RDS instance class"
}

variable "db_name" {
  type        = string
  default     = "task_manager"
  description = "Database name"
}

variable "db_username" {
  type        = string
  default     = "postgres"
  description = "Database username"
}

variable "db_password" {
  type        = string
  description = "Database password"
  sensitive   = true
}

# Cache
variable "cache_node_type" {
  type        = string
  default     = "cache.t3.micro"
  description = "ElastiCache node type"
}

# App
variable "app_port" {
  type        = number
  default     = 8080
  description = "Application port"
}

variable "allowed_ssh_cidr" {
  type        = string
  default     = "0.0.0.0/0"
  description = "CIDR allowed for SSH access"
}
```

```hcl
# locals.tf

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.10.0/24", "10.0.20.0/24"]
}
```

```hcl
# main.tf — root module

# ==========================================
# NETWORKING
# ==========================================
module "networking" {
  source = "./modules/networking"

  project_name         = var.project_name
  environment          = var.environment
  vpc_cidr             = var.vpc_cidr
  public_subnet_cidrs  = local.public_subnet_cidrs
  private_subnet_cidrs = local.private_subnet_cidrs
}

# ==========================================
# SECURITY
# ==========================================
module "security" {
  source = "./modules/security"

  project_name     = var.project_name
  environment      = var.environment
  vpc_id           = module.networking.vpc_id
  vpc_cidr         = var.vpc_cidr
  app_port         = var.app_port
  allowed_ssh_cidr = var.allowed_ssh_cidr
}

# ==========================================
# DATABASE (RDS PostgreSQL)
# ==========================================
module "database" {
  source = "./modules/database"

  project_name       = var.project_name
  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  db_sg_id           = module.security.db_sg_id
  db_instance_class  = var.db_instance_class
  db_name            = var.db_name
  db_username        = var.db_username
  db_password        = var.db_password
}

# ==========================================
# CACHE (ElastiCache Redis)
# ==========================================
module "cache" {
  source = "./modules/cache"

  project_name       = var.project_name
  environment        = var.environment
  private_subnet_ids = module.networking.private_subnet_ids
  cache_sg_id        = module.security.cache_sg_id
  cache_node_type    = var.cache_node_type
}

# ==========================================
# STORAGE (S3 + ECR)
# ==========================================
module "storage" {
  source = "./modules/storage"

  project_name = var.project_name
  environment  = var.environment
}

# ==========================================
# COMPUTE (EC2)
# ==========================================
module "compute" {
  source = "./modules/compute"

  project_name      = var.project_name
  environment       = var.environment
  instance_type     = var.instance_type
  key_pair_name     = var.key_pair_name
  public_subnet_id  = module.networking.public_subnet_ids[0]
  app_sg_id         = module.security.app_sg_id
  app_port          = var.app_port
  db_endpoint       = module.database.db_endpoint
  db_name           = var.db_name
  db_username       = var.db_username
  db_password       = var.db_password
  redis_endpoint    = module.cache.redis_endpoint
  ecr_repository_url = module.storage.ecr_repository_url
  iam_instance_profile = module.storage.iam_instance_profile_name
}
```

```hcl
# outputs.tf

# Networking
output "vpc_id" {
  value       = module.networking.vpc_id
  description = "VPC ID"
}

output "public_subnet_ids" {
  value       = module.networking.public_subnet_ids
  description = "Public subnet IDs"
}

# Compute
output "app_public_ip" {
  value       = module.compute.public_ip
  description = "Application server public IP"
}

output "ssh_command" {
  value       = "ssh -i ${var.key_pair_name}.pem ubuntu@${module.compute.public_ip}"
  description = "SSH command to connect"
}

# Database
output "db_endpoint" {
  value       = module.database.db_endpoint
  description = "RDS PostgreSQL endpoint"
  sensitive   = true
}

# Cache
output "redis_endpoint" {
  value       = module.cache.redis_endpoint
  description = "ElastiCache Redis endpoint"
  sensitive   = true
}

# Storage
output "ecr_repository_url" {
  value       = module.storage.ecr_repository_url
  description = "ECR repository URL"
}

output "s3_bucket_name" {
  value       = module.storage.s3_bucket_name
  description = "S3 bucket name"
}
```

### Module: Networking

```hcl
# modules/networking/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  type    = list(string)
  default = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  type    = list(string)
  default = ["10.0.10.0/24", "10.0.20.0/24"]
}
```

```hcl
# modules/networking/main.tf

data "aws_availability_zones" "available" {
  state = "available"
}

# ==========================================
# VPC
# ==========================================
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-${var.environment}-vpc"
  }
}

# ==========================================
# INTERNET GATEWAY
# ==========================================
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-${var.environment}-igw"
  }
}

# ==========================================
# PUBLIC SUBNETS
# ==========================================
resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-${var.environment}-public-${count.index + 1}"
    Type = "public"
  }
}

# ==========================================
# PRIVATE SUBNETS
# ==========================================
resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${var.project_name}-${var.environment}-private-${count.index + 1}"
    Type = "private"
  }
}

# ==========================================
# NAT GATEWAY
# ==========================================
resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name = "${var.project_name}-${var.environment}-nat-eip"
  }
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name = "${var.project_name}-${var.environment}-nat"
  }

  depends_on = [aws_internet_gateway.main]
}

# ==========================================
# ROUTE TABLES
# ==========================================

# Public route table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-public-rt"
  }
}

# Private route table
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-private-rt"
  }
}

# Associations
resource "aws_route_table_association" "public" {
  count = length(var.public_subnet_cidrs)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count = length(var.private_subnet_cidrs)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
```

```hcl
# modules/networking/outputs.tf

output "vpc_id" {
  value = aws_vpc.main.id
}

output "vpc_cidr" {
  value = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  value = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  value = aws_subnet.private[*].id
}

output "nat_gateway_ip" {
  value = aws_eip.nat.public_ip
}
```

### Module: Security

```hcl
# modules/security/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "app_port" {
  type    = number
  default = 8080
}

variable "allowed_ssh_cidr" {
  type    = string
  default = "0.0.0.0/0"
}
```

```hcl
# modules/security/main.tf

# ==========================================
# APP SECURITY GROUP
# ==========================================
resource "aws_security_group" "app" {
  name        = "${var.project_name}-${var.environment}-app-sg"
  description = "Security group for Task Manager app server"
  vpc_id      = var.vpc_id

  # SSH
  ingress {
    description = "SSH access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.allowed_ssh_cidr]
  }

  # App port
  ingress {
    description = "App port"
    from_port   = var.app_port
    to_port     = var.app_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # All outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-app-sg"
  }
}

# ==========================================
# DATABASE SECURITY GROUP
# ==========================================
resource "aws_security_group" "db" {
  name        = "${var.project_name}-${var.environment}-db-sg"
  description = "Security group for RDS PostgreSQL"
  vpc_id      = var.vpc_id

  # PostgreSQL — HANYA dari VPC
  ingress {
    description = "PostgreSQL from VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-db-sg"
  }
}

# ==========================================
# CACHE SECURITY GROUP
# ==========================================
resource "aws_security_group" "cache" {
  name        = "${var.project_name}-${var.environment}-cache-sg"
  description = "Security group for ElastiCache Redis"
  vpc_id      = var.vpc_id

  # Redis — HANYA dari VPC
  ingress {
    description = "Redis from VPC"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-cache-sg"
  }
}
```

```hcl
# modules/security/outputs.tf

output "app_sg_id" {
  value = aws_security_group.app.id
}

output "db_sg_id" {
  value = aws_security_group.db.id
}

output "cache_sg_id" {
  value = aws_security_group.cache.id
}
```

### Module: Database (RDS PostgreSQL)

```hcl
# modules/database/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "db_sg_id" {
  type = string
}

variable "db_instance_class" {
  type    = string
  default = "db.t3.micro"
}

variable "db_name" {
  type    = string
  default = "task_manager"
}

variable "db_username" {
  type    = string
  default = "postgres"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "db_engine_version" {
  type    = string
  default = "16.6"
}

variable "db_storage_gb" {
  type    = number
  default = 20
}

variable "multi_az" {
  type    = bool
  default = false
}
```

```hcl
# modules/database/main.tf

# DB Subnet Group — RDS butuh minimal 2 subnet di 2 AZ berbeda
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-db-subnet"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-db-subnet"
  }
}

# RDS PostgreSQL Instance
resource "aws_db_instance" "postgres" {
  identifier = "${var.project_name}-${var.environment}-db"

  # Engine
  engine         = "postgres"
  engine_version = var.db_engine_version
  instance_class = var.db_instance_class

  # Storage
  allocated_storage     = var.db_storage_gb
  max_allocated_storage = var.db_storage_gb * 2  # auto-scaling storage
  storage_type          = "gp3"
  storage_encrypted     = true

  # Database
  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  # Network
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.db_sg_id]
  publicly_accessible    = false  # JANGAN bisa diakses dari internet!

  # Availability
  multi_az = var.multi_az

  # Backup
  backup_retention_period = var.environment == "production" ? 7 : 1
  backup_window           = "03:00-04:00"     # UTC, jam 10-11 WIB pagi
  maintenance_window      = "sun:04:00-sun:05:00"

  # Monitoring
  performance_insights_enabled = var.environment == "production"

  # Deletion protection
  deletion_protection = var.environment == "production"
  skip_final_snapshot = var.environment != "production"

  final_snapshot_identifier = var.environment == "production" ? "${var.project_name}-${var.environment}-final-snapshot" : null

  tags = {
    Name = "${var.project_name}-${var.environment}-db"
  }
}
```

```hcl
# modules/database/outputs.tf

output "db_endpoint" {
  value       = aws_db_instance.postgres.endpoint
  description = "RDS endpoint (host:port)"
}

output "db_address" {
  value       = aws_db_instance.postgres.address
  description = "RDS hostname"
}

output "db_port" {
  value       = aws_db_instance.postgres.port
  description = "RDS port"
}

output "db_name" {
  value       = aws_db_instance.postgres.db_name
  description = "Database name"
}
```

### Module: Cache (ElastiCache Redis)

```hcl
# modules/cache/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "cache_sg_id" {
  type = string
}

variable "cache_node_type" {
  type    = string
  default = "cache.t3.micro"
}

variable "redis_engine_version" {
  type    = string
  default = "7.1"
}

variable "num_cache_nodes" {
  type    = number
  default = 1
}
```

```hcl
# modules/cache/main.tf

# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-cache-subnet"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-cache-subnet"
  }
}

# ElastiCache Redis Cluster
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.project_name}-${var.environment}-redis"
  engine               = "redis"
  engine_version       = var.redis_engine_version
  node_type            = var.cache_node_type
  num_cache_nodes      = var.num_cache_nodes
  port                 = 6379
  parameter_group_name = "default.redis7"

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [var.cache_sg_id]

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"

  # Snapshots (for production)
  snapshot_retention_limit = var.environment == "production" ? 3 : 0
  snapshot_window          = "04:00-05:00"

  tags = {
    Name = "${var.project_name}-${var.environment}-redis"
  }
}
```

```hcl
# modules/cache/outputs.tf

output "redis_endpoint" {
  value       = aws_elasticache_cluster.redis.cache_nodes[0].address
  description = "Redis endpoint"
}

output "redis_port" {
  value       = aws_elasticache_cluster.redis.cache_nodes[0].port
  description = "Redis port"
}
```

### Module: Storage (S3 + ECR)

```hcl
# modules/storage/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}
```

```hcl
# modules/storage/main.tf

# ==========================================
# S3 BUCKET — App assets, uploads, etc.
# ==========================================
resource "aws_s3_bucket" "app" {
  bucket = "${var.project_name}-${var.environment}-assets"

  tags = {
    Name = "${var.project_name}-${var.environment}-assets"
  }
}

resource "aws_s3_bucket_versioning" "app" {
  bucket = aws_s3_bucket.app.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app" {
  bucket = aws_s3_bucket.app.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "app" {
  bucket = aws_s3_bucket.app.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ==========================================
# ECR — Docker image repository
# ==========================================
resource "aws_ecr_repository" "app" {
  name                 = "${var.project_name}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-ecr"
  }
}

# ECR Lifecycle Policy — auto-delete old images
resource "aws_ecr_lifecycle_policy" "app" {
  repository = aws_ecr_repository.app.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ==========================================
# IAM ROLE — For EC2 to access ECR, S3
# ==========================================
resource "aws_iam_role" "app" {
  name = "${var.project_name}-${var.environment}-app-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-app-role"
  }
}

# Policy: ECR access
resource "aws_iam_role_policy" "ecr_access" {
  name = "${var.project_name}-${var.environment}-ecr-policy"
  role = aws_iam_role.app.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
        Resource = "*"
      }
    ]
  })
}

# Policy: S3 access
resource "aws_iam_role_policy" "s3_access" {
  name = "${var.project_name}-${var.environment}-s3-policy"
  role = aws_iam_role.app.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.app.arn,
          "${aws_s3_bucket.app.arn}/*"
        ]
      }
    ]
  })
}

# Instance Profile — attach IAM role ke EC2
resource "aws_iam_instance_profile" "app" {
  name = "${var.project_name}-${var.environment}-app-profile"
  role = aws_iam_role.app.name
}
```

```hcl
# modules/storage/outputs.tf

output "s3_bucket_name" {
  value = aws_s3_bucket.app.id
}

output "s3_bucket_arn" {
  value = aws_s3_bucket.app.arn
}

output "ecr_repository_url" {
  value = aws_ecr_repository.app.repository_url
}

output "iam_instance_profile_name" {
  value = aws_iam_instance_profile.app.name
}
```

### Module: Compute (EC2)

```hcl
# modules/compute/variables.tf

variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.micro"
}

variable "key_pair_name" {
  type = string
}

variable "public_subnet_id" {
  type = string
}

variable "app_sg_id" {
  type = string
}

variable "app_port" {
  type    = number
  default = 8080
}

variable "db_endpoint" {
  type = string
}

variable "db_name" {
  type = string
}

variable "db_username" {
  type = string
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "redis_endpoint" {
  type = string
}

variable "ecr_repository_url" {
  type = string
}

variable "iam_instance_profile" {
  type = string
}
```

```hcl
# modules/compute/main.tf

# Cari AMI Ubuntu terbaru
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]  # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# EC2 Instance
resource "aws_instance" "app" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = var.instance_type
  key_name             = var.key_pair_name
  subnet_id            = var.public_subnet_id
  iam_instance_profile = var.iam_instance_profile

  vpc_security_group_ids = [var.app_sg_id]

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    encrypted   = true
  }

  # User data script — auto-setup Docker dan deploy app
  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    app_port           = var.app_port
    db_host            = split(":", var.db_endpoint)[0]
    db_port            = "5432"
    db_name            = var.db_name
    db_username        = var.db_username
    db_password        = var.db_password
    redis_host         = var.redis_endpoint
    redis_port         = "6379"
    ecr_repository_url = var.ecr_repository_url
    environment        = var.environment
  }))

  tags = {
    Name = "${var.project_name}-${var.environment}-app"
  }
}
```

```bash
# modules/compute/user_data.sh

#!/bin/bash
set -e

# Update system
apt-get update -y
apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Install AWS CLI
apt-get install -y awscli

# Login ke ECR
aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin ${ecr_repository_url}

# Pull dan run app
docker pull ${ecr_repository_url}:latest

docker run -d \
  --name task-manager \
  --restart unless-stopped \
  -p ${app_port}:${app_port} \
  -e APP_ENV=${environment} \
  -e PORT=${app_port} \
  -e DB_HOST=${db_host} \
  -e DB_PORT=${db_port} \
  -e DB_NAME=${db_name} \
  -e DB_USER=${db_username} \
  -e DB_PASSWORD='${db_password}' \
  -e REDIS_HOST=${redis_host} \
  -e REDIS_PORT=${redis_port} \
  ${ecr_repository_url}:latest

echo "Task Manager API deployed successfully!"
```

```hcl
# modules/compute/outputs.tf

output "instance_id" {
  value = aws_instance.app.id
}

output "public_ip" {
  value = aws_instance.app.public_ip
}

output "public_dns" {
  value = aws_instance.app.public_dns
}
```

### Environment-Specific Variables

```hcl
# environments/dev.tfvars

environment       = "dev"
instance_type     = "t3.micro"
db_instance_class = "db.t3.micro"
cache_node_type   = "cache.t3.micro"
key_pair_name     = "task-manager-dev-key"
allowed_ssh_cidr  = "0.0.0.0/0"  # dev aja — production harus restrict
```

```hcl
# environments/staging.tfvars

environment       = "staging"
instance_type     = "t3.small"
db_instance_class = "db.t3.small"
cache_node_type   = "cache.t3.micro"
key_pair_name     = "task-manager-staging-key"
allowed_ssh_cidr  = "YOUR_OFFICE_IP/32"
```

```hcl
# environments/production.tfvars

environment       = "production"
instance_type     = "t3.medium"
db_instance_class = "db.t3.medium"
cache_node_type   = "cache.t3.small"
key_pair_name     = "task-manager-prod-key"
allowed_ssh_cidr  = "YOUR_OFFICE_IP/32"
```

### Deploy Commands

```bash
# Deploy development
terraform init
terraform plan -var-file="environments/dev.tfvars" -var="db_password=dev-password-123"
terraform apply -var-file="environments/dev.tfvars" -var="db_password=dev-password-123"

# Deploy staging
terraform plan -var-file="environments/staging.tfvars" -var="db_password=staging-password-456"
terraform apply -var-file="environments/staging.tfvars" -var="db_password=staging-password-456"

# Deploy production (hati-hati!)
terraform plan -var-file="environments/production.tfvars" -var="db_password=prod-super-secret"
terraform apply -var-file="environments/production.tfvars" -var="db_password=prod-super-secret"
```

### Infrastructure Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AWS (ap-southeast-1)                              │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    VPC (10.0.0.0/16)                                │ │
│  │                                                                     │ │
│  │  ┌──────────────────────────┐  ┌──────────────────────────┐       │ │
│  │  │    PUBLIC SUBNET          │  │    PUBLIC SUBNET          │       │ │
│  │  │    10.0.1.0/24           │  │    10.0.2.0/24           │       │ │
│  │  │    AZ: ap-southeast-1a   │  │    AZ: ap-southeast-1b   │       │ │
│  │  │                          │  │                          │       │ │
│  │  │  ┌──────────────────┐   │  │                          │       │ │
│  │  │  │  EC2 (App Server) │   │  │                          │       │ │
│  │  │  │  t3.micro/medium  │   │  │                          │       │ │
│  │  │  │  Docker + Go App  │   │  │                          │       │ │
│  │  │  │  Port 8080        │   │  │                          │       │ │
│  │  │  └──────────────────┘   │  │                          │       │ │
│  │  │                          │  │                          │       │ │
│  │  │  ┌──────────────────┐   │  │                          │       │ │
│  │  │  │  NAT Gateway      │   │  │                          │       │ │
│  │  │  └──────────────────┘   │  │                          │       │ │
│  │  └──────────┬───────────────┘  └──────────────────────────┘       │ │
│  │             │                                                      │ │
│  │  ┌──────────▼───────────────┐  ┌──────────────────────────┐       │ │
│  │  │    PRIVATE SUBNET        │  │    PRIVATE SUBNET        │       │ │
│  │  │    10.0.10.0/24          │  │    10.0.20.0/24          │       │ │
│  │  │    AZ: ap-southeast-1a   │  │    AZ: ap-southeast-1b   │       │ │
│  │  │                          │  │                          │       │ │
│  │  │  ┌──────────────────┐   │  │                          │       │ │
│  │  │  │  RDS PostgreSQL   │   │  │                          │       │ │
│  │  │  │  db.t3.micro      │   │  │                          │       │ │
│  │  │  │  Port 5432        │   │  │                          │       │ │
│  │  │  └──────────────────┘   │  │                          │       │ │
│  │  │                          │  │                          │       │ │
│  │  │  ┌──────────────────┐   │  │                          │       │ │
│  │  │  │  ElastiCache Redis│   │  │                          │       │ │
│  │  │  │  cache.t3.micro   │   │  │                          │       │ │
│  │  │  │  Port 6379        │   │  │                          │       │ │
│  │  │  └──────────────────┘   │  │                          │       │ │
│  │  └──────────────────────────┘  └──────────────────────────┘       │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                             │
│  │  S3      │  │  ECR     │  │  IAM     │                             │
│  │  Bucket  │  │  Repo    │  │  Role    │                             │
│  └──────────┘  └──────────┘  └──────────┘                             │
│                                                                          │
│  Internet Gateway                                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 17.14 Terraform Workspace — Multi-Environment

### Problem

Gimana kalau kamu mau punya **dev, staging, production** tapi pakai **code yang sama**? Copy-paste folder? Itu ga DRY.

### Workspace Solution

Terraform **workspace** itu fitur built-in untuk manage multiple environments dari satu set configuration files. Setiap workspace punya **state file sendiri**.

```bash
# Lihat workspace yang ada
terraform workspace list
# * default

# Bikin workspace baru
terraform workspace new dev
terraform workspace new staging
terraform workspace new production

# Pindah workspace
terraform workspace select dev

# Lihat workspace active
terraform workspace show
# dev
```

### Workspace di Code

```hcl
# Gunakan terraform.workspace untuk conditional logic

locals {
  environment = terraform.workspace

  instance_types = {
    dev        = "t3.micro"
    staging    = "t3.small"
    production = "t3.medium"
  }

  db_instance_classes = {
    dev        = "db.t3.micro"
    staging    = "db.t3.small"
    production = "db.t3.medium"
  }

  multi_az = {
    dev        = false
    staging    = false
    production = true
  }
}

resource "aws_instance" "app" {
  instance_type = local.instance_types[terraform.workspace]
  # dev → t3.micro, staging → t3.small, production → t3.medium
}

resource "aws_db_instance" "postgres" {
  instance_class = local.db_instance_classes[terraform.workspace]
  multi_az       = local.multi_az[terraform.workspace]
}
```

### Workspace vs Separate Directories

```
WORKSPACE approach:
━━━━━━━━━━━━━━━━━
terraform/
├── main.tf         ← 1 set code
├── variables.tf
└── outputs.tf

terraform workspace select dev → state: dev
terraform workspace select prod → state: prod

Pros: DRY, satu codebase
Cons: Semua environment share sama code
      Susah kalau environment beda signifikan

SEPARATE DIRECTORIES approach:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
terraform/
├── modules/        ← shared modules
├── dev/
│   ├── main.tf     ← reference modules
│   └── terraform.tfvars
├── staging/
│   ├── main.tf     ← reference modules
│   └── terraform.tfvars
└── production/
    ├── main.tf     ← reference modules
    └── terraform.tfvars

Pros: Setiap environment bisa beda
      Lebih explicit, more control
Cons: Sedikit duplication

RECOMMENDATION:
━━━━━━━━━━━━━━
- Small projects: Workspace cukup
- Large/complex projects: Separate directories + shared modules
- Enterprise: Separate directories (lebih aman, isolated state)
```

### Backend Key per Workspace

```hcl
# Kalau pakai workspace + remote backend:
backend "s3" {
  bucket         = "task-manager-terraform-state"
  key            = "task-manager/terraform.tfstate"  # Terraform auto-prefix dengan workspace
  region         = "ap-southeast-1"
  dynamodb_table = "task-manager-terraform-locks"
  encrypt        = true
}

# State file locations:
# dev:        task-manager/env:/dev/terraform.tfstate
# staging:    task-manager/env:/staging/terraform.tfstate
# production: task-manager/env:/production/terraform.tfstate
```

---

## 17.15 Terraform with Kubernetes — Provision EKS

Di Bab 16, kamu belajar Kubernetes. Sekarang kita provision **EKS cluster** pakai Terraform — bukan klik-klik di Console.

### EKS Cluster dengan Terraform

```hcl
# eks.tf — Provision EKS cluster

# IAM Role for EKS Cluster
resource "aws_iam_role" "eks_cluster" {
  name = "${var.project_name}-${var.environment}-eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster.name
}

# EKS Cluster
resource "aws_eks_cluster" "main" {
  name     = "${var.project_name}-${var.environment}"
  role_arn = aws_iam_role.eks_cluster.arn
  version  = "1.32"

  vpc_config {
    subnet_ids = concat(
      module.networking.public_subnet_ids,
      module.networking.private_subnet_ids
    )

    endpoint_private_access = true
    endpoint_public_access  = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
  ]

  tags = {
    Name = "${var.project_name}-${var.environment}-eks"
  }
}

# IAM Role for EKS Node Group
resource "aws_iam_role" "eks_nodes" {
  name = "${var.project_name}-${var.environment}-eks-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_nodes.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_nodes.name
}

resource "aws_iam_role_policy_attachment" "eks_ecr_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_nodes.name
}

# EKS Node Group
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.project_name}-${var.environment}-nodes"
  node_role_arn   = aws_iam_role.eks_nodes.arn
  subnet_ids      = module.networking.private_subnet_ids

  scaling_config {
    desired_size = var.environment == "production" ? 3 : 2
    max_size     = var.environment == "production" ? 5 : 3
    min_size     = var.environment == "production" ? 2 : 1
  }

  instance_types = [var.environment == "production" ? "t3.medium" : "t3.small"]

  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_ecr_policy,
  ]

  tags = {
    Name = "${var.project_name}-${var.environment}-node-group"
  }
}

# Outputs
output "eks_cluster_endpoint" {
  value = aws_eks_cluster.main.endpoint
}

output "eks_cluster_name" {
  value = aws_eks_cluster.main.name
}

output "eks_cluster_ca_certificate" {
  value     = aws_eks_cluster.main.certificate_authority[0].data
  sensitive = true
}
```

### Pakai Community Module (Lebih Simpel)

```hcl
# Pakai terraform-aws-modules/eks/aws — jauh lebih simpel

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = "${var.project_name}-${var.environment}"
  cluster_version = "1.32"

  vpc_id     = module.networking.vpc_id
  subnet_ids = module.networking.private_subnet_ids

  # Cluster endpoint access
  cluster_endpoint_public_access = true

  # EKS Managed Node Groups
  eks_managed_node_groups = {
    general = {
      desired_size = 2
      min_size     = 1
      max_size     = 5

      instance_types = ["t3.small"]
      capacity_type  = "ON_DEMAND"

      labels = {
        role = "general"
      }
    }
  }

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}
```

### Kubernetes + Helm Provider

Setelah EKS cluster di-provision, kamu bisa manage K8s resources langsung dari Terraform:

```hcl
# kubernetes-resources.tf

# Configure Kubernetes provider
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}

# Configure Helm provider
provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
    }
  }
}

# Deploy Nginx Ingress Controller via Helm
resource "helm_release" "nginx_ingress" {
  name       = "nginx-ingress"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  namespace  = "ingress-nginx"

  create_namespace = true

  set {
    name  = "controller.service.type"
    value = "LoadBalancer"
  }
}

# Deploy app namespace
resource "kubernetes_namespace" "app" {
  metadata {
    name = var.project_name
  }
}

# Deploy ConfigMap
resource "kubernetes_config_map" "app" {
  metadata {
    name      = "${var.project_name}-config"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    APP_ENV    = var.environment
    PORT       = tostring(var.app_port)
    DB_HOST    = module.database.db_address
    DB_PORT    = "5432"
    DB_NAME    = var.db_name
    REDIS_HOST = module.cache.redis_endpoint
    REDIS_PORT = "6379"
  }
}

# Deploy Secret
resource "kubernetes_secret" "app" {
  metadata {
    name      = "${var.project_name}-secret"
    namespace = kubernetes_namespace.app.metadata[0].name
  }

  data = {
    DB_PASSWORD = var.db_password
    JWT_SECRET  = var.jwt_secret
  }
}
```

---

## 17.16 CI/CD Integration — Terraform di GitHub Actions

### Workflow: Plan on PR, Apply on Merge

```
PR dibuat:
━━━━━━━━━
1. terraform fmt -check (format check)
2. terraform init
3. terraform validate
4. terraform plan
5. Post plan output as PR comment ← reviewer bisa lihat apa yang berubah!

PR di-merge ke main:
━━━━━━━━━━━━━━━━━━━
1. terraform init
2. terraform plan
3. terraform apply -auto-approve
```

### Full GitHub Actions Workflow

```yaml
# .github/workflows/terraform.yml

name: Terraform

on:
  push:
    branches: [main]
    paths:
      - 'terraform/**'
  pull_request:
    branches: [main]
    paths:
      - 'terraform/**'

env:
  TF_VERSION: '1.12.0'
  AWS_REGION: 'ap-southeast-1'

permissions:
  contents: read
  pull-requests: write  # untuk post comment di PR

jobs:
  # ==================================
  # JOB 1: Terraform Format & Validate
  # ==================================
  format:
    name: Format & Validate
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: terraform

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        run: terraform fmt -check -recursive
        continue-on-error: false

      - name: Terraform Init
        run: terraform init -backend=false

      - name: Terraform Validate
        run: terraform validate

  # ==================================
  # JOB 2: Terraform Plan (on PR)
  # ==================================
  plan:
    name: Plan
    runs-on: ubuntu-latest
    needs: format
    if: github.event_name == 'pull_request'
    defaults:
      run:
        working-directory: terraform

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        run: terraform init

      - name: Terraform Plan
        id: plan
        run: terraform plan -var-file="environments/production.tfvars" -no-color -out=tfplan
        continue-on-error: true

      - name: Post Plan to PR
        uses: actions/github-script@v7
        if: github.event_name == 'pull_request'
        env:
          PLAN: ${{ steps.plan.outputs.stdout }}
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const output = `#### Terraform Plan
            \`\`\`
            ${process.env.PLAN}
            \`\`\`

            *Pushed by: @${{ github.actor }}, Action: \`${{ github.event_name }}\`*`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            });

      - name: Plan Status
        if: steps.plan.outcome == 'failure'
        run: exit 1

  # ==================================
  # JOB 3: Terraform Apply (on merge)
  # ==================================
  apply:
    name: Apply
    runs-on: ubuntu-latest
    needs: format
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment: production  # require manual approval (optional)
    defaults:
      run:
        working-directory: terraform

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Init
        run: terraform init

      - name: Terraform Plan
        run: terraform plan -var-file="environments/production.tfvars" -out=tfplan

      - name: Terraform Apply
        run: terraform apply -auto-approve tfplan
```

### GitHub Secrets yang Perlu Disetting

```
Repository → Settings → Secrets and variables → Actions

Secrets yang dibutuhkan:
━━━━━━━━━━━━━━━━━━━━━━
AWS_ACCESS_KEY_ID       → IAM user access key
AWS_SECRET_ACCESS_KEY   → IAM user secret key
TF_VAR_db_password      → Database password (optional, bisa juga via tfvars)
```

### Atlantis — Alternative Approach

```
Atlantis itu tool yang automate Terraform workflow di PR:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Developer bikin PR yang ubah .tf files
2. Atlantis otomatis run terraform plan
3. Plan output di-post sebagai PR comment
4. Reviewer cek plan
5. Reviewer comment "atlantis apply" di PR
6. Atlantis run terraform apply
7. Done!

Benefits:
- Ga perlu setup GitHub Actions sendiri
- Built-in locking (1 apply at a time)
- PR-based workflow
- Team bisa review plan sebelum apply

Install: self-host di EC2 atau pakai managed (env0, Spacelift)
```

---

## 17.17 Import Existing Resources

### Scenario

Kamu udah punya resources di AWS yang dibuat manual (klik-klik Console). Sekarang mau manage pakai Terraform. Gimana?

### terraform import

```bash
# Format: terraform import <RESOURCE_ADDRESS> <RESOURCE_ID>

# Import EC2 instance
terraform import aws_instance.app i-0abc123def456789

# Import RDS instance
terraform import aws_db_instance.postgres task-manager-production-db

# Import S3 bucket
terraform import aws_s3_bucket.app task-manager-production-assets

# Import Security Group
terraform import aws_security_group.app sg-0abc123def456789

# Import VPC
terraform import aws_vpc.main vpc-0abc123def456789
```

### Step-by-Step Import Workflow

```
Step 1: Tulis resource block yang KOSONG (placeholder)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

```hcl
# main.tf — tulis resource placeholder
resource "aws_instance" "app" {
  # kosong dulu — akan diisi setelah import
}
```

```
Step 2: Import
━━━━━━━━━━━━━
```

```bash
terraform import aws_instance.app i-0abc123def456789
# Import successful!
```

```
Step 3: Lihat state
━━━━━━━━━━━━━━━━━━
```

```bash
terraform state show aws_instance.app
# ami                          = "ami-0abcdef1234567890"
# instance_type                = "t3.small"
# key_name                     = "my-key"
# subnet_id                    = "subnet-0abc123"
# vpc_security_group_ids       = ["sg-0abc123"]
# tags                         = {
#     "Name" = "task-manager-app"
# }
```

```
Step 4: Isi resource block berdasarkan state output
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

```hcl
resource "aws_instance" "app" {
  ami                    = "ami-0abcdef1234567890"
  instance_type          = "t3.small"
  key_name               = "my-key"
  subnet_id              = "subnet-0abc123"
  vpc_security_group_ids = ["sg-0abc123"]

  tags = {
    Name = "task-manager-app"
  }
}
```

```
Step 5: Verify — terraform plan harus "No changes"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

```bash
terraform plan
# No changes. Your infrastructure matches the configuration.
```

### Import Block (Terraform 1.5+)

Terraform 1.5 ke atas punya cara import yang lebih clean:

```hcl
# import.tf — declarative import

import {
  to = aws_instance.app
  id = "i-0abc123def456789"
}

import {
  to = aws_s3_bucket.app
  id = "task-manager-production-assets"
}

import {
  to = aws_security_group.app
  id = "sg-0abc123def456789"
}
```

```bash
# Generate config dari imported resources
terraform plan -generate-config-out=generated.tf

# Review generated.tf — Terraform auto-generate resource configuration!
# Lalu pindahkan ke file yang tepat dan cleanup
```

---

## 17.18 Lifecycle & Meta-Arguments

### depends_on — Explicit Dependencies

Biasanya Terraform otomatis detect dependency. Tapi kadang perlu explicit:

```hcl
resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  # Explicit: tunggu RDS dan ElastiCache ready dulu
  depends_on = [
    aws_db_instance.postgres,
    aws_elasticache_cluster.redis,
  ]
}
```

### count — Create Multiple Resources

```hcl
# Bikin 3 EC2 instances
resource "aws_instance" "worker" {
  count = 3

  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "worker-${count.index + 1}"
    # worker-1, worker-2, worker-3
  }
}

# Reference: aws_instance.worker[0], aws_instance.worker[1], aws_instance.worker[2]

# Conditional creation
resource "aws_instance" "bastion" {
  count = var.environment == "production" ? 1 : 0
  # production → bikin 1. staging/dev → ga bikin (0)

  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "bastion-host"
  }
}
```

### for_each — Create Resources from Map/Set

```hcl
# for_each lebih bagus dari count karena kamu punya KEY sebagai identifier

variable "instances" {
  type = map(object({
    instance_type = string
    subnet_id     = string
  }))
  default = {
    web = {
      instance_type = "t3.small"
      subnet_id     = "subnet-abc123"
    }
    api = {
      instance_type = "t3.micro"
      subnet_id     = "subnet-def456"
    }
    worker = {
      instance_type = "t3.medium"
      subnet_id     = "subnet-ghi789"
    }
  }
}

resource "aws_instance" "app" {
  for_each = var.instances

  ami           = data.aws_ami.ubuntu.id
  instance_type = each.value.instance_type
  subnet_id     = each.value.subnet_id

  tags = {
    Name = "${var.project_name}-${each.key}"
    # task-manager-web, task-manager-api, task-manager-worker
  }
}

# Reference: aws_instance.app["web"], aws_instance.app["api"], aws_instance.app["worker"]

# for_each with set of strings
resource "aws_iam_user" "developers" {
  for_each = toset(["alice", "bob", "charlie"])

  name = each.value
}
```

### count vs for_each

```
count:
━━━━━
- Pakai INDEX (0, 1, 2, ...)
- Kalau delete item di tengah → semua index setelahnya bergeser → recreate!
- Bagus untuk: conditional creation (count = 0 or 1)

for_each:
━━━━━━━━
- Pakai KEY (string identifier)
- Delete item → cuma item itu yang dihapus. Yang lain ga terpengaruh.
- Bagus untuk: multiple resources dari map/set

RULE OF THUMB:
- Pakai count untuk conditional (count = var.enabled ? 1 : 0)
- Pakai for_each untuk lists of resources
```

### lifecycle — Control Resource Lifecycle

```hcl
resource "aws_db_instance" "postgres" {
  identifier     = "task-manager-db"
  engine         = "postgres"
  instance_class = "db.t3.micro"
  # ...

  lifecycle {
    # 1. create_before_destroy
    # Bikin yang baru DULU, baru hapus yang lama
    # Berguna untuk zero-downtime updates
    create_before_destroy = true

    # 2. prevent_destroy
    # JANGAN BISA di-destroy! Terraform error kalau coba destroy.
    # Berguna untuk database, S3 bucket berisi data penting
    prevent_destroy = true

    # 3. ignore_changes
    # Abaikan perubahan pada attribute tertentu
    # Berguna kalau attribute di-manage di luar Terraform
    ignore_changes = [
      tags,                    # tags diubah manual? ga apa-apa
      engine_version,          # auto-update oleh AWS? abaikan
    ]
  }
}

resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  lifecycle {
    create_before_destroy = true
    # Update instance type? Bikin instance baru dulu,
    # pindah traffic, baru hapus yang lama
  }
}
```

### Dynamic Blocks

```hcl
# Dynamic block = looping di dalam resource block

variable "ingress_rules" {
  type = list(object({
    port        = number
    protocol    = string
    cidr_blocks = list(string)
    description = string
  }))
  default = [
    {
      port        = 80
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "HTTP"
    },
    {
      port        = 443
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "HTTPS"
    },
    {
      port        = 8080
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "App port"
    },
    {
      port        = 22
      protocol    = "tcp"
      cidr_blocks = ["10.0.0.0/8"]
      description = "SSH from VPC"
    },
  ]
}

resource "aws_security_group" "app" {
  name   = "app-sg"
  vpc_id = var.vpc_id

  # Tanpa dynamic block: harus nulis 4 ingress block manual
  # Dengan dynamic block: loop!

  dynamic "ingress" {
    for_each = var.ingress_rules

    content {
      from_port   = ingress.value.port
      to_port     = ingress.value.port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
      description = ingress.value.description
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

---

## 17.19 Terraform Functions — Built-in Functions

Terraform punya banyak built-in functions. Ini yang paling sering dipakai:

### String Functions

```hcl
# format — string formatting
locals {
  server_name = format("%s-%s-server-%02d", var.project_name, var.environment, 1)
  # "task-manager-production-server-01"
}

# join — join list jadi string
locals {
  subnet_list = join(", ", var.public_subnet_cidrs)
  # "10.0.1.0/24, 10.0.2.0/24"
}

# split — split string jadi list
locals {
  db_host = split(":", var.db_endpoint)[0]
  # "task-manager-db.abc123.ap-southeast-1.rds.amazonaws.com"
}

# upper, lower, title
locals {
  env_upper = upper(var.environment)   # "PRODUCTION"
  env_lower = lower(var.environment)   # "production"
}

# replace
locals {
  clean_name = replace(var.project_name, "_", "-")
  # "task-manager" (kalau input "task_manager")
}

# trimspace, trimprefix, trimsuffix
locals {
  clean = trimspace("  hello  ")           # "hello"
  no_prefix = trimprefix("vpc-main", "vpc-")  # "main"
}

# substr
locals {
  short_hash = substr(var.commit_hash, 0, 7)  # "abc1234"
}
```

### Collection Functions

```hcl
# length — hitung jumlah element
locals {
  subnet_count = length(var.public_subnet_cidrs)  # 2
}

# element — ambil element by index
locals {
  first_az = element(data.aws_availability_zones.available.names, 0)
  # "ap-southeast-1a"
}

# lookup — cari value di map (with default)
locals {
  instance_type = lookup(var.instance_types, var.environment, "t3.micro")
  # Kalau var.environment ga ada di map → return "t3.micro"
}

# merge — gabung maps
locals {
  all_tags = merge(
    var.common_tags,
    {
      Name        = "my-resource"
      Environment = var.environment
    }
  )
}

# concat — gabung lists
locals {
  all_subnets = concat(var.public_subnet_cidrs, var.private_subnet_cidrs)
  # ["10.0.1.0/24", "10.0.2.0/24", "10.0.10.0/24", "10.0.20.0/24"]
}

# flatten — flatten nested lists
locals {
  all_ips = flatten([
    var.office_ips,
    var.vpn_ips,
    var.developer_ips,
  ])
}

# contains — cek apakah list contains value
locals {
  is_prod = contains(["production", "prod"], var.environment)  # true/false
}

# distinct — hapus duplicates
locals {
  unique_ports = distinct([80, 443, 80, 8080, 443])
  # [80, 443, 8080]
}

# keys, values — ambil keys/values dari map
locals {
  env_names = keys(var.instance_types)     # ["dev", "staging", "production"]
  env_types = values(var.instance_types)   # ["t3.micro", "t3.small", "t3.medium"]
}

# zipmap — bikin map dari 2 lists
locals {
  server_ips = zipmap(
    ["web", "api", "worker"],
    ["10.0.1.10", "10.0.1.11", "10.0.1.12"]
  )
  # { web = "10.0.1.10", api = "10.0.1.11", worker = "10.0.1.12" }
}
```

### Numeric Functions

```hcl
# min, max
locals {
  min_size = min(var.desired_count, 1)
  max_size = max(var.desired_count, 10)
}

# ceil, floor
locals {
  workers = ceil(var.total_tasks / 10)  # round up
}

# abs
locals {
  diff = abs(var.a - var.b)
}
```

### Network Functions

```hcl
# cidrsubnet — calculate subnet CIDRs
locals {
  # VPC CIDR: 10.0.0.0/16
  # cidrsubnet(prefix, newbits, netnum)
  subnet_1 = cidrsubnet("10.0.0.0/16", 8, 1)   # "10.0.1.0/24"
  subnet_2 = cidrsubnet("10.0.0.0/16", 8, 2)   # "10.0.2.0/24"
  subnet_10 = cidrsubnet("10.0.0.0/16", 8, 10)  # "10.0.10.0/24"
}

# Practical: auto-generate subnet CIDRs
resource "aws_subnet" "public" {
  count      = 3
  cidr_block = cidrsubnet(var.vpc_cidr, 8, count.index + 1)
  # 10.0.1.0/24, 10.0.2.0/24, 10.0.3.0/24
}
```

### File Functions

```hcl
# file — baca file content
locals {
  user_data_script = file("${path.module}/scripts/user_data.sh")
}

# templatefile — baca file dan replace variables
locals {
  user_data = templatefile("${path.module}/scripts/user_data.sh", {
    db_host     = var.db_host
    db_password = var.db_password
    app_port    = var.app_port
  })
}

# fileexists — cek file ada ga
locals {
  has_custom_config = fileexists("${path.module}/custom.conf")
}

# basename, dirname
locals {
  filename = basename("/path/to/file.txt")  # "file.txt"
  dir      = dirname("/path/to/file.txt")   # "/path/to"
}
```

### Type Conversion Functions

```hcl
# tostring, tonumber, tobool
locals {
  port_string = tostring(var.app_port)     # "8080"
  port_number = tonumber("8080")            # 8080
  enabled     = tobool("true")              # true
}

# tolist, toset, tomap
locals {
  unique_cidrs = toset(var.cidr_blocks)
}

# jsonencode, jsondecode
locals {
  policy_json = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "s3:GetObject"
        Resource = "*"
      }
    ]
  })
}

# yamlencode
locals {
  config_yaml = yamlencode({
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      name = "app-config"
    }
  })
}
```

### Conditional & Loop Expressions

```hcl
# Conditional (ternary)
locals {
  instance_type = var.environment == "production" ? "t3.medium" : "t3.micro"
}

# For expression — transform lists/maps
locals {
  # List comprehension
  upper_names = [for name in var.names : upper(name)]
  # ["ALICE", "BOB", "CHARLIE"]

  # Filter
  production_instances = [
    for k, v in var.instances : v
    if v.environment == "production"
  ]

  # Map comprehension
  instance_map = {
    for name in var.names : name => {
      role = "developer"
      team = "backend"
    }
  }

  # Transform list to map
  subnet_by_az = {
    for s in aws_subnet.public :
    s.availability_zone => s.id
  }
}
```

---

## 17.20 Security Best Practices

### 1. Jangan Commit State File

```gitignore
# .gitignore
*.tfstate
*.tfstate.*
```

```
KENAPA:
State file berisi SEMUA data resource, termasuk:
- Database password (PLAINTEXT!)
- API keys
- Private IP addresses
- Resource ARNs

Commit ke Git = expose secrets ke siapa aja yang punya akses repo.
SELALU pakai remote backend (S3).
```

### 2. Encrypt State at Rest

```hcl
# S3 backend dengan encryption
backend "s3" {
  bucket         = "my-terraform-state"
  key            = "terraform.tfstate"
  region         = "ap-southeast-1"
  encrypt        = true                    # SSE-S3 encryption
  dynamodb_table = "terraform-locks"       # state locking
}

# S3 bucket pakai KMS encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.terraform.arn
    }
  }
}
```

### 3. Pakai IAM Roles, Bukan Access Keys

```hcl
# BURUK — hardcode credentials
provider "aws" {
  region     = "ap-southeast-1"
  access_key = "AKIAIOSFODNN7EXAMPLE"        # JANGAN!
  secret_key = "wJalrXUtnFEMI/K7MDENG/bPxR"  # JANGAN!
}

# BAIK — pakai IAM role atau env vars
provider "aws" {
  region = "ap-southeast-1"
  # Credentials dari: IAM role / env vars / shared credentials file
}
```

### 4. Least Privilege IAM Policy

```hcl
# BURUK — terlalu broad
resource "aws_iam_role_policy" "too_broad" {
  policy = jsonencode({
    Statement = [{
      Effect   = "Allow"
      Action   = "*"            # JANGAN! Ini admin access
      Resource = "*"
    }]
  })
}

# BAIK — specific actions dan resources
resource "aws_iam_role_policy" "specific" {
  policy = jsonencode({
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:GetObject",
        "s3:PutObject",
      ]
      Resource = [
        "arn:aws:s3:::my-bucket/*"  # specific bucket aja
      ]
    }]
  })
}
```

### 5. Sensitive Variables dan Outputs

```hcl
# Mark sensitive variables
variable "db_password" {
  type      = string
  sensitive = true  # ga muncul di plan/apply output
}

# Mark sensitive outputs
output "db_endpoint" {
  value     = aws_db_instance.postgres.endpoint
  sensitive = true  # ga muncul di terminal
}
```

### 6. Security Scanning — tfsec/checkov

```bash
# Install tfsec
brew install tfsec

# Scan Terraform code
tfsec .

# Output:
# Result 1:
#   [AWS018] Resource 'aws_security_group.app' defines a fully open ingress
#   security group rule.
#
#   Impact:     The port is accessible from the internet
#   Resolution: Set a more restrictive cidr range

# Install checkov
pip install checkov

# Scan
checkov -d .

# Output:
# Passed checks: 15, Failed checks: 3, Skipped checks: 0
# Check: CKV_AWS_79: Ensure Instance Metadata Service Version 1 is not enabled
# FAILED for resource: aws_instance.app
```

### 7. Pre-commit Hooks

```yaml
# .pre-commit-config.yaml

repos:
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.96.0
    hooks:
      - id: terraform_fmt          # auto-format
      - id: terraform_validate     # validate
      - id: terraform_tflint       # linting
      - id: terraform_tfsec        # security scan
      - id: terraform_docs         # auto-generate docs
```

```bash
# Install pre-commit
pip install pre-commit
pre-commit install

# Sekarang setiap git commit → otomatis check!
```

### Security Checklist

```
Security Best Practices:
━━━━━━━━━━━━━━━━━━━━━━
[ ] State file di remote backend (S3) — BUKAN di Git
[ ] State encrypted at rest (S3 encryption + KMS)
[ ] State locking enabled (DynamoDB)
[ ] S3 bucket public access blocked
[ ] Pakai IAM roles, bukan access keys
[ ] Sensitive variables marked sensitive = true
[ ] Sensitive outputs marked sensitive = true
[ ] No hardcoded credentials di .tf files
[ ] .gitignore includes *.tfstate, *.tfvars (yang berisi secrets)
[ ] Least privilege IAM policies
[ ] Security scanning (tfsec/checkov) di CI/CD
[ ] Pre-commit hooks untuk format + validate
[ ] Provider version pinned (~> x.0)
[ ] Terraform version pinned (required_version)
```

---

## 17.21 Common Issues & Debugging

### 1. State Lock — "Error acquiring the state lock"

```
Error: Error acquiring the state lock
Lock Info:
  ID:        12345-abcde-67890
  Path:      task-manager/terraform.tfstate
  Operation: OperationTypeApply
  Who:       developer-a@laptop
  Created:   2026-05-01 10:00:00
```

```bash
# Penyebab:
# - Terraform apply lagi jalan oleh developer lain
# - Terraform apply sebelumnya crash/timeout, lock ga ke-release

# Solusi 1: Tunggu yang lagi apply selesai

# Solusi 2: Force unlock (HATI-HATI — pastikan ga ada yang lagi apply!)
terraform force-unlock 12345-abcde-67890
```

### 2. Drift Detection — "Resource has been changed outside Terraform"

```bash
# Penyebab: ada orang yang ubah resource langsung di Console, bukan via Terraform

# Detect drift
terraform plan
# ~ aws_instance.app will be updated in-place
#   ~ instance_type: "t3.micro" → "t3.small"
#   → artinya seseorang ubah instance type di Console

# Solusi 1: Sync state ke actual (keep manual change)
terraform apply -refresh-only
# "Update state to match actual"

# Solusi 2: Revert ke Terraform config (override manual change)
terraform apply
# "Change instance_type back to t3.micro (sesuai .tf file)"
```

### 3. Dependency Cycle — "Cycle detected"

```
Error: Cycle: aws_security_group.a, aws_security_group.b
```

```hcl
# Penyebab: A depends on B, tapi B juga depends on A

# BURUK — cycle!
resource "aws_security_group" "a" {
  ingress {
    security_groups = [aws_security_group.b.id]
  }
}
resource "aws_security_group" "b" {
  ingress {
    security_groups = [aws_security_group.a.id]
  }
}

# BAIK — pisah rule ke aws_security_group_rule
resource "aws_security_group" "a" {
  name = "sg-a"
}
resource "aws_security_group" "b" {
  name = "sg-b"
}
resource "aws_security_group_rule" "a_to_b" {
  type                     = "ingress"
  security_group_id        = aws_security_group.a.id
  source_security_group_id = aws_security_group.b.id
  from_port                = 8080
  to_port                  = 8080
  protocol                 = "tcp"
}
resource "aws_security_group_rule" "b_to_a" {
  type                     = "ingress"
  security_group_id        = aws_security_group.b.id
  source_security_group_id = aws_security_group.a.id
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
}
```

### 4. Provider Version Conflicts

```
Error: Failed to query available provider packages
Could not retrieve the list of available versions for provider hashicorp/aws
```

```bash
# Penyebab: .terraform.lock.hcl outdated atau conflict

# Solusi: delete lock file dan re-init
rm .terraform.lock.hcl
terraform init -upgrade
```

### 5. TF_LOG — Debug Mode

```bash
# Set log level
export TF_LOG=DEBUG    # TRACE, DEBUG, INFO, WARN, ERROR
export TF_LOG_PATH=terraform.log

terraform plan
# Full debug log di terraform.log

# Unset setelah selesai debug
unset TF_LOG
unset TF_LOG_PATH
```

### 6. Common Errors & Solutions

```
Error: "Error creating X: EntityAlreadyExists"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ Resource udah ada di AWS tapi ga ada di state
→ Solusi: terraform import

Error: "Error deleting X: DependencyViolation"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ Resource ga bisa dihapus karena masih ada yang depend
→ Solusi: hapus dependent resource dulu, atau pakai depends_on

Error: "Invalid count argument"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ count ga bisa reference resource yang belum dibuat
→ Solusi: pakai variable atau data source

Error: "Unsupported attribute"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ Attribute ga ada di resource type tersebut
→ Solusi: cek docs provider, mungkin typo atau versi beda

Error: "No matching AMI found"
━━━━━━━━━━━━━━━━━━━━━━━━━━━
→ Filter di data source ga match AMI manapun
→ Solusi: cek filter name, owners, dan region
```

### Debugging Flowchart

```
Terraform Error?
     │
     ├─ "State lock" ─→ Tunggu / force-unlock
     │
     ├─ "EntityAlreadyExists" ─→ terraform import
     │
     ├─ "DependencyViolation" ─→ Check depends_on
     │
     ├─ "Cycle detected" ─→ Break cycle (pisah rule)
     │
     ├─ "Provider version" ─→ terraform init -upgrade
     │
     ├─ "No changes" tapi harusnya ada ─→ terraform refresh
     │
     ├─ Ga jelas ─→ TF_LOG=DEBUG terraform plan
     │
     └─ Masih ga jelas ─→ terraform state show <resource>
                           → compare state vs actual vs config
```

---

## 17.22 Terraform vs Others — Cheat Sheet

### Quick Comparison

| | Terraform | CloudFormation | Pulumi | Ansible | CDK |
|---|-----------|---------------|--------|---------|-----|
| **Provisioning** | Ya | Ya | Ya | Limited | Ya |
| **Config Mgmt** | Limited | Tidak | Limited | Ya | Tidak |
| **Multi-cloud** | Ya | Tidak (AWS only) | Ya | Ya | AWS-focused |
| **Language** | HCL | JSON/YAML | Any PL | YAML | TypeScript/Python |
| **State** | File (local/remote) | AWS-managed | File | Stateless | AWS-managed |
| **Plan/Preview** | Ya (plan) | Ya (changeset) | Ya (preview) | Ya (--check) | Ya (diff) |
| **Community** | Huge | AWS only | Growing | Huge | Growing |
| **Learning Curve** | Medium | Medium | Low* | Low | Medium |
| **Drift Detection** | Plan | Drift detection | Refresh | Limited | Drift |
| **Best For** | Multi-cloud IaC | AWS-only shops | Dev-friendly IaC | Server config | AWS dev teams |

*Pulumi low learning curve kalau udah bisa programming language-nya

### When to Use What

```
Terraform: Default choice untuk IaC. Multi-cloud, mature, huge community.
           → Pilihan terbaik untuk kebanyakan use case.

CloudFormation: Kalau perusahaan 100% AWS dan ga mau manage state file.
                → Gratis, no state management overhead.

Pulumi: Kalau tim developer ga mau belajar HCL dan prefer Python/TypeScript.
        → "IaC tapi pakai bahasa programming biasa."

Ansible: Kalau butuh configure EXISTING servers (install packages, copy files).
         → Complementary sama Terraform, bukan replacement.

CDK: Kalau perusahaan AWS-focused dan tim comfortable sama TypeScript/Python.
     → Compiles ke CloudFormation under the hood.
```

---

## 17.23 CLI Cheat Sheet

### Initialization & Setup

```bash
# Initialize Terraform (download providers, setup backend)
terraform init

# Re-initialize (upgrade providers)
terraform init -upgrade

# Initialize tanpa backend (untuk validate/format aja)
terraform init -backend=false
```

### Planning & Applying

```bash
# Preview changes
terraform plan

# Preview with specific var file
terraform plan -var-file="production.tfvars"

# Preview specific resource
terraform plan -target=aws_instance.app

# Save plan to file
terraform plan -out=tfplan

# Apply changes
terraform apply

# Apply saved plan (skip re-plan)
terraform apply tfplan

# Apply with auto-approve (no prompt)
terraform apply -auto-approve

# Apply specific variable
terraform apply -var="environment=production"

# Apply specific resource only
terraform apply -target=aws_instance.app
```

### Destroying

```bash
# Destroy ALL resources
terraform destroy

# Destroy with auto-approve
terraform destroy -auto-approve

# Destroy specific resource
terraform destroy -target=aws_instance.app

# Plan destroy (preview apa yang bakal dihapus)
terraform plan -destroy
```

### State Management

```bash
# List all resources in state
terraform state list

# Show specific resource
terraform state show aws_instance.app

# Move/rename resource in state
terraform state mv aws_instance.old aws_instance.new

# Remove resource from state (ga delete di cloud)
terraform state rm aws_instance.app

# Pull remote state to local
terraform state pull

# Push local state to remote
terraform state push

# Replace provider in state
terraform state replace-provider hashicorp/aws registry.terraform.io/hashicorp/aws
```

### Workspace

```bash
# List workspaces
terraform workspace list

# Show current workspace
terraform workspace show

# Create new workspace
terraform workspace new dev

# Switch workspace
terraform workspace select production

# Delete workspace
terraform workspace delete staging
```

### Import

```bash
# Import existing resource
terraform import aws_instance.app i-0abc123def456789

# Import with specific state file
terraform import -state=terraform.tfstate aws_instance.app i-0abc123

# Generate config from import (Terraform 1.5+)
terraform plan -generate-config-out=generated.tf
```

### Format & Validate

```bash
# Format all .tf files
terraform fmt

# Format recursively
terraform fmt -recursive

# Check format (CI/CD — return error kalau not formatted)
terraform fmt -check

# Validate configuration
terraform validate
```

### Output & Info

```bash
# Show all outputs
terraform output

# Show specific output
terraform output public_ip

# Show output without quotes
terraform output -raw public_ip

# Show outputs as JSON
terraform output -json

# Show current state (all resources)
terraform show

# Show specific plan file
terraform show tfplan

# Show Terraform version
terraform version

# Show providers used
terraform providers
```

### Advanced

```bash
# Refresh state (sync with actual cloud resources)
terraform refresh
# Atau lebih aman:
terraform apply -refresh-only

# Taint resource (force recreate on next apply)
terraform taint aws_instance.app
# Deprecated — pakai ini:
terraform apply -replace="aws_instance.app"

# Graph (generate dependency graph in DOT format)
terraform graph | dot -Tpng > graph.png

# Console (interactive expression evaluator)
terraform console
> var.project_name
"task-manager"
> length(var.public_subnet_cidrs)
2
> cidrsubnet("10.0.0.0/16", 8, 1)
"10.0.1.0/24"

# Force unlock state
terraform force-unlock LOCK_ID

# Login to Terraform Cloud
terraform login
```

---

## 17.24 Checklist

### Setup & Basics

- [ ] Terraform installed (`terraform version`)
- [ ] VS Code extension installed (HashiCorp Terraform)
- [ ] AWS CLI configured (`aws configure`)
- [ ] Paham IaC concept: declarative, version controlled, reproducible
- [ ] Paham kenapa Terraform: cloud-agnostic, HCL, plan before apply

### HCL & Configuration

- [ ] Paham HCL blocks: resource, variable, output, data, locals, provider
- [ ] Paham variable types: string, number, bool, list, map, object
- [ ] Bisa set variables via: default, tfvars, env vars, CLI flags
- [ ] Paham locals vs variables
- [ ] Bisa pakai string interpolation (`"${var.name}"`)
- [ ] Paham conditional expressions (`condition ? true_val : false_val`)
- [ ] Bisa pakai comments (#, //, /* */)

### Core Workflow

- [ ] Bisa `terraform init` (download providers)
- [ ] Bisa `terraform plan` (preview changes)
- [ ] Bisa baca plan output (+, ~, -, -/+)
- [ ] Bisa `terraform apply` (execute changes)
- [ ] Bisa `terraform destroy` (delete resources)
- [ ] Paham flow: Write → Plan → Apply

### Provider & Resources

- [ ] Provider configured (AWS, version pinned)
- [ ] Credentials configured (shared creds / env vars / IAM role)
- [ ] Bisa bikin EC2 instance
- [ ] Bisa bikin Security Group
- [ ] Bisa pakai data sources (AMI, VPC, AZ)
- [ ] Paham resource vs data source

### State

- [ ] Paham apa itu state file dan kenapa penting
- [ ] Remote state configured (S3 backend)
- [ ] State locking enabled (DynamoDB)
- [ ] State file TIDAK di-commit ke Git
- [ ] Bisa `terraform state list/show`
- [ ] Paham `terraform state mv/rm`

### Modules

- [ ] Paham module concept (reusable infrastructure package)
- [ ] Bisa bikin custom module (main.tf, variables.tf, outputs.tf)
- [ ] Bisa pakai module dari Terraform Registry
- [ ] Bisa pass variables ke module
- [ ] Bisa reference module outputs

### Task Manager Infrastructure

- [ ] VPC + subnets (public + private) provisioned
- [ ] Security groups configured (app, db, cache)
- [ ] EC2 instance provisioned
- [ ] RDS PostgreSQL provisioned
- [ ] ElastiCache Redis provisioned
- [ ] ECR repository created
- [ ] S3 bucket created
- [ ] IAM roles and policies created
- [ ] Organized into modules

### Advanced

- [ ] Paham workspace (multi-environment)
- [ ] Paham lifecycle: create_before_destroy, prevent_destroy, ignore_changes
- [ ] Bisa pakai count dan for_each
- [ ] Bisa pakai dynamic blocks
- [ ] Paham terraform import
- [ ] Bisa pakai built-in functions (lookup, merge, join, format, cidrsubnet, templatefile)

### CI/CD

- [ ] GitHub Actions workflow untuk Terraform
- [ ] Plan on PR (post output as comment)
- [ ] Apply on merge to main
- [ ] Secrets configured di GitHub

### Security

- [ ] State file di remote backend (bukan Git)
- [ ] State encrypted at rest
- [ ] No hardcoded credentials
- [ ] Sensitive variables marked `sensitive = true`
- [ ] Least privilege IAM policies
- [ ] .gitignore configured properly
- [ ] Security scanning (tfsec/checkov) setup

### Debugging

- [ ] Paham common errors dan solutions
- [ ] Bisa pakai TF_LOG untuk debug
- [ ] Paham state lock dan force-unlock
- [ ] Paham drift detection dan resolution

---

## Terraform Mastered!

Selamat! Kamu baru aja menyelesaikan **Infrastructure as Code dengan Terraform**. Ini skill yang **game-changer** di karier DevOps kamu. Dari yang tadinya klik-klik di Console, sekarang kamu bisa:

```
Yang bisa kamu lakuin sekarang:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Define infrastructure as CODE (HCL)
2. Version control infrastructure (Git)
3. Review infrastructure changes via PR (terraform plan)
4. Provision ENTIRE AWS stack dari 1 command (terraform apply)
5. Bikin multiple environments yang identik (workspace / tfvars)
6. Organize infrastructure jadi reusable modules
7. Provision EKS cluster with Terraform
8. Automate infrastructure deployment via CI/CD
9. Import existing resources ke Terraform management
10. Secure infrastructure code (encryption, IAM, scanning)
11. Debug Terraform issues (state lock, drift, cycles)

Journey so far:
━━━━━━━━━━━━━━
Bab 1-10:  Backend Developer (Go API)
Bab 11-15: DevOps Core (Docker, CI/CD, AWS, Nginx, Monitoring)
Bab 16:    Container Orchestration (Kubernetes)
Bab 17:    Infrastructure as Code (Terraform) ← KAMU DI SINI

What's next:
━━━━━━━━━━━
Bab 18: Microservices — Distributed System Patterns
Bab 19: Performance — Load Testing & Optimization
Bab 20: Security — Best Practices
```

Dari developer yang bikin infrastructure manual di Console, sekarang kamu bisa **provision, manage, dan version control** seluruh AWS infrastructure dari kode. Setiap perubahan di-review lewat PR, di-preview lewat `terraform plan`, dan di-apply secara konsisten. No more "siapa yang ubah security group kemarin?" — semuanya ada di Git history.

> **Next up: Bab 18 — Microservices.** Kalau Terraform provision infrastructure dan Kubernetes orchestrate containers, Microservices itu tentang **arsitektur aplikasi** — gimana decompose monolith jadi services yang independent, scalable, dan maintainable. Distributed systems, service mesh, event-driven architecture — the next frontier.

---