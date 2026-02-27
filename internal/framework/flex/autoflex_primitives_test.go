// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten using generic-style roundtrip testing of strings,
// bools, int64, int32, float64, and float32 with various variants: standard, legacy, pointers.

import (
	"bytes"
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// TestPrimitivesRoundtrip is the proof of concept for string roundtrip testing
// This replaces TestExpandString and TestFlattenString with comprehensive roundtrip validation
func TestPrimitivesRoundtrip(t *testing.T) {
	t.Parallel()

	// Test string roundtrips with all variants
	t.Run("String", func(t *testing.T) {
		t.Parallel()
		testStringRoundtrip(t)
	})

	// Test bool roundtrips with all variants
	t.Run("Bool", func(t *testing.T) {
		t.Parallel()
		testBoolRoundtrip(t)
	})

	// Test int64 roundtrips with all variants
	t.Run("Int64", func(t *testing.T) {
		t.Parallel()
		testInt64Roundtrip(t)
	})

	// Test int32 roundtrips with all variants
	t.Run("Int32", func(t *testing.T) {
		t.Parallel()
		testInt32Roundtrip(t)
	})

	// Test float64 roundtrips with all variants
	t.Run("Float64", func(t *testing.T) {
		t.Parallel()
		testFloat64Roundtrip(t)
	})

	// Test float32 roundtrips with all variants
	t.Run("Float32", func(t *testing.T) {
		t.Parallel()
		testFloat32Roundtrip(t)
	})
}

func testStringRoundtrip(t *testing.T) {
	// Define String-specific type info
	stringTypeInfo := PrimitiveTypeInfo[string]{
		TFType:         reflect.TypeFor[types.String](),
		CreateValue:    func(v string) any { return types.StringValue(v) },
		CreateNull:     func() any { return types.StringNull() },
		CreateAWSValue: func(v string) any { return aws.String(v) },
		GetAWSNil:      func() any { return (*string)(nil) },
		GetZeroValue:   func() string { return "" },
	}

	// Test cases covering all scenarios from original TestExpandString and TestFlattenString
	testCases := []struct {
		name        string
		stringValue string
		isNull      bool
		isEmpty     bool
		variants    []string // which variants to test: "standard", "legacy"
		skipExpand  bool     // skip expand direction (flatten-only test)
	}{
		{
			name:        "normal_value",
			stringValue: "test_value",
			variants:    []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:        "empty_string",
			stringValue: "",
			isEmpty:     true,
			variants:    []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:        "special_characters",
			stringValue: "test with spaces & symbols!",
			variants:    []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:        "unicode_content",
			stringValue: "测试内容 🚀",
			variants:    []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		// Random value for property-based testing feel
		{
			name:        "random_value",
			stringValue: sdkacctest.RandomWithPrefix("tf-test"), // nosemgrep:ci.semgrep.acctest.vcr.use-acctest-randomwithprefix
			variants:    []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		// Omitempty tests - flatten-only (expand direction not defined in original tests)
		{
			name:        "omitempty_normal_value",
			stringValue: "test_value",
			variants:    []string{"omitempty"},
			skipExpand:  true, // Only test flatten direction for omitempty
		},
		{
			name:        "omitempty_empty_string",
			stringValue: "",
			isEmpty:     true,
			variants:    []string{"omitempty"},
			skipExpand:  true,
		},
		{
			name:       "omitempty_null_value",
			isNull:     true,
			variants:   []string{"omitempty"},
			skipExpand: true,
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Special handling for omitempty (flatten-only) cases
				if tc.skipExpand {
					// Generate structs for this variant
					var factory func(reflect.Type) (any, any)
					for _, v := range primitiveTestVariants {
						if v.Name == variant {
							factory = v.Factory
							break
						}
					}

					tfStruct, awsStruct := factory(reflect.TypeFor[types.String]())

					// Set up AWS struct based on omitempty behavior
					if tc.isNull {
						setFieldValue(awsStruct, "Field1", (*string)(nil))
					} else if tc.isEmpty {
						// Omitempty: empty becomes nil
						setFieldValue(awsStruct, "Field1", (*string)(nil))
					} else {
						setFieldValue(awsStruct, "Field1", aws.String(tc.stringValue))
					}

					// Set up the expected TF result based on omitempty behavior
					expectedTFResult := reflect.New(reflect.TypeOf(tfStruct).Elem()).Interface()
					if tc.isNull || tc.isEmpty {
						// Omitempty: nil/empty AWS values become null TF values
						setFieldValue(expectedTFResult, "Field1", types.StringNull())
					} else {
						setFieldValue(expectedTFResult, "Field1", types.StringValue(tc.stringValue))
					}
					runFlattenOnlyTest(t, testName, awsStruct, expectedTFResult)
				} else {
					// Use helper for all standard roundtrip cases
					runBasicRoundtripTest(t, testName, variant, stringTypeInfo, tc.stringValue, tc.isNull, false, tc.isEmpty, runChecks{CompareTarget: true})
				}
			})
		}
	}
}

func testBoolRoundtrip(t *testing.T) {
	// Define Bool-specific type info
	boolTypeInfo := PrimitiveTypeInfo[bool]{
		TFType:         reflect.TypeFor[types.Bool](),
		CreateValue:    func(v bool) any { return types.BoolValue(v) },
		CreateNull:     func() any { return types.BoolNull() },
		CreateAWSValue: func(v bool) any { return aws.Bool(v) },
		GetAWSNil:      func() any { return (*bool)(nil) },
		GetZeroValue:   func() bool { return false },
	}

	// Test cases covering all scenarios from original TestExpandBool and TestFlattenBool
	testCases := []struct {
		name       string
		boolValue  bool
		isNull     bool
		variants   []string // which variants to test: "standard", "legacy"
		skipExpand bool     // skip expand direction (flatten-only test)
	}{
		{
			name:      "true_value",
			boolValue: true,
			variants:  []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:      "false_value",
			boolValue: false,
			variants:  []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Check for unsupported skipExpand cases
				if tc.skipExpand {
					t.Fatalf("skipExpand=true for Bool tests requires special handling implementation")
				}

				// Use helper for all standard roundtrip cases
				// Note: false value should be treated as "zero" for legacy mode
				isZero := !tc.boolValue && !tc.isNull
				runBasicRoundtripTest(t, testName, variant, boolTypeInfo, tc.boolValue, tc.isNull, isZero, false, runChecks{CompareTarget: true})
			})
		}
	}
}

