// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2InstanceEventWindowAssociation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccEC2InstanceEventWindowAssociation_instanceIDs,
		acctest.CtDisappears: testAccEC2InstanceEventWindowAssociation_disappears,
		"instanceTags":       testAccEC2InstanceEventWindowAssociation_instanceTags,
		"mutuallyExclusive":  testAccEC2InstanceEventWindowAssociation_mutuallyExclusive,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEC2InstanceEventWindowAssociation_instanceIDs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_event_window_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowAssociationConfig_instanceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "association_target.0.instance_ids.#", "1"),
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

func testAccEC2InstanceEventWindowAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_event_window_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowAssociationConfig_instanceIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, tfec2.ResourceInstanceEventWindowAssociation, resourceName, func(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
						windowID, ok := is.Attributes["instance_event_window_id"]
						if !ok {
							return errors.New(`Identifying attribute "instance_event_window_id" not defined`)
						}
						state.SetAttribute(ctx, path.Root("instance_event_window_id"), windowID)

						countStr, ok := is.Attributes["association_target.0.instance_ids.#"]
						if !ok {
							return errors.New(`Identifying attribute "association_target.0.instance_ids.#" not defined`)
						}
						count, err := strconv.Atoi(countStr)
						if err != nil {
							return err
						}

						instanceIDValues := make([]attr.Value, count)
						for i := range count {
							instanceIDValues[i] = types.StringValue(is.Attributes[fmt.Sprintf("association_target.0.instance_ids.%d", i)])
						}

						state.SetAttribute(ctx, path.Root("association_target").AtListIndex(0).AtName("instance_ids"), fwtypes.NewListValueOfMust[types.String](ctx, instanceIDValues))
						state.SetAttribute(ctx, path.Root("association_target").AtListIndex(0).AtName("dedicated_host_ids"), fwtypes.NewListValueOfNull[types.String](ctx))
						state.SetAttribute(ctx, path.Root("association_target").AtListIndex(0).AtName("instance_tags"), fwtypes.NewMapValueOfNull[types.String](ctx))

						return nil
					}),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccEC2InstanceEventWindowAssociation_instanceTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_instance_event_window_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceEventWindowAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceEventWindowAssociationConfig_instanceTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceEventWindowAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "association_target.0.instance_tags.%", "1"),
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

func testAccEC2InstanceEventWindowAssociation_mutuallyExclusive(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceEventWindowAssociationConfig_both(rName),
				ExpectError: regexache.MustCompile(`(?s)instance_ids.*instance_tags|instance_tags.*instance_ids`),
			},
		},
	})
}

func testAccCheckInstanceEventWindowAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindInstanceEventWindowByID(ctx, conn, rs.Primary.Attributes["instance_event_window_id"])
		return err
	}
}

func testAccCheckInstanceEventWindowAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_instance_event_window_association" {
				continue
			}

			out, err := tfec2.FindInstanceEventWindowByID(ctx, conn, rs.Primary.Attributes["instance_event_window_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			target := out.AssociationTarget
			if target != nil && (len(target.InstanceIds) > 0 || len(target.DedicatedHostIds) > 0 || len(target.Tags) > 0) {
				return fmt.Errorf("EC2 Instance Event Window Association %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccInstanceEventWindowAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_instance_event_window" "test" {
  name = %[1]q

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}
`, rName))
}

func testAccInstanceEventWindowAssociationConfig_instanceIDs(rName string) string {
	return acctest.ConfigCompose(testAccInstanceEventWindowAssociationConfig_base(rName), `
resource "aws_ec2_instance_event_window_association" "test" {
  instance_event_window_id = aws_ec2_instance_event_window.test.id

  association_target {
    instance_ids = [aws_instance.test.id]
  }
}
`)
}

func testAccInstanceEventWindowAssociationConfig_instanceTags(rName string) string {
	return acctest.ConfigCompose(testAccInstanceEventWindowAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_instance_event_window_association" "test" {
  instance_event_window_id = aws_ec2_instance_event_window.test.id

  association_target {
    instance_tags = {
      Name = %[1]q
    }
  }
}
`, rName))
}

func testAccInstanceEventWindowAssociationConfig_both(rName string) string {
	return acctest.ConfigCompose(testAccInstanceEventWindowAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_instance_event_window_association" "test" {
  instance_event_window_id = aws_ec2_instance_event_window.test.id

  association_target {
    instance_ids = [aws_instance.test.id]
    instance_tags = {
      Name = %[1]q
    }
  }
}
`, rName))
}
