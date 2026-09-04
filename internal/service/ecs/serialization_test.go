// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlattenContainerDefinitions(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"empty": `[]`,
		"populated": `
[
  {
    "command": ["serve", "--port", "8080"],
    "cpu": 256,
    "dependsOn": [{"condition": "HEALTHY", "containerName": "database"}],
    "disableNetworking": false,
    "dnsSearchDomains": ["example.test"],
    "dnsServers": ["10.0.0.2"],
    "dockerLabels": {"component": "api"},
    "dockerSecurityOptions": ["no-new-privileges"],
    "entryPoint": ["/entrypoint.sh"],
    "environment": [{"name": "LOG_LEVEL", "value": "debug"}],
    "environmentFiles": [{"type": "s3", "value": "arn:aws:s3:::bucket/app.env"}],
    "essential": true,
    "extraHosts": [{"hostname": "database", "ipAddress": "10.0.0.10"}],
    "firelensConfiguration": {"type": "fluentbit", "options": {"enable-ecs-log-metadata": "true"}},
    "healthCheck": {"command": ["CMD-SHELL", "curl -f http://localhost/"], "interval": 30, "retries": 3, "startPeriod": 5, "timeout": 10},
    "hostname": "api",
    "image": "example/api:latest",
    "interactive": true,
    "links": ["database"],
    "logConfiguration": {"logDriver": "awslogs", "options": {"awslogs-group": "/ecs/api"}, "secretOptions": [{"name": "token", "valueFrom": "arn:aws:ssm:region:account:parameter/token"}]},
    "memory": 512,
    "memoryReservation": 256,
    "mountPoints": [{"containerPath": "/data", "readOnly": false, "sourceVolume": "data"}],
    "name": "api",
    "portMappings": [{"appProtocol": "http", "containerPort": 8080, "containerPortRange": "8080-8081", "hostPort": 8080, "name": "http", "protocol": "tcp"}],
    "privileged": false,
    "pseudoTerminal": true,
    "readonlyRootFilesystem": true,
    "repositoryCredentials": {"credentialsParameter": "arn:aws:secretsmanager:region:account:secret:registry"},
    "resourceRequirements": [{"type": "GPU", "value": "1"}],
    "secrets": [{"name": "PASSWORD", "valueFrom": "arn:aws:secretsmanager:region:account:secret:password"}],
    "startTimeout": 20,
    "stopTimeout": 30,
    "systemControls": [{"namespace": "net.core.somaxconn", "value": "1024"}],
    "ulimits": [{"hardLimit": 65535, "name": "nofile", "softLimit": 65535}],
    "user": "1000",
    "versionConsistency": "enabled",
    "volumesFrom": [{"readOnly": true, "sourceContainer": "database"}],
    "workingDirectory": "/app"
  }
]`,
	}

	for name, want := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input []awstypes.ContainerDefinition
			if err := tfjson.DecodeFromString(want, &input); err != nil {
				t.Fatal(err)
			}

			got, err := flattenContainerDefinitions(input)
			if err != nil {
				t.Fatal(err)
			}

			assertContainerDefinitionsJSONEqual(t, want, got)
		})
	}
}

func assertContainerDefinitionsJSONEqual(t *testing.T, want, got string) {
	t.Helper()

	if diff := jsoncmp.Diff(`{"value":`+want+`}`, `{"value":`+got+`}`); diff != "" {
		t.Errorf("unexpected JSON diff (+want, -got):\n%s", diff)
	}
}
