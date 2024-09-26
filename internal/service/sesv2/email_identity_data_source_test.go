// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_email_identity.test"
	dataSourceName := "data.aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "email_identity", dataSourceName, "email_identity"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.#", dataSourceName, "dkim_signing_attributes.#"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.current_signing_key_length", dataSourceName, "dkim_signing_attributes.0.current_signing_key_length"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.last_key_generation_timestamp", dataSourceName, "dkim_signing_attributes.0.last_key_generation_timestamp"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.next_signing_key_length", dataSourceName, "dkim_signing_attributes.0.next_signing_key_length"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.signing_attributes_origin", dataSourceName, "dkim_signing_attributes.0.signing_attributes_origin"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.status", dataSourceName, "dkim_signing_attributes.0.status"),
					resource.TestCheckResourceAttrPair(resourceName, "dkim_signing_attributes.0.tokens.#", dataSourceName, "dkim_signing_attributes.0.tokens.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_type", dataSourceName, "identity_type"),
					resource.TestCheckResourceAttrPair(resourceName, "verified_for_sending_status", dataSourceName, "verified_for_sending_status"),
				),
			},
		},
	})
}

func testAccEmailIdentityDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

data "aws_sesv2_email_identity" "test" {
  email_identity = aws_sesv2_email_identity.test.email_identity
}
`, rName)
}
