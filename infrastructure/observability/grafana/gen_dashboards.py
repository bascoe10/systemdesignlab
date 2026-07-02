#!/usr/bin/env python3
"""Generates the four SystemDesignLab dashboards into dashboards/.

Dashboards are committed artifacts (Grafana provisioning reads the JSON),
but edit THIS file and re-run it rather than editing JSON by hand:

    python3 gen_dashboards.py
"""
import json
import os

DS = {"type": "prometheus", "uid": "prometheus"}
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "dashboards")

_pid = 0


def next_id():
    global _pid
    _pid += 1
    return _pid


def targets(*exprs):
    return [
        {"datasource": DS, "expr": expr, "legendFormat": legend, "refId": chr(65 + i)}
        for i, (expr, legend) in enumerate(exprs)
    ]


def timeseries(title, exprs, x, y, w=12, h=8, unit="short", description="", max_val=None):
    fc = {"defaults": {"unit": unit, "custom": {"fillOpacity": 10, "showPoints": "never"}}, "overrides": []}
    if max_val is not None:
        fc["defaults"]["max"] = max_val
        fc["defaults"]["min"] = 0
    return {
        "id": next_id(), "type": "timeseries", "title": title, "description": description,
        "datasource": DS, "gridPos": {"x": x, "y": y, "w": w, "h": h},
        "fieldConfig": fc, "targets": targets(*exprs),
        "options": {"legend": {"displayMode": "list", "placement": "bottom"},
                    "tooltip": {"mode": "multi"}},
    }


def stat(title, expr, x, y, w=6, h=5, unit="percentunit", description="", thresholds=None):
    steps = thresholds or [{"color": "red", "value": None}, {"color": "green", "value": 0.995}]
    return {
        "id": next_id(), "type": "stat", "title": title, "description": description,
        "datasource": DS, "gridPos": {"x": x, "y": y, "w": w, "h": h},
        "fieldConfig": {"defaults": {"unit": unit, "decimals": 2,
                                     "thresholds": {"mode": "absolute", "steps": steps}},
                        "overrides": []},
        "targets": targets((expr, "")),
        "options": {"reduceOptions": {"calcs": ["lastNotNull"]}, "colorMode": "value",
                    "graphMode": "area"},
    }


def text(content, x, y, w=24, h=3):
    return {
        "id": next_id(), "type": "text", "title": "",
        "gridPos": {"x": x, "y": y, "w": w, "h": h},
        "options": {"mode": "markdown", "content": content},
    }


def dashboard(uid, title, panels, time_from="now-15m", refresh="5s"):
    return {
        "uid": uid, "title": title, "tags": ["systemdesignlab"],
        "timezone": "browser", "schemaVersion": 39, "version": 1, "editable": True,
        "refresh": refresh, "time": {"from": time_from, "to": "now"},
        "panels": panels,
    }


HIT_RATE = ('sum(rate(cache_requests_total{op="get",result="hit"}[1m])) / '
            'clamp_min(sum(rate(cache_requests_total{op="get",result=~"hit|miss|bypass"}[1m])), 0.001)')
# `or vector(0)`: sum() over an empty vector is "no data"; zero errors must
# draw a zero line, not an empty panel.
ERROR_RATIO = ('(sum(rate(http_requests_total{job="api-gateway",code=~"5.."}[1m])) or vector(0)) / '
               'clamp_min(sum(rate(http_requests_total{job="api-gateway"}[1m])), 0.001)')
P99 = ('histogram_quantile(0.99, sum by (le) '
       '(rate(http_request_duration_seconds_bucket{job="api-gateway"}[1m])))')
P50 = ('histogram_quantile(0.50, sum by (le) '
       '(rate(http_request_duration_seconds_bucket{job="api-gateway"}[1m])))')
NODE_OPS = 'sum by (node) (rate(cache_node_requests_total[1m]))'
DB_QPS = 'sum by (query) (rate(db_query_duration_seconds_count[1m]))'

golden = dashboard("sdl-golden-signals", "Golden Signals Overview", [
    text("## The Four Golden Signals — Latency · Traffic · Errors · Saturation\n"
         "Start every investigation here. Which signal moved first? "
         "Drill into **Component Deep Dive** to find out why.", 0, 0),
    timeseries("Traffic — requests/sec by service",
               [('sum by (job) (rate(http_requests_total{route!=""}[1m]))', "{{job}}")],
               0, 3, unit="reqps",
               description="Demand on the system, measured at each service. The gateway sees everything; the read/write split shows up as redirector vs shortener traffic."),
    timeseries("Latency — gateway p50 / p99",
               [(P99, "p99"), (P50, "p50")],
               12, 3, unit="s",
               description="End-to-end request latency at the public edge. A wide p50/p99 gap means a slow minority — usually cache misses taking the database path."),
    timeseries("Errors — 5xx ratio and rate-limited requests",
               [(ERROR_RATIO, "5xx ratio"),
                ('sum(rate(gateway_ratelimit_rejected_total[1m]))', "429/sec (right axis)")],
               0, 11,
               description="Failed request ratio at the gateway, plus requests rejected by the rate limiter."),
    timeseries("Saturation — cache memory used / limit",
               [('cache_node_memory_used_bytes / clamp_min(cache_node_memory_max_bytes, 1)', "{{node}}")],
               12, 11, unit="percentunit", max_val=1.1,
               description="How full each cache node is. Pinned at 100% with a rising eviction rate means the working set no longer fits."),
    timeseries("Saturation — DB connection pool",
               [('db_pool_in_use', "{{job}} in use"), ('db_pool_max', "{{job}} max")],
               0, 19,
               description="Connections in use vs pool limit. A saturated pool queues queries and shows up as tail latency, not errors."),
    timeseries("Cache hit rate",
               [(HIT_RATE, "hit rate")],
               12, 19, unit="percentunit", max_val=1.05,
               description="Fraction of reads served from cache. Healthy steady state is >85%. Bypass (ring unavailable) counts as a miss here."),
])

