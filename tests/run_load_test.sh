#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORTS_DIR="$SCRIPT_DIR/reports"
LOCUSTFILE="$SCRIPT_DIR/load/locustfile.py"

USERS=${USERS:-500}
SPAWN_RATE=${SPAWN_RATE:-50}
RUN_TIME=${RUN_TIME:-120s}
CORE_HOST=${CORE_HOST:-http://localhost:8080}

mkdir -p "$REPORTS_DIR"

echo "=== Running seed ==="
python "$SCRIPT_DIR/seed.py"

echo ""
echo "=== Starting load test ==="
echo "  users=$USERS  spawn-rate=$SPAWN_RATE  run-time=$RUN_TIME  host=$CORE_HOST"

locust \
  --headless \
  --users "$USERS" \
  --spawn-rate "$SPAWN_RATE" \
  --run-time "$RUN_TIME" \
  --host "$CORE_HOST" \
  --locustfile "$LOCUSTFILE" \
  --csv "$REPORTS_DIR/load" \
  --html "$REPORTS_DIR/load_report.html" \
  --exit-code-on-error 0

echo ""
echo "=== Asserting results ==="
python - <<'EOF'
import csv, sys

stats_file = "tests/reports/load_stats.csv"
try:
    rows = list(csv.DictReader(open(stats_file)))
except FileNotFoundError:
    print(f"FAIL: {stats_file} not found")
    sys.exit(1)

agg = next((r for r in rows if r["Name"] == "Aggregated"), None)
if not agg:
    print("FAIL: no Aggregated row in CSV")
    sys.exit(1)

p99_ms   = float(agg.get("99%", agg.get("99th Percentile Response Time", 0)) or 0)
rps      = float(agg.get("Requests/s", 0) or 0)
failures = float(agg.get("Failure Count", 0) or 0)
total    = float(agg.get("Request Count", 1) or 1)
fail_pct = (failures / total) * 100

results = [
    ("p99 < 10ms",         p99_ms  < 10,   f"{p99_ms:.1f}ms"),
    ("RPS > 5000",         rps     > 5000,  f"{rps:.0f} RPS"),
    ("Failure rate < 0.1%", fail_pct < 0.1, f"{fail_pct:.3f}%"),
]

all_pass = True
for label, passed, value in results:
    status = "PASS" if passed else "FAIL"
    print(f"  [{status}] {label}: {value}")
    if not passed:
        all_pass = False

print()
print("LOAD TEST PASSED" if all_pass else "LOAD TEST FAILED — see above")
sys.exit(0 if all_pass else 1)
EOF
