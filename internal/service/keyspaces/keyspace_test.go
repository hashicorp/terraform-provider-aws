// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyspaces_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkeyspaces "github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheck(t *testing.T) {
	acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
}

func TestAccKeyspacesKeyspace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_keyspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyspaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyspaceExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cassandra", "/keyspace/"+rName+"/"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccKeyspacesKeyspace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_keyspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyspaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyspaceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkeyspaces.ResourceKeyspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKeyspacesKeyspace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf_acc_test_" + sdkacctest.RandString(20)
	resourceName := "aws_keyspaces_keyspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KeyspacesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyspaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyspaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyspaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyspaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKeyspaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyspaceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckKeyspaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_keyspaces_keyspace" {
				continue
			}

			_, err := tfkeyspaces.FindKeyspaceByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Keyspaces Keyspace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyspaceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Keyspaces Keyspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KeyspacesClient(ctx)

		_, err := tfkeyspaces.FindKeyspaceByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccKeyspaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q
}
`, rName)
}

func testAccKeyspaceConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccKeyspaceConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_keyspaces_keyspace" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
