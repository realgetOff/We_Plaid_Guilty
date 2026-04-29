*This project has been created as part of the 42 curriculum by namichel, pmilner-, lviravon, mforest-.*

---

# ft_transcendence — We Plaid Guilty 🎮

> *"This project was brought to you with hate by pmilner-, mforest-, namichel & lviravon!"*

<p align="center">
  <img src="./readme_img/team.png">
</p>

---

## Description

**We Plaid Guilty** is a real-time multiplayer web application built for the 42 curriculum final project. The project is a **Gartic Phone**-inspired game where players write prompts, draw them, guess what others drew — and laugh at the results.

It also features an **AI Game mode** where the Grok LLM generates the prompts and players draw and vote on each other's creations.

The UI is deliberately styled after **Apple HyperCard (1987)**, giving the whole app a retro Mac aesthetic: pixel-perfect windows, a simulated menu bar with a live clock, and a footer with legal links.

### A word on the journey

Everything started from a **Docker Compose** file — a single `docker-compose.yml` with a Go backend, a PostgreSQL database, and a Vite frontend. Simple, fast, and deployable in minutes. As the project grew, so did the ambitions, and we migrated the entire stack to a production-grade Kubernetes infrastructure on AWS.

We are pushing the project now because one of our team members unfortunately has to leave — but this is not the end. We plan to keep working on it together and implement everything that we didn't have time to finish. The roadmap is real.

### Key Features

- Real-time multiplayer drawing game (classic Gartic Phone flow: write → draw → guess → gallery)
- AI Game mode powered by the Grok API (AI-generated prompts, star-vote panel, ranked gallery)
- WebSocket-based real-time communication (lobby chat, game phases, friend invites)
- Friends system with real-time online status
- Profile system with avatar upload, username color and font customization
- Standard registration and login (email + hashed password) + 42 OAuth login + guest session
- Production-grade infrastructure on AWS with Kubernetes, Vault PKI, mTLS, CI/CD, monitoring and logging
- HyperCard retro design system with 10+ reusable components

---

## Instructions

### Prerequisites

