// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/tools/check-basic-test/attrnames"
)

func TestParseFrameworkSchema_inline(t *testing.T) {
	src := `package example

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *res) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseFrameworkSchema(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{
		"id":           true,
		"name":         true,
		"description":  true,
		"config":       true,
		"config.key":   true,
		"config.value": true,
	}

	got := make(map[string]bool)
	for _, a := range attrs {
		got[a.Path] = true
	}

	for path := range want {
		if !got[path] {
			t.Errorf("missing expected attribute: %s", path)
		}
	}
	for path := range got {
		if !want[path] {
			t.Errorf("unexpected attribute: %s", path)
		}
	}

	// Check properties
	for _, a := range attrs {
		switch a.Path {
		case "id":
			if !a.Computed || a.Required || a.Optional {
				t.Errorf("id: want Computed only, got R=%v O=%v C=%v", a.Required, a.Optional, a.Computed)
			}
		case "name":
			if !a.Required || a.Optional || a.Computed {
				t.Errorf("name: want Required only, got R=%v O=%v C=%v", a.Required, a.Optional, a.Computed)
			}
		case "description":
			if !a.Optional || !a.Computed || a.Required {
				t.Errorf("description: want Optional+Computed, got R=%v O=%v C=%v", a.Required, a.Optional, a.Computed)
			}
		}
	}
}

func TestParseFrameworkSchema_varAssignment(t *testing.T) {
	src := `package example

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func (r *res) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
		},
	}
	response.Schema = s
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseFrameworkSchema(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	paths := make([]string, len(attrs))
	for i, a := range attrs {
		paths[i] = a.Path
	}

	if !slices.Contains(paths, "arn") {
		t.Error("missing 'arn'")
	}
	if !slices.Contains(paths, "name") {
		t.Error("missing 'name'")
	}
}

func TestParseFrameworkSchema_timeoutsSkipped(t *testing.T) {
	src := `package example

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
)

func (r *res) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}
`
	tmpFile := t.TempDir() + "/test.go"
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	attrs, err := ParseFrameworkSchema(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	for _, a := range attrs {
		if a.Path == "timeouts" {
			t.Error("timeouts should be skipped")
		}
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"ARN", "arn"},
		{"TableName", "table_name"},
		{"ID", "id"},
		{"NetworkInterfaceID", "network_interface_id"},
		{"DeleteOnTermination", "delete_on_termination"},
		{"Bucket", "bucket"},
		{"ExpectedBucketOwner", "expected_bucket_owner"},
		{"Status", "status"},
		{"CIDRBlock", "cidr_block"},
	}
	for _, tt := range tests {
		got := attrnames.CamelToSnake(tt.in)
		if got != tt.want {
			t.Errorf("CamelToSnake(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
