package kinesis

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

func TestFlattenShardLevelMetrics(t *testing.T) {
	expanded := []*kinesis.EnhancedMetrics{
		{
			ShardLevelMetrics: []*string{
				aws.String("IncomingBytes"),
				aws.String("IncomingRecords"),
			},
		},
	}
	result := flattenShardLevelMetrics(expanded)
	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}
	if result[0] != "IncomingBytes" {
		t.Fatalf("expected element 0 to be IncomingBytes, but was %s", result[0])
	}
	if result[1] != "IncomingRecords" {
		t.Fatalf("expected element 0 to be IncomingRecords, but was %s", result[1])
	}
}
