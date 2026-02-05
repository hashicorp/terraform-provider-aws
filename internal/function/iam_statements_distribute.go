// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// splitResultAttrTypes defines the return type structure for the IAM statements distribute function
var splitResultAttrTypes = map[string]attr.Type{
	"policies": types.ListType{ElemType: types.StringType},
	"metadata": types.ObjectType{AttrTypes: metadataAttrTypes},
}

// metadataAttrTypes defines the metadata structure
var metadataAttrTypes = map[string]attr.Type{
	"original_size":        types.Int64Type,
	"average_size":         types.Int64Type,
	"largest_policy":       types.Int64Type,
	"smallest_policy":      types.Int64Type,
	"total_size_reduction": types.Int64Type,
}

var _ function.Function = iamStatementsDistributeFunction{}

// NewIAMStatementsDistributeFunction creates a new instance of the IAM statements distribute function
func NewIAMStatementsDistributeFunction() function.Function {
	return &iamStatementsDistributeFunction{}
}

type iamStatementsDistributeFunction struct{}

// Metadata sets the function name
func (f iamStatementsDistributeFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "iam_statements_distribute"
}

// Definition defines the function parameters and return type
func (f iamStatementsDistributeFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "iam_statements_distribute Function",
		MarkdownDescription: "Distributes statements from large IAM policy documents across multiple smaller policies that comply with AWS size limits. " +
			"This function helps manage policies that exceed AWS service-specific size constraints by intelligently " +
			"distributing statements across multiple valid policy documents.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "policy_json",
				MarkdownDescription: "IAM policy JSON document to distribute statements from",
			},
			function.StringParameter{
				Name:                "policy_type",
				MarkdownDescription: "AWS policy type for size limits. Valid values: 'customer-managed' (6144 bytes), 'inline-user' (2048 bytes), 'inline-role' (10240 bytes), 'inline-group' (5120 bytes) and 'service-control-policy' (5120 bytes).",
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: splitResultAttrTypes,
		},
	}
}

// Run executes the IAM statements distribution logic
func (f iamStatementsDistributeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var policyJSON string
	var policyType string

	// Get required parameters
	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &policyJSON, &policyType))
	if resp.Error != nil {
		return
	}

	// Validate that policy_type is provided (no default)
	if policyType == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError("policy_type parameter is required"))
		return
	}

	// For now, return a placeholder result to establish the structure
	result, err := f.distributeStatements(policyJSON, policyType)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	// Convert result to Terraform types
	value := map[string]attr.Value{
		"policies": result.Policies,
		"metadata": result.Metadata,
	}

	resultObj, d := types.ObjectValue(splitResultAttrTypes, value)
	if d.HasError() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.FuncErrorFromDiags(ctx, d))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, resultObj))
}

// DistributeResult represents the result of the statements distribution operation
type DistributeResult struct {
	Policies types.List
	Metadata types.Object
}

// SplitMetadata contains additional information about the distribution operation
type SplitMetadata struct {
	OriginalSize       int64
	AverageSize        int64
	LargestPolicy      int64
	SmallestPolicy     int64
	TotalSizeReduction int64
}

// PolicyDocument represents an IAM policy document structure
type PolicyDocument struct {
	Version   string      `json:"Version"`
	Id        string      `json:"Id,omitempty"`
	Statement []Statement `json:"Statement"`
}

// Statement represents an individual IAM policy statement
type Statement struct {
	Sid          string         `json:"Sid,omitempty"`
	Effect       string         `json:"Effect"`
	Action       any            `json:"Action,omitempty"`
	NotAction    any            `json:"NotAction,omitempty"`
	Resource     any            `json:"Resource,omitempty"`
	NotResource  any            `json:"NotResource,omitempty"`
	Principal    any            `json:"Principal,omitempty"`
	NotPrincipal any            `json:"NotPrincipal,omitempty"`
	Condition    map[string]any `json:"Condition,omitempty"`
}

// Validate checks if the policy document has required fields
func (p *PolicyDocument) Validate() error {
	if p.Version == "" {
		return fmt.Errorf("policy document missing required field: Version")
	}

	if len(p.Statement) == 0 {
		return fmt.Errorf("policy document missing required field: Statement (must contain at least one statement)")
	}

	// Validate each statement
	for i, stmt := range p.Statement {
		if err := stmt.Validate(); err != nil {
			return fmt.Errorf("statement %d: %w", i, err)
		}
	}

	return nil
}

