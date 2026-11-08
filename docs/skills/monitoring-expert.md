# monitoring-expert

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/monitoring-expert (MIT). Рабочая копия скилла: `.claude/skills/monitoring-expert/`.

---
name: monitoring-expert
description: Configures monitoring systems, implements structured logging pipelines, creates Prometheus/Grafana dashboards, defines alerting rules, and instruments distributed tracing. Implements Prometheus/Grafana stacks, conducts load testing, performs application profiling, and plans infrastructure capacity. Use when setting up application monitoring, adding observability to services, debugging production issues with logs/metrics/traces, running load tests with k6 or Artillery, profiling CPU/memory bottlenecks, or forecasting capacity needs.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: devops
  triggers: monitoring, observability, logging, metrics, tracing, alerting, Prometheus, Grafana, DataDog, APM, performance testing, load testing, profiling, capacity planning, bottleneck
  role: specialist
  scope: implementation
  output-format: code
  related-skills: devops-engineer, debugging-wizard, architecture-designer
---

# Monitoring Expert

Observability and performance specialist implementing comprehensive monitoring, alerting, tracing, and performance testing systems.

## Core Workflow

1. **Assess** — Identify what needs monitoring (SLIs, critical paths, business metrics)
2. **Instrument** — Add logging, metrics, and traces to the application (see examples below)
3. **Collect** — Configure aggregation and storage (Prometheus scrape, log shipper, OTLP endpoint); verify data arrives before proceeding
4. **Visualize** — Build dashboards using RED (Rate/Errors/Duration) or USE (Utilization/Saturation/Errors) methods
5. **Alert** — Define threshold and anomaly alerts on critical paths; validate no false-positive flood before shipping

## Quick-Start Examples

### Structured Logging (Node.js / Pino)
```js
import pino from 'pino';

const logger = pino({ level: 'info' });

// Good — structured fields, includes correlation ID
logger.info({ requestId: req.id, userId: req.user.id, durationMs: elapsed }, 'order.created');

// Bad — string interpolation, no correlation
console.log(`Order created for user ${userId}`);
```

### Prometheus Metrics (Node.js)
```js
import { Counter, Histogram, register } from 'prom-client';

const httpRequests = new Counter({
  name: 'http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'route', 'status'],
});

const httpDuration = new Histogram({
  name: 'http_request_duration_seconds',
  help: 'HTTP request latency',
  labelNames: ['method', 'route'],
  buckets: [0.05, 0.1, 0.3, 0.5, 1, 2, 5],
});

// Instrument a route
app.use((req, res, next) => {
  const end = httpDuration.startTimer({ method: req.method, route: req.path });
  res.on('finish', () => {
    httpRequests.inc({ method: req.method, route: req.path, status: res.statusCode });
    end();
  });
  next();
});

// Expose scrape endpoint
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});
```

### OpenTelemetry Tracing (Node.js)
```js
import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { trace } from '@opentelemetry/api';

const sdk = new NodeSDK({
  traceExporter: new OTLPTraceExporter({ url: 'http://jaeger:4318/v1/traces' }),
});
sdk.start();

// Manual span around a critical operation
const tracer = trace.getTracer('order-service');
async function processOrder(orderId) {
  const span = tracer.startSpan('order.process');
  span.setAttribute('order.id', orderId);
  try {
    const result = await db.saveOrder(orderId);
    span.setStatus({ code: SpanStatusCode.OK });
    return result;
  } catch (err) {
    span.recordException(err);
    span.setStatus({ code: SpanStatusCode.ERROR });
    throw err;
  } finally {
    span.end();
  }
}
```

### Prometheus Alerting Rule
```yaml
groups:
  - name: api.rules
    rules:
      - alert: HighErrorRate
        expr: |
          rate(http_requests_total{status=~"5.."}[5m])
          / rate(http_requests_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 5% on {{ $labels.route }}"
```

### k6 Load Test
```js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },   // ramp up
    { duration: '5m', target: 50 },   // sustained load
    { duration: '1m', target: 0 },    // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95th percentile < 500 ms
    http_req_failed:   ['rate<0.01'],  // error rate < 1%
  },
};

export default function () {
  const res = http.get('https://api.example.com/orders');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
```

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Logging | `references/structured-logging.md` | Pino, JSON logging |
| Metrics | `references/prometheus-metrics.md` | Counter, Histogram, Gauge |
| Tracing | `references/opentelemetry.md` | OpenTelemetry, spans |
| Alerting | `references/alerting-rules.md` | Prometheus alerts |
| Dashboards | `references/dashboards.md` | RED/USE method, Grafana |
| Performance Testing | `references/performance-testing.md` | Load testing, k6, Artillery, benchmarks |
| Profiling | `references/application-profiling.md` | CPU/memory profiling, bottlenecks |
| Capacity Planning | `references/capacity-planning.md` | Scaling, forecasting, budgets |

