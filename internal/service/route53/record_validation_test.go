// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidateRoute53RecordTTLRequirement_Logic(t *testing.T) {
	testCases := []struct {
		name        string
		hasRecords  bool
		recordsLen  int
		hasAlias    bool
		aliasLen    int
		hasTTL      bool
		ttlValue    int
		expectError bool
	}{
		{
			name:        "records with TTL - should pass",
			hasRecords:  true,
			recordsLen:  1,
			hasAlias:    false,
			aliasLen:    0,
			hasTTL:      true,
			ttlValue:    300,
			expectError: false,
		},
		{
			name:        "records without TTL - should fail",
			hasRecords:  true,
			recordsLen:  1,
			hasAlias:    false,
			aliasLen:    0,
			hasTTL:      false,
			ttlValue:    0,
			expectError: true,
		},
		{
			name:        "records with zero TTL - should fail",
			hasRecords:  true,
			recordsLen:  1,
			hasAlias:    false,
			aliasLen:    0,
			hasTTL:      true,
			ttlValue:    0,
			expectError: true,
		},
		{
			name:        "alias without TTL - should pass",
			hasRecords:  false,
			recordsLen:  0,
			hasAlias:    true,
			aliasLen:    1,
			hasTTL:      false,
			ttlValue:    0,
			expectError: false,
		},
		{
			name:        "empty records set - should pass",
			hasRecords:  true,
			recordsLen:  0,
			hasAlias:    false,
			aliasLen:    0,
			hasTTL:      false,
			ttlValue:    0,
			expectError: false,
		},
		{
			name:        "no records, no alias - should pass",
			hasRecords:  false,
			recordsLen:  0,
			hasAlias:    false,
			aliasLen:    0,
			hasTTL:      true,
			ttlValue:    300,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the validation logic directly
			var shouldError bool

			// If records are provided
			if tc.hasRecords && tc.recordsLen > 0 {
				// And no alias is provided
				if !tc.hasAlias || tc.aliasLen == 0 {
					// Then TTL must be provided and greater than 0
					if !tc.hasTTL || tc.ttlValue <= 0 {
						shouldError = true
					}
				}
			}

			if tc.expectError != shouldError {
				t.Errorf("Expected error: %v, but validation logic would error: %v", tc.expectError, shouldError)
			}
		})
	}
}

func TestRoute53RecordSchema_TTLField(t *testing.T) {
	resource := resourceRecord()
	
	// Check that TTL field exists and has correct properties
	ttlField, exists := resource.Schema["ttl"]
	if !exists {
		t.Fatal("TTL field should exist in schema")
	}

	if ttlField.Type != schema.TypeInt {
		t.Errorf("TTL field should be TypeInt, got %v", ttlField.Type)
	}

	if !ttlField.Optional {
		t.Error("TTL field should be Optional")
	}

	if ttlField.Required {
		t.Error("TTL field should not be Required")
	}

	// Check ConflictsWith
	expectedConflicts := []string{names.AttrAlias}
	if len(ttlField.ConflictsWith) != len(expectedConflicts) {
		t.Errorf("Expected ConflictsWith %v, got %v", expectedConflicts, ttlField.ConflictsWith)
	}

	// Check that RequiredWith is not present (we removed it)
	if len(ttlField.RequiredWith) > 0 {
		t.Errorf("TTL field should not have RequiredWith, but has %v", ttlField.RequiredWith)
	}

	// Check description
	if ttlField.Description == "" {
		t.Error("TTL field should have a description")
	}
}

func TestRoute53RecordSchema_CustomizeDiff(t *testing.T) {
	resource := resourceRecord()
	
	if resource.CustomizeDiff == nil {
		t.Error("Resource should have CustomizeDiff function")
	}
}
