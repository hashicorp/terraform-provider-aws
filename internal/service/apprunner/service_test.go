package apprunner_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
)

func TestAccAppRunnerService_ImageRepository_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`service/%s/.+`, rName))),
					acctest.MatchResourceAttrRegionalARN(resourceName, "auto_scaling_configuration_arn", "apprunner", regexp.MustCompile(`autoscalingconfiguration/DefaultConfiguration/1/.+`)),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.path", "/"),
					// Only check the following attribute values for health_check and instance configurations
					// are set as their defaults differ in the API documentation and API itself
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.interval"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.timeout"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.healthy_threshold"),
					resource.TestCheckResourceAttrSet(resourceName, "health_check_configuration.0.unhealthy_threshold"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.cpu"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.memory"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.instance_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.egress_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.vpc_connector_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttrSet(resourceName, "service_url"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.auto_deployments_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_identifier", "public.ecr.aws/nginx/nginx:latest"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_repository_type", apprunner.ImageRepositoryTypeEcrPublic),
					resource.TestCheckResourceAttr(resourceName, "status", apprunner.ServiceStatusRunning),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAppRunnerService_ImageRepository_autoScaling(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	autoScalingResourceName := "aws_apprunner_auto_scaling_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
				),
			},
			{
				Config: testAccServiceConfig_ImageRepository_autoScalingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_configuration_arn", autoScalingResourceName, "arn"),
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

func TestAccAppRunnerService_ImageRepository_encryption(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_encryptionConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation; EncryptionConfiguration (or lack thereof) Forces New resource
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAppRunnerService_ImageRepository_healthCheck(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_healthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test resource recreation; HealthConfiguration Forces New resource
				Config: testAccServiceConfig_ImageRepository_updateHealthCheckConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.protocol", apprunner.HealthCheckProtocolTcp),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.timeout", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check_configuration.0.unhealthy_threshold", "4"),
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

func TestAccAppRunnerService_ImageRepository_instance_NoInstanceRole(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_InstanceConfiguration_noInstanceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "1024"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.instance_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "3072"),
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

func TestAccAppRunnerService_ImageRepository_instance_Update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_instanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "1024"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "3072"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_ImageRepository_updateInstanceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "2048"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.instance_role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "4096"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.cpu", "2048"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.memory", "4096"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_configuration.0.instance_role_arn"), // The IAM Role is not unset
				),
			},
		},
	})
}

func TestAccAppRunnerService_ImageRepository_networkConfiguration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"
	vpcConnectorResourceName := "aws_apprunner_vpc_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_networkConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.egress_configuration.0.egress_type", "VPC"),
					resource.TestCheckResourceAttrPair(resourceName, "network_configuration.0.egress_configuration.0.vpc_connector_arn", vpcConnectorResourceName, "arn"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19469
func TestAccAppRunnerService_ImageRepository_runtimeEnvironmentVars(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_ImageRepository_runtimeEnvVars(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_configuration.0.image_repository.0.image_configuration.0.runtime_environment_variables.APP_NAME", rName),
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

func TestAccAppRunnerService_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_imageRepository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapprunner.ResourceService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerService_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
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
				Config: testAccServiceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckServiceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_service" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceWithContext(context.Background(), input)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Service != nil && aws.StringValue(output.Service.Status) != apprunner.ServiceStatusDeleted {
			return fmt.Errorf("App Runner Service (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckServiceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		input := &apprunner.DescribeServiceInput{
			ServiceArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.Service == nil {
			return fmt.Errorf("App Runner Service (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn
	ctx := context.Background()

	input := &apprunner.ListServicesInput{}

	_, err := conn.ListServicesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccServiceConfig_imageRepository(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_runtimeEnvVars(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
        runtime_environment_variables = {
          APP_NAME = %[1]q
        }
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_autoScalingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_auto_scaling_configuration_version" "test" {
  auto_scaling_configuration_name = %[1]q
}

resource "aws_apprunner_service" "test" {
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.test.arn

  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[1]q
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  encryption_configuration {
    kms_key = aws_kms_key.test.arn
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_healthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold = 2
    timeout           = 5
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_updateHealthCheckConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  health_check_configuration {
    healthy_threshold   = 2
    timeout             = 10
    unhealthy_threshold = 4
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_networkConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_apprunner_vpc_connector" "test" {
  vpc_connector_name = %[1]q
  subnets            = [aws_subnet.test.id]
  security_groups    = [aws_security_group.test.id]
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  network_configuration {
    egress_configuration {
      egress_type       = "VPC"
      vpc_connector_arn = aws_apprunner_vpc_connector.test.arn
    }
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "tasks.apprunner.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
`, rName)
}

func testAccServiceConfig_ImageRepository_instanceConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "1 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "3 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_ImageRepository_InstanceConfiguration_noInstanceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu    = "1 vCPU"
    memory = "3 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName)
}

func testAccServiceConfig_ImageRepository_updateInstanceConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIAMRole(rName),
		fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  instance_configuration {
    cpu               = "2 vCPU"
    instance_role_arn = aws_iam_role.test.arn
    memory            = "4 GB"
  }

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName))
}

func testAccServiceConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
