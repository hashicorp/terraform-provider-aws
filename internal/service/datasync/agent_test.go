// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncAgent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("datasync", regexache.MustCompile(`agent/agent-.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_link_endpoint"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("security_group_arns"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subnet_arns"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVPCEndpointID), knownvalue.StringExact("")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress},
			},
		},
	})
}

func TestAccDataSyncAgent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdatasync.ResourceAgent(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDataSyncAgent_agentName(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_name(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
				},
			},
			{
				Config: testAccAgentConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress},
			},
		},
	})
}

func TestAccDataSyncAgent_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress},
			},
			{
				Config: testAccAgentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccAgentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccDataSyncAgent_vpcEndpointID(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_vpcEndpointID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("security_group_arns"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("security_group_arns").AtSliceIndex(0), securityGroupResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subnet_arns"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("subnet_arns").AtSliceIndex(0), subnetResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("names.AttrVPCEndpointID"), vpcEndpointResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress, "private_link_ip"},
			},
		},
	})
}

func TestAccDataSyncAgent_advancedMode(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_advancedMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("datasync", regexache.MustCompile(`agent/agent-.+`))),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress},
			},
		},
	})
}

func TestAccDataSyncAgent_AdvancedMode_vpcEndpointID(t *testing.T) {
	ctx := acctest.Context(t)
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datasync_agent.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentConfig_advancedModeVPCEndpointID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(ctx, t, resourceName, &agent),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("datasync", regexache.MustCompile(`agent/agent-.+`))),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", names.AttrIPAddress},
			},
		},
	})
}

func testAccCheckAgentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_agent" {
				continue
			}

			_, err := tfdatasync.FindAgentByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Agent %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAgentExists(ctx context.Context, t *testing.T, n string, v *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DataSyncClient(ctx)

		output, err := tfdatasync.FindAgentByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAgentAgentConfig_base(rName, ssmParameterName string, preferredInstanceTypes ...string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		// See https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html#ec2-instance-types.
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", preferredInstanceTypes...),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = %[2]q
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test[0].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_internet_gateway.test]

  ami                    = data.aws_ssm_parameter.aws_service_datasync_ami.value
  instance_type          = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test[0].id

  # The Instance must have a public IP address because the aws_datasync_agent retrieves
  # the activation key by making an HTTP request to the instance
  associate_public_ip_address = true

  tags = {
    Name = %[1]q
  }
}
`, rName, ssmParameterName))
}

func testAccAgentAgentConfig_baseBasicMode(rName string) string {
	return testAccAgentAgentConfig_base(rName, "/aws/service/datasync/ami", "m5.2xlarge", "m5.4xlarge")
}

func testAccAgentAgentConfig_baseAdvancedMode(rName string) string {
	return testAccAgentAgentConfig_base(rName, "/aws/service/datasync/ami/v3", "m6a.2xlarge")
}

func testAccAgentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseBasicMode(rName), `
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
}
`)
}

func testAccAgentConfig_name(rName, agentName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseBasicMode(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, agentName))
}

func testAccAgentConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseBasicMode(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccAgentConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseBasicMode(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccAgentConfig_vpcEndpointID(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseBasicMode(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  name                  = %[1]q
  security_group_arns   = [aws_security_group.test.arn]
  subnet_arns           = [aws_subnet.test[0].arn]
  vpc_endpoint_id       = aws_vpc_endpoint.test.id
  ip_address            = aws_instance.test.public_ip
  private_link_endpoint = data.aws_network_interface.test.private_ip
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  service_name       = "com.amazonaws.${data.aws_region.current.region}.datasync"
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test[0].id]
  vpc_endpoint_type  = "Interface"

  tags = {
    Name = %[1]q
  }
}

data "aws_network_interface" "test" {
  id = tolist(aws_vpc_endpoint.test.network_interface_ids)[0]
}
`, rName))
}

func testAccAgentConfig_advancedMode(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseAdvancedMode(rName), `
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
}
`)
}

func testAccAgentConfig_advancedModeVPCEndpointID(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_baseAdvancedMode(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  name                  = %[1]q
  security_group_arns   = [aws_security_group.test.arn]
  subnet_arns           = [aws_subnet.test[0].arn]
  vpc_endpoint_id       = aws_vpc_endpoint.test.id
  ip_address            = aws_instance.test.public_ip
  private_link_endpoint = data.aws_network_interface.test.private_ip
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  service_name       = "com.amazonaws.${data.aws_region.current.region}.datasync"
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test[0].id]
  vpc_endpoint_type  = "Interface"

  tags = {
    Name = %[1]q
  }
}

data "aws_network_interface" "test" {
  id = tolist(aws_vpc_endpoint.test.network_interface_ids)[0]
}
`, rName))
}
