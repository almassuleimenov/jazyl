# security-reviewer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/security-reviewer (MIT). Рабочая копия скилла: `.claude/skills/security-reviewer/`.

---
name: security-reviewer
description: Identifies security vulnerabilities, generates structured audit reports with severity ratings, and provides actionable remediation guidance. Use when conducting security audits, reviewing code for vulnerabilities, or analyzing infrastructure security. Invoke for SAST scans, penetration testing, DevSecOps practices, cloud security reviews, dependency audits, secrets scanning, or compliance checks. Produces vulnerability reports, prioritized recommendations, and compliance checklists.
license: MIT
allowed-tools: Read Grep Glob
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.1"
  domain: security
  triggers: security review, vulnerability scan, SAST, security audit, penetration test, code audit, security analysis, infrastructure security, DevSecOps, cloud security, compliance audit
  role: specialist
  scope: review
  output-format: report
  related-skills: secure-code-guardian, code-reviewer, devops-engineer, cloud-architect, kubernetes-specialist, api-designer, mcp-developer
---

# Security Reviewer

Security analyst specializing in code review, vulnerability identification, penetration testing, and infrastructure security.

## When to Use This Skill

- Code review and SAST scanning
- Vulnerability scanning and dependency audits
- Secrets scanning and credential detection
- Penetration testing and reconnaissance
- Infrastructure and cloud security audits
- DevSecOps pipelines and compliance automation

## Core Workflow

1. **Scope** — Map attack surface and critical paths. Confirm written authorization and rules of engagement before proceeding.
2. **Scan** — Run SAST, dependency, and secrets tools. Example commands:
   - `semgrep --config=auto .`
   - `bandit -r ./src`
   - `gitleaks detect --source=.`
   - `npm audit --audit-level=moderate`
   - `trivy fs .`
3. **Review** — Manual review of auth, input handling, and crypto. Tools miss context — manual review is mandatory.
4. **Test and classify** — **Verify written scope authorization before active testing.** Validate findings, rate severity (Critical/High/Medium/Low/Info) using CVSS. Confirm exploitability with proof-of-concept only; do not exceed it.
5. **Report** — Confirm findings with stakeholder before finalizing. Document with location, impact, and remediation. Report critical findings immediately.

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| SAST Tools | `references/sast-tools.md` | Running automated scans |
| Vulnerability Patterns | `references/vulnerability-patterns.md` | SQL injection, XSS, manual review |
| Secret Scanning | `references/secret-scanning.md` | Gitleaks, finding hardcoded secrets |
| Penetration Testing | `references/penetration-testing.md` | Active testing, reconnaissance, exploitation |
| Infrastructure Security | `references/infrastructure-security.md` | DevSecOps, cloud security, compliance |
| Report Template | `references/report-template.md` | Writing security report |

## Constraints

### MUST DO
- Check authentication/authorization first
- Run automated tools before manual review
- Provide specific file/line locations
- Include remediation for each finding
- Rate severity consistently
- Check for secrets in code
- Verify scope and authorization before active testing
- Document all testing activities
- Follow rules of engagement
- Report critical findings immediately

### MUST NOT DO
- Skip manual review (tools miss things)
- Test on production systems without authorization
- Ignore "low" severity issues
- Assume frameworks handle everything
- Share detailed exploits publicly
- Exploit beyond proof of concept
- Cause service disruption or data loss
- Test outside defined scope

## Output Templates

1. Executive summary with risk assessment
2. Findings table with severity counts
3. Detailed findings with location, impact, and remediation
4. Prioritized recommendations

### Example Finding Entry

```
ID: FIND-001
Severity: High (CVSS 8.1)
Title: SQL Injection in user search endpoint
File: src/api/users.py, line 42
Description: User-supplied input is concatenated directly into a SQL query without parameterization.
Impact: An attacker can read, modify, or delete database contents.
Remediation: Use parameterized queries or an ORM. Replace `cursor.execute(f"SELECT * FROM users WHERE name='{name}'")`
             with `cursor.execute("SELECT * FROM users WHERE name=%s", (name,))`.
References: CWE-89, OWASP A03:2021
```

## Knowledge Reference

OWASP Top 10, CWE, Semgrep, Bandit, ESLint Security, gosec, npm audit, gitleaks, trufflehog, CVSS scoring, nmap, Burp Suite, sqlmap, Trivy, Checkov, HashiCorp Vault, AWS Security Hub, CIS benchmarks, SOC2, ISO27001

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/security/security-reviewer/)