func testInt64Roundtrip(t *testing.T) {
	// Define Int64-specific type info
	int64TypeInfo := PrimitiveTypeInfo[int64]{
		TFType:         reflect.TypeFor[types.Int64](),
		CreateValue:    func(v int64) any { return types.Int64Value(v) },
		CreateNull:     func() any { return types.Int64Null() },
		CreateAWSValue: func(v int64) any { return aws.Int64(v) },
		GetAWSNil:      func() any { return (*int64)(nil) },
		GetZeroValue:   func() int64 { return 0 },
	}

	// Test cases covering all scenarios from original TestExpandInt64 and TestFlattenInt64
	testCases := []struct {
		name       string
		int64Value int64
		isNull     bool
		isZero     bool
		variants   []string // which variants to test: "standard", "legacy"
	}{
		{
			name:       "value",
			int64Value: 42,
			variants:   []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:       "zero_value",
			int64Value: 0,
			isZero:     true,
			variants:   []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Use helper for all roundtrip cases
				runBasicRoundtripTest(t, testName, variant, int64TypeInfo, tc.int64Value, tc.isNull, tc.isZero, false, runChecks{CompareTarget: true})
			})
		}
	}
}

func testInt32Roundtrip(t *testing.T) {
	// Define Int32-specific type info
	int32TypeInfo := PrimitiveTypeInfo[int32]{
		TFType:         reflect.TypeFor[types.Int32](),
		CreateValue:    func(v int32) any { return types.Int32Value(v) },
		CreateNull:     func() any { return types.Int32Null() },
		CreateAWSValue: func(v int32) any { return aws.Int32(v) },
		GetAWSNil:      func() any { return (*int32)(nil) },
		GetZeroValue:   func() int32 { return 0 },
	}

	// Test cases covering all scenarios from original TestExpandInt32 and TestFlattenInt32
	testCases := []struct {
		name       string
		int32Value int32
		isNull     bool
		isZero     bool
		variants   []string // which variants to test: "standard", "legacy"
	}{
		{
			name:       "value",
			int32Value: 42,
			variants:   []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:       "zero_value",
			int32Value: 0,
			isZero:     true,
			variants:   []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Use helper for all roundtrip cases
				runBasicRoundtripTest(t, testName, variant, int32TypeInfo, tc.int32Value, tc.isNull, tc.isZero, false, runChecks{CompareTarget: true})
			})
		}
	}
}

