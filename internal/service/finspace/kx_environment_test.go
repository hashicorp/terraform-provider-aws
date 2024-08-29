// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxEnvironment_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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

func TestAccFinSpaceKxEnvironment_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_updateName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_description(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
				),
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_customDNS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_dnsConfig(rName, "example.finspace.amazon.aws.com", "10.0.0.76"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_dns_configuration.*", map[string]string{
						"custom_dns_server_name": "example.finspace.amazon.aws.com",
						"custom_dns_server_ip":   "10.0.0.76",
					}),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_dnsConfig(rName, "updated.finspace.amazon.com", "10.0.0.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_dns_configuration.*", map[string]string{
						"custom_dns_server_name": "updated.finspace.amazon.com",
						"custom_dns_server_ip":   "10.0.0.24",
					}),
				),
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_transitGateway(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_tgwConfig(rName, "100.64.0.0/26"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.*", map[string]string{
						"routable_cidr_space": "100.64.0.0/26",
					}),
				),
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_attachmentNetworkACLConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_attachmentNetworkACLConfig(rName, "100.64.0.0/26"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.*", map[string]string{
						"routable_cidr_space": "100.64.0.0/26",
					}),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_action":       "allow",
						names.AttrCIDRBlock: "0.0.0.0/0",
						"rule_number":       acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_attachmentNetworkACLConfig2(rName, "100.64.0.0/26"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.*", map[string]string{
						"routable_cidr_space": "100.64.0.0/26",
					}),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_action":       "allow",
						names.AttrCIDRBlock: "0.0.0.0/0",
						"rule_number":       acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.*", map[string]string{
						names.AttrProtocol:  acctest.Ct4,
						"rule_action":       "allow",
						names.AttrCIDRBlock: "0.0.0.0/0",
						"rule_number":       "20",
					}),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_attachmentNetworkACLConfig(rName, "100.64.0.0/26"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.*", map[string]string{
						"routable_cidr_space": "100.64.0.0/26",
					}),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "transit_gateway_configuration.0.attachment_network_acl_configuration.*", map[string]string{
						names.AttrProtocol:  "6",
						"rule_action":       "allow",
						names.AttrCIDRBlock: "0.0.0.0/0",
						"rule_number":       acctest.Ct1,
					}),
				),
			},
		},
	})
}

func TestAccFinSpaceKxEnvironment_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxenvironment finspace.GetKxEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxEnvironmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKxEnvironmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxEnvironmentExists(ctx, resourceName, &kxenvironment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckKxEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_finspace_kx_environment" {
				continue
			}

			input := &finspace.GetKxEnvironmentInput{
				EnvironmentId: aws.String(rs.Primary.ID),
			}
			out, err := conn.GetKxEnvironment(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}
			if out.Status == types.EnvironmentStatusDeleted {
				return nil
			}
			return create.Error(names.FinSpace, create.ErrActionCheckingDestroyed, tffinspace.ResNameKxEnvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKxEnvironmentExists(ctx context.Context, name string, kxenvironment *finspace.GetKxEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxEnvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxEnvironment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)
		resp, err := conn.GetKxEnvironment(ctx, &finspace.GetKxEnvironmentInput{
			EnvironmentId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxEnvironment, rs.Primary.ID, err)
		}

		*kxenvironment = *resp

		return nil
	}
}

func testAccKxEnvironmentConfigBase() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}
`
}

func testAccKxEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}
`, rName))
}

func testAccKxEnvironmentConfig_description(rName, desc string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_finspace_kx_environment" "test" {
  name        = %[1]q
  kms_key_id  = aws_kms_key.test.arn
  description = %[2]q
}
`, rName, desc))
}

func testAccKxEnvironmentConfig_tgwConfig(rName, cidr string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = "test"
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.test.id
    routable_cidr_space = %[2]q
  }
}
`, rName, cidr))
}

func testAccKxEnvironmentConfig_attachmentNetworkACLConfig(rName, cidr string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = "test"
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.test.id
    routable_cidr_space = %[2]q
    attachment_network_acl_configuration {
      rule_number = 1
      protocol    = "6"
      rule_action = "allow"
      cidr_block  = "0.0.0.0/0"
      port_range {
        from = 53
        to   = 53
      }
      icmp_type_code {
        type = -1
        code = -1
      }
    }
  }
}
`, rName, cidr))
}

func testAccKxEnvironmentConfig_attachmentNetworkACLConfig2(rName, cidr string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = "test"
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  transit_gateway_configuration {
    transit_gateway_id  = aws_ec2_transit_gateway.test.id
    routable_cidr_space = %[2]q
    attachment_network_acl_configuration {
      rule_number = 1
      protocol    = "6"
      rule_action = "allow"
      cidr_block  = "0.0.0.0/0"
      port_range {
        from = 53
        to   = 53
      }
      icmp_type_code {
        type = -1
        code = -1
      }
    }
    attachment_network_acl_configuration {
      rule_number = 20
      protocol    = "4"
      rule_action = "allow"
      cidr_block  = "0.0.0.0/0"
      port_range {
        from = 51
        to   = 51
      }
      icmp_type_code {
        type = -1
        code = -1
      }
    }
  }
}
`, rName, cidr))
}

func testAccKxEnvironmentConfig_dnsConfig(rName, serverName, serverIP string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  custom_dns_configuration {
    custom_dns_server_name = %[2]q
    custom_dns_server_ip   = %[3]q
  }
}
`, rName, serverName, serverIP))
}

func testAccKxEnvironmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccKxEnvironmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccKxEnvironmentConfigBase(),
		fmt.Sprintf(`
resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
