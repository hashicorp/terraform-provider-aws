#!/usr/bin/env python3
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

"""
Audit remote branches on origin and rank them as cleanup candidates.

Ordering is "low-hanging fruit" first: low effort to validate + high likelihood
deletable at the top, descending to high effort / low likelihood at the bottom.
Branches that cannot be deleted (org rule protects `release*`) and protected
infrastructure (`gh-pages`) are listed separately and never flagged.

Requirements: `git` and an authenticated `gh` CLI.
Usage:
    .ci/scripts/audit-branches.py            # human-readable ranked report
    .ci/scripts/audit-branches.py --markdown # GitHub-issue markdown (uses #<num> refs)
"""

import subprocess
import re
import os
import sys
import time
from collections import Counter

REPO = "hashicorp/terraform-provider-aws"

# --- Roster (GitHub logins) -------------------------------------------------
CURRENT = {"johnsonaj", "yakdriver", "gdavison", "jar-b",
           "ewbankkit", "subham-ibmhc", "taruntej-a", "brosas07"}
TEAM = {"breathingdust", "justinretzolk"}

# committer email -> login (best-effort attribution for no-PR branches)
EMAIL_TO_LOGIN = {
    "gdavison@hashicorp.com": "gdavison",
    "jared.baker@hashicorp.com": "jar-b",
    "kit_ewbank@hotmail.com": "ewbankkit",
    "adrian.johnson@hashicorp.com": "johnsonaj",
    "subham.mukhopadhyay@hashicorp.com": "subham-ibmhc",
    "dirk.avery@gmail.com": "yakdriver",
    "yakdriver@users.noreply.github.com": "yakdriver",
    "31492422+yakdriver@users.noreply.github.com": "yakdriver",
    "tarunsraina483@gmail.com": "taruntej-a",
    "tarunteja@taruns-macbook-pro.local": "taruntej-a",
    "sdavis@hashicorp.com": "breathingdust",
    "44710313+justinretzolk@users.noreply.github.com": "justinretzolk",
}
# Known departed HashiCorp folks / former maintainers (no longer stewards)
FORMER_EMPLOYEE_EMAILS = {
    "thomas@thomaszwskismbp.lan", "thomas.zalewski@hashicorp.com",
    "sharon.nam@hashicorp.com", "sharon.nam@sharon.nam-pjcj4f7jh7",
    "jitendra.gangwar1@hashicorp.com",
}
BOT_EMAIL_MARKERS = ("github-actions", "noreply@github.com",
                     "changelogbot@hashicorp.com", "update-schemas@github.com")

# Append-only ledger of deleted branches (survives regeneration; git can no
# longer report deleted branches). Maintained in the sibling TSV file.
DELETED_LEDGER = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                              "deleted-branches.tsv")


def load_deleted():
    """Return list of dicts {date, branch, sha, pr, bucket} from the ledger."""
    recs = []
    try:
        with open(DELETED_LEDGER) as f:
            for line in f:
                line = line.rstrip("\n")
                if not line or line.startswith("#"):
                    continue
                parts = line.split("\t")
                if len(parts) < 5:
                    continue
                recs.append(dict(date=parts[0], branch=parts[1], sha=parts[2],
                                 pr=parts[3], bucket=parts[4]))
    except FileNotFoundError:
        pass
    return recs


def sh(args):
    return subprocess.run(args, capture_output=True, text=True).stdout.strip()


def email_to_status(email):
    e = email.lower().strip()
    if any(m in e for m in BOT_EMAIL_MARKERS):
        return "bot", "bot"
    login = EMAIL_TO_LOGIN.get(e)
    if login:
        return status_of(login), login
    if e in FORMER_EMPLOYEE_EMAILS:
        return "other", "former-employee"
    return None, None


def status_of(login):
    l = (login or "").lower()
    if l in CURRENT:
        return "current"
    if l in TEAM:
        return "team"
    return "other"