func testFloat64Roundtrip(t *testing.T) {
	// Define Float64-specific type info
	float64TypeInfo := PrimitiveTypeInfo[float64]{
		TFType:         reflect.TypeFor[types.Float64](),
		CreateValue:    func(v float64) any { return types.Float64Value(v) },
		CreateNull:     func() any { return types.Float64Null() },
		CreateAWSValue: func(v float64) any { return aws.Float64(v) },
		GetAWSNil:      func() any { return (*float64)(nil) },
		GetZeroValue:   func() float64 { return 0.0 },
	}

	// Test cases covering all scenarios from original TestExpandFloat64 and TestFlattenFloat64
	testCases := []struct {
		name         string
		float64Value float64
		isNull       bool
		isZero       bool
		variants     []string // which variants to test: "standard", "legacy"
		skipExpand   bool     // for future expansion if needed
	}{
		{
			name:         "value",
			float64Value: 42.0,
			variants:     []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:         "zero_value",
			float64Value: 0.0,
			isZero:       true,
			variants:     []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Check for unsupported skipExpand cases
				if tc.skipExpand {
					t.Fatalf("skipExpand=true for Float64 tests requires special handling implementation")
				}

				// Use helper for all roundtrip cases
				runBasicRoundtripTest(t, testName, variant, float64TypeInfo, tc.float64Value, tc.isNull, tc.isZero, false, runChecks{CompareTarget: true})
			})
		}
	}
}

