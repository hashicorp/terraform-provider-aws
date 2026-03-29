// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TargetGroupAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_idInstance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_idInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelbv2.ResourceTargetGroupAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_alb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_port(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_port(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_quic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"
	quicServerID := testAccTargetGroupAttachment_generateQUICServerID()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_quicServerID(rName, awstypes.ProtocolEnumQuic, quicServerID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "quic_server_id", quicServerID),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_quicServerId_tcpQuic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"
	quicServerID := testAccTargetGroupAttachment_generateQUICServerID()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_quicServerID(rName, awstypes.ProtocolEnumTcpQuic, quicServerID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "quic_server_id", quicServerID),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_ipAddress(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_idIPAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lb_target_group_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupAttachmentConfig_idLambda(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_Identity_noPort(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_lb_target_group_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:        tfknownvalue.AccountID(),
						names.AttrRegion:           knownvalue.StringExact(acctest.Region()),
						"target_group_arn":         knownvalue.NotNull(),
						"target_id":                knownvalue.NotNull(),
						names.AttrPort:             knownvalue.Null(),
						names.AttrAvailabilityZone: knownvalue.Null(),
						"quic_server_id":           knownvalue.Null(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_id")),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_Identity_quic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_lb_target_group_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	quicServerID := testAccTargetGroupAttachment_generateQUICServerID()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:        tfknownvalue.AccountID(),
						names.AttrRegion:           knownvalue.StringExact(acctest.Region()),
						"target_group_arn":         knownvalue.NotNull(),
						"target_id":                knownvalue.NotNull(),
						names.AttrPort:             knownvalue.NotNull(),
						names.AttrAvailabilityZone: knownvalue.Null(),
						"quic_server_id":           knownvalue.StringExact(quicServerID),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrPort)),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("quic_server_id")),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPort), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("quic_server_id"), knownvalue.StringExact(quicServerID)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPort), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("quic_server_id"), knownvalue.StringExact(quicServerID)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func TestAccELBV2TargetGroupAttachment_Identity_quicNoPort(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_lb_target_group_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	quicServerID := testAccTargetGroupAttachment_generateQUICServerID()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		CheckDestroy:             testAccCheckTargetGroupAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic_no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetGroupAttachmentExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID:        tfknownvalue.AccountID(),
						names.AttrRegion:           knownvalue.StringExact(acctest.Region()),
						"target_group_arn":         knownvalue.NotNull(),
						"target_id":                knownvalue.NotNull(),
						names.AttrPort:             knownvalue.Null(),
						names.AttrAvailabilityZone: knownvalue.Null(),
						"quic_server_id":           knownvalue.StringExact(quicServerID),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_group_arn")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("target_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New("quic_server_id")),
				},
			},

			// Step 2: Import command
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic_no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Step 3: Import block with Import ID
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic_no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPort), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("quic_server_id"), knownvalue.StringExact(quicServerID)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},

			// Step 4: Import block with Resource Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/TargetGroupAttachment/quic_no_port/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"quic_server_id": config.StringVariable(quicServerID),
				},
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_group_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPort), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("quic_server_id"), knownvalue.StringExact(quicServerID)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
		},
	})
}

func testAccCheckTargetGroupAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ELBV2Client(ctx)

		input := &elasticloadbalancingv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(rs.Primary.Attributes["target_group_arn"]),
			Targets: []awstypes.TargetDescription{{
				Id: aws.String(rs.Primary.Attributes["target_id"]),
			}},
		}

		if v := rs.Primary.Attributes[names.AttrAvailabilityZone]; v != "" {
			input.Targets[0].AvailabilityZone = aws.String(v)
		}

		if v := rs.Primary.Attributes[names.AttrPort]; v != "" {
			input.Targets[0].Port = flex.StringValueToInt32(v)
		}

		_, err := tfelbv2.FindTargetHealthDescription(ctx, conn, input)

		return err
	}
}

func testAccCheckTargetGroupAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_target_group_attachment" && rs.Type != "aws_alb_target_group_attachment" {
				continue
			}

			input := &elasticloadbalancingv2.DescribeTargetHealthInput{
				TargetGroupArn: aws.String(rs.Primary.Attributes["target_group_arn"]),
				Targets: []awstypes.TargetDescription{{
					Id: aws.String(rs.Primary.Attributes["target_id"]),
				}},
			}

			if v := rs.Primary.Attributes[names.AttrAvailabilityZone]; v != "" {
				input.Targets[0].AvailabilityZone = aws.String(v)
			}

			if v := rs.Primary.Attributes[names.AttrPort]; v != "" {
				input.Targets[0].Port = flex.StringValueToInt32(v)
			}

			_, err := tfelbv2.FindTargetHealthDescription(ctx, conn, input)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Target Group Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTargetGroupAttachment_generateQUICServerID() string {
	s := make([]byte, 8)
	if _, err := rand.Read(s); err != nil {
		return ""
	}
	return "0x" + hex.EncodeToString(s)
}

func testAccTargetGroupAttachmentCongig_baseEC2Instance(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_idInstance(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentCongig_baseEC2Instance(rName), fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_port(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentCongig_baseEC2Instance(rName), fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_quicServerID(rName string, protocol awstypes.ProtocolEnum, quicServerID string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentCongig_baseEC2Instance(rName), fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = %[2]q
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 443
  quic_server_id   = %[3]q
}
`, rName, protocol, quicServerID))
}

func testAccTargetGroupAttachmentConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentCongig_baseEC2Instance(rName), fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_alb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_idIPAddress(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupAttachmentCongig_baseEC2Instance(rName), fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  port        = 443
  protocol    = "HTTPS"
  target_type = "ip"
  vpc_id      = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  availability_zone = aws_instance.test.availability_zone
  target_group_arn  = aws_lb_target_group.test.arn
  target_id         = aws_instance.test.private_ip
}
`, rName))
}

func testAccTargetGroupAttachmentConfig_idLambda(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "elasticloadbalancing.${data.aws_partition.current.dns_suffix}"
  qualifier     = aws_lambda_alias.test.name
  source_arn    = aws_lb_target_group.test.arn
  statement_id  = "AllowExecutionFromlb"
}

resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda_elb.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_elb.lambda_handler"
  runtime       = "python3.12"
}

resource "aws_lambda_alias" "test" {
  name             = "test"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "$LATEST"
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
	EOF

}

resource "aws_lb_target_group_attachment" "test" {
  depends_on = [aws_lambda_permission.test]

  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_lambda_alias.test.arn
}
`, rName)
}
