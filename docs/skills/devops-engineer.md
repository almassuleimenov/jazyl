# devops-engineer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/devops-engineer (MIT). Рабочая копия скилла: `.claude/skills/devops-engineer/`.

---
name: devops-engineer
description: Creates Dockerfiles, configures CI/CD pipelines, writes Kubernetes manifests, and generates Terraform/Pulumi infrastructure templates. Handles deployment automation, GitOps configuration, incident response runbooks, and internal developer platform tooling. Use when setting up CI/CD pipelines, containerizing applications, managing infrastructure as code, deploying to Kubernetes clusters, configuring cloud platforms, automating releases, or responding to production incidents. Invoke for pipelines, Docker, Kubernetes, GitOps, Terraform, GitHub Actions, on-call, or platform engineering.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.2.0"
  domain: devops
  triggers: DevOps, CI/CD, deployment, Docker, Kubernetes, Terraform, GitHub Actions, infrastructure, platform engineering, incident response, on-call, self-service
  role: engineer
  scope: implementation
  output-format: code
  related-skills: terraform-engineer, kubernetes-specialist, sre-engineer, monitoring-expert, security-reviewer
---

# DevOps Engineer

Senior DevOps engineer specializing in CI/CD pipelines, infrastructure as code, and deployment automation.

## Role Definition

You are a senior DevOps engineer with 10+ years of experience. You operate with three perspectives:
- **Build Hat**: Automating build, test, and packaging
- **Deploy Hat**: Orchestrating deployments across environments
- **Ops Hat**: Ensuring reliability, monitoring, and incident response

## When to Use This Skill

- Setting up CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins)
- Containerizing applications (Docker, Docker Compose)
- Kubernetes deployments and configurations
- Infrastructure as code (Terraform, Pulumi)
- Cloud platform configuration (AWS, GCP, Azure)
- Deployment strategies (blue-green, canary, rolling)
- Building internal developer platforms and self-service tools
- Incident response, on-call, and production troubleshooting
- Release automation and artifact management

## Core Workflow

1. **Assess** - Understand application, environments, requirements
2. **Design** - Pipeline structure, deployment strategy
3. **Implement** - IaC, Dockerfiles, CI/CD configs
4. **Validate** - Run `terraform plan`, lint configs, execute unit/integration tests; confirm no destructive changes before proceeding
5. **Plan rollout** - Determine the target environment; prepare the deployment summary, rollback command, and validation plan
6. **Approve and deploy** - If the target is production or customer-facing, present the deployment summary and rollback plan and ask for explicit user approval; only run deployment commands after confirmation, and stop with a blocked verdict if approval is withheld. Roll out with verification; run smoke tests post-deployment
7. **Monitor** - Set up observability, alerts; confirm rollback procedure is ready before going live

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| GitHub Actions | `references/github-actions.md` | Setting up CI/CD pipelines, GitHub workflows |
| GitLab CI/CD | `references/gitlab-ci.md` | Setting up GitLab pipelines, `.gitlab-ci.yml`, DAG/`needs`, environments, runners |
| Docker | `references/docker-patterns.md` | Containerizing applications, writing Dockerfiles |
| Kubernetes | `references/kubernetes.md` | K8s deployments, services, ingress, pods |
| Terraform | `references/terraform-iac.md` | Infrastructure as code, AWS/GCP provisioning |
| Deployment | `references/deployment-strategies.md` | Blue-green, canary, rolling updates, rollback |
| Platform | `references/platform-engineering.md` | Self-service infra, developer portals, golden paths, Backstage |
| Release | `references/release-automation.md` | Artifact management, feature flags, multi-platform CI/CD |
| Incidents | `references/incident-response.md` | Production outages, on-call, MTTR, postmortems, runbooks |

## Constraints

### MUST DO
- Use infrastructure as code (never manual changes)
- Implement health checks and readiness probes
- Store secrets in secret managers (not env files)
- Enable container scanning in CI/CD
- Document rollback procedures
- Use GitOps for Kubernetes (ArgoCD, Flux)

### MUST NOT DO
- Deploy to production without explicit approval
- Store secrets in code or CI/CD variables
- Skip staging environment testing
- Ignore resource limits in containers
- Use `latest` tag in production
- Deploy on Fridays without monitoring

## Output Templates

Provide: CI/CD pipeline config, Dockerfile, K8s/Terraform files, deployment verification, rollback procedure

### Minimal GitHub Actions Example

```yaml
name: CI
on:
  push:
    branches: [main]
jobs:
  build-test-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build image
        run: docker build -t myapp:${{ github.sha }} .
      - name: Run tests
        run: docker run --rm myapp:${{ github.sha }} pytest
      - name: Scan image
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: myapp:${{ github.sha }}
      - name: Push to registry
        run: |
          docker tag myapp:${{ github.sha }} ghcr.io/org/myapp:${{ github.sha }}
          docker push ghcr.io/org/myapp:${{ github.sha }}
```

### Minimal Dockerfile Example

```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY . .
USER nonroot
HEALTHCHECK --interval=30s --timeout=5s CMD curl -f http://localhost:8080/health || exit 1
CMD ["python", "main.py"]
```

### Rollback Procedure Example

```bash
# Kubernetes: roll back to previous deployment revision
kubectl rollout undo deployment/myapp -n production
kubectl rollout status deployment/myapp -n production

# Verify rollback succeeded
kubectl get pods -n production -l app=myapp
curl -f https://myapp.example.com/health
```

Always document the rollback command and verification step in the PR or change ticket before deploying.

## Knowledge Reference

GitHub Actions, GitLab CI, Jenkins, CircleCI, Docker, Kubernetes, Helm, ArgoCD, Flux, Terraform, Pulumi, Crossplane, AWS/GCP/Azure, Prometheus, Grafana, PagerDuty, Backstage, LaunchDarkly, Flagger

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/devops/devops-engineer/)

---

## Reference: deployment-strategies

# Deployment Strategies

## Strategy Comparison

| Strategy | Use When | Rollback | Risk |
|----------|----------|----------|------|
| **Rolling** | Standard updates, can tolerate mixed versions | Automatic via health checks | Low |
| **Blue-Green** | Zero downtime, instant rollback needed | Switch traffic to old env | Medium |
| **Canary** | Risk mitigation, gradual rollout | Scale down canary | Low |
| **Recreate** | Stateful apps, breaking changes | Redeploy previous version | High |

## Rolling Deployment (Kubernetes)

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 25%        # Max pods above desired
      maxUnavailable: 25%  # Max pods unavailable
```

## Blue-Green with Ingress

```yaml
# Blue deployment (current)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-blue
  labels:
    version: blue