---

## Reference: infrastructure-security

# Infrastructure Security

## DevSecOps Integration

### CI/CD Security Pipeline

```yaml
# GitHub Actions - Security scanning
name: Security Pipeline
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: returntocorp/semgrep-action@v1
      - uses: gitleaks/gitleaks-action@v2
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'CRITICAL,HIGH'
```

### Infrastructure as Code Security

```bash
# Terraform/CloudFormation scanning
checkov -d terraform/ --framework terraform
tfsec terraform/
terrascan scan -d terraform/

# Kubernetes manifest scanning
kubesec scan deployment.yaml
```

## Cloud Security Controls

### AWS Security Hardening

```bash
# Enable security services
aws guardduty create-detector --enable
aws securityhub enable-security-hub
aws cloudtrail create-trail --name security-trail --s3-bucket-name logs

# Check S3 bucket security
aws s3api list-buckets --query "Buckets[].Name" | \
  xargs -I {} aws s3api get-bucket-acl --bucket {}

# IAM password policy
aws iam update-account-password-policy \
  --minimum-password-length 14 \
  --require-symbols --require-numbers \
  --require-uppercase-characters --require-lowercase-characters
```

### Azure Security

```bash
# Enable Security Center
az security auto-provisioning-setting update --name default --auto-provision on

# Enable disk encryption
az vm encryption enable --resource-group myRG --name myVM --disk-encryption-keyvault myKV
```

### GCP Security

```bash
# Enable Security Command Center
gcloud services enable securitycenter.googleapis.com

# Enable VPC Flow Logs
gcloud compute networks subnets update SUBNET --enable-flow-logs
```

## Container Security

### Secure Dockerfile

```dockerfile
FROM node:18-alpine
RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001
WORKDIR /app
COPY --chown=nodejs:nodejs package*.json ./
RUN npm ci --only=production
USER nodejs
EXPOSE 3000
HEALTHCHECK --interval=30s CMD node healthcheck.js
CMD ["node", "server.js"]
```

### Kubernetes Security

```yaml
# Pod Security Standards
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: app
    image: myapp:1.0
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop: [ALL]
    resources:
      limits:
        memory: "128Mi"
        cpu: "500m"
---
# Network Policy - Default deny
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

## Compliance Automation

### CIS Benchmark Scanning

```bash
# Docker CIS benchmark
docker run --net host --pid host --cap-add audit_control \
  -v /var/lib:/var/lib -v /var/run/docker.sock:/var/run/docker.sock \
  docker/docker-bench-security

# Kubernetes CIS benchmark
kube-bench run --targets master,node

# Linux system hardening
lynis audit system --quick
```

### Compliance as Code (InSpec)

```ruby
# controls/baseline.rb
control 'ssh-hardening' do
  impact 1.0
  title 'SSH Security Configuration'

  describe sshd_config do
    its('Protocol') { should eq '2' }
    its('PermitRootLogin') { should eq 'no' }
    its('PasswordAuthentication') { should eq 'no' }
  end
end

control 'encryption-at-rest' do
  impact 1.0
  title 'S3 Encryption Enabled'

  describe aws_s3_bucket('my-bucket') do
    it { should have_default_encryption_enabled }
  end
end
```

## Secrets Management

### HashiCorp Vault

```bash
# Initialize and configure
vault operator init
vault secrets enable -path=secret kv-v2

# Store secrets
vault kv put secret/app/config api_key="<api-key>"

# Dynamic database credentials
vault secrets enable database
vault write database/config/postgresql \
  plugin_name=postgresql-database-plugin \
  allowed_roles="app" \
  connection_url="postgresql://{{username}}:{{password}}@localhost:5432/" \
  username="vault" password="<db-password>"

vault write database/roles/app \
  db_name=postgresql \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}';" \
  default_ttl="1h" max_ttl="24h"
```

### Kubernetes Secrets with External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      auth:
        kubernetes:
          role: "app-role"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
  target:
    name: app-secrets
  data:
  - secretKey: api_key
    remoteRef:
      key: secret/app/config
      property: api_key
```

## Security Monitoring

### SIEM Log Shipping (Filebeat)

