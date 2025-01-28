// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
)

// ConfigCompose can be called to concatenate multiple strings to build test configurations
func ConfigCompose(config ...string) string {
	var str strings.Builder

	for _, conf := range config {
		str.WriteString(conf)
	}

	return str.String()
}

func ConfigAlternateAccountProvider() string {
	//lintignore:AT004
	return ConfigNamedAccountProvider(
		ProviderNameAlternate,
		os.Getenv(envvar.AlternateAccessKeyId),
		os.Getenv(envvar.AlternateProfile),
		os.Getenv(envvar.AlternateSecretAccessKey),
	)
}

func ConfigMultipleAccountProvider(t *testing.T, accounts int) string {
	t.Helper()

	var config strings.Builder

	if accounts > 3 {
		t.Fatalf("invalid number of Account configurations: %d", accounts)
	}

	if accounts >= 2 {
		config.WriteString(
			ConfigNamedAccountProvider(
				ProviderNameAlternate,
				os.Getenv(envvar.AlternateAccessKeyId),
				os.Getenv(envvar.AlternateProfile),
				os.Getenv(envvar.AlternateSecretAccessKey),
			),
		)
	}
	if accounts == 3 {
		config.WriteString(
			ConfigNamedAccountProvider(
				ProviderNameThird,
				os.Getenv(envvar.ThirdAccessKeyId),
				os.Getenv(envvar.ThirdProfile),
				os.Getenv(envvar.ThirdSecretAccessKey),
			),
		)
	}

	return config.String()
}

// ConfigNamedAccountProvider creates a new provider named configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func ConfigNamedAccountProvider(providerName, accessKey, profile, secretKey string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  access_key = %[2]q
  profile    = %[3]q
  secret_key = %[4]q
}
`, providerName, accessKey, profile, secretKey)
}

// Deprecated: Use ConfigMultipleRegionProvider instead
func ConfigAlternateRegionProvider() string {
	return ConfigNamedRegionalProvider(ProviderNameAlternate, AlternateRegion())
}

func ConfigMultipleRegionProvider(regions int) string {
	var config strings.Builder

	config.WriteString(ConfigNamedRegionalProvider(ProviderNameAlternate, AlternateRegion()))

	if regions >= 3 {
		config.WriteString(ConfigNamedRegionalProvider(ProviderNameThird, ThirdRegion()))
	}

	return config.String()
}

// ConfigNamedRegionalProvider creates a new named provider configuration with a region.
//
// This can be used to build multiple provider configuration testing.
func ConfigNamedRegionalProvider(providerName string, region string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  region = %[2]q
}
`, providerName, region)
}

func ConfigDefaultAndIgnoreTagsKeyPrefixes1(key1, value1, keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %[2]q
    }
  }
  ignore_tags {
    key_prefixes = [%[3]q]
  }
}
`, key1, value1, keyPrefix1)
}

func ConfigDefaultAndIgnoreTagsKeys1(key1, value1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %[2]q
    }
  }
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1, value1)
}

func ConfigIgnoreTagsKeyPrefixes1(keyPrefix1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    key_prefixes = [%[1]q]
  }
}
`, keyPrefix1)
}

func ConfigIgnoreTagsKeys(key1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  ignore_tags {
    keys = [%[1]q]
  }
}
`, key1)
}

func ConfigWithEchoProvider(ephemeralResourceData string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "echo" {
  data = %[1]s
}
resource "echo" "test" {}
`, ephemeralResourceData)
}

// ConfigRegionalProvider creates a new provider configuration with a region.
//
// This can only be used for single provider configuration testing as it
// overwrites the "aws" provider configuration.
func ConfigRegionalProvider(region string) string {
	return ConfigNamedRegionalProvider(ProviderName, region)
}

func ConfigAlternateAccountAlternateRegionProvider() string {
	return ConfigNamedAlternateAccountAlternateRegionProvider(ProviderNameAlternate)
}

func ConfigNamedAlternateAccountAlternateRegionProvider(providerName string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider %[1]q {
  access_key = %[2]q
  profile    = %[3]q
  region     = %[4]q
  secret_key = %[5]q
}
`, providerName, os.Getenv(envvar.AlternateAccessKeyId), os.Getenv(envvar.AlternateProfile), AlternateRegion(), os.Getenv(envvar.AlternateSecretAccessKey))
}

func ConfigDefaultTags_Tags0() string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		`
provider "aws" {
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`)
}

func ConfigDefaultTags_Tags1(tag1, value1 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %[2]q
    }
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, value1))
}

func ConfigDefaultTags_Tags2(tag1, value1, tag2, value2 string) string {
	//lintignore:AT004
	return ConfigCompose(
		testAccProviderConfigBase,
		fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]q = %[2]q
      %[3]q = %[4]q
    }
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
`, tag1, value1, tag2, value2))
}

