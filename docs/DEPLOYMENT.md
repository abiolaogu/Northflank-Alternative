# Multi-Cloud Deployment Guide

NorthStack supports deployment across multiple cloud platforms. This guide covers deployment to each supported platform.

---

## Supported Platforms

| Platform | K8s Distribution | Storage | LoadBalancer | Secrets |
|----------|------------------|---------|--------------|---------|
| **Harvester** | RKE2 | Longhorn | MetalLB/Nginx | K8s Secrets |
| **AWS** | EKS | EBS CSI | ALB | Secrets Manager |
| **GCP** | GKE | PD CSI | Gateway API | Secret Manager |
| **Azure** | AKS | Disk CSI | AGIC | Key Vault |
| **OpenStack** | Magnum | Cinder | Octavia | Barbican |

---

## Quick Start

### Option 1: Helm Chart (Recommended)

```bash
# Add Helm repo
helm repo add northstack https://charts.northstack.io
helm repo update

# Install with defaults (Harvester)
helm install northstack northstack/northstack -n northstack --create-namespace

# Install for AWS EKS
helm install northstack northstack/northstack -n northstack --create-namespace \
  --set global.cloudProvider=aws \
  --set aws.enabled=true \
  --set aws.region=us-east-1

# Install for GCP GKE
helm install northstack northstack/northstack -n northstack --create-namespace \
  --set global.cloudProvider=gcp \
  --set gcp.enabled=true \
  --set gcp.project=my-project

# Install for Azure AKS
helm install northstack northstack/northstack -n northstack --create-namespace \
  --set global.cloudProvider=azure \
  --set azure.enabled=true \
  --set azure.subscriptionId=<GUID>

# Install for OpenStack
helm install northstack northstack/northstack -n northstack --create-namespace \
  --set global.cloudProvider=openstack \
  --set openstack.enabled=true
```

### Option 2: Raw Manifests

```bash
# Harvester
kubectl apply -f deploy/harvester/

# AWS EKS
kubectl apply -f deploy/aws/

# GCP GKE
kubectl apply -f deploy/gcp/

# Azure AKS
kubectl apply -f deploy/azure/

# OpenStack
kubectl apply -f deploy/openstack/
```

---

## Platform-Specific Configuration

### Harvester (On-Premise)

**Prerequisites:**
- Harvester 1.1+ cluster
- Longhorn storage provisioned
- Nginx Ingress Controller
- cert-manager

**Features:**
- VM Images for hybrid workloads
- Longhorn replicated storage
- Network policies
- HPA auto-scaling

### AWS EKS

**Prerequisites:**
- EKS cluster 1.27+
- AWS Load Balancer Controller
- EBS CSI Driver
- External Secrets Operator
- Karpenter (optional)

**Features:**
- IRSA (IAM Roles for Service Accounts)
- ALB with WAFv2
- Secrets Manager integration
- Karpenter auto-scaling
- Spot instance support

### GCP GKE

**Prerequisites:**
- GKE cluster (Autopilot or Standard)
- Gateway API enabled
- Workload Identity
- Config Connector (optional)

**Features:**
- Workload Identity Federation
- Gateway API with Cloud Armor
- Secret Manager integration
- Regional PD storage
- Spot VM support

### Azure AKS

**Prerequisites:**
- AKS cluster 1.27+
- Application Gateway Ingress Controller
- Azure Disk CSI
- Secrets Store CSI Driver

**Features:**
- Workload Identity
- Application Gateway with WAF
- Key Vault integration
- Premium SSD storage
- Spot node pools

### OpenStack

**Prerequisites:**
- Magnum Kubernetes cluster
- Cinder CSI
- Octavia Load Balancer
- Barbican (for secrets)

**Features:**
- Cinder block storage
- Octavia LBaaS
- Barbican secrets
- Floating IP support

---

## Security Scan Before Deployment

```bash
# Run security scan
./scripts/security-scan.sh

# Review reports
cat security-reports/gosec-report.json
cat security-reports/govulncheck-report.txt
```

---

## Verify Deployment

```bash
# Check pods
kubectl get pods -n northstack

# Check health
curl https://api.northstack.io/health/ready

# View logs
kubectl logs -n northstack -l app=northstack-api -f
```

---

## Architecture Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                      LOAD BALANCER                             │
│        (ALB / Gateway / AGIC / Octavia / Nginx)               │
└─────────────────────────┬──────────────────────────────────────┘
                          │
┌─────────────────────────▼──────────────────────────────────────┐
│                     KUBERNETES CLUSTER                         │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    northstack namespace                  │  │
│  │                                                          │  │
│  │   ┌──────────┐  ┌──────────┐  ┌──────────┐              │  │
│  │   │ API x3   │  │ API x3   │  │ API x3   │  (HPA)       │  │
│  │   └────┬─────┘  └────┬─────┘  └────┬─────┘              │  │
│  │        │             │             │                     │  │
│  │   ┌────▼─────────────▼─────────────▼────────────────┐   │  │
│  │   │                Internal Services                │   │  │
│  │   │  YugabyteDB  │  NATS  │  Redpanda  │  Dragonfly │   │  │
│  │   └─────────────────────────────────────────────────┘   │  │
│  │                                                          │  │
│  │   ┌────────────────────────────────────────────────┐    │  │
│  │   │              Observability Stack                │    │  │
│  │   │  Prometheus  │  Grafana  │  Loki  │  Jaeger    │    │  │
│  │   └────────────────────────────────────────────────┘    │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    Platform Tools                        │  │
│  │    ArgoCD    │    Backstage    │    Coolify              │  │
│  └─────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

---

## Upgrade

```bash
# Helm upgrade
helm upgrade northstack northstack/northstack -n northstack -f custom-values.yaml

# Rolling restart
kubectl rollout restart deployment/northstack-api -n northstack
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Pods pending | Check node resources, storage provisioner |
| LoadBalancer pending | Verify cloud controller, annotations |
| Secrets not mounting | Check External Secrets Operator, permissions |
| Health checks failing | Verify database connectivity, env vars |