## Constraints

### MUST DO
- Use structured logging (JSON)
- Include request IDs for correlation
- Set up alerts for critical paths
- Monitor business metrics, not just technical
- Use appropriate metric types (counter/gauge/histogram)
- Implement health check endpoints

### MUST NOT DO
- Log sensitive data (passwords, tokens, PII)
- Alert on every error (alert fatigue)
- Use string interpolation in logs (use structured fields)
- Skip correlation IDs in distributed systems

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/devops/monitoring-expert/)

---

## Reference: alerting-rules

# Alerting Rules

## Prometheus Alert Rules

```yaml
# alerts.yml
groups:
  - name: application
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          description: Error rate is {{ $value | humanizePercentage }}

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High latency detected
          description: 95th percentile latency is {{ $value }}s

      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: Service {{ $labels.instance }} is down

  - name: infrastructure
    rules:
      - alert: HighMemoryUsage
        expr: |
          (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes)
          / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High memory usage on {{ $labels.instance }}

      - alert: HighCPUUsage
        expr: |
          100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: High CPU usage on {{ $labels.instance }}

      - alert: DiskSpaceLow
        expr: |
          (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: Disk space low on {{ $labels.instance }}
```

## Alert Design Principles

```yaml
# Good alert: Actionable, specific
- alert: DatabaseConnectionPoolExhausted
  expr: db_pool_available_connections == 0
  for: 2m
  annotations:
    runbook_url: https://wiki.example.com/runbooks/db-pool

# Bad alert: Too noisy, not actionable
- alert: AnyError
  expr: errors_total > 0  # Will always fire
```

## Severity Levels

| Severity | Response | Example |
|----------|----------|---------|
| `critical` | Page immediately | Service down, data loss |
| `warning` | Investigate soon | High latency, low disk |
| `info` | Check in morning | Unusual traffic pattern |

## Alertmanager Configuration

```yaml
# alertmanager.yml
global:
  slack_api_url: '<slack-webhook-url>'

route:
  receiver: 'slack-notifications'
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
    - match:
        severity: warning
      receiver: 'slack-notifications'

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - channel: '#alerts'
        send_resolved: true

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '<pagerduty-service-key>'
```

## Quick Reference

| Field | Purpose |
|-------|---------|
| `expr` | PromQL query |
| `for` | Duration before firing |
| `labels` | Classification (severity) |
| `annotations` | Human-readable info |

| Threshold | Use |
|-----------|-----|
| Error rate > 5% | Critical |
| p95 latency > 1s | Warning |
| Disk < 10% | Critical |
| Memory > 90% | Warning |

---

## Reference: application-profiling

# Application Profiling

## Node.js Profiling

### CPU Profiling with clinic.js

```bash
# Install
npm install -g clinic

# CPU profiling
clinic doctor -- node app.js

# Flame graph
clinic flame -- node app.js

# Bubble profiling
clinic bubbleprof -- node app.js

# Generate report
clinic doctor --collect-only -- node app.js
clinic doctor --visualize-only PID.clinic-doctor
```

### Built-in Node.js Profiler

```javascript
// Start profiling
node --prof app.js

# Process the output
node --prof-process isolate-0x*.log > processed.txt

# Chrome DevTools
node --inspect app.js
# Open chrome://inspect
```

### Memory Profiling

```javascript
import v8 from 'v8';
import fs from 'fs';

// Heap snapshot
const snapshot = v8.writeHeapSnapshot();
console.log('Snapshot written to:', snapshot);

// Memory usage
const usage = process.memoryUsage();
console.log({
  rss: `${Math.round(usage.rss / 1024 / 1024)}MB`,
  heapTotal: `${Math.round(usage.heapTotal / 1024 / 1024)}MB`,
  heapUsed: `${Math.round(usage.heapUsed / 1024 / 1024)}MB`,
  external: `${Math.round(usage.external / 1024 / 1024)}MB`,
});
```

### Custom Performance Marks

```javascript
import { performance, PerformanceObserver } from 'perf_hooks';

// Mark start
performance.mark('operation-start');

// ... do work ...
await processOrder(orderId);

// Mark end
performance.mark('operation-end');

// Measure
performance.measure('operation', 'operation-start', 'operation-end');

// Observer
const obs = new PerformanceObserver((items) => {
  items.getEntries().forEach((entry) => {
    console.log(`${entry.name}: ${entry.duration}ms`);
  });
});
obs.observe({ entryTypes: ['measure'] });
```

## Python Profiling

### cProfile

