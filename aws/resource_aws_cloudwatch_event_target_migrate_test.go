package aws

import (
	"context"
	"reflect"
	"testing"
)

func testresourceAwsCloudWatchEventTargetStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"arn":       "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"rule":      "testrule",
		"target_id": "testtargetid",
	}
}

func testresourceAwsCloudWatchEventTargetStateDataV0EventBusName() map[string]interface{} {
	return map[string]interface{}{
		"arn":            "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"event_bus_name": "testbus",
		"rule":           "testrule",
		"target_id":      "testtargetid",
	}
}

func testresourceAwsCloudWatchEventTargetStateDataV1() map[string]interface{} {
	v0 := testresourceAwsCloudWatchEventTargetStateDataV0()
	return map[string]interface{}{
		"arn":            v0["arn"],
		"event_bus_name": "default",
		"rule":           v0["rule"],
		"target_id":      v0["target_id"],
	}
}

func testresourceAwsCloudWatchEventTargetStateDataV1EventBusName() map[string]interface{} {
	v0 := testresourceAwsCloudWatchEventTargetStateDataV0EventBusName()
	return map[string]interface{}{
		"arn":            v0["arn"],
		"event_bus_name": v0["event_bus_name"],
		"rule":           v0["rule"],
		"target_id":      v0["target_id"],
	}
}

func TestresourceAwsCloudWatchEventTargetStateUpgradeV0(t *testing.T) {
	expected := testresourceAwsCloudWatchEventTargetStateDataV1()
	actual, err := resourceAwsCloudWatchEventTargetStateUpgradeV0(context.Background(), testresourceAwsCloudWatchEventTargetStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func TestresourceAwsCloudWatchEventTargetStateUpgradeV0EventBusName(t *testing.T) {
	expected := testresourceAwsCloudWatchEventTargetStateDataV1EventBusName()
	actual, err := resourceAwsCloudWatchEventTargetStateUpgradeV0(context.Background(), testresourceAwsCloudWatchEventTargetStateDataV0EventBusName(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
