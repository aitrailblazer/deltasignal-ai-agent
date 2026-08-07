#!/usr/bin/env python3
"""Record the post-fix retest pass for tracker-backed UI issues."""

from __future__ import annotations

import csv
from pathlib import Path

tracker = Path("feature_status_tracker.csv")
rows = list(csv.DictReader(tracker.open(newline="", encoding="utf-8")))
fieldnames = list(rows[0].keys())

fixes = {
    "WEB-002": {
        "implemented": "Accessible name now includes the visible “Menu” label.",
        "files": "index.html",
        "retest": "PASS — Lighthouse mobile: accessibility 100, 55 passed, 0 failed.",
        "evidence": "qa_evidence/lighthouse/retest-landing-mobile/report.json",
    },
    "WEB-014": {
        "implemented": "Cloud Run now continues the Apple-first Evidence OS narrative and labels HUT as a second research-memory proof.",
        "files": "cmd/server/demo_ui.go; cmd/server/demo_ui_test.go",
        "retest": "PASS — all final actions resolve into coherent, explicitly prioritized product surfaces.",
        "evidence": "qa_evidence/screenshots/cloud-root-1440x1000.png; qa_evidence/logs/retest-go-test-all.log",
    },
    "TOUR-001": {
        "implemented": "Tour slide 4 now explicitly labels HUT as the second proof beyond the flagship Apple workflow.",
        "files": "index.html",
        "retest": "PASS — six-slide navigation preserves the Apple-first story and explains the issuer transition.",
        "evidence": "qa_evidence/logs/retest-feature_ui_audit.json",
    },
    "CLOUD-001": {
        "implemented": "Root service page was reframed as DeltaSignal Evidence OS with Apple SEC/XBRL as stage one and HUT memory as a secondary action.",
        "files": "cmd/server/demo_ui.go; cmd/server/demo_ui_test.go",
        "retest": "PASS — desktop/mobile render without overflow; stale competition language absent; Apple and HUT roles explicit.",
        "evidence": "qa_evidence/screenshots/cloud-root-1440x1000.png; qa_evidence/lighthouse/retest-cloud-root-desktop/report.json",
    },
    "CLOUD-002": {
        "implemented": "Demo landing now uses a complete architecture visual, sub-second reveal sequence, and explicit second-proof context.",
        "files": "cmd/server/demo_ui.go; cmd/server/demo_ui_test.go",
        "retest": "PASS — first viewport is populated, readable, and visually explains User → Cloud Run → MCP/River/Evidence → packet.",
        "evidence": "qa_evidence/screenshots/cloud-demo-1440x1000.png",
    },
    "CLOUD-012": {
        "implemented": "Mobile CSS now removes fixed-width constraints, uses minmax grid tracks, contains decorative overflow, and stacks operator fields.",
        "files": "cmd/server/demo_ui.go; cmd/server/demo_ui_test.go",
        "retest": "PASS — 390px viewport has innerWidth=390, document/body scrollWidth=390, and an uncropped full-page screenshot.",
        "evidence": "qa_evidence/screenshots/retest-cloud-run-390x844.png; qa_evidence/logs/retest-go-server.log",
    },
}

for row in rows:
    item = fixes.get(row["Feature ID"])
    if not item:
        continue
    row["Actual Behaviour"] = item["implemented"]
    row["Status"] = "Fixed"
    row["Test Result"] = row["Test Result"] + " | FIX VERIFIED"
    row["Fix Implemented"] = item["implemented"]
    row["Files Changed"] = item["files"]
    row["Retest Result"] = item["retest"]
    row["Final Status"] = "Verified"
    row["Notes"] = row["Notes"] + f" | Retest evidence: {item['evidence']}"

with tracker.open("w", newline="", encoding="utf-8") as handle:
    writer = csv.DictWriter(handle, fieldnames=fieldnames)
    writer.writeheader()
    writer.writerows(rows)

print(f"Recorded {len(fixes)} fixed and retested stories.")