```python
import cProfile
import pstats

# Profile a function
def main():
    # Your code here
    process_data()

if __name__ == '__main__':
    profiler = cProfile.Profile()
    profiler.enable()

    main()

    profiler.disable()
    stats = pstats.Stats(profiler)
    stats.sort_stats('cumulative')
    stats.print_stats(20)  # Top 20 functions
```

### Line Profiler

```python
from line_profiler import LineProfiler

@profile
def expensive_function():
    # Code to profile
    result = []
    for i in range(10000):
        result.append(i ** 2)
    return result

# Run with: kernprof -l -v script.py
```

### Memory Profiler

```python
from memory_profiler import profile

@profile
def process_large_data():
    data = [i for i in range(1000000)]
    result = [x * 2 for x in data]
    return result

# Run with: python -m memory_profiler script.py
```

### py-spy

```bash
# CPU sampling (live process)
py-spy top --pid 12345

# Generate flame graph
py-spy record -o profile.svg --pid 12345

# Record for duration
py-spy record -o profile.svg --duration 60 -- python app.py
```

## Go Profiling

### pprof

```go
import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
)

func main() {
    // Enable profiling endpoint
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Your application code
}
```

```bash
# CPU profile
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Memory profile
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Goroutine profile
curl http://localhost:6060/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof

# Web interface
go tool pprof -http=:8080 cpu.prof
```

## Java Profiling

### VisualVM

```bash
# Start application with JMX
java -Dcom.sun.management.jmxremote \
     -Dcom.sun.management.jmxremote.port=9010 \
     -Dcom.sun.management.jmxremote.authenticate=false \
     -Dcom.sun.management.jmxremote.ssl=false \
     -jar app.jar

# Connect with VisualVM
jvisualvm
```

### async-profiler

```bash
# CPU profiling
./profiler.sh -d 30 -f cpu.html <pid>

# Allocation profiling
./profiler.sh -d 30 -e alloc -f alloc.html <pid>

# Flame graph
./profiler.sh -d 30 -f flamegraph.svg <pid>
```

## Database Query Profiling

### PostgreSQL

```sql
-- Enable query logging
ALTER SYSTEM SET log_min_duration_statement = 100;  -- Log queries > 100ms
SELECT pg_reload_conf();

-- Explain analyze
EXPLAIN ANALYZE
SELECT * FROM orders
WHERE user_id = 123
AND created_at > NOW() - INTERVAL '30 days';

-- Track slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

### MySQL

```sql
-- Enable slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 0.1;  -- 100ms

-- Explain query
EXPLAIN ANALYZE
SELECT * FROM orders
WHERE user_id = 123;

-- Performance schema
SELECT * FROM performance_schema.events_statements_summary_by_digest
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 10;
```

## APM Integration

### New Relic

```javascript
import newrelic from 'newrelic';

// Custom transaction
newrelic.startBackgroundTransaction('process-orders', async () => {
  const orders = await getOrders();

  // Custom segment
  await newrelic.startSegment('validate-orders', true, async () => {
    return validateOrders(orders);
  });
});

// Custom metrics
newrelic.recordMetric('Custom/OrderValue', orderTotal);
```

### DataDog APM

```javascript
import tracer from 'dd-trace';
tracer.init();

// Custom span
const span = tracer.startSpan('process.order', {
  resource: orderId,
  tags: {
    'order.total': orderTotal,
    'user.id': userId,
  },
});

try {
  await processOrder(orderId);
  span.setTag('status', 'success');
} catch (err) {
  span.setTag('error', err);
} finally {
  span.finish();
}
```

## Quick Reference

| Tool | Language | Type |
|------|----------|------|
| clinic.js | Node.js | CPU, Event loop |
| Chrome DevTools | Node.js | CPU, Memory |
| cProfile | Python | CPU |
| py-spy | Python | CPU (sampling) |
| pprof | Go | CPU, Memory, Goroutines |
| VisualVM | Java | CPU, Memory, Threads |
| async-profiler | Java | CPU, Allocation |

| Metric | What to Look For |
|--------|------------------|
| CPU time | Hot functions, tight loops |
| Memory | Large allocations, leaks |
| I/O wait | Blocking operations |
| GC time | Excessive collections |
| Thread count | Thread pool saturation |

| Problem | Symptom |
|---------|---------|
| CPU bound | High CPU usage, slow processing |
| Memory leak | Growing memory, eventual crash |
| I/O bound | Low CPU, high wait time |
| Lock contention | Idle threads, poor scaling |

---

## Reference: capacity-planning

# Capacity Planning

## Growth Projection

### Linear Projection

```python
import pandas as pd
import numpy as np
from sklearn.linear_model import LinearRegression

# Historical data
data = pd.DataFrame({
    'month': range(1, 13),
    'requests_per_second': [100, 120, 145, 160, 180, 200, 220, 245, 270, 290, 310, 330]
})

