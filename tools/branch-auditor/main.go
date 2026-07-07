// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Command branch-auditor is a convenience tool for occasional cleanup of remote
// branches on origin. It ranks branches "low-hanging fruit first" (low effort to
// validate + high likelihood deletable at the top) and prints a GitHub-issue
// markdown worksheet. It is read-only and never deletes anything.
//
// It is not part of any CI workflow. Run it from this directory:
//
//	cd tools/branch-auditor && go run .
//
// Requires git and an authenticated gh CLI.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const repo = "hashicorp/terraform-provider-aws"

// Roster of current maintainers, loaded at runtime from docs/faq.md.
// Bucketing only distinguishes maintainer-vs-not, so a single set suffices.
var maintainers = map[string]bool{}

var (
	// committer email -> login (best-effort attribution for branches with no PR;
	// docs/faq.md lists logins but not commit emails, so this stays local).
	emailToLogin = map[string]string{
		"gdavison@hashicorp.com":                          "gdavison",
		"jared.baker@hashicorp.com":                       "jar-b",
		"kit_ewbank@hotmail.com":                          "ewbankkit",
		"adrian.johnson@hashicorp.com":                    "johnsonaj",
		"subham.mukhopadhyay@hashicorp.com":               "subham-ibmhc",
		"dirk.avery@gmail.com":                            "yakdriver",
		"yakdriver@users.noreply.github.com":              "yakdriver",
		"31492422+yakdriver@users.noreply.github.com":     "yakdriver",
		"tarunsraina483@gmail.com":                        "taruntej-a",
		"tarunteja@taruns-macbook-pro.local":              "taruntej-a",
		"sdavis@hashicorp.com":                            "breathingdust",
		"44710313+justinretzolk@users.noreply.github.com": "justinretzolk",
	}

	// Departed HashiCorp folks / former maintainers (no longer stewards).
	formerEmployees = []string{
		"thomas@thomaszwskismbp.lan", "thomas.zalewski@hashicorp.com",
		"sharon.nam@hashicorp.com", "sharon.nam@sharon.nam-pjcj4f7jh7",
		"jitendra.gangwar1@hashicorp.com",
	}
	botEmailMarkers = []string{"github-actions", "noreply@github.com", "changelogbot@hashicorp.com", "update-schemas@github.com"}

	protected = regexp.MustCompile(`^(release|v\d)`) // org rule protects release* and v* branches
)

// rank buckets, ordered: lower = delete sooner (low-hanging fruit first).
type rank struct{ key, label, effort, likelihood string }

var ranks = []rank{
	{"R1", "Automation/bot, no open PR", "trivial (bot, job done)", "very high"},
	{"R2", "Content already in `main` (patch-equivalent), no open PR", "trivial (content preserved)", "very high"},
	{"R3", "No current owner (departed/external), no open PR", "low–medium (skim commits)", "high"},
	{"R4", "Current maintainer/team, stale >1yr, no open PR", "medium (confirm w/ owner)", "medium"},
	{"R5", "Current maintainer/team, recent <1yr, no open PR", "medium", "low"},
	{"R6", "Active — open PR", "n/a", "keep"},
	{"X1", "Cannot delete — `release*` / `v*` (org rule)", "n/a", "n/a"},
	{"X2", "Protected infra — `gh-pages`", "n/a", "n/a"},
}

var actionable = []string{"R1", "R2", "R3", "R4", "R5"}

type branch struct {
	name           string
	date           string
	age            int
	inMain         bool
	prNum, prState string
	ownerStatus    string
	ownerWho       string
	bucket         string
}

// run executes a command and fails fast with stderr context on error. A silent
// git/gh failure could otherwise misclassify branches (e.g. an empty `git cherry`
// looking like "in main", or a failed `gh` call looking like "no PR") and produce
// a misleading worksheet.
func run(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		stderr := ""
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		fmt.Fprintf(os.Stderr, "branch-auditor: `%s %s` failed: %v\n%s\n", name, strings.Join(args, " "), err, stderr)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func git(args ...string) string { return run("git", args...) }

// loadMaintainers parses current maintainer logins from the "Who are the
// maintainers?" section of docs/faq.md (the authoritative roster).
func loadMaintainers(path string) map[string]bool {
	m := map[string]bool{}
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer f.Close()
	re := regexp.MustCompile(`github\.com/([A-Za-z0-9-]+)`)
	inSection := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "## ") {
			inSection = strings.Contains(strings.ToLower(line), "who are the maintainers")
			continue
		}
		if inSection {
			if mt := re.FindStringSubmatch(line); mt != nil {
				m[strings.ToLower(mt[1])] = true
			}
		}
	}
	return m
}

