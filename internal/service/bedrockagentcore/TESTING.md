# Testing BedRock Agent Core Resource Types

The resource type `aws_bedrockagentcore_agent_runtime` requires code delivered in either
a zipfile containing code and dependencies stored in S3 (`agent_runtime_artifact.code_configuration`) or
a container stored in ECR (`agent_runtime_artifact.container_configuration`).

Instructions for creating either type can be found at https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/runtime-get-started-cli.html.

## S3 Zipfile

The current S3 zipfile, located at `./test-fixtures/agent-runtime-codezip.zip`, was copied from a manually-created Agent Runtime.
If it needs to be recreated, follow the instructions above.

## ECR Container

The provider tests do not supply a container.
To test with containers, create them using the instructions above and upload them to ECR.
The tests `TestAccBedrockAgentCoreAgentRuntime_artifactContainer` and `TestAccBedrockAgentCoreAgentRuntime_artifactTypeChanged` both use the environment variable `AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI` to specify a container.
The test `TestAccBedrockAgentCoreAgentRuntime_artifactContainer` also uses the environment variable `AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V2_URI` to test changing the container configuration.
