#!/bin/bash
set -euo pipefail

FIXED=0
ADDED=0
PKG="${PKG:-}"

# Build path filter based on PKG
if [[ -n "$PKG" ]]; then
	PATH_FILTER="./internal/service/$PKG"
	echo "Fixing copyright headers for service: $PKG"
else
	PATH_FILTER="."
fi

fix_go_file() {
	local file="$1"
	local first_line=$(head -1 "$file")
	
	# Skip generated files
	[[ "$first_line" =~ ^//\ Code\ generated ]] && return 0
	
	# Fix HashiCorp headers
	if [[ "$first_line" =~ ^//\ Copyright.*HashiCorp ]]; then
		perl -i -pe 's|^// Copyright.*HashiCorp.*$|// Copyright IBM Corp. 2014, 2025|' "$file"
		echo "Fixed: $file"
		FIXED=$((FIXED + 1))
		return 0
	fi
	
	# Add missing header
	if [[ ! "$first_line" =~ ^//\ Copyright\ IBM\ Corp ]]; then
		{ echo "// Copyright IBM Corp. 2014, 2025"; echo "// SPDX-License-Identifier: MPL-2.0"; echo ""; cat "$file"; } > "$file.tmp"
		mv "$file.tmp" "$file"
		echo "Added: $file"
		ADDED=$((ADDED + 1))
	fi
}

fix_hash_comment_file() {
	local file="$1"
	local has_shebang=false
	local first_line=$(head -1 "$file")
	
	# Check for shebang
	if [[ "$first_line" =~ ^#! ]]; then
		has_shebang=true
		first_line=$(sed -n '2p' "$file")
	fi
	
	# Skip generated files
	[[ "$first_line" =~ Code\ generated ]] && return 0
	
	# Fix HashiCorp headers
	if [[ "$first_line" =~ Copyright.*HashiCorp ]]; then
		if $has_shebang; then
			perl -i -pe 's|^# Copyright.*HashiCorp.*$|# Copyright IBM Corp. 2014, 2025| if $. == 2' "$file"
		else
			perl -i -pe 's|^# Copyright.*HashiCorp.*$|# Copyright IBM Corp. 2014, 2025| if $. == 1' "$file"
		fi
		echo "Fixed: $file"
		FIXED=$((FIXED + 1))
		return 0
	fi
	
	# Fix wrong comment style (// in shell/hcl/tf/py)
	if [[ "$first_line" =~ ^//\ Copyright ]]; then
		if $has_shebang; then
			perl -i -pe 's|^// Copyright IBM Corp\. 2014, 2025$|# Copyright IBM Corp. 2014, 2025| if $. == 2; s|^// SPDX-License-Identifier: MPL-2.0$|# SPDX-License-Identifier: MPL-2.0| if $. == 3' "$file"
		else
			perl -i -pe 's|^// Copyright IBM Corp\. 2014, 2025$|# Copyright IBM Corp. 2014, 2025| if $. == 1; s|^// SPDX-License-Identifier: MPL-2.0$|# SPDX-License-Identifier: MPL-2.0| if $. == 2' "$file"
		fi
		echo "Fixed comment style: $file"
		FIXED=$((FIXED + 1))
		return 0
	fi
	
	# Add missing header
	if [[ ! "$first_line" =~ Copyright\ IBM\ Corp ]]; then
		if $has_shebang; then
			{ head -1 "$file"; echo "# Copyright IBM Corp. 2014, 2025"; echo "# SPDX-License-Identifier: MPL-2.0"; echo ""; tail -n +2 "$file"; } > "$file.tmp"
		else
			{ echo "# Copyright IBM Corp. 2014, 2025"; echo "# SPDX-License-Identifier: MPL-2.0"; echo ""; cat "$file"; } > "$file.tmp"
		fi
		mv "$file.tmp" "$file"
		echo "Added: $file"
		ADDED=$((ADDED + 1))
	fi
}

# Fix Go files
while IFS= read -r file; do
	fix_go_file "$file"
done < <(find "$PATH_FILTER" -name "*.go" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" 2>/dev/null || true)

# Only fix non-Go files if checking everything
if [[ -z "$PKG" ]]; then
	# Fix shell scripts
	while IFS= read -r file; do
		fix_hash_comment_file "$file"
	done < <(find . -name "*.sh" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")

	# Fix HCL/Terraform files
	while IFS= read -r file; do
		fix_hash_comment_file "$file"
	done < <(find . \( -name "*.hcl" -o -name "*.tf" \) -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")

	# Fix Python files
	while IFS= read -r file; do
		fix_hash_comment_file "$file"
	done < <(find . -name "*.py" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")
fi

if [[ $FIXED -eq 0 && $ADDED -eq 0 ]]; then
	echo "✓ No files needed fixing"
else
	echo ""
	[[ $FIXED -gt 0 ]] && echo "✓ Fixed $FIXED files"
	[[ $ADDED -gt 0 ]] && echo "✓ Added headers to $ADDED files"
fi
