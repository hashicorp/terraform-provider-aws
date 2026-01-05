// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestContactFlowModuleSchema_Settings(t *testing.T) {
	resource := resourceContactFlowModule()

	if resource == nil {
		t.Fatal("Resource is nil")
	}

	resourceSchema := resource.Schema
	if resourceSchema == nil {
		t.Fatal("Schema is nil")
	}

	// Check if settings parameter exists
	settingsSchema, exists := resourceSchema["settings"]
	if !exists {
		t.Fatal("Settings parameter NOT found in schema")
	}

	// Verify settings schema properties
	if settingsSchema.Type != schema.TypeString {
		t.Errorf("Expected settings type to be TypeString, got %v", settingsSchema.Type)
	}

	if !settingsSchema.Optional {
		t.Error("Expected settings to be optional")
	}

	if settingsSchema.ValidateFunc == nil {
		t.Error("Expected settings to have ValidateFunc (JSON validation)")
	}

	if settingsSchema.DiffSuppressFunc == nil {
		t.Error("Expected settings to have DiffSuppressFunc (JSON diff suppression)")
	}

	if settingsSchema.StateFunc == nil {
		t.Error("Expected settings to have StateFunc (JSON normalization)")
	}

	t.Log("✅ Settings parameter schema validation passed")
}

func TestContactFlowModuleSchema_AllParameters(t *testing.T) {
	resource := resourceContactFlowModule()
	resourceSchema := resource.Schema

	expectedParams := []string{
		"arn",
		"contact_flow_module_id",
		"content",
		"content_hash",
		"description",
		"filename",
		"instance_id",
		"name",
		"settings", // Our new parameter
		"tags",
		"tags_all",
	}

	for _, param := range expectedParams {
		if _, exists := resourceSchema[param]; !exists {
			t.Errorf("Expected parameter %s not found in schema", param)
		}
	}

	t.Logf("✅ All %d expected parameters found in schema", len(expectedParams))
}
