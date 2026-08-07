#!/usr/bin/env python3
"""Apply the pre-fix discovery/test pass to the canonical feature tracker."""

from __future__ import annotations

import csv
from pathlib import Path

TRACKER = Path("feature_status_tracker.csv")

PASS_EVIDENCE = {
    "WEB": "qa_evidence/logs/feature_ui_audit.json; qa_evidence/lighthouse/landing-desktop/report.json",
    "TOUR": "qa_evidence/logs/feature_ui_audit.json; qa_evidence/screenshots/tour-final-1920x1080.png",
    "AAPL": "qa_evidence/logs/aapl-walkthrough-test.log; qa_evidence/logs/feature_ui_audit.json",
    "CLOUD": "qa_evidence/logs/go-test-all.log; qa_evidence/logs/feature_ui_audit.json",
    "API": "qa_evidence/logs/go-test-all.log",
    "DOC": "qa_evidence/logs/feature_ui_audit.json",
}

FAILURES = {
    "WEB-002": {
        "bug": "Mobile menu displays “Menu” but its aria-label is “Toggle navigation”; Lighthouse flags label-content-name-mismatch.",
        "severity": "Medium",
        "actual": "The menu opens and no horizontal overflow occurs, but the accessible name does not contain the visible label.",
        "evidence": "qa_evidence/lighthouse/landing-mobile/report.json",
    },
    "WEB-014": {
        "bug": "The final Cloud Run action opens a stale competition-branded, HUT-first surface instead of continuing the Apple-first Evidence OS narrative.",
        "severity": "High",
        "actual": "All actions resolve, but the Cloud Run destination changes product identity and priority unexpectedly.",
        "evidence": "qa_evidence/screenshots/cloud-root-1440x1000.png",
    },
    "TOUR-001": {
        "bug": "Slide 4 switches from the Apple reference to HUT without an explicit ‘second proof’ transition, weakening the Apple-first story.",
        "severity": "Medium",
        "actual": "The tour activates and is navigable, but the issuer switch is abrupt and can be mistaken for the primary proof.",
        "evidence": "qa_evidence/logs/feature_ui_audit.json",
    },
    "CLOUD-001": {
        "bug": "Root landing still presents ‘Google for Startups AI Agents Challenge 2026 · Track 3’ and HUT as the product priority; this is stale competition framing.",
        "severity": "High",
        "actual": "The root loads and both actions work, but it does not represent the current Apple-first DeltaSignal Evidence OS product.",
        "evidence": "qa_evidence/screenshots/cloud-root-1440x1000.png",
    },
    "CLOUD-002": {
        "bug": "Demo landing is visually sparse after the narrow hero card, delays important controls behind long reveal animations, and retains stale competition branding.",
        "severity": "High",
        "actual": "The page loads, but first-view comprehension and product continuity are weaker than the public landing.",
        "evidence": "qa_evidence/screenshots/cloud-demo-1440x1000.png",
    },
    "CLOUD-012": {
        "bug": "The /demo/run interface is wider than the 390px viewport while horizontal overflow is suppressed, clipping headings, explanatory copy, inputs, and controls.",
        "severity": "Critical",
        "actual": "Desktop is readable; the mobile viewport shows only the left portion of the operator panel and is not reliably operable.",
        "evidence": "qa_evidence/screenshots/cloud-run-390x844.png",
    },
}

rows = list(csv.DictReader(TRACKER.open(newline="", encoding="utf-8")))
fieldnames = list(rows[0].keys())

for row in rows:
    feature_id = row["Feature ID"]
    prefix = feature_id.split("-", 1)[0]
    evidence = PASS_EVIDENCE[prefix]
    if feature_id in FAILURES:
        issue = FAILURES[feature_id]
        row["Actual Behaviour"] = issue["actual"]
        row["Status"] = "Failed Test"
        row["Test Result"] = f"FAIL — {issue['evidence']}"
        row["Error / Bug Found"] = issue["bug"]
        row["Severity"] = issue["severity"]
        row["Fix Required"] = "Yes"
        row["Retest Steps"] = row["Test Steps"]
        row["Final Status"] = ""
        row["Notes"] = f"Pre-fix discovery evidence: {issue['evidence']}"
    else:
        row["Actual Behaviour"] = "Observed behavior matched the tracker-backed expectation in the discovery test pass."
        row["Status"] = "Verified"
        row["Test Result"] = f"PASS — {evidence}"
        row["Error / Bug Found"] = ""
        row["Severity"] = ""
        row["Fix Required"] = "No"
        row["Retest Result"] = "Not required in discovery pass."
        row["Final Status"] = "Verified"
        row["Notes"] = f"Discovery evidence: {evidence}"

with TRACKER.open("w", newline="", encoding="utf-8") as handle:
    writer = csv.DictWriter(handle, fieldnames=fieldnames)
    writer.writeheader()
    writer.writerows(rows)

print(f"Updated {len(rows)} stories; failures={len(FAILURES)}")
