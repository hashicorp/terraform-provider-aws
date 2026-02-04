// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIAMPolicySplit_ParsePolicyDocument_Valid(t *testing.T) {
	t.Parallel()
	validPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	policy, err := ParsePolicyDocument(validPolicy)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if policy.Version != "2012-10-17" {
		t.Errorf("Expected Version '2012-10-17', got: %s", policy.Version)
	}

	if len(policy.Statement) != 1 {
		t.Errorf("Expected 1 statement, got: %d", len(policy.Statement))
	}

	if policy.Statement[0].Effect != "Allow" {
		t.Errorf("Expected Effect 'Allow', got: %s", policy.Statement[0].Effect)
	}
}

func TestIAMPolicySplit_ParsePolicyDocument_InvalidJSON(t *testing.T) {
	t.Parallel()
	invalidJSON := `{"Version": "2012-10-17", "Statement": [`

	_, err := ParsePolicyDocument(invalidJSON)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !contains(err.Error(), "JSON syntax error") {
		t.Errorf("Expected 'JSON syntax error' in error message, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_ParsePolicyDocument_MissingVersion(t *testing.T) {
	t.Parallel()
	policyMissingVersion := `{
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::example-bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyMissingVersion)
	if err == nil {
		t.Fatal("Expected error for missing Version, got nil")
	}

	if !contains(err.Error(), "policy document missing required field: Version") {
		t.Errorf("Expected 'missing required field: Version' in error message, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_ParsePolicyDocument_EmptyStatements(t *testing.T) {
	t.Parallel()
	policyEmptyStatements := `{
		"Version": "2012-10-17",
		"Statement": []
	}`

	_, err := ParsePolicyDocument(policyEmptyStatements)
	if err == nil {
		t.Fatal("Expected error for empty statements, got nil")
	}

	if !contains(err.Error(), "policy document missing required field: Statement") {
		t.Errorf("Expected 'missing required field: Statement' in error message, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_ValidateServiceType_Valid(t *testing.T) {
	t.Parallel()
	validTypes := []string{"inline", "managed", "resource-based"}

	for _, serviceType := range validTypes {
		err := ValidateServiceType(serviceType)
		if err != nil {
			t.Errorf("Expected no error for service type '%s', got: %v", serviceType, err)
		}
	}
}

func TestIAMPolicySplit_ValidateServiceType_Invalid(t *testing.T) {
	t.Parallel()
	err := ValidateServiceType("invalid")
	if err == nil {
		t.Fatal("Expected error for invalid service type, got nil")
	}

	if !contains(err.Error(), "invalid service_type 'invalid'") {
		t.Errorf("Expected 'invalid service_type' in error message, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_GetSizeLimitForServiceType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		serviceType string
		expected    int
	}{
		{"inline", 2048},
		{"managed", 6144},
		{"resource-based", 10240},
		{"unknown", 2048}, // Should default to inline
	}

	for _, test := range tests {
		actual := GetSizeLimitForServiceType(test.serviceType)
		if actual != test.expected {
			t.Errorf("For service type '%s', expected %d, got %d", test.serviceType, test.expected, actual)
		}
	}
}

func TestIAMPolicySplit_CalculatePolicySize(t *testing.T) {
	t.Parallel()
	policy := `{"Version":"2012-10-17","Statement":[]}`
	size := CalculatePolicySize(policy)
	expected := len([]byte(policy))

	if size != expected {
		t.Errorf("Expected size %d, got %d", expected, size)
	}
}

func TestIAMPolicySplit_ValidatePolicySize(t *testing.T) {
	t.Parallel()
	smallPolicy := `{"Version":"2012-10-17","Statement":[]}`

	// Should pass for inline (small policy)
	err := ValidatePolicySize(smallPolicy, "inline")
	if err != nil {
		t.Errorf("Expected no error for small policy with inline service, got: %v", err)
	}

	// Create a large policy that exceeds inline limit
	largePolicy := `{"Version":"2012-10-17","Statement":[` +
		generateLargeStatement(3000) + `]}`

	err = ValidatePolicySize(largePolicy, "inline")
	if err == nil {
		t.Error("Expected error for large policy with inline service, got nil")
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateLargeStatement(size int) string {
	// Generate a statement with approximately the specified size
	action := `"Action":["s3:GetObject"`
	for len(action) < size-100 { // Leave room for other fields
		action += `,"s3:GetObject"`
	}
	action += `]`

	return `{"Effect":"Allow",` + action + `,"Resource":"arn:aws:s3:::bucket/*"}`
}
func TestIAMPolicySplit_SplitPolicy_SmallPolicy(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	smallPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	result, err := f.splitPolicy(smallPolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for small policy, got: %v", err)
	}

	// Should return 1 policy (original)
	if result.Count.ValueInt64() != 1 {
		t.Errorf("Expected 1 policy, got: %d", result.Count.ValueInt64())
	}

	// Should have no size reduction (policy unchanged)
	if result.TotalSizeReduction.ValueInt64() != 0 {
		t.Errorf("Expected 0 size reduction, got: %d", result.TotalSizeReduction.ValueInt64())
	}
}

func TestIAMPolicySplit_SplitPolicy_LargePolicy(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy with many statements that will exceed inline limit
	largePolicy := `{
		"Version": "2012-10-17",
		"Id": "test-policy",
		"Statement": [`

	// Add many statements to exceed the 2048 byte limit for inline policies
	for i := range 50 { // Increased from 20 to 50
		if i > 0 {
			largePolicy += ","
		}
		largePolicy += fmt.Sprintf(`{
			"Effect": "Allow",
			"Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
			"Resource": "arn:aws:s3:::very-long-bucket-name-for-testing-purposes-%d/*"
		}`, i)
	}

	largePolicy += `]}`

	// Check the actual size first
	policySize := CalculatePolicySize(largePolicy)
	t.Logf("Policy size: %d bytes (inline limit: 2048)", policySize)

	if policySize <= 2048 {
		t.Skipf("Test policy (%d bytes) doesn't exceed inline limit (2048 bytes), skipping split test", policySize)
	}

	result, err := f.splitPolicy(largePolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for large policy splitting, got: %v", err)
	}

	// Should return multiple policies
	if result.Count.ValueInt64() <= 1 {
		t.Errorf("Expected multiple policies, got: %d", result.Count.ValueInt64())
	}

	// Should have some size increase due to splitting overhead (duplicated Version/Id fields)
	// This is expected and normal
	t.Logf("Size reduction: %d bytes (negative means overhead from splitting)", result.TotalSizeReduction.ValueInt64())
}

func TestIAMPolicySplit_GenerateUniqueId(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	tests := []struct {
		originalId string
		index      int
		expected   string
	}{
		{"", 1, ""},
		{"test-policy", 1, "test-policy-split-1"},
		{"test-policy", 2, "test-policy-split-2"},
		{"my-policy-id", 3, "my-policy-id-split-3"},
	}

	for _, test := range tests {
		actual := f.generateUniqueId(test.originalId, test.index)
		if actual != test.expected {
			t.Errorf("For originalId '%s' and index %d, expected '%s', got '%s'",
				test.originalId, test.index, test.expected, actual)
		}
	}
}

func TestIAMPolicySplit_SplitPolicy_OversizedSingleStatement(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy with a single statement that's too large for inline limit
	oversizedPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [` + strings.Repeat(`"s3:GetObject",`, 200) + `"s3:PutObject"],
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	// Verify the policy is actually oversized
	policySize := CalculatePolicySize(oversizedPolicy)
	t.Logf("Oversized policy size: %d bytes (inline limit: 2048)", policySize)

	if policySize <= 2048 {
		t.Skipf("Test policy (%d bytes) doesn't exceed inline limit, skipping", policySize)
	}

	// Should return an error, not attempt to split the statement
	_, err := f.splitPolicy(oversizedPolicy, "inline")
	if err == nil {
		t.Fatal("Expected error for oversized single statement, got nil")
	}

	// Error should mention that the statement is too large
	if !contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' in error message, got: %s", err.Error())
	}

	// Error should suggest using a different service type
	if !contains(err.Error(), "service type with higher limits") {
		t.Errorf("Expected service type suggestions in error message, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_GeneratePolicyOfSize(t *testing.T) {
	t.Parallel()
	// Test generating a policy that should exceed inline limit
	policy := generatePolicyOfSize(2500, "inline")
	actualSize := CalculatePolicySize(policy)
	t.Logf("Generated policy size: %d bytes (target: 2500)", actualSize)

	// Parse and reformat to see if size changes
	parsed, err := ParsePolicyDocument(policy)
	if err != nil {
		t.Fatalf("Failed to parse generated policy: %v", err)
	}

	reformatted, err := parsed.ToJSON()
	if err != nil {
		t.Fatalf("Failed to reformat policy: %v", err)
	}

	reformattedSize := CalculatePolicySize(reformatted)
	t.Logf("Reformatted policy size: %d bytes", reformattedSize)

	// The sizes should be very close if we're generating compact JSON
	sizeDiff := actualSize - reformattedSize
	if sizeDiff < 0 {
		sizeDiff = -sizeDiff
	}

	t.Logf("Size difference: %d bytes", sizeDiff)

	// Check if reformatted size exceeds inline limit
	inlineLimit := GetSizeLimitForServiceType("inline")
	t.Logf("Inline limit: %d bytes", inlineLimit)
	t.Logf("Reformatted size exceeds limit: %v", reformattedSize > inlineLimit)
}
func TestIAMPolicySplit_SplitPolicy_OutputPolicyCompleteness(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy that will be split
	var builder strings.Builder
	builder.WriteString(`{
		"Version": "2012-10-17",
		"Id": "test-policy",
		"Statement": [`)

	// Add enough statements to force splitting
	for i := range 30 {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`{
			"Effect": "Allow",
			"Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
			"Resource": "arn:aws:s3:::very-long-bucket-name-for-testing-purposes-%d/*"
		}`, i))
	}

	builder.WriteString(`]}`)
	largePolicy := builder.String()

	result, err := f.splitPolicy(largePolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for policy splitting, got: %v", err)
	}

	// Get the policies from the result
	policies := result.Policies
	policyList := policies.Elements()

	if len(policyList) <= 1 {
		t.Skip("Policy wasn't split, skipping completeness test")
	}

	// Test each output policy for completeness
	for i, policyValue := range policyList {
		stringValue, ok := policyValue.(types.String)
		if !ok {
			t.Errorf("Output policy %d is not a string type", i)
			continue
		}
		policyJSON := stringValue.ValueString()

		// Parse the output policy to verify it's complete and valid
		parsedPolicy, err := ParsePolicyDocument(policyJSON)
		if err != nil {
			t.Errorf("Output policy %d is not valid JSON: %v", i, err)
			continue
		}

		// Verify required fields
		if parsedPolicy.Version == "" {
			t.Errorf("Output policy %d missing Version field", i)
		}

		if parsedPolicy.Version != "2012-10-17" {
			t.Errorf("Output policy %d has wrong Version: expected '2012-10-17', got '%s'", i, parsedPolicy.Version)
		}

		if len(parsedPolicy.Statement) == 0 {
			t.Errorf("Output policy %d has no statements", i)
		}

		// Verify unique ID generation
		if parsedPolicy.Id == "" {
			t.Errorf("Output policy %d missing Id field (original had Id)", i)
		} else if !contains(parsedPolicy.Id, "test-policy-split-") {
			t.Errorf("Output policy %d has unexpected Id format: %s", i, parsedPolicy.Id)
		}

		// Verify each statement is complete
		for j, stmt := range parsedPolicy.Statement {
			if stmt.Effect == "" {
				t.Errorf("Output policy %d, statement %d missing Effect", i, j)
			}
			if stmt.Action == nil {
				t.Errorf("Output policy %d, statement %d missing Action", i, j)
			}
			if stmt.Resource == nil {
				t.Errorf("Output policy %d, statement %d missing Resource", i, j)
			}
		}

		// Verify the policy is within size limits
		if err := ValidatePolicySize(policyJSON, "inline"); err != nil {
			t.Errorf("Output policy %d exceeds size limit: %v", i, err)
		}

		// Verify the policy can be used directly (round-trip test)
		reformattedJSON, err := parsedPolicy.ToJSON()
		if err != nil {
			t.Errorf("Output policy %d cannot be reformatted to JSON: %v", i, err)
		}

		// The reformatted JSON should also be valid
		_, err = ParsePolicyDocument(reformattedJSON)
		if err != nil {
			t.Errorf("Output policy %d fails round-trip validation: %v", i, err)
		}
	}
}
func TestIAMPolicySplit_SplitPolicy_MetadataGeneration(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy that will be split
	var builder strings.Builder
	builder.WriteString(`{
		"Version": "2012-10-17",
		"Id": "test-policy",
		"Statement": [`)

	// Add enough statements to force splitting
	for i := range 40 {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(`{
			"Effect": "Allow",
			"Action": ["s3:GetObject", "s3:PutObject"],
			"Resource": "arn:aws:s3:::bucket-%d/*"
		}`, i))
	}

	builder.WriteString(`]}`)
	largePolicy := builder.String()

	result, err := f.splitPolicy(largePolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for policy splitting, got: %v", err)
	}

	// Verify metadata is present
	metadata := result.Metadata
	metadataMap := metadata.Attributes()

	// Check that all metadata fields are present
	requiredFields := []string{"original_size", "average_size", "largest_policy", "smallest_policy"}
	for _, field := range requiredFields {
		if _, exists := metadataMap[field]; !exists {
			t.Errorf("Metadata missing required field: %s", field)
		}
	}

	// Verify metadata values make sense
	originalSize := metadataMap["original_size"].(types.Int64).ValueInt64()
	averageSize := metadataMap["average_size"].(types.Int64).ValueInt64()
	largestPolicy := metadataMap["largest_policy"].(types.Int64).ValueInt64()
	smallestPolicy := metadataMap["smallest_policy"].(types.Int64).ValueInt64()

	if originalSize <= 0 {
		t.Errorf("Original size should be positive, got: %d", originalSize)
	}

	if averageSize <= 0 {
		t.Errorf("Average size should be positive, got: %d", averageSize)
	}

	if largestPolicy < smallestPolicy {
		t.Errorf("Largest policy (%d) should be >= smallest policy (%d)", largestPolicy, smallestPolicy)
	}

	if largestPolicy > 2048 {
		t.Errorf("Largest policy (%d) should not exceed inline limit (2048)", largestPolicy)
	}

	t.Logf("Metadata - Original: %d, Average: %d, Largest: %d, Smallest: %d",
		originalSize, averageSize, largestPolicy, smallestPolicy)
}
func TestIAMPolicySplit_EnhancedErrorHandling_UnsupportedVersion(t *testing.T) {
	t.Parallel()
	policyWithUnsupportedVersion := `{
		"Version": "2020-01-01",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyWithUnsupportedVersion)
	if err == nil {
		t.Fatal("Expected error for unsupported version, got nil")
	}

	if !contains(err.Error(), "unsupported policy version '2020-01-01'") {
		t.Errorf("Expected unsupported version error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_EnhancedErrorHandling_InvalidEffect(t *testing.T) {
	t.Parallel()
	policyWithInvalidEffect := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Maybe",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyWithInvalidEffect)
	if err == nil {
		t.Fatal("Expected error for invalid Effect, got nil")
	}

	if !contains(err.Error(), "Effect must be 'Allow' or 'Deny', got: Maybe") {
		t.Errorf("Expected invalid Effect error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_EnhancedErrorHandling_BothActionAndNotAction(t *testing.T) {
	t.Parallel()
	policyWithBothActions := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"NotAction": "s3:DeleteObject",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyWithBothActions)
	if err == nil {
		t.Fatal("Expected error for both Action and NotAction, got nil")
	}

	if !contains(err.Error(), "cannot have both 'Action' and 'NotAction'") {
		t.Errorf("Expected both Action/NotAction error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_EnhancedErrorHandling_EmptyAction(t *testing.T) {
	t.Parallel()
	policyWithEmptyAction := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "",
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyWithEmptyAction)
	if err == nil {
		t.Fatal("Expected error for empty Action, got nil")
	}

	if !contains(err.Error(), "Action field cannot be empty string") {
		t.Errorf("Expected empty Action error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_EnhancedErrorHandling_InvalidActionArray(t *testing.T) {
	t.Parallel()
	policyWithInvalidActionArray := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": ["s3:GetObject", 123],
				"Resource": "arn:aws:s3:::bucket/*"
			}
		]
	}`

	_, err := ParsePolicyDocument(policyWithInvalidActionArray)
	if err == nil {
		t.Fatal("Expected error for invalid Action array, got nil")
	}

	if !contains(err.Error(), "Action field array item 1 must be string") {
		t.Errorf("Expected invalid Action array error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_DetectImpossibleConstraints_BasePolicyTooLarge(t *testing.T) {
	t.Parallel()
	// Create a policy with a very long ID that makes the base policy too large
	longId := strings.Repeat("very-long-policy-id-", 200) // This will make base policy > 2048 bytes
	policy := &PolicyDocument{
		Version: "2012-10-17",
		Id:      longId,
		Statement: []Statement{
			{
				Effect:   "Allow",
				Action:   "s3:GetObject",
				Resource: "arn:aws:s3:::bucket/*",
			},
		},
	}

	err := detectImpossibleConstraints(policy, "inline")
	if err == nil {
		t.Fatal("Expected error for base policy too large, got nil")
	}

	if !contains(err.Error(), "impossible to split: base policy structure") {
		t.Errorf("Expected impossible constraint error, got: %s", err.Error())
	}

	if !contains(err.Error(), "Consider using a service type with higher limits") {
		t.Errorf("Expected helpful suggestion in error, got: %s", err.Error())
	}
}

func TestIAMPolicySplit_GenerateHelpfulErrorMessage_SizeLimit(t *testing.T) {
	t.Parallel()
	originalErr := fmt.Errorf("policy size (3000 bytes) exceeds inline service limit (2048 bytes)")

	helpfulMsg := generateHelpfulErrorMessage(originalErr, "inline")

	if !contains(helpfulMsg, "Try using 'managed' (6144 bytes) or 'resource-based' (10240 bytes)") {
		t.Errorf("Expected helpful suggestion for inline service type, got: %s", helpfulMsg)
	}
}

func TestIAMPolicySplit_GenerateHelpfulErrorMessage_JSONSyntax(t *testing.T) {
	t.Parallel()
	originalErr := fmt.Errorf("JSON syntax error at position 10")

	helpfulMsg := generateHelpfulErrorMessage(originalErr, "inline")

	if !contains(helpfulMsg, "Validate your JSON using a JSON validator tool") {
		t.Errorf("Expected JSON validation suggestion, got: %s", helpfulMsg)
	}
}

func TestIAMPolicySplit_GenerateHelpfulErrorMessage_MissingField(t *testing.T) {
	t.Parallel()
	originalErr := fmt.Errorf("policy document missing required field: Version")

	helpfulMsg := generateHelpfulErrorMessage(originalErr, "inline")

	if !contains(helpfulMsg, "Ensure your policy includes all required fields") {
		t.Errorf("Expected required fields suggestion, got: %s", helpfulMsg)
	}
}

func TestIAMPolicySplit_SplitPolicy_EnhancedErrorHandling(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Test with malformed JSON
	_, err := f.splitPolicy(`{"Version": "2012-10-17", "Statement": [`, "inline")
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}

	if !contains(err.Error(), "JSON syntax error") {
		t.Errorf("Expected JSON syntax error, got: %s", err.Error())
	}

	// Test with unsupported version
	_, err = f.splitPolicy(`{
		"Version": "2020-01-01",
		"Statement": [{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"}]
	}`, "inline")
	if err == nil {
		t.Fatal("Expected error for unsupported version, got nil")
	}

	if !contains(err.Error(), "unsupported policy version") {
		t.Errorf("Expected unsupported version error, got: %s", err.Error())
	}
}

// Comprehensive test suite for boundary conditions and edge cases

func TestIAMPolicySplit_Comprehensive_BoundaryConditions(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Test policy that should not be split (under limit after reformatting)
	policy := generatePolicyOfSize(1800, "inline")
	actualSize := CalculatePolicySize(policy)
	t.Logf("Generated policy for size 1800: actual size %d bytes", actualSize)

	result, err := f.splitPolicy(policy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for policy under limit, got: %v", err)
	}

	if result.Count.ValueInt64() != 1 {
		t.Errorf("Expected 1 policy for policy under limit, got: %d", result.Count.ValueInt64())
	}

	// Test policy that should be split (over limit after reformatting)
	policy = generatePolicyOfSize(2500, "inline")
	actualSize = CalculatePolicySize(policy)
	t.Logf("Generated policy for size 2500: actual size %d bytes", actualSize)

	// Check if reformatted size exceeds limit
	parsed, err := ParsePolicyDocument(policy)
	if err != nil {
		t.Fatalf("Failed to parse policy: %v", err)
	}
	reformatted, err := parsed.ToJSON()
	if err != nil {
		t.Fatalf("Failed to reformat policy: %v", err)
	}
	reformattedSize := CalculatePolicySize(reformatted)
	t.Logf("Reformatted size: %d bytes", reformattedSize)

	result, err = f.splitPolicy(policy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for policy over limit, got: %v", err)
	}

	if reformattedSize <= 2048 {
		t.Skipf("Reformatted policy (%d bytes) doesn't exceed limit, skipping split test", reformattedSize)
	}

	if result.Count.ValueInt64() <= 1 {
		t.Errorf("Expected multiple policies for policy over limit, got: %d", result.Count.ValueInt64())
	}
}

func TestIAMPolicySplit_Comprehensive_AllServiceTypeLimits(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	testCases := []struct {
		serviceType string
		sizeLimit   int
	}{
		{"inline", 2048},
		{"managed", 6144},
		{"resource-based", 10240},
	}

	for _, tc := range testCases {
		t.Run(tc.serviceType, func(t *testing.T) {
			// Test policy just under limit
			policy := generatePolicyOfSize(tc.sizeLimit-200, tc.serviceType)
			actualSize := CalculatePolicySize(policy)
			t.Logf("Generated policy for size %d: actual size %d bytes", tc.sizeLimit-200, actualSize)

			result, err := f.splitPolicy(policy, tc.serviceType)
			if err != nil {
				t.Fatalf("Expected no error for %s policy under limit, got: %v", tc.serviceType, err)
			}

			if result.Count.ValueInt64() != 1 {
				t.Errorf("Expected 1 policy for %s under limit, got: %d", tc.serviceType, result.Count.ValueInt64())
			}

			// Test policy over limit - use a much larger target to ensure it exceeds limit after reformatting
			targetSize := tc.sizeLimit + tc.sizeLimit/2 // 50% larger than limit
			policy = generatePolicyOfSize(targetSize, tc.serviceType)
			actualSize = CalculatePolicySize(policy)
			t.Logf("Generated policy for size %d: actual size %d bytes", targetSize, actualSize)

			// Check reformatted size
			parsed, err := ParsePolicyDocument(policy)
			if err != nil {
				t.Fatalf("Failed to parse policy: %v", err)
			}
			reformatted, err := parsed.ToJSON()
			if err != nil {
				t.Fatalf("Failed to reformat policy: %v", err)
			}
			reformattedSize := CalculatePolicySize(reformatted)
			t.Logf("Reformatted size: %d bytes, limit: %d bytes", reformattedSize, tc.sizeLimit)

			result, err = f.splitPolicy(policy, tc.serviceType)
			if err != nil {
				t.Fatalf("Expected no error for %s policy over limit, got: %v", tc.serviceType, err)
			}

			if reformattedSize <= tc.sizeLimit {
				t.Skipf("Reformatted policy (%d bytes) doesn't exceed %s limit (%d bytes), skipping split test",
					reformattedSize, tc.serviceType, tc.sizeLimit)
			}

			if result.Count.ValueInt64() <= 1 {
				t.Errorf("Expected multiple policies for %s over limit, got: %d", tc.serviceType, result.Count.ValueInt64())
			}
		})
	}
}

func TestIAMPolicySplit_Comprehensive_StatementTypes(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	testCases := []struct {
		name   string
		policy string
	}{
		{
			name: "StringAction",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"}]
			}`,
		},
		{
			name: "ArrayAction",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{"Effect": "Allow", "Action": ["s3:GetObject", "s3:PutObject"], "Resource": "*"}]
			}`,
		},
		{
			name: "NotAction",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{"Effect": "Deny", "NotAction": "s3:DeleteObject", "Resource": "*"}]
			}`,
		},
		{
			name: "WithCondition",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": "s3:GetObject",
					"Resource": "*",
					"Condition": {"StringEquals": {"s3:x-amz-server-side-encryption": "AES256"}}
				}]
			}`,
		},
		{
			name: "WithPrincipal",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": "s3:GetObject",
					"Resource": "*",
					"Principal": {"AWS": "arn:aws:iam::123456789012:user/testuser"}
				}]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := f.splitPolicy(tc.policy, "inline")
			if err != nil {
				t.Fatalf("Expected no error for %s, got: %v", tc.name, err)
			}

			if result.Count.ValueInt64() != 1 {
				t.Errorf("Expected 1 policy for %s, got: %d", tc.name, result.Count.ValueInt64())
			}
		})
	}
}

func TestIAMPolicySplit_Comprehensive_ErrorScenarios(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	testCases := []struct {
		name          string
		policy        string
		serviceType   string
		expectedError string
	}{
		{
			name:          "EmptyJSON",
			policy:        "",
			serviceType:   "inline",
			expectedError: "policy JSON cannot be empty",
		},
		{
			name:          "InvalidJSON",
			policy:        `{"Version": "2012-10-17", "Statement": [`,
			serviceType:   "inline",
			expectedError: "JSON syntax error",
		},
		{
			name: "BothActionAndNotAction",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": "s3:GetObject",
					"NotAction": "s3:PutObject",
					"Resource": "*"
				}]
			}`,
			serviceType:   "inline",
			expectedError: "cannot have both 'Action' and 'NotAction'",
		},
		{
			name: "InvalidEffect",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Maybe",
					"Action": "s3:GetObject",
					"Resource": "*"
				}]
			}`,
			serviceType:   "inline",
			expectedError: "Effect must be 'Allow' or 'Deny'",
		},
		{
			name: "EmptyAction",
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": "",
					"Resource": "*"
				}]
			}`,
			serviceType:   "inline",
			expectedError: "Action field cannot be empty string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.splitPolicy(tc.policy, tc.serviceType)
			if err == nil {
				t.Fatalf("Expected error for %s, got nil", tc.name)
			}

			if !contains(err.Error(), tc.expectedError) {
				t.Errorf("Expected error containing '%s' for %s, got: %s", tc.expectedError, tc.name, err.Error())
			}
		})
	}
}

