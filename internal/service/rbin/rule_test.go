// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rbin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rbin"
	"github.com/aws/aws-sdk-go-v2/service/rbin/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrbin "github.com/hashicorp/terraform-provider-aws/internal/service/rbin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRBinRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule rbin.GetRuleOutput
	description := "my test description"
	resourceType := "EBS_SNAPSHOT"
	resourceName := "aws_rbin_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rbin.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, rbin.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic1(description, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, resourceType),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_period.*", map[string]string{
						"retention_period_value": acctest.Ct10,
						"retention_period_unit":  "DAYS",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_tags.*", map[string]string{
						"resource_tag_key":   "some_tag1",
						"resource_tag_value": "some_value1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_basic2(description, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, resourceType),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "retention_period.*", map[string]string{
						"retention_period_value": acctest.Ct10,
						"retention_period_unit":  "DAYS",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_tags.*", map[string]string{
						"resource_tag_key":   "some_tag3",
						"resource_tag_value": "some_value3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "resource_tags.*", map[string]string{
						"resource_tag_key":   "some_tag4",
						"resource_tag_value": "some_value4",
					}),
				),
			},
		},
	})
}

func TestAccRBinRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rbinrule rbin.GetRuleOutput
	description := "my test description"
	resourceType := "EBS_SNAPSHOT"
	resourceName := "aws_rbin_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rbin.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, rbin.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic1(description, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rbinrule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrbin.ResourceRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRBinRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rule rbin.GetRuleOutput
	resourceType := "EBS_SNAPSHOT"
	resourceName := "aws_rbin_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RBin)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RBin),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfigTags1(resourceType, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
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
				Config: testAccRuleConfigTags2(resourceType, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleConfigTags1(resourceType, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccRBinRule_lock_config(t *testing.T) {
	ctx := acctest.Context(t)
	var rule rbin.GetRuleOutput
	resourceType := "EBS_SNAPSHOT"
	resourceName := "aws_rbin_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rbin.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, rbin.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_lockConfig(resourceType, "DAYS", "7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lock_configuration.0.unlock_delay.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "lock_configuration.0.unlock_delay.0.unlock_delay_unit", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "lock_configuration.0.unlock_delay.0.unlock_delay_value", "7"),
				),
			},
		},
	})
}

func testAccCheckRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RBinClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rbin_rule" {
				continue
			}

			_, err := conn.GetRule(ctx, &rbin.GetRuleInput{
				Identifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.RBin, create.ErrActionCheckingDestroyed, tfrbin.ResNameRule, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRuleExists(ctx context.Context, name string, rbinrule *rbin.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RBin, create.ErrActionCheckingExistence, tfrbin.ResNameRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.RBin, create.ErrActionCheckingExistence, tfrbin.ResNameRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RBinClient(ctx)

		resp, err := conn.GetRule(ctx, &rbin.GetRuleInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.RBin, create.ErrActionCheckingExistence, tfrbin.ResNameRule, rs.Primary.ID, err)
		}

		*rbinrule = *resp

		return nil
	}
}

func testAccRuleConfig_basic1(description, resourceType string) string {
	return fmt.Sprintf(`
resource "aws_rbin_rule" "test" {
  description   = %[1]q
  resource_type = %[2]q

  resource_tags {
    resource_tag_key   = "some_tag1"
    resource_tag_value = "some_value1"
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }
}
`, description, resourceType)
}

func testAccRuleConfig_basic2(description, resourceType string) string {
	return fmt.Sprintf(`
resource "aws_rbin_rule" "test" {
  description   = %[1]q
  resource_type = %[2]q

  resource_tags {
    resource_tag_key   = "some_tag3"
    resource_tag_value = "some_value3"
  }

  resource_tags {
    resource_tag_key   = "some_tag4"
    resource_tag_value = "some_value4"
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }
}
`, description, resourceType)
}

func testAccRuleConfig_lockConfig(resourceType, delay_unit1, delay_value1 string) string {
	return fmt.Sprintf(`
resource "aws_rbin_rule" "test" {
  resource_type = %[1]q

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }

  lock_configuration {
    unlock_delay {
      unlock_delay_unit  = %[2]q
      unlock_delay_value = %[3]q
    }
  }
}
`, resourceType, delay_unit1, delay_value1)
}

func testAccRuleConfigTags1(resourceType, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_rbin_rule" "test" {
  resource_type = %[1]q

  resource_tags {
    resource_tag_key   = "some_tag"
    resource_tag_value = ""
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, resourceType, tag1Key, tag1Value)
}

func testAccRuleConfigTags2(resourceType, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_rbin_rule" "test" {
  resource_type = %[1]q

  resource_tags {
    resource_tag_key   = "some_tag"
    resource_tag_value = ""
  }

  retention_period {
    retention_period_value = 10
    retention_period_unit  = "DAYS"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, resourceType, tag1Key, tag1Value, tag2Key, tag2Value)
}