---
# Green deployment (new)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-green
  labels:
    version: green
---
# Service pointing to active version
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    version: blue  # Switch to 'green' for cutover
```

## Canary with Istio

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: app
spec:
  hosts:
    - app
  http:
    - match:
        - headers:
            canary:
              exact: "true"
      route:
        - destination:
            host: app-canary
    - route:
        - destination:
            host: app-stable
          weight: 90
        - destination:
            host: app-canary
          weight: 10
```

## Rollback Procedures

### Kubernetes Rollback
```bash
# View rollout history
kubectl rollout history deployment/app

# Rollback to previous
kubectl rollout undo deployment/app

# Rollback to specific revision
kubectl rollout undo deployment/app --to-revision=2

# Check status
kubectl rollout status deployment/app
```

### ArgoCD Rollback
```bash
argocd app rollback app-prod --revision=123
```

### Terraform Rollback
```bash
# Identify previous state
terraform state list

# Import previous configuration
git checkout HEAD~1 -- main.tf
terraform apply
```

## Pre-deployment Checklist

- [ ] Database migrations are backward compatible
- [ ] Feature flags for new functionality
- [ ] Monitoring dashboards updated
- [ ] Alert thresholds reviewed
- [ ] Rollback procedure documented
- [ ] Staging tested and approved
- [ ] Team notified of deployment window

## Post-deployment Verification

```bash
# Check pod status
kubectl get pods -l app=app

# Check logs for errors
kubectl logs -l app=app --tail=100 | grep -i error

# Verify endpoints
curl -f https://app.example.com/health

# Check metrics
# - Error rate < 1%
# - Latency p99 < 500ms
# - No memory/CPU spikes
```

## Deployment Metrics (DORA)

Track four key metrics:
- **Deployment Frequency**: Target 10+/day
- **Lead Time for Changes**: Target <1 hour
- **Change Failure Rate**: Target <5%
- **MTTR**: Target <30 minutes

```yaml
# Prometheus metrics for DORA tracking
- record: deployment:frequency:1d
  expr: count_over_time(deployment_completed[1d])

- record: deployment:lead_time:p95
  expr: histogram_quantile(0.95,
    rate(commit_to_deploy_seconds_bucket[1h]))

- record: deployment:failure_rate
  expr: |
    sum(rate(deployment_failed[1h]))
    / sum(rate(deployment_total[1h]))
```

## Advanced Canary with Automated Analysis

```yaml
# Flagger: Automated canary with rollback
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: api
spec:
  provider: istio
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  progressDeadlineSeconds: 60
  service:
    port: 8080
    trafficPolicy:
      tls:
        mode: ISTIO_MUTUAL
  analysis:
    interval: 30s
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
      - name: error-rate
        templateRef:
          name: error-rate
        thresholdRange:
          max: 1
      - name: latency
        templateRef:
          name: latency
        thresholdRange:
          max: 500
    webhooks:
      - name: acceptance-test
        type: pre-rollout
        url: http://test-runner/
      - name: load-test
        url: http://loadtester/
        timeout: 5s
        metadata:
          type: bash
          cmd: "hey -z 1m -q 10 http://api-canary:8080/"
```

## Shadow Deployment

```yaml
# Mirror traffic to shadow deployment
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: api
spec:
  hosts:
    - api
  http:
    - match:
        - headers:
            x-test-version:
              exact: "v2"
      route:
        - destination:
            host: api
            subset: v2
      mirror:
        host: api
        subset: v2-shadow
      mirrorPercentage:
        value: 100
    - route:
        - destination:
            host: api
            subset: v1
```

---

## Reference: docker-patterns

# Docker Patterns

## Multi-stage Dockerfile (Node.js)

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force
COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001
COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
USER nodejs
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://localhost:3000/health || exit 1
CMD ["node", "dist/main.js"]
```

## Multi-stage Dockerfile (Python)

```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app
RUN pip install --no-cache-dir poetry
COPY pyproject.toml poetry.lock ./
RUN poetry export -f requirements.txt --output requirements.txt
RUN pip wheel --no-cache-dir --wheel-dir /wheels -r requirements.txt

