// Copyright IBM Corp. 2014, 2026
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

// Only include configuration that is used in 3+ or more services.
// Go idiom: "A little copying is better than a little dependency."

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
	if accounts == 4 {
		config.WriteString(
			ConfigNamedAccountProvider(
				ProviderNameFourth,
				os.Getenv(envvar.FourthAccessKeyId),
				os.Getenv(envvar.FourthProfile),
				os.Getenv(envvar.FourthSecretAccessKey),
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

	if regions == 3 {
		config.WriteString(ConfigNamedRegionalProvider(ProviderNameThird, ThirdRegion()))
	}

	if regions == 4 {
		config.WriteString(ConfigNamedRegionalProvider(ProviderNameFourth, FourthRegion()))
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

// ConfigTagPolicyCompliance enables tag policy enforcement with the provided severity
func ConfigTagPolicyCompliance(severity string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  tag_policy_compliance = %[1]q
}
`, severity)
}

// ConfigTagPolicyComplianceAndDefaultTags1 enables tag policy enforcement with the
// provided severity and a default tag
func ConfigTagPolicyComplianceAndDefaultTags1(severity, key1, value1 string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  tag_policy_compliance = %[1]q

  default_tags {
    tags = {
      %[2]s = %[3]q
    }
  }
}
`, severity, key1, value1)
}

func ConfigSkipCredentialsValidationAndRequestingAccountID() string {
	//lintignore:AT004
	return `
provider "aws" {
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}
`
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

func ConfigAssumeRole() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn = %[1]q
  }
}
`, os.Getenv(envvar.AccAssumeRoleARN))
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

// ConfigProviderMeta returns a terraform block with provider_meta configured
func ConfigProviderMeta() string {
	return `
terraform {
  provider_meta "aws" {
    user_agent = [
      "test-module/0.0.1 (test comment)",
      "github.com/hashicorp/terraform-provider-aws/v0.0.0-acctest",
    ]
  }
}
`
}

const testAccProviderConfigBase = `
data "aws_region" "provider_test" {}

# Required to initialize the provider.
data "aws_service" "provider_test" {
  region     = data.aws_region.provider_test.region
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

func ConfigAvailableAZsNoOptInDefaultExclude_RegionOverride(region string) string {
	// Exclude usw2-az4 (us-west-2d) as it has limited instance types.
	return ConfigAvailableAZsNoOptInExclude_RegionOverride(region, "usw2-az4", "usgw1-az2")
}

func ConfigAvailableAZsNoOptInExclude_RegionOverride(region string, excludeZoneIds ...string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  region = %[2]q

  exclude_zone_ids = ["%[1]s"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, strings.Join(excludeZoneIds, "\", \""), region)
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
  availability_zone               = data.aws_availability_zones.available.names[1]

  cidr_block      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 1)
  ipv6_cidr_block = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 1)

  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

# This is defined here, rather than only in test cases where it's needed is to
# prevent a timeout issue when fully removing Lambda Filesystems
resource "aws_subnet" "subnet_for_lambda_az2" {
  vpc_id                          = aws_vpc.vpc_for_lambda.id
  availability_zone               = data.aws_availability_zones.available.names[1]

  cidr_block      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 2)
  ipv6_cidr_block = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 2)

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
		ConfigSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName),
	)
}

func ConfigVPCWithSubnets_RegionOverride(rName string, subnetCount int, region string) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude_RegionOverride(region),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  region = %[3]q

  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = %[2]d

  region = %[3]q

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}
`, rName, subnetCount, region),
	)
}

func ConfigVPCWithSubnetsEnableDNSHostnames(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}
`, rName),
	)
}

func ConfigVPCWithSubnetsIPv6(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigSubnetsIPv6(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}
`, rName),
	)
}

func ConfigSubnets(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
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

func ConfigSubnetsIPv6(rName string, subnetCount int) string {
	return ConfigCompose(
		ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]

  cidr_block      = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)

  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role_bedrock" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "bedrock" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:InvokeModel"]
    resources = ["*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket", "s3:GetObject"]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3vectors:GetIndex",
      "s3vectors:QueryVectors",
      "s3vectors:PutVectors",
      "s3vectors:GetVectors",
      "s3vectors:DeleteVectors"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_bedrock.json
  name               = %[1]q
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.bedrock.json
}

resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = %[1]q
  force_destroy      = true
}

resource "aws_s3vectors_index" "test" {
  index_name         = %[1]q
  vector_bucket_name = aws_s3vectors_vector_bucket.test.vector_bucket_name

  data_type       = "float32"
  dimension       = 256
  distance_metric = "euclidean"
}
`, rName)
}

// ConfigRandomPassword returns the configuration for an ephemeral resource that
// describes a random password.
//
// The ephemeral resource is named 'test'. Use
// ephemeral.aws_secretsmanager_random_password.test.random_password to
// reference the password value, assigning it to a write-only argument ("_wo").
//
// The function accepts a variable number of string arguments in the format
// "key=value". The following keys are supported:
//   - password_length: The length of the password. Default is 20.
//   - exclude_punctuation: Whether to exclude punctuation characters. Default is true.
//   - exclude_characters: A string of characters to exclude from the password.
//   - exclude_lowercase: Whether to exclude lowercase letters. Default is false.
//   - exclude_numbers: Whether to exclude numbers. Default is false.
//   - exclude_uppercase: Whether to exclude uppercase letters. Default is false.
//   - include_space: Whether to include a space character. Default is false.
//   - require_each_included_type: Whether to require at least one character from each included type. Default is false.
//
// Called without overrides, the function returns the default configuration:
//
//	ephemeral "aws_secretsmanager_random_password" "test" {
//	  password_length     = 20
//	  exclude_punctuation = true
//	}
func ConfigRandomPassword(overrides ...string) string {
	// Default configuration values
	config := map[string]string{
		"password_length":     "20",
		"exclude_punctuation": "true",
	}

	// Additional keys without defaults
	optionalKeys := []string{
		"exclude_characters",
		"exclude_lowercase",
		"exclude_numbers",
		"exclude_uppercase",
		"include_space",
		"require_each_included_type",
	}

	// Parse overrides and update the config map
	for _, override := range overrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	// Build the Terraform configuration string
	var builder strings.Builder
	builder.WriteString(`
ephemeral "aws_secretsmanager_random_password" "test" {
`)

	// Add default keys
	fmt.Fprintf(&builder, "  password_length     = %s\n", config["password_length"])
	fmt.Fprintf(&builder, "  exclude_punctuation = %s\n", config["exclude_punctuation"])

	// Add optional keys in a consistent order
	for _, key := range optionalKeys {
		if value, exists := config[key]; exists {
			if key == "exclude_characters" {
				// Special handling for exclude_characters
				if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
					// Trim surrounding quotes
					value = value[1 : len(value)-1]
				}
				value = strings.ReplaceAll(value, `\"`, `"`)
				fmt.Fprintf(&builder, "  %s = %q\n", key, value)
				continue
			}

			// Default handling for other keys
			fmt.Fprintf(&builder, "  %s = %s\n", key, value)
		}
	}

	builder.WriteString("}\n")
	return builder.String()
}
