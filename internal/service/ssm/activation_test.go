// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSSMActivation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmActivation ssm.Activation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivationConfig_basic(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivationExists(ctx, resourceName, &ssmActivation),
					resource.TestCheckResourceAttrSet(resourceName, "activation_code"),
					acctest.CheckResourceAttrRFC3339(resourceName, "expiration_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccSSMActivation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmActivation ssm.Activation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivationConfig_tags(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivationExists(ctx, resourceName, &ssmActivation),
					resource.TestCheckResourceAttrSet(resourceName, "activation_code"),
					acctest.CheckResourceAttrRFC3339(resourceName, "expiration_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccSSMActivation_expirationDate(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmActivation ssm.Activation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	expirationDate := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivationConfig_expirationDate(rName, expirationDate, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivationExists(ctx, resourceName, &ssmActivation),
					resource.TestCheckResourceAttr(resourceName, "expiration_date", expirationDate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccSSMActivation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmActivation ssm.Activation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	roleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActivationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActivationConfig_basic(rName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivationExists(ctx, resourceName, &ssmActivation),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceActivation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckActivationExists(ctx context.Context, n string, v *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Activation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		output, err := tfssm.FindActivationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckActivationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_activation" {
				continue
			}

			_, err := tfssm.FindActivationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Activation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccActivationConfig_base(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}
`, roleName)
}

func testAccActivationConfig_basic(rName string, roleName string) string {
	return acctest.ConfigCompose(testAccActivationConfig_base(roleName), fmt.Sprintf(`
resource "aws_ssm_activation" "test" {
  name               = %[1]q
  description        = "Test"
  iam_role           = aws_iam_role.test.name
  registration_limit = "5"

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccActivationConfig_tags(rName string, roleName string) string {
	return acctest.ConfigCompose(testAccActivationConfig_base(roleName), fmt.Sprintf(`
resource "aws_ssm_activation" "test" {
  name               = %[1]q
  description        = "Test"
  iam_role           = aws_iam_role.test.name
  registration_limit = "5"

  depends_on = [aws_iam_role_policy_attachment.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccActivationConfig_expirationDate(rName string, expirationDate string, roleName string) string {
	return acctest.ConfigCompose(testAccActivationConfig_base(roleName), fmt.Sprintf(`
resource "aws_ssm_activation" "test" {
  name               = %[1]q
  description        = "Test"
  expiration_date    = "%[2]s"
  iam_role           = aws_iam_role.test.name
  registration_limit = "5"

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, expirationDate))
}
