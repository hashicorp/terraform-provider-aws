package aws

import (
	"testing"
)

func TestArn_iamRootUser(t *testing.T) {
	arn := iamArnString("aws", "1234567890", "root")
	expectedArn := "arn:aws:iam::1234567890:root"
	if arn != expectedArn {
		t.Fatalf("Expected ARN: %s, got: %s", expectedArn, arn)
	}
}