```yaml
filebeat.inputs:
- type: log
  paths:
    - /var/log/auth.log
    - /var/log/nginx/*.log
  fields:
    environment: production

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "security-logs-%{+yyyy.MM.dd}"
```

## Quick Reference

| Area | Tool | Purpose |
|------|------|---------|
| Cloud Security | Prowler, ScoutSuite | AWS/Azure/GCP audit |
| Container | Trivy, Clair | Image scanning |
| IaC | Checkov, tfsec | Terraform/CloudFormation |
| Secrets | Vault, Sealed Secrets | Secret management |
| Compliance | InSpec, OpenSCAP | CIS benchmarks |
| Monitoring | ELK, Splunk | SIEM |

| Framework | Focus | Key Controls |
|-----------|-------|--------------|
| SOC 2 | Security controls | Access, encryption, monitoring |
| ISO 27001 | ISMS | Policy, risk, audit |
| PCI DSS | Payment security | Network segmentation, encryption |
| HIPAA | Healthcare | Encryption, access logs |
| GDPR | Data privacy | Consent, retention, DLP |

---

## Reference: penetration-testing

# Penetration Testing

## Reconnaissance

### Passive Information Gathering

```bash
# DNS enumeration
dig example.com ANY
nslookup -type=any example.com

# Subdomain discovery
subfinder -d example.com
amass enum -d example.com

# Certificate transparency
curl -s "https://crt.sh/?q=%.example.com&output=json"
```

### Active Scanning

```bash
# Port scanning
nmap -sV -p- target.com
nmap -sC -sV -oA scan target.com

# Web technology detection
whatweb target.com
```

## Web Application Testing

### Authentication & Authorization

```bash
# Session analysis - Check for:
# - Session timeout, Secure/HttpOnly flags
# - Session fixation, concurrent sessions

# IDOR testing
GET /api/users/123  # Your ID
GET /api/users/124  # Another user - should fail

# Privilege escalation
GET /api/admin/users  # As standard user
```

### Input Validation

```bash
# SQL injection
sqlmap -u "http://target.com/search?q=test" --batch

# XSS payloads
<script>alert(document.domain)</script>
<img src=x onerror=alert(1)>
<svg onload=alert(1)>

# Command injection
; ls -la
| whoami
$(whoami)

# XXE
<?xml version="1.0"?>
<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
<root>&xxe;</root>
```

## API Security Testing

### JWT & Token Security

```bash
# Decode JWT
echo "<jwt>" | base64 -d

# Test none algorithm
# Modify header: {"alg": "none"}

# Weak secret brute force
hashcat -m 16500 jwt.txt wordlist.txt
```

### Rate Limiting & Data Exposure

```bash
# Test rate limits
for i in {1..1000}; do
  curl https://api.target.com/login -d "user=<username>&pass=<password>"
done

# Check for excessive data exposure
GET /api/users/me
# Look for: password hashes, internal IDs, sensitive PII

# Mass assignment
POST /api/users/profile
{"email": "new@email.com", "isAdmin": true}
```

## Network Penetration

### Privilege Escalation (Linux)

```bash
# SUID binaries
find / -perm -4000 -type f 2>/dev/null

# Sudo permissions
sudo -l

# Writable paths in PATH
echo $PATH | tr ':' '\n' | xargs -I {} ls -ld {}

# Kernel exploits
uname -a
searchsploit linux kernel $(uname -r)
```

### Lateral Movement

```bash
# Network enumeration
arp -a
netstat -ant

# Service discovery
nmap -sV 192.168.1.0/24

# Credential harvesting
grep -r "password" /home/*/
cat ~/.bash_history | grep -i "pass\|pwd\|secret"
```

## Mobile Application Testing

### Android

```bash
# Decompile APK
apktool d app.apk
jadx -d output app.apk

# Check for secrets
grep -r "api_key\|secret\|password" .

# Insecure storage
adb shell
run-as com.app.package
find . -type f -exec cat {} \;
```

### iOS

```bash
# Class dump
class-dump App.app

# Check data storage
sqlite3 /var/mobile/Applications/.../Library/Caches/data.db
```

## Cloud Security Testing

### AWS

```bash
# S3 bucket enumeration
aws s3 ls s3://bucket-name --no-sign-request
aws s3api get-bucket-acl --bucket bucket-name

# IAM enumeration
aws iam get-user
aws iam list-attached-user-policies --user-name username
```

### Container & Kubernetes