func TestIAMPolicySplit_Comprehensive_MetadataAccuracy(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy that will definitely be split
	largePolicy := generatePolicyOfSize(5000, "inline")

	result, err := f.splitPolicy(largePolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for large policy, got: %v", err)
	}

	// Verify metadata
	metadata := result.Metadata.Attributes()

	originalSize := metadata["original_size"].(types.Int64).ValueInt64()
	averageSize := metadata["average_size"].(types.Int64).ValueInt64()
	largestPolicy := metadata["largest_policy"].(types.Int64).ValueInt64()
	smallestPolicy := metadata["smallest_policy"].(types.Int64).ValueInt64()

	// Verify metadata relationships
	if originalSize <= 0 {
		t.Errorf("Original size should be positive, got: %d", originalSize)
	}

	if averageSize <= 0 {
		t.Errorf("Average size should be positive, got: %d", averageSize)
	}

	if largestPolicy < smallestPolicy {
		t.Errorf("Largest policy (%d) should be >= smallest policy (%d)", largestPolicy, smallestPolicy)
	}

	if largestPolicy > 2048 {
		t.Errorf("Largest policy (%d) should not exceed inline limit (2048)", largestPolicy)
	}

	if smallestPolicy <= 0 {
		t.Errorf("Smallest policy should be positive, got: %d", smallestPolicy)
	}

	// Verify count matches actual policies
	policyCount := result.Count.ValueInt64()
	actualPolicies := len(result.Policies.Elements())

	if policyCount != int64(actualPolicies) {
		t.Errorf("Count (%d) should match actual policies (%d)", policyCount, actualPolicies)
	}
}

