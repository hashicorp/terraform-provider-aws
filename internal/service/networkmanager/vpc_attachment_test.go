// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerVPCAttachment_basic(t *testing.T) {
	const (
		resourceName            = "aws_networkmanager_vpc_attachment.test"
		coreNetworkResourceName = "aws_networkmanager_core_network.test"
		vpcResourceName         = "aws_vpc.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
		expectedState      string
	}{
		"acceptance_required": {
			acceptanceRequired: true,
			expectedState:      networkmanager.AttachmentStatePendingAttachmentAcceptance,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
			expectedState:      networkmanager.AttachmentStateAvailable,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
							acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
							resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "attachment_type", "VPC"),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
							resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
							resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
							acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
							resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, vpcResourceName, names.AttrARN),
							resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
							resource.TestCheckResourceAttr(resourceName, names.AttrState, testcase.expectedState),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
							resource.TestCheckResourceAttrPair(resourceName, "vpc_arn", vpcResourceName, names.AttrARN),
						),
					},
					{
						ResourceName:      resourceName,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerVPCAttachment_Attached_basic(t *testing.T) {
	const (
		resourceName            = "aws_networkmanager_vpc_attachment.test"
		coreNetworkResourceName = "aws_networkmanager_core_network.test"
		vpcResourceName         = "aws_vpc.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_Attached_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
							acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
							resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "attachment_type", "VPC"),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetworkResourceName, names.AttrARN),
							resource.TestCheckResourceAttrPair(resourceName, "core_network_id", coreNetworkResourceName, names.AttrID),
							resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
							resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
							acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
							resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, vpcResourceName, names.AttrARN),
							resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
							resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
							resource.TestCheckResourceAttrPair(resourceName, "vpc_arn", vpcResourceName, names.AttrARN),
						),
					},
					{
						ResourceName:      resourceName,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerVPCAttachment_disappears(t *testing.T) {
	const (
		resourceName = "aws_networkmanager_vpc_attachment.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
							acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceVPCAttachment(), resourceName),
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
		})
	}
}