# Train model
model = LinearRegression()
X = data[['month']].values
y = data['requests_per_second'].values
model.fit(X, y)

# Forecast next 6 months
future_months = np.array([[13], [14], [15], [16], [17], [18]])
predictions = model.predict(future_months)

print("Projected RPS in 6 months:", predictions[-1])
```

### Prometheus Queries for Trends

```promql
# Monthly growth rate
(
  rate(http_requests_total[30d])
  /
  rate(http_requests_total[30d] offset 30d)
) - 1

# Predict resource exhaustion
predict_linear(
  node_memory_MemAvailable_bytes[1h],
  3600 * 24 * 30  # 30 days ahead
)

# Storage growth
predict_linear(
  node_filesystem_avail_bytes[7d],
  3600 * 24 * 90  # 90 days ahead
)
```

## Resource Forecasting

### CPU Requirements

```javascript
// Current capacity
const currentRPS = 1000;
const currentCPU = 0.65;  // 65% utilization
const targetCPU = 0.70;   // Target 70% max

// Projected load
const projectedRPS = 2500;

// Required CPU capacity
const cpuScalingFactor = projectedRPS / currentRPS;
const requiredCPU = (currentCPU * cpuScalingFactor) / targetCPU;

console.log(`Current: ${currentRPS} RPS @ ${currentCPU * 100}% CPU`);
console.log(`Projected: ${projectedRPS} RPS requires ${requiredCPU.toFixed(2)}x CPU`);
```

### Memory Requirements

```javascript
// Memory per request (average)
const avgMemoryPerRequest = 2048;  // bytes
const concurrentRequests = 500;
const overhead = 1.3;  // 30% overhead for GC, OS, etc.

const requiredMemory = (avgMemoryPerRequest * concurrentRequests * overhead) / (1024 ** 3);
console.log(`Required memory: ${requiredMemory.toFixed(2)} GB`);
```

### Database Connections

```javascript
// Connections per instance
const connectionsPerInstance = 100;
const instances = 5;
const utilizationTarget = 0.75;

// Available connections
const totalConnections = connectionsPerInstance * instances;
const effectiveConnections = totalConnections * utilizationTarget;

// RPS capacity
const avgRequestsPerConnection = 10;
const maxRPS = effectiveConnections * avgRequestsPerConnection;

console.log(`Max sustainable RPS: ${maxRPS}`);
```

## Scaling Strategies

### Horizontal Scaling Calculator

```javascript
function calculateInstances(targetRPS, instanceCapacity, bufferPercent = 20) {
  // Account for buffer
  const effectiveCapacity = instanceCapacity * (1 - bufferPercent / 100);

  // Calculate required instances
  const requiredInstances = Math.ceil(targetRPS / effectiveCapacity);

  // Account for availability zones
  const minInstancesPerAZ = 2;
  const zones = 3;
  const minTotal = minInstancesPerAZ * zones;

  return Math.max(requiredInstances, minTotal);
}

console.log(calculateInstances(5000, 1000));  // 7 instances
```

### Auto-scaling Configuration

```yaml
# Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app
  minReplicas: 3
  maxReplicas: 20
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
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "1000"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 30
        - type: Pods
          value: 4
          periodSeconds: 30
      selectPolicy: Max
```

### AWS Auto Scaling

```json
{
  "AutoScalingGroupName": "app-asg",
  "MinSize": 3,
  "MaxSize": 20,
  "DesiredCapacity": 5,
  "TargetTrackingScalingPolicies": [
    {
      "TargetValue": 70.0,
      "PredefinedMetricSpecification": {
        "PredefinedMetricType": "ASGAverageCPUUtilization"
      },
      "ScaleInCooldown": 300,
      "ScaleOutCooldown": 60
    },
    {
      "TargetValue": 1000.0,
      "CustomizedMetricSpecification": {
        "MetricName": "RequestCountPerTarget",
        "Namespace": "AWS/ApplicationELB",
        "Statistic": "Sum"
      }
    }
  ]
}
```

## Performance Budgets

### Response Time Budget

```javascript
const performanceBudget = {
  // Page load budgets
  ttfb: 200,          // Time to First Byte (ms)
  fcp: 1000,          // First Contentful Paint (ms)
  lcp: 2500,          // Largest Contentful Paint (ms)

  // API budgets
  apiP50: 100,        // 50th percentile (ms)
  apiP95: 500,        // 95th percentile (ms)
  apiP99: 1000,       // 99th percentile (ms)

  // Resource budgets
  jsBundle: 200,      // JavaScript bundle size (KB)
  cssBundle: 50,      // CSS bundle size (KB)
  images: 500,        // Total images (KB)

  // Infrastructure budgets
  cpuUtilization: 70,     // Max % during normal load
  memoryUtilization: 80,  // Max % during normal load
  errorRate: 0.01,        // Max 1% error rate
};

