// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2OptOutList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var optOutList pinpointsmsvoicev2.DescribeOptOutListsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOptOutList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptOutListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptOutListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptOutListExists(ctx, resourceName, &optOutList),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccPinpointSMSVoiceV2OptOutList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var optOutList pinpointsmsvoicev2.DescribeOptOutListsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOptOutList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptOutListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptOutListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptOutListExists(ctx, resourceName, &optOutList),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpointsmsvoicev2.ResourceOptOutList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOptOutListDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_opt_out_list" {
				continue
			}

			input := &pinpointsmsvoicev2.DescribeOptOutListsInput{
				OptOutListNames: aws.StringSlice([]string{rs.Primary.ID}),
			}

			_, err := conn.DescribeOptOutListsWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			return fmt.Errorf("expected PinpointSMSVoiceV2 OptOutList to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOptOutListExists(ctx context.Context, n string, v *pinpointsmsvoicev2.DescribeOptOutListsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

		resp, err := conn.DescribeOptOutListsWithContext(ctx, &pinpointsmsvoicev2.DescribeOptOutListsInput{
			OptOutListNames: aws.StringSlice([]string{rs.Primary.ID}),
		})

		if err != nil {
			return fmt.Errorf("error describing PinpointSMSVoiceV2 OptOutList: %s", err.Error())
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckOptOutList(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

	input := &pinpointsmsvoicev2.DescribeOptOutListsInput{}

	_, err := conn.DescribeOptOutListsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOptOutListConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName)
}
