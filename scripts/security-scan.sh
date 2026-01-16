#!/bin/bash
# NorthStack Security Scanner
# Runs comprehensive security and vulnerability scans

set -e

echo "=========================================="
echo "  NorthStack Security & Vulnerability Scan"
echo "=========================================="
echo ""

REPORT_DIR="./security-reports"
mkdir -p "$REPORT_DIR"

# ========================================
# 1. Go Security Scanner (gosec)
# ========================================
echo "[1/6] Running Go Security Scanner (gosec)..."
if command -v gosec &> /dev/null; then
    gosec -fmt=json -out="$REPORT_DIR/gosec-report.json" ./... 2>/dev/null || true
    gosec -fmt=html -out="$REPORT_DIR/gosec-report.html" ./... 2>/dev/null || true
    echo "  ✓ gosec report saved to $REPORT_DIR/gosec-report.json"
else
    echo "  ⚠ gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi

# ========================================
# 2. Dependency Vulnerability Check (govulncheck)
# ========================================
echo "[2/6] Running Dependency Vulnerability Check (govulncheck)..."
if command -v govulncheck &> /dev/null; then
    govulncheck ./... > "$REPORT_DIR/govulncheck-report.txt" 2>&1 || true
    echo "  ✓ govulncheck report saved to $REPORT_DIR/govulncheck-report.txt"
else
    echo "  ⚠ govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

# ========================================
# 3. Container Image Scan (trivy)
# ========================================
echo "[3/6] Running Container Image Scan (trivy)..."
if command -v trivy &> /dev/null; then
    # Scan Dockerfile for misconfigurations
    trivy config --format json --output "$REPORT_DIR/trivy-dockerfile.json" ./Dockerfile 2>/dev/null || true
    # Scan file system
    trivy fs --format json --output "$REPORT_DIR/trivy-fs.json" . 2>/dev/null || true
    echo "  ✓ trivy reports saved to $REPORT_DIR/"
else
    echo "  ⚠ trivy not installed. Install from: https://trivy.dev"
fi

# ========================================
# 4. Secret Detection (gitleaks)
# ========================================
echo "[4/6] Running Secret Detection (gitleaks)..."
if command -v gitleaks &> /dev/null; then
    gitleaks detect --source=. --report-format=json --report-path="$REPORT_DIR/gitleaks-report.json" 2>/dev/null || true
    echo "  ✓ gitleaks report saved to $REPORT_DIR/gitleaks-report.json"
else
    echo "  ⚠ gitleaks not installed. Install from: https://github.com/gitleaks/gitleaks"
fi

# ========================================
# 5. Kubernetes Manifest Scan (kubesec)
# ========================================
echo "[5/6] Running Kubernetes Manifest Scan (kubesec)..."
if command -v kubesec &> /dev/null; then
    for file in deploy/k8s/*.yaml; do
        if [ -f "$file" ]; then
            kubesec scan "$file" > "$REPORT_DIR/kubesec-$(basename "$file").json" 2>/dev/null || true
        fi
    done
    echo "  ✓ kubesec reports saved to $REPORT_DIR/"
else
    echo "  ⚠ kubesec not installed. Install from: https://kubesec.io"
fi

# ========================================
# 6. Helm Chart Scan
# ========================================
echo "[6/6] Running Helm Chart Lint..."
if command -v helm &> /dev/null; then
    helm lint deploy/helm/northstack > "$REPORT_DIR/helm-lint.txt" 2>&1 || true
    echo "  ✓ helm lint report saved to $REPORT_DIR/helm-lint.txt"
else
    echo "  ⚠ helm not installed"
fi

# ========================================
# Summary Report
# ========================================
echo ""
echo "=========================================="
echo "  Scan Complete!"
echo "=========================================="
echo ""
echo "Reports saved to: $REPORT_DIR/"
ls -la "$REPORT_DIR/"
echo ""

# Check if any critical issues found
CRITICAL_COUNT=0

if [ -f "$REPORT_DIR/gosec-report.json" ]; then
    GOSEC_ISSUES=$(jq '.Issues | length' "$REPORT_DIR/gosec-report.json" 2>/dev/null || echo "0")
    echo "gosec issues: $GOSEC_ISSUES"
    CRITICAL_COUNT=$((CRITICAL_COUNT + GOSEC_ISSUES))
fi

if [ -f "$REPORT_DIR/gitleaks-report.json" ]; then
    LEAKS=$(jq 'length' "$REPORT_DIR/gitleaks-report.json" 2>/dev/null || echo "0")
    echo "secrets detected: $LEAKS"
    CRITICAL_COUNT=$((CRITICAL_COUNT + LEAKS))
fi

echo ""
if [ "$CRITICAL_COUNT" -gt 0 ]; then
    echo "⚠ Found $CRITICAL_COUNT potential issues. Review reports."
    exit 1
else
    echo "✓ No critical issues found!"
    exit 0
fi