// Validate checks if the statement has required fields
func (s *Statement) Validate() error {
	if s.Effect == "" {
		return fmt.Errorf("statement missing required field: Effect")
	}

	if s.Effect != "Allow" && s.Effect != "Deny" {
		return fmt.Errorf("statement Effect must be 'Allow' or 'Deny', got: %s", s.Effect)
	}

	// At least one of Action/NotAction must be present
	if s.Action == nil && s.NotAction == nil {
		return fmt.Errorf("statement must have either Action or NotAction")
	}

	return nil
}

// ToJSON converts the policy document to JSON string
func (p *PolicyDocument) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy to JSON: %w", err)
	}
	return string(data), nil
}

// ParsePolicyDocument parses a JSON string into a PolicyDocument
func ParsePolicyDocument(jsonStr string) (*PolicyDocument, error) {
	// Check for empty input
	if strings.TrimSpace(jsonStr) == "" {
		return nil, fmt.Errorf("policy JSON cannot be empty")
	}

	var policy PolicyDocument

	if err := json.Unmarshal([]byte(jsonStr), &policy); err != nil {
		// Provide more descriptive JSON parsing errors
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("JSON syntax error at position %d: %w", syntaxErr.Offset, err)
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return nil, fmt.Errorf("JSON type error: expected %s but got %s for field %s", typeErr.Type, typeErr.Value, typeErr.Field)
		}
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Basic validation
	if err := policy.Validate(); err != nil {
		return nil, fmt.Errorf("policy validation failed: %w", err)
	}

	// Enhanced structure validation
	if err := validatePolicyStructure(&policy); err != nil {
		return nil, fmt.Errorf("policy structure validation failed: %w", err)
	}

	return &policy, nil
}

// ValidatePolicyType checks if the policy type is valid
func ValidatePolicyType(policyType string) error {
	validTypes := []string{
		"customer-managed",
		"inline-user",
		"inline-role",
		"inline-group",
		"service-control-policy",
	}

	if !slices.Contains(validTypes, policyType) {
		return fmt.Errorf("invalid policy_type '%s': must be one of %s", policyType, strings.Join(validTypes, ", "))
	}

	return nil
}

// GetSizeLimitForPolicyType returns the size limit in bytes for the given policy type
func GetSizeLimitForPolicyType(policyType string) int {
	switch policyType {
	case "customer-managed":
		return 6144
	case "inline-user":
		return 2048
	case "inline-role":
		return 10240
	case "inline-group":
		return 5120
	case "service-control-policy":
		return 5120
	default:
		// Default to inline-user policy limit for safety (most restrictive)
		return 2048
	}
}

// CalculatePolicySize calculates the byte size of a policy JSON string
func CalculatePolicySize(policyJSON string) int {
	return len([]byte(policyJSON))
}

// CalculateBasePolicySize calculates the minimum size of a policy structure
// This includes Version field and empty Statement array with JSON formatting overhead
func CalculateBasePolicySize(version, id string) int {
	basePolicy := PolicyDocument{
		Version:   version,
		Id:        id,
		Statement: []Statement{},
	}

	jsonStr, err := basePolicy.ToJSON()
	if err != nil {
		// Fallback calculation if JSON marshaling fails
		overhead := len(`{"Version":"","Statement":[]}`) + len(version)
		if id != "" {
			overhead += len(`,"Id":""`) + len(id)
		}
		return overhead
	}

	return CalculatePolicySize(jsonStr)
}

// EstimateStatementSize estimates the JSON size of a single statement
func EstimateStatementSize(stmt Statement) (int, error) {
	jsonStr, err := json.Marshal(stmt)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal statement: %w", err)
	}
	return len(jsonStr), nil
}

// ValidatePolicySize checks if a policy exceeds the size limit for the policy type
func ValidatePolicySize(policyJSON, policyType string) error {
	size := CalculatePolicySize(policyJSON)
	limit := GetSizeLimitForPolicyType(policyType)

	if size > limit {
		return fmt.Errorf("policy size (%d bytes) exceeds %s policy limit (%d bytes)", size, policyType, limit)
	}

	return nil
}

