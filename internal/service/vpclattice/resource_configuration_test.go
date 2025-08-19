// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeResourceConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceconfiguration vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"
	resourceGatewayName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &resourceconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.ip_address_type", "IPV4"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCLatticeResourceConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"
	resourceGatewayName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_update(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allow_association_to_shareable_service_network", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.ip_address_type", "IPV4"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceConfigurationConfig_update(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allow_association_to_shareable_service_network", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.1", "8080"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.dns_resource.0.ip_address_type", "IPV4"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCLatticeResourceConfiguration_ipAddress(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"
	resourceGatewayName := "aws_vpclattice_resource_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_ipAddress(rName, "10.0.0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allow_association_to_shareable_service_network", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.ip_resource.0.ip_address", "10.0.0.1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceConfigurationConfig_ipAddress(rName, "10.0.0.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "allow_association_to_shareable_service_network", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.ip_resource.0.ip_address", "10.0.0.2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCLatticeResourceConfiguration_parentChild(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceconfiguration vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"
	resourceGatewayName := "aws_vpclattice_resource_gateway.test"
	resourceParentName := "aws_vpclattice_resource_configuration.parent"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_parentChild(rName, "10.0.0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &resourceconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_group_id", resourceParentName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.ResourceConfigurationTypeChild)),
					resource.TestCheckResourceAttr(resourceParentName, names.AttrType, string(types.ResourceConfigurationTypeGroup)),
					resource.TestCheckResourceAttr(resourceName, "port_ranges.0", "80"),
					resource.TestCheckResourceAttr(resourceName, "resource_configuration_definition.0.ip_resource.0.ip_address", "10.0.0.1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCLatticeResourceConfiguration_arnResource(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceconfiguration vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"
	resourceGatewayName := "aws_vpclattice_resource_gateway.test"
	resourceArnName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_arnResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &resourceconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_gateway_identifier", resourceGatewayName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "resource_configuration_definition.0.arn_resource.0.arn", resourceArnName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`resourceconfiguration/+.`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCLatticeResourceConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceconfiguration vpclattice.GetResourceConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_resource_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceConfigurationExists(ctx, resourceName, &resourceconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceResourceConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_resource_configuration" {
				continue
			}

			_, err := tfvpclattice.FindResourceConfigurationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Resource Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceConfigurationExists(ctx context.Context, n string, v *vpclattice.GetResourceConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindResourceConfigurationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourceConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

`, rName))
}

func testAccResourceConfigurationConfig_update(rName, shareable string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  allow_association_to_shareable_service_network = %[2]s

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80", "8080"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

`, rName, shareable))
}

func testAccResourceConfigurationConfig_ipAddress(rName, ip string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    ip_resource {
      ip_address = %[2]q
    }
  }
}

`, rName, ip))
}

func testAccResourceConfigurationConfig_parentChild(rName, ip string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_basic(rName),
		fmt.Sprintf(`

resource "aws_vpclattice_resource_configuration" "parent" {
  name = "%[1]s-parent"

  protocol = "TCP"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id
  type                        = "GROUP"
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  port_ranges = ["80"]

  resource_configuration_group_id = aws_vpclattice_resource_configuration.parent.id
  type                            = "CHILD"

  resource_configuration_definition {
    ip_resource {
      ip_address = %[2]q
    }
  }
}

`, rName, ip))
}

func testAccResourceConfigurationConfig_arnResource(rName string) string {
	return acctest.ConfigCompose(testAccResourceGatewayConfig_multipleSubnets(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  type = "ARN"

  resource_configuration_definition {
    arn_resource {
      arn = aws_rds_cluster_instance.test.arn
    }
  }
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
  publicly_accessible = false
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]
}
`, rName))
}