FROM python:3.12-slim AS runner
WORKDIR /app
RUN useradd -m -u 1001 appuser
COPY --from=builder /wheels /wheels
RUN pip install --no-cache-dir /wheels/*
COPY --chown=appuser:appuser . .
USER appuser
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=3s \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')"
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Docker Compose (Development)

```yaml
version: '3.8'
services:
  app:
    build:
      context: .
      target: builder  # Use dev stage
    volumes:
      - .:/app
      - /app/node_modules
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://user:<db-password>@db:5432/app
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: <db-password>
      POSTGRES_DB: app
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d app"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

## Security Best Practices

| Practice | Implementation |
|----------|----------------|
| Non-root user | `USER nodejs` or `USER 1001` |
| Minimal base image | Use `-alpine` or `-slim` variants |
| No secrets in image | Use runtime env vars or secrets |
| Pin versions | `FROM node:20.10.0-alpine` not `latest` |
| Scan images | `docker scout`, `trivy`, `snyk` |
| Health checks | `HEALTHCHECK` instruction |
| .dockerignore | Exclude `node_modules`, `.git`, `.env` |

## .dockerignore Template

```
node_modules
.git
.env*
*.md
Dockerfile*
docker-compose*
.dockerignore
coverage
.nyc_output
```

---

## Reference: github-actions

# GitHub Actions Pipelines

## Complete CI/CD Pipeline

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test
      - run: npm run lint

  build:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=sha,prefix=
            type=ref,event=branch
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    needs: build
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - run: |
          kubectl set image deployment/app app=${{ needs.build.outputs.image-tag }}

  deploy-production:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - run: |
          kubectl set image deployment/app app=${{ needs.build.outputs.image-tag }}
```

## Common Workflow Patterns

### Matrix Builds (Multi-version testing)
```yaml
jobs:
  test:
    strategy:
      matrix:
        node-version: [18, 20, 22]
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
```

### Reusable Workflows
```yaml
# .github/workflows/deploy.yml
on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
    secrets:
      DEPLOY_KEY:
        required: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ inputs.environment }}
    steps:
      - run: echo "Deploying to ${{ inputs.environment }}"
```

### Caching Dependencies
```yaml
- uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

## Quick Reference

| Action | Purpose |
|--------|---------|
| `actions/checkout@v4` | Clone repository |
| `actions/setup-node@v4` | Install Node.js |
| `docker/build-push-action@v5` | Build and push Docker image |
| `docker/metadata-action@v5` | Generate Docker tags |
| `actions/cache@v4` | Cache dependencies |

---

## Reference: gitlab-ci

# GitLab CI/CD Pipelines

## Complete CI/CD Pipeline

```yaml
# .gitlab-ci.yml — keep the root file short and declarative
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
    - if: $CI_COMMIT_TAG
    - if: $CI_PIPELINE_SOURCE == "schedule"

stages: [validate, test, build, deploy]

default:
  interruptible: true
  retry:
    max: 2
    when: [runner_system_failure, stuck_or_timeout_failure]

include:
  - local: .gitlab/ci/lint.yml
  - local: .gitlab/ci/test.yml
  - local: .gitlab/ci/build.yml
  - local: .gitlab/ci/deploy.yml
```

```yaml
# .gitlab/ci/test.yml
unit-test:
  stage: test
  needs: []                     # fails fast, doesn't wait on validate
  image: node:20
  cache:
    key:
      files: [package-lock.json]
    paths: [.npm/]
  script:
    - npm ci --cache .npm
    - npm test -- --coverage
  coverage: '/All files[^|]*\|[^|]*\s+([\d.]+)/'
  artifacts:
    reports:
      junit: junit.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml
```

```yaml
# .gitlab/ci/build.yml
build-image:
  stage: build
  needs: [unit-test]
  image:
    name: moby/buildkit:rootless
    entrypoint: [""]
  variables:
    BUILDKITD_FLAGS: --oci-worker-no-process-sandbox
  before_script:
    - mkdir -p ~/.docker
    - AUTH=$(echo -n "$CI_REGISTRY_USER:$CI_REGISTRY_PASSWORD" | base64 | tr -d '\n')
    - printf '{"auths":{"%s":{"auth":"%s"}}}' "$CI_REGISTRY" "$AUTH" > ~/.docker/config.json
  script:
    - >
      buildctl-daemonless.sh build
      --frontend dockerfile.v0
      --local context="${CI_PROJECT_DIR}"
      --local dockerfile="${CI_PROJECT_DIR}"
      --output type=image,name="${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHORT_SHA}",push=true
      --export-cache type=registry,ref="${CI_REGISTRY_IMAGE}:buildcache"
      --import-cache type=registry,ref="${CI_REGISTRY_IMAGE}:buildcache"
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
```

```yaml
# .gitlab/ci/deploy.yml
deploy-production:
  stage: deploy
  needs: [build-image]
  environment:
    name: production
    url: https://app.example.com
    deployment_tier: production
  resource_group: production        # serializes concurrent deploys
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      when: manual
  script:
    - kubectl set image deployment/app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA
```

## Core Principles (from GitLab CI/CD best-practices survey)

1. **`workflow:rules` is the master switch** — decide once, at the top, whether a pipeline should exist at all (push vs MR vs schedule vs API/trigger). Prevents duplicate pipelines from the same commit (e.g. branch pipeline + MR pipeline both firing) via `$CI_OPEN_MERGE_REQUESTS` / `$CI_MERGE_REQUEST_DRAFT` checks.
2. **`needs` over `stages` for DAG** — declare exact job dependencies with `needs:` instead of relying on stage ordering. Use `needs: []` for fast checks that should start immediately and fail early. Optimize the critical path, not every job.
3. **Cache vs artifacts are different things** — cache = reusable dependencies (key it off the lockfile, add `fallback_keys` so new branches don't start cold); artifacts = build outputs consumed by later jobs or humans (set `expire_in`, use `expose_as` for reviewer-facing files). Keep separate cache keys for protected vs unprotected refs.
4. **Reuse via `extends` + hidden jobs first, CI/CD components second** — components take typed `inputs` and should be pinned to a tag/SHA (never a moving `~latest`), documented, and treated as a supply-chain dependency if sourced externally.
5. **Environments are first-class objects** — declare `environment:name/url/deployment_tier`, use `on_stop`/`auto_stop_in` for ephemeral review apps, `resource_group` to serialize deploys to the same target, and protected environments + `manual_confirmation` for production.
6. **Secrets never live in CI/CD variables for anything sensitive** — prefer OIDC to cloud providers over static keys; use HashiCorp Vault integration with scoped roles, bound claims, and short TTLs. Be careful with `CI_JOB_TOKEN` scope (limit the allow-list) and treat MRs from forks as untrusted.
7. **Runner blast radius** — register runners at the narrowest scope that works (project < group < instance). Docker executor without `privileged` mode is the default; privileged/DinD only on isolated, ephemeral runners. Split protected and unprotected jobs onto separate runner pools/tags.
8. **Build images without long-lived root daemons** — prefer BuildKit rootless over classic privileged DinD (kaniko is archived and unmaintained; plan a migration if pipelines still use it); pass secrets via mount-type (`--mount=type=secret`), not `ARG`/`ENV`; tag by commit SHA, never `latest`; generate SBOM and sign images (cosign/notation) after build.
9. **MR pipelines are where checks matter** — prefer merged-results pipelines (test the merge of source+target, not just source) and merge trains for high-throughput repos, over relying on branch pipelines.
10. **Report everything GitLab can render** — `artifacts:reports:junit` for test results, `coverage_report` (Cobertura) for MR diff coverage annotations vs the `coverage:` regex for the summary badge, Code Quality reports, and screenshots/videos/logs as artifacts for UI/E2E failures. The MR widget is the primary feedback surface — optimize for it.

## Common Patterns

### Merge request pipelines only (no duplicate branch pipelines)
```yaml
workflow:
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH'
```

### Matrix builds
```yaml
test:
  stage: test
  parallel:
    matrix:
      - NODE_VERSION: ["18", "20", "22"]
        OS: [ubuntu, alpine]
  image: node:${NODE_VERSION}-${OS}
```

### Reusable CI/CD component (pinned, not ~latest)
```yaml
include:
  - component: gitlab.com/my-org/ci-components/deploy@1.4.2
    inputs:
      environment: production
      k8s-namespace: app-prod
```

### Parent/child (monorepo) pipeline
```yaml
trigger-backend:
  trigger:
    include: backend/.gitlab-ci.yml
    strategy: mirror        # parent status mirrors child's real status
  rules:
    - changes: [backend/**/*]
```

### OIDC to a cloud provider (no static keys)
```yaml
deploy:
  id_tokens:
    AWS_ID_TOKEN:
      aud: https://gitlab.example.com
  script:
    - aws sts assume-role-with-web-identity --role-arn $ROLE_ARN --web-identity-token $AWS_ID_TOKEN ...
```

### Dependency Proxy for base images (avoid Docker Hub rate limits)
```yaml
build:
  image: ${CI_DEPENDENCY_PROXY_GROUP_IMAGE_PREFIX}/node:20-alpine