def owner(pr_author, cemail):
    cand = []
    if pr_author and pr_author not in ("", "null"):
        if pr_author.startswith("app/"):
            cand.append(("bot", pr_author))
        else:
            cand.append((status_of(pr_author), pr_author))
    st, who = email_to_status(cemail)
    if st:
        cand.append((st, who))
    rank = {"current": 0, "team": 1, "other": 2, "bot": 3, None: 4}
    if not cand:
        if cemail.lower().strip() in FORMER_EMPLOYEE_EMAILS:
            return ("other", "former-employee")
        return ("unknown", "?")
    cand.sort(key=lambda c: rank.get(c[0], 4))
    return cand[0]


# --- Ranking buckets (index = priority; lower = delete sooner) --------------
RANKS = [
    ("R1", "Automation/bot, PR closed", "trivial (bot, job done)", "very high"),
    ("R2", "Already in `main` — lossless, no open PR", "trivial (content preserved)", "very high"),
    ("R3", "No current owner (departed/external), no open PR", "low–medium (skim commits)", "high"),
    ("R4", "Current maintainer/team, stale >1yr, no open PR", "medium (confirm w/ owner)", "medium"),
    ("R5", "Current maintainer/team, recent <1yr, no open PR", "medium", "low"),
    ("R6", "Active — open PR", "n/a", "keep"),
    ("X1", "Cannot delete — `release*` / `v*` (org rule)", "n/a", "n/a"),
    ("X2", "Protected infra — `gh-pages`", "n/a", "n/a"),
]


def bucket(r):
    b = r["branch"]
    if b == "gh-pages":
        return "X2"
    if re.match(r"^(release|v\d)", b):  # org rule protects release* and v* branches
        return "X1"
    if r["prstate"] == "OPEN":
        return "R6"
    if r["owner"][0] == "bot":
        return "R1"
    if r["inmain"]:
        return "R2"
    if r["owner"][0] in ("other", "unknown"):
        return "R3"
    if r["age"] > 365:
        return "R4"
    return "R5"


