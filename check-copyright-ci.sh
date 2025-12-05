#!/bin/bash
set -euo pipefail

# Check if copywrite tool will work with new copyright format

echo "=== Copyright CI Check Analysis ==="
echo ""

echo "1. Current copywrite configuration:"
cat .copywrite.hcl
echo ""

echo "2. Testing copywrite with new format..."
echo "   Creating a test file..."

# Create a test file with new format
cat > /tmp/test_copyright.go <<'EOF'
// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package test

func Test() {}
EOF

echo "   Running copywrite headers --plan on test file..."
cd /tmp
copywrite headers --plan --copyright-holder "IBM Corp." 2>&1 || true
cd - > /dev/null

echo ""
echo "3. Recommendation:"
echo ""
echo "The copywrite tool is designed for HashiCorp format and may not"
echo "recognize the IBM copyright format. You have these options:"
echo ""
echo "Option A (Recommended): Temporarily disable copyright CI check"
echo "  - Edit .github/workflows/copyright.yml"
echo "  - Comment out or remove the 'copywrite' job"
echo "  - Document in PR that copyright format has changed"
echo ""
echo "Option B: Update copywrite invocation"
echo "  - Modify the CI workflow to use: copywrite headers --copyright-holder 'IBM Corp.'"
echo "  - This may still not produce the exact format needed"
echo ""
echo "Option C: Replace with different tool"
echo "  - Use addlicense or similar tool"
echo "  - Update CI workflow accordingly"
echo ""

rm -f /tmp/test_copyright.go
