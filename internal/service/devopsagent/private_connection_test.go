// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdevopsagent "github.com/hashicorp/terraform-provider-aws/internal/service/devopsagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDevOpsAgentPrivateConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var privateconnection devopsagent.DescribePrivateConnectionOutput
	rName := acctest.RandomWithPrefix(t, "tf-test")
	resourceName := "aws_devopsagent_private_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &privateconnection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aidevops", regexache.MustCompile(`private-connection/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(awstypes.PrivateConnectionTypeSelfManaged)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportStateVerifyIgnore: []string{
					"resource_configuration_id",
					names.AttrCertificate,
				},
			},
		},
	})
}

func TestAccDevOpsAgentPrivateConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var privateconnection devopsagent.DescribePrivateConnectionOutput
	rName := acctest.RandomWithPrefix(t, "tf-test")
	resourceName := "aws_devopsagent_private_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &privateconnection),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsagent.ResourcePrivateConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDevOpsAgentPrivateConnection_certificate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var privateconnection devopsagent.DescribePrivateConnectionOutput
	rName := acctest.RandomWithPrefix(t, "tf-test")
	resourceName := "aws_devopsagent_private_connection.test"

	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCert := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	cert1 := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCert, key, "example.com")
	cert2 := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCert, key, "example2.com")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateConnectionConfig_certificate(rName, cert1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &privateconnection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportStateVerifyIgnore: []string{
					"resource_configuration_id",
					names.AttrCertificate,
				},
			},
			{
				Config: testAccPrivateConnectionConfig_certificate(rName, cert2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &privateconnection),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
				),
			},
		},
	})
}

func TestAccDevOpsAgentPrivateConnection_nameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 devopsagent.DescribePrivateConnectionOutput
	rName1 := acctest.RandomWithPrefix(t, "tf-test")
	rName2 := acctest.RandomWithPrefix(t, "tf-test")
	resourceName := "aws_devopsagent_private_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateConnectionConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccPrivateConnectionConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccDevOpsAgentPrivateConnection_serviceManaged(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var privateconnection devopsagent.DescribePrivateConnectionOutput
	rName := acctest.RandomWithPrefix(t, "tf-test")
	resourceName := "aws_devopsagent_private_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateConnectionConfig_serviceManaged(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateConnectionExists(ctx, t, resourceName, &privateconnection),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "aidevops", regexache.MustCompile(`private-connection/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, string(awstypes.PrivateConnectionTypeServiceManaged)),
					resource.TestCheckResourceAttrSet(resourceName, "host_address"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateId:                        rName,
				ImportStateVerifyIgnore: []string{
					"resource_configuration_id",
					names.AttrCertificate,
					names.AttrSubnetIDs,
				},
			},
		},
	})
}

func testAccCheckPrivateConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsagent_private_connection" {
				continue
			}

			_, err := tfdevopsagent.FindPrivateConnectionByName(ctx, conn, rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNamePrivateConnection, rs.Primary.Attributes[names.AttrName], err)
			}

			return create.Error(names.DevOpsAgent, create.ErrActionCheckingDestroyed, tfdevopsagent.ResNamePrivateConnection, rs.Primary.Attributes[names.AttrName], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPrivateConnectionExists(ctx context.Context, t *testing.T, name string, privateconnection *devopsagent.DescribePrivateConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNamePrivateConnection, name, errors.New("not found"))
		}

		connName := rs.Primary.Attributes[names.AttrName]
		if connName == "" {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNamePrivateConnection, name, errors.New("name not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

		resp, err := tfdevopsagent.FindPrivateConnectionByName(ctx, conn, connName)
		if err != nil {
			return create.Error(names.DevOpsAgent, create.ErrActionCheckingExistence, tfdevopsagent.ResNamePrivateConnection, connName, err)
		}

		*privateconnection = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DevOpsAgentClient(ctx)

	input := devopsagent.ListPrivateConnectionsInput{}

	_, err := conn.ListPrivateConnections(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPrivateConnectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPrivateConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_devopsagent_private_connection" "test" {
  name                      = %[1]q
  mode                      = "SELF_MANAGED"
  resource_configuration_id = aws_vpclattice_resource_configuration.test.id
}
`, rName))
}

func testAccPrivateConnectionConfig_certificate(rName, certificate string) string {
	return acctest.ConfigCompose(testAccPrivateConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_devopsagent_private_connection" "test" {
  name                      = %[1]q
  mode                      = "SELF_MANAGED"
  resource_configuration_id = aws_vpclattice_resource_configuration.test.id
  certificate               = %[2]q
}
`, rName, certificate))
}

func testAccPrivateConnectionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}
`, rName)
}

func testAccPrivateConnectionConfig_serviceManaged(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_devopsagent_private_connection" "test" {
  name         = %[1]q
  mode         = "SERVICE_MANAGED"
  host_address = "10.0.0.1"
  vpc_id       = aws_vpc.test.id
  subnet_ids   = aws_subnet.test[*].id
}
`, rName))
}

func randomPrivateConnectionName(t *testing.T) string {
	return acctest.RandString(t, 20)
}
