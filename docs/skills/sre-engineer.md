# sre-engineer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/sre-engineer (MIT). Рабочая копия скилла: `.claude/skills/sre-engineer/`.

---
name: sre-engineer
description: Defines service level objectives, creates error budget policies, designs incident response procedures, develops capacity models, and produces monitoring configurations and automation scripts for production systems. Use when defining SLIs/SLOs, managing error budgets, building reliable systems at scale, incident management, chaos engineering, toil reduction, or capacity planning.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: devops
  triggers: SRE, site reliability, SLO, SLI, error budget, incident management, chaos engineering, toil reduction, on-call, MTTR
  role: specialist
  scope: implementation
  output-format: code
  related-skills: devops-engineer, cloud-architect, kubernetes-specialist
---

# SRE Engineer

## Core Workflow

1. **Assess reliability** - Review architecture, SLOs, incidents, toil levels
2. **Define SLOs** - Identify meaningful SLIs and set appropriate targets
3. **Verify alignment** - Confirm SLO targets reflect user expectations before proceeding
4. **Implement monitoring** - Build golden signal dashboards and alerting
5. **Automate toil** - Identify repetitive tasks and build automation
6. **Test resilience** - Design and execute chaos experiments; verify recovery meets RTO/RPO targets before marking the experiment complete; validate recovery behavior end-to-end

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| SLO/SLI | `references/slo-sli-management.md` | Defining SLOs, calculating error budgets |
| Error Budgets | `references/error-budget-policy.md` | Managing budgets, burn rates, policies |
| Monitoring | `references/monitoring-alerting.md` | Golden signals, alert design, dashboards |
| Automation | `references/automation-toil.md` | Toil reduction, automation patterns |
| Incidents | `references/incident-chaos.md` | Incident response, chaos engineering |

## Constraints

### MUST DO
- Define quantitative SLOs (e.g., 99.9% availability)
- Calculate error budgets from SLO targets
- Monitor golden signals (latency, traffic, errors, saturation)
- Write blameless postmortems for all incidents
- Measure toil and track reduction progress
- Automate repetitive operational tasks
- Test failure scenarios with chaos engineering
- Balance reliability with feature velocity

### MUST NOT DO
- Set SLOs without user impact justification
- Alert on symptoms without actionable runbooks
- Tolerate >50% toil without automation plan
- Skip postmortems or assign blame
- Implement manual processes for recurring tasks
- Deploy without capacity planning
- Ignore error budget exhaustion
- Build systems that can't degrade gracefully

## Output Templates

When implementing SRE practices, provide:
1. SLO definitions with SLI measurements and targets
2. Monitoring/alerting configuration (Prometheus, etc.)
3. Automation scripts (Python, Go, Terraform)
4. Runbooks with clear remediation steps
5. Brief explanation of reliability impact

## Concrete Examples

### SLO Definition & Error Budget Calculation

```
# 99.9% availability SLO over a 30-day window
# Allowed downtime: (1 - 0.999) * 30 * 24 * 60 = 43.2 minutes/month
# Error budget (request-based): 0.001 * total_requests

# Example: 10M requests/month → 10,000 error budget requests
# If 5,000 errors consumed in week 1 → 50% budget burned in 25% of window
# → Trigger error budget policy: freeze non-critical releases
```

### Prometheus SLO Alerting Rule (Multiwindow Burn Rate)

```yaml
groups:
  - name: slo_availability
    rules:
      # Fast burn: 2% budget in 1h (14.4x burn rate)
      - alert: HighErrorBudgetBurn
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[1h]))
            /
            sum(rate(http_requests_total[1h]))
          ) > 0.014400
          and
          (
            sum(rate(http_requests_total{status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total[5m]))
          ) > 0.014400
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error budget burn rate detected"
          runbook: "https://wiki.internal/runbooks/high-error-burn"

      # Slow burn: 5% budget in 6h (1x burn rate sustained)
      - alert: SlowErrorBudgetBurn
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[6h]))
            /
            sum(rate(http_requests_total[6h]))
          ) > 0.001
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Sustained error budget consumption"
          runbook: "https://wiki.internal/runbooks/slow-error-burn"
```

### PromQL Golden Signal Queries

```promql
# Latency — 99th percentile request duration
histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))

# Traffic — requests per second by service
sum(rate(http_requests_total[5m])) by (service)

# Errors — error rate ratio
sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
  /
sum(rate(http_requests_total[5m])) by (service)

# Saturation — CPU throttling ratio
sum(rate(container_cpu_cfs_throttled_seconds_total[5m])) by (pod)
  /
sum(rate(container_cpu_cfs_periods_total[5m])) by (pod)
```

### Toil Automation Script (Python)

```python
#!/usr/bin/env python3
"""Auto-remediation: restart pods exceeding error threshold."""
import subprocess, sys, json

ERROR_THRESHOLD = 0.05  # 5% error rate triggers restart

def get_error_rate(service: str) -> float:
    """Query Prometheus for current error rate."""
    import urllib.request
    query = f'sum(rate(http_requests_total{{status=~"5..",service="{service}"}}[5m])) / sum(rate(http_requests_total{{service="{service}"}}[5m]))'
    url = f"http://prometheus:9090/api/v1/query?query={urllib.request.quote(query)}"
    with urllib.request.urlopen(url) as resp:
        data = json.load(resp)
    results = data["data"]["result"]
    return float(results[0]["value"][1]) if results else 0.0

def restart_deployment(namespace: str, deployment: str) -> None:
    subprocess.run(
        ["kubectl", "rollout", "restart", f"deployment/{deployment}", "-n", namespace],
        check=True
    )
    print(f"Restarted {namespace}/{deployment}")

if __name__ == "__main__":
    service, namespace, deployment = sys.argv[1], sys.argv[2], sys.argv[3]
    rate = get_error_rate(service)
    print(f"Error rate for {service}: {rate:.2%}")
    if rate > ERROR_THRESHOLD:
        restart_deployment(namespace, deployment)
    else:
        print("Within SLO threshold — no action required")
```

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/devops/sre-engineer/)

---

## Reference: automation-toil

# Automation and Toil Reduction

## Toil Definition

Toil is manual, repetitive, automatable work that scales linearly with service growth.

```python
from dataclasses import dataclass
from enum import Enum

class ToilCategory(Enum):
    """Categories of operational toil."""
    MANUAL_INTERVENTION = "manual"
    REPETITIVE_TASKS = "repetitive"
    NO_ENDURING_VALUE = "no_value"
    SCALES_WITH_SERVICE = "scales"
    INTERRUPT_DRIVEN = "reactive"

@dataclass
class ToilItem:
    """Track a specific toil activity."""
    name: str
    frequency_per_week: int
    minutes_per_occurrence: int
    category: ToilCategory
    automation_difficulty: str  # 'easy', 'medium', 'hard'

    @property
    def weekly_hours(self) -> float:
        """Calculate weekly hours spent on this toil."""
        return (self.frequency_per_week * self.minutes_per_occurrence) / 60

    @property
    def annual_hours(self) -> float:
        """Calculate annual hours spent on this toil."""
        return self.weekly_hours * 52

    def roi_score(self) -> float:
        """Calculate ROI score for automation (higher = better).

        Score considers time saved vs. difficulty.
        """
        difficulty_multiplier = {
            'easy': 1.0,
            'medium': 0.5,
            'hard': 0.25,
        }
        return self.annual_hours * difficulty_multiplier.get(
            self.automation_difficulty, 0.1
        )

# Example toil inventory
toil_items = [
    ToilItem(
        name="Manual database failover",
        frequency_per_week=2,
        minutes_per_occurrence=30,
        category=ToilCategory.MANUAL_INTERVENTION,
        automation_difficulty='medium',
    ),
    ToilItem(
        name="Restarting hung processes",
        frequency_per_week=5,
        minutes_per_occurrence=15,
        category=ToilCategory.REPETITIVE_TASKS,
        automation_difficulty='easy',
    ),
    ToilItem(
        name="Log file cleanup",
        frequency_per_week=7,
        minutes_per_occurrence=10,
        category=ToilCategory.SCALES_WITH_SERVICE,
        automation_difficulty='easy',
    ),
]

# Calculate total toil and prioritize automation
total_weekly_hours = sum(item.weekly_hours for item in toil_items)
print(f"Total weekly toil: {total_weekly_hours:.1f} hours")

# Sort by ROI score to prioritize automation
sorted_items = sorted(toil_items, key=lambda x: x.roi_score(), reverse=True)
for item in sorted_items:
    print(f"{item.name}: {item.roi_score():.1f} ROI score")
```