func ConfigAssumeRolePolicy(policy string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %[1]q
    policy   = %[2]q
  }
}
`, os.Getenv(envvar.AccAssumeRoleARN), policy)
}

const testAccProviderConfigBase = `
data "aws_region" "provider_test" {}

# Required to initialize the provider.
data "aws_service" "provider_test" {
  region     = data.aws_region.provider_test.name
  service_id = "s3"
}
`

func ConfigAvailableAZsNoOptIn() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`
}

func ConfigAvailableAZsNoOptInDefaultExclude() string {
	// Exclude usw2-az4 (us-west-2d) as it has limited instance types.
	return ConfigAvailableAZsNoOptInExclude("usw2-az4", "usgw1-az2")
}

func ConfigAvailableAZsNoOptInExclude(excludeZoneIds ...string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  exclude_zone_ids = ["%[1]s"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, strings.Join(excludeZoneIds, "\", \""))
}

// AvailableEC2InstanceTypeForAvailabilityZone returns the configuration for a data source that describes
// the first available EC2 instance type offering in the specified availability zone from a list of preferred instance types.
// The first argument is either an Availability Zone name or Terraform configuration reference to one, e.g.
//   - data.aws_availability_zones.available.names[0]
//   - aws_subnet.test.availability_zone
//   - us-west-2a
//
// The data source is named 'available'.
func AvailableEC2InstanceTypeForAvailabilityZone(availabilityZoneName string, preferredInstanceTypes ...string) string {
	if !strings.Contains(availabilityZoneName, ".") {
		availabilityZoneName = strconv.Quote(availabilityZoneName)
	}

	return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["%[2]s"]
  }

  filter {
    name   = "location"
    values = [%[1]s]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["%[2]s"]
}
`, availabilityZoneName, strings.Join(preferredInstanceTypes, "\", \""))
}

// AvailableEC2InstanceTypeForRegion returns the configuration for a data source that describes
// the first available EC2 instance type offering in the current region from a list of preferred instance types.
// The data source is named 'available'.
func AvailableEC2InstanceTypeForRegion(preferredInstanceTypes ...string) string {
	return AvailableEC2InstanceTypeForRegionNamed("available", preferredInstanceTypes...)
}

// AvailableEC2InstanceTypeForRegionNamed returns the configuration for a data source that describes
// the first available EC2 instance type offering in the current region from a list of preferred instance types.
// The data source name is configurable.
func AvailableEC2InstanceTypeForRegionNamed(name string, preferredInstanceTypes ...string) string {
	return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "%[1]s" {
  filter {
    name   = "instance-type"
    values = ["%[2]s"]
  }

  preferred_instance_types = ["%[2]s"]
}
`, name, strings.Join(preferredInstanceTypes, "\", \""))
}

func configLatestAmazonLinux2HVMEBSAMI(architecture ec2types.ArchitectureValues) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn2-ami-minimal-hvm-ebs-%[1]s" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = [%[1]q]
  }
}
`, architecture)
}

// ConfigLatestAmazonLinux2HVMEBSX8664AMI returns the configuration for a data source that
// describes the latest Amazon Linux 2 x86_64 AMI using HVM virtualization and an EBS root device.
// The data source is named 'amzn2-ami-minimal-hvm-ebs-x86_64'.
func ConfigLatestAmazonLinux2HVMEBSX8664AMI() string {
	return configLatestAmazonLinux2HVMEBSAMI(ec2types.ArchitectureValuesX8664)
}

// ConfigLatestAmazonLinux2HVMEBSARM64AMI returns the configuration for a data source that
// describes the latest Amazon Linux 2 arm64 AMI using HVM virtualization and an EBS root device.
// The data source is named 'amzn2-ami-minimal-hvm-ebs-arm64'.
func ConfigLatestAmazonLinux2HVMEBSARM64AMI() string {
	return configLatestAmazonLinux2HVMEBSAMI(ec2types.ArchitectureValuesArm64)
}

func ConfigLambdaBase(policyName, roleName, sgName string) string {
	return ConfigCompose(ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:AssignPrivateIpAddresses",
        "ec2:UnassignPrivateIpAddresses"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = %[2]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_vpc" "vpc_for_lambda" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[3]q
  }
}

resource "aws_subnet" "subnet_for_lambda" {
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  cidr_block                      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 1)
  availability_zone               = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

# This is defined here, rather than only in test cases where it's needed is to
# prevent a timeout issue when fully removing Lambda Filesystems
resource "aws_subnet" "subnet_for_lambda_az2" {
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  cidr_block                      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 2)
  availability_zone               = data.aws_availability_zones.available.names[1]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 2)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

resource "aws_security_group" "sg_for_lambda" {
  name        = %[3]q
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = aws_vpc.vpc_for_lambda.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[3]q
  }
}
`, policyName, roleName, sgName))
}