func testFloat32Roundtrip(t *testing.T) {
	// Define Float32-specific type info
	float32TypeInfo := PrimitiveTypeInfo[float32]{
		TFType:         reflect.TypeFor[types.Float32](),
		CreateValue:    func(v float32) any { return types.Float32Value(v) },
		CreateNull:     func() any { return types.Float32Null() },
		CreateAWSValue: func(v float32) any { return aws.Float32(v) },
		GetAWSNil:      func() any { return (*float32)(nil) },
		GetZeroValue:   func() float32 { return 0.0 },
	}

	// Test cases covering all scenarios from original TestExpandFloat32 and TestFlattenFloat32
	testCases := []struct {
		name         string
		float32Value float32
		isNull       bool
		isZero       bool
		variants     []string // which variants to test: "standard", "legacy"
		skipExpand   bool     // for future expansion if needed
	}{
		{
			name:         "value",
			float32Value: 42.0,
			variants:     []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:         "zero_value",
			float32Value: 0.0,
			isZero:       true,
			variants:     []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
		{
			name:     "null_value",
			isNull:   true,
			variants: []string{"standard", "legacy", "tf_to_aws_pointer", "legacy_tf_to_aws_pointer"},
		},
	}

	for _, tc := range testCases {
		for _, variant := range tc.variants {
			testName := tc.name + "_" + variant
			t.Run(testName, func(t *testing.T) {
				// Check for unsupported skipExpand cases
				if tc.skipExpand {
					t.Fatalf("skipExpand=true for Float32 tests requires special handling implementation")
				}

				// Use helper for all roundtrip cases
				runBasicRoundtripTest(t, testName, variant, float32TypeInfo, tc.float32Value, tc.isNull, tc.isZero, false, runChecks{CompareTarget: true})
			})
		}
	}
}

// runBasicRoundtripTest runs a single roundtrip test with standardized struct setup
func runBasicRoundtripTest[T any](t *testing.T, testName string, variant string, typeInfo PrimitiveTypeInfo[T], value T, isNull, isZero, isEmpty bool, checks runChecks) {
	t.Helper()

	// Generate structs for this variant
	var factory func(reflect.Type) (any, any)
	for _, v := range primitiveTestVariants {
		if v.Name == variant {
			factory = v.Factory
			break
		}
	}

	tfStruct, awsStruct := factory(typeInfo.TFType)

	// Set up TF struct with test value (always use value types for TF)
	v := reflect.ValueOf(tfStruct).Elem()
	field := v.FieldByName("Field1")
	if !field.IsValid() || !field.CanSet() {
		t.Fatalf("Field1 is not valid or cannot be set")
	}

	if isNull {
		field.Set(reflect.ValueOf(typeInfo.CreateNull()))
	} else {
		field.Set(reflect.ValueOf(typeInfo.CreateValue(value)))
	}

	// Set up expected AWS struct based on variant - this is the common pattern
	switch {
	case variant == "legacy" || strings.HasPrefix(variant, "legacy_"):
		if isNull {
			// Legacy null behavior: null -> nil for pointers, zero for values
			v := reflect.ValueOf(awsStruct).Elem()
			field := v.FieldByName("Field1")
			awsFieldType := field.Type()
			if awsFieldType.Kind() == reflect.Ptr {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.GetAWSNil()))
				}
			} else {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.GetZeroValue()))
				}
			}
		} else if isZero || isEmpty {
			// Legacy zero/empty behavior: usually -> nil for pointers
			v := reflect.ValueOf(awsStruct).Elem()
			field := v.FieldByName("Field1")
			awsFieldType := field.Type()
			if awsFieldType.Kind() == reflect.Ptr {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.GetAWSNil()))
				}
			} else {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(value))
				}
			}
		} else {
			// Legacy non-zero behavior
			v := reflect.ValueOf(awsStruct).Elem()
			field := v.FieldByName("Field1")
			awsFieldType := field.Type()
			if awsFieldType.Kind() == reflect.Ptr {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.CreateAWSValue(value)))
				}
			} else {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(value))
				}
			}
		}
	default: // standard
		if isNull {
			// Standard null behavior: null -> nil for pointers, zero for values
			v := reflect.ValueOf(awsStruct).Elem()
			field := v.FieldByName("Field1")
			awsFieldType := field.Type()
			// For null values with non-pointer AWS fields, set to zero value
			// For pointer fields, leave unset (nil is already the zero value)
			if awsFieldType.Kind() != reflect.Ptr {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.GetZeroValue()))
				}
			}
		} else {
			// Standard behavior: value -> aws.Xxx(value) or value
			v := reflect.ValueOf(awsStruct).Elem()
			field := v.FieldByName("Field1")
			awsFieldType := field.Type()
			if awsFieldType.Kind() == reflect.Ptr {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(typeInfo.CreateAWSValue(value)))
				}
			} else {
				if field.IsValid() && field.CanSet() {
					field.Set(reflect.ValueOf(value))
				}
			}
		}
	}

	// Full roundtrip test
	rtc := RoundtripTestCase[T]{
		Name:          testName,
		OriginalValue: value,
		TFStruct:      tfStruct,
		AWSStruct:     awsStruct,
		ExpectError:   false,
		Options:       nil,
		ExpectedDiags: nil, // No diagnostics expected for basic roundtrip tests
	}

	runRoundtripTest(t, rtc, checks)
}

