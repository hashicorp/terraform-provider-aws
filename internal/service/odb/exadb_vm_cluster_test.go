// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExaDBVMClusterResourceSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resourceUnderTest, err := newExaDBVMClusterResource(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}

	var response resource.SchemaResponse
	resourceUnderTest.Schema(ctx, resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("building schema: %v", response.Diagnostics)
	}
	if diagnostics := response.Schema.ValidateImplementation(ctx); diagnostics.HasError() {
		t.Fatalf("validating schema implementation: %v", diagnostics)
	}

	testCases := map[string]struct {
		required bool
		optional bool
		computed bool
	}{
		names.AttrARN:                   {computed: true},
		names.AttrClusterName:           {optional: true, computed: true},
		names.AttrCreatedAt:             {computed: true},
		names.AttrDisplayName:           {required: true},
		names.AttrDomain:                {computed: true},
		"enabled_ecpu_count":            {required: true},
		"exascale_db_storage_vault_arn": {computed: true},
		"exascale_db_storage_vault_id":  {required: true},
		"gi_version":                    {computed: true},
		"grid_image_id":                 {required: true},
		"grid_image_type":               {computed: true},
		"hostname":                      {required: true},
		"iam_roles":                     {computed: true},
		"iorm_config_cache":             {computed: true},
		"last_update_history_entry_id":  {computed: true},
		"license_model":                 {optional: true, computed: true},
		"listener_port":                 {computed: true},
		"memory_size_in_gbs":            {computed: true},
		"node_count":                    {required: true},
		"ocid":                          {computed: true},
		"oci_resource_anchor_name":      {computed: true},
		"oci_url":                       {computed: true},
		"odb_network_arn":               {computed: true},
		"odb_network_id":                {required: true},
		"percent_progress":              {computed: true},
		"scan_dns_name":                 {computed: true},
		"scan_dns_record_id":            {computed: true},
		"scan_ip_ids":                   {computed: true},
		"scan_listener_port_tcp":        {optional: true, computed: true},
		"scan_listener_port_tcp_ssl":    {optional: true, computed: true},
		"shape":                         {required: true},
		"shape_attribute":               {optional: true, computed: true},
		"snapshot_file_system_storage":  {computed: true},
		"ssh_public_keys":               {required: true},
		names.AttrStatus:                {computed: true},
		names.AttrStatusReason:          {computed: true},
		"system_version":                {optional: true, computed: true},
		names.AttrTags:                  {optional: true},
		names.AttrTagsAll:               {computed: true},
		"time_zone":                     {optional: true, computed: true},
		"total_ecpu_count":              {required: true},
		"total_file_system_storage":     {computed: true},
		"vip_ids":                       {computed: true},
		"vm_file_system_storage":        {computed: true},
		"vm_file_system_storage_total_size_in_gbs": {required: true},
		names.AttrID: {computed: true},
	}

	if got, want := len(response.Schema.Attributes), len(testCases); got != want {
		t.Errorf("schema attribute count = %d, want %d", got, want)
	}

	for name, expected := range testCases {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			t.Errorf("expected schema attribute %q", name)
			continue
		}

		if got := attribute.IsRequired(); got != expected.required {
			t.Errorf("attribute %q Required = %t, want %t", name, got, expected.required)
		}
		if got := attribute.IsOptional(); got != expected.optional {
			t.Errorf("attribute %q Optional = %t, want %t", name, got, expected.optional)
		}
		if got := attribute.IsComputed(); got != expected.computed {
			t.Errorf("attribute %q Computed = %t, want %t", name, got, expected.computed)
		}
	}

	shapeAttribute, ok := response.Schema.Attributes["shape"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("shape attribute has type %T, want schema.StringAttribute", response.Schema.Attributes["shape"])
	}
	if shapeAttribute.CustomType == nil || !shapeAttribute.CustomType.Equal(fwtypes.CaseInsensitiveStringType) {
		t.Errorf("shape custom type = %T, want fwtypes.CaseInsensitiveStringType", shapeAttribute.CustomType)
	}

	dataCollectionBlock, ok := response.Schema.Blocks["data_collection_options"]
	if !ok {
		t.Fatal("expected schema block \"data_collection_options\"")
	}

	dataCollectionOptions, ok := dataCollectionBlock.(schema.ListNestedBlock)
	if !ok {
		t.Fatalf("data_collection_options block has type %T, want schema.ListNestedBlock", dataCollectionBlock)
	}

	for _, name := range []string{
		"is_diagnostics_events_enabled",
		"is_health_monitoring_enabled",
		"is_incident_logs_enabled",
	} {
		attribute, ok := dataCollectionOptions.NestedObject.Attributes[name]
		if !ok {
			t.Errorf("expected data_collection_options attribute %q", name)
			continue
		}
		if !attribute.IsOptional() || !attribute.IsComputed() {
			t.Errorf("data_collection_options attribute %q must be optional and computed", name)
		}
	}

	if _, ok := response.Schema.Blocks[names.AttrTimeouts]; !ok {
		t.Errorf("expected schema block %q", names.AttrTimeouts)
	}
}

