// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlattenECSProperties(t *testing.T) {
	t.Parallel()

	if got, err := flattenECSProperties(nil); err != nil || got != "" {
		t.Fatalf("flattenECSProperties(nil) = %q, %v; want empty string, nil", got, err)
	}

	const want = `
{
  "taskProperties": [
    {
      "containers": [
        {
          "command": ["serve"],
          "environment": [{"name": "LOG_LEVEL", "value": "debug"}],
          "image": "example/api:latest",
          "name": "api",
          "resourceRequirements": [{"type": "VCPU", "value": "1"}]
        }
      ],
      "enableExecuteCommand": true,
      "ephemeralStorage": {"sizeInGiB": 40},
      "executionRoleArn": "arn:aws:iam::123456789012:role/execution",
      "ipcMode": "task",
      "networkConfiguration": {"assignPublicIp": "ENABLED"},
      "networkMode": "host",
      "pidMode": "task",
      "platformVersion": "1.4.0",
      "runtimePlatform": {"cpuArchitecture": "ARM64", "operatingSystemFamily": "LINUX"},
      "taskRoleArn": "arn:aws:iam::123456789012:role/task",
      "volumes": [{"name": "data", "host": {"sourcePath": "/data"}}]
    }
  ]
}`

	var input awstypes.EcsProperties
	if err := tfjson.DecodeFromString(want, &input); err != nil {
		t.Fatal(err)
	}

	got, err := flattenECSProperties(&input)
	if err != nil {
		t.Fatal(err)
	}

	assertSerializationJSONEqual(t, want, got)
}

func TestFlattenContainerProperties(t *testing.T) {
	t.Parallel()

	if got, err := flattenContainerProperties(nil); err != nil || got != "" {
		t.Fatalf("flattenContainerProperties(nil) = %q, %v; want empty string, nil", got, err)
	}

	const want = `
{
  "command": ["serve", "--port", "8080"],
  "enableExecuteCommand": true,
  "environment": [{"name": "LOG_LEVEL", "value": "debug"}],
  "ephemeralStorage": {"sizeInGiB": 40},
  "executionRoleArn": "arn:aws:iam::123456789012:role/execution",
  "fargatePlatformConfiguration": {"platformVersion": "1.4.0"},
  "image": "example/api:latest",
  "instanceType": "m5.large",
  "jobRoleArn": "arn:aws:iam::123456789012:role/job",
  "logConfiguration": {"logDriver": "awslogs", "options": {"awslogs-group": "/aws/batch/job"}, "secretOptions": [{"name": "token", "valueFrom": "arn:aws:ssm:region:account:parameter/token"}]},
  "memory": 512,
  "mountPoints": [{"containerPath": "/data", "readOnly": false, "sourceVolume": "data"}],
  "networkConfiguration": {"assignPublicIp": "ENABLED"},
  "privileged": false,
  "readonlyRootFilesystem": true,
  "repositoryCredentials": {"credentialsParameter": "arn:aws:secretsmanager:region:account:secret:registry"},
  "resourceRequirements": [{"type": "MEMORY", "value": "1024"}],
  "runtimePlatform": {"cpuArchitecture": "X86_64", "operatingSystemFamily": "LINUX"},
  "secrets": [{"name": "PASSWORD", "valueFrom": "arn:aws:secretsmanager:region:account:secret:password"}],
  "ulimits": [{"hardLimit": 65535, "name": "nofile", "softLimit": 65535}],
  "user": "1000",
  "vcpus": 1,
  "volumes": [{"name": "data", "host": {"sourcePath": "/data"}}]
}`

	var input awstypes.ContainerProperties
	if err := tfjson.DecodeFromString(want, &input); err != nil {
		t.Fatal(err)
	}

	got, err := flattenContainerProperties(&input)
	if err != nil {
		t.Fatal(err)
	}

	assertSerializationJSONEqual(t, want, got)
}

func TestFlattenNodeProperties(t *testing.T) {
	t.Parallel()

	if got, err := flattenNodeProperties(nil); err != nil || got != "" {
		t.Fatalf("flattenNodeProperties(nil) = %q, %v; want empty string, nil", got, err)
	}

	const want = `
{
  "mainNode": 1,
  "nodeRangeProperties": [
    {
      "container": {
        "command": ["worker"],
        "image": "example/worker:latest",
        "resourceRequirements": [{"type": "VCPU", "value": "2"}]
      },
      "ecsProperties": {
        "taskProperties": [
          {
            "containers": [{"image": "example/sidecar:latest", "name": "sidecar"}],
            "platformVersion": "1.4.0"
          }
        ]
      },
      "instanceTypes": ["m5.large"],
      "targetNodes": "0:3"
    }
  ],
  "numNodes": 4
}`

	var input awstypes.NodeProperties
	if err := tfjson.DecodeFromString(want, &input); err != nil {
		t.Fatal(err)
	}

	got, err := flattenNodeProperties(&input)
	if err != nil {
		t.Fatal(err)
	}

	assertSerializationJSONEqual(t, want, got)
}

func assertSerializationJSONEqual(t *testing.T, want, got string) {
	t.Helper()

	if diff := jsoncmp.Diff(`{"value":`+want+`}`, `{"value":`+got+`}`); diff != "" {
		t.Errorf("unexpected JSON diff (+want, -got):\n%s", diff)
	}
}
