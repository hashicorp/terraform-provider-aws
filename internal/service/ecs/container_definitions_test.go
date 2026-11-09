// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"encoding/json"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlattenContainerDefinitions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil":              {input: "null", want: "[]"},
		"empty":            {input: "[]"},
		"empty definition": {input: "[{}]"},
		"all fields": {
			input: `[{
				"command":["sh","-c","echo \"hello\" <&>\n\u2603"],"cpu":256,"credentialSpecs":["credential-spec"],
				"dependsOn":[{"condition":"START","containerName":"other"}],"disableNetworking":true,
				"dnsSearchDomains":["example.com"],"dnsServers":["10.0.0.2"],
				"dockerLabels":{"MixedCase.Key":"<&>","empty":""},"dockerSecurityOptions":["no-new-privileges"],
				"entryPoint":["sh"],"environment":[{"name":"Z","value":"last"},{"name":"A","value":""}],
				"environmentFiles":[{"type":"s3","value":"environment-file"}],"essential":true,
				"extraHosts":[{"hostname":"host","ipAddress":"10.0.0.3"}],
				"firelensConfiguration":{"options":{"enable-ecs-log-metadata":"true"},"type":"fluentbit"},
				"healthCheck":{"command":["CMD-SHELL","exit 0"],"interval":30,"retries":3,"startPeriod":10,"timeout":5},
				"hostname":"hostname","image":"image","interactive":true,"links":["other"],
				"linuxParameters":{
					"capabilities":{"add":["SYS_PTRACE"],"drop":["NET_RAW"]},
					"devices":[{"containerPath":"/dev/x","hostPath":"/dev/y","permissions":["read","write","mknod"]}],
					"initProcessEnabled":true,"maxSwap":1024,"sharedMemorySize":64,"swappiness":60,
					"tmpfs":[{"containerPath":"/tmp","mountOptions":["rw","noexec"],"size":128}]
				},
				"logConfiguration":{"logDriver":"awslogs","options":{"awslogs-group":"logs"},"secretOptions":[{"name":"log-secret","valueFrom":"log-value"}]},
				"memory":512,"memoryReservation":256,"mountPoints":[{"containerPath":"/data","readOnly":true,"sourceVolume":"data"}],
				"name":"Z","portMappings":[{"appProtocol":"http","containerPort":80,"containerPortRange":"8000-8010","hostPort":8080,"name":"http","protocol":"tcp"}],
				"privileged":true,"pseudoTerminal":true,"readonlyRootFilesystem":true,
				"repositoryCredentials":{"credentialsParameter":"credentials"},"resourceRequirements":[{"type":"GPU","value":"1"}],
				"restartPolicy":{"enabled":true,"ignoredExitCodes":[0,1],"restartAttemptPeriod":60},
				"secrets":[{"name":"Z","valueFrom":"last"},{"name":"A","valueFrom":"first"}],
				"startTimeout":10,"stopTimeout":20,"systemControls":[{"namespace":"net.ipv4.ip_forward","value":"1"}],
				"ulimits":[{"hardLimit":1024,"name":"nofile","softLimit":512}],"user":"1000","versionConsistency":"enabled",
				"volumesFrom":[{"readOnly":true,"sourceContainer":"other"}],"workingDirectory":"/app"
			},{"name":"A","image":"other-image"}]`,
		},
		"explicit zero values": {
			input: `[{
				"command":[""],"credentialSpecs":[""],"dependsOn":[{"containerName":""}],"disableNetworking":false,
				"dnsSearchDomains":[""],"dnsServers":[""],"dockerLabels":{"":""},"dockerSecurityOptions":[""],
				"entryPoint":[""],"environment":[{"name":"","value":""}],"environmentFiles":[{"value":""}],"essential":false,
				"extraHosts":[{"hostname":"","ipAddress":""}],"firelensConfiguration":{"options":{"":""}},
				"healthCheck":{"command":[""],"interval":0,"retries":0,"startPeriod":0,"timeout":0},
				"hostname":"","image":"","interactive":false,"links":[""],
				"linuxParameters":{"capabilities":{"add":[""],"drop":[""]},"devices":[{"containerPath":"","hostPath":"","permissions":[""]}],
					"initProcessEnabled":false,"maxSwap":0,"sharedMemorySize":0,"swappiness":0,"tmpfs":[{"containerPath":"","mountOptions":[""],"size":0}]},
				"logConfiguration":{"options":{"":""},"secretOptions":[{"name":"","valueFrom":""}]},
				"memory":0,"memoryReservation":0,"mountPoints":[{"containerPath":"","readOnly":false,"sourceVolume":""}],"name":"",
				"portMappings":[{"containerPort":0,"containerPortRange":"","hostPort":0,"name":""}],
				"privileged":false,"pseudoTerminal":false,"readonlyRootFilesystem":false,
				"repositoryCredentials":{"credentialsParameter":""},"resourceRequirements":[{"value":""}],
				"restartPolicy":{"enabled":false,"ignoredExitCodes":[0],"restartAttemptPeriod":0},"secrets":[{"name":"","valueFrom":""}],
				"startTimeout":0,"stopTimeout":0,"systemControls":[{"namespace":"","value":""}],
				"ulimits":[{"hardLimit":0,"softLimit":0}],"user":"","volumesFrom":[{"readOnly":false,"sourceContainer":""}],"workingDirectory":""
			}]`,
		},
		"empty collections": {
			input: `[{
				"command":[],"credentialSpecs":[],"dependsOn":[],"dnsSearchDomains":[],"dnsServers":[],"dockerLabels":{},"dockerSecurityOptions":[],
				"entryPoint":[],"environment":[],"environmentFiles":[],"extraHosts":[],"firelensConfiguration":{"options":{}},"healthCheck":{"command":[]},
				"links":[],"linuxParameters":{"capabilities":{"add":[],"drop":[]},"devices":[],"tmpfs":[]},
				"logConfiguration":{"options":{},"secretOptions":[]},"mountPoints":[],"portMappings":[],"repositoryCredentials":{},
				"resourceRequirements":[],"restartPolicy":{"ignoredExitCodes":[]},"secrets":[],"systemControls":[],"ulimits":[],"volumesFrom":[]
			}]`,
		},
		"empty nested objects": {
			input: `[{
				"dependsOn":[{}],"environment":[{}],"environmentFiles":[{}],"extraHosts":[{}],"firelensConfiguration":{},"healthCheck":{},
				"linuxParameters":{"capabilities":{},"devices":[{},{"permissions":[]}]},"logConfiguration":{"secretOptions":[{}]},
				"mountPoints":[{}],"portMappings":[{}],"repositoryCredentials":{},"resourceRequirements":[{}],"restartPolicy":{},
				"secrets":[{}],"systemControls":[{}],"volumesFrom":[{}]
			}]`,
		},
		"required zero fields retained": {
			input: `[{"linuxParameters":{"tmpfs":[{},{"mountOptions":[]}]},"ulimits":[{}]}]`,
			want:  `[{"linuxParameters":{"tmpfs":[{"size":0},{"mountOptions":[],"size":0}]},"ulimits":[{"hardLimit":0,"softLimit":0}]}]`,
		},
		"zero cpu and empty enums omitted": {
			input: `[{"cpu":0,"versionConsistency":"","dependsOn":[{"condition":""}],"environmentFiles":[{"type":""}],"firelensConfiguration":{"type":""},"logConfiguration":{"logDriver":""},"portMappings":[{"appProtocol":"","protocol":""}],"resourceRequirements":[{"type":""}],"ulimits":[{"name":""}]}]`,
			want:  `[{"dependsOn":[{}],"environmentFiles":[{}],"firelensConfiguration":{},"logConfiguration":{},"portMappings":[{}],"resourceRequirements":[{}],"ulimits":[{"hardLimit":0,"softLimit":0}]}]`,
		},
		"unknown enums retained": {
			input: `[{"versionConsistency":"FUTURE","dependsOn":[{"condition":"FUTURE"}],"environmentFiles":[{"type":"FUTURE"}],"firelensConfiguration":{"type":"FUTURE"},"logConfiguration":{"logDriver":"FUTURE"},"portMappings":[{"appProtocol":"FUTURE","protocol":"FUTURE"}],"resourceRequirements":[{"type":"FUTURE"}],"ulimits":[{"name":"FUTURE","hardLimit":0,"softLimit":0}]}]`,
		},
		"nil fields omitted": {
			input: `[{"command":null,"environment":null,"essential":null,"dockerLabels":null,"linuxParameters":null,"healthCheck":null,"restartPolicy":null}]`,
			want:  `[{}]`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input []awstypes.ContainerDefinition
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			got, err := flattenContainerDefinitions(input)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.want
			if want == "" {
				want = tc.input
			}
			if !json.Valid([]byte(got)) {
				t.Fatalf("invalid JSON: %s", got)
			}
			// jsoncmp compares object roots, so wrap the container arrays.
			if diff := jsoncmp.Diff(`{"definitions":`+want+`}`, `{"definitions":`+got+`}`); diff != "" {
				t.Errorf("unexpected JSON (+got, -want): %s", diff)
			}
		})
	}
}