function checkBudget(actual, budget, metric) {
  if (actual > budget) {
    console.warn(`Budget exceeded for ${metric}: ${actual} > ${budget}`);
    return false;
  }
  return true;
}
```

## Cost Optimization

### Instance Sizing

```javascript
function optimizeInstanceSize(workload) {
  const instances = [
    { type: 't3.small', vcpu: 2, memory: 2, cost: 0.0208 },
    { type: 't3.medium', vcpu: 2, memory: 4, cost: 0.0416 },
    { type: 't3.large', vcpu: 2, memory: 8, cost: 0.0832 },
    { type: 'm5.large', vcpu: 2, memory: 8, cost: 0.096 },
    { type: 'm5.xlarge', vcpu: 4, memory: 16, cost: 0.192 },
  ];

  const filtered = instances.filter(i =>
    i.vcpu >= workload.requiredVCPU &&
    i.memory >= workload.requiredMemory
  );

  // Sort by cost efficiency
  return filtered.sort((a, b) => {
    const scoreA = (a.vcpu * a.memory) / a.cost;
    const scoreB = (b.vcpu * b.memory) / b.cost;
    return scoreB - scoreA;
  })[0];
}

const recommendation = optimizeInstanceSize({
  requiredVCPU: 2,
  requiredMemory: 4,
});

console.log('Recommended instance:', recommendation);
```

## Capacity Alerts

```yaml
# Prometheus alerting rules
groups:
  - name: capacity
    rules:
      - alert: HighCPUPrediction
        expr: |
          predict_linear(
            node_cpu_seconds_total{mode="idle"}[1h],
            3600 * 24 * 7  # 7 days ahead
          ) < 0.2
        for: 1h
        annotations:
          summary: CPU capacity will be exhausted in 7 days

      - alert: DiskSpaceProjection
        expr: |
          predict_linear(
            node_filesystem_avail_bytes[7d],
            3600 * 24 * 30
          ) < 1e9  # Less than 1GB in 30 days
        annotations:
          summary: Disk space will run out in 30 days

      - alert: DatabaseConnectionsNearLimit
        expr: |
          pg_stat_database_numbackends / pg_settings_max_connections > 0.8
        for: 10m
        annotations:
          summary: Database connections at 80% capacity

      - alert: ScalingRecommendation
        expr: |
          rate(http_requests_total[5m]) >
          (instance_capacity * instance_count * 0.7)
        annotations:
          summary: Consider scaling up - traffic approaching capacity
```

## Quick Reference

| Metric | Buffer | Reasoning |
|--------|--------|-----------|
| CPU | 30% | Headroom for spikes |
| Memory | 20% | GC and OS overhead |
| Connections | 25% | Connection churn |
| Storage | 40% | Growth + snapshots |

| Planning Horizon | Update Frequency |
|------------------|------------------|
| 3 months | Weekly |
| 6 months | Bi-weekly |
| 12 months | Monthly |

| Scaling Trigger | Action |
|-----------------|--------|
| 70% CPU | Start planning |
| 80% CPU | Scale up |
| 90% CPU | Emergency scaling |
| 60% CPU for 24h | Scale down |

---

## Reference: dashboards

# Dashboards

## RED Method (Request-focused)

```
Rate     - Requests per second
Errors   - Failed requests per second
Duration - Response time distribution
```

```promql
# Rate
sum(rate(http_requests_total[5m]))

# Errors
sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
sum(rate(http_requests_total[5m]))

# Duration (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

## USE Method (Resource-focused)

```
Utilization - % time resource is busy
Saturation  - Queue depth, backlog
Errors      - Error events
```

```promql
# CPU Utilization
100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory Saturation
node_memory_SwapTotal_bytes - node_memory_SwapFree_bytes

# Disk Errors
rate(node_disk_io_time_weighted_seconds_total[5m])
```

## Dashboard Structure

```
┌─────────────────────────────────────────────────────────────┐
│                    SERVICE OVERVIEW                         │
│  Request Rate │ Error Rate │ p50 Latency │ p99 Latency     │
├─────────────────────────────────────────────────────────────┤
│                    REQUEST METRICS                          │
│  [Graph: Requests/s by endpoint]                           │
│  [Graph: Error rate over time]                             │
├─────────────────────────────────────────────────────────────┤
│                    LATENCY METRICS                          │
│  [Heatmap: Latency distribution]                           │
│  [Graph: p50, p95, p99 over time]                          │
├─────────────────────────────────────────────────────────────┤
│                    INFRASTRUCTURE                           │
│  CPU │ Memory │ Disk │ Network                             │
└─────────────────────────────────────────────────────────────┘
```

