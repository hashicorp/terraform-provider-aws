package arcregionswitch

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
)

func TestExtractHealthCheckIds(t *testing.T) {
	tests := []struct {
		name     string
		events   []types.ExecutionEvent
		expected []string
	}{
		{
			name:     "empty events",
			events:   []types.ExecutionEvent{},
			expected: nil,
		},
		{
			name: "single health check",
			events: []types.ExecutionEvent{
				{
					EventId:   aws.String("event-1"),
					StepName:  aws.String("Route53HealthCheck"),
					Resources: []string{"health-check-123"},
				},
			},
			expected: []string{"health-check-123"},
		},
		{
			name: "multiple health checks",
			events: []types.ExecutionEvent{
				{
					EventId:   aws.String("event-1"),
					StepName:  aws.String("Route53HealthCheck"),
					Resources: []string{"health-check-123", "health-check-456"},
				},
				{
					EventId:   aws.String("event-2"),
					StepName:  aws.String("Route53HealthCheck"),
					Resources: []string{"health-check-789"},
				},
			},
			expected: []string{"health-check-123", "health-check-456", "health-check-789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHealthCheckIds(tt.events)
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("ExtractHealthCheckIds() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFlattenExecutionEvents(t *testing.T) {
	tests := []struct {
		name     string
		events   []types.ExecutionEvent
		expected []interface{}
	}{
		{
			name:     "empty events",
			events:   []types.ExecutionEvent{},
			expected: nil,
		},
		{
			name: "single event",
			events: []types.ExecutionEvent{
				{
					EventId:   aws.String("event-1"),
					StepName:  aws.String("Route53HealthCheck"),
					Resources: []string{"health-check-123"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"event_id":  "event-1",
					"step_name": "Route53HealthCheck",
					"resources": []string{"health-check-123"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenExecutionEvents(tt.events)
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("FlattenExecutionEvents() = %v, want %v", result, tt.expected)
			}
		})
	}
}
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
