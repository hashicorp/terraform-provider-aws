# smarterr-parser

Finds and optionally fixes non-smarterr error patterns in Terraform AWS Provider resource/datasource files.

## Usage

```bash
go build .

# Analyze only
./smarterr-parser <folder_path>

# Analyze and auto-fix bare return statements
./smarterr-parser <folder_path> --fix
```

## Output

Lists files and line numbers with non-smarterr error patterns:

```
/path/to/file1.go:134
/path/to/file1.go:139
/path/to/file2.go:82
```

## Auto-fix

The `--fix` flag automatically wraps bare `return nil, err` statements with `smarterr.NewError(err)` and adds the required import.