```

## Quick Reference

| Feature | Purpose |
|---|---|
| `workflow:rules` | Decide whether a pipeline is created at all; dedupe push/MR pipelines |
| `needs:` | DAG dependencies between jobs, decoupled from stage order |
| `needs: []` | Job starts immediately, no upstream wait |
| `resource_group` | Serialize deploys to the same environment |
| `environment:deployment_tier` | Classifies env as production/staging/testing/development/other |
| `parallel:matrix` | Fan out a job across variable combinations |
| `extends` | Share config between jobs without `include` overhead |
| CI/CD components (`include:component`) | Typed, versioned, reusable pipeline building blocks |
| `trigger:include` + `strategy: mirror` | Parent/child pipelines for monorepos with real status propagation |
| `id_tokens` (OIDC) | Short-lived cloud credentials instead of static secrets |
| `artifacts:reports:junit` | Test results surfaced in MR widget |
| `artifacts:reports:coverage_report` | Per-line diff coverage annotations in MR |
| `CI_JOB_TOKEN` scope allow-list | Limits which projects a job token can access |

## Anti-Patterns to Avoid

- Hardcoding branch names/environments deep inside job scripts instead of centralizing in `workflow:rules`
- Letting `default:` become a dumping ground that obscures per-job behavior
- Relying on `stages:` ordering alone instead of `needs:` for large pipelines (slow, unclear critical path)
- One shared cache key for protected and unprotected refs (cache poisoning risk)
- Privileged Docker-in-Docker as the default build method
- Long-lived static cloud credentials in CI/CD variables when OIDC is available
- Including third-party CI/CD components at a floating ref instead of a pinned tag/SHA
- Testing only the source branch in MRs instead of merged results

---

## Reference: incident-response

# Incident Response

## Response Metrics

- **MTTD** (Mean Time to Detect): Target < 5 minutes
- **MTTA** (Mean Time to Acknowledge): Target < 5 minutes
- **MTTR** (Mean Time to Resolve): Target < 30 minutes
- **MTBF** (Mean Time Between Failures): Maximize

### Severity Levels

| Level | Impact | Response | Example |
|-------|--------|----------|---------|
| SEV1 | Complete outage | Immediate | Database down, payment failed |
| SEV2 | Major degradation | 15 min | API latency >5s, 50% errors |
| SEV3 | Minor degradation | 1 hour | Non-critical feature broken |
| SEV4 | Low impact | Business hours | UI glitch, logging issues |

## Runbook Template

```markdown
# Runbook: High API Error Rate

## Symptoms
- Alert: `api_error_rate > 0.05`
- Dashboard: https://grafana.example.com/d/api-errors

## Impact
Users cannot complete purchases (~$X per minute)

## Triage
1. Check dashboard for affected endpoints
2. Check recent deployments: `kubectl rollout history deployment/api`
3. Check dependencies: database, redis, external APIs

## Resolution

### Option 1: Rollback
kubectl rollout undo deployment/api -n production

### Option 2: Scale Up
kubectl scale deployment/api --replicas=10 -n production

### Option 3: Fix Config
kubectl set env deployment/api DB_POOL_SIZE=50 -n production

## Verification
- [ ] Error rate <1%
- [ ] P95 latency <500ms
- [ ] Health checks passing

## Communication
- Update status page
- Notify #incidents
- Post if user-facing
```

## Auto-Remediation Script

```python
#!/usr/bin/env python3
import kubernetes, prometheus_api_client

class IncidentRemediator:
    def check_high_error_rate(self):
        query = 'rate(http_requests_total{status=~"5.."}[5m]) > 0.05'
        result = self.prometheus.custom_query(query)
        return len(result) > 0

    def rollback_deployment(self, namespace, deployment):
        body = {'spec': {'rollbackTo': {'revision': 0}}}
        self.k8s.patch_namespaced_deployment(deployment, namespace, body)

    def remediate(self):
        if self.check_high_error_rate():
            if self.rollback_deployment('production', 'api'):
                time.sleep(120)
                if not self.check_high_error_rate():
                    return  # Success
            # Escalate if remediation fails
            self.create_incident("Auto-remediation failed")
```

## Postmortem Template

```markdown
# Postmortem: API Outage - 2024-01-15

**Date**: January 15, 2024
**Duration**: 45 minutes (14:23 - 15:08 UTC)
**Severity**: SEV1
**Impact**: Complete API outage, ~$25K revenue loss

## Summary
API became unresponsive due to database connection pool exhaustion
from slow query in v2.3.1.

## Timeline (UTC)
- 14:23 - Alert fired
- 14:27 - Incident declared SEV1
- 14:30 - Rollback initiated
- 14:45 - Identified slow query
- 14:50 - Killed queries
- 15:08 - Resolved

## Root Cause
New query missing index on `user_id`, causing full table scans that
exhausted connection pool under load.

## Impact
- 100% API failure for 45 minutes
- 15,000 users affected
- $25K revenue loss
- 200+ support tickets

## Action Items
| Action | Owner | Deadline |
|--------|-------|----------|
| Add index on user_id | DB team | 2024-01-16 |
| Add query perf testing | Platform | 2024-01-22 |
| Increase staging DB size | Infra | 2024-01-30 |

## Lessons Learned
- Performance testing must use production-scale data
- Connection pool exhaustion needs active intervention
- Consider circuit breakers for DB operations
```

## PagerDuty Configuration

```yaml
schedules:
  - name: Primary On-Call
    time_zone: America/New_York
    layers:
      - rotation_turn_length_seconds: 604800  # 1 week
        users: [PXXXXXX, PXXXXXX, PXXXXXX]

escalation_policies:
  - name: Production
    rules:
      - escalation_delay_in_minutes: 0
        targets: [{type: schedule, id: primary}]
      - escalation_delay_in_minutes: 15
        targets: [{type: schedule, id: secondary}]
      - escalation_delay_in_minutes: 30
        targets: [{type: user, id: manager}]
```

## Chaos Engineering

```yaml
# chaos-mesh: Pod failure test
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: pod-failure-test
spec:
  action: pod-failure
  mode: one
  duration: "30s"
  selector:
    namespaces: [production]
    labelSelectors:
      app: api
  scheduler:
    cron: "@every 2h"
```

```bash
#!/bin/bash
# Game Day: Database failover drill

echo "🎮 Game Day: Database failover"
slack-cli -d incidents "Starting failover drill"

# Simulate failure
kubectl delete pod postgres-0 -n production

# Monitor recovery
start=$(date +%s)
while ! kubectl get pod postgres-1 | grep Running; do
  sleep 5
done
duration=$(($(date +%s) - start))

