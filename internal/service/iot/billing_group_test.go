// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTBillingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.DescribeBillingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", regexache.MustCompile(fmt.Sprintf("billinggroup/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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

func TestAccIoTBillingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.DescribeBillingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceBillingGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTBillingGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.DescribeBillingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBillingGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBillingGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIoTBillingGroup_properties(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.DescribeBillingGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_billing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBillingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBillingGroupConfig_properties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBillingGroupConfig_propertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
				),
			},
		},
	})
}

func testAccCheckBillingGroupExists(ctx context.Context, n string, v *iot.DescribeBillingGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		output, err := tfiot.FindBillingGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBillingGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_billing_group" {
				continue
			}

			_, err := tfiot.FindBillingGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Billing Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBillingGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccBillingGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccBillingGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccBillingGroupConfig_properties(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  properties {
    description = "test description 1"
  }
}
`, rName)
}

func testAccBillingGroupConfig_propertiesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_billing_group" "test" {
  name = %[1]q

  properties {
    description = "test description 2"
  }
}
`, rName)
}
