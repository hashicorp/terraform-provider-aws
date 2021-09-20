package apigateway

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
)

func TestFlattenThrottleSettings(t *testing.T) {
	expectedBurstLimit := int64(140)
	expectedRateLimit := 120.0

	ts := &apigateway.ThrottleSettings{
		BurstLimit: aws.Int64(expectedBurstLimit),
		RateLimit:  aws.Float64(expectedRateLimit),
	}
	result := FlattenThrottleSettings(ts)

	if len(result) != 1 {
		t.Fatalf("Expected map to have exactly 1 element, got %d", len(result))
	}

	burstLimit, ok := result[0]["burst_limit"]
	if !ok {
		t.Fatal("Expected 'burst_limit' key in the map")
	}
	burstLimitInt, ok := burstLimit.(int64)
	if !ok {
		t.Fatal("Expected 'burst_limit' to be int")
	}
	if burstLimitInt != expectedBurstLimit {
		t.Fatalf("Expected 'burst_limit' to equal %d, got %d", expectedBurstLimit, burstLimitInt)
	}

	rateLimit, ok := result[0]["rate_limit"]
	if !ok {
		t.Fatal("Expected 'rate_limit' key in the map")
	}
	rateLimitFloat, ok := rateLimit.(float64)
	if !ok {
		t.Fatal("Expected 'rate_limit' to be float64")
	}
	if rateLimitFloat != expectedRateLimit {
		t.Fatalf("Expected 'rate_limit' to equal %f, got %f", expectedRateLimit, rateLimitFloat)
	}
}
