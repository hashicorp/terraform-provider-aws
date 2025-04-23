// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2DefaultCreditSpecification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var defaultcreditspecification awstypes.InstanceFamilyCreditSpecification
	_ = sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_instance_default_credit_specification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultCreditSpecificationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultCreditSpecificationExists(ctx, resourceName, &defaultcreditspecification),
					resource.TestCheckResourceAttr(resourceName, "cpu_credits", "unlimited"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "t2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func testAccCheckDefaultCreditSpecificationExists(ctx context.Context, name string, defaultcreditspecification *awstypes.InstanceFamilyCreditSpecification) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameDefaultCreditSpecification, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameDefaultCreditSpecification, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindDefaultCreditSpecificationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameDefaultCreditSpecification, rs.Primary.ID, err)
		}

		*defaultcreditspecification = *resp

		return nil
	}
}
func testAccDefaultCreditSpecificationConfig_basic() string {
	return `
resource "aws_ec2_instance_default_credit_specification" "test" {
  cpu_credits     = "standard"
  instance_family = "t2"
}
`
}