func TestAccNetworkManagerVPCAttachment_Attached_disappears(t *testing.T) { // nosemgrep:ci.acceptance-test-naming-parent-disappears
	const (
		resourceName           = "aws_networkmanager_vpc_attachment.test"
		attachmentResourceName = "aws_networkmanager_attachment_accepter.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_Attached_basic(rName, testcase.acceptanceRequired),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
							acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceVPCAttachment(), resourceName),
						),
						ExpectNonEmptyPlan: true,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PostApplyPostRefresh: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
								plancheck.ExpectResourceAction(attachmentResourceName, plancheck.ResourceActionCreate),
							},
						},
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerVPCAttachment_Attached_disappearsAccepter(t *testing.T) {
	const (
		resourceName           = "aws_networkmanager_vpc_attachment.test"
		attachmentResourceName = "aws_networkmanager_attachment_accepter.test"
	)

	ctx := acctest.Context(t)
	var v networkmanager.VpcAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_Attached_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceAttachmentAccepter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(attachmentResourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNetworkManagerVPCAttachment_tags(t *testing.T) {
	const (
		resourceName = "aws_networkmanager_vpc_attachment.test"
	)

	ctx := acctest.Context(t)
	var v networkmanager.VpcAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_tags2(rName, "segment", "shared", "Name", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
				),
			},
			{
				Config: testAccVPCAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
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

func TestAccNetworkManagerVPCAttachment_update(t *testing.T) {
	const (
		resourceName = "aws_networkmanager_vpc_attachment.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
		expectedState      string
		expectRecreation   bool
	}{
		"acceptance_required": {
			acceptanceRequired: true,
			expectedState:      networkmanager.AttachmentStatePendingAttachmentAcceptance,
			expectRecreation:   true,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
			expectedState:      networkmanager.AttachmentStateAvailable,
			expectRecreation:   false,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v1, v2, v3, v4 networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_updates(rName, testcase.acceptanceRequired, 2, true, false),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v1),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtTrue),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_updates(rName, testcase.acceptanceRequired, 1, false, true),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v2),
							testAccCheckVPCAttachmentRecreated(&v1, &v2, testcase.expectRecreation),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtTrue),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_updates(rName, testcase.acceptanceRequired, 2, false, false),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v3),
							testAccCheckVPCAttachmentRecreated(&v2, &v3, testcase.expectRecreation),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_updates(rName, testcase.acceptanceRequired, 2, false, true),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v4),
							testAccCheckVPCAttachmentRecreated(&v3, &v4, testcase.expectRecreation),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtTrue),
						),
					},
					{
						ResourceName:      resourceName,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccNetworkManagerVPCAttachment_Attached_update(t *testing.T) {
	const (
		resourceName = "aws_networkmanager_vpc_attachment.test"
	)

	t.Parallel()

	testcases := map[string]struct {
		acceptanceRequired bool
		expectedState      string
	}{
		"acceptance_required": {
			acceptanceRequired: true,
			expectedState:      networkmanager.AttachmentStatePendingAttachmentAcceptance,
		},

		"acceptance_not_required": {
			acceptanceRequired: false,
			expectedState:      networkmanager.AttachmentStateAvailable,
		},
	}

	for name, testcase := range testcases { //nolint:paralleltest // false positive
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			var v1, v2, v3, v4 networkmanager.VpcAttachment
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccVPCAttachmentConfig_Attached_updates(rName, testcase.acceptanceRequired, 2, true, false),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v1),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtTrue),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_Attached_updates(rName, testcase.acceptanceRequired, 1, false, true),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v2),
							testAccCheckVPCAttachmentRecreated(&v1, &v2, false),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct1),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtTrue),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_Attached_updates(rName, testcase.acceptanceRequired, 2, false, false),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v3),
							testAccCheckVPCAttachmentRecreated(&v2, &v3, false),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtFalse),
						),
					},
					{
						Config: testAccVPCAttachmentConfig_Attached_updates(rName, testcase.acceptanceRequired, 2, false, true),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVPCAttachmentExists(ctx, resourceName, &v4),
							testAccCheckVPCAttachmentRecreated(&v3, &v4, false),
							resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", acctest.Ct2),
							resource.TestCheckResourceAttr(resourceName, "options.0.appliance_mode_support", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, "options.0.ipv6_support", acctest.CtTrue),
						),
					},
					{
						ResourceName:      resourceName,
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func testAccCheckVPCAttachmentExists(ctx context.Context, n string, v *networkmanager.VpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager VPC Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		output, err := tfnetworkmanager.FindVPCAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_vpc_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindVPCAttachmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager VPC Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCAttachmentRecreated(v1, v2 *networkmanager.VpcAttachment, expectRecreation bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckAttachmentRecreated(v1.Attachment, v2.Attachment, expectRecreation)
	}
}

func testAccCheckAttachmentRecreated(v1, v2 *networkmanager.Attachment, expectRecreation bool) error {
	v1CreatedAt := aws.TimeValue(v1.CreatedAt)
	v2CreatedAt := aws.TimeValue(v2.CreatedAt)
	cmp := v1CreatedAt.Compare(v2CreatedAt)
	if expectRecreation && cmp != -1 {
		return fmt.Errorf("Attachment not recreated: v1.CreatedAt=%q, v2.CreatedAt=%q", v1CreatedAt, v2CreatedAt)
	} else if !expectRecreation && cmp != 0 {
		return fmt.Errorf("Attachment recreated: v1.CreatedAt=%q, v2.CreatedAt=%q", v1CreatedAt, v2CreatedAt)
	}
	return nil
}

func testAccVPCAttachmentConfig_basic(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, requireAcceptance), `
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn
}
`)
}

func testAccVPCAttachmentConfig_Attached_basic(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, requireAcceptance), `
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`)
}

func testAccVPCAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, false),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, false),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCAttachmentConfig_updates(rName string, requireAcceptance bool, nSubnets int, applianceModeSupport, ipv6Support bool) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, requireAcceptance),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = slice(aws_subnet.test[*].arn, 0, %[2]d)
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  options {
    appliance_mode_support = %[3]t
    ipv6_support           = %[4]t
  }
}
`, rName, nSubnets, applianceModeSupport, ipv6Support))
}

func testAccVPCAttachmentConfig_Attached_updates(rName string, requireAcceptance bool, nSubnets int, applianceModeSupport, ipv6Support bool) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, requireAcceptance),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = slice(aws_subnet.test[*].arn, 0, %[2]d)
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  options {
    appliance_mode_support = %[3]t
    ipv6_support           = %[4]t
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`, rName, nSubnets, applianceModeSupport, ipv6Support))
}

func testAccVPCAttachmentConfig_base(rName string, requireAcceptance bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.name
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = %[2]t
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }

  attachment_policies {
    rule_number = 1

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}
`, rName, requireAcceptance))
}
