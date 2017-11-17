package aws

import (
	"testing"
)

func TestAwsEcsContainerDefinitionsAreEquivalent_basic(t *testing.T) {
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

	equal, err := ecsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_portMappings(t *testing.T) {
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

	equal, err := ecsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Expected definitions to be equal.")
	}
}

func TestAwsEcsContainerDefinitionsAreEquivalent_negative(t *testing.T) {
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

	equal, err := ecsContainerDefinitionsAreEquivalent(cfgRepresention, apiRepresentation)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Expected definitions to differ.")
	}
}