```bash
# Docker escape testing
docker inspect container_id | grep -i privileged
docker inspect container_id | grep -A 5 Mounts

# Kubernetes
kubectl get pods --all-namespaces
kubectl get secrets --all-namespaces
kubectl auth can-i --list
```

## Exploitation Validation

### Proof of Concept Guidelines

```python
# Always demonstrate impact SAFELY

# SQL injection PoC
# DON'T: Extract actual data
# DO: Prove injection with sleep
payload = "' OR SLEEP(5)--"

# DON'T: Delete/modify production data
# DO: Show you COULD with SELECT
payload = "' UNION SELECT 'proof_of_concept'--"
```

### Rules of Engagement

1. **Scope verification** - Only test authorized targets
2. **Time windows** - Respect testing hours
3. **DoS prevention** - Avoid resource exhaustion
4. **Data handling** - Don't exfiltrate real data
5. **Stop on discovery** - Don't exploit beyond proof
6. **Immediate reporting** - Report critical findings ASAP
7. **Documentation** - Record all actions
8. **Cleanup** - Remove test artifacts

## Vulnerability Classification

### Severity Scoring

| Severity | Exploitability | Impact | CVSS Range |
|----------|---------------|---------|------------|
| Critical | Easy | Full compromise | 9.0-10.0 |
| High | Medium | Significant access | 7.0-8.9 |
| Medium | Hard | Limited access | 4.0-6.9 |
| Low | Very hard | Minimal impact | 0.1-3.9 |

### Impact Assessment

- **Critical**: Remote code execution, full data access, admin takeover
- **High**: Authentication bypass, privilege escalation, sensitive data exposure
- **Medium**: CSRF, XSS (non-admin), information disclosure
- **Low**: Missing security headers, verbose errors, rate limiting issues

## Testing Checklist

### OWASP Top 10 Coverage

- [ ] Broken Access Control (IDOR, path traversal)
- [ ] Cryptographic Failures (weak encryption, plaintext)
- [ ] Injection (SQL, XSS, command)
- [ ] Insecure Design (missing auth flows)
- [ ] Security Misconfiguration (defaults, debug mode)
- [ ] Vulnerable Components (outdated dependencies)
- [ ] Authentication Failures (weak passwords, session issues)
- [ ] Data Integrity (deserialization, lack of verification)
- [ ] Logging Failures (missing logs, exposed sensitive data)
- [ ] SSRF (unvalidated URLs)

## Quick Reference

| Test Type | Tools | Focus |
|-----------|-------|-------|
| Web App | Burp Suite, OWASP ZAP | OWASP Top 10 |
| API | Postman, curl | AuthN/AuthZ, data exposure |
| Network | nmap, Metasploit | Services, exploits |
| Mobile | MobSF, Frida | Data storage, crypto |
| Cloud | ScoutSuite, Prowler | Misconfigurations |

| Finding Type | Validation Method | Evidence Required |
|--------------|------------------|-------------------|
| SQL Injection | Sleep-based, error-based | Request/response, timing |
| XSS | Alert box, DOM manipulation | Screenshot, payload |
| IDOR | Access other user's resource | Two user accounts, IDs |
| Auth Bypass | Unauthorized access | Before/after screenshots |
| RCE | Command output (safe) | Whoami, id command output |

---

## Reference: report-template

# Security Report Template

## Full Report Template

