package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestExpandReplicationRuleFilterTag(t *testing.T) {
	t.Parallel()

	expectedKey := "TestKey1"
	expectedValue := "TestValue1"

	tagMap := map[string]interface{}{
		"key":   expectedKey,
		"value": expectedValue,
	}

	result := ExpandReplicationRuleFilterTag([]interface{}{tagMap})

	if result == nil {
		t.Fatalf("Expected *s3.Tag to not be nil")
	}

	if actualKey := aws.StringValue(result.Key); actualKey != expectedKey {
		t.Fatalf("Expected key %s, got %s", expectedKey, actualKey)
	}

	if actualValue := aws.StringValue(result.Value); actualValue != expectedValue {
		t.Fatalf("Expected value %s, got %s", expectedValue, actualValue)
	}
}

func TestFlattenReplicationRuleFilterTag(t *testing.T) {
	t.Parallel()

	expectedKey := "TestKey1"
	expectedValue := "TestValue1"

	tag := &s3.Tag{
		Key:   aws.String(expectedKey),
		Value: aws.String(expectedValue),
	}

	result := FlattenReplicationRuleFilterTag(tag)

	if len(result) != 1 {
		t.Fatalf("Expected array to have exactly 1 element, got %d", len(result))
	}

	tagMap, ok := result[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected element in array to be a map[string]interface{}")
	}

	actualKey, ok := tagMap["key"].(string)
	if !ok {
		t.Fatal("Expected string 'key' key in the map")
	}

	if actualKey != expectedKey {
		t.Fatalf("Expected 'key' to equal %s, got %s", expectedKey, actualKey)
	}

	actualValue, ok := tagMap["value"].(string)
	if !ok {
		t.Fatal("Expected string 'value' key in the map")
	}

	if actualValue != expectedValue {
		t.Fatalf("Expected 'value' to equal %s, got %s", expectedValue, actualValue)
	}
}