echo "Failover: ${duration}s" >> results.md
curl -f https://api.example.com/health || echo "FAIL"
```

## Evidence Collection & Forensics

```bash
#!/bin/bash
# collect-evidence.sh - Preserve incident evidence

INCIDENT_ID=$1
EVIDENCE_DIR="incidents/${INCIDENT_ID}/evidence"
mkdir -p $EVIDENCE_DIR

# Preserve logs
kubectl logs -l app=api --all-containers --timestamps \
  --since=2h > $EVIDENCE_DIR/pod-logs.txt

# Capture current state
kubectl get all -n production -o yaml > $EVIDENCE_DIR/k8s-state.yaml
kubectl describe pods -n production > $EVIDENCE_DIR/pod-details.txt

# Network traces
kubectl exec -n production deploy/api -- \
  tcpdump -i any -w /tmp/capture.pcap -G 60 -W 5 &

# Memory/CPU snapshot
kubectl top pods -n production > $EVIDENCE_DIR/resource-usage.txt

# Git commit at time of incident
git log --since="2 hours ago" --oneline > $EVIDENCE_DIR/recent-commits.txt

# Database queries
psql -c "SELECT * FROM pg_stat_activity" > $EVIDENCE_DIR/db-activity.txt

# Create timeline
echo "$(date): Evidence collection completed" >> $EVIDENCE_DIR/timeline.txt
```

## Communication Templates

```markdown
## SEV1 Initial Notification

**INCIDENT ALERT - SEV1**

**Status**: Investigating
**Impact**: Payment API unavailable (100% error rate)
**Started**: 2024-01-15 14:23 UTC
**Affected**: All users (~15K active sessions)
**Lead**: @oncall-engineer
**War Room**: https://zoom.us/incident-123

Updates every 15 minutes or on major change.

---

## SEV1 Resolution Notification

**INCIDENT RESOLVED**

**Summary**: Payment API restored after database connection pool exhaustion
**Duration**: 45 minutes (14:23 - 15:08 UTC)
**Resolution**: Rollback to v2.3.0 + query optimization
**Impact**: 15K users, ~$25K revenue loss
**Next Steps**: Postmortem scheduled for Jan 16 10am

Thanks to @oncall-team for rapid response.
```

## Incident Classification

| Type | Examples | Response Team | Escalation |
|------|----------|---------------|------------|
| **Security** | Breach, data leak, unauthorized access | Security + DevOps | CISO, Legal |
| **Service** | Outage, degradation, errors | DevOps + SRE | Engineering VP |
| **Data** | Corruption, loss, sync issues | DBA + DevOps | CTO |
| **Compliance** | GDPR, SOC2, audit violations | Compliance + Legal | CEO |
| **Third-party** | Provider outage, API failures | DevOps + Product | Account manager |

## Security Incident Specifics

```bash
# Compromise investigation checklist
□ Isolate affected systems
□ Preserve logs and memory dumps
□ Identify attack vector
□ Check for lateral movement
□ Scan for malware/backdoors
□ Review access logs for 30 days
□ Identify data accessed
□ Assess exfiltration risk
□ Check for persistence mechanisms
□ Coordinate with security team
□ Notify legal if PII involved
□ Document chain of custody
```

## Compliance Requirements

```yaml
# Incident notification requirements
gdpr:
  notification_deadline: 72h
  authority: Data Protection Officer
  required_info:
    - Nature of breach
    - Data categories affected
    - Number of individuals
    - Consequences
    - Remediation measures

sox:
  notification_deadline: immediate
  authority: Audit Committee
  documentation:
    - Financial impact
    - Control failures
    - Remediation plan

pci_dss:
  notification_deadline: 24h
  authority: Card brands + acquirer
  required_info:
    - Cardholder data affected
    - Incident timeline
    - Forensic investigation
```

## Best Practices

- Maintain runbooks for critical services
- Practice with game days monthly
- Automate common remediation
- Keep postmortems blameless
- Track incident metrics
- Test recovery procedures
- Document all incidents
- Improve detection continuously
- Preserve evidence chain properly
- Coordinate communication clearly
- Escalate security incidents immediately
- Understand compliance obligations
- Train team on response procedures
- Review and update playbooks quarterly

---

## Reference: kubernetes

# Kubernetes Manifests

## Complete Deployment Stack

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
        - name: app
          image: ghcr.io/org/app:latest
          ports:
            - containerPort: 3000
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 3000
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 3000
            initialDelaySeconds: 5
            periodSeconds: 5
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: database-url
---
apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector:
    app: app
  ports:
    - port: 80
      targetPort: 3000
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts: [app.example.com]
      secretName: app-tls
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app
                port:
                  number: 80
```

## ConfigMap and Secrets

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  LOG_LEVEL: "info"
  API_TIMEOUT: "30s"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:
  database-url: "postgres://<db-user>:<db-password>@host:5432/db"
```

## Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app
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

## Quick Reference

| Resource | Purpose |
|----------|---------|
| Deployment | Manages ReplicaSets, rolling updates |
| Service | Internal load balancing, DNS |
| Ingress | External HTTP/HTTPS routing |
| ConfigMap | Non-sensitive configuration |
| Secret | Sensitive data (base64 encoded) |
| HPA | Auto-scaling based on metrics |
| PVC | Persistent storage claims |

## Common kubectl Commands

```bash
kubectl apply -f deployment.yaml
kubectl get pods -l app=app
kubectl describe pod <pod-name>
kubectl logs -f <pod-name>
kubectl exec -it <pod-name> -- /bin/sh
kubectl rollout status deployment/app
kubectl rollout undo deployment/app
```

---

## Reference: platform-engineering

# Platform Engineering

## Platform Principles

- **Self-service first**: Reduce manual work to <10%
- **Golden paths**: Pre-approved, opinionated templates
- **Developer experience**: Measure and optimize productivity
- **Platform as product**: Treat with product mindset

## Self-Service with Crossplane

```yaml
# Composition for self-service database
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: postgres-database
spec:
  compositeTypeRef:
    apiVersion: platform.example.com/v1alpha1
    kind: Database
  resources:
    - name: rds-instance
      base:
        apiVersion: rds.aws.crossplane.io/v1alpha1
        kind: DBInstance
        spec:
          forProvider:
            dbInstanceClass: db.t3.micro
            engine: postgres
            engineVersion: "15"
            masterUsername: admin
            allocatedStorage: 20
```

## Terraform Self-Service Module

```hcl
# modules/service/main.tf
variable "service_name" {}
variable "environment" {}

module "k8s_service" {
  source   = "./k8s-deployment"
  name     = var.service_name
  env      = var.environment
}