```markdown
# Security Review Report

## Executive Summary

| Field | Value |
|-------|-------|
| **Application** | [Application Name] |
| **Review Date** | [YYYY-MM-DD] |
| **Reviewer** | [Name] |
| **Scope** | [Files/modules reviewed] |
| **Overall Risk Level** | [Critical/High/Medium/Low] |

### Key Findings
- X Critical vulnerabilities requiring immediate attention
- Y High-severity issues to address before deployment
- Z Medium/Low issues for future consideration

## Findings Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | X | Requires immediate fix |
| High | X | Fix before deployment |
| Medium | X | Fix in next sprint |
| Low | X | Backlog |

## Detailed Findings

### [CRITICAL] SQL Injection in User Search

| Field | Value |
|-------|-------|
| **ID** | SEC-001 |
| **Location** | `src/api/users.ts:45` |
| **CWE** | CWE-89 |
| **CVSS** | 9.8 (Critical) |

**Description**
User input directly concatenated into SQL query without sanitization.

**Vulnerable Code**
```typescript
const query = `SELECT * FROM users WHERE name LIKE '%${searchTerm}%'`;
```

**Proof of Concept**
```
GET /api/users?search=' OR '1'='1
```

**Impact**
- Full database access
- Data exfiltration
- Data modification/deletion
- Potential RCE via SQL features

**Remediation**
Use parameterized queries:
```typescript
const query = 'SELECT * FROM users WHERE name LIKE $1';
db.query(query, [`%${searchTerm}%`]);
```

**Effort**: 1 hour
**Priority**: Immediate

---

### [HIGH] Weak Password Requirements

| Field | Value |
|-------|-------|
| **ID** | SEC-002 |
| **Location** | `src/auth/validation.ts:12` |
| **CWE** | CWE-521 |
| **CVSS** | 7.5 (High) |

**Description**
Password policy requires only 6 characters with no complexity requirements.

**Current Policy**
```typescript
const isValid = password.length >= 6;
```

**Impact**
- Susceptible to brute force attacks
- Dictionary attack vulnerability

**Remediation**
Implement stronger requirements:
```typescript
const isValid =
  password.length >= 12 &&
  /[A-Z]/.test(password) &&
  /[a-z]/.test(password) &&
  /[0-9]/.test(password) &&
  /[^A-Za-z0-9]/.test(password);
```

**Effort**: 30 minutes
**Priority**: Before deployment

## Automated Scan Results

### Dependency Vulnerabilities
| Package | Severity | CVE | Fix |
|---------|----------|-----|-----|
| lodash | High | CVE-2021-xxxx | Upgrade to 4.17.21 |

### SAST Findings
| Tool | Critical | High | Medium | Low |
|------|----------|------|--------|-----|
| Semgrep | 1 | 3 | 5 | 8 |
| npm audit | 0 | 2 | 4 | 10 |

## Recommendations

### Immediate (This Sprint)
1. Fix SQL injection vulnerability (SEC-001)
2. Implement parameterized queries globally
3. Update vulnerable dependencies

### Short-term (Next Sprint)
1. Strengthen password policy (SEC-002)
2. Add input validation middleware
3. Enable security headers

### Long-term
1. Implement SAST in CI/CD pipeline
2. Schedule regular security reviews
3. Security training for developers

## Appendix

### Tools Used
- Semgrep v1.x
- npm audit
- Gitleaks v8.x
- Manual code review

### References
- OWASP Top 10 2021
- CWE Database
- CVSS Calculator
```

## Severity Definitions

| Severity | CVSS Score | Response Time |
|----------|------------|---------------|
| Critical | 9.0 - 10.0 | Immediate |
| High | 7.0 - 8.9 | 24-48 hours |
| Medium | 4.0 - 6.9 | 1-2 weeks |
| Low | 0.1 - 3.9 | Next release |

## Quick Reference

| Section | Purpose |
|---------|---------|
| Executive Summary | Management overview |
| Findings Summary | Quick count by severity |
| Detailed Findings | Technical details |
| Scan Results | Automated tool output |
| Recommendations | Prioritized action items |

---

## Reference: sast-tools

# SAST Tools

## JavaScript/TypeScript

```bash
# Dependency vulnerabilities
npm audit
npm audit --json > npm-audit.json

# ESLint security plugin
npm install eslint-plugin-security --save-dev
npx eslint --ext .js,.ts . --plugin security

# Snyk
npx snyk test
npx snyk code test
```

## Python

```bash
# Bandit - Python SAST
pip install bandit
bandit -r . -f json -o bandit-report.json
bandit -r . -ll  # Only high severity

# Safety - Dependency check
pip install safety
safety check
safety check -r requirements.txt --json > safety-report.json

# Pyup Safety
pip install pyupio-safety
pyupio-safety check
```

## Go

```bash
# GoSec - Go security checker
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
gosec -fmt=json -out=gosec-report.json ./...

# Go vulnerability database
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Multi-Language Tools

```bash
# Semgrep - Universal SAST
pip install semgrep
semgrep --config=auto .
semgrep --config=p/security-audit .
semgrep --config=p/owasp-top-ten .

# Trivy - Comprehensive scanner
brew install trivy
trivy fs .
trivy fs --security-checks vuln,secret,config .

# SonarQube (requires server)
sonar-scanner -Dsonar.projectKey=myproject
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Semgrep
  uses: returntocorp/semgrep-action@v1
  with:
    config: p/security-audit