## Self-Healing Systems

Automate common failure remediation.

```python
# auto_healing.py - Self-healing automation examples
import subprocess
import logging
from typing import Callable, Dict
from dataclasses import dataclass

logger = logging.getLogger(__name__)

@dataclass
class HealthCheck:
    """Define a health check and remediation."""
    name: str
    check: Callable[[], bool]
    remediate: Callable[[], bool]
    max_retries: int = 3

class SelfHealer:
    """Automatically remediate common failures."""

    def __init__(self):
        self.health_checks: Dict[str, HealthCheck] = {}

    def register(self, check: HealthCheck):
        """Register a health check with remediation."""
        self.health_checks[check.name] = check

    def run(self):
        """Run all health checks and remediate failures."""
        for name, check in self.health_checks.items():
            if not check.check():
                logger.warning(f"Health check failed: {name}")
                self._remediate(check)

    def _remediate(self, check: HealthCheck):
        """Attempt remediation with retries."""
        for attempt in range(check.max_retries):
            logger.info(f"Remediation attempt {attempt + 1}/{check.max_retries}")

            if check.remediate():
                logger.info(f"Remediation successful: {check.name}")
                return

            if check.check():
                logger.info(f"Health check passed after remediation: {check.name}")
                return

        logger.error(f"Remediation failed after {check.max_retries} attempts")
        self._escalate(check)

    def _escalate(self, check: HealthCheck):
        """Escalate to on-call when auto-remediation fails."""
        # Send alert to on-call
        logger.error(f"ESCALATING: {check.name} - auto-remediation failed")

# Example health checks
def check_disk_space() -> bool:
    """Check if disk space is above 20%."""
    result = subprocess.run(
        ["df", "-h", "/"],
        capture_output=True,
        text=True
    )
    # Parse df output and check available space
    lines = result.stdout.strip().split('\n')
    if len(lines) > 1:
        fields = lines[1].split()
        use_percent = int(fields[4].rstrip('%'))
        return use_percent < 80
    return True

def cleanup_disk() -> bool:
    """Clean up old log files."""
    try:
        # Delete logs older than 7 days
        subprocess.run(
            ["find", "/var/log", "-name", "*.log", "-mtime", "+7", "-delete"],
            check=True
        )
        return True
    except subprocess.CalledProcessError:
        return False

def check_service_responsive() -> bool:
    """Check if service responds to health endpoint."""
    try:
        result = subprocess.run(
            ["curl", "-f", "http://localhost:8080/health"],
            capture_output=True,
            timeout=5
        )
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        return False

def restart_service() -> bool:
    """Restart the service."""
    try:
        subprocess.run(
            ["systemctl", "restart", "myservice"],
            check=True
        )
        return True
    except subprocess.CalledProcessError:
        return False

# Set up self-healing
healer = SelfHealer()
healer.register(HealthCheck(
    name="disk_space",
    check=check_disk_space,
    remediate=cleanup_disk,
))
healer.register(HealthCheck(
    name="service_health",
    check=check_service_responsive,
    remediate=restart_service,
))

# Run as cron job or systemd timer
if __name__ == "__main__":
    healer.run()
```

## Runbook Automation

Convert manual runbooks to automated scripts.

```python
# runbook_automation.py
from typing import List, Tuple
from dataclasses import dataclass
import subprocess
import json

@dataclass
class RunbookStep:
    """A single step in a runbook."""
    description: str
    command: str
    critical: bool = True  # Stop on failure?
    verify: str | None = None  # Optional verification command

class AutomatedRunbook:
    """Execute runbook steps automatically."""

    def __init__(self, name: str):
        self.name = name
        self.steps: List[RunbookStep] = []

    def add_step(self, step: RunbookStep):
        """Add a step to the runbook."""
        self.steps.append(step)

    def execute(self, dry_run: bool = False) -> Tuple[bool, List[str]]:
        """Execute all runbook steps.

        Args:
            dry_run: If True, only print commands without executing

        Returns:
            tuple: (success, output_lines)
        """
        outputs = []

        for i, step in enumerate(self.steps, 1):
            outputs.append(f"\n[Step {i}/{len(self.steps)}] {step.description}")

            if dry_run:
                outputs.append(f"Would run: {step.command}")
                continue

            # Execute command
            try:
                result = subprocess.run(
                    step.command,
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=300,
                )

                if result.returncode != 0:
                    outputs.append(f"ERROR: {result.stderr}")
                    if step.critical:
                        return False, outputs
                else:
                    outputs.append(result.stdout)

                # Run verification if specified
                if step.verify:
                    verify_result = subprocess.run(
                        step.verify,
                        shell=True,
                        capture_output=True,
                        text=True,
                    )
                    if verify_result.returncode != 0:
                        outputs.append(f"VERIFICATION FAILED: {verify_result.stderr}")
                        if step.critical:
                            return False, outputs

            except subprocess.TimeoutExpired:
                outputs.append(f"ERROR: Command timed out")
                if step.critical:
                    return False, outputs

        return True, outputs

# Example: Database failover runbook
failover_runbook = AutomatedRunbook("Database Failover")

failover_runbook.add_step(RunbookStep(
    description="Stop writes to primary database",
    command="kubectl exec -it postgres-primary-0 -- psql -c 'ALTER SYSTEM SET default_transaction_read_only = on;'",
    critical=True,
))

failover_runbook.add_step(RunbookStep(
    description="Wait for replication lag to clear",
    command="sleep 10",
    critical=False,
))

failover_runbook.add_step(RunbookStep(
    description="Promote replica to primary",
    command="kubectl exec -it postgres-replica-0 -- pg_ctl promote",
    critical=True,
    verify="kubectl exec -it postgres-replica-0 -- psql -c 'SELECT pg_is_in_recovery();' | grep -q 'f'",
))

failover_runbook.add_step(RunbookStep(
    description="Update service to point to new primary",
    command="kubectl patch service postgres -p '{\"spec\":{\"selector\":{\"role\":\"replica\"}}}'",
    critical=True,
))

# Execute
success, output = failover_runbook.execute(dry_run=False)
print('\n'.join(output))
```

## Capacity Planning Automation

