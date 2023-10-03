// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessTrustProviderAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.VerifiedAccessInstance
	resourceName := "aws_verifiedaccess_trust_provider_attachment.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"
	trustProviderResourceName := "aws_verifiedaccess_trust_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstance(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessTrustProviderAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessTrustProviderAttachmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessTrustProviderAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "trust_provider_id", trustProviderResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckVerifiedAccessTrustProviderAttachmentExists(ctx context.Context, n string, v *types.VerifiedAccessInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instanceId, trustProviderId, err := tfec2.VerifiedAccessTrustProviderAttachmentParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfec2.FindVerifiedAccessTrustProviderAttachmentByID(ctx, conn, instanceId, trustProviderId)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVerifiedAccessTrustProviderAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_trust_provider_attachment" {
				continue
			}

			instanceId, trustProviderId, err := tfec2.VerifiedAccessTrustProviderAttachmentParseId(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfec2.FindVerifiedAccessTrustProviderAttachmentByID(ctx, conn, instanceId, trustProviderId)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Trust Provider Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVerifiedAccessTrustProviderAttachmentConfig_basic() string {
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

resource "aws_verifiedaccess_trust_provider_attachment" "test" {
  instance_id       = aws_verifiedaccess_instance.test.id
  trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}
`
}