## Key Panels

### Stat Panel (Single Value)

```promql
# Current RPS
sum(rate(http_requests_total[5m]))

# Error percentage
sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
sum(rate(http_requests_total[5m])) * 100
```

### Time Series

```promql
# Requests by status
sum by (status) (rate(http_requests_total[5m]))

# Latency percentiles
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### Table

```promql
# Top endpoints by error rate
topk(10,
  sum by (path) (rate(http_requests_total{status=~"5.."}[5m]))
  /
  sum by (path) (rate(http_requests_total[5m]))
)
```

## Business Metrics Dashboard

```promql
# Orders per minute
sum(rate(orders_created_total[5m])) * 60

# Revenue (if tracked)
sum(increase(order_value_dollars_sum[1h]))

# Active users (gauge)
active_users_total
```

## Quick Reference

| Method | Focus | Metrics |
|--------|-------|---------|
| RED | Services | Rate, Errors, Duration |
| USE | Resources | Utilization, Saturation, Errors |

| Panel Type | Use Case |
|------------|----------|
| Stat | Single KPI |
| Time Series | Trends over time |
| Heatmap | Latency distribution |
| Table | Top N, details |
| Gauge | Current vs threshold |

---

## Reference: opentelemetry

# OpenTelemetry Tracing

## Node.js Setup

```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';

const sdk = new NodeSDK({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: 'my-service',
    [SemanticResourceAttributes.SERVICE_VERSION]: '1.0.0',
  }),
  traceExporter: new OTLPTraceExporter({
    url: 'http://jaeger:4318/v1/traces',
  }),
  instrumentations: [getNodeAutoInstrumentations()],
});

sdk.start();
```

## Manual Spans

```typescript
import { trace, SpanStatusCode } from '@opentelemetry/api';

const tracer = trace.getTracer('my-service');

async function processOrder(orderId: string) {
  return tracer.startActiveSpan('processOrder', async (span) => {
    span.setAttribute('order.id', orderId);

    try {
      // Child span for database
      await tracer.startActiveSpan('db.getOrder', async (dbSpan) => {
        const order = await db.orders.findById(orderId);
        dbSpan.setAttribute('db.rows_affected', 1);
        dbSpan.end();
        return order;
      });

      // Child span for external API
      await tracer.startActiveSpan('api.processPayment', async (apiSpan) => {
        await paymentService.process(order);
        apiSpan.end();
      });

      span.setStatus({ code: SpanStatusCode.OK });
    } catch (error) {
      span.setStatus({
        code: SpanStatusCode.ERROR,
        message: error.message,
      });
      span.recordException(error);
      throw error;
    } finally {
      span.end();
    }
  });
}
```

## Context Propagation

```typescript
import { propagation, context } from '@opentelemetry/api';

// Extract from incoming request
app.use((req, res, next) => {
  const ctx = propagation.extract(context.active(), req.headers);
  context.with(ctx, next);
});

// Inject into outgoing request
async function callExternalService() {
  const headers = {};
  propagation.inject(context.active(), headers);

  await fetch('http://other-service/api', { headers });
}
```

## Python Setup

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint="http://jaeger:4318/v1/traces"))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

tracer = trace.get_tracer(__name__)

def process_order(order_id: str):
    with tracer.start_as_current_span("process_order") as span:
        span.set_attribute("order.id", order_id)
        # ... process order
```

## Quick Reference

| Concept | Purpose |
|---------|---------|
| Span | Single operation |
| Trace | Full request flow |
| Context | Correlation across services |
| Attributes | Metadata on spans |
| Events | Timestamped logs in span |

| Attribute | Example |
|-----------|---------|
| `http.method` | GET, POST |
| `http.status_code` | 200, 500 |
| `db.system` | postgresql |
| `db.statement` | SELECT ... |

---

## Reference: performance-testing

# Performance Testing

## Load Testing with k6

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp-up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp-up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.1'],
  },
};

export default function () {
  const res = http.get('https://api.example.com/products');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  sleep(1);
}
```

## Test Types

### Load Test
```javascript
// Gradual ramp-up to expected production load
export const options = {
  stages: [
    { duration: '5m', target: 100 },
    { duration: '30m', target: 100 },
    { duration: '5m', target: 0 },
  ],
};
```

### Stress Test
```javascript
// Push beyond normal capacity to find breaking point
export const options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 200 },
    { duration: '5m', target: 300 },
    { duration: '5m', target: 400 },
    { duration: '2m', target: 0 },
  ],
};
```

### Spike Test
```javascript
// Sudden increase in load
export const options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '30s', target: 1000 }, // Spike
    { duration: '3m', target: 100 },
    { duration: '1m', target: 0 },
  ],
};
```

### Soak Test
```javascript
// Extended duration at normal load
export const options = {
  stages: [
    { duration: '5m', target: 100 },
    { duration: '8h', target: 100 },  // Long duration
    { duration: '5m', target: 0 },
  ],
};
```

## Artillery.io

```yaml
# load-test.yml
config:
  target: 'https://api.example.com'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 300
      arrivalRate: 50
      name: "Sustained load"

  processor: "./custom-functions.js"

  variables:
    userId:
      - "user1"
      - "user2"