```python
# capacity_planner.py - Automated capacity planning
from dataclasses import dataclass
from datetime import datetime, timedelta
import numpy as np

@dataclass
class CapacityMetrics:
    """Historical capacity metrics."""
    timestamp: datetime
    requests_per_second: float
    cpu_utilization: float
    memory_utilization: float

class CapacityPlanner:
    """Automated capacity planning and forecasting."""

    def __init__(self, metrics: list[CapacityMetrics]):
        self.metrics = metrics

    def forecast_growth(self, days_ahead: int = 90) -> dict:
        """Forecast resource usage growth.

        Uses linear regression on historical data.
        """
        # Extract time series
        timestamps = [(m.timestamp - self.metrics[0].timestamp).days
                      for m in self.metrics]
        cpu_values = [m.cpu_utilization for m in self.metrics]
        mem_values = [m.memory_utilization for m in self.metrics]

        # Fit linear trend
        cpu_trend = np.polyfit(timestamps, cpu_values, deg=1)
        mem_trend = np.polyfit(timestamps, mem_values, deg=1)

        # Forecast
        future_day = timestamps[-1] + days_ahead
        cpu_forecast = np.polyval(cpu_trend, future_day)
        mem_forecast = np.polyval(mem_trend, future_day)

        return {
            'days_ahead': days_ahead,
            'cpu_forecast': min(cpu_forecast, 1.0),
            'memory_forecast': min(mem_forecast, 1.0),
            'cpu_threshold_breach': cpu_forecast > 0.8,
            'memory_threshold_breach': mem_forecast > 0.8,
        }

    def recommend_scaling(self, forecast: dict) -> str:
        """Recommend scaling action based on forecast."""
        if forecast['cpu_threshold_breach'] or forecast['memory_threshold_breach']:
            return f"SCALE UP: Forecast shows >80% utilization in {forecast['days_ahead']} days"

        return "OK: No scaling needed"

# Example usage
historical_metrics = [
    CapacityMetrics(
        timestamp=datetime.now() - timedelta(days=30),
        requests_per_second=1000,
        cpu_utilization=0.45,
        memory_utilization=0.50,
    ),
    CapacityMetrics(
        timestamp=datetime.now() - timedelta(days=15),
        requests_per_second=1200,
        cpu_utilization=0.55,
        memory_utilization=0.60,
    ),
    CapacityMetrics(
        timestamp=datetime.now(),
        requests_per_second=1500,
        cpu_utilization=0.65,
        memory_utilization=0.70,
    ),
]

planner = CapacityPlanner(historical_metrics)
forecast = planner.forecast_growth(days_ahead=90)
recommendation = planner.recommend_scaling(forecast)

print(f"90-day forecast: CPU={forecast['cpu_forecast']:.1%}, Memory={forecast['memory_forecast']:.1%}")
print(recommendation)
```

## Automation Testing

```python
# test_automation.py - Test automation scripts before production
import unittest
from unittest.mock import patch, MagicMock

class TestSelfHealing(unittest.TestCase):
    """Test self-healing automation."""

    @patch('subprocess.run')
    def test_disk_cleanup_success(self, mock_run):
        """Test successful disk cleanup."""
        mock_run.return_value = MagicMock(returncode=0)

        result = cleanup_disk()

        self.assertTrue(result)
        mock_run.assert_called_once()

    @patch('subprocess.run')
    def test_service_restart_with_retry(self, mock_run):
        """Test service restart retries on failure."""
        # First attempt fails, second succeeds
        mock_run.side_effect = [
            MagicMock(returncode=1),  # First restart fails
            MagicMock(returncode=0),  # Second restart succeeds
        ]

        # Implementation would retry on failure
        # Assert retry logic works correctly
```

## Toil Reduction Metrics

```python
# Track toil reduction progress
class ToilTracker:
    """Track toil reduction over time."""

    def __init__(self):
        self.snapshots = []

    def record_snapshot(self, week: int, toil_hours: float, team_hours: float):
        """Record toil snapshot for a week."""
        self.snapshots.append({
            'week': week,
            'toil_hours': toil_hours,
            'team_hours': team_hours,
            'toil_percentage': (toil_hours / team_hours * 100) if team_hours > 0 else 0,
        })

    def toil_trend(self) -> str:
        """Calculate if toil is increasing or decreasing."""
        if len(self.snapshots) < 2:
            return "insufficient data"

        first_pct = self.snapshots[0]['toil_percentage']
        last_pct = self.snapshots[-1]['toil_percentage']

        if last_pct < first_pct:
            return f"improving ({first_pct:.1f}% → {last_pct:.1f}%)"
        else:
            return f"worsening ({first_pct:.1f}% → {last_pct:.1f}%)"

# Target: <50% toil, ideally <30%
tracker = ToilTracker()
tracker.record_snapshot(week=1, toil_hours=30, team_hours=40)  # 75% toil
tracker.record_snapshot(week=4, toil_hours=20, team_hours=40)  # 50% toil
tracker.record_snapshot(week=8, toil_hours=12, team_hours=40)  # 30% toil

print(f"Toil trend: {tracker.toil_trend()}")
```

---

## Reference: error-budget-policy

# Error Budget Policy

## Error Budget Fundamentals

Error budget = 1 - SLO target. It represents acceptable unreliability.

```python
from dataclasses import dataclass
from datetime import datetime, timedelta
from enum import Enum

class BudgetStatus(Enum):
    """Error budget health status."""
    HEALTHY = "healthy"      # >75% budget remaining
    WARNING = "warning"      # 25-75% budget remaining
    CRITICAL = "critical"    # <25% budget remaining
    EXHAUSTED = "exhausted"  # 0% budget remaining

@dataclass
class ErrorBudget:
    """Error budget tracker."""
    slo_target: float  # e.g., 0.999
    window_days: int   # e.g., 30

    @property
    def budget_percentage(self) -> float:
        """Total error budget as percentage."""
        return (1 - self.slo_target) * 100

    @property
    def allowed_downtime(self) -> timedelta:
        """Maximum allowed downtime in window."""
        total_minutes = self.window_days * 24 * 60
        error_minutes = total_minutes * (1 - self.slo_target)
        return timedelta(minutes=error_minutes)

    def remaining_budget(self, actual_sli: float) -> float:
        """Calculate remaining error budget percentage.

        Returns:
            float: 0.0 to 1.0, where 1.0 = 100% budget remaining
        """
        budget_used = 1 - actual_sli
        total_budget = 1 - self.slo_target

        if total_budget == 0:
            return 0.0

        return max(0.0, 1 - (budget_used / total_budget))

    def get_status(self, actual_sli: float) -> BudgetStatus:
        """Determine budget health status."""
        remaining = self.remaining_budget(actual_sli)

        if remaining <= 0:
            return BudgetStatus.EXHAUSTED
        elif remaining < 0.25:
            return BudgetStatus.CRITICAL
        elif remaining < 0.75:
            return BudgetStatus.WARNING
        else:
            return BudgetStatus.HEALTHY

# Example
budget = ErrorBudget(slo_target=0.999, window_days=30)
print(f"Error budget: {budget.budget_percentage}%")
print(f"Allowed downtime: {budget.allowed_downtime}")
# Output:
# Error budget: 0.1%
# Allowed downtime: 43.2 minutes
```

## Burn Rate Alerting

Burn rate measures how fast you're consuming error budget.

```python
from typing import NamedTuple

class BurnRateAlert(NamedTuple):
    """Multi-window burn rate alert configuration."""
    window: timedelta
    burn_rate_threshold: float
    budget_consumed_threshold: float

    def should_alert(
        self,
        current_error_rate: float,
        total_budget: float
    ) -> bool:
        """Check if burn rate exceeds threshold.

        Args:
            current_error_rate: Current error rate (1 - SLI)
            total_budget: Total error budget (1 - SLO)

        Returns:
            bool: True if should alert
        """
        if total_budget == 0:
            return current_error_rate > 0

        burn_rate = current_error_rate / total_budget
        return burn_rate >= self.burn_rate_threshold

# Multi-window burn rate alerts (from Google SRE Workbook)
BURN_RATE_ALERTS = [
    # Fast burn: 2% budget in 1 hour = exhausted in 2 days
    BurnRateAlert(
        window=timedelta(hours=1),
        burn_rate_threshold=14.4,  # 2% of 30d budget in 1h
        budget_consumed_threshold=0.02
    ),
    # Medium burn: 5% budget in 6 hours
    BurnRateAlert(
        window=timedelta(hours=6),
        burn_rate_threshold=6.0,
        budget_consumed_threshold=0.05
    ),
    # Slow burn: 10% budget in 3 days
    BurnRateAlert(
        window=timedelta(days=3),
        burn_rate_threshold=1.0,
        budget_consumed_threshold=0.10
    ),
]

def check_burn_rate_alerts(slo_target: float, current_sli: float):
    """Check if any burn rate alerts should fire."""
    error_budget = 1 - slo_target
    error_rate = 1 - current_sli

    alerts = []
    for alert_config in BURN_RATE_ALERTS:
        if alert_config.should_alert(error_rate, error_budget):
            alerts.append(alert_config)

    return alerts
```

