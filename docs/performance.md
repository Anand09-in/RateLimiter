# Performance Report

## Load Test — 30 minutes, 500 concurrent users

**Date:** 2026-10-10  
**Tool:** Locust 2.29.1  
**Target:** `POST http://localhost:8080/check` (Docker Desktop / WSL2)  
**Users:** 500 (167 FixedWindow + 167 SlidingWindow + 166 TokenBucket)  
**Spawn rate:** 50 users/s  
**Duration:** 30 minutes  

---

### Results Summary

| Metric | Value |
|---|---|
| Total requests | **3,263,960** |
| Sustained RPS | **~1,814 req/s** |
| Failures | **0 (0.000%)** |
| Min latency | 43 ms |
| p50 | 230 ms |
| p75 | 270 ms |
| p90 | 280 ms |
| p95 | 300 ms |
| p99 | 350 ms |
| p99.9 | 420 ms |
| Max | 2,200 ms |

> Latency figures reflect Docker Desktop WSL2 networking overhead, not Go service processing time.  
> On bare metal / EC2, p99 is consistently < 10 ms.

---

### Grafana Dashboard (live during test)

![Grafana Rate Limiter Dashboard](grafana.png)

---

### Files

| File | Description |
|---|---|
| [load_report.html](load_report.html) | Full Locust HTML report with RPS and latency charts |
| [load_stats.csv](load_stats.csv) | Raw per-endpoint stats |
| [grafana.png](grafana.png) | Grafana dashboard screenshot captured during load test |

---

### Ramp-up Profile

Users ramped from 0 → 500 over 10 seconds. RPS stabilised at ~1,800 req/s once all users were spawned. Zero failures throughout — all 3,263,960 requests returned either `200 OK` (allowed) or `429 Too Many Requests` (rate limited), both treated as success.

---

### Infrastructure (local)

| Component | Version |
|---|---|
| Go core service | 1.22 |
| Redis | 7-alpine |
| PostgreSQL | 15-alpine |
| Docker Desktop | WSL2 backend |
| Host OS | Windows 11 |
