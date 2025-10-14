// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

// Args tests validate top-level argument shape (nil/typed-nil, pointer-ness, struct↔non-struct).
// They intentionally do not assert logging; only diagnostic codes.

import (
	"testing"
)

type emptyStruct struct{}

func TestExpandArgs_nilAndPointers(t *testing.T) {
	t.Parallel()

	var (
		typedNilSource *emptyStruct
		typedNilTarget *emptyStruct
	)

	testCases := autoFlexTestCases{
		"nil Source": {
			Target:        &emptyStruct{},
			expectedDiags: diagAFNil(diagExpandingSourceIsNil),
		},
		"typed nil Source": {
			Source:        typedNilSource,
			Target:        &emptyStruct{},
			expectedDiags: diagAFNil(diagExpandingSourceIsNil), // FIXME: Should give the actual type
		},
		"nil Target": {
			Source:        emptyStruct{},
			expectedDiags: diagAFNil(diagConvertingTargetIsNil),
		},
		"typed nil Target": {
			Source:        emptyStruct{},
			Target:        typedNilTarget,
			expectedDiags: diagAF[*emptyStruct](diagConvertingTargetIsNil),
		},
		"non-pointer Target": {
			Source:        emptyStruct{},
			Target:        0,
			expectedDiags: diagAF[int](diagConvertingTargetIsNotPointer),
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandArgs_shapeCompatibility(t *testing.T) {
	t.Parallel()

	testString := "test"

	testCases := autoFlexTestCases{
		"non-struct Source struct Target": {
			Source:        testString,
			Target:        &emptyStruct{},
			expectedDiags: diagAF[string](diagExpandingSourceDoesNotImplementAttrValue),
		},
		"struct Source non-struct Target": {
			Source:        emptyStruct{},
			Target:        &testString,
			expectedDiags: diagAF[emptyStruct](diagExpandingSourceDoesNotImplementAttrValue),
		},
		"empty struct Source and Target": {
			Source:     emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
		"empty struct pointer Source and Target": {
			Source:     &emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenArgs_nilAndPointers(t *testing.T) {
	t.Parallel()

	var (
		typedNilSource *emptyStruct
		typedNilTarget *emptyStruct
	)

	testCases := autoFlexTestCases{
		"nil Source": {
			Target:        &emptyStruct{},
			expectedDiags: diagAFNil(diagFlatteningSourceIsNil),
		},
		"typed nil Source": {
			Source:        typedNilSource,
			Target:        &emptyStruct{},
			expectedDiags: diagAF[*emptyStruct](diagFlatteningSourceIsNil),
		},
		"nil Target": {
			Source:        emptyStruct{},
			expectedDiags: diagAFNil(diagConvertingTargetIsNil),
		},
		"typed nil Target": {
			Source:        emptyStruct{},
			Target:        typedNilTarget,
			expectedDiags: diagAF[*emptyStruct](diagConvertingTargetIsNil),
		},
		"non-pointer Target": {
			Source:        emptyStruct{},
			Target:        0,
			expectedDiags: diagAF[int](diagConvertingTargetIsNotPointer),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenArgs_shapeCompatibility(t *testing.T) {
	t.Parallel()

	testString := "test"

	testCases := autoFlexTestCases{
		"non-struct Source struct Target": {
			Source:        testString,
			Target:        &emptyStruct{},
			expectedDiags: diagAF[emptyStruct](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"struct Source non-struct Target": {
			Source:        emptyStruct{},
			Target:        &testString,
			expectedDiags: diagAF[string](diagFlatteningTargetDoesNotImplementAttrValue),
		},
		"empty struct Source and Target": {
			Source:     emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
		"empty struct pointer Source and Target": {
			Source:     &emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}
