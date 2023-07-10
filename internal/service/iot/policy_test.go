// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
)

func TestAccIoTPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iot.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy_basic(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_initialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccIoTPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iot.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy_basic(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_initialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy_basic(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_policy" {
				continue
			}

			// Try to find the Policy
			GetPolicyOpts := &iot.GetPolicyInput{
				PolicyName: aws.String(rs.Primary.Attributes["name"]),
			}

			resp, err := conn.GetPolicyWithContext(ctx, GetPolicyOpts)

			if err == nil {
				if resp.PolicyName != nil {
					return fmt.Errorf("IoT Policy still exists")
				}
			}

			if err != nil {
				if !tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
					return err
				}
			}
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *iot.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		resp, err := conn.GetPolicyWithContext(ctx, &iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPolicyConfig_initialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}
`, rName)
}