func TestContainerDefinitionsAreEquivalent_basic(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "links": [
        "mysql"
      ],
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10
    },
    {
      "environment": [
        {
          "name": "MYSQL_ROOT_PASSWORD",
          "value": "password"
        }
      ],
      "name": "mysql",
      "image": "mysql",
      "cpu": 10,
      "memory": 500,
      "essential": true
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "links": [
            "mysql"
        ],
        "portMappings": [
            {
                "containerPort": 80,
                "hostPort": 80,
                "protocol": "tcp"
            }
        ],
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    },
    {
        "name": "mysql",
        "image": "mysql",
        "cpu": 10,
        "memory": 500,
        "portMappings": [],
        "essential": true,
        "environment": [
            {
                "name": "MYSQL_ROOT_PASSWORD",
                "value": "password"
            }
        ],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_portMappings(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "portMappings": [
            {
                "containerPort": 80,
                "hostPort": 0,
                "protocol": "tcp"
            }
        ],
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_portMappingsIgnoreHostPort(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80
        }
      ]
    }
]`

	apiRepresentation := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "portMappings": [
        {
          "containerPort": 80
        }
      ]
    }
]`

	var (
		equal bool
		err   error
	)

	equal, err = containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Expected definitions to differ.")
	}

	equal, err = containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, true)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_arrays(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "links": ["container1", "container2", "container3"],
      "portMappings": [
        {"containerPort": 80},
        {"containerPort": 81},
        {"containerPort": 82}
      ],
      "environment": [
        {"name": "VARNAME1", "value": "VARVAL1"},
        {"name": "VARNAME2", "value": "VARVAL2"},
        {"name": "VARNAME3", "value": "VARVAL3"}
      ],
      "extraHosts": [
        {"hostname": "host1", "ipAddress": "127.0.0.1"},
        {"hostname": "host2", "ipAddress": "127.0.0.2"},
        {"hostname": "host3", "ipAddress": "127.0.0.3"}
      ],
      "mountPoints": [
        {"sourceVolume": "vol1", "containerPath": "/vol1"},
        {"sourceVolume": "vol2", "containerPath": "/vol2"},
        {"sourceVolume": "vol3", "containerPath": "/vol3"}
      ],
      "volumesFrom": [
        {"sourceContainer": "container1"},
        {"sourceContainer": "container2"},
        {"sourceContainer": "container3"}
      ],
      "ulimits": [
        {
          "name": "core",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "cpu",
          "softLimit": 10, "hardLimit": 20
        },
        {
          "name": "fsize",
          "softLimit": 10, "hardLimit": 20
        }
      ],
      "linuxParameters": {
        "capabilities": {
          "add": ["AUDIT_CONTROL", "AUDIT_WRITE", "BLOCK_SUSPEND"],
          "drop": ["CHOWN", "IPC_LOCK", "KILL"]
        }
      },
      "devices": [
        {
          "hostPath": "/path1",
          "permissions": ["read", "write", "mknod"]
        },
        {
          "hostPath": "/path2",
          "permissions": ["read", "write"]
        },
        {
          "hostPath": "/path3",
          "permissions": ["read", "mknod"]
        }
      ],
      "dockerSecurityOptions": ["label:one", "label:two", "label:three"],
      "memory": 500,
      "cpu": 10
    },
    {
      "name": "container1",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container2",
      "image": "busybox",
      "memory": 100
    },
    {
      "name": "container3",
      "image": "busybox",
      "memory": 100
    }
]`

	apiRepresentation := `