// PrimitiveTestCase represents a test case for any primitive type
type PrimitiveTestCase[T any] struct {
	Name       string
	Value      T
	IsNull     bool
	IsZero     bool
	IsEmpty    bool // for strings only
	Variants   []string
	SkipExpand bool // for flatten-only tests
}

// PrimitiveTypeInfo contains type-specific information for testing primitives
type PrimitiveTypeInfo[T any] struct {
	TFType         reflect.Type
	CreateValue    func(T) any // creates types.XxxValue(v)
	CreateNull     func() any  // creates types.XxxNull()
	CreateAWSValue func(T) any // creates aws.Xxx(v)
	GetAWSNil      func() any  // creates (*type)(nil)
	GetZeroValue   func() T    // creates zero value for the type
}

type RoundtripTestCase[T any] struct {
	Name          string
	OriginalValue T
	TFStruct      any
	AWSStruct     any
	ExpectError   bool
	Options       []AutoFlexOptionsFunc
	ExpectedDiags diag.Diagnostics // Expected diagnostics for expand/flatten operations
}

// PrimitiveTestVariant defines different struct variants for testing
type PrimitiveTestVariant struct {
	Name    string
	Tag     string
	Factory func(fieldType reflect.Type) (tf, aws any)
}

// runRoundtripTest executes a complete roundtrip test: TF -> AWS -> TF
func runRoundtripTest[T any](t *testing.T, tc RoundtripTestCase[T], checks runChecks) {
	t.Helper()

	ctx := context.Background()

	// Set up logging if golden logs are requested
	var buf bytes.Buffer
	if !checks.SkipGoldenLogs {
		ctx = tflogtest.RootLogger(ctx, &buf)
		ctx = registerTestingLogger(ctx)
	}

	// Step 1: Expand TF -> AWS
	expandedAWS := reflect.New(reflect.TypeOf(tc.AWSStruct).Elem()).Interface()
	expandDiags := Expand(ctx, tc.TFStruct, expandedAWS, tc.Options...)

	// Check diagnostics if requested
	if checks.CompareDiags {
		if diff := cmp.Diff(expandDiags, tc.ExpectedDiags); diff != "" {
			t.Errorf("unexpected expand diagnostics difference: %s", diff)
		}
	}

	if tc.ExpectError {
		if !expandDiags.HasError() {
			t.Errorf("Expected error during expand, but got none")
		}
		return
	}

	if expandDiags.HasError() {
		t.Fatalf("Unexpected error during expand: %v", expandDiags)
	}

	// Step 2: Flatten the AWS struct back to TF
	actualTF := reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
	flattenDiags := Flatten(ctx, expandedAWS, actualTF, tc.Options...)

	// Check flatten diagnostics if requested (and we have expected flatten diags)
	// Note: We only check expand diagnostics above since that's the primary use case
	// but we could extend this to also check flatten diagnostics if needed

	if len(flattenDiags) > 0 {
		if tc.ExpectError {
			return
		}
		t.Fatalf("Unexpected flatten errors for %s: %v", tc.Name, flattenDiags)
	}

	flattenedTF := actualTF

	// Step 3: Verify roundtrip consistency (with known behavioral exceptions)
	expectedTF := tc.TFStruct

	// Handle known behavioral differences for null values
	if tc.Name != "" {
		// For null values: conversion to default values for variants that don't maintain null
		// (standard, legacy, omitempty use non-pointer AWS fields; legacy_pointer also converts null→default due to legacy flatten behavior)
		if strings.Contains(tc.Name, "null_value") && (!strings.Contains(tc.Name, "pointer") || strings.Contains(tc.Name, "legacy_tf_to_aws_pointer")) {
			// Detect field type from the struct
			fieldValue := reflect.ValueOf(tc.TFStruct).Elem().FieldByName("Field1")
			fieldType := fieldValue.Type()

			if fieldType == reflect.TypeFor[types.String]() {
				// null -> empty string for legacy string variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.StringValue("")))
			} else if fieldType == reflect.TypeFor[types.Bool]() {
				// null -> false for legacy bool variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.BoolValue(false)))
			} else if fieldType == reflect.TypeFor[types.Int64]() {
				// null -> 0 for legacy int64 variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.Int64Value(0)))
			} else if fieldType == reflect.TypeFor[types.Int32]() {
				// null -> 0 for legacy int32 variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.Int32Value(0)))
			} else if fieldType == reflect.TypeFor[types.Float64]() {
				// null -> 0.0 for all float64 variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.Float64Value(0.0)))
			} else if fieldType == reflect.TypeFor[types.Float32]() {
				// null -> 0.0 for all float32 variants
				expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
				reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.Float32Value(0.0)))
			}
		}
		// For omitempty: empty string -> null behavior
		if strings.Contains(tc.Name, "omitempty") && strings.Contains(tc.Name, "empty") {
			// Create expected TF with null instead of empty string
			expectedTF = reflect.New(reflect.TypeOf(tc.TFStruct).Elem()).Interface()
			reflect.ValueOf(expectedTF).Elem().FieldByName("Field1").Set(reflect.ValueOf(types.StringNull()))
		}
	}

	if checks.CompareTarget {
		if diff := cmp.Diff(expectedTF, flattenedTF); diff != "" {
			t.Errorf("Roundtrip mismatch for %s (+got, -want): %s", tc.Name, diff)
		}

		// Step 4: Verify AWS structure matches expected
		if diff := cmp.Diff(tc.AWSStruct, expandedAWS); diff != "" {
			t.Errorf("AWS structure mismatch for %s (+got, -want): %s", tc.Name, diff)
		}
	}

	// Golden log validation (if requested)
	if !checks.SkipGoldenLogs {
		lines, err := tflogtest.MultilineJSONDecode(&buf)
		if err != nil {
			t.Fatalf("decoding log lines: %s", err)
		}
		normalizedLines := normalizeLogs(lines)

		goldenFileName := autoGenerateGoldenPath(t, t.Name())
		goldenPath := filepath.Join("testdata", goldenFileName)
		compareWithGolden(t, goldenPath, normalizedLines)
	}
}

