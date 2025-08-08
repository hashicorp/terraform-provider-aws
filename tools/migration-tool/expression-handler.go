package main

import (
	"fmt"
	"strings"
)

const (
	smarterrImport = "\"github.com/!yak!driver/smarterr\""
	smerrImport    = "\"github.com/!yak!driver/smerr\""
	fmtImport      = "\"fmt\""
)

type errorHandler func(string, map[string]bool) (string, error)

func transformAssertSingleValueResult(call string, imports map[string]bool) string {
	imports[smarterrImport] = true
	return "return smarterr.Assert(tfresource.AssertSingleValueResult(...))"
}

func processCaseReturnfmtError(line string, imports map[string]bool) (string, error) {
	imports[fmtImport] = true
	imports[smarterrImport] = true
	// Example: return nil, fmt.Errorf("unexpected format for ID (%v), expected more than one part", idParts)
	if !strings.Contains(line, "fmt.Errorf(") {
		fmt.Println("Not a fmt.Errorf return line")
		return line, nil
	}
	// Extract the fmt.Errorf arguments
	start := strings.Index(line, "fmt.Errorf(")
	if start == -1 {
		fmt.Println("No fmt.Errorf found")
		return line, nil
	}
	argsText := line[start+len("fmt.Errorf(") : strings.LastIndex(line, ")")]
	// Build the smarterr.Errorf replacement
	newLine := "return nil, smarterr.Errorf(" + argsText + ")"
	return newLine, nil
}

// sdkdiag.AppendFromErr(diags, err) -> smerr.Append(ctx, diags, err, smerr.ID, ...)
func processAppendFromErr(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	args := extractFunctionArgs(line)
	if len(args) < 2 {
		return "", fmt.Errorf("not enough arguments in sdkdiag.AppendFromErr: %s", line)
	}
	return fmt.Sprintf("smerr.Append(ctx, diags, smerr.ID, %s)", args[1]), nil
}

// sdkdiag.AppendErrorf(diags, ..., err) -> smerr.Append(ctx, diags, err, smerr.ID, ...)
func processAppendErrorf(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	args := extractFunctionArgs(line)
	if len(args) < 3 {
		return "", fmt.Errorf("not enough arguments in sdkdiag.AppendErrorf: %s", line)
	}
	return fmt.Sprintf("smerr.Append(ctx, diags, err, smerr.ID, %s)", args[2]), nil
}

// create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, resNameFleet, d.Get(names.AttrName).(string), err) → smerr.Append(ctx, diags, err, smerr.ID, d.Get(names.AttrName).(string))
func processAppendDiagError(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	return "smerr.Append(ctx, diags, err, smerr.ID, d.Get(names.AttrName).(string))", nil
}

// response.Diagnostics.AddError("creating EC2 EBS Fast Snapshot Restore", err.Error()) → smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())
func processResponseDiagnosticsAddError(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	return fmt.Sprintf("smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())"), nil
}

// resp.Diagnostics.AddError(create.ProblemStandardMessage(..., err), err.Error()) → smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ...)
func processRespDiagnosticsAddError(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	argsLevelOne := extractFunctionArgs(line)
	if len(argsLevelOne) < 2 {
		return "", fmt.Errorf("not enough arguments in resp.Diagnostics.AddError: %s", line)
	}
	argsLevelTwo := extractFunctionArgs(argsLevelOne[1])
	if len(argsLevelTwo) < 2 {
		return "", fmt.Errorf("not enough arguments in create.ProblemStandardMessage: %s", argsLevelOne[1])
	}
	argsLevelTwoString := strings.Join(argsLevelTwo, ", ")
	return fmt.Sprintf("smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, %s)", argsLevelTwoString), nil
}

// response.Diagnostics.Append
func processResponseDiagnosticsAppend(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	args := extractFunctionArgs(line)
	fmt.Println(args, len(args), "ddddddd")
	if len(args) < 1 {
		return "", fmt.Errorf("not enough arguments in response.Diagnostics.Append: %s", line)
	}
	return fmt.Sprintf("smerr.EnrichAppend(ctx, &response.Diagnostics, %s)", strings.Join(args, ", ")), nil
}

