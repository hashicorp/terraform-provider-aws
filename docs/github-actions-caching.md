<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# GitHub Actions Caching Strategy

This document explains the caching strategy used in the Terraform AWS Provider's GitHub Actions workflows and why it's designed this way.

## The Problem

The Terraform AWS Provider is a massive codebase with unique caching challenges:

- **261 services** with complex interdependencies
- **30-50 AWS SDK package updates per week**
- **500+ active pull requests** at any given time
- **8+ workflows** that compile Go code
- **10GB GitHub Actions cache limit** for the entire repository

### Why `internal/**` in Cache Keys Doesn't Work

A common pattern is to include source code in cache keys:

```yaml
key: ${{ runner.os }}-GOCACHE-${{ hashFiles('go.sum') }}-${{ hashFiles('internal/**') }}
```

**This creates catastrophic cache thrashing:**

```
8 workflows × 500 PRs × 8GB cache = 32,000 GB demand
GitHub limit: 10 GB (per repo)
Result: 0.03% cache hit rate (constant misses)
```

Every PR changes `internal/**`, creating a unique cache key. With hundreds of PRs, caches are constantly evicted before they can be reused.

### Why `go.sum` in Cache Keys Is Problematic

Including `go.sum` in the cache key seems logical but causes issues:

```yaml
key: ${{ runner.os }}-go-build-${{ hashFiles('go.sum') }}
```

**Problems:**

- AWS SDK updates 30-50 packages/week
- Each update changes `go.sum` → new cache key → full recompile
- Wastes the 90% of packages that didn't change

**Go's build cache is self-invalidating** - it automatically detects when dependencies change and only recompiles affected packages. Including `go.sum` in the key defeats this optimization.

## The Solution: Daily Rotation with Shared Cache

### Cache Key Strategy

```yaml
key: ${{ runner.os }}-go-build-${{ env.CACHE_DATE }}
restore-keys: |
  ${{ runner.os }}-go-build-
```

Where `CACHE_DATE=$(date +%Y-%m-%d)`

**Why this works:**

- **One cache per day** (not per PR or per commit)
- **All PRs share the same cache** on a given day
- **Daily rotation** prevents unbounded growth
- **Restore-keys** provide fallback to yesterday's cache (`restore-keys` is prefix, i.e., `${{ runner.os }}-go-build-*`, GitHub returns most recent match)
- **Go's internal cache** handles incremental compilation

### Cache Architecture

```
┌─────────────────┐
│  go_build job   │  ← Only job that SAVES cache
│  (provider.yml) │
└────────┬────────┘
         │ saves
         ▼
    ┌─────────┐
    │  Cache  │  8GB, daily rotation
    │ Storage │  key: go-build-2025-12-15
    └────┬────┘
         │ restores (read-only)
         ▼
┌────────────────────────────────┐
│ All other jobs restore cache:  │
│ - go_generate                  │
│ - go_test                      │
│ - import-lint                  │
│ - validate_sweepers            │
│ - copyright                    │
│ - dependencies                 │
│ - modern_go                    │
│ - providerlint                 │
│ - pull_request_target          │
│ - skaff                        │
│ - smarterr                     │
└────────────────────────────────┘
```

### Single Producer Pattern

**Only `provider.yml`'s `go_build` job saves cache:**

```yaml
- name: Save Go Build Cache
  uses: actions/cache/save@v5.0.1
  if: always() && steps.cache-go-build.outputs.cache-hit != 'true'
  with:
    path: ${{ env.GOCACHE }}
    key: ${{ runner.os }}-go-build-${{ env.CACHE_DATE }}
```

**All other jobs restore-only:**

```yaml
- name: Restore Go Build Cache
  uses: actions/cache/restore@v5.0.1
  with:
    path: ${{ env.GOCACHE }}
    key: ${{ runner.os }}-go-build-${{ env.CACHE_DATE }}
    restore-keys: |
      ${{ runner.os }}-go-build-
```

**Benefits:**