- name: Run npm audit
  run: npm audit --audit-level=high

- name: Run Trivy
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    severity: 'CRITICAL,HIGH'
```

### GitLab CI

```yaml
security-scan:
  image: returntocorp/semgrep
  script:
    - semgrep --config=auto --json -o semgrep.json .
  artifacts:
    reports:
      sast: semgrep.json
```

## Quick Reference

| Language | Primary Tool | Dependency Check |
|----------|--------------|------------------|
| JavaScript | ESLint + security | npm audit |
| TypeScript | ESLint + security | npm audit |
| Python | Bandit | Safety |
| Go | GoSec | govulncheck |
| Java | SpotBugs | OWASP Dependency-Check |
| Ruby | Brakeman | bundler-audit |

| Tool | Strengths | Best For |
|------|-----------|----------|
| Semgrep | Multi-language, custom rules | General SAST |
| Trivy | Container + code + secrets | Comprehensive |
| Bandit | Python-specific | Python projects |
| GoSec | Go-specific | Go projects |
| npm audit | Built-in, fast | Node.js deps |

---

## Reference: secret-scanning

# Secret Scanning

## Gitleaks

```bash
# Install
brew install gitleaks

# Scan current directory
gitleaks detect --source . --verbose

# Scan with report
gitleaks detect --source . -f json -r gitleaks-report.json

# Scan git history
gitleaks detect --source . --log-opts="--all"

# Use baseline (ignore known)
gitleaks detect --baseline-path .gitleaks-baseline.json
```

## TruffleHog

```bash
# Install
pip install trufflehog

# Scan filesystem
trufflehog filesystem .

# Scan git repo
trufflehog git file://. --since-commit HEAD~100

# Scan with JSON output
trufflehog filesystem . --json > trufflehog-report.json
```

## Manual Grep Patterns

```bash
# Common secret patterns
grep -rn "api_key\|apikey\|api-key" --include="*.{ts,js,py}" .
grep -rn "secret\|password\|passwd" --include="*.{ts,js,py}" .
grep -rn "private_key\|privatekey" --include="*.{ts,js,py}" .
grep -rn "access_token\|accesstoken" --include="*.{ts,js,py}" .

# AWS credentials
grep -rn "AKIA[0-9A-Z]{16}" .
grep -rn "aws_secret_access_key" .

# Base64 encoded (potential secrets)
grep -rn "[A-Za-z0-9+/]{40,}=" .

# JWT tokens
grep -rn "eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\." .
```

## Common Secret Patterns

| Type | Pattern | Example |
|------|---------|---------|
| AWS Access Key | `AKIA[0-9A-Z]{16}` | `<aws-access-key-id>` |
| AWS Secret Key | 40 char base64 | `<aws-secret-access-key>` |
| GitHub Token | `ghp_[A-Za-z0-9]{36}` | `<github-token>` |
| Slack Token | `xox[baprs]-` | `<slack-bot-token>` |
| Stripe Key | `sk_live_[A-Za-z0-9]{24}` | `<stripe-secret-key>` |
| Private Key | `-----BEGIN.*PRIVATE KEY-----` | RSA/EC keys |
| JWT | `eyJ[A-Za-z0-9_-]*\.eyJ` | Encoded tokens |

## Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks
```

## CI/CD Integration

```yaml
# GitHub Actions
- name: Gitleaks
  uses: gitleaks/gitleaks-action@v2
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

# GitLab CI
secret_detection:
  image: zricethezav/gitleaks
  script:
    - gitleaks detect --source . -f sarif -r gl-secret-detection-report.sarif
  artifacts:
    reports:
      secret_detection: gl-secret-detection-report.sarif
```

## Remediation Steps

1. **Rotate immediately** - Consider secret compromised
2. **Remove from history** - Use git filter-branch or BFG
3. **Add to .gitignore** - Prevent future commits
4. **Use env variables** - Move to environment
5. **Use secret manager** - AWS Secrets Manager, Vault

```bash
# Remove from git history (BFG)
bfg --replace-text passwords.txt repo.git

# Or git filter-branch
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch path/to/secret" \
  --prune-empty --tag-name-filter cat -- --all
```

## Quick Reference

| Tool | Best For | Speed |
|------|----------|-------|
| Gitleaks | Git history | Fast |
| TruffleHog | Deep scanning | Medium |
| grep | Quick checks | Fast |
| GitHub Secret Scanning | GitHub repos | Auto |

