// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSOptionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`og:.+`)),
					resource.TestCheckResourceAttr(resourceName, "engine_name", "mysql"),
					resource.TestCheckResourceAttr(resourceName, "major_engine_version", "8.0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "option_group_description", "Managed by Terraform"),
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

func TestAccRDSOptionGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceOptionGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSOptionGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
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

func TestAccRDSOptionGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
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

func TestAccRDSOptionGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var optionGroup1, optionGroup2, optionGroup3 rds.OptionGroup
	resourceName := "aws_db_option_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
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
				Config: testAccOptionGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOptionGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSOptionGroup_timeoutBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_timeoutBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
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

func TestAccRDSOptionGroup_optionGroupDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var optionGroup1 rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option_group_description", "description1"),
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

func TestAccRDSOptionGroup_basicDestroyWithInstance(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_destroy(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDSOptionGroup_Option_optionSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_optionSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*.option_settings.*", map[string]string{
						names.AttrValue: "UTC",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore option since our current logic skips "unconfigured" default option settings
				// Even with Config set, ImportState TestStep does not "see" the configuration to check against
				ImportStateVerifyIgnore: []string{"option"},
			},
			{
				Config: testAccOptionGroupConfig_optionSettingsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*.option_settings.*", map[string]string{
						names.AttrValue: "US/Pacific",
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

func TestAccRDSOptionGroup_OptionOptionSettings_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_optionSettingsIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					testAccCheckOptionGroupOptionSettingsIAMRole(&v),
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

func TestAccRDSOptionGroup_sqlServerOptionsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_sqlServerEEOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOptionGroupConfig_sqlServerEEOptionsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSOptionGroup_oracleOptionsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_oracleEEOptionSettings(rName, "13.2.0.0.v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					testAccCheckOptionGroupOptionVersionAttribute(&v, "13.2.0.0.v2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore option since API responds with **** instead of password
				ImportStateVerifyIgnore: []string{"option"},
			},
			{
				Config: testAccOptionGroupConfig_oracleEEOptionSettings(rName, "13.3.0.0.v2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					testAccCheckOptionGroupOptionVersionAttribute(&v, "13.3.0.0.v2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/1876
func TestAccRDSOptionGroup_OptionOptionSettings_multipleNonDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var optionGroup1, optionGroup2 rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_settingsMultiple(rName, "example1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOptionGroupConfig_settingsMultiple(rName, "example1,example2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccRDSOptionGroup_multipleOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v rds.OptionGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_option_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_multipleOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct2),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7114
func TestAccRDSOptionGroup_Tags_withOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var optionGroup1, optionGroup2, optionGroup3 rds.OptionGroup
	resourceName := "aws_db_option_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_tagsOption1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
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
				Config: testAccOptionGroupConfig_tagsOption2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup2),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOptionGroupConfig_tagsOption1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup3),
					resource.TestCheckResourceAttr(resourceName, "option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/21367
func TestAccRDSOptionGroup_badDiffs(t *testing.T) {
	ctx := acctest.Context(t)
	var optionGroup1 rds.OptionGroup
	resourceName := "aws_db_option_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOptionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOptionGroupConfig_badDiffs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						names.AttrPort: "3872",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						"option_name": "SQLT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						"option_name": "S3_INTEGRATION",
					}),
				),
			},
			{
				Config:   testAccOptionGroupConfig_badDiffs1(rName),
				PlanOnly: true,
			},
			{
				Config: testAccOptionGroupConfig_badDiffs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptionGroupExists(ctx, resourceName, &optionGroup1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						names.AttrPort: "3873",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						"option_name":     "SQLT",
						names.AttrVersion: "2018-07-25.v1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "option.*", map[string]string{
						"option_name":     "S3_INTEGRATION",
						names.AttrVersion: "1.0",
					}),
				),
			},
		},
	})
}