// create.AddError(&response.Diagnostics, ..., err) -> smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ...)
func processCreateAddError(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	args := extractFunctionArgs(line)
	if len(args) < 3 {
		return "", fmt.Errorf("not enough arguments in create.AddError: %s", line)
	}
	filteredArgs := make([]string, 0, len(args)-2)
	for _, arg := range args {
		if arg != "&response.Diagnostics" && !strings.Contains(arg, "err") {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	fmt.Printf("filteredArgs: %v\n", filteredArgs)
	return fmt.Sprintf("smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, %s)", args[2]), nil
}

// return nil, err -> return nil, smarterr.NewError(err)
func processNakedErrorReturn(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true
	return "return nil, smarterr.NewError(err)", nil
}

// return nil, &retry.NotFoundError{ LastError: err, LastRequest: ..., } -> return nil, smarterr.NewError(&retry.NotFoundError{ LastError: err, LastRequest: ..., })
func processRetryNotFoundErrorReturn(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true

	// For return statements, we need to parse differently than function calls
	// Look for the pattern after "return "
	returnPrefix := "return "
	if !strings.HasPrefix(strings.TrimSpace(line), returnPrefix) {
		return "", fmt.Errorf("expected return statement, got: %s", line)
	}

	// Extract everything after "return "
	afterReturn := strings.TrimSpace(line[strings.Index(line, returnPrefix)+len(returnPrefix):])

	// Find the &retry.NotFoundError{ part and extract the complete struct
	retryIndex := strings.Index(afterReturn, "&retry.NotFoundError{")
	if retryIndex == -1 {
		return "", fmt.Errorf("could not find &retry.NotFoundError{ in return statement: %s", line)
	}

	// Extract everything from &retry.NotFoundError{ to the end
	structPart := afterReturn[retryIndex:]

	// Replace the struct with smarterr.NewError(struct)
	result := fmt.Sprintf("return nil, smarterr.NewError(%s)", structPart)
	return result, nil
}

// return tfresource.AssertSingleValueResult(...) -> return smarterr.Assert(tfresource.AssertSingleValueResult(...))
func processAssertSingleValueResultReturn(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true
	// Remove leading 'return ' if present
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "return ") {
		trimmed = strings.TrimSpace(trimmed[len("return "):])
	}
	return "return smarterr.Assert(tfresource.AssertSingleValueResult(" + trimmed + "))", nil
}

// return nil, tfresource.NewEmptyResultError(...) -> return nil, smarterr.NewError(tfresource.NewEmptyResultError(...))
func processReturnNewEmptyResultError(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "return nil, ") {
		trimmed = strings.TrimSpace(trimmed[len("return nil, "):])
	}
	return "return nil, smarterr.NewError(" + trimmed + ")", nil
}

// return nil, tfresource.NewEmptyResultError(...) -> return nil, smarterr.NewError(tfresource.NewEmptyResultError(...))
func processTfResourceNewEmptyResultError(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true
	start := strings.Index(line, "tfresource.NewEmptyResultError(")
	if start == -1 {
		return line, nil // not a match
	}
	end := strings.LastIndex(line, ")")
	if end == -1 || end < start {
		return line, nil // not a match
	}
	args := line[start+len("tfresource.NewEmptyResultError(") : end]
	return fmt.Sprintf("return nil, smarterr.NewError(tfresource.NewEmptyResultError(%s))", args), nil
}

// return tfresource.AssertSingleValueResult(...) -> return smarterr.Assert(tfresource.AssertSingleValueResult(...))
func processTFresourceAssertSingleValueResult(line string, imports map[string]bool) (string, error) {
	imports[smarterrImport] = true
	return fmt.Sprintf("return smarterr.Assert(tfresource.AssertSingleValueResult(%s))", line), nil
}

// resp.Diagnostics.Append(...) -> smerr.EnrichAppend(ctx, &resp.Diagnostics, ...)
func processDiagnosticsAppend(line string, imports map[string]bool) (string, error) {
	imports[smerrImport] = true
	args := extractFunctionArgs(line)
	return fmt.Sprintf("smerr.EnrichApped(ctx, &resp.Diagnostics, %s)", strings.Join(args, ", ")), nil
}

// Pattern-to-handler mapping
// Make sure
var handlerMap = map[string]func(string, map[string]bool) (string, error){
	"&retry.NotFoundError{":               processRetryNotFoundErrorReturn,
	"create.AddError(":                    processCreateAddError,
	"create.AppendDiagError(":             processAppendDiagError,
	"resp.Diagnostics.Append(":            processDiagnosticsAppend,
	"resp.Diagnostics.AddError(":          processRespDiagnosticsAddError,
	"response.Diagnostics.AddError(":      processResponseDiagnosticsAddError,
	"response.Diagnostics.Append(":        processResponseDiagnosticsAppend,
	"return nil, err":                     processNakedErrorReturn,
	"sdkdiag.AppendErrorf(":               processAppendErrorf,
	"sdkdiag.AppendFromErr(":              processAppendFromErr,
	"tfresource.AssertSingleValueResult(": processAssertSingleValueResultReturn,
	"tfresource.NewEmptyResultError(":     processReturnNewEmptyResultError,
	"schema.Schema{":                      processSchemaFormatting,
	"fmt.Errorf(":                         processCaseReturnfmtError,
}

// processSchemaFormatting formats schema.Schema{} structs properly
func processSchemaFormatting(line string, imports map[string]bool) (string, error) {
	// This handler triggers the DST reconstruction which will use our enhanced nodeToString
	// The actual formatting happens in the nodeToString function for CompositeLit
	// We just return the line unchanged to trigger the reconstruction
	return line, nil
}

func chooseHandler(line string, imports map[string]bool) (string, error) {
	fmt.Println("Choosing handler for line:", line)
	var matched bool = false
	for pattern, handler := range handlerMap {
		if strings.Contains(line, pattern) {
			transformedLine, err := handler(line, imports)
			if err != nil {
				return "", err
			}
			line = transformedLine
			matched = true
			break
		}
	}
	if !matched {
		return "", fmt.Errorf("no handler found for expression: %s", line)
	}

	sb := strings.Builder{}
	sb.WriteString(line)
	return sb.String(), nil
}