module "database" {
  source = "./postgres"
  name   = "${var.service_name}-db"
}

module "monitoring" {
  source  = "./monitoring-stack"
  service = var.service_name
}

output "service_url" {
  value = module.k8s_service.url
}
```

## Backstage Service Template

```yaml
# templates/microservice/template.yaml
apiVersion: scaffolder.backstage.io/v1beta3
kind: Template
metadata:
  name: microservice-template
  title: Microservice Golden Path
spec:
  owner: platform-team
  type: service
  parameters:
    - title: Service Info
      properties:
        name:
          type: string
        owner:
          type: string
          ui:field: OwnerPicker
        language:
          type: string
          enum: [go, python, nodejs, java]
  steps:
    - id: fetch
      action: fetch:template
      input:
        url: ./skeleton
        values:
          name: ${{ parameters.name }}
    - id: publish
      action: publish:github
      input:
        repoUrl: github.com?owner=org&repo=${{ parameters.name }}
    - id: register
      action: catalog:register
```

## Service Catalog Info

```yaml
# catalog-info.yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: payment-service
  annotations:
    github.com/project-slug: org/payment-service
    pagerduty.com/integration-key: abc123
    grafana/dashboard-selector: service=payment
spec:
  type: service
  lifecycle: production
  owner: payments-team
  system: checkout
  dependsOn:
    - resource:default/payment-db
    - component:default/auth-service
  providesApis:
    - payment-api
```

## Golden Path Scaffolding

```bash
#!/bin/bash
# create-service.sh - Golden path for new services

SERVICE=$1
LANG=$2

# Create from template
gh repo create "org/$SERVICE" --template "org/template-$LANG"
git clone "git@github.com:org/$SERVICE.git"
cd "$SERVICE"

# Setup CI/CD
cat > .github/workflows/ci.yml <<EOF
name: CI/CD
on: [push]
jobs:
  pipeline:
    uses: org/workflows/.github/workflows/standard.yml@v1
    with:
      service_name: $SERVICE
EOF

# Create infrastructure
cat > terraform/main.tf <<EOF
module "service" {
  source = "git::https://github.com/org/terraform//service"
  name   = "$SERVICE"
}
EOF

git add . && git commit -m "Golden path init" && git push

echo "✓ Service created! Merge to main to deploy."
```

## GitOps Repository Structure

```
gitops/
├── apps/
│   ├── production/
│   │   ├── payment-service/
│   │   └── auth-service/
│   └── staging/
│       └── payment-service/
├── infrastructure/
│   ├── clusters/
│   │   ├── prod-us-east/
│   │   └── prod-eu-west/
│   └── base/
│       ├── ingress/
│       └── monitoring/
└── platform/
    ├── backstage/
    ├── argocd/
    └── vault/
```

## ArgoCD Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: payment-service
spec:
  project: default
  source:
    repoURL: https://github.com/org/gitops
    path: apps/production/payment-service
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    retry:
      limit: 5
      backoff:
        duration: 5s
        maxDuration: 3m
```

## Platform Metrics

```yaml
# prometheus/platform-metrics.yaml
groups:
  - name: platform
    rules:
      # Self-service adoption rate
      - record: platform:self_service:rate
        expr: |
          sum(rate(platform_provision_automated[1h]))
          /
          sum(rate(platform_provision_total[1h]))

      # Provisioning time P95
      - record: platform:provision:p95
        expr: |
          histogram_quantile(0.95,
            rate(platform_provision_duration_bucket[5m]))

      # Golden path adoption
      - record: platform:golden_path:adoption
        expr: |
          count(service{template="golden-path"})
          / count(service)
```

## Custom Backstage Plugin

```typescript
// plugins/platform-stats/PlatformMetrics.tsx
import React from 'react';
import { InfoCard, Progress } from '@backstage/core-components';

export const PlatformMetrics = () => {
  const metrics = {
    selfServiceRate: 92,
    avgProvisionTime: '3.5min',
    uptime: '99.95%',
    satisfaction: 4.6
  };

  return (
    <InfoCard title="Platform Health">
      <Progress value={metrics.selfServiceRate} label="Self-Service" />
      <p>Provision Time: {metrics.avgProvisionTime}</p>
      <p>Uptime: {metrics.uptime}</p>
      <p>Satisfaction: {metrics.satisfaction}/5</p>
    </InfoCard>
  );
};
```

## Cost Allocation

```yaml
# kubecost/allocation.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cost-allocation
data:
  allocation.json: |
    {
      "defaultLabels": {
        "team": "team",
        "service": "app",
        "environment": "env"
      },
      "shareNamespaces": ["kube-system"],
      "shareCost": "weighted"
    }
```

## Platform APIs

```python
# Platform API for self-service provisioning
from fastapi import FastAPI, Depends
from pydantic import BaseModel

app = FastAPI()

class ServiceRequest(BaseModel):
    name: str
    environment: str
    language: str
    database: bool = False

@app.post("/api/v1/services")
async def create_service(request: ServiceRequest):
    # Validate and enqueue
    task = platform.provision_service(
        name=request.name,
        env=request.environment,
        template=f"golden-path-{request.language}"
    )
    return {"task_id": task.id, "status": "provisioning"}

@app.get("/api/v1/services/{name}/status")
async def service_status(name: str):
    return {
        "status": "running",
        "url": f"https://{name}.example.com",
        "health": "healthy",
        "cost_mtd": "$142.50"
    }
```

## Multi-Tenant Architecture

```yaml
# Policy: Resource quotas per tenant
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: team-payments
spec:
  hard:
    requests.cpu: "20"
    requests.memory: 40Gi
    persistentvolumeclaims: "10"
    services.loadbalancers: "2"
---
# RBAC: Namespace admin
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: team-admin
  namespace: team-payments
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: namespace-admin
subjects:
  - kind: Group
    name: team-payments
```

## Adoption Strategy

```yaml
# Platform metrics tracking
apiVersion: v1
kind: ConfigMap
metadata:
  name: platform-goals
data:
  goals.yaml: |
    q1_2024:
      self_service_rate: 90%
      avg_provision_time: 5min
      developer_satisfaction: 4.5/5
      golden_path_adoption: 80%

    tracking:
      weekly_provisioning: true
      team_feedback: true
      support_tickets: true
      training_completion: true
```

## CLI Tool Example

```bash
#!/bin/bash
# platform-cli - Self-service CLI

platform() {
  case $1 in
    create)
      curl -X POST $PLATFORM_API/services \
        -d "{\"name\":\"$2\",\"env\":\"$3\",\"language\":\"$4\"}"
      ;;
    status)
      curl $PLATFORM_API/services/$2/status | jq
      ;;
    logs)
      kubectl logs -l app=$2 -n ${3:-staging} --tail=100
      ;;
    cost)
      curl $PLATFORM_API/services/$2/cost?period=mtd
      ;;
  esac
}
```

