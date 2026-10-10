// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"strings"
	"testing"

	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
)

func TestValidEndpointID(t *testing.T) {
	t.Parallel()

	validIds := []string{
		"tf-test-endpoint-1",
		"tfTestEndpoint",
	}

	for _, s := range validIds {
		_, errors := tfdms.ValidEndpointID(s, "endpoint_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid endpoint id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_endpoint_1",
		"tf.test.endpoint.1",
		"tf test endpoint 1",
		"tf-test-endpoint-1!",
		"tf-test-endpoint-1-",
		"tf-test-endpoint--1",
		"tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1tf-test-endpoint-1",
	}

	for _, s := range invalidIds {
		_, errors := tfdms.ValidEndpointID(s, "endpoint_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid endpoint id: %v", s, errors)
		}
	}
}

func TestValidReplicationInstanceID(t *testing.T) {
	t.Parallel()

	validIds := []string{
		"tf-test-replication-instance-1",
		"tfTestReplicaitonInstance",
	}

	for _, s := range validIds {
		_, errors := tfdms.ValidReplicationInstanceID(s, "replicaiton_instance_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication instance id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_replication-instance_1",
		"tf.test.replication.instance.1",
		"tf test replication instance 1",
		"tf-test-replication-instance-1!",
		"tf-test-replication-instance-1-",
		"tf-test-replication-instance--1",
		"tf-test-replication-instance-1tf-test-replication-instance-1tf-test-replication-instance-1",
	}

	for _, s := range invalidIds {
		_, errors := tfdms.ValidReplicationInstanceID(s, "replication_instance_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication instance id: %v", s, errors)
		}
	}
}

func TestValidReplicationSubnetGroupID(t *testing.T) {
	t.Parallel()

	validIds := []string{
		"tf-test-replication-subnet-group-1",
		"tf_test_replication_subnet_group_1",
		"tf.test.replication.subnet.group.1",
		"tf test replication subnet group 1",
		"tfTestReplicationSubnetGroup",
	}

	for _, s := range validIds {
		_, errors := tfdms.ValidReplicationSubnetGroupID(s, "replication_subnet_group_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication subnet group id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"default",
		"tf-test-replication-subnet-group-1!",
		"tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1tf-test-replication-subnet-group-1",
	}

	for _, s := range invalidIds {
		_, errors := tfdms.ValidReplicationSubnetGroupID(s, "replication_subnet_group_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication subnet group id: %v", s, errors)
		}
	}
}

func TestValidReplicationTaskID(t *testing.T) {
	t.Parallel()

	validIds := []string{
		"tf-test-replication-task-1",
		"tfTestReplicationTask",
	}

	for _, s := range validIds {
		_, errors := tfdms.ValidReplicationTaskID(s, "replication_task_id")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid replication task id: %v", s, errors)
		}
	}

	invalidIds := []string{
		"tf_test_replication_task_1",
		"tf.test.replication.task.1",
		"tf test replication task 1",
		"tf-test-replication-task-1!",
		"tf-test-replication-task-1-",
		"tf-test-replication-task--1",
		"tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1tf-test-replication-task-1",
	}

	for _, s := range invalidIds {
		_, errors := tfdms.ValidReplicationTaskID(s, "replication_task_id")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid replication task id: %v", s, errors)
		}
	}
}

// Test ARN components are declared once, with lintignore, and composed by
// secretARN so that individual test cases do not need their own directives.
const (
	testPartition   = "aws"           //lintignore:AWSAT005
	testPartitionGC = "aws-us-gov"    //lintignore:AWSAT005
	testPartitionCN = "aws-cn"        //lintignore:AWSAT005
	testRegion      = "us-east-1"     //lintignore:AWSAT003
	testRegionAlt   = "us-west-2"     //lintignore:AWSAT003
	testRegionGC    = "us-gov-west-1" //lintignore:AWSAT003
	testAccount     = "123456789012"
	testAccountAlt  = "210987654321"
)

func secretARN(partition, service, region, account, resource string) string {
	return strings.Join([]string{"arn", partition, service, region, account, resource}, ":")
}

