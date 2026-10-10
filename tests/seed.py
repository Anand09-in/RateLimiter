"""Seed test data for load testing via the Admin API."""
import json
import sys
import requests

ADMIN_BASE = "http://localhost:8000"
KEYS_FILE = "tests/keys.json"

RULES = [
    {"name": "fixed-rule",   "algorithm": "fixed",   "limit_count": 1000, "window_secs": 60, "endpoint": "/api/test"},
    {"name": "sliding-rule", "algorithm": "sliding",  "limit_count": 1000, "window_secs": 60, "endpoint": "/api/test"},
    {"name": "token-rule",   "algorithm": "token",    "limit_count": 1000, "window_secs": 60, "endpoint": "/api/test"},
]

CLIENT_NAMES = ["load-client-fixed", "load-client-sliding", "load-client-token"]


def get_or_create_rule(rule_payload: dict) -> str:
    existing = requests.get(f"{ADMIN_BASE}/rules").json()
    for r in existing:
        if r["name"] == rule_payload["name"]:
            print(f"  rule exists: {r['name']} ({r['id']})")
            return r["id"]
    resp = requests.post(f"{ADMIN_BASE}/rules", json=rule_payload)
    resp.raise_for_status()
    rule = resp.json()
    print(f"  created rule: {rule['name']} ({rule['id']})")
    return rule["id"]


def get_or_create_client(name: str, rule_id: str) -> str:
    existing = requests.get(f"{ADMIN_BASE}/clients").json()
    for c in existing:
        if c["name"] == name:
            print(f"  client exists: {c['name']} — key: {c['api_key'][:12]}...")
            return c["api_key"]
    resp = requests.post(f"{ADMIN_BASE}/clients", json={"name": name, "rule_id": rule_id})
    resp.raise_for_status()
    client = resp.json()
    print(f"  created client: {client['name']} — key: {client['api_key'][:12]}...")
    return client["api_key"]


def main():
    print("=== Seeding test data ===")
    keys = {}

    for rule_payload, client_name in zip(RULES, CLIENT_NAMES):
        algo = rule_payload["algorithm"]
        print(f"\n[{algo}]")
        rule_id = get_or_create_rule(rule_payload)
        api_key = get_or_create_client(client_name, rule_id)
        keys[algo] = api_key

    with open(KEYS_FILE, "w") as f:
        json.dump(keys, f, indent=2)

    print(f"\nKeys saved to {KEYS_FILE}")
    print("Done.")


if __name__ == "__main__":
    try:
        main()
    except requests.exceptions.ConnectionError:
        print("ERROR: Cannot connect to admin API at", ADMIN_BASE)
        sys.exit(1)