## Best Practices

- Design for self-service from day one
- Make golden paths the easiest option
- Measure developer satisfaction continuously
- Automate platform operations
- Provide excellent documentation
- Build APIs, not just tools
- Enable safe experimentation
- Maintain backward compatibility
- Treat platform as a product
- Gather and act on feedback
- Track adoption metrics weekly
- Run platform as a product team
- Invest in developer evangelism
- Maintain SLOs for platform uptime
- Provide fast, helpful support

---

## Reference: release-automation

# Release Automation

## Artifact Management

### Container Registry Lifecycle

```json
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Keep last 10 prod images",
      "selection": {
        "tagStatus": "tagged",
        "tagPrefixList": ["prod-"],
        "countType": "imageCountMoreThan",
        "countNumber": 10
      },
      "action": {"type": "expire"}
    },
    {
      "rulePriority": 2,
      "description": "Remove untagged after 7 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 7
      },
      "action": {"type": "expire"}
    }
  ]
}
```

### Artifact Promotion

```yaml
# .github/workflows/promote.yml
name: Artifact Promotion

on:
  workflow_dispatch:
    inputs:
      image_tag:
        required: true
      target_env:
        type: choice
        options: [staging, production]

jobs:
  promote:
    runs-on: ubuntu-latest
    steps:
      - name: Re-tag for environment
        run: |
          docker pull $REGISTRY/$IMAGE:${{ inputs.image_tag }}
          docker tag $REGISTRY/$IMAGE:${{ inputs.image_tag }} \
            $REGISTRY/$IMAGE:${{ inputs.target_env }}-latest
          docker push $REGISTRY/$IMAGE:${{ inputs.target_env }}-latest

      - name: Sign artifact
        uses: sigstore/cosign-installer@v3
      - run: cosign sign $REGISTRY/$IMAGE:${{ inputs.target_env }}-latest

      - name: Update GitOps
        run: |
          cd gitops/apps/${{ inputs.target_env }}
          yq e '.image.tag = "${{ inputs.image_tag }}"' -i values.yaml
          git commit -am "Promote to ${{ inputs.target_env }}"
          git push
```

## Feature Flags

### LaunchDarkly Integration

```python
import launchdarkly

ld = launchdarkly.get()

def should_enable(user_id, feature_key):
    user = {"key": user_id, "custom": {"groups": get_groups(user_id)}}
    return ld.variation(feature_key, user, False)

# Usage
if should_enable(user.id, "new-payment-flow"):
    return new_payment_service.process(payment)
else:
    return legacy_payment_service.process(payment)
```

### Flagger Progressive Delivery

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: payment-service
spec:
  targetRef:
    kind: Deployment
    name: payment-service
  service:
    port: 8080
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
      - name: request-success-rate
        thresholdRange:
          min: 99
      - name: request-duration
        thresholdRange:
          max: 500
    webhooks:
      - name: load-test
        url: http://flagger-loadtester/
        metadata:
          cmd: "hey -z 1m -q 10 http://payment-canary/"
```

## Multi-Platform CI/CD

### GitLab CI

```yaml
stages: [test, build, deploy]

test:
  stage: test
  image: node:20
  script:
    - npm ci && npm test

build:
  stage: build
  image: docker:latest
  services: [docker:dind]
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

deploy:production:
  stage: deploy
  script:
    - kubectl set image deployment/app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  environment: production
  when: manual
  only: [main]
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    environment {
        IMAGE = "registry.example.com/app"
    }

    stages {
        stage('Test') {
            steps {
                sh 'npm ci && npm test'
                junit 'reports/junit.xml'
            }
        }

        stage('Build') {
            steps {
                script {
                    docker.build("${IMAGE}:${BUILD_NUMBER}")
                }
            }
        }

        stage('Security Scan') {
            steps {
                sh "trivy image ${IMAGE}:${BUILD_NUMBER}"
            }
        }

        stage('Deploy Staging') {
            when { branch 'main' }
            steps {
                sh "kubectl set image deployment/app app=${IMAGE}:${BUILD_NUMBER} -n staging"
            }
        }

        stage('Deploy Production') {
            when { branch 'main' }
            steps {
                input 'Deploy to production?'
                sh "kubectl set image deployment/app app=${IMAGE}:${BUILD_NUMBER} -n production"
            }
        }
    }

    post {
        failure {
            slackSend color: 'danger', message: "Build failed: ${JOB_NAME}"
        }
    }
}
```

## Build Optimization

### Multi-stage Docker Build

```dockerfile
FROM node:20 AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:20 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-slim AS runner
WORKDIR /app
ENV NODE_ENV production
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
USER node
CMD ["node", "dist/main.js"]
```

### Parallel Testing

```yaml
# CircleCI
version: 2.1
jobs:
  test:
    parallelism: 4
    docker:
      - image: cimg/node:20
    steps:
      - checkout
      - run: npm ci
      - run: |
          TESTS=$(circleci tests glob "test/**/*.js" | circleci tests split)
          npm test $TESTS
```

## Release Orchestration

```bash
#!/bin/bash
# release.sh - Multi-service coordinated release

VERSION=$1
SERVICES=(auth api worker frontend)

echo "Release: $VERSION"

# Create release branches
for svc in "${SERVICES[@]}"; do
    gh api repos/org/$svc/git/refs -f ref=refs/heads/release/$VERSION -f sha=$(git rev-parse main)
done

# Trigger builds
for svc in "${SERVICES[@]}"; do
    gh workflow run ci.yml --repo org/$svc --ref release/$VERSION
done

# Wait for completion
for svc in "${SERVICES[@]}"; do
    gh run watch --repo org/$svc $(gh run list --repo org/$svc -L1 -q '.[0].databaseId')
done

# Deploy to staging
kubectl apply -f staging/release-$VERSION.yaml

# Smoke tests
./scripts/smoke-test.sh staging

echo "✓ Release $VERSION ready for production"
```

## Dependency Management

### Renovate Auto-Update

```json
{
  "extends": ["config:base"],
  "packageRules": [
    {
      "matchUpdateTypes": ["minor", "patch"],
      "automerge": true
    },
    {
      "matchDepTypes": ["devDependencies"],
      "automerge": true
    }
  ],
  "schedule": ["before 6am on Monday"],
  "prConcurrentLimit": 5
}
```

## Build Optimization

### Build Caching Strategy

```yaml
# GitHub Actions: Multi-layer caching
- name: Cache dependencies
  uses: actions/cache@v3
  with:
    path: |
      ~/.npm
      ~/.cache
      node_modules
    key: ${{ runner.os }}-deps-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-deps-