def collect():
    refs = sh(["git", "branch", "-r", "--format=%(refname:short)"]).splitlines()
    branches = [x.split("origin/", 1)[1] for x in refs
                if x.startswith("origin/") and "->" not in x and x != "origin/main"]
    now = time.time()
    rows = []
    for b in branches:
        meta = sh(["git", "log", "-1", "--format=%ct%x1f%ce", f"origin/{b}"])
        if not meta:
            continue
        cepoch, cemail = (meta.split("\x1f") + [""])[:2]
        age = int((now - int(cepoch)) // 86400)
        cdate = time.strftime("%Y-%m-%d", time.localtime(int(cepoch)))
        cherry = sh(["git", "cherry", "origin/main", f"origin/{b}"])
        ahead = sum(1 for ln in cherry.splitlines() if ln.startswith("+"))
        pr = sh(["gh", "pr", "list", "--repo", REPO, "--head", b, "--state", "all",
                 "--json", "number,state,author",
                 "--jq", "sort_by(.number)|last|\"\\(.number)\\^_\\(.state)\\^_\\(.author.login)\""])
        prnum, prstate, prauthor = (pr.split("\x1f") + ["", "none", ""])[:3] if pr else ("", "none", "")
        row = dict(branch=b, cdate=cdate, age=age, inmain=(ahead == 0),
                   prnum=prnum, prstate=prstate,
                   owner=owner(prauthor, cemail))
        row["bucket"] = bucket(row)
        rows.append(row)
    return rows


def refline(r, md, checkbox=False):
    pr = f"#{r['prnum']}" if r["prnum"] and r["prnum"] != "null" else "no PR"
    lossless = " **[in-main → lossless]**" if r["inmain"] else ""
    own = f"`{r['owner'][1]}` ({r['owner'][0]})"
    mark = "- [ ] " if checkbox else "- "
    return f"{mark}`{r['branch']}` — {r['cdate']} ({r['age']}d), {pr}, owner {own}{lossless}"


ACTIONABLE = {"R1", "R2", "R3", "R4", "R5"}


def report(rows, md=False):
    counts = Counter(r["bucket"] for r in rows)
    labels = {k: (lbl, eff, lik) for k, lbl, eff, lik in RANKS}
    out = []
    if md:
        out.append(f"Follow-up to #48797. Living worksheet auditing the **{len(rows)}** branches on "
                   "`origin` (excluding `main`), ordered **low-hanging fruit first**: "
                   "low effort to validate + high likelihood deletable at the top. "
                   "Actionable buckets use `- [ ]`; check items off as branches are deleted "
                   "(re-run the script to refresh the live list).\n")
        deleted = load_deleted()
        if deleted:
            def dline(d):
                pr = f"#{d['pr']}" if d["pr"] and d["pr"] != "-" else "no PR"
                return (f"- [x] `{d['branch']}` — deleted {d['date']}, was {d['bucket']}, "
                        f"{pr}, tip `{d['sha']}`")
            comp = [d for d in deleted if d["branch"].startswith("compliance/update-headers-batch-")]
            rest = [d for d in deleted if not d["branch"].startswith("compliance/update-headers-batch-")]
            out.append(f"## Deleted (record) — {len(deleted)} branches\n")
            out.append("_Recorded here because git cannot report deleted branches. "
                       "`tip` is the pre-deletion SHA (best-effort restore point)._\n")
            out += [dline(d) for d in rest]
            if comp:
                out.append(f"\n<details><summary>{len(comp)}× `compliance/update-headers-batch-*` "
                           "(bot copyright-header batches)</summary>\n")
                out += [dline(d) for d in comp]
                out.append("\n</details>")
            out.append("")
        out.append("## Ranking\n")
        out.append("| Rank | Bucket | Count | Effort to validate | Likelihood deletable |")
        out.append("|---|---|--:|---|---|")
        for k, lbl, eff, lik in RANKS:
            out.append(f"| {k} | {lbl} | {counts.get(k,0)} | {eff} | {lik} |")
        out.append("\n_Owner resolved from PR author + committer email vs roster; best-effort. "
                   "Current engineer maintainers: `@johnsonaj @YakDriver @gdavison @jar-b "
                   "@ewbankkit @subham-ibmhc @taruntej-a @brosas07`; also on team: "
                   "`@breathingdust @justinretzolk`. `release*` / `v*` cannot be deleted (org rule)._\n")
        for k, lbl, eff, lik in RANKS:
            group = sorted([r for r in rows if r["bucket"] == k], key=lambda x: -x["age"])
            if not group:
                continue
            cb = k in ACTIONABLE
            if k in ("R5", "R6"):
                out.append(f"<details><summary>{k} — {lbl} ({len(group)})</summary>\n")
                out += [refline(r, md, cb) for r in group]
                out.append("\n</details>\n")
            elif k == "R1":
                comp = [r for r in group if r["branch"].startswith("compliance/update-headers-batch-")]
                rest = [r for r in group if not r["branch"].startswith("compliance/update-headers-batch-")]
                out.append(f"## {k} — {lbl}\n")
                if comp:
                    out.append(f"- [ ] `compliance/update-headers-batch-*` — ~{comp[0]['cdate']}, "
                               f"PRs closed, bot-owned — **{len(comp)} branches**, bulk-deletable")
                out += [refline(r, md, cb) for r in rest]
                out.append("")
            else:
                out.append(f"## {k} — {lbl}\n")
                out += [refline(r, md, cb) for r in group]
                out.append("")
        out.append("### Caveats\n"
                   "- Owner attribution is best-effort; verify anything ambiguous before deleting.\n"
                   "- `release*` are protected by an org rule (cannot be deleted) — listed for completeness, not flagged.\n"
                   "- The PR check matches head-ref on this repo; a same-named fork PR is not detected.\n"
                   "- Regenerate with `.ci/scripts/audit-branches.py --markdown`.")
    else:
        for k, lbl, eff, lik in RANKS:
            group = sorted([r for r in rows if r["bucket"] == k], key=lambda x: -x["age"])
            if not group:
                continue
            out.append(f"\n== {k}: {lbl} ({len(group)}) ==")
            for r in group:
                out.append(refline(r, False))
    return "\n".join(out)


if __name__ == "__main__":
    md = "--markdown" in sys.argv
    print(report(collect(), md=md))
