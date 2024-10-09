// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorResourceGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.ResourceGroup
	resourceName := "aws_inspector_resource_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "inspector", regexache.MustCompile(`resourcegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "foo"),
				),
			},
			{
				Config: testAccResourceGroupConfig_modified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceGroupExists(ctx, resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "inspector", regexache.MustCompile(`resourcegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "bar"),
					testAccCheckResourceGroupRecreated(&v1, &v2),
				),
			},
		},
	})
}

func testAccCheckResourceGroupExists(ctx context.Context, name string, rg *awstypes.ResourceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := conn.DescribeResourceGroups(ctx, &inspector.DescribeResourceGroupsInput{
			ResourceGroupArns: []string{rs.Primary.ID},
		})
		if err != nil {
			return err
		}
		if len(output.ResourceGroups) == 0 {
			return fmt.Errorf("No matching Inspector Classic Resource Groups")
		}

		*rg = output.ResourceGroups[0]

		return nil
	}
}

func testAccCheckResourceGroupRecreated(v1, v2 *awstypes.ResourceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v2.CreatedAt.Equal(*v1.CreatedAt) {
			return fmt.Errorf("Inspector Classic Resource Group not recreated when changing tags")
		}

		return nil
	}
}

var testAccResourceGroupConfig_basic = `
resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "foo"
  }
}
`

var testAccResourceGroupConfig_modified = `
resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "bar"
  }
}
`