[
  {
    "cpu": 10,
    "dockerSecurityOptions": [
      "label:one",
      "label:two",
      "label:three"
    ],
    "environment": [
      {
        "name": "VARNAME3",
        "value": "VARVAL3"
      },
      {
        "name": "VARNAME2",
        "value": "VARVAL2"
      },
      {
        "name": "VARNAME1",
        "value": "VARVAL1"
      }
    ],
    "essential": true,
    "extraHosts": [
      {
        "hostname": "host1",
        "ipAddress": "127.0.0.1"
      },
      {
        "hostname": "host2",
        "ipAddress": "127.0.0.2"
      },
      {
        "hostname": "host3",
        "ipAddress": "127.0.0.3"
      }
    ],
    "image": "wordpress",
    "links": [
      "container1",
      "container2",
      "container3"
    ],
    "linuxParameters": {
      "capabilities": {
        "add": [
          "AUDIT_CONTROL",
          "AUDIT_WRITE",
          "BLOCK_SUSPEND"
        ],
        "drop": [
          "CHOWN",
          "IPC_LOCK",
          "KILL"
        ]
      }
    },
    "memory": 500,
    "mountPoints": [
      {
        "containerPath": "/vol1",
        "sourceVolume": "vol1"
      },
      {
        "containerPath": "/vol2",
        "sourceVolume": "vol2"
      },
      {
        "containerPath": "/vol3",
        "sourceVolume": "vol3"
      }
    ],
    "name": "wordpress",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 0,
        "protocol": "tcp"
      },
      {
        "containerPort": 81,
        "hostPort": 0,
        "protocol": "tcp"
      },
      {
        "containerPort": 82,
        "hostPort": 0,
        "protocol": "tcp"
      }
    ],
    "ulimits": [
      {
        "hardLimit": 20,
        "name": "core",
        "softLimit": 10
      },
      {
        "hardLimit": 20,
        "name": "cpu",
        "softLimit": 10
      },
      {
        "hardLimit": 20,
        "name": "fsize",
        "softLimit": 10
      }
    ],
    "volumesFrom": [
      {
        "sourceContainer": "container1"
      },
      {
        "sourceContainer": "container2"
      },
      {
        "sourceContainer": "container3"
      }
    ]
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container1",
    "portMappings": [],
    "volumesFrom": []
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container2",
    "portMappings": [],
    "volumesFrom": []
  },
  {
    "cpu": 0,
    "environment": [],
    "essential": true,
    "image": "busybox",
    "memory": 100,
    "mountPoints": [],
    "name": "container3",
    "portMappings": [],
    "volumesFrom": []
  }
]
`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_negative(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "memory": 500,
      "cpu": 10,
      "environment": [
        {"name": "EXAMPLE_NAME", "value": "foobar"}
      ]
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Expected definitions to differ.")
	}
}

func TestContainerDefinitionsAreEquivalent_missingEnvironmentName(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "links": [
        "mysql"
      ],
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10
    },
    {
      "environment": [
        {
          "value": "password"
        },
        {
          "value": "password2"
        }
      ],
      "name": "mysql",
      "image": "mysql",
      "cpu": 10,
      "memory": 500,
      "essential": true
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "links": [
            "mysql"
        ],
        "portMappings": [
            {
                "containerPort": 80,
                "hostPort": 80,
                "protocol": "tcp"
            }
        ],
        "essential": true,
        "environment": [],
        "mountPoints": [],
        "volumesFrom": []
    },
    {
        "name": "mysql",
        "image": "mysql",
        "cpu": 10,
        "memory": 500,
        "portMappings": [],
        "essential": true,
        "environment": [
          {
            "value": "password"
          },
          {
            "value": "password2"
          }
        ],
        "mountPoints": [],
        "volumesFrom": []
    }
]`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_sparseArrays(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "links": [
        "mysql"
      ],
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {}
      ],
      "memory": 500,
      "cpu": 10,
      "environment": [null],
      "mountPoints": [{"containerPath": null}],
      "command": [""]
    }
]`

	apiRepresentation := `
