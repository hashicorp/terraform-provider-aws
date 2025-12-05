# Copyright Migration Summary

## Quick Start

```bash
# 1. Dry run to preview
./update-copyright.sh --dry-run

# 2. Execute the migration
./update-copyright.sh

# 3. Update config
mv .copywrite.hcl.new .copywrite.hcl

# 4. Format code
make fmt

# 5. Verify
rg '^// Copyright.*HashiCorp' --type go  # Should be empty
```

## What's Changing

- **From**: `// Copyright (c) HashiCorp, Inc.`
- **To**: `// Copyright IBM Corp. 2014, 2025`
- **Files affected**: ~6,300 Go files
- **SPDX license**: Remains `MPL-2.0` (unchanged)

## Files Created

1. **update-copyright.sh** - Main migration script
2. **COPYRIGHT-MIGRATION.md** - Detailed step-by-step guide
3. **COPYRIGHT-SUMMARY.md** - This file (quick reference)
4. **check-copyright-ci.sh** - CI compatibility checker
5. **.copywrite.hcl.new** - Updated copywrite configuration

## Critical Issue: CI Workflow

The `copywrite` tool is HashiCorp's tool and may not recognize the new IBM format.

**Impact**: The `.github/workflows/copyright.yml` CI check will likely fail.

**Solution**: You must choose one:

### Option A: Disable Copyright CI Check (Fastest)

Edit `.github/workflows/copyright.yml` and comment out the job:

```yaml
jobs:
  # copywrite:
  #   name: add headers check
  #   runs-on: ubuntu-latest
  #   steps:
  #     ...
```

### Option B: Update CI to Use Custom Format

Modify the workflow to pass custom flags:

```yaml
- run: copywrite headers --copyright-holder "IBM Corp."
```

**Warning**: This may still not produce the exact format `// Copyright IBM Corp. 2014, 2025`

### Option C: Replace with Different Tool

Replace `copywrite` with `addlicense` or similar:

```bash
go install github.com/google/addlicense@latest
addlicense -c "IBM Corp." -y "2014-2025" .
```

## Execution Order

1. ✅ Create feature branch
2. ✅ Run `./update-copyright.sh --dry-run`
3. ✅ Run `./update-copyright.sh`
4. ✅ Run `mv .copywrite.hcl.new .copywrite.hcl`
5. ⚠️  **DECIDE**: Update or disable copyright CI workflow
6. ✅ Run `make fmt`
7. ✅ Review with `git diff --stat`
8. ✅ Commit and push
9. ⚠️  **MONITOR**: CI may fail on copyright check

## Verification Commands

```bash
# Count old format (should be 0)
rg '^// Copyright.*HashiCorp' --type go | wc -l

# Count new format (should be ~6,300)
rg '^// Copyright IBM Corp\. 2014, 2025' --type go | wc -l

# Check SPDX intact (should be ~6,300)
rg 'SPDX-License-Identifier: MPL-2.0' --type go | wc -l

# Spot check files
head -5 main.go
head -5 internal/service/s3/bucket.go
```

## Rollback

```bash
git reset --hard HEAD~1  # If not pushed
git revert HEAD          # If already pushed
```

## Timeline Estimate

- Script execution: ~30 seconds
- Verification: ~5 minutes
- CI workflow update: ~10 minutes
- Code review: ~15 minutes
- **Total**: ~30 minutes

## Risk Assessment

- **Low Risk**: The change is purely cosmetic (comments only)
- **No functional impact**: Code behavior unchanged
- **Main risk**: CI workflow compatibility
- **Mitigation**: Test in feature branch first

## Questions?

See `COPYRIGHT-MIGRATION.md` for detailed documentation.
