// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccVerifiedAccessInstanceTrustProviderAttachment_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_verifiedaccess_instance_trust_provider_attachment.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	trustProviderResourceName := "aws_verifiedaccess_trust_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstance(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceTrustProviderAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessInstanceTrustProviderAttachmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceTrustProviderAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_trust_provider_id", trustProviderResourceName, names.AttrID),
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

func testAccVerifiedAccessInstanceTrustProviderAttachment_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_verifiedaccess_instance_trust_provider_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVerifiedAccessSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstance(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceTrustProviderAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessInstanceTrustProviderAttachmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceTrustProviderAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessInstanceTrustProviderAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessInstanceTrustProviderAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		err := tfec2.FindVerifiedAccessInstanceTrustProviderAttachmentExists(ctx, conn, rs.Primary.Attributes["verifiedaccess_instance_id"], rs.Primary.Attributes["verifiedaccess_trust_provider_id"])

		return err
	}
}

func testAccCheckVerifiedAccessInstanceTrustProviderAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_instance_trust_provider_attachment" {
				continue
			}

			err := tfec2.FindVerifiedAccessInstanceTrustProviderAttachmentExists(ctx, conn, rs.Primary.Attributes["verifiedaccess_instance_id"], rs.Primary.Attributes["verifiedaccess_trust_provider_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Instance Trust Provider Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVerifiedAccessInstanceTrustProviderAttachmentConfig_basic() string {
	return `
resource "aws_verifiedaccess_instance" "test" {}

resource "aws_verifiedaccess_trust_provider" "test" {
  device_trust_provider_type = "jamf"
  policy_reference_name      = "test"
  trust_provider_type        = "device"

  device_options {
    tenant_id = "test"
  }
}

resource "aws_verifiedaccess_instance_trust_provider_attachment" "test" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.test.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}
`
}