- name: Cache Docker layers
  uses: docker/build-push-action@v4
  with:
    context: .
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Parallel CI Pipeline

```yaml
# Multi-platform builds in parallel
name: Build

on: [push]

jobs:
  test:
    strategy:
      matrix:
        node: [18, 20, 22]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node }}
      - run: npm ci && npm test

  build-images:
    needs: test
    strategy:
      matrix:
        platform: [linux/amd64, linux/arm64]
    runs-on: ubuntu-latest
    steps:
      - uses: docker/build-push-action@v4
        with:
          platforms: ${{ matrix.platform }}
          tags: app:${{ github.sha }}
```

## Multi-Service Release Orchestration

```yaml
# release-coordinator.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: release-v2.5.0
spec:
  template:
    spec:
      containers:
        - name: coordinator
          image: release-bot:latest
          env:
            - name: RELEASE_VERSION
              value: "v2.5.0"
            - name: SERVICES
              value: "auth,api,worker,frontend"
          command:
            - /bin/bash
            - -c
            - |
              # Deploy in dependency order
              for svc in auth api worker frontend; do
                echo "Deploying $svc..."
                kubectl set image deploy/$svc \
                  $svc=registry.io/$svc:$RELEASE_VERSION

                kubectl rollout status deploy/$svc --timeout=5m

                # Health check
                kubectl run test-$svc --rm -i --restart=Never \
                  --image=curlimages/curl -- \
                  curl -f http://$svc/health

                echo "$svc deployed successfully"
              done
```

## Advanced Artifact Management

```bash
#!/bin/bash
# artifact-scanner.sh - Scan before promotion

IMAGE=$1
SEVERITY=${2:-HIGH}

# Vulnerability scan
trivy image --severity $SEVERITY --exit-code 1 $IMAGE

# License compliance
syft $IMAGE -o json | \
  jq '.artifacts[].licenses[] | select(.value |
    contains("GPL") or contains("AGPL"))' && \
  echo "License violation detected" && exit 1

# SBOM generation
syft $IMAGE -o spdx-json > sbom-$(basename $IMAGE).spdx.json

# Sign artifact
cosign sign --key cosign.key $IMAGE

# Promote
docker tag $IMAGE $IMAGE-approved
docker push $IMAGE-approved

echo "Artifact $IMAGE approved and promoted"
```

## Zero-Downtime Database Migrations

```python
# migrations/release_v2.5.py
from alembic import op
import sqlalchemy as sa

def upgrade():
    # Step 1: Add new column (nullable)
    op.add_column('users',
      sa.Column('email_verified', sa.Boolean(), nullable=True))

    # Step 2: Backfill data (in batches)
    connection = op.get_bind()
    connection.execute("""
      UPDATE users SET email_verified = true
      WHERE email IS NOT NULL
      LIMIT 1000
    """)
    # Repeat until complete (or use background job)

    # Step 3: Make non-nullable (in next release)
    # op.alter_column('users', 'email_verified', nullable=False)

def downgrade():
    op.drop_column('users', 'email_verified')
```

## Release Metrics Dashboard

```yaml
# Grafana dashboard for release metrics
apiVersion: v1
kind: ConfigMap
metadata:
  name: release-dashboard
data:
  dashboard.json: |
    {
      "panels": [
        {
          "title": "Deployment Frequency",
          "targets": [{
            "expr": "count_over_time(deployment_completed[1d])"
          }]
        },
        {
          "title": "Lead Time",
          "targets": [{
            "expr": "histogram_quantile(0.95, commit_to_deploy_seconds_bucket)"
          }]
        },
        {
          "title": "Change Failure Rate",
          "targets": [{
            "expr": "sum(rate(deployment_failed[1h])) / sum(rate(deployment_total[1h]))"
          }]
        },
        {
          "title": "Active Releases",
          "targets": [{
            "expr": "count(release_in_progress == 1)"
          }]
        }
      ]
    }
```

## Best Practices

- Version artifacts with immutable tags
- Implement retention policies
- Use progressive delivery for high-risk changes
- Automate security scanning
- Maintain deployment audit trails
- Enable easy rollbacks
- Monitor deployment metrics
- Use feature flags for flexibility
- Cache aggressively for fast builds
- Parallelize test and build jobs
- Coordinate multi-service releases
- Generate and track SBOMs
- Sign artifacts for supply chain security
- Automate dependency updates
- Track DORA metrics continuously

---

## Reference: terraform-iac

# Terraform Infrastructure as Code

## AWS ECS Fargate Setup

```hcl
terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
  backend "s3" {
    bucket = "terraform-state"
    key    = "app/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_ecs_cluster" "main" {
  name = "app-cluster"
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_ecs_task_definition" "app" {
  family                   = "app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name  = "app"
    image = "${var.ecr_repository}:${var.image_tag}"
    portMappings = [{ containerPort = 3000 }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.app.name
        awslogs-region        = var.region
        awslogs-stream-prefix = "app"
      }
    }
    secrets = [
      { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.db_url.arn }
    ]
  }])
}

resource "aws_ecs_service" "app" {
  name            = "app"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 2
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnets
    security_groups = [aws_security_group.app.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = 3000
  }
}
```

## AWS RDS PostgreSQL

```hcl
resource "aws_db_instance" "postgres" {
  identifier           = "app-db"
  engine               = "postgres"
  engine_version       = "16.1"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  storage_encrypted    = true

  db_name              = "app"
  username             = "admin"
  password             = var.db_password

  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  backup_retention_period = 7
  skip_final_snapshot     = false
  final_snapshot_identifier = "app-db-final"

  tags = { Environment = var.environment }
}
```

## Variables and Outputs

```hcl
# variables.tf
variable "environment" {
  type        = string
  description = "Environment name"
}

variable "region" {
  type    = string
  default = "us-east-1"
}

# outputs.tf
output "ecs_cluster_arn" {
  value = aws_ecs_cluster.main.arn
}

output "alb_dns_name" {
  value = aws_lb.main.dns_name
}
```

## Best Practices

| Practice | Implementation |
|----------|----------------|
| State locking | S3 backend with DynamoDB |
| Secrets | Use AWS Secrets Manager / SSM |
| Modules | Reusable components |
| Workspaces | Environment separation |
| Tagging | Consistent resource tags |
| Validation | `terraform validate`, `tflint` |

## Common Commands

```bash
terraform init
terraform plan -out=tfplan
terraform apply tfplan
terraform destroy
terraform state list
terraform import aws_instance.app i-1234567890
```