func TestExaDBVMClusterFlattenDataCollectionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName                     string
		Mode                         exaDBVMClusterFlattenMode
		WantNull                     bool
		WantDiagnosticsEventsEnabled bool
		WantHealthMonitoringEnabled  bool
		WantIncidentLogsEnabled      bool
	}{
		{
			TestName: "managed resource preserves omitted configuration",
			Mode:     exaDBVMClusterFlattenModeManagedResource,
			WantNull: true,
		},
		{
			TestName:                     "list resource includes API values",
			Mode:                         exaDBVMClusterFlattenModeListResource,
			WantDiagnosticsEventsEnabled: true,
			WantHealthMonitoringEnabled:  false,
			WantIncidentLogsEnabled:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			data := exaDBVMClusterResourceModel{
				DataCollectionOptions: fwtypes.NewListNestedObjectValueOfNull[exaDBVMClusterDataCollectionOptionsModel](ctx),
			}
			apiObject := &odbtypes.ExadbVmCluster{
				DataCollectionOptions: &odbtypes.DataCollectionOptions{
					IsDiagnosticsEventsEnabled: aws.Bool(true),
					IsHealthMonitoringEnabled:  aws.Bool(false),
					IsIncidentLogsEnabled:      aws.Bool(true),
				},
			}

			diags := (&exaDBVMClusterResource{}).flatten(ctx, apiObject, &data, tc.Mode)
			if diags.HasError() {
				t.Fatalf("flattening ExaDB VM Cluster: %v", diags)
			}
			if got := data.DataCollectionOptions.IsNull(); got != tc.WantNull {
				t.Fatalf("data_collection_options null = %t, want %t", got, tc.WantNull)
			}
			if tc.WantNull {
				return
			}

			options, diags := data.DataCollectionOptions.ToPtr(ctx)
			if diags.HasError() {
				t.Fatalf("converting data_collection_options: %v", diags)
			}
			if options == nil {
				t.Fatal("data_collection_options is nil")
			}
			if got := options.IsDiagnosticsEventsEnabled.ValueBool(); got != tc.WantDiagnosticsEventsEnabled {
				t.Errorf("is_diagnostics_events_enabled = %t, want %t", got, tc.WantDiagnosticsEventsEnabled)
			}
			if got := options.IsHealthMonitoringEnabled.ValueBool(); got != tc.WantHealthMonitoringEnabled {
				t.Errorf("is_health_monitoring_enabled = %t, want %t", got, tc.WantHealthMonitoringEnabled)
			}
			if got := options.IsIncidentLogsEnabled.ValueBool(); got != tc.WantIncidentLogsEnabled {
				t.Errorf("is_incident_logs_enabled = %t, want %t", got, tc.WantIncidentLogsEnabled)
			}
		})
	}
}
