// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// lintignore:AWSAT003,AWSAT005
package arcregionswitch

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFlattenRoute53HealthChecks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		healthChecks []types.Route53HealthCheck
		expected     []any
	}{
		{
			name:         "empty health checks",
			healthChecks: []types.Route53HealthCheck{},
			expected:     nil,
		},
		{
			name: "single health check",
			healthChecks: []types.Route53HealthCheck{
				{
					HealthCheckId: aws.String("hc-123"),
					HostedZoneId:  aws.String("Z123"),
					RecordName:    aws.String("api.example.com"),
					Region:        aws.String("us-east-1"),
				},
			},
			expected: []any{
				map[string]any{
					"health_check_id":      "hc-123",
					names.AttrHostedZoneID: "Z123",
					"record_name":          "api.example.com",
					names.AttrRegion:       "us-east-1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := flattenRoute53HealthChecks(tt.healthChecks)
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("flattenRoute53HealthChecks() = %v, want %v", result, tt.expected)
			}
		})
	}
}