// runFlattenOnlyTest executes only the flatten direction: AWS -> TF
func runFlattenOnlyTest(t *testing.T, testName string, awsStruct, expectedTF any) {
	t.Helper()

	ctx := context.Background()

	// Flatten AWS -> TF
	actualTF := reflect.New(reflect.TypeOf(expectedTF).Elem()).Interface()
	flattenDiags := Flatten(ctx, awsStruct, actualTF)

	if flattenDiags.HasError() {
		t.Fatalf("Unexpected error during flatten: %v", flattenDiags)
	}

	// Verify TF structure matches expected
	if diff := cmp.Diff(expectedTF, actualTF); diff != "" {
		t.Errorf("Flatten result mismatch for %s (+got, -want): %s", testName, diff)
	}
}

// Standard primitive test variants
var primitiveTestVariants = []PrimitiveTestVariant{
	{
		Name: "standard",
		Tag:  `tfsdk:"field1"`,
		Factory: func(fieldType reflect.Type) (tf, aws any) {
			return generateStandardStructs(fieldType)
		},
	},
	{
		Name: "legacy",
		Tag:  `tfsdk:"field1" autoflex:",legacy"`,
		Factory: func(fieldType reflect.Type) (tf, aws any) {
			return generateLegacyStructs(fieldType)
		},
	},
	{
		Name: "omitempty",
		Tag:  `tfsdk:"field1" autoflex:",omitempty"`,
		Factory: func(fieldType reflect.Type) (tf, aws any) {
			return generateOmitEmptyStructs(fieldType)
		},
	},
	{
		Name: "tf_to_aws_pointer",
		Tag:  `tfsdk:"field1"`,
		Factory: func(fieldType reflect.Type) (tf, aws any) {
			return generateTFToAWSPointerStructs(fieldType)
		},
	},
	{
		Name: "legacy_tf_to_aws_pointer",
		Tag:  `tfsdk:"field1" autoflex:",legacy"`,
		Factory: func(fieldType reflect.Type) (tf, aws any) {
			return generateLegacyTFToAWSPointerStructs(fieldType)
		},
	},
}