- AWS account with appropriate permissions (EC2, KMS, S3, ECR, Route53)
- [Terraform](https://www.terraform.io/) >= 1.6
- [Ansible](https://www.ansible.com/) >= 2.15
- [Taskfile](https://taskfile.dev/) >= 3.0
- [Docker](https://www.docker.com/) with buildx
- SSH key pair configured in GitHub secrets (`SSH_PRIVATE_KEY`)
- Ansible Vault password file

### Environment Setup

Secrets are managed by HashiCorp Vault — no `.env` file is committed to the repository. Credentials are stored in an encrypted `ansible/secrets.yml` file:

```bash
# Edit secrets (requires Ansible Vault password)
ansible-vault edit ansible/secrets.yml --vault-password-file ~/.vault_pass
```

Required secrets: `db_password`, `db_user`, `db_name`, `app_jwt_secret`, `app_api_key`, `client_id` (42 OAuth), `client_secret` (42 OAuth), `elastic_pass`, `kibana_pass`, `grafana_pass`, `elk_encrypt_key`.

### Deploy via CI/CD (recommended)

```bash
# Tag a release to trigger the full deployment pipeline
task release -- v1.x.x
```

This commits, tags, and pushes — GitHub Actions handles the build and deploy automatically.

### After cloning from vogsphere
```bash
task setup   # Configure github remote (run once)
```

### First-time Infrastructure Setup

```bash
# 1. Create S3 bucket for Terraform state (run once)
task bootstrap

# 2. Deploy full infrastructure
task deploy-ci
```

### Available Tasks

```
task release       # Commit + tag + push (task release -- v1.x.x)
task deploy-ci     # Full infrastructure deployment (CI/CD)
task destroy       # Destroy EC2 instances (keeps KMS, S3, Route53)
task destroy-full  # Destroy everything (Route53 zone preserved)
task tf-infra      # Apply Terraform infra only
task tf-vault      # Apply Terraform vault only
task ansible       # Run Ansible (task ansible -- <tags>)
task kubectl       # Run kubectl on master (task kubectl -- get nodes)
task output        # Show Terraform outputs
task bootstrap     # Create S3 secrets bucket (run once)
```

### Access

| Service     | URL                                      |
|-------------|------------------------------------------|
| Application | `https://play-stupid.games:30443`        |
| Grafana     | `http://WORKER2_IP:30300`               |
| Kibana      | `http://WORKER1_IP:30601`               |
| Vault UI    | `http://MASTER_IP:30821`                |

---

## Team Information

| Member | Formal Role | Responsibilities |
|---|---|---|
| `namichel` | Tech Lead + PM | Defines technical architecture, oversees all infrastructure decisions, manages timelines and blockers. Terraform, Ansible, Vault, CI/CD, Kubernetes, Monitoring. |
| `pmilner-` | PO + Developer | Defines product features and priorities, validates completed work. REST API, WebSocket management, server routing, database communication & queries, code cleanup. |
| `lviravon` | Developer | Game logic (classic & AI), pipeline & handler system, SonarQube integration. |
| `mforest-` | Developer | Frontend React application, HyperCard design system, WebSocket client integration. |

---

## Project Management

- **Task distribution**: each member worked on their own branch per domain (infra, backend API, game logic, frontend). Task assignment was tracked via **SonarQube** — code quality and issues were visible to the whole team after each push.
- **CI/CD as quality gate**: the GitHub Actions security pipeline runs on every push — Checkov, Trivy, Gosec and SonarQube run automatically, keeping the codebase clean across all branches.
- **Communication**: Discord for async exchanges + a lot of in-person discussion since we were mostly working together on-site at 42.
- **Code reviews**: pull requests reviewed by at least one other team member before merge.
- **Version control**: git tags (`v*`) trigger CI/CD — every production deployment is traceable to a tagged commit.

---

## Technical Stack

### Frontend
| Technology | Justification |
|---|---|
| React 18 + Vite | Component-based architecture, fast HMR, large ecosystem |
| Custom CSS (BEM) | Full control over the HyperCard retro aesthetic — no framework could reproduce it |
| WebSocket singleton | Shared real-time connection across all game phases without re-connecting |

### Backend
| Technology | Justification |
|---|---|
| Go + Gin | Performance, strong typing, excellent concurrency for WebSocket rooms |
| pgx v5 | Native PostgreSQL driver with connection pooling, better than `database/sql` for concurrent goroutines |
| PostgreSQL 16 | Relational model fits the data (users, friends, profiles), UUID support via `uuid-ossp` |

### Infrastructure
| Technology | Justification |
|---|---|
| AWS EC2 ARM64 (t4g) | Cost-effective, good performance for a student project |
| K3s + Flannel WireGuard | Lightweight Kubernetes with encrypted pod-to-pod traffic |
| HashiCorp Vault | Centralized secrets management, PKI, auto-unseal via AWS KMS |
| Terraform | Reproducible infrastructure, state stored in S3 |
| Ansible | Idempotent configuration management for all 3 nodes |
| Taskfile | Single entry point for all operations (replaces Makefile) |
| GitHub Actions | CI/CD: security scanning on every push, deploy on tags |

---

## Database Schema

Three tables managed by PostgreSQL with the `uuid-ossp` extension:

```
users
├── id          UUID PRIMARY KEY (generated via uuid-ossp)
├── username    TEXT UNIQUE NOT NULL
├── email       TEXT (required for standard + api42 users)
├── password    TEXT (hashed, required for standard users only)
├── type        ENUM ('standard', 'guest', 'api42')
└── created_at  TIMESTAMP

profiles
├── user_id     UUID REFERENCES users(id) ON DELETE CASCADE
├── display_name TEXT
├── color       TEXT (chat username color)
└── font        TEXT (chat username font style)

friends
├── user_a      UUID REFERENCES users(id) ON DELETE CASCADE
├── user_b      UUID REFERENCES users(id) ON DELETE CASCADE
├── status      ENUM ('pending', 'accepted')
└── PRIMARY KEY (user_a, user_b)
```

**Constraints:**
- Standard users must have email + hashed password
- Guest users have no password
- API42 users must have their 42 email
- Friends table is bidirectional — if A is friends with B, B is friends with A

---

## Features List

| Feature | Member(s) | Description |
|---|---|---|
| Guest login | `pmilner-` | One-click anonymous session via `POST /api/auth/player` |
| 42 OAuth | `pmilner-` | OAuth 2.0 flow with 42 API, JWT returned via query param |
| User profile | `pmilner-`, `mforest-` | Avatar upload (base64), username color, font style |
| Friends system | `pmilner-`, `mforest-` | Add/remove friends, real-time online status via WebSocket |
| Classic game | `lviravon`, `pmilner-` | Full Gartic Phone flow: write → draw → guess → gallery |
| AI game mode | `lviravon` | Grok LLM generates prompts, players draw and vote |
| Drawing canvas | `mforest-` | Pen, eraser, fill, shapes, color picker, undo (30 steps) |
| WebSocket dispatcher | `lviravon`, `pmilner-` | Pipeline-based message routing with pre-execution validation |
| Game room management | `lviravon` | Hub, BaseRoom, Room, AIRoom with struct embedding |
| HyperCard UI | `mforest-` | Retro Mac design system with 10+ reusable components |
| Notification system | `mforest-` | Game invites with 15s auto-dismiss progress bar |
| Infrastructure | `namichel` | AWS, K3s, Terraform, Ansible, Vault, Helm |
| mTLS PostgreSQL | `namichel` | Vault PKI certs + `pg_ident.conf` mapping |
| CI/CD pipeline | `namichel` | GitHub Actions: security scan + build ECR + deploy |
| Monitoring | `namichel` | Prometheus, Grafana, node-exporter, custom dashboards |
| Logging | `namichel` | Elasticsearch, Filebeat, Kibana |
| Security scanning | `namichel`, `lviravon` | Checkov, Trivy, Gosec, SonarQube |
| Privacy Policy + ToS | `mforest-` | Accessible from footer, relevant content |
| Easter egg | `mforest-` | 404 "Find lviravon" Where's Waldo mini-game |

---

## Modules

**Total: 24 points** (14 mandatory + 10 bonus)

### Major Modules (2 pts each)

| Module | Category | Points | Member(s) | Implementation |
|---|---|---|---|---|
| Web-based multiplayer game | Gaming | 2 | `lviravon`, `pmilner-` | Classic Gartic Phone with real-time phases |
| Remote players | Gaming | 2 | `lviravon`, `pmilner-` | WebSocket, reconnection logic, latency handling |
| Multiplayer 3+ players | Gaming | 2 | `lviravon` | Room supports N players with round rotation |
| Real-time features (WebSocket) | Web | 2 | `pmilner-`, `lviravon`, `mforest-` | Singleton client, auth on connect, pub/sub |
| User interaction (chat, profile, friends) | Web | 2 | `pmilner-`, `mforest-` | Lobby chat, profile page, friends list, invites |
| LLM interface (Grok API) | AI | 2 | `lviravon` | AI generates prompts, streaming handled, rate limited |
| WAF + HashiCorp Vault | Cybersecurity | 2 | `namichel` | Vault PKI, mTLS, KMS auto-unseal, nginx ingress |
| ELK log management | DevOps | 2 | `namichel` | Elasticsearch + Filebeat + Kibana on dedicated node |
| Prometheus + Grafana monitoring | DevOps | 2 | `namichel` | kube-prometheus-stack, node-exporter, custom dashboards |
| Custom infrastructure module | Modules of choice | 2 | `namichel` | See justification below |

### Minor Modules (1 pt each)

| Module | Category | Points | Member(s) | Implementation |
|---|---|---|---|---|
| Frontend framework (React) | Web | 1 | `mforest-` | React 18 + Vite |
| Backend framework (Gin) | Web | 1 | `pmilner-` | Gin HTTP framework + pgx |
| Custom design system | Web | 1 | `mforest-` | HyperCard theme, 10+ reusable components |
| OAuth 2.0 (42) | User Management | 1 | `pmilner-` | Full OAuth flow with 42 API |
| CI/CD custom module | Modules of choice | 1 | `namichel` | See justification below |

---

### Custom Module Justifications

#### Major — Infrastructure as Code (Terraform + Ansible)

**Why this module:** The project required deploying a production-grade Kubernetes cluster on AWS with secrets management, mTLS, monitoring and logging. No existing DevOps module covered this full stack.

**Technical challenges:**
- Vault PKI with a single `trans-ca` root CA signing all service certificates
- vault-agent sidecar pattern injecting certs and secrets into every pod
- mTLS between the Go app and PostgreSQL using `pg_ident.conf` to map certificate CN to database username
- iptables DNAT + Flannel WireGuard routing to allow K3s pods to reach Docker-hosted PostgreSQL
- AWS KMS auto-unseal for Vault across destroy/redeploy cycles

**Value added:** The entire infrastructure is reproducible from a single `task release -- v1.x.x`. No manual steps. Every secret is managed by Vault and never touches the filesystem in plaintext.

**Why Major (2 pts):** Terraform (infra + vault), Ansible (6 roles, 3 nodes), Helm (5 charts), custom Go binary (`split-certs`), custom Docker image (`namichel/vault-custom`) — this is a substantial standalone engineering effort equivalent to a full DevOps module.

---

#### Minor — CI/CD Pipeline (GitHub Actions)

**Why this module:** Automated security scanning and deployment was essential for a 4-person team to iterate safely.

**What it does:**
- `security.yml` — Checkov (IaC), Trivy (CVE), Gosec (Go source), SonarQube on every push
- `deploy.yml` — builds ARM64 Docker image, pushes to ECR, triggers Ansible deploy on `v*` tags
- `destroy.yml` — manual infrastructure teardown with environment protection

**Value added:** Every production deployment is gated behind security scans. No image is deployed without passing Checkov, Trivy and Gosec.

**Why Minor (1 pt):** Complements the infrastructure module — focused scope (3 workflows), but meaningful automation.

---

## Individual Contributions

### `namichel` — Tech Lead + PM
- Full AWS infrastructure: EC2, KMS, S3, ECR, Route53, IAM roles (Terraform)
- K3s cluster setup on 3 ARM64 nodes with Flannel WireGuard
- HashiCorp Vault: PKI, policies, Kubernetes auth, AWS auth, KMS auto-unseal
- vault-agent sidecar pattern for all services
- mTLS between app Go and PostgreSQL (`pg_ident.conf`, `pg_hba.conf`)
- iptables DNAT + firewalld to route K3s pods to Docker-hosted PostgreSQL
- Helm charts: gartic-app, ingress-nginx, Vault, Prometheus, Grafana, ELK
- GitHub Actions: security (Checkov, Trivy, Gosec, SonarQube), deploy, destroy
- Taskfile replacing Makefile
- `split-certs` Go binary + `namichel/vault-custom` Docker image
- Grafana dashboards (cluster + app metrics)
- **Challenge**: Routing K3s pods to a Docker-hosted PostgreSQL on the same host required multiple layers of iptables rules, DNAT, and NetworkPolicy tuning over several debug sessions.

### `pmilner-` — PO + Developer
- REST API design and implementation (Gin)
- WebSocket management and authentication flow
- Server routing and middleware
- PostgreSQL database design, connection pooling (pgx), all SQL queries
- User registration, login, session management
- Friends system backend
- Code cleanup and refactoring across the codebase
- **Challenge**: Connection pooling with pgx required careful handling of concurrent goroutines during high-traffic game sessions.

### `lviravon` — Developer
- Classic game logic: round calculation, player shuffle, drawing rotation
- AI game logic: Grok API integration, prompt generation, voting system
- WebSocket dispatcher and pipeline system
- Handler registration and validation pipeline
- Go channels for round transition management (both buffered and unbuffered)
- SonarQube integration and code quality enforcement
- **Challenge**: Designing a pipeline system that keeps handlers clean while enforcing strict validation order required multiple architectural iterations.

### `mforest-` — Developer
- React 18 frontend with Vite
- HyperCard retro design system (10+ reusable components, BEM CSS)
- All game UI: WritePrompt, DrawBoard, GuessPrompt, Gallery, AIVotePanel, AIGallery
- WebSocket singleton client with auth queue
- Friends system UI with real-time status
- Profile page with avatar upload
- MacWindow, Navbar, NotificationBell, ToastContainer components
- Privacy Policy and Terms of Service pages
- 404 Easter egg "Find lviravon" mini-game
- **Challenge**: The DrawBoard canvas with undo (30 steps), flood fill, and multiple shape tools required significant state management without any drawing library.

---

## Architecture Overview

```
Internet
    │
    ▼
Route53 (play-stupid.games)
    │
    ▼
AWS EC2 (eu-north-1, ARM64)
    │
    ├── gartic-master   → K3s control plane + Vault + PostgreSQL (Docker) + App Go
    ├── gartic-elk      → Kibana + Filebeat
    └── gartic-grafana  → Grafana + Prometheus + node-exporter
```

### Security Architecture

**Vault PKI (`trans-ca`)** — Single root CA signs all certificates. Each service has a PKI role with scoped `allowed_domains`. vault-agent sidecars inject and renew certificates automatically.

**mTLS App Go ↔ PostgreSQL** — vault-agent injects a PEM bundle. PostgreSQL verifies client certs via `hostssl ... cert clientcert=verify-full map=appmap`. `pg_ident.conf` maps `gartic-app.app.svc.cluster.local` → DB username.

**HTTPS** — nginx ingress terminates TLS from the browser. nginx communicates with the Go app over HTTPS (port 8080). K8s liveness/readiness probes use HTTP on port 8081.

---

## Roadmap

Even though we have to submit the project now, we plan to keep working on it to implement things we care about:

- **`trivy-to-junit` + `gosec-to-junit`** — Go binaries to convert Trivy and Gosec JSON output into proper JUnit XML for CI reporting
- **Ansible roles with Molecule tests** — proper role testing and idempotency validation
- **cert-manager + Let's Encrypt** — public HTTPS on `play-stupid.games` without browser certificate warnings
- **mTLS everywhere** — Grafana, Kibana, Prometheus all behind mutual TLS, not just the app ↔ database path
- **Inspektor Gadget eBPF** — syscall tracing on the app Go pod and nginx (trace_open, trace_tcp, trace_exec) — we want to practice profiling and runtime security
- **Seccomp / AppArmor / KubeArmor** — tighten the kernel-level security posture of every workload in the cluster
- **Tower Defense + Tron** — two new games we plan to add to the platform
- **Terratest** — infrastructure testing to validate Terraform modules automatically
- **Transition to Hubble** — replace the ELK stack with Hubble for network observability, and migrate metrics to **Loki/Mimir** (ELK was required by the subject, but Loki/Mimir is lighter and more Kubernetes-native)
- **Wolfi-based images** — replace current Docker images with [Wolfi](https://wolfi.dev/) (apko + melange) for minimal, reproducible, CVE-free base images
- **Talos Linux on master** — replace AlmaLinux with Talos Linux on the master node for an immutable, API-driven OS
- **Kepler** — add [Kepler](https://github.com/sustainable-computing-io/kepler) for energy consumption monitoring of Kubernetes workloads

---

## Resources

### Documentation
- [HashiCorp Vault PKI](https://developer.hashicorp.com/vault/docs/secrets/pki)
- [K3s Documentation](https://docs.k3s.io/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Ansible Documentation](https://docs.ansible.com/)
- [pgx PostgreSQL Driver](https://github.com/jackc/pgx)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [Grok API Documentation](https://docs.x.ai/api)
- [Checkov Documentation](https://www.checkov.io/1.Welcome/What%20is%20Checkov.html)
- [Prometheus + Grafana](https://prometheus.io/docs/)
- [Elastic Stack](https://www.elastic.co/guide/index.html)
- [Go Documentation](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Go Language — GeeksforGeeks](https://www.geeksforgeeks.org/go-language)
- [42 API](https://api.intra.42.fr/apidoc/guides/getting_started)
- [SQL Tutorial](https://www.w3schools.com/sql/default.asp)

### AI Usage

AI (Claude by Anthropic) was used during this project mainly as a **thinking partner** — helping us understand what solutions were available for a given problem and choosing the best one, rather than generating code directly.

When facing complex infrastructure or architectural decisions (Vault PKI topology, iptables routing, mTLS configuration, Kubernetes networking), AI helped map out the option space and reason through trade-offs. The actual implementation decisions, debugging, and validation were always done by the team.

AI was also used to help write and structure this README.

---

## Known Limitations

- Grafana and Kibana are served over HTTP (HTTPS integration in roadmap via `split-certs`)
- Mobile devices are blocked intentionally (the drawing canvas is not touch-optimized)
- CSS not fully compatible with Firefox
- 42 OAuth callback URL must be configured on intra.42.fr for each deployment IP
