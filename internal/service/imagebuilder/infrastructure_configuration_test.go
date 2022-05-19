package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
)

func TestAccImageBuilderInfrastructureConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("infrastructure-configuration/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceInfrastructureConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_instanceMetadataOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationInstanceMetadataOptionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.0.http_put_response_hop_limit", "64"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.0.http_tokens", "required"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_instanceProfileName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	iamInstanceProfileResourceName2 := "aws_iam_instance_profile.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationInstanceProfileName1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationInstanceProfileName2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName2, "name"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_instanceTypes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationInstanceTypes1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationInstanceTypes2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_keyPair(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyPairResourceName := "aws_key_pair.test"
	keyPairResourceName2 := "aws_key_pair.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	publicKey1, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey2, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationKeyPair1Config(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName, "key_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationKeyPair2Config(rName, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName2, "key_name"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_LoggingS3Logs_s3BucketName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketResourceName2 := "aws_s3_bucket.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationLoggingS3LogsS3BucketName1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName, "bucket"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationLoggingS3LogsS3BucketName2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName2, "bucket"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_LoggingS3Logs_s3KeyPrefix(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationLoggingS3LogsS3KeyPrefixConfig(rName, "/prefix1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationLoggingS3LogsS3KeyPrefixConfig(rName, "/prefix2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix2/"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_resourceTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationResourceTagsConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationResourceTagsConfig(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_securityGroupIDs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	securityGroupResourceName := "aws_security_group.test"
	securityGroupResourceName2 := "aws_security_group.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationSecurityGroupIds1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationSecurityGroupIds2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_snsTopicARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snsTopicResourceName := "aws_sns_topic.test"
	snsTopicResourceName2 := "aws_sns_topic.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationSNSTopicARN1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationSNSTopicARN2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_subnetID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	subnetResourceName := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationSubnetId1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationSubnetId2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInfrastructureConfigurationTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_terminateInstanceOnFailure(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationTerminateInstanceOnFailureConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationTerminateInstanceOnFailureConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "false"),
				),
			},
		},
	})
}

func testAccCheckInfrastructureConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_infrastructure_configuration" {
			continue
		}

		input := &imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetInfrastructureConfiguration(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Infrastructure Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckInfrastructureConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetInfrastructureConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccInfrastructureConfigurationBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}
`, rName)
}

func testAccInfrastructureConfigurationDescriptionConfig(rName string, description string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  description           = %[2]q
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName, description))
}

func testAccInfrastructureConfigurationInstanceMetadataOptionsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  instance_metadata_options {
    http_put_response_hop_limit = 64
    http_tokens                 = "required"
  }
}
`, rName))
}

func testAccInfrastructureConfigurationInstanceProfileName1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationInstanceProfileName2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_iam_instance_profile" "test2" {
  name = aws_iam_role.role2.name
  role = aws_iam_role.role2.name
}

resource "aws_iam_role" "role2" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test2.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationInstanceTypes1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		acctest.AvailableEC2InstanceTypeForRegion("t3.medium", "t2.medium"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationInstanceTypes2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		acctest.AvailableEC2InstanceTypeForRegion("t3.large", "t2.large"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationKeyPair1Config(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test.key_name
  name                  = %[1]q
}
`, rName, publicKey))
}

func testAccInfrastructureConfigurationKeyPair2Config(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test2" {
  key_name   = "%[1]s-2"
  public_key = %[2]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test2.key_name
  name                  = %[1]q
}
`, rName, publicKey))
}

func testAccInfrastructureConfigurationLoggingS3LogsS3BucketName1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
    }
  }
}
`, rName))
}

func testAccInfrastructureConfigurationLoggingS3LogsS3BucketName2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test2.bucket
    }
  }
}
`, rName))
}

func testAccInfrastructureConfigurationLoggingS3LogsS3KeyPrefixConfig(rName string, s3KeyPrefix string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
      s3_key_prefix  = %[2]q
    }
  }
}
`, rName, s3KeyPrefix))
}

func testAccInfrastructureConfigurationNameConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationResourceTagsConfig(rName string, resourceTagKey string, resourceTagValue string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  resource_tags = {
    %[2]q = %[3]q
  }
}
`, rName, resourceTagKey, resourceTagValue))
}

func testAccInfrastructureConfigurationSecurityGroupIds1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id]
}
`, rName))
}

func testAccInfrastructureConfigurationSecurityGroupIds2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test2.id]
}
`, rName))
}

func testAccInfrastructureConfigurationSubnetId1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test.id
}
`, rName))
}

func testAccInfrastructureConfigurationSubnetId2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 2)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test2.id
}
`, rName))
}

func testAccInfrastructureConfigurationSNSTopicARN1Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test.arn
}
`, rName))
}

func testAccInfrastructureConfigurationSNSTopicARN2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test2.arn
}
`, rName))
}

func testAccInfrastructureConfigurationTags1Config(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccInfrastructureConfigurationTags2Config(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInfrastructureConfigurationTerminateInstanceOnFailureConfig(rName string, terminateInstanceOnFailure bool) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name         = aws_iam_instance_profile.test.name
  name                          = %[1]q
  terminate_instance_on_failure = %[2]t
}
`, rName, terminateInstanceOnFailure))
}