// distributeStatements is the main distribution logic (renamed from splitPolicy)
func (f iamStatementsDistributeFunction) distributeStatements(policyJSON, policyType string) (*DistributeResult, error) {
	// Validate policy type
	if err := ValidatePolicyType(policyType); err != nil {
		return nil, err
	}

	// Parse and validate the input policy
	policy, err := ParsePolicyDocument(policyJSON)
	if err != nil {
		return nil, fmt.Errorf("%s", generateHelpfulErrorMessage(err, policyType))
	}

	// Check for impossible constraints before attempting to distribute
	if err := detectImpossibleConstraints(policy, policyType); err != nil {
		return nil, err
	}

	// Convert back to JSON to ensure proper formatting
	outputJSON, err := policy.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to format output policy: %w", err)
	}

	// Check if the original policy already fits within the size limit
	originalSize := CalculatePolicySize(policyJSON)
	sizeLimit := GetSizeLimitForPolicyType(policyType)

	if originalSize <= sizeLimit {
		// Policy is already within limits, return as-is (but use reformatted JSON for consistency)
		policies, _ := types.ListValue(types.StringType, []attr.Value{
			types.StringValue(outputJSON),
		})

		// Generate metadata for single policy
		metadata := f.calculateMetadata(originalSize, []int{}, 0)

		return &DistributeResult{
			Policies: policies,
			Metadata: metadata,
		}, nil
	}

	// Policy exceeds size limit, need to distribute statements
	distributedPolicies, err := f.distributePolicyStatements(policy, policyType)
	if err != nil {
		return nil, err
	}

	// Convert distributed policies to JSON strings and collect metadata
	var policyJSONs []attr.Value
	totalDistributedSize := 0
	var policySizes []int

	for _, distributedPolicy := range distributedPolicies {
		distributedJSON, err := distributedPolicy.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to format distributed policy: %w", err)
		}

		// Validate that each distributed policy is within limits
		if err := ValidatePolicySize(distributedJSON, policyType); err != nil {
			return nil, fmt.Errorf("distributed policy still exceeds size limit: %w", err)
		}

		policySize := CalculatePolicySize(distributedJSON)
		policyJSONs = append(policyJSONs, types.StringValue(distributedJSON))
		totalDistributedSize += policySize
		policySizes = append(policySizes, policySize)
	}

	// Calculate size reduction and metadata
	sizeReduction := originalSize - totalDistributedSize
	metadata := f.calculateMetadata(originalSize, policySizes, sizeReduction)

	policies, _ := types.ListValue(types.StringType, policyJSONs)

	return &DistributeResult{
		Policies: policies,
		Metadata: metadata,
	}, nil
}

// distributePolicyStatements implements an accurate size-based bin-packing algorithm
func (f iamStatementsDistributeFunction) distributePolicyStatements(policy *PolicyDocument, policyType string) ([]*PolicyDocument, error) {
	sizeLimit := GetSizeLimitForPolicyType(policyType)

	// Calculate base policy size (Version + empty Statement array + optional Id)
	basePolicySize := CalculateBasePolicySize(policy.Version, policy.Id)

	// Check if base policy structure exceeds limit
	if basePolicySize >= sizeLimit {
		return nil, fmt.Errorf("base policy structure (%d bytes) exceeds %s policy limit (%d bytes)",
			basePolicySize, policyType, sizeLimit)
	}

	// Check if any individual statement is too large
	for i, stmt := range policy.Statement {
		tempPolicy := &PolicyDocument{
			Version:   policy.Version,
			Id:        policy.Id,
			Statement: []Statement{stmt},
		}

		tempJSON, err := tempPolicy.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal statement %d: %w", i, err)
		}

		if CalculatePolicySize(tempJSON) > sizeLimit {
			return nil, fmt.Errorf("statement %d (%d bytes) is too large for %s policy limit (%d bytes)",
				i, CalculatePolicySize(tempJSON), policyType, sizeLimit)
		}
	}

	// Use greedy bin-packing algorithm with actual size validation
	var distributedPolicies []*PolicyDocument
	remainingStatements := make([]Statement, len(policy.Statement))
	copy(remainingStatements, policy.Statement)

	for len(remainingStatements) > 0 {
		// Create new policy
		currentPolicy := &PolicyDocument{
			Version:   policy.Version,
			Id:        f.generateUniqueId(policy.Id, len(distributedPolicies)+1),
			Statement: []Statement{},
		}

		var usedIndices []int

		// Try to fit as many statements as possible
		for i, stmt := range remainingStatements {
			// Try adding this statement
			testPolicy := &PolicyDocument{
				Version:   currentPolicy.Version,
				Id:        currentPolicy.Id,
				Statement: append(currentPolicy.Statement, stmt),
			}

			testJSON, err := testPolicy.ToJSON()
			if err != nil {
				continue // Skip this statement if it causes marshaling issues
			}

			// Check if it fits
			if CalculatePolicySize(testJSON) <= sizeLimit {
				currentPolicy.Statement = append(currentPolicy.Statement, stmt)
				usedIndices = append(usedIndices, i)
			}
		}

		// Remove used statements from remaining list (in reverse order to maintain indices)
		for i := len(usedIndices) - 1; i >= 0; i-- {
			idx := usedIndices[i]
			remainingStatements = append(remainingStatements[:idx], remainingStatements[idx+1:]...)
		}

		// Ensure we made progress
		if len(usedIndices) == 0 && len(remainingStatements) > 0 {
			return nil, fmt.Errorf("unable to fit remaining statements - algorithm error")
		}

		distributedPolicies = append(distributedPolicies, currentPolicy)
	}

	// Ensure we have at least one policy
	if len(distributedPolicies) == 0 {
		return nil, fmt.Errorf("no statements could be distributed - all statements may be too large")
	}

	return distributedPolicies, nil
}