deep = dashboard("sdl-component-deep-dive", "Component Deep Dive", [
    timeseries("Cache hit rate", [(HIT_RATE, "hit rate")], 0, 0, unit="percentunit", max_val=1.05),
    timeseries("Cache read outcomes",
               [('sum by (result) (rate(cache_requests_total{op="get"}[1m]))', "{{result}}")],
               12, 0, unit="reqps",
               description="hit / miss / bypass / error. 'bypass' means the hash ring could not resolve a node — on Level 3 that is your stub."),
    timeseries("Consistent hashing — ops per cache node",
               [(NODE_OPS, "{{node}}")],
               0, 8, unit="reqps",
               description="How the ring spreads keys. Healthy: three roughly even lines. One node taking most traffic = too few virtual nodes or a hot key."),
    timeseries("Items per cache node",
               [('cache_node_items', "{{node}}")],
               12, 8,
               description="Key count per node — the storage view of ring balance."),
    timeseries("Evictions per cache node",
               [('rate(cache_node_evictions_total[1m])', "{{node}}")],
               0, 16, unit="ops",
               description="Sustained evictions mean the working set exceeds max_memory."),
    timeseries("Cache memory used per node",
               [('cache_node_memory_used_bytes', "{{node}} used"),
                ('cache_node_memory_max_bytes', "{{node}} limit")],
               12, 16, unit="bytes"),
    timeseries("DB query latency p99",
               [('histogram_quantile(0.99, sum by (le, query) (rate(db_query_duration_seconds_bucket[1m])))', "{{query}}")],
               0, 24, unit="s"),
    timeseries("DB queries/sec",
               [(DB_QPS, "{{query}}")],
               12, 24, unit="reqps",
               description="Every cache miss becomes a get_url query. This panel is the mirror image of the hit rate."),
])

chaos = dashboard("sdl-chaos-impact", "Chaos Impact", [
    text("## Chaos Impact\nRun a `make chaos-*` command, keep a load test running, and watch "
         "the blast radius here. Ask: which signal degraded first, how far did it cascade, "
         "and how long did recovery take after `make chaos-restore`?", 0, 0),
    timeseries("Error ratio (gateway)", [(ERROR_RATIO, "5xx ratio")], 0, 3, unit="percentunit"),
    timeseries("Latency p99 (gateway)", [(P99, "p99")], 12, 3, unit="s"),
    timeseries("Cache hit rate", [(HIT_RATE, "hit rate")], 0, 11, unit="percentunit", max_val=1.05),
    timeseries("DB queries/sec", [(DB_QPS, "{{query}}")], 12, 11, unit="reqps",
               description="When the cache dies, this is where the traffic goes."),
    timeseries("Ops per cache node", [(NODE_OPS, "{{node}}")], 0, 19, unit="reqps",
               description="Kill one node and watch the ring reroute its share."),
    timeseries("Throughput by service",
               [('sum by (job) (rate(http_requests_total{route!=""}[1m]))', "{{job}}")],
               12, 19, unit="reqps"),
], time_from="now-30m")

FAST_RATIO = ('sum(rate(http_request_duration_seconds_bucket{job="redirector",route="redirect",le="0.1"}[30m])) / '
              'clamp_min(sum(rate(http_request_duration_seconds_count{job="redirector",route="redirect"}[30m])), 0.001)')
AVAIL = ('1 - ((sum(rate(http_requests_total{job="redirector",route="redirect",code=~"5.."}[30m])) or vector(0)) / '
         'clamp_min(sum(rate(http_requests_total{job="redirector",route="redirect"}[30m])), 0.001))')

slo = dashboard("sdl-sli-slo", "SLI / SLO Tracker", [
    text("## Service Level Objectives — URL Shortener\n"
         "| SLI | SLO |\n|---|---|\n"
         "| Redirect availability (non-5xx) | **99.5%** over 30m |\n"
         "| Fast redirects (< 100ms) | **99.5%** over 30m |\n\n"
         "On Level 5 you define and justify your own SLOs in the decision journal.", 0, 0),
    stat("Redirect availability (30m)", AVAIL, 0, 3, w=12, h=6,
         description="SLO: 99.5% of redirects succeed."),
    stat("Redirects under 100ms (30m)", FAST_RATIO, 12, 3, w=12, h=6,
         description="SLO: 99.5% of redirects complete in under 100ms."),
    timeseries("SLIs over time (5m window)",
               [(AVAIL.replace("[30m]", "[5m]"), "availability"),
                (FAST_RATIO.replace("[30m]", "[5m]"), "fast redirect ratio")],
               0, 9, w=24, unit="percentunit", max_val=1.01,
               description="The 0.995 objective is your floor. Time spent below it burns error budget."),
], time_from="now-1h", refresh="10s")

os.makedirs(OUT, exist_ok=True)
for name, dash in [("golden-signals", golden), ("component-deep-dive", deep),
                   ("chaos-impact", chaos), ("sli-slo", slo)]:
    path = os.path.join(OUT, f"{name}.json")
    with open(path, "w") as f:
        json.dump(dash, f, indent=2)
        f.write("\n")
    print(f"wrote {path}")
