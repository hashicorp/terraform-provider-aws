// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedProfile
	resourceName := "aws_transfer_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "as2_id", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "profile_id"),
					resource.TestCheckResourceAttr(resourceName, "profile_type", "LOCAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccTransferProfile_certificateIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedProfile
	resourceName := "aws_transfer_profile.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_certificateIDs(rName, certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "certificate_ids.#", acctest.Ct1),
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

func TestAccTransferProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedProfile
	resourceName := "aws_transfer_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTransferProfile_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.DescribedProfile
	resourceName := "aws_transfer_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProfileConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProfileConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckProfileExists(ctx context.Context, n string, v *awstypes.DescribedProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Profile ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		output, err := tftransfer.FindProfileByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_profile" {
				continue
			}

			_, err := tftransfer.FindProfileByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_profile" "test" {
  as2_id       = %[1]q
  profile_type = "LOCAL"
}
`, rName)
}

func testAccProfileConfig_certificateIDs(rName, certificate, privateKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[2]q
  private_key = %[3]q
  usage       = "SIGNING"
}

resource "aws_transfer_profile" "test" {
  as2_id          = %[1]q
  certificate_ids = [aws_transfer_certificate.test.certificate_id]
  profile_type    = "LOCAL"
}
`, rName, certificate, privateKey)
}

func testAccProfileConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_profile" "test" {
  as2_id       = %[1]q
  profile_type = "LOCAL"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccProfileConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_profile" "test" {
  as2_id       = %[1]q
  profile_type = "LOCAL"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
