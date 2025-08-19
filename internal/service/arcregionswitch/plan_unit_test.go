package arcregionswitch

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
)

func TestFlattenRoute53HealthChecks(t *testing.T) {
	tests := []struct {
		name         string
		healthChecks []types.Route53HealthCheck
		expected     []interface{}
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
			expected: []interface{}{
				map[string]interface{}{
					"health_check_id": "hc-123",
					"hosted_zone_id":  "Z123",
					"record_name":     "api.example.com",
					"region":          "us-east-1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenRoute53HealthChecks(tt.healthChecks)
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("flattenRoute53HealthChecks() = %v, want %v", result, tt.expected)
			}
		})
	}
}
