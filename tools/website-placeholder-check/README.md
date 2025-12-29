# Website Placeholder Check

This tool checks for the presence of the placeholder value `example_id_arg` in resource documentation files. This placeholder is used in the skaff template (`skaff/resource/websitedoc.gtpl`) and should be replaced with actual import ID examples in the generated documentation.

## Usage

```bash
go run main.go <directory>
```

### Examples

Check the resource documentation:

```bash
go run main.go ../../website/docs/r
```

Check a specific file:

```bash
go run main.go ../../website/docs/r/s3_bucket.html.markdown
```

## What it checks

The linter scans all `.md` and `.markdown` files in the specified directory and its subdirectories for the literal string `example_id_arg`. When found, it reports:

- The file path and line number
- The content of the line containing the placeholder

## Exit codes

- `0`: No placeholders found (success)
- `1`: Placeholders found or error occurred (failure)

## Integration

This linter can be integrated into CI/CD pipelines to ensure that documentation doesn't contain unresolved placeholders from the skaff template.

### Example Makefile target

```makefile
website-placeholder-check: ## Check for unresolved documentation placeholders
	@echo "make: Website Checks / example_id_arg placeholder check..."
	@cd tools/website-placeholder-check && go run main.go ../../website/docs/r
```
