// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxDatabase_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxdatabase finspace.GetKxDatabaseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccFinSpaceKxDatabase_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxdatabase finspace.GetKxDatabaseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDatabaseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxDatabase_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxdatabase finspace.GetKxDatabaseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDatabaseConfig_description(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 1"),
				),
			},
			{
				Config: testAccKxDatabaseConfig_description(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description 2"),
				),
			},
		},
	})
}

func TestAccFinSpaceKxDatabase_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxdatabase finspace.GetKxDatabaseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxDatabaseConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccKxDatabaseConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKxDatabaseConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxDatabaseExists(ctx, resourceName, &kxdatabase),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckKxDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_finspace_kx_database" {
				continue
			}

			input := &finspace.GetKxDatabaseInput{
				DatabaseName:  aws.String(rs.Primary.Attributes[names.AttrName]),
				EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
			}
			_, err := conn.GetKxDatabase(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.FinSpace, create.ErrActionCheckingDestroyed, tffinspace.ResNameKxDatabase, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKxDatabaseExists(ctx context.Context, name string, kxdatabase *finspace.GetKxDatabaseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDatabase, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDatabase, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)
		resp, err := conn.GetKxDatabase(ctx, &finspace.GetKxDatabaseInput{
			DatabaseName:  aws.String(rs.Primary.Attributes[names.AttrName]),
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
		})

		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxDatabase, rs.Primary.ID, err)
		}

		*kxdatabase = *resp

		return nil
	}
}

func testAccKxDatabaseConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccKxDatabaseConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccKxDatabaseConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
}
`, rName))
}

func testAccKxDatabaseConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccKxDatabaseConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
  description    = %[2]q
}
`, rName, description))
}

func testAccKxDatabaseConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccKxDatabaseConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccKxDatabaseConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccKxDatabaseConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