func ConfigVPCWithSubnets(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigVPCWithSubnetsEnableDNSHostnames(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigVPCWithSubnetsIPv6(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigBedrockAgentKnowledgeBaseRDSBase(rName, model string) string {
	return ConfigCompose(
		ConfigVPCWithSubnetsEnableDNSHostnames(rName, 2), //nolint:mnd // 2 subnets required
		fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": "sts:AssumeRole",
		"Principal": {
		"Service": "bedrock.amazonaws.com"
		},
		"Effect": "Allow"
	}]
}
POLICY
}

# See https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html.
resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:ListFoundationModels",
        "bedrock:ListCustomModels"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "rds_data_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/AmazonRDSDataFullAccess"
}

resource "aws_iam_role_policy_attachment" "secrets_manager_read_write" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/SecretsManagerReadWrite"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier          = %[1]q
  engine                      = "aurora-postgresql"
  engine_mode                 = "provisioned"
  engine_version              = "15.4"
  database_name               = "test"
  master_username             = "test"
  manage_master_user_password = true
  enable_http_endpoint        = true
  vpc_security_group_ids      = [aws_security_group.test.id]
  skip_final_snapshot         = true
  db_subnet_group_name        = aws_db_subnet_group.test.name

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier  = aws_rds_cluster.test.id
  instance_class      = "db.serverless"
  engine              = aws_rds_cluster.test.engine
  engine_version      = aws_rds_cluster.test.engine_version
  publicly_accessible = true
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_rds_cluster.test.master_user_secret[0].secret_arn
  version_stage = "AWSCURRENT"
  depends_on    = [aws_rds_cluster.test]
}

resource "null_resource" "db_setup" {
  depends_on = [aws_rds_cluster_instance.test, aws_rds_cluster.test, data.aws_secretsmanager_secret_version.test]

  provisioner "local-exec" {
    command = <<EOT
      sleep 60
      export PGPASSWORD=$(aws secretsmanager get-secret-value --secret-id '${aws_rds_cluster.test.master_user_secret[0].secret_arn}' --version-stage AWSCURRENT --region ${data.aws_region.current.name} --query SecretString --output text | jq -r '."password"')
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE EXTENSION IF NOT EXISTS vector;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_integration;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_new;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE ROLE bedrock_user WITH PASSWORD '$PGPASSWORD' LOGIN;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "GRANT ALL ON SCHEMA bedrock_integration TO bedrock_user;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE TABLE bedrock_integration.bedrock_kb (id uuid PRIMARY KEY, embedding vector(1536), chunks text, metadata json);"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE INDEX ON bedrock_integration.bedrock_kb USING hnsw (embedding vector_cosine_ops);"
    EOT
  }
}
`, rName, model))
}

func ConfigBedrockAgentKnowledgeBaseRDSUpdateBase(rName, model string) string {
	return ConfigCompose(
		ConfigBedrockAgentKnowledgeBaseRDSBase(rName, model), //nolint:mnd
		fmt.Sprintf(`
resource "aws_iam_role" "test_update" {
  name               = "%[1]s-update"
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": "sts:AssumeRole",
		"Principal": {
		"Service": "bedrock.amazonaws.com"
		},
		"Effect": "Allow"
	}]
}
POLICY
}

# See https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html.
resource "aws_iam_role_policy" "test_update" {
  name   = "%[1]s-update"
  role   = aws_iam_role.test_update.name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:ListFoundationModels",
        "bedrock:ListCustomModels"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/%[2]s"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "rds_data_full_access_update" {
  role       = aws_iam_role.test_update.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/AmazonRDSDataFullAccess"
}

resource "aws_iam_role_policy_attachment" "secrets_manager_read_write_update" {
  role       = aws_iam_role.test_update.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/SecretsManagerReadWrite"
}
`, rName, model))
}