- Prevents race conditions
- Ensures consistency
- Reduces cache save time
- Avoids duplicate cache entries

## Implementation Details

### Setting Up Cache Date

All jobs that use caching must set `CACHE_DATE`:

```yaml
- name: go env
  run: |
    echo "GOCACHE=$(go env GOCACHE)" >> $GITHUB_ENV
    echo "CACHE_DATE=$(date +%Y-%m-%d)" >> $GITHUB_ENV
```

### Cache Cleanup in Tests

The `go_test` job includes cleanup to prevent test artifacts from bloating the cache:

```yaml
- name: Cleanup Test Artifacts
  if: always()
  run: |
    if [ -d "$GOCACHE" ]; then
      # Remove test binaries - huge and rarely reused
      find $GOCACHE -name "*.test" -type f -delete 2>/dev/null || true

      # Remove entries older than 2 days
      find $GOCACHE -type f -mtime +2 -delete 2>/dev/null || true
      find $GOCACHE -type d -empty -delete 2>/dev/null || true
    fi
```

### Dependency Cache

The `go/pkg/mod` cache uses a different strategy since dependencies are stable:

```yaml
- uses: actions/cache@v5.0.1
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-pkg-mod-${{ hashFiles('go.sum') }}
```

This cache:

- **Does** use `go.sum` in the key (dependencies change infrequently)
- Is shared across all workflows
- Typically ~2GB
- Rarely invalidated

## Expected Results

### Cache Performance

| Metric | Before | After |
|--------|--------|-------|
| Cache demand | 32,000 GB | 10 GB |
| Cache hit rate | 0.03% | 80-90% |
| Build time | 10-15 min | 2-3 min |
| Cache stability | Constant thrashing | Stable |

### Daily Workflow

**First run of the day:**

- Cold cache (or restores yesterday's)
- Full compilation: ~10 minutes
- Saves new cache for the day

**Subsequent PRs same day:**

- Warm cache hit
- Incremental compilation: ~2-3 minutes
- No cache save (already exists)

**Next day:**

- New cache key (new date)
- Fresh start prevents unbounded growth
- Old cache auto-expires after 7 days

## Local Development

The same strategy is used in the `GNUmakefile` for local testing:

```bash
# On macOS (with CrowdStrike), uses temp cache to avoid scanning
make test-fast

# Automatically detects:
# - macOS: Uses /tmp cache to avoid security software overhead
# - Linux: Uses default cache location
```

See [Makefile Cheat Sheet](makefile-cheat-sheet.md) for details.

## Monitoring

Monitor cache effectiveness in GitHub Actions:

1. **Cache hit rate**: Check workflow logs for "Cache restored from key"
2. **Build times**: Compare first run of day vs. subsequent runs
3. **Cache size**: Should stay around 8-10GB total

If cache hit rates drop below 70%, investigate:

- Are multiple workflows saving cache? (should only be `go_build`)
- Is cache size approaching 10GB limit?
- Are there new workflows not following the pattern?

## Troubleshooting

### Cache Miss on Same Day

**Symptom:** PR shows cache miss even though another PR ran earlier same day.

**Cause:** Different runner OS or cache was evicted due to size limits.

**Solution:** This is expected occasionally. The restore-keys will fall back to a recent cache.

### Cache Size Growing

**Symptom:** Cache approaching 10GB limit.

**Cause:** Test artifacts or stale entries accumulating.

**Solution:** The cleanup step in `go_test` should handle this. If not, adjust cleanup thresholds.

### Slow Builds Despite Cache Hit

**Symptom:** Cache hit but build still takes 10+ minutes.

**Cause:** Major dependency update invalidated most of Go's internal cache.

**Solution:** This is expected after large AWS SDK updates. Subsequent builds will be fast.

## References

- [GitHub Actions Cache Documentation](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
- [Go Build Cache](https://pkg.go.dev/cmd/go#hdr-Build_and_test_caching)
- [Makefile Cheat Sheet](makefile-cheat-sheet.md)
- [Continuous Integration](continuous-integration.md)
