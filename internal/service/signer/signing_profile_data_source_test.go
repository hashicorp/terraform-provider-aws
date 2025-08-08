// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSignerSigningProfileDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_signer_signing_profile.test"
	resourceName := "aws_signer_signing_profile.test"
	rString := sdkacctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSingerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, signer.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningProfileDataSourceConfig_basic(profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform_id", resourceName, "platform_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_validity_period.value", resourceName, "signature_validity_period.value"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signature_validity_period.type", resourceName, "signature_validity_period.type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform_display_name", resourceName, "platform_display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccSigningProfileDataSourceConfig_basic(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}

data "aws_signer_signing_profile" "test" {
  name = aws_signer_signing_profile.test.name
}`, profileName)
}