func testAccCheckOptionGroupOptionSettingsIAMRole(optionGroup *rds.OptionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if optionGroup == nil {
			return errors.New("Option Group does not exist")
		}
		if len(optionGroup.Options) == 0 {
			return errors.New("Option Group does not have any options")
		}
		if len(optionGroup.Options[0].OptionSettings) == 0 {
			return errors.New("Option Group does not have any option settings")
		}

		settingName := aws.StringValue(optionGroup.Options[0].OptionSettings[0].Name)
		if settingName != "IAM_ROLE_ARN" {
			return fmt.Errorf("Expected option setting IAM_ROLE_ARN and received %s", settingName)
		}

		settingValue := aws.StringValue(optionGroup.Options[0].OptionSettings[0].Value)
		iamArnRegExp := regexache.MustCompile(fmt.Sprintf(`^arn:%s:iam::\d{12}:role/.+`, acctest.Partition()))
		if !iamArnRegExp.MatchString(settingValue) {
			return fmt.Errorf("Expected option setting to be a valid IAM role but received %s", settingValue)
		}
		return nil
	}
}

func testAccCheckOptionGroupOptionVersionAttribute(optionGroup *rds.OptionGroup, optionVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if optionGroup == nil {
			return errors.New("Option Group does not exist")
		}
		if len(optionGroup.Options) == 0 {
			return errors.New("Option Group does not have any options")
		}
		foundOptionVersion := aws.StringValue(optionGroup.Options[0].OptionVersion)
		if foundOptionVersion != optionVersion {
			return fmt.Errorf("Expected option version %q and received %q", optionVersion, foundOptionVersion)
		}
		return nil
	}
}

func testAccCheckOptionGroupExists(ctx context.Context, n string, v *rds.OptionGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindOptionGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOptionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_option_group" {
				continue
			}

			_, err := tfrds.FindOptionGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Option Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOptionGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  name                 = %[1]q
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
}
`, rName)
}

func testAccOptionGroupConfig_nameGenerated() string {
	return `
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
}
`
}

func testAccOptionGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  name_prefix              = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
}
`, namePrefix)
}

func testAccOptionGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                 = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccOptionGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                 = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccOptionGroupConfig_timeoutBlock(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)

  timeouts {
    delete = "10m"
  }
}
`, rName)
}

func testAccOptionGroupConfig_destroy(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = [%[1]s]
}

resource "aws_db_instance" "test" {
  allocated_storage = 10
  engine            = data.aws_rds_orderable_db_instance.test.engine
  engine_version    = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class    = data.aws_rds_orderable_db_instance.test.instance_class
  identifier        = %[2]q
  password          = "avoid-plaintext-passwords"
  username          = "tfacctest"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"

  backup_retention_period = 0
  skip_final_snapshot     = true

  option_group_name = aws_db_option_group.test.name
}

resource "aws_db_option_group" "test" {
  name                     = %[2]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
}
`, mainInstanceClasses, rName)
}

func testAccOptionGroupConfig_optionSettings(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "UTC"
    }
  }
}
`, rName)
}

func testAccOptionGroupConfig_optionSettingsIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_rds_engine_version" "default" {
  engine = "sqlserver-ex"
}

data "aws_iam_policy_document" "rds_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "sql_server_backup" {
  name               = "rds-backup-%[1]s"
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role.json
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "SQLSERVER_BACKUP_RESTORE"

    option_settings {
      name  = "IAM_ROLE_ARN"
      value = aws_iam_role.sql_server_backup.arn
    }
  }
}
`, rName)
}

func testAccOptionGroupConfig_optionSettingsUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "US/Pacific"
    }
  }
}
`, rName)
}

func testAccOptionGroupConfig_sqlServerEEOptions(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "sqlserver-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
}
`, rName)
}

func testAccOptionGroupConfig_sqlServerEEOptionsUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "sqlserver-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "TDE"
  }
}
`, rName)
}

func testAccOptionGroupConfig_oracleEEOptionSettings(rName, optionVersion string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_security_group" "foo" {
  name = %[1]q
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform issue 748"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "OEM_AGENT"
    port        = "3872"
    version     = %[2]q

    vpc_security_group_memberships = [aws_security_group.foo.id]

    option_settings {
      name  = "OMS_PORT"
      value = "4903"
    }

    option_settings {
      name  = "OMS_HOST"
      value = "oem.host.value"
    }

    option_settings {
      name  = "AGENT_REGISTRATION_PASSWORD"
      value = "password"
    }
  }
}
`, rName, optionVersion)
}