scenarios:
  - name: "Product browsing"
    weight: 70
    flow:
      - get:
          url: "/products"
      - think: 2
      - get:
          url: "/products/{{ $randomNumber(1, 100) }}"

  - name: "Checkout"
    weight: 30
    flow:
      - post:
          url: "/cart"
          json:
            productId: "{{ $randomNumber(1, 100) }}"
      - post:
          url: "/checkout"
          json:
            userId: "{{ userId }}"
```

## Locust (Python)

```python
from locust import HttpUser, task, between

class WebsiteUser(HttpUser):
    wait_time = between(1, 3)

    @task(3)
    def view_products(self):
        self.client.get("/products")

    @task(1)
    def view_product(self):
        product_id = random.randint(1, 100)
        self.client.get(f"/products/{product_id}")

    @task(1)
    def create_order(self):
        self.client.post("/orders", json={
            "product_id": random.randint(1, 100),
            "quantity": random.randint(1, 5)
        })

    def on_start(self):
        # Login or setup
        self.client.post("/login", json={
            "username": "test",
            "password": "<test-password>"
        })
```

## JMeter Thread Groups

```xml
<!-- Basic HTTP Request -->
<ThreadGroup>
  <stringProp name="ThreadGroup.num_threads">100</stringProp>
  <stringProp name="ThreadGroup.ramp_time">60</stringProp>
  <stringProp name="ThreadGroup.duration">300</stringProp>
  <boolProp name="ThreadGroup.scheduler">true</boolProp>
</ThreadGroup>
```

## Performance Metrics to Track

```javascript
// k6 custom metrics
import { Counter, Trend, Gauge } from 'k6/metrics';

const checkoutDuration = new Trend('checkout_duration');
const cartSize = new Gauge('cart_size');
const orderCounter = new Counter('orders_created');

export default function () {
  const startTime = Date.now();

  const res = http.post('https://api.example.com/checkout', payload);

  checkoutDuration.add(Date.now() - startTime);
  orderCounter.add(1);
  cartSize.add(payload.items.length);
}
```

## Test Scenario Design

```javascript
// Realistic user journey
import { scenario } from 'k6/execution';

export const options = {
  scenarios: {
    browser_users: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5m', target: 100 },
        { duration: '10m', target: 100 },
      ],
      gracefulRampDown: '30s',
    },
    api_users: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '15m',
      preAllocatedVUs: 100,
    },
  },
};

export default function () {
  // Homepage
  http.get('https://example.com/');
  sleep(Math.random() * 3);

  // Search
  http.get('https://example.com/search?q=laptop');
  sleep(Math.random() * 5);

  // Product page
  http.get('https://example.com/products/123');
  sleep(Math.random() * 10);

  // Add to cart (30% conversion)
  if (Math.random() < 0.3) {
    http.post('https://example.com/cart', { productId: 123 });
  }
}
```

## Quick Reference

| Test Type | Purpose | Duration |
|-----------|---------|----------|
| Load | Normal capacity | 30m - 2h |
| Stress | Find limits | 1h - 4h |
| Spike | Sudden traffic | 15m - 30m |
| Soak | Memory leaks | 4h - 24h |

| Tool | Language | Best For |
|------|----------|----------|
| k6 | JavaScript | API testing, CI/CD |
| Artillery | YAML/JS | Simple scenarios |
| Locust | Python | Complex scenarios |
| JMeter | GUI/XML | Legacy systems |

| Metric | Target |
|--------|--------|
| p95 latency | < 500ms |
| p99 latency | < 1s |
| Error rate | < 1% |
| RPS | 10x normal |

---

## Reference: prometheus-metrics

# Prometheus Metrics

## Metric Types

```typescript
import { Registry, Counter, Histogram, Gauge, Summary } from 'prom-client';

const register = new Registry();

// Counter - cumulative, only increases
const httpRequests = new Counter({
  name: 'http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'path', 'status'],
  registers: [register],
});

// Histogram - distribution with buckets
const httpDuration = new Histogram({
  name: 'http_request_duration_seconds',
  help: 'HTTP request duration in seconds',
  labelNames: ['method', 'path'],
  buckets: [0.01, 0.05, 0.1, 0.5, 1, 5],
  registers: [register],
});