func statusOf(login string) string {
	if maintainers[strings.ToLower(login)] {
		return "maintainer"
	}
	return "other"
}

// emailStatus resolves a committer email to (status, who).
func emailStatus(email string) (string, string) {
	e := strings.ToLower(strings.TrimSpace(email))
	for _, m := range botEmailMarkers {
		if strings.Contains(e, m) {
			return "bot", "bot"
		}
	}
	if login, ok := emailToLogin[e]; ok {
		return statusOf(login), login
	}
	if slices.Contains(formerEmployees, e) {
		return "other", "former-employee"
	}
	return "", ""
}

// owner combines the PR author and committer email, preferring the most-current.
func owner(prAuthor, cemail string) (status, who string) {
	rankOf := map[string]int{"maintainer": 0, "other": 1, "bot": 2, "": 3}
	type cand struct{ st, who string }
	var cands []cand
	if prAuthor != "" && prAuthor != "null" {
		if strings.HasPrefix(prAuthor, "app/") {
			cands = append(cands, cand{"bot", prAuthor})
		} else {
			cands = append(cands, cand{statusOf(prAuthor), prAuthor})
		}
	}
	if st, w := emailStatus(cemail); st != "" {
		cands = append(cands, cand{st, w})
	}
	if len(cands) == 0 {
		return "unknown", "?"
	}
	sort.SliceStable(cands, func(i, j int) bool { return rankOf[cands[i].st] < rankOf[cands[j].st] })
	return cands[0].st, cands[0].who
}

func bucketOf(b branch) string {
	switch {
	case b.name == "gh-pages":
		return "X2"
	case protected.MatchString(b.name):
		return "X1"
	case b.prState == "OPEN":
		return "R6"
	case b.ownerStatus == "bot":
		return "R1"
	case b.inMain:
		return "R2"
	case b.ownerStatus == "other" || b.ownerStatus == "unknown":
		return "R3"
	case b.age > 365:
		return "R4"
	default:
		return "R5"
	}
}

func collect() []branch {
	var out []branch
	now := time.Now()
	for _, ref := range strings.Split(git("branch", "-r", "--format=%(refname:short)"), "\n") {
		name, ok := strings.CutPrefix(strings.TrimSpace(ref), "origin/")
		if !ok || name == "" || name == "HEAD" || name == "main" || strings.Contains(ref, "->") {
			continue
		}
		meta := git("log", "-1", "--format=%ct\x1f%ce", "origin/"+name)
		epochStr, cemail, _ := strings.Cut(meta, "\x1f")
		epoch, _ := strconv.ParseInt(epochStr, 10, 64)
		if epoch == 0 {
			continue
		}
		ct := time.Unix(epoch, 0)
		ahead := 0
		for _, ln := range strings.Split(git("cherry", "origin/main", "origin/"+name), "\n") {
			if strings.HasPrefix(ln, "+") {
				ahead++
			}
		}
		prNum, prState, prAuthor := prInfo(name)
		b := branch{
			name:    name,
			date:    ct.Format("2006-01-02"),
			age:     int(now.Sub(ct).Hours() / 24),
			inMain:  ahead == 0,
			prNum:   prNum,
			prState: prState,
		}
		b.ownerStatus, b.ownerWho = owner(prAuthor, cemail)
		b.bucket = bucketOf(b)
		out = append(out, b)
	}
	return out
}

func prInfo(head string) (num, state, author string) {
	// A genuine "no PR" result exits 0 with empty fields; an actual gh failure
	// (missing/unauthenticated/rate-limited) exits non-zero and is caught by run.
	s := run("gh", "pr", "list", "--repo", repo, "--head", head, "--state", "all",
		"--json", "number,state,author",
		"--jq", `sort_by(.number)|last|"\(.number)\u001f\(.state)\u001f\(.author.login)"`)
	if s == "" {
		return "", "none", ""
	}
	num, rest, _ := strings.Cut(s, "\x1f")
	state, author, _ = strings.Cut(rest, "\x1f")
	return num, state, author
}