func testAccOptionGroupConfig_multipleOptions(rName string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Test option group for terraform"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "SPATIAL"
  }

  option {
    option_name = "STATSPACK"
  }
}
`, rName)
}

func testAccOptionGroupConfig_description(rName, optionGroupDescription string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                     = %[1]q
  option_group_description = %[2]q
}
`, rName, optionGroupDescription)
}

func testAccOptionGroupConfig_settingsMultiple(rName, value string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                 = %[1]q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_EXCL_USERS"
      value = %[2]q
    }

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "15"
    }

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATE_SIZE"
      value = "52428800"
    }
  }
}
`, rName, value)
}

func testAccOptionGroupConfig_tagsOption1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                 = %[1]q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "0"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccOptionGroupConfig_tagsOption2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

resource "aws_db_option_group" "test" {
  engine_name          = data.aws_rds_engine_version.default.engine
  major_engine_version = regex("^\\d+\\.\\d+", data.aws_rds_engine_version.default.version)
  name                 = %[1]q

  option {
    option_name = "MARIADB_AUDIT_PLUGIN"

    option_settings {
      name  = "SERVER_AUDIT_FILE_ROTATIONS"
      value = "0"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccOptionGroupConfig_badDiffs1(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Option Group for Numagove"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "S3_INTEGRATION"
  }

  option {
    option_name = "SQLT"
    option_settings {
      name  = "LICENSE_PACK"
      value = "T"
    }
  }

  option {
    option_name                    = "OEM_AGENT"
    version                        = "13.5.0.0.v1"
    port                           = 3872
    vpc_security_group_memberships = [aws_security_group.test.id]

    option_settings {
      name  = "AGENT_REGISTRATION_PASSWORD"
      value = "TESTPASSWORDBGY"
    }
    option_settings {
      name  = "MINIMUM_TLS_VERSION"
      value = "TLSv1.2"
    }
    option_settings {
      name  = "TLS_CIPHER_SUITE"
      value = "TLS_RSA_WITH_AES_128_CBC_SHA"
    }
    option_settings {
      name  = "OMS_HOST"
      value = "BGY-TEST"
    }
    option_settings {
      name  = "OMS_PORT"
      value = "1159"
    }
  }
}
`, rName)
}

func testAccOptionGroupConfig_badDiffs2(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

data "aws_rds_engine_version" "default" {
  engine = "oracle-ee"
}

resource "aws_db_option_group" "test" {
  name                     = %[1]q
  option_group_description = "Option Group for Numagove"
  engine_name              = data.aws_rds_engine_version.default.engine
  major_engine_version     = regex("^\\d+", data.aws_rds_engine_version.default.version)

  option {
    option_name = "S3_INTEGRATION"
    version     = "1.0"
  }

  option {
    option_name = "SQLT"
    option_settings {
      name  = "LICENSE_PACK"
      value = "T"
    }
    version = "2018-07-25.v1"
  }

  option {
    option_name                    = "OEM_AGENT"
    version                        = "13.5.0.0.v1"
    port                           = 3873
    vpc_security_group_memberships = [aws_security_group.test.id]

    option_settings {
      name  = "AGENT_REGISTRATION_PASSWORD"
      value = "TESTPASSWORDBGY"
    }
    option_settings {
      name  = "MINIMUM_TLS_VERSION"
      value = "TLSv1.2"
    }
    option_settings {
      name  = "TLS_CIPHER_SUITE"
      value = "TLS_RSA_WITH_AES_128_CBC_SHA"
    }
    option_settings {
      name  = "OMS_HOST"
      value = "BGY-TEST"
    }
    option_settings {
      name  = "OMS_PORT"
      value = "1159"
    }
  }
}
`, rName)
}
