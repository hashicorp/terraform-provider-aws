// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaLiveInputSecurityGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var inputSecurityGroup medialive.DescribeInputSecurityGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_input_security_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputSecurityGroupsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSecurityGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInputSecurityGroupConfig_basic(rName, "10.0.0.8/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSecurityGroupExists(ctx, t, resourceName, &inputSecurityGroup),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "inputSecurityGroup:{id}"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "whitelist_rules.*", map[string]string{
						"cidr": "10.0.0.8/32",
					}),
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

func TestAccMediaLiveInputSecurityGroup_updateCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var inputSecurityGroup medialive.DescribeInputSecurityGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_input_security_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputSecurityGroupsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSecurityGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInputSecurityGroupConfig_basic(rName, "10.0.0.8/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSecurityGroupExists(ctx, t, resourceName, &inputSecurityGroup),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "inputSecurityGroup:{id}"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "whitelist_rules.*", map[string]string{
						"cidr": "10.0.0.8/32",
					}),
				),
			},
			{
				Config: testAccInputSecurityGroupConfig_basic(rName, "10.2.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSecurityGroupExists(ctx, t, resourceName, &inputSecurityGroup),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "medialive", "inputSecurityGroup:{id}"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "whitelist_rules.*", map[string]string{
						"cidr": "10.2.0.0/16",
					}),
				),
			},
		},
	})
}

func TestAccMediaLiveInputSecurityGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var inputSecurityGroup medialive.DescribeInputSecurityGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_medialive_input_security_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputSecurityGroupsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputSecurityGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInputSecurityGroupConfig_basic(rName, "10.0.0.8/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputSecurityGroupExists(ctx, t, resourceName, &inputSecurityGroup),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmedialive.ResourceInputSecurityGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInputSecurityGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_medialive_input_security_group" {
				continue
			}

			_, err := tfmedialive.FindInputSecurityGroupByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameInputSecurityGroup, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckInputSecurityGroupExists(ctx context.Context, t *testing.T, name string, inputSecurityGroup *medialive.DescribeInputSecurityGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindInputSecurityGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, rs.Primary.ID, err)
		}

		*inputSecurityGroup = *resp

		return nil
	}
}

func testAccInputSecurityGroupsPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MediaLiveClient(ctx)

	input := &medialive.ListInputSecurityGroupsInput{}
	_, err := conn.ListInputSecurityGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInputSecurityGroupConfig_basic(rName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, cidr)
}