// generateStandardStructs creates standard TF and AWS structs for testing
func generateStandardStructs(fieldType reflect.Type) (tf, aws any) {
	// Create TF struct with framework type
	tfStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: fieldType,
			Tag:  `tfsdk:"field1"`,
		},
	})
	tfStruct := reflect.New(tfStructType).Interface()

	// Create AWS struct based on field type
	var awsFieldType reflect.Type
	switch fieldType {
	case reflect.TypeFor[types.String]():
		awsFieldType = reflect.TypeFor[string]()
	case reflect.TypeFor[types.Bool]():
		awsFieldType = reflect.TypeFor[bool]()
	case reflect.TypeFor[types.Int64]():
		awsFieldType = reflect.TypeFor[int64]()
	case reflect.TypeFor[types.Int32]():
		awsFieldType = reflect.TypeFor[int32]()
	case reflect.TypeFor[types.Float64]():
		awsFieldType = reflect.TypeFor[float64]()
	case reflect.TypeFor[types.Float32]():
		awsFieldType = reflect.TypeFor[float32]()
	default:
		panic("unsupported field type")
	}

	awsStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: awsFieldType,
		},
	})
	awsStruct := reflect.New(awsStructType).Interface()

	return tfStruct, awsStruct
}

// generateLegacyStructs creates legacy-tagged TF structs paired with pointer AWS structs
func generateLegacyStructs(fieldType reflect.Type) (tf, aws any) {
	// Create TF struct with legacy tag
	tfStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: fieldType,
			Tag:  `tfsdk:"field1" autoflex:",legacy"`,
		},
	})
	tfStruct := reflect.New(tfStructType).Interface()

	// Create AWS struct with pointer field for legacy behavior
	var awsFieldType reflect.Type
	switch fieldType {
	case reflect.TypeFor[types.String]():
		awsFieldType = reflect.TypeFor[*string]()
	case reflect.TypeFor[types.Bool]():
		awsFieldType = reflect.TypeFor[*bool]()
	case reflect.TypeFor[types.Int64]():
		awsFieldType = reflect.TypeFor[*int64]()
	case reflect.TypeFor[types.Int32]():
		awsFieldType = reflect.TypeFor[*int32]()
	case reflect.TypeFor[types.Float64]():
		awsFieldType = reflect.TypeFor[*float64]()
	case reflect.TypeFor[types.Float32]():
		awsFieldType = reflect.TypeFor[*float32]()
	default:
		panic("unsupported field type")
	}

	awsStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: awsFieldType,
		},
	})
	awsStruct := reflect.New(awsStructType).Interface()

	return tfStruct, awsStruct
}

