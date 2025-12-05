# Copyright Header Migration: HashiCorp → IBM

## Overview

Migrate ~6,300 Go files from:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
```

To:
```go
// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0
```

## Systematic Approach

### Phase 1: Preparation & Validation

1. **Create a feature branch**
   ```bash
   git checkout -b copyright-ibm-migration
   ```

2. **Verify current state**
   ```bash
   rg "^// Copyright \(c\) HashiCorp, Inc\." --type go -c | wc -l
   # Should show ~6,300 files
   ```

3. **Test on a single file first**
   ```bash
   # Backup a test file
   cp main.go main.go.backup
   
   # Test the replacement
   perl -i -pe 's|^// Copyright \(c\) HashiCorp, Inc\.$|// Copyright IBM Corp. 2014, 2025|' main.go
   
   # Verify the change
   head -5 main.go
   
   # Restore if needed
   mv main.go.backup main.go
   ```

### Phase 2: Bulk Update

4. **Run dry-run to preview changes**
   ```bash
   ./update-copyright.sh --dry-run
   ```

5. **Execute the bulk update**
   ```bash
   ./update-copyright.sh
   ```

6. **Verify the changes**
   ```bash
   # Check that old format is gone
   rg "Copyright \(c\) HashiCorp, Inc\." --type go
   # Should return no results
   
   # Check new format is present
   rg "^// Copyright IBM Corp\. 2014, 2025" --type go -c | wc -l
   # Should show ~6,300 files
   
   # Spot check a few files
   head -5 main.go
   head -5 internal/service/s3/bucket.go
   ```

### Phase 3: Update Tooling

7. **Update copywrite configuration**
   ```bash
   mv .copywrite.hcl.new .copywrite.hcl
   ```

8. **Test copywrite tool behavior**
   ```bash
   # The copywrite tool may not recognize the new format
   # and might try to add HashiCorp headers again
   copywrite headers --plan
   ```

   **IMPORTANT**: If copywrite tries to add headers, you have two options:
   
   **Option A**: Disable the copyright CI check temporarily
   - Comment out the copyright job in `.github/workflows/copyright.yml`
   - Document this in the PR
   
   **Option B**: Fork/modify copywrite tool
   - More complex, not recommended for this migration

9. **Run formatting**
   ```bash
   make fmt
   ```

### Phase 4: Validation

10. **Check for any issues**
    ```bash
    # Look for any malformed headers
    rg "^// Copyright" --type go | rg -v "^[^:]+:// Copyright IBM Corp\. 2014, 2025$" | head -20
    
    # Check SPDX lines are intact
    rg "SPDX-License-Identifier: MPL-2.0" --type go -c | wc -l
    # Should still show ~6,300 files
    ```

11. **Run a subset of tests**
    ```bash
    # Quick sanity check
    make test T=TestAccS3Bucket_basic K=s3
    ```

12. **Review git diff statistics**
    ```bash
    git diff --stat
    git diff --shortstat
    ```

### Phase 5: Commit & CI

13. **Stage and commit changes**
    ```bash
    git add -A
    git commit -m "Update copyright headers from HashiCorp to IBM
    
    - Changed copyright format from 'Copyright (c) HashiCorp, Inc.' to 'Copyright IBM Corp. 2014, 2025'
    - Updated .copywrite.hcl configuration
    - Affected ~6,300 Go files
    - SPDX-License-Identifier remains MPL-2.0"
    ```

14. **Push and create PR**
    ```bash
    git push origin copyright-ibm-migration
    ```

15. **Monitor CI**
    - The copyright check will likely fail if copywrite tool is not updated
    - Document this in the PR description
    - May need to update CI workflow to skip copyright check or use different tool

## Rollback Plan

If issues are discovered:

```bash
# Revert the commit
git revert HEAD

# Or reset if not pushed
git reset --hard HEAD~1

# Or restore from backup branch
git checkout main
git branch -D copyright-ibm-migration
```

## Known Issues & Solutions

### Issue 1: copywrite tool adds HashiCorp headers back

**Solution**: Update `.github/workflows/copyright.yml` to either:
- Skip the check temporarily
- Use a different copyright checking tool
- Pass custom flags to copywrite: `--copyright-holder "IBM Corp."`

### Issue 2: Generated files get updated

**Solution**: Add generated file patterns to `.copywrite.hcl` header_ignore list

### Issue 3: Some files have different copyright formats

**Solution**: Run additional targeted replacements:
```bash
# Find any variations
rg "Copyright.*HashiCorp" --type go | rg -v "Copyright \(c\) HashiCorp, Inc\."

# Handle them case-by-case
```

## Verification Checklist

- [ ] All HashiCorp copyright headers replaced
- [ ] SPDX license identifiers intact
- [ ] .copywrite.hcl updated
- [ ] make fmt passes
- [ ] Sample tests pass
- [ ] Git diff reviewed
- [ ] CI workflow updated (if needed)
- [ ] PR created with clear description