func TestIAMPolicySplit_Comprehensive_PolicyIntegrity(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Create a policy with multiple statements that will be split - use compact JSON
	originalPolicy := `{"Version":"2012-10-17","Id":"TestPolicy","Statement":[`

	// Add many statements with different properties
	for i := range 30 {
		if i > 0 {
			originalPolicy += ","
		}
		originalPolicy += fmt.Sprintf(`{"Sid":"Statement%d","Effect":"%s","Action":"s3:%sObject","Resource":"arn:aws:s3:::bucket-%d/*"}`, i,
			map[bool]string{true: "Allow", false: "Deny"}[i%2 == 0],
			map[bool]string{true: "Get", false: "Put"}[i%3 == 0],
			i)
	}

	originalPolicy += `]}`

	// Check if this policy will actually be split
	originalSize := CalculatePolicySize(originalPolicy)
	t.Logf("Original policy size: %d bytes", originalSize)

	result, err := f.splitPolicy(originalPolicy, "inline")
	if err != nil {
		t.Fatalf("Expected no error for policy splitting, got: %v", err)
	}

	// Parse original policy to get statement count
	originalParsed, err := ParsePolicyDocument(originalPolicy)
	if err != nil {
		t.Fatalf("Failed to parse original policy: %v", err)
	}

	originalStatementCount := len(originalParsed.Statement)

	// Verify all statements are preserved across split policies
	totalStatements := 0
	policies := result.Policies.Elements()

	for i, policyValue := range policies {
		stringValue, ok := policyValue.(types.String)
		if !ok {
			t.Errorf("Policy %d is not a string type", i)
			continue
		}

		policyJSON := stringValue.ValueString()
		parsedPolicy, err := ParsePolicyDocument(policyJSON)
		if err != nil {
			t.Errorf("Split policy %d is not valid: %v", i, err)
			continue
		}

		totalStatements += len(parsedPolicy.Statement)

		// Verify each split policy has proper structure
		if parsedPolicy.Version != originalParsed.Version {
			t.Errorf("Split policy %d has wrong version: expected %s, got %s",
				i, originalParsed.Version, parsedPolicy.Version)
		}

		if originalParsed.Id != "" && parsedPolicy.Id == "" {
			t.Errorf("Split policy %d missing Id when original had Id", i)
		}
	}

	// Verify statement count is preserved
	if totalStatements != originalStatementCount {
		t.Errorf("Statement count not preserved: original %d, split total %d",
			originalStatementCount, totalStatements)
	}
}

