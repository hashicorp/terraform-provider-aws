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

func TestValidSecretsManagerSecretID(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value   string
		wantErr bool
	}{
		"full ARN":                  {value: "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret-a1b2c3"},            //lintignore:AWSAT003,AWSAT005
		"partial ARN":               {value: "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret"},                   //lintignore:AWSAT003,AWSAT005
		"GovCloud ARN":              {value: "arn:aws-us-gov:secretsmanager:us-gov-west-1:123456789012:secret:my-secret-a1b2c3"}, //lintignore:AWSAT003,AWSAT005
		"friendly name":             {value: "my-secret"},
		"friendly name with path":   {value: "/test/example/my-secret"},
		"friendly name all charset": {value: "a-Z_0/9+=.@"},
		"friendly name max length":  {value: strings.Repeat("a", 512)},
		"empty":                     {value: "", wantErr: true},
		"name too long":             {value: strings.Repeat("a", 513), wantErr: true},
		"name with space":           {value: "my secret", wantErr: true},
		"name with invalid char":    {value: "my-secret!", wantErr: true},
		"invalid ARN":               {value: "arn:aws:secretsmanager", wantErr: true},
		"ARN with invalid account":  {value: "arn:aws:secretsmanager:us-east-1:12345:secret:my-secret", wantErr: true}, //lintignore:AWSAT003,AWSAT005
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

func TestSuppressSecretsManagerSecretIDDiffs(t *testing.T) {
	t.Parallel()

	const (
		fullARN    = "arn:aws:secretsmanager:us-east-1:123456789012:secret:test/example/my-secret-a1b2c3" //lintignore:AWSAT003,AWSAT005
		partialARN = "arn:aws:secretsmanager:us-east-1:123456789012:secret:test/example/my-secret"        //lintignore:AWSAT003,AWSAT005
	)

	testCases := map[string]struct {
		old, new string
		want     bool
	}{
		"identical":                         {old: fullARN, new: fullARN, want: true},
		"full ARN in state, name in config": {old: fullARN, new: "test/example/my-secret", want: true},
		"name in state, full ARN in config": {old: "test/example/my-secret", new: fullARN, want: true},
		"full ARN in state, partial ARN":    {old: fullARN, new: partialARN, want: true},
		"partial ARN in state, full ARN":    {old: partialARN, new: fullARN, want: true},
		"different name":                    {old: fullARN, new: "test/example/other-secret"},
		"name is prefix only":               {old: fullARN, new: "test/example/my"},
		"different partial ARN account":     {old: fullARN, new: "arn:aws:secretsmanager:us-east-1:210987654321:secret:test/example/my-secret"},        //lintignore:AWSAT003,AWSAT005
		"different full ARN suffix":         {old: fullARN, new: "arn:aws:secretsmanager:us-east-1:123456789012:secret:test/example/my-secret-d4e5f6"}, //lintignore:AWSAT003,AWSAT005
		"name vs partial ARN":               {old: partialARN, new: "test/example/my-secret"},
		"new value removed":                 {old: fullARN, new: ""},
		"new value added":                   {old: "", new: "test/example/my-secret"},
		"non-Secrets Manager ARN":           {old: "arn:aws:ssm:us-east-1:123456789012:secret:my-secret-a1b2c3", new: "my-secret"}, //lintignore:AWSAT003,AWSAT005
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := tfdms.SuppressSecretsManagerSecretIDDiffs("secrets_manager_arn", testCase.old, testCase.new, nil), testCase.want; got != want {
				t.Errorf("SuppressSecretsManagerSecretIDDiffs(%q, %q) = %t, want %t", testCase.old, testCase.new, got, want)
			}
		})
	}
}