func TestValidSecretsManagerSecretID(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value   string
		wantErr bool
	}{
		"full ARN":                  {value: secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:my-secret-a1b2c3")},
		"partial ARN":               {value: secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:my-secret")},
		"GovCloud ARN":              {value: secretARN(testPartitionGC, "secretsmanager", testRegionGC, testAccount, "secret:my-secret-a1b2c3")},
		"friendly name":             {value: "my-secret"},
		"friendly name with path":   {value: "/test/example/my-secret"},
		"friendly name all charset": {value: "a-Z_0/9+=.@"},
		"friendly name max length":  {value: strings.Repeat("a", 512)},
		"empty":                     {value: ""},
		"name too long":             {value: strings.Repeat("a", 513), wantErr: true},
		"name with space":           {value: "my secret", wantErr: true},
		"name with invalid char":    {value: "my-secret!", wantErr: true},
		"invalid ARN":               {value: "arn:" + testPartition + ":secretsmanager", wantErr: true},
		"ARN with invalid account":  {value: secretARN(testPartition, "secretsmanager", testRegion, "12345", "secret:my-secret"), wantErr: true},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, errs := tfdms.ValidSecretsManagerSecretID(testCase.value, "secrets_manager_arn")

			if got, want := len(errs) > 0, testCase.wantErr; got != want {
				t.Errorf("ValidSecretsManagerSecretID(%q) errors = %v, wantErr %t", testCase.value, errs, want)
			}
		})
	}
}

func TestSecretsManagerSecretIDMatchesFullARN(t *testing.T) {
	t.Parallel()

	var (
		fullARN        = secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:test/example/my-secret-a1b2c3")
		partialARN     = secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:test/example/my-secret")
		govCloudFull   = secretARN(testPartitionGC, "secretsmanager", testRegionGC, testAccount, "secret:my-secret-a1b2c3")
		govCloudShort  = secretARN(testPartitionGC, "secretsmanager", testRegionGC, testAccount, "secret:my-secret")
		otherAccount   = secretARN(testPartition, "secretsmanager", testRegion, testAccountAlt, "secret:test/example/my-secret")
		otherRegion    = secretARN(testPartition, "secretsmanager", testRegionAlt, testAccount, "secret:test/example/my-secret")
		otherPartition = secretARN(testPartitionCN, "secretsmanager", testRegion, testAccount, "secret:test/example/my-secret")
		otherSuffix    = secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:test/example/my-secret-d4e5f6")
		emptyName      = secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:-a1b2c3")
		otherService   = secretARN(testPartition, "ssm", testRegion, testAccount, "secret:my-secret-a1b2c3")
		ambiguous      = secretARN(testPartition, "secretsmanager", testRegion, testAccount, "secret:db-source")
	)

	testCases := map[string]struct {
		fullARN, id string
		want        bool
	}{
		"friendly name":               {fullARN: fullARN, id: "test/example/my-secret", want: true},
		"partial ARN":                 {fullARN: fullARN, id: partialARN, want: true},
		"GovCloud partial ARN":        {fullARN: govCloudFull, id: govCloudShort, want: true},
		"identical full ARN":          {fullARN: fullARN, id: fullARN},
		"different name":              {fullARN: fullARN, id: "test/example/other-secret"},
		"name is prefix only":         {fullARN: fullARN, id: "test/example/my"},
		"name with suffix":            {fullARN: fullARN, id: "test/example/my-secret-a1b2c3"},
		"partial ARN other account":   {fullARN: fullARN, id: otherAccount},
		"partial ARN other region":    {fullARN: fullARN, id: otherRegion},
		"partial ARN other partition": {fullARN: fullARN, id: otherPartition},
		"full ARN other suffix":       {fullARN: fullARN, id: otherSuffix},
		"name too short for suffix":   {fullARN: emptyName, id: ""},
		"not an ARN":                  {fullARN: "test/example/my-secret-a1b2c3", id: "test/example/my-secret"},
		"non-Secrets Manager ARN":     {fullARN: otherService, id: "my-secret"},
		// A partial ARN whose name ends in a hyphen and six characters cannot be told
		// apart from a full ARN, which is why only values returned by the API are
		// treated as full ARNs.
		"ambiguous partial ARN": {fullARN: ambiguous, id: "db", want: true},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := tfdms.SecretsManagerSecretIDMatchesFullARN(testCase.fullARN, testCase.id), testCase.want; got != want {
				t.Errorf("SecretsManagerSecretIDMatchesFullARN(%q, %q) = %t, want %t", testCase.fullARN, testCase.id, got, want)
			}
		})
	}
}
