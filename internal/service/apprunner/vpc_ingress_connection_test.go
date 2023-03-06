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

func TestAccAppRunnerVPCIngressConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"
	vpcResourceName := "aws_vpc.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.apprunner"
	appRunnerServiceResourceName := "aws_apprunner_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`vpcingressconnection/%s/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", tfapprunner.VPCIngressConnectionStatusActive),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", appRunnerServiceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "ingress_vpc_configuration.0.vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ingress_vpc_configuration.0.vpc_endpoint_id", vpcEndpointResourceName, "id"),
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

func TestAccAppRunnerVPCIngressConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceVPCIngressConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerVPCIngressConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(ctx, resourceName),
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
				Config: testAccVPCIngressConnectionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCIngressConnectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVPCIngressConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_vpc_ingress_connection" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn()

			input := &apprunner.DescribeVpcIngressConnectionInput{
				VpcIngressConnectionArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeVpcIngressConnectionWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			if output != nil && output.VpcIngressConnection != nil && aws.StringValue(output.VpcIngressConnection.Status) != apprunner.VpcIngressConnectionStatusDeleted {
				return fmt.Errorf("App Runner VPC Ingress Connection (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckVPCIngressConnectionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn()

		input := &apprunner.DescribeVpcIngressConnectionInput{
			VpcIngressConnectionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeVpcIngressConnectionWithContext(ctx, input)

		if err != nil {
			return err
		}

		if output == nil || output.VpcIngressConnection == nil {
			return fmt.Errorf("App Runner VPC Ingress Connection (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCIngressConnectionConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/aws-containers/hello-app-runner:latest"
      image_repository_type = "ECR_PUBLIC"
    }
    auto_deployments_enabled = false
  }

  network_configuration {
    ingress_configuration {
      is_publicly_accessible = false
    }
  }
}

resource "aws_vpc_endpoint" "apprunner" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.apprunner.requests"
  vpc_endpoint_type = "Interface"

  subnet_ids = aws_subnet.test[*].id

  security_group_ids = [
    aws_vpc.test.default_security_group_id,
  ]
}
`, rName))
}

func testAccVPCIngressConnectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCIngressConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name        = %[1]q
  service_arn = aws_apprunner_service.test.arn

  ingress_vpc_configuration {
    vpc_id          = aws_vpc.test.id
    vpc_endpoint_id = aws_vpc_endpoint.apprunner.id
  }
}
`, rName))
}

func testAccVPCIngressConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCIngressConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name        = %[1]q
  service_arn = aws_apprunner_service.test.arn

  ingress_vpc_configuration {
    vpc_id          = aws_vpc.test.id
    vpc_endpoint_id = aws_vpc_endpoint.apprunner.id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccVPCIngressConnectionConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCIngressConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name        = %[1]q
  service_arn = aws_apprunner_service.test.arn

  ingress_vpc_configuration {
    vpc_id          = aws_vpc.test.id
    vpc_endpoint_id = aws_vpc_endpoint.apprunner.id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