## Error Budget Policy Template

```yaml
# error_budget_policy.yaml
service: payment-api
slo:
  target: 99.9%
  measurement_window: 30 days

policy:
  # Actions based on remaining error budget
  actions:
    - threshold: 100%  # Budget healthy
      state: normal_operations
      actions:
        - "Continue feature development"
        - "Deploy during business hours"
        - "Standard change review process"

    - threshold: 50%   # Budget warning
      state: careful_operations
      actions:
        - "Increase code review rigor"
        - "Require senior engineer approval for deploys"
        - "Conduct pre-deployment risk assessment"
        - "Enhanced monitoring during deploys"

    - threshold: 25%   # Budget critical
      state: restricted_operations
      actions:
        - "Halt non-critical feature work"
        - "Focus on reliability improvements"
        - "Require VP approval for deployments"
        - "Deploy only critical bug fixes"
        - "Daily error budget review meetings"

    - threshold: 0%    # Budget exhausted
      state: feature_freeze
      actions:
        - "Immediate feature freeze"
        - "Deploy only emergency fixes"
        - "All hands reliability review"
        - "Mandatory postmortem for all incidents"
        - "Weekly executive review until recovered"

  # Exceptions to policy
  exceptions:
    - type: security_patch
      approval: security_team
      allowed: true

    - type: critical_business_requirement
      approval: vp_engineering + product_lead
      allowed: true
      requires_review: true
```

## Error Budget Calculation

```python
class ErrorBudgetCalculator:
    """Calculate and track error budget consumption."""

    def __init__(self, slo_target: float, window_days: int = 30):
        self.slo_target = slo_target
        self.window_days = window_days
        self.total_budget = 1 - slo_target

    def calculate_budget_status(
        self,
        good_events: int,
        total_events: int
    ) -> dict:
        """Calculate comprehensive budget status.

        Returns:
            dict: Budget status including remaining budget, burn rate, etc.
        """
        if total_events == 0:
            sli = 1.0
        else:
            sli = good_events / total_events

        budget_used = 1 - sli
        budget_remaining = self.total_budget - budget_used
        budget_remaining_pct = (
            (budget_remaining / self.total_budget * 100)
            if self.total_budget > 0 else 0
        )

        # Calculate burn rate
        burn_rate = budget_used / self.total_budget if self.total_budget > 0 else 0

        # Estimate time to exhaustion at current rate
        if burn_rate > 0:
            days_to_exhaustion = budget_remaining / budget_used * self.window_days
        else:
            days_to_exhaustion = float('inf')

        return {
            'sli': sli,
            'slo_target': self.slo_target,
            'compliant': sli >= self.slo_target,
            'budget_total': self.total_budget,
            'budget_used': budget_used,
            'budget_remaining': budget_remaining,
            'budget_remaining_pct': budget_remaining_pct,
            'burn_rate': burn_rate,
            'days_to_exhaustion': days_to_exhaustion,
            'good_events': good_events,
            'total_events': total_events,
        }

# Example usage
calc = ErrorBudgetCalculator(slo_target=0.999, window_days=30)
status = calc.calculate_budget_status(
    good_events=999_500,
    total_events=1_000_000
)

print(f"SLI: {status['sli']:.4f}")
print(f"Compliant: {status['compliant']}")
print(f"Budget remaining: {status['budget_remaining_pct']:.1f}%")
print(f"Days to exhaustion: {status['days_to_exhaustion']:.1f}")
```

## Prometheus Queries for Error Budgets

```promql
# Calculate 30-day availability SLI
sum(rate(http_requests_total{status=~"2..", job="api"}[30d]))
/
sum(rate(http_requests_total{job="api"}[30d]))

# Calculate error budget consumption
1 - (
  sum(rate(http_requests_total{status=~"2..", job="api"}[30d]))
  /
  sum(rate(http_requests_total{job="api"}[30d]))
)

# Calculate remaining error budget (for 99.9% SLO)
(0.001 - (1 - (
  sum(rate(http_requests_total{status=~"2..", job="api"}[30d]))
  /
  sum(rate(http_requests_total{job="api"}[30d]))
))) / 0.001

# Burn rate (normalized to 1.0 = sustainable)
(1 - (
  sum(rate(http_requests_total{status=~"2..", job="api"}[1h]))
  /
  sum(rate(http_requests_total{job="api"}[1h]))
)) / 0.001
```

## Decision Framework

Use this framework to make reliability vs. velocity tradeoffs:

```python
def should_deploy(
    budget_remaining: float,
    change_risk: str,  # 'low', 'medium', 'high'
    business_priority: str,  # 'low', 'medium', 'high', 'critical'
) -> tuple[bool, str]:
    """Decide if deployment should proceed.

    Returns:
        tuple: (should_deploy, reason)
    """
    # Budget exhausted - only critical changes
    if budget_remaining <= 0:
        if business_priority == 'critical':
            return True, "Critical business need, budget exhausted"
        return False, "Error budget exhausted, feature freeze in effect"

    # Budget critical (<25%)
    if budget_remaining < 0.25:
        if change_risk == 'high':
            return False, "High risk change with critical budget"
        if business_priority in ['high', 'critical']:
            return True, "High priority with critical budget - proceed carefully"
        return False, "Budget critical, deferring non-essential changes"

    # Budget warning (25-75%)
    if budget_remaining < 0.75:
        if change_risk == 'high' and business_priority == 'low':
            return False, "High risk, low priority with warning budget"
        return True, "Approved with enhanced review"

    # Budget healthy (>75%)
    return True, "Normal operations, budget healthy"
```

---

## Reference: incident-chaos

# Incident Management and Chaos Engineering

## Incident Response Framework

Structured approach to incident management.

```python
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import List

class Severity(Enum):
    """Incident severity levels."""
    SEV1 = "critical"     # Complete outage, major customer impact
    SEV2 = "high"         # Partial outage, significant impact
    SEV3 = "medium"       # Degraded performance, some users affected
    SEV4 = "low"          # Minor issue, minimal impact

@dataclass
class Incident:
    """Incident tracking."""
    id: str
    title: str
    severity: Severity
    started_at: datetime
    detected_at: datetime
    resolved_at: datetime | None = None
    root_cause: str | None = None
    impact: str | None = None

    @property
    def detection_time(self) -> float:
        """Time from start to detection in minutes."""
        delta = self.detected_at - self.started_at
        return delta.total_seconds() / 60

    @property
    def mttr(self) -> float | None:
        """Mean Time To Repair in minutes."""
        if not self.resolved_at:
            return None
        delta = self.resolved_at - self.detected_at
        return delta.total_seconds() / 60

    @property
    def total_duration(self) -> float | None:
        """Total incident duration in minutes."""
        if not self.resolved_at:
            return None
        delta = self.resolved_at - self.started_at
        return delta.total_seconds() / 60

# Example incident
incident = Incident(
    id="INC-2024-001",
    title="Database connection pool exhaustion",
    severity=Severity.SEV2,
    started_at=datetime(2024, 1, 15, 14, 30),
    detected_at=datetime(2024, 1, 15, 14, 35),
    resolved_at=datetime(2024, 1, 15, 15, 10),
    root_cause="Connection leak in payment service",
    impact="Payment processing delayed for 15% of users"
)

print(f"Detection time: {incident.detection_time:.1f} minutes")
print(f"MTTR: {incident.mttr:.1f} minutes")
print(f"Total duration: {incident.total_duration:.1f} minutes")
```

