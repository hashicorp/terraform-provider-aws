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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`vpcingressconnection/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", tfapprunner.VPCIngressConnectionStatusActive),
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

func TestAccAppRunnerVPCIngressConnection_ingressVpcConfiguration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_traceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`vpcingressconnection/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", tfapprunner.VPCIngressConnectionStatusActive),
					resource.TestCheckResourceAttr(resourceName, "ingress_vpc_configuration.#", "1"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapprunner.ResourceVPCIngressConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerVPCIngressConnection_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_vpc_ingress_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apprunner.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIngressConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIngressConnectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(resourceName),
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
					testAccCheckVPCIngressConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCIngressConnectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIngressConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVPCIngressConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_vpc_ingress_connection" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		input := &apprunner.DescribeVpcIngressConnectionInput{
			VpcIngressConnectionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeVpcIngressConnectionWithContext(context.Background(), input)

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

func testAccCheckVPCIngressConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		input := &apprunner.DescribeVpcIngressConnectionInput{
			VpcIngressConnectionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeVpcIngressConnectionWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.VpcIngressConnection == nil {
			return fmt.Errorf("App Runner VPC Ingress Connection (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPCIngressConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name = %[1]q
}
`, rName)
}

func testAccVPCIngressConnectionConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCIngressConnectionConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_vpc_ingress_connection" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