// generateUniqueId generates a unique ID for distributed policies
func (f iamStatementsDistributeFunction) generateUniqueId(originalId string, index int) string {
	if originalId == "" {
		return ""
	}

	return fmt.Sprintf("%s-split-%d", originalId, index)
}

// calculateMetadata generates metadata about the distribution operation
func (f iamStatementsDistributeFunction) calculateMetadata(originalSize int, policySizes []int, sizeReduction int) types.Object {
	if len(policySizes) == 0 {
		// Return empty metadata for single policy case
		metadataValue := map[string]attr.Value{
			"original_size":        types.Int64Value(int64(originalSize)),
			"average_size":         types.Int64Value(int64(originalSize)),
			"largest_policy":       types.Int64Value(int64(originalSize)),
			"smallest_policy":      types.Int64Value(int64(originalSize)),
			"total_size_reduction": types.Int64Value(int64(sizeReduction)),
		}

		result, _ := types.ObjectValue(metadataAttrTypes, metadataValue)
		return result
	}

	// Calculate statistics
	totalSize := 0
	largest := policySizes[0]
	smallest := policySizes[0]

	for _, size := range policySizes {
		totalSize += size
		if size > largest {
			largest = size
		}
		if size < smallest {
			smallest = size
		}
	}

	averageSize := totalSize / len(policySizes)

	metadataValue := map[string]attr.Value{
		"original_size":        types.Int64Value(int64(originalSize)),
		"average_size":         types.Int64Value(int64(averageSize)),
		"largest_policy":       types.Int64Value(int64(largest)),
		"smallest_policy":      types.Int64Value(int64(smallest)),
		"total_size_reduction": types.Int64Value(int64(sizeReduction)),
	}

	result, _ := types.ObjectValue(metadataAttrTypes, metadataValue)
	return result
}

// validatePolicyStructure performs comprehensive validation of policy structure
func validatePolicyStructure(policy *PolicyDocument) error {
	// Check for supported version
	supportedVersions := []string{"2012-10-17", "2008-10-17"}

	if !slices.Contains(supportedVersions, policy.Version) {
		return fmt.Errorf("unsupported policy version '%s': supported versions are %s",
			policy.Version, strings.Join(supportedVersions, ", "))
	}

	// Validate each statement more thoroughly
	for i, stmt := range policy.Statement {
		if err := validateStatementStructure(stmt); err != nil {
			return fmt.Errorf("statement %d: %w", i, err)
		}
	}

	return nil
}

// validateStatementStructure performs detailed validation of individual statements
func validateStatementStructure(stmt Statement) error {
	// Check Effect field
	if stmt.Effect != "Allow" && stmt.Effect != "Deny" {
		return fmt.Errorf("invalid Effect '%s': must be 'Allow' or 'Deny'", stmt.Effect)
	}

	// Check that we have either Action or NotAction (but not both)
	hasAction := stmt.Action != nil
	hasNotAction := stmt.NotAction != nil

	if !hasAction && !hasNotAction {
		return fmt.Errorf("statement must have either 'Action' or 'NotAction' field")
	}

	if hasAction && hasNotAction {
		return fmt.Errorf("statement cannot have both 'Action' and 'NotAction' fields")
	}

	// Validate Action/NotAction format
	if hasAction {
		if err := validateActionField(stmt.Action, "Action"); err != nil {
			return err
		}
	}

	if hasNotAction {
		if err := validateActionField(stmt.NotAction, "NotAction"); err != nil {
			return err
		}
	}

	// Check Resource/NotResource (similar logic)
	hasResource := stmt.Resource != nil
	hasNotResource := stmt.NotResource != nil

	if hasResource && hasNotResource {
		return fmt.Errorf("statement cannot have both 'Resource' and 'NotResource' fields")
	}

	// Check Principal/NotPrincipal (similar logic)
	hasPrincipal := stmt.Principal != nil
	hasNotPrincipal := stmt.NotPrincipal != nil

	if hasPrincipal && hasNotPrincipal {
		return fmt.Errorf("statement cannot have both 'Principal' and 'NotPrincipal' fields")
	}

	return nil
}