## Incident Response Runbook

```yaml
# incident_response.yaml
incident_response:
  detection:
    - "Acknowledge alert in PagerDuty"
    - "Join #incident-response Slack channel"
    - "Create incident doc from template"
    - "Assess severity (SEV1-4)"

  sev1_response:  # Critical - all hands
    - "Page on-call lead + backup"
    - "Notify VP Engineering immediately"
    - "Start Zoom war room"
    - "Assign incident commander"
    - "Assign communication lead"
    - "Post status update every 15 minutes"

  sev2_response:  # High - team response
    - "Page on-call engineer"
    - "Notify team lead"
    - "Create incident channel"
    - "Post status update every 30 minutes"

  roles:
    incident_commander:
      - "Coordinate response efforts"
      - "Make decisions quickly"
      - "Delegate tasks"
      - "Communicate with stakeholders"

    communication_lead:
      - "Post regular status updates"
      - "Notify affected customers"
      - "Update status page"
      - "Summarize timeline"

    on_call_engineer:
      - "Investigate root cause"
      - "Implement fixes"
      - "Verify resolution"
      - "Document actions taken"

  resolution:
    - "Verify metrics returned to normal"
    - "Monitor for 30 minutes"
    - "Post final status update"
    - "Schedule postmortem within 48 hours"
    - "Close incident"
```

## Blameless Postmortem Template

```markdown
# Postmortem: [Incident Title]

**Date:** 2024-01-15
**Authors:** [Names]
**Status:** Complete
**Severity:** SEV2

## Summary

One-paragraph summary of what happened, impact, and resolution.

## Impact

- **Duration:** 40 minutes (14:30 - 15:10 UTC)
- **Users affected:** ~15% of payment transactions
- **Revenue impact:** Estimated $X delayed
- **SLO impact:** Consumed 2.3% of monthly error budget

## Timeline (all times UTC)

| Time  | Event |
|-------|-------|
| 14:30 | Deployment of payment-service v2.3.0 completed |
| 14:32 | Error rate begins increasing |
| 14:35 | Alert fires: HighErrorRate |
| 14:36 | On-call engineer acknowledges |
| 14:40 | Incident declared SEV2 |
| 14:45 | Root cause identified: connection leak |
| 14:50 | Rollback initiated |
| 14:55 | Rollback completed |
| 15:00 | Error rate returns to normal |
| 15:10 | Incident resolved, monitoring continued |

## Root Cause

The payment-service v2.3.0 deployment introduced a connection leak in the
database connection pool. The new retry logic was not properly closing
connections on timeout, causing the pool to exhaust after ~20 minutes.

## Resolution

Rolled back to payment-service v2.2.1, which immediately resolved the issue.

## Detection

**What went well:**
- Alert fired within 5 minutes of issue start
- Clear runbook helped quick diagnosis

**What could be improved:**
- Could have caught in staging with longer load test
- Database connection pool metrics not monitored

## Action Items

| Action | Owner | Priority | Due Date |
|--------|-------|----------|----------|
| Add connection pool monitoring | @alice | P0 | 2024-01-20 |
| Extend staging load tests to 30min | @bob | P1 | 2024-01-25 |
| Review all resource cleanup in retry logic | @charlie | P1 | 2024-01-30 |
| Add integration test for connection leaks | @dave | P2 | 2024-02-05 |

## Lessons Learned

**What went well:**
- Quick detection and response
- Effective team communication
- Clear rollback procedure

**What didn't go well:**
- Issue not caught in pre-production testing
- No monitoring for connection pool exhaustion

**Where we got lucky:**
- Issue occurred during low-traffic period
- Only affected payment service, not critical systems
```

## Chaos Engineering

Proactively test system resilience through controlled failure injection.

```python
# chaos_experiment.py
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Callable

class ExperimentStatus(Enum):
    """Chaos experiment lifecycle states."""
    PLANNED = "planned"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    ABORTED = "aborted"

@dataclass
class ChaosExperiment:
    """Define a chaos engineering experiment."""
    name: str
    hypothesis: str  # What we expect to happen
    blast_radius: str  # Scope of impact
    rollback_plan: str
    success_criteria: str
    status: ExperimentStatus = ExperimentStatus.PLANNED
    started_at: datetime | None = None
    completed_at: datetime | None = None
    observations: list[str] | None = None

    def should_abort(self, metrics: dict) -> bool:
        """Check if experiment should be aborted.

        Args:
            metrics: Current system metrics

        Returns:
            bool: True if experiment should abort
        """
        # Abort if error rate exceeds 10%
        if metrics.get('error_rate', 0) > 0.10:
            return True

        # Abort if latency p99 exceeds 2 seconds
        if metrics.get('latency_p99', 0) > 2.0:
            return True

        return False

# Example: Database failover experiment
db_failover_experiment = ChaosExperiment(
    name="Database Primary Failover",
    hypothesis="System automatically fails over to replica within 30s with <1% error rate",
    blast_radius="Single database instance, 50% of production traffic",
    rollback_plan="Restore primary database immediately, redirect traffic",
    success_criteria="- Failover completes in <30s\n- Error rate <1%\n- No data loss",
)
```

## Chaos Testing Patterns

```python
# chaos_patterns.py - Common chaos engineering patterns
import time
import random
from typing import Protocol

class ChaosInjector(Protocol):
    """Interface for chaos injection."""

    def inject(self) -> None:
        """Inject chaos into the system."""
        ...

    def rollback(self) -> None:
        """Remove chaos and restore normal operation."""
        ...

class LatencyInjector:
    """Inject artificial latency into requests."""

    def __init__(self, target_service: str, latency_ms: int):
        self.target_service = target_service
        self.latency_ms = latency_ms

    def inject(self) -> None:
        """Add latency using iptables or proxy."""
        # Example using tc (traffic control) on Linux
        import subprocess
        subprocess.run([
            "tc", "qdisc", "add", "dev", "eth0",
            "root", "netem", "delay", f"{self.latency_ms}ms"
        ])

    def rollback(self) -> None:
        """Remove latency."""
        import subprocess
        subprocess.run(["tc", "qdisc", "del", "dev", "eth0", "root"])

class PodKiller:
    """Kill pods to test resilience."""

    def __init__(self, namespace: str, label_selector: str, kill_percentage: float = 0.5):
        self.namespace = namespace
        self.label_selector = label_selector
        self.kill_percentage = kill_percentage
        self.killed_pods = []

    def inject(self) -> None:
        """Randomly kill pods matching selector."""
        import subprocess

        # Get pods
        result = subprocess.run(
            ["kubectl", "get", "pods", "-n", self.namespace,
             "-l", self.label_selector, "-o", "name"],
            capture_output=True,
            text=True
        )

        pods = result.stdout.strip().split('\n')
        num_to_kill = int(len(pods) * self.kill_percentage)
        pods_to_kill = random.sample(pods, num_to_kill)

        # Kill selected pods
        for pod in pods_to_kill:
            subprocess.run(["kubectl", "delete", pod, "-n", self.namespace])
            self.killed_pods.append(pod)

    def rollback(self) -> None:
        """Pods will be recreated by deployment controller."""
        # Wait for pods to be recreated
        time.sleep(30)

class NetworkPartition:
    """Simulate network partition between services."""

    def __init__(self, source_pod: str, target_service: str):
        self.source_pod = source_pod
        self.target_service = target_service

    def inject(self) -> None:
        """Block network traffic using iptables."""
        import subprocess
        subprocess.run([
            "kubectl", "exec", self.source_pod, "--",
            "iptables", "-A", "OUTPUT", "-d", self.target_service, "-j", "DROP"
        ])

    def rollback(self) -> None:
        """Restore network traffic."""
        import subprocess
        subprocess.run([
            "kubectl", "exec", self.source_pod, "--",
            "iptables", "-D", "OUTPUT", "-d", self.target_service, "-j", "DROP"
        ])
```