type deleted struct{ date, name, sha, pr, bucket string }

func loadDeleted(path string) []deleted {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var recs []deleted
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) < 5 {
			continue
		}
		recs = append(recs, deleted{p[0], p[1], p[2], p[3], p[4]})
	}
	return recs
}

func prRef(num string) string {
	if num == "" || num == "null" || num == "-" {
		return "no PR"
	}
	return "#" + num
}

func main() {
	maintainers = loadMaintainers(filepath.Join(git("rev-parse", "--show-toplevel"), "docs", "faq.md"))
	if len(maintainers) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no maintainers parsed from docs/faq.md; owners will show as non-maintainers")
	}
	rows := collect()
	counts := map[string]int{}
	for _, b := range rows {
		counts[b.bucket]++
	}

	var out []string
	add := func(s string) { out = append(out, s) }

	add(fmt.Sprintf("Follow-up to #48797. Living worksheet auditing the **%d** branches on `origin` "+
		"(excluding `main`), ordered **low-hanging fruit first**: low effort to validate + high "+
		"likelihood deletable at the top. Actionable buckets use `- [ ]`. Regenerate with "+
		"`cd tools/branch-auditor && go run .`.\n", len(rows)))

	// Deleted record (ledger survives regeneration; git cannot report deleted branches).
	if del := loadDeleted("deleted-branches.tsv"); len(del) > 0 {
		add(fmt.Sprintf("## Deleted (record) — %d branches\n", len(del)))
		add("_`tip` is the pre-deletion SHA (best-effort restore point)._\n")
		var comp []deleted
		for _, d := range del {
			if strings.HasPrefix(d.name, "compliance/update-headers-batch-") {
				comp = append(comp, d)
				continue
			}
			add(fmt.Sprintf("- [x] `%s` — deleted %s, was %s, %s, tip `%s`", d.name, d.date, d.bucket, prRef(d.pr), d.sha))
		}
		if len(comp) > 0 {
			add(fmt.Sprintf("\n<details><summary>%d× `compliance/update-headers-batch-*` (bot copyright-header batches)</summary>\n", len(comp)))
			for _, d := range comp {
				add(fmt.Sprintf("- [x] `%s` — deleted %s, was %s, %s, tip `%s`", d.name, d.date, d.bucket, prRef(d.pr), d.sha))
			}
			add("\n</details>")
		}
		add("")
	}

	add("## Ranking\n")
	add("| Rank | Bucket | Count | Effort to validate | Likelihood deletable |")
	add("|---|---|--:|---|---|")
	for _, r := range ranks {
		add(fmt.Sprintf("| %s | %s | %d | %s | %s |", r.key, r.label, counts[r.key], r.effort, r.likelihood))
	}
	add("\n_Owner resolved from PR author + committer email vs the maintainer list in `docs/faq.md`; best-effort. `release*` / `v*` cannot be deleted (org rule)._\n")

	for _, r := range ranks {
		group := filterSort(rows, r.key)
		if len(group) == 0 {
			continue
		}
		cb := slices.Contains(actionable, r.key)
		if r.key == "R5" || r.key == "R6" {
			add(fmt.Sprintf("<details><summary>%s — %s (%d)</summary>\n", r.key, r.label, len(group)))
			for _, b := range group {
				add(line(b, cb))
			}
			add("\n</details>\n")
			continue
		}
		add(fmt.Sprintf("## %s — %s\n", r.key, r.label))
		for _, b := range group {
			add(line(b, cb))
		}
		add("")
	}
	fmt.Println(strings.Join(out, "\n"))
}

func filterSort(rows []branch, key string) []branch {
	var g []branch
	for _, b := range rows {
		if b.bucket == key {
			g = append(g, b)
		}
	}
	sort.SliceStable(g, func(i, j int) bool { return g[i].age > g[j].age })
	return g
}

func line(b branch, checkbox bool) string {
	mark := "- "
	if checkbox {
		mark = "- [ ] "
	}
	inMainMark := ""
	if b.inMain {
		inMainMark = " **[content in `main`]**"
	}
	return fmt.Sprintf("%s`%s` — %s (%dd), %s, owner `%s` (%s)%s",
		mark, b.name, b.date, b.age, prRef(b.prNum), b.ownerWho, b.ownerStatus, inMainMark)
}