// Gauge - point-in-time value, can go up/down
const activeConnections = new Gauge({
  name: 'active_connections',
  help: 'Number of active connections',
  registers: [register],
});

// Summary - similar to histogram with percentiles
const responseSummary = new Summary({
  name: 'http_response_size_bytes',
  help: 'HTTP response size',
  percentiles: [0.5, 0.9, 0.99],
  registers: [register],
});
```

## HTTP Middleware

```typescript
app.use((req, res, next) => {
  const end = httpDuration.startTimer({
    method: req.method,
    path: req.route?.path || req.path,
  });

  res.on('finish', () => {
    httpRequests.inc({
      method: req.method,
      path: req.route?.path || req.path,
      status: res.statusCode,
    });
    end();
  });

  next();
});

// Metrics endpoint
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.send(await register.metrics());
});
```

## Business Metrics

```typescript
// Orders
const ordersCreated = new Counter({
  name: 'orders_created_total',
  help: 'Total orders created',
  labelNames: ['status', 'payment_method'],
});

const orderValue = new Histogram({
  name: 'order_value_dollars',
  help: 'Order value in dollars',
  buckets: [10, 50, 100, 500, 1000],
});

// Usage
ordersCreated.inc({ status: 'completed', payment_method: 'card' });
orderValue.observe(order.total);
```

## Default Metrics

```typescript
import { collectDefaultMetrics } from 'prom-client';

// Collect Node.js metrics (memory, CPU, etc.)
collectDefaultMetrics({ register });
```

## Python (prometheus_client)

```python
from prometheus_client import Counter, Histogram, Gauge, generate_latest

http_requests = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'path', 'status']
)

http_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration',
    ['method', 'path']
)

@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type="text/plain")
```

## Quick Reference

| Type | Use Case | Example |
|------|----------|---------|
| Counter | Cumulative totals | Requests, errors |
| Gauge | Current value | Active users, queue size |
| Histogram | Distributions | Response times |
| Summary | Percentiles | Similar to histogram |

| Naming | Convention |
|--------|------------|
| Unit suffix | `_seconds`, `_bytes`, `_total` |
| Base unit | Use seconds, bytes (not ms, KB) |
| Prefix | App/service name |

---

## Reference: structured-logging

# Structured Logging

## Pino (Node.js)

```typescript
import pino from 'pino';

const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  formatters: {
    level: (label) => ({ level: label }),
  },
  redact: ['password', 'token', 'authorization'],
});

// Structured logging
logger.info({
  event: 'user.login',
  userId: user.id,
  ip: req.ip,
  userAgent: req.headers['user-agent'],
  duration: Date.now() - start,
});

// Error logging with context
logger.error({
  event: 'payment.failed',
  error: err.message,
  stack: err.stack,
  orderId: order.id,
  amount: order.total,
  userId: user.id,
});
```

## Request Logging Middleware

```typescript
import { randomUUID } from 'crypto';

app.use((req, res, next) => {
  const requestId = req.headers['x-request-id'] || randomUUID();
  const start = Date.now();

  res.setHeader('x-request-id', requestId);

  res.on('finish', () => {
    logger.info({
      event: 'http.request',
      requestId,
      method: req.method,
      path: req.path,
      status: res.statusCode,
      duration: Date.now() - start,
      userAgent: req.headers['user-agent'],
      ip: req.ip,
    });
  });

  next();
});
```

## Python (structlog)

```python
import structlog

structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer()
    ],
)

logger = structlog.get_logger()

# Structured logging
logger.info(
    "user.login",
    user_id=user.id,
    ip=request.client.host,
    duration=elapsed_time,
)

# Error logging
logger.error(
    "payment.failed",
    error=str(exc),
    order_id=order.id,
    amount=order.total,
)
```

## Log Levels

| Level | Use Case |
|-------|----------|
| `error` | Failures needing attention |
| `warn` | Potential problems |
| `info` | Business events, requests |
| `debug` | Development details |
| `trace` | Verbose debugging |

## Best Practices

```typescript
// Good: Structured fields
logger.info({ event: 'order.created', orderId: '123', total: 99.99 });

// Bad: String interpolation
logger.info(`Order 123 created with total 99.99`);

// Good: Consistent event names
logger.info({ event: 'user.registered' });
logger.info({ event: 'user.login' });
logger.info({ event: 'user.logout' });

// Good: Include correlation ID
logger.info({ event: 'request.processed', requestId, userId });
```

## Quick Reference

| Field | Purpose |
|-------|---------|
| `event` | Event name |
| `requestId` | Correlation ID |
| `userId` | User context |
| `duration` | Timing info |
| `error` / `stack` | Error details |
| `timestamp` | When (auto-added) |

| Library | Language |
|---------|----------|
| pino | Node.js |
| structlog | Python |
| slog | Go |
| logrus | Go |