// generateOmitEmptyStructs creates omitempty-tagged TF structs for testing
func generateOmitEmptyStructs(fieldType reflect.Type) (tf, aws any) {
	// Create TF struct with omitempty tag
	tfStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: fieldType,
			Tag:  `tfsdk:"field1" autoflex:",omitempty"`,
		},
	})
	tfStruct := reflect.New(tfStructType).Interface()

	// For omitempty, AWS side uses pointer types
	var awsFieldType reflect.Type
	switch fieldType {
	case reflect.TypeFor[types.String]():
		awsFieldType = reflect.TypeFor[*string]()
	case reflect.TypeFor[types.Bool]():
		awsFieldType = reflect.TypeFor[*bool]()
	case reflect.TypeFor[types.Int64]():
		awsFieldType = reflect.TypeFor[*int64]()
	case reflect.TypeFor[types.Int32]():
		awsFieldType = reflect.TypeFor[*int32]()
	case reflect.TypeFor[types.Float64]():
		awsFieldType = reflect.TypeFor[*float64]()
	case reflect.TypeFor[types.Float32]():
		awsFieldType = reflect.TypeFor[*float32]()
	default:
		panic("unsupported field type")
	}

	awsStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: awsFieldType,
		},
	})
	awsStruct := reflect.New(awsStructType).Interface()

	return tfStruct, awsStruct
}

// generateTFToAWSPointerStructs creates value TF structs paired with pointer AWS structs
// Tests: types.String -> *string, types.Bool -> *bool, etc.
func generateTFToAWSPointerStructs(fieldType reflect.Type) (tf, aws any) { // nosemgrep:ci.aws-in-func-name
	// Create TF struct with value field type
	tfStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: fieldType, // Value type (types.String, types.Bool, etc.)
			Tag:  `tfsdk:"field1"`,
		},
	})
	tfStruct := reflect.New(tfStructType).Interface()

	// Create AWS struct with pointer field type
	var awsFieldType reflect.Type
	switch fieldType {
	case reflect.TypeFor[types.String]():
		awsFieldType = reflect.TypeFor[*string]()
	case reflect.TypeFor[types.Bool]():
		awsFieldType = reflect.TypeFor[*bool]()
	case reflect.TypeFor[types.Int64]():
		awsFieldType = reflect.TypeFor[*int64]()
	case reflect.TypeFor[types.Int32]():
		awsFieldType = reflect.TypeFor[*int32]()
	case reflect.TypeFor[types.Float64]():
		awsFieldType = reflect.TypeFor[*float64]()
	case reflect.TypeFor[types.Float32]():
		awsFieldType = reflect.TypeFor[*float32]()
	default:
		panic("unsupported field type")
	}

	awsStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: awsFieldType,
		},
	})
	awsStruct := reflect.New(awsStructType).Interface()

	return tfStruct, awsStruct
}

// generateLegacyTFToAWSPointerStructs creates legacy TF structs paired with pointer AWS structs
// Tests: types.String (legacy) -> *string, types.Bool (legacy) -> *bool, etc.
func generateLegacyTFToAWSPointerStructs(fieldType reflect.Type) (tf, aws any) { // nosemgrep:ci.aws-in-func-name
	// Create TF struct with legacy tag
	tfStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: fieldType, // Value type (types.String, types.Bool, etc.)
			Tag:  `tfsdk:"field1" autoflex:",legacy"`,
		},
	})
	tfStruct := reflect.New(tfStructType).Interface()

	// Create AWS struct with pointer field type
	var awsFieldType reflect.Type
	switch fieldType {
	case reflect.TypeFor[types.String]():
		awsFieldType = reflect.TypeFor[*string]()
	case reflect.TypeFor[types.Bool]():
		awsFieldType = reflect.TypeFor[*bool]()
	case reflect.TypeFor[types.Int64]():
		awsFieldType = reflect.TypeFor[*int64]()
	case reflect.TypeFor[types.Int32]():
		awsFieldType = reflect.TypeFor[*int32]()
	case reflect.TypeFor[types.Float64]():
		awsFieldType = reflect.TypeFor[*float64]()
	case reflect.TypeFor[types.Float32]():
		awsFieldType = reflect.TypeFor[*float32]()
	default:
		panic("unsupported field type")
	}

	awsStructType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Field1",
			Type: awsFieldType,
		},
	})
	awsStruct := reflect.New(awsStructType).Interface()

	return tfStruct, awsStruct
}