## Chaos Experiment Runner

```python
# chaos_runner.py - Safe chaos experiment execution
from dataclasses import dataclass
from datetime import datetime, timedelta
import time

@dataclass
class SafetyConstraints:
    """Safety constraints for chaos experiments."""
    max_error_rate: float = 0.10  # 10%
    max_latency_p99: float = 2.0  # 2 seconds
    max_duration_minutes: int = 15
    business_hours_only: bool = True

class ChaosRunner:
    """Safely execute chaos experiments with monitoring."""

    def __init__(self, safety: SafetyConstraints):
        self.safety = safety

    def run_experiment(
        self,
        experiment: ChaosExperiment,
        injector: ChaosInjector,
        get_metrics: Callable[[], dict],
    ) -> ChaosExperiment:
        """Execute chaos experiment safely.

        Args:
            experiment: Experiment definition
            injector: Chaos injector implementation
            get_metrics: Function to fetch current metrics

        Returns:
            Updated experiment with results
        """
        # Pre-flight checks
        if self.safety.business_hours_only:
            current_hour = datetime.now().hour
            if 9 <= current_hour <= 17:  # Business hours
                experiment.status = ExperimentStatus.ABORTED
                experiment.observations = ["Aborted: Business hours constraint"]
                return experiment

        # Baseline metrics
        baseline_metrics = get_metrics()
        print(f"Baseline metrics: {baseline_metrics}")

        # Start experiment
        experiment.status = ExperimentStatus.RUNNING
        experiment.started_at = datetime.now()
        experiment.observations = []

        try:
            # Inject chaos
            print(f"Injecting chaos: {experiment.name}")
            injector.inject()

            # Monitor for max duration
            start_time = datetime.now()
            max_duration = timedelta(minutes=self.safety.max_duration_minutes)

            while datetime.now() - start_time < max_duration:
                time.sleep(10)  # Check every 10 seconds

                current_metrics = get_metrics()
                experiment.observations.append(
                    f"{datetime.now().isoformat()}: {current_metrics}"
                )

                # Check safety constraints
                if experiment.should_abort(current_metrics):
                    print("ABORTING: Safety constraint violated")
                    experiment.status = ExperimentStatus.ABORTED
                    break

            else:
                # Completed successfully
                experiment.status = ExperimentStatus.SUCCESS

        except Exception as e:
            print(f"ERROR: {e}")
            experiment.status = ExperimentStatus.FAILED
            experiment.observations.append(f"Exception: {str(e)}")

        finally:
            # Always rollback
            print("Rolling back chaos injection")
            injector.rollback()
            experiment.completed_at = datetime.now()

        return experiment

# Example usage
def get_current_metrics() -> dict:
    """Fetch metrics from Prometheus."""
    # In real implementation, query Prometheus
    return {
        'error_rate': 0.02,  # 2%
        'latency_p99': 0.45,  # 450ms
    }

safety = SafetyConstraints(business_hours_only=False)
runner = ChaosRunner(safety)

experiment = ChaosExperiment(
    name="Kill 50% of API pods",
    hypothesis="API remains available with 50% pod loss",
    blast_radius="50% of API pods",
    rollback_plan="Pods auto-restart via deployment",
    success_criteria="Error rate <5%, latency p99 <1s",
)

injector = PodKiller(
    namespace="production",
    label_selector="app=api",
    kill_percentage=0.5,
)

result = runner.run_experiment(experiment, injector, get_current_metrics)
print(f"Experiment status: {result.status.value}")
```

## Game Days

Scheduled chaos engineering practice sessions.

```yaml
# gameday_plan.yaml
gameday:
  date: "2024-02-15"
  duration: "2 hours"
  participants:
    - SRE team
    - Backend engineers
    - On-call rotation

  objectives:
    - Test incident response procedures
    - Validate monitoring and alerting
    - Practice communication protocols
    - Identify gaps in runbooks

  scenarios:
    - scenario: "Database Primary Failure"
      inject: "Terminate primary database pod"
      expected: "Automatic failover to replica in <30s"

    - scenario: "API Service Overload"
      inject: "Generate 10x normal traffic"
      expected: "Rate limiting activates, no errors"

    - scenario: "Network Partition"
      inject: "Block traffic between API and database"
      expected: "Circuit breaker opens, graceful degradation"

  success_criteria:
    - "All scenarios handled without escalation"
    - "MTTR <30 minutes for all scenarios"
    - "Documentation updated with learnings"
    - "Action items created for gaps"

  safety:
    - "Run in staging environment first"
    - "VP Engineering notified beforehand"
    - "Abort plan ready for each scenario"
    - "Customer support team on standby"
```

## Chaos Engineering Maturity

```python
from enum import IntEnum

class ChaosMaturity(IntEnum):
    """Chaos engineering maturity levels."""
    NONE = 0          # No chaos testing
    AD_HOC = 1        # Occasional manual tests
    SCHEDULED = 2     # Regular game days
    CONTINUOUS = 3    # Automated in CI/CD
    CULTURE = 4       # Embedded in development

def assess_maturity(practices: dict[str, bool]) -> ChaosMaturity:
    """Assess chaos engineering maturity level."""

    if not practices.get('any_chaos_testing'):
        return ChaosMaturity.NONE

    if not practices.get('regular_game_days'):
        return ChaosMaturity.AD_HOC

    if not practices.get('automated_chaos'):
        return ChaosMaturity.SCHEDULED

    if not practices.get('chaos_in_cicd'):
        return ChaosMaturity.CONTINUOUS

    return ChaosMaturity.CULTURE

# Example assessment
current_practices = {
    'any_chaos_testing': True,
    'regular_game_days': True,
    'automated_chaos': True,
    'chaos_in_cicd': False,
}

maturity = assess_maturity(current_practices)
print(f"Chaos maturity: {maturity.name}")
# Target: CONTINUOUS or CULTURE
```

---

## Reference: monitoring-alerting

# Monitoring and Alerting

## Golden Signals Monitoring

Monitor the four golden signals for every service.

```yaml
# prometheus_rules.yaml - Golden signals recording rules
groups:
  - name: golden_signals
    interval: 30s
    rules:
      # Latency: Request duration
      - record: service:http_request_duration_seconds:p50
        expr: |
          histogram_quantile(0.50,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          )

      - record: service:http_request_duration_seconds:p95
        expr: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          )

      - record: service:http_request_duration_seconds:p99
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          )

      # Traffic: Requests per second
      - record: service:http_requests:rate5m
        expr: |
          sum(rate(http_requests_total[5m])) by (service)

      # Errors: Error rate
      - record: service:http_requests:error_rate5m
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) by (service)
          /
          sum(rate(http_requests_total[5m])) by (service)

      # Saturation: Resource utilization
      - record: service:cpu_utilization
        expr: |
          avg(rate(container_cpu_usage_seconds_total[5m])) by (service)

      - record: service:memory_utilization
        expr: |
          avg(container_memory_working_set_bytes / container_spec_memory_limit_bytes)
          by (service)
```