---

## Reference: vulnerability-patterns

# Vulnerability Patterns

## SQL Injection

```typescript
// VULNERABLE
const query = `SELECT * FROM users WHERE id = ${userId}`;
const query = `SELECT * FROM users WHERE name = '${name}'`;

// SECURE - Parameterized queries
const query = 'SELECT * FROM users WHERE id = $1';
db.query(query, [userId]);

// SECURE - ORM
const user = await User.findOne({ where: { id: userId } });
```

## XSS (Cross-Site Scripting)

```typescript
// VULNERABLE - Direct HTML injection
element.innerHTML = userInput;
document.write(userInput);

// SECURE - Text content
element.textContent = userInput;

// SECURE - Sanitization
import DOMPurify from 'dompurify';
element.innerHTML = DOMPurify.sanitize(userInput);

// SECURE - React (auto-escaped)
return <div>{userInput}</div>;

// VULNERABLE - React dangerouslySetInnerHTML
return <div dangerouslySetInnerHTML={{ __html: userInput }} />;
```

## Path Traversal

```typescript
// VULNERABLE
const file = path.join(uploadDir, req.query.filename);
res.sendFile(file);

// SECURE - Validate and normalize
const filename = path.basename(req.query.filename);
const file = path.resolve(uploadDir, filename);

if (!file.startsWith(path.resolve(uploadDir))) {
  throw new Error('Invalid path');
}
res.sendFile(file);
```

## Command Injection

```typescript
// VULNERABLE
exec(`ls ${userInput}`);
exec('git clone ' + repoUrl);

// SECURE - Use arrays, avoid shell
execFile('ls', [userInput]);
spawn('git', ['clone', repoUrl]);

// SECURE - Validation
const allowedCommands = ['status', 'log'];
if (!allowedCommands.includes(cmd)) throw new Error('Invalid');
```

## IDOR (Insecure Direct Object Reference)

```typescript
// VULNERABLE - No authorization check
app.get('/documents/:id', async (req, res) => {
  const doc = await Document.findById(req.params.id);
  res.json(doc);
});

// SECURE - Verify ownership
app.get('/documents/:id', async (req, res) => {
  const doc = await Document.findOne({
    _id: req.params.id,
    userId: req.user.id  // Ensure user owns document
  });
  if (!doc) return res.status(404).json({ error: 'Not found' });
  res.json(doc);
});
```

## Insecure Deserialization

```typescript
// VULNERABLE - Python pickle
import pickle
data = pickle.loads(user_input)

// SECURE - Use JSON
import json
data = json.loads(user_input)

// VULNERABLE - Node.js
const obj = eval('(' + userInput + ')');

// SECURE
const obj = JSON.parse(userInput);
```

## Sensitive Data Exposure

```typescript
// VULNERABLE - Logging sensitive data
logger.info('User login', { email, password });
console.log('Token:', authToken);

// SECURE - Redact sensitive fields
logger.info('User login', { email, password: '[REDACTED]' });

// VULNERABLE - Error response exposes internals
res.status(500).json({ error: err.stack });

// SECURE - Generic error
res.status(500).json({ error: 'Internal server error' });
```

## Quick Reference

| Vulnerability | Input Vector | Prevention |
|---------------|--------------|------------|
| SQL Injection | Query params | Parameterized queries |
| XSS | User content | Output encoding |
| Path Traversal | File paths | path.basename + validation |
| Command Injection | Shell args | execFile, no shell |
| IDOR | Resource IDs | Authorization checks |
| Deserialization | Serialized data | JSON only |
| Data Exposure | Logs, errors | Redaction, generic errors |

## OWASP Top 10 Mapping

| OWASP | Vulnerabilities |
|-------|-----------------|
| A01 Broken Access Control | IDOR, path traversal |
| A02 Cryptographic Failures | Weak encryption, plaintext |
| A03 Injection | SQL, XSS, command |
| A04 Insecure Design | Missing auth, IDOR |
| A05 Security Misconfiguration | Debug mode, default creds |
| A06 Vulnerable Components | Outdated dependencies |
| A07 Auth Failures | Weak passwords, session issues |
| A08 Data Integrity | Insecure deserialization |
| A09 Logging Failures | Missing logs, exposed data |
| A10 SSRF | Unvalidated URLs |
