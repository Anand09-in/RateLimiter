"""Locust load test for the Rate Limiter core service."""
import json
import os

from locust import HttpUser, task, between, events

KEYS_FILE = os.path.join(os.path.dirname(__file__), "..", "keys.json")
CORE_HOST = os.getenv("CORE_HOST", "http://localhost:8080")

_keys: dict = {}


@events.init.add_listener
def on_locust_init(environment, **_kwargs):
    global _keys
    try:
        with open(KEYS_FILE) as f:
            _keys = json.load(f)
    except FileNotFoundError:
        print(f"WARNING: {KEYS_FILE} not found — run tests/seed.py first")
        _keys = {"fixed": "missing", "sliding": "missing", "token": "missing"}


class FixedWindowUser(HttpUser):
    host = CORE_HOST
    wait_time = between(0, 0)

    @task
    def check(self):
        with self.client.post(
            "/check",
            json={"endpoint": "/api/test"},
            headers={"X-API-Key": _keys.get("fixed", "")},
            catch_response=True,
        ) as resp:
            # 429 is expected behaviour, not a failure
            if resp.status_code in (200, 429):
                resp.success()
            else:
                resp.failure(f"unexpected status {resp.status_code}")


class SlidingWindowUser(HttpUser):
    host = CORE_HOST
    wait_time = between(0, 0)

    @task
    def check(self):
        with self.client.post(
            "/check",
            json={"endpoint": "/api/test"},
            headers={"X-API-Key": _keys.get("sliding", "")},
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 429):
                resp.success()
            else:
                resp.failure(f"unexpected status {resp.status_code}")


class TokenBucketUser(HttpUser):
    host = CORE_HOST
    wait_time = between(0, 0)

    @task
    def check(self):
        with self.client.post(
            "/check",
            json={"endpoint": "/api/test"},
            headers={"X-API-Key": _keys.get("token", "")},
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 429):
                resp.success()
            else:
                resp.failure(f"unexpected status {resp.status_code}")