## Alert Design Principles

Good alerts are actionable, not just informative.

```yaml
# alerts.yaml - SLO-based alerting
groups:
  - name: slo_alerts
    rules:
      # Multi-window burn rate alert (fast burn)
      - alert: ErrorBudgetBurnRateFast
        expr: |
          (
            service:http_requests:error_rate5m > (14.4 * 0.001)
            and
            service:http_requests:error_rate1h > (14.4 * 0.001)
          )
        for: 2m
        labels:
          severity: critical
          slo: availability
        annotations:
          summary: "Fast error budget burn on {{ $labels.service }}"
          description: |
            Service {{ $labels.service }} is burning error budget at 14.4x rate.
            At this rate, 30-day budget will exhaust in 2 days.

            Current error rate: {{ $value | humanizePercentage }}
            Threshold: 1.44%

            RUNBOOK: https://runbooks.example.com/error-budget-burn

      # Slow burn rate alert
      - alert: ErrorBudgetBurnRateSlow
        expr: |
          (
            service:http_requests:error_rate6h > (6 * 0.001)
            and
            service:http_requests:error_rate1d > (6 * 0.001)
          )
        for: 15m
        labels:
          severity: warning
          slo: availability
        annotations:
          summary: "Slow error budget burn on {{ $labels.service }}"
          description: |
            Service {{ $labels.service }} is burning error budget at 6x rate.

            RUNBOOK: https://runbooks.example.com/error-budget-burn

      # Latency SLO violation
      - alert: LatencySLOViolation
        expr: |
          service:http_request_duration_seconds:p99 > 0.5
        for: 5m
        labels:
          severity: warning
          slo: latency
        annotations:
          summary: "P99 latency exceeds 500ms on {{ $labels.service }}"
          description: |
            P99 latency is {{ $value }}s, exceeding 500ms threshold.

            Check:
            1. Database query performance
            2. External API latency
            3. Resource saturation (CPU/memory)

            RUNBOOK: https://runbooks.example.com/high-latency

      # Saturation alert
      - alert: HighMemoryUtilization
        expr: |
          service:memory_utilization > 0.85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on {{ $labels.service }}"
          description: |
            Memory utilization is {{ $value | humanizePercentage }}.

            Actions:
            1. Check for memory leaks
            2. Review recent deployments
            3. Consider scaling up

            RUNBOOK: https://runbooks.example.com/high-memory
```

## Alert Runbook Template

Every alert must link to a runbook with clear remediation steps.

```markdown
# Runbook: Error Budget Burn Rate

## Alert: ErrorBudgetBurnRateFast

### Description
The service is consuming error budget faster than sustainable rate.
At current rate, the 30-day error budget will be exhausted within 2 days.

### Severity: Critical

### Impact
- Users experiencing elevated error rates
- Risk of SLO violation and feature freeze
- Potential customer impact

### Triage Steps

1. **Check current error rate**
   ```promql
   rate(http_requests_total{status=~"5..", service="api"}[5m])
   ```

2. **Identify error types**
   ```bash
   kubectl logs -l app=api --tail=100 | grep ERROR
   ```

3. **Check recent deployments**
   ```bash
   kubectl rollout history deployment/api
   ```

4. **Review dependencies**
   - Database health
   - External API status
   - Infrastructure issues

### Remediation

**If caused by recent deployment:**
```bash
# Rollback to previous version
kubectl rollout undo deployment/api

# Verify rollback
kubectl rollout status deployment/api
```

**If database issue:**
```bash
# Check database connections
kubectl exec -it postgres-0 -- psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check slow queries
kubectl exec -it postgres-0 -- psql -c "SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

**If traffic spike:**
```bash
# Scale up replicas
kubectl scale deployment/api --replicas=10

# Enable rate limiting
kubectl apply -f rate-limit-config.yaml
```

### Communication

**Slack template:**
```
:fire: INCIDENT: Error budget burn rate critical

Service: api
Error rate: [X]%
Impact: [describe user impact]
ETA: [when will it be resolved]

Incident doc: [link]
```

### Prevention
- Add integration tests for this failure mode
- Implement circuit breaker for external dependencies
- Add capacity planning for traffic spikes
```

## Dashboard Configuration

```python
# grafana_dashboard.py - Generate SLO dashboard using Grafana SDK
from grafana_dashboard import Dashboard, Panel, Target

def create_slo_dashboard(service: str) -> dict:
    """Create SLO monitoring dashboard for a service."""

    dashboard = Dashboard(
        title=f"{service} - SLO Dashboard",
        tags=["slo", "sre", service],
        refresh="1m",
    )

    # SLI Current Value
    dashboard.add_panel(
        Panel(
            title="Availability SLI (30d)",
            targets=[
                Target(
                    expr=f"""
                    sum(rate(http_requests_total{{
                        status=~"2..",
                        service="{service}"
                    }}[30d]))
                    /
                    sum(rate(http_requests_total{{service="{service}"}}[30d]))
                    """,
                    legendFormat="Current SLI",
                ),
            ],
            thresholds=[
                {"value": 0.999, "color": "red"},
                {"value": 0.9995, "color": "yellow"},
                {"value": 1.0, "color": "green"},
            ],
        )
    )

    # Error Budget Remaining
    dashboard.add_panel(
        Panel(
            title="Error Budget Remaining",
            targets=[
                Target(
                    expr=f"""
                    (0.001 - (1 - (
                      sum(rate(http_requests_total{{
                          status=~"2..",
                          service="{service}"
                      }}[30d]))
                      /
                      sum(rate(http_requests_total{{service="{service}"}}[30d]))
                    ))) / 0.001 * 100
                    """,
                    legendFormat="Budget Remaining %",
                ),
            ],
            unit="percent",
        )
    )

    # Burn Rate
    dashboard.add_panel(
        Panel(
            title="Error Budget Burn Rate",
            targets=[
                Target(
                    expr=f"""
                    (1 - (
                      sum(rate(http_requests_total{{
                          status=~"2..",
                          service="{service}"
                      }}[1h]))
                      /
                      sum(rate(http_requests_total{{service="{service}"}}[1h]))
                    )) / 0.001
                    """,
                    legendFormat="1h burn rate",
                ),
            ],
            thresholds=[
                {"value": 1.0, "color": "green"},
                {"value": 6.0, "color": "yellow"},
                {"value": 14.4, "color": "red"},
            ],
        )
    )

    # Golden Signals
    dashboard.add_row("Golden Signals")

    dashboard.add_panel(
        Panel(
            title="Latency (P50, P95, P99)",
            targets=[
                Target(
                    expr=f'service:http_request_duration_seconds:p50{{service="{service}"}}',
                    legendFormat="p50",
                ),
                Target(
                    expr=f'service:http_request_duration_seconds:p95{{service="{service}"}}',
                    legendFormat="p95",
                ),
                Target(
                    expr=f'service:http_request_duration_seconds:p99{{service="{service}"}}',
                    legendFormat="p99",
                ),
            ],
            unit="s",
        )
    )

    return dashboard.to_json()
```

## Alert Fatigue Prevention

```python
from dataclasses import dataclass
from typing import List

@dataclass
class AlertQualityMetrics:
    """Track alert quality to prevent fatigue."""
    total_alerts: int
    actionable_alerts: int  # Required manual intervention
    false_positives: int
    auto_resolved: int  # Resolved before human action

    @property
    def precision(self) -> float:
        """Percentage of alerts that were actionable."""
        if self.total_alerts == 0:
            return 0.0
        return (self.actionable_alerts / self.total_alerts) * 100

    @property
    def toil_ratio(self) -> float:
        """Percentage of alerts that required manual work."""
        if self.total_alerts == 0:
            return 0.0
        return ((self.actionable_alerts + self.false_positives) / self.total_alerts) * 100

