// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityMailFromAttributesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomain()
	mailFromDomain1 := domain.Subdomain("test1")

	rName := domain.String()
	resourceName := "aws_sesv2_email_identity_mail_from_attributes.test"
	dataSourceName := "data.aws_sesv2_email_identity_mail_from_attributes.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityMailFromAttributesDataSourceConfig_basic(rName, string(types.BehaviorOnMxFailureRejectMessage), mailFromDomain1.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityMailFromAttributesExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(resourceName, "email_identity", dataSourceName, "email_identity"),
					resource.TestCheckResourceAttrPair(resourceName, "behavior_on_mx_failure", dataSourceName, "behavior_on_mx_failure"),
					resource.TestCheckResourceAttrPair(resourceName, "mail_from_domain", dataSourceName, "mail_from_domain"),
				),
			},
		},
	})
}

func testAccEmailIdentityMailFromAttributesDataSourceConfig_basic(rName, behaviorOnMXFailure, mailFromDomain string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_mail_from_attributes" "test" {
  email_identity         = aws_sesv2_email_identity.test.email_identity
  behavior_on_mx_failure = %[2]q
  mail_from_domain       = %[3]q
}

data "aws_sesv2_email_identity_mail_from_attributes" "test" {
  email_identity = aws_sesv2_email_identity_mail_from_attributes.test.email_identity
}
`, rName, behaviorOnMXFailure, mailFromDomain)
}