[
    {
        "name": "wordpress",
        "image": "wordpress",
        "cpu": 10,
        "memory": 500,
        "links": [
            "mysql"
        ],
        "portMappings": [],
        "essential": true,
        "command": [""]
    }
]`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestContainerDefinitionsAreEquivalent_healthCheck(t *testing.T) {
	t.Parallel()

	cfgRepresentation := `
[
    {
        "cpu": 512,
        "environment": [],
        "healthCheck": {
            "command": [
                "CMD-SHELL",
                "curl -f http://localhost:8080/health || exit 1"
            ]
        },
        "image": "nginx",
        "memory": 2048,
        "name": "nginx",
        "startTimeout": 10,
        "logConfiguration": {
            "logDriver": "awslogs",
            "options": {
                "awslogs-group": "foo-bar-e196c99",
                "awslogs-region": "region-1",
                "awslogs-stream-prefix": "nginx"
            }
        }
    }
]`

	apiRepresentation := `
[
    {
        "cpu": 512,
        "environment": [],
        "essential": true,
        "healthCheck": {
            "command": [
                "CMD-SHELL",
                "curl -f http://localhost:8080/health || exit 1"
            ],
            "interval": 30,
            "retries": 3,
            "timeout": 5
        },
        "image": "nginx",
        "logConfiguration": {
            "logDriver": "awslogs",
            "options": {
                "awslogs-group": "foo-bar-e196c99",
                "awslogs-region": "region-1",
                "awslogs-stream-prefix": "nginx"
            }
        },
        "memory": 2048,
        "mountPoints": [],
        "name": "nginx",
        "portMappings": [],
        "startTimeout": 10,
        "systemControls": [],
        "volumesFrom": []
    }
]
`

	equal, err := containerDefinitionsAreEquivalent(cfgRepresentation, apiRepresentation, false)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestExpandContainerDefinitions_InvalidVersionConsistency(t *testing.T) {
	t.Parallel()

	cfgRepresention := `
[
    {
      "name": "wordpress",
      "image": "wordpress",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80
        }
      ],
      "memory": 500,
      "cpu": 10,
      "versionConsistency": "invalid"
    }
]`
	_, err := expandContainerDefinitions(cfgRepresention)
	if err == nil {
		t.Fatal("Expected error")
	}

	expectedErr := "invalid version consistency value (invalid) for container definition supplied at index (0)"
	if err.Error() != expectedErr {
		t.Fatalf("Expected message '%[1]s', got '%[2]s'", expectedErr, err.Error())
	}
}