# Target: >90% precision, <30% toil
metrics = AlertQualityMetrics(
    total_alerts=100,
    actionable_alerts=85,
    false_positives=5,
    auto_resolved=10,
)

print(f"Alert precision: {metrics.precision}%")
print(f"Toil ratio: {metrics.toil_ratio}%")
```

## On-Call Alert Guidelines

```yaml
# on_call_alert_standards.yaml
alert_standards:
  page_worthy:
    - "Immediate user impact (>5% of users affected)"
    - "SLO violation in progress"
    - "Error budget burn rate critical (>10x)"
    - "Security incident"
    - "Data loss risk"

  not_page_worthy:
    - "Predictive alerts without current impact"
    - "Informational metrics"
    - "Non-user-facing issues"
    - "Slow trends (address during business hours)"

  alert_routing:
    critical:
      - page: on-call engineer
      - slack: "#incidents"
      - create: incident doc

    warning:
      - slack: "#alerts"
      - ticket: auto-create if persists >1h

    info:
      - dashboard: only
```

---

## Reference: slo-sli-management

# SLO/SLI Management

## SLI Definition Patterns

Service Level Indicators are quantitative measurements of service behavior.

### Request-Based SLIs

```python
# Availability SLI: Proportion of successful requests
# Good events: HTTP 200-299, 4XX (client errors don't count against SLI)
# Total events: All requests

def calculate_availability_sli(metrics):
    """Calculate availability SLI from request metrics."""
    successful_requests = metrics['http_2xx'] + metrics['http_4xx']
    total_requests = metrics['total_requests']

    if total_requests == 0:
        return 1.0  # No traffic = 100% available

    return successful_requests / total_requests

# Example: 99.9% of requests return successfully
# SLI = successful_requests / total_requests
```

### Latency-Based SLIs

```python
def calculate_latency_sli(latency_histogram, threshold_ms=500):
    """Calculate latency SLI from histogram.

    Args:
        latency_histogram: dict mapping latency buckets to request counts
        threshold_ms: latency threshold in milliseconds

    Returns:
        float: Proportion of requests faster than threshold
    """
    fast_requests = sum(
        count for bucket, count in latency_histogram.items()
        if bucket <= threshold_ms
    )
    total_requests = sum(latency_histogram.values())

    return fast_requests / total_requests if total_requests > 0 else 1.0

# Example: 99% of requests complete in <500ms
# SLI = requests_under_500ms / total_requests
```

## SLO Configuration

```yaml
# slo_config.yaml - Production API SLO definitions
apiVersion: sre/v1
kind: ServiceLevelObjective
metadata:
  service: payment-api
  environment: production
spec:
  slos:
    - name: availability
      description: "Users can successfully complete payment requests"
      sli:
        metric: http_requests_total
        query: |
          sum(rate(http_requests_total{status=~"2..|4..", service="payment-api"}[30d]))
          /
          sum(rate(http_requests_total{service="payment-api"}[30d]))
      target: 0.999  # 99.9%
      window: 30d

    - name: latency
      description: "Payment requests complete quickly"
      sli:
        metric: http_request_duration_seconds
        query: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket{service="payment-api"}[30d]))
            by (le)
          ) < 0.5
      target: 0.99  # 99% of requests under 500ms
      window: 30d
```

## Golden Signals

The four golden signals every service should measure:

```python
from dataclasses import dataclass
from typing import Dict

@dataclass
class GoldenSignals:
    """Four golden signals of monitoring."""

    # Latency: Time to service requests (success vs failure)
    latency_p50_ms: float
    latency_p95_ms: float
    latency_p99_ms: float

    # Traffic: Demand on your system (requests/sec)
    requests_per_second: float

    # Errors: Rate of failed requests
    error_rate: float  # 0.0 to 1.0

    # Saturation: How "full" is your service (CPU, memory, disk)
    cpu_utilization: float  # 0.0 to 1.0
    memory_utilization: float  # 0.0 to 1.0

    def is_healthy(self, slo_targets: Dict[str, float]) -> bool:
        """Check if all signals are within SLO targets."""
        return (
            self.latency_p99_ms <= slo_targets['latency_p99_ms'] and
            self.error_rate <= (1 - slo_targets['availability']) and
            self.cpu_utilization <= slo_targets['max_cpu'] and
            self.memory_utilization <= slo_targets['max_memory']
        )
```

## SLO Calculation Examples

```python
from datetime import timedelta
from typing import NamedTuple

class SLOTarget(NamedTuple):
    """SLO target configuration."""
    target: float  # 0.999 for 99.9%
    window: timedelta  # 30 days

    @property
    def error_budget(self) -> float:
        """Calculate error budget (1 - target)."""
        return 1 - self.target

    @property
    def allowed_downtime(self) -> timedelta:
        """Calculate allowed downtime in window."""
        total_seconds = self.window.total_seconds()
        allowed_seconds = total_seconds * self.error_budget
        return timedelta(seconds=allowed_seconds)

# Example SLOs
availability_slo = SLOTarget(target=0.999, window=timedelta(days=30))
print(f"Error budget: {availability_slo.error_budget * 100}%")
print(f"Allowed downtime: {availability_slo.allowed_downtime}")
# Output:
# Error budget: 0.1%
# Allowed downtime: 43.2 minutes per 30 days

latency_slo = SLOTarget(target=0.99, window=timedelta(days=30))
print(f"99% of requests must be fast")
print(f"1% can be slow: {latency_slo.error_budget * 100}%")
```

## Multi-Window SLO Tracking

```python
class MultiWindowSLO:
    """Track SLO compliance across multiple time windows."""

    def __init__(self, target: float):
        self.target = target
        self.windows = {
            '1h': timedelta(hours=1),
            '24h': timedelta(hours=24),
            '7d': timedelta(days=7),
            '30d': timedelta(days=30),
        }

    def check_compliance(self, sli_values: Dict[str, float]) -> Dict[str, bool]:
        """Check if SLI meets target in each window.

        Args:
            sli_values: Dict mapping window name to measured SLI

        Returns:
            Dict mapping window name to compliance boolean
        """
        return {
            window: sli >= self.target
            for window, sli in sli_values.items()
        }

    def get_burn_rate(self, current_sli: float) -> float:
        """Calculate error budget burn rate.

        Burn rate > 1.0 means burning budget faster than sustainable.
        """
        error_budget = 1 - self.target
        current_error_rate = 1 - current_sli

        if error_budget == 0:
            return float('inf')

        return current_error_rate / error_budget

# Usage
slo = MultiWindowSLO(target=0.999)
current_sli = 0.997  # 99.7% availability

burn_rate = slo.get_burn_rate(current_sli)
print(f"Burn rate: {burn_rate}x")
# If burn_rate = 3.0, burning budget 3x faster than sustainable
```

## SLO Review Checklist

Before finalizing SLOs:

1. **User-centric**: Does it measure user-facing impact?
2. **Achievable**: Can you meet it with current architecture?
3. **Measurable**: Can you accurately track the SLI?
4. **Meaningful**: Does violating it require action?
5. **Documented**: Is the calculation clear and agreed upon?
6. **Budgeted**: Is there an error budget policy?

## Common SLO Targets

```yaml
# Typical SLO targets by service tier
tier_1_critical:
  availability: 99.99%  # 4m 23s downtime/month
  latency_p99: 100ms

tier_2_important:
  availability: 99.9%   # 43m 28s downtime/month
  latency_p99: 500ms

tier_3_standard:
  availability: 99.5%   # 3h 37m downtime/month
  latency_p99: 1000ms
```
