// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediastore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmediastore "github.com/hashicorp/terraform-provider-aws/internal/service/mediastore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaStoreContainer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_store_container.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = strings.ReplaceAll(rName, "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
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

func TestAccMediaStoreContainer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_store_container.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = strings.ReplaceAll(rName, "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig_tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccContainerConfig_tags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccMediaStoreContainer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_store_container.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName = strings.ReplaceAll(rName, "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmediastore.ResourceContainer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContainerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_store_container" {
				continue
			}

			_, err := tfmediastore.FindContainerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("container (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckContainerExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreClient(ctx)

		_, err := tfmediastore.FindContainerByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("retrieving MediaStore Container (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreClient(ctx)

	input := &mediastore.ListContainersInput{}

	_, err := conn.ListContainers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccContainerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = %[1]q
}
`, rName)
}

func testAccContainerConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q

    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
