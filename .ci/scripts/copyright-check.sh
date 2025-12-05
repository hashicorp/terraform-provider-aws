#!/bin/bash
set -euo pipefail

MISSING=0

check_file() {
	local file="$1"
	local comment_style="$2"
	local expected="$3"
	
	local first_line=$(head -1 "$file")
	
	# Skip generated files
	[[ "$first_line" =~ Code\ generated ]] && return 0
	
	# Skip shebang lines
	if [[ "$first_line" =~ ^#! ]]; then
		first_line=$(sed -n '2p' "$file")
	fi
	
	if [[ ! "$first_line" =~ $expected ]]; then
		echo "Missing or incorrect copyright header: $file"
		return 1
	fi
	return 0
}

# Check Go files
while IFS= read -r file; do
	check_file "$file" "//" "^// Copyright IBM Corp\. 2014, 2025$" || MISSING=$((MISSING + 1))
done < <(find . -name "*.go" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*")

# Check shell scripts
while IFS= read -r file; do
	check_file "$file" "#" "^# Copyright IBM Corp\. 2014, 2025$" || MISSING=$((MISSING + 1))
done < <(find . -name "*.sh" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")

# Check HCL/Terraform files
while IFS= read -r file; do
	check_file "$file" "#" "^# Copyright IBM Corp\. 2014, 2025$" || MISSING=$((MISSING + 1))
done < <(find . \( -name "*.hcl" -o -name "*.tf" \) -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")

# Check Python files
while IFS= read -r file; do
	check_file "$file" "#" "^# Copyright IBM Corp\. 2014, 2025$" || MISSING=$((MISSING + 1))
done < <(find . -name "*.py" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*")

# Check Markdown files (use HTML comments)
while IFS= read -r file; do
	check_file "$file" "<!--" "^<!-- Copyright IBM Corp\. 2014, 2025 -->$" || MISSING=$((MISSING + 1))
done < <(find . -name "*.md" -type f ! -path "./vendor/*" ! -path "./.ci/*" ! -path "./.github/*" ! -path "./.teamcity/*" ! -path "./.release/*" ! -path "./examples/*" ! -path "./website/*" ! -path "./CHANGELOG.md" ! -path "./README.md")

if [[ $MISSING -gt 0 ]]; then
	echo ""
	echo "Error: $MISSING files missing correct copyright header"
	exit 1
fi

echo "✓ All files have correct copyright headers"