// validateActionField validates that Action/NotAction fields are properly formatted
func validateActionField(action any, fieldName string) error {
	if action == nil {
		return fmt.Errorf("%s field cannot be null", fieldName)
	}

	switch v := action.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("%s field cannot be empty string", fieldName)
		}
	case []any:
		if len(v) == 0 {
			return fmt.Errorf("%s field cannot be empty array", fieldName)
		}
		for i, item := range v {
			if str, ok := item.(string); ok {
				if strings.TrimSpace(str) == "" {
					return fmt.Errorf("%s field array item %d cannot be empty string", fieldName, i)
				}
			} else {
				return fmt.Errorf("%s field array item %d must be string, got %T", fieldName, i, item)
			}
		}
	default:
		return fmt.Errorf("%s field must be string or array of strings, got %T", fieldName, v)
	}

	return nil
}

// detectImpossibleConstraints checks for scenarios where distribution cannot succeed
func detectImpossibleConstraints(policy *PolicyDocument, policyType string) error {
	sizeLimit := GetSizeLimitForPolicyType(policyType)
	basePolicySize := CalculateBasePolicySize(policy.Version, policy.Id)

	// Check if base policy structure is too large
	if basePolicySize >= sizeLimit {
		return fmt.Errorf("impossible to distribute: base policy structure (%d bytes) exceeds %s policy limit (%d bytes). "+
			"Consider using a policy type with higher limits: 'inline-group' (5120 bytes), 'customer-managed' (6144 bytes), or 'inline-role' (10240 bytes)",
			basePolicySize, policyType, sizeLimit)
	}

	// Check if any individual statement is too large
	availableSpace := sizeLimit - basePolicySize - 10 // Reserve some space for JSON formatting

	for i, stmt := range policy.Statement {
		stmtSize, err := EstimateStatementSize(stmt)
		if err != nil {
			return fmt.Errorf("failed to estimate size of statement %d: %w", i, err)
		}

		if stmtSize > availableSpace {
			return fmt.Errorf("impossible to distribute: statement %d (%d bytes) is too large for %s policy limit. "+
				"Statement size exceeds available space (%d bytes). "+
				"Consider simplifying the statement or using a policy type with higher limits",
				i, stmtSize, policyType, availableSpace)
		}
	}

	return nil
}

// generateHelpfulErrorMessage creates user-friendly error messages with suggestions
func generateHelpfulErrorMessage(err error, policyType string) string {
	errMsg := err.Error()

	// Add helpful suggestions based on error type
	if strings.Contains(errMsg, "exceeds") && strings.Contains(errMsg, "policy limit") {
		switch policyType {
		case "inline-user":
			return errMsg + ". Suggestion: Try using 'inline-group' (5120 bytes), 'customer-managed' (6144 bytes), or 'inline-role' (10240 bytes) policy type for larger limits."
		case "inline-group":
			return errMsg + ". Suggestion: Try using 'customer-managed' (6144 bytes) or 'inline-role' (10240 bytes) policy type for larger limits."
		case "service-control-policy":
			return errMsg + ". Suggestion: Try using 'customer-managed' (6144 bytes) or 'inline-role' (10240 bytes) policy type for larger limits."
		case "customer-managed":
			return errMsg + ". Suggestion: Try using 'inline-role' (10240 bytes) policy type for larger limits."
		default:
			return errMsg + ". Suggestion: Consider simplifying your policy statements or breaking them into smaller, more focused statements."
		}
	}

	if strings.Contains(errMsg, "JSON syntax error") {
		return errMsg + ". Suggestion: Validate your JSON using a JSON validator tool."
	}

	if strings.Contains(errMsg, "missing required field") {
		return errMsg + ". Suggestion: Ensure your policy includes all required fields: Version and Statement array."
	}

	return errMsg
}
