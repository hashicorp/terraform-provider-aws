// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/medialive"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaLiveInput_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "input_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UDP_PUSH"),
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

func TestAccMediaLiveInput_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttrSet(resourceName, "input_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UDP_PUSH"),
				),
			},
			{
				Config: testAccInputConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrSet(resourceName, "input_class"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UDP_PUSH"),
				),
			},
		},
	})
}

func TestAccMediaLiveInput_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccInputConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInputConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMediaLiveInput_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var input medialive.DescribeInputOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_medialive_input.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaLiveEndpointID)
			testAccInputsPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInputDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInputConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInputExists(ctx, resourceName, &input),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmedialive.ResourceInput(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInputDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_medialive_input" {
				continue
			}

			_, err := tfmedialive.FindInputByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameInput, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccCheckInputExists(ctx context.Context, name string, input *medialive.DescribeInputOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

		resp, err := tfmedialive.FindInputByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInput, rs.Primary.ID, err)
		}

		*input = *resp

		return nil
	}
}

func testAccInputsPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveClient(ctx)

	input := &medialive.ListInputsInput{}
	_, err := conn.ListInputs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInputBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_medialive_input_security_group" "test" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccInputConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInputBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInputConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccInputBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccInputConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccInputBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_medialive_input" "test" {
  name                  = %[1]q
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
