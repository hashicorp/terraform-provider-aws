// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package swf_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfswf "github.com/hashicorp/terraform-provider-aws/internal/service/swf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheckDomainTestingEnabled(t *testing.T) {
	if os.Getenv("SWF_DOMAIN_TESTING_ENABLED") == "" {
		t.Skip(
			"Environment variable SWF_DOMAIN_TESTING_ENABLED is not set. " +
				"SWF limits domains per region and the API does not support " +
				"deletions. Set the environment variable to any value to enable.")
	}
}

func TestAccSWFDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "swf", regexache.MustCompile(`/domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "workflow_execution_retention_period_in_days", acctest.Ct1),
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

func TestAccSWFDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfswf.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSWFDomain_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccSWFDomain_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccSWFDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
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
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSWFDomain_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SWFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SWFClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_swf_domain" {
				continue
			}

			// Retrying as Read after Delete is not always consistent.
			_, err := tfresource.RetryUntilNotFound(ctx, 2*time.Minute, func() (interface{}, error) {
				return tfswf.FindDomainByName(ctx, conn, rs.Primary.ID)
			})

			return err
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SWF Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SWFClient(ctx)

		_, err := tfswf.FindDomainByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, rName)
}

func testAccDomainConfig_nameGenerated() string {
	return `
resource "aws_swf_domain" "test" {
  workflow_execution_retention_period_in_days = 1
}
`
}

func testAccDomainConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name_prefix                                 = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, namePrefix)
}

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  description                                 = %[2]q
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, rName, description)
}