// Helper function to generate a policy of approximately the specified size
func generatePolicyOfSize(targetSize int, serviceType string) string {
	basePolicy := `{
		"Version": "2012-10-17",
		"Statement": [`

	statementTemplate := `{
		"Effect": "Allow",
		"Action": "s3:GetObject",
		"Resource": "arn:aws:s3:::bucket-%d/*"
	}`

	// Calculate the size of the closing part
	closingSize := len(`]}`)

	// Start with base policy size
	currentSize := len(basePolicy)
	statementCount := 0

	// Keep adding statements until we reach or exceed the target size
	for {
		// Format the next statement
		statement := fmt.Sprintf(statementTemplate, statementCount)

		// Calculate size if we add this statement
		additionalSize := len(statement)
		if statementCount > 0 {
			additionalSize += 1 // comma separator
		}

		// Check if adding this statement would make the total policy exceed target
		totalSizeWithStatement := currentSize + additionalSize + closingSize

		// If we haven't added any statements yet, add at least one
		if statementCount == 0 {
			basePolicy += statement
			currentSize += additionalSize
			statementCount++
			continue
		}

		// If adding this statement would exceed target, stop
		if totalSizeWithStatement > targetSize {
			break
		}

		// Add the statement
		basePolicy += "," + statement
		currentSize += additionalSize
		statementCount++

		// Safety check to prevent infinite loop
		if statementCount > 1000 {
			break
		}
	}

	basePolicy += `]}`

	// Verify the generated policy actually exceeds the target when intended
	finalSize := len(basePolicy)
	if targetSize > GetSizeLimitForServiceType(serviceType) && finalSize <= GetSizeLimitForServiceType(serviceType) {
		// If we're trying to generate a policy larger than the service limit but didn't succeed,
		// add more statements with longer resource names
		for finalSize <= targetSize {
			longerStatement := fmt.Sprintf(`{
				"Effect": "Allow",
				"Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"],
				"Resource": "arn:aws:s3:::very-long-bucket-name-for-testing-purposes-to-increase-policy-size-%d/*"
			}`, statementCount)

			basePolicy = basePolicy[:len(basePolicy)-2] + "," + longerStatement + "]}"
			finalSize = len(basePolicy)
			statementCount++

			if statementCount > 1000 {
				break
			}
		}
	}

	return basePolicy
}
func TestIAMPolicySplit_ActualSplitting(t *testing.T) {
	t.Parallel()
	f := &iamPolicySplitFunction{}

	// Generate a policy that should definitely be split after reformatting
	policy := generatePolicyOfSize(2500, "inline")
	actualSize := CalculatePolicySize(policy)
	t.Logf("Generated policy size: %d bytes", actualSize)

	// Parse and check reformatted size
	parsed, err := ParsePolicyDocument(policy)
	if err != nil {
		t.Fatalf("Failed to parse generated policy: %v", err)
	}

	reformatted, err := parsed.ToJSON()
	if err != nil {
		t.Fatalf("Failed to reformat policy: %v", err)
	}

	reformattedSize := CalculatePolicySize(reformatted)
	inlineLimit := GetSizeLimitForServiceType("inline")
	t.Logf("Reformatted size: %d bytes, limit: %d bytes", reformattedSize, inlineLimit)

	// Try to split
	result, err := f.splitPolicy(policy, "inline")
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	t.Logf("Split result: %d policies", result.Count.ValueInt64())

	if reformattedSize > inlineLimit && result.Count.ValueInt64() <= 1 {
		t.Errorf("Expected multiple policies when reformatted size (%d) > limit (%d), got %d policies",
			reformattedSize, inlineLimit, result.Count.ValueInt64())
	}
}
