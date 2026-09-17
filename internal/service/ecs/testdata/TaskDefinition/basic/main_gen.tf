# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_task_definition" "test" {
  family = var.rName

  container_definitions = <<TASK_DEFINITION
[
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"environment": [
			{"name": "VARNAME", "value": "VARVAL"}
		],
		"essential": true,
		"image": "jenkins",
		"links": ["mongodb"],
		"memory": 128,
		"name": "jenkins",
		"portMappings": [
			{
				"containerPort": 80,
				"hostPort": 8080
			}
		]
	},
	{
		"cpu": 10,
		"command": ["sleep", "10"],
		"entryPoint": ["/"],
		"essential": true,
		"image": "mongodb",
		"memory": 128,
		"name": "mongodb",
		"portMappings": [
			{
				"containerPort": 28017,
				"hostPort": 28017
			}
		]
	}
]
TASK_DEFINITION

  volume {
    name      = "jenkins-home"
    host_path = "/ecs/jenkins-home"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
