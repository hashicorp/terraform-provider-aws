// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccSSMParameter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("parameter/%s", name)),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "data_type", "text"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_updateValue(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "String", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basic(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_description(name, "description", "String", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_description(name, "updated description", "String", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "test"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_tier(t *testing.T) {
	ctx := acctest.Context(t)
	var parameter1, parameter2, parameter3 ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter1),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter2),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter3),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
		},
	})
}

func TestAccSSMParameter_Tier_intelligentTieringToStandard(t *testing.T) {
	ctx := acctest.Context(t)
	var parameter ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Tier_intelligentTieringToAdvanced(t *testing.T) {
	ctx := acctest.Context(t)
	var parameter1, parameter2 ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter1),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter1),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
			{
				// Intelligent-Tiering will not downgrade an existing parameter to Standard
				Config: testAccParameterConfig_tier(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter2),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Tier_intelligentTieringOnCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var parameter ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	value := sdkacctest.RandString(5000) // Maximum size for Standard tier is 4 KB

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_tierWithValue(rName, ssm.ParameterTierIntelligentTiering, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Tier_intelligentTieringOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var parameter ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	standardSizedValue := sdkacctest.RandString(10)
	advancedSizedValue := sdkacctest.RandString(5000)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_tierWithValue(rName, ssm.ParameterTierIntelligentTiering, standardSizedValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				Config: testAccParameterConfig_tierWithValue(rName, ssm.ParameterTierIntelligentTiering, advancedSizedValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
		},
	})
}

func TestAccSSMParameter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceParameter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMParameter_Overwrite_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{

				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

					input := &ssm.PutParameterInput{
						Name:  aws.String(fmt.Sprintf("%s-%s", "test_parameter", name)),
						Type:  aws.String(ssm.ParameterTypeString),
						Value: aws.String("This value is set using the SDK"),
					}

					_, err := conn.PutParameterWithContext(ctx, input)
					if err != nil {
						t.Fatalf("creating SSM Parameter: (%s):, %s", name, err)
					}
				},
				Config: testAccParameterConfig_basicOverwrite(name, "String", "This value is set using Terraform"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "overwrite", "true"),
				),
			},
			{
				Config: testAccParameterConfig_basicOverwrite(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
					resource.TestCheckResourceAttr(resourceName, "overwrite", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basicOverwrite(name, "String", "test3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "test3"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "version", "4"),
					resource.TestCheckResourceAttr(resourceName, "overwrite", "true"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12213
func TestAccSSMParameter_Overwrite_cascade(t *testing.T) {
	ctx := acctest.Context(t)
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_cascadeOverwrite(name, "test1"),
			},
			{
				Config: testAccParameterConfig_cascadeOverwrite(name, "test2"),
			},
			{
				Config:             testAccParameterConfig_cascadeOverwrite(name, "test2"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18550
func TestAccSSMParameter_Overwrite_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_overwriteTags1(rName, true, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18550
func TestAccSSMParameter_Overwrite_noOverwriteTags(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_overwriteTags1(rName, false, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18550
func TestAccSSMParameter_Overwrite_updateToTags(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basicTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
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
				Config: testAccParameterConfig_overwriteTags1(rName, true, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
		},
	})
}
func TestAccSSMParameter_Overwrite_removeAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		CheckDestroy: testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.67.0",
					},
				},
				Config: testAccParameterConfig_overwriteRemove_Setup(rName, "String", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "overwrite", "true"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccParameterConfig_overwriteRemove_Remove(rName, "String", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "overwrite", "false"),
				),
			},
		},
	})
}

func TestAccSSMParameter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basicTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basicTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccParameterConfig_basicTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSSMParameter_updateType(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "SecureString", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basic(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
		},
	})
}

func TestAccSSMParameter_Overwrite_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basicOverwrite(name, "String", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basicOverwriteNoDescription(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccSSMParameter_changeNameForcesNew(t *testing.T) {
	ctx := acctest.Context(t)
	var beforeParam, afterParam ssm.Parameter
	before := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	after := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(before, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &beforeParam),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_basic(after, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &afterParam),
					testAccCheckParameterRecreated(t, &beforeParam, &afterParam),
				),
			},
		},
	})
}

func TestAccSSMParameter_fullPath(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("/path/%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("parameter%s", name)),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Secure_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_basic(name, "SecureString", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Secure_insecure(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_insecure(rName, "String", "notsecret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "insecure_value", "notsecret"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
			{
				Config: testAccParameterConfig_insecure(rName, "String", "newvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "insecure_value", "newvalue"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
			{
				Config:             testAccParameterConfig_insecure(rName, "String", "diff"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:      testAccParameterConfig_insecure(rName, "SecureString", "notsecret"),
				ExpectError: regexp.MustCompile("invalid configuration"),
			},
		},
	})
}

func TestAccSSMParameter_Secure_insecureChangeSecure(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_insecure(rName, "String", "notsecret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "insecure_value", "notsecret"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
			{
				Config: testAccParameterConfig_secure(rName, "newvalue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "newvalue"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
				),
			},
			{
				Config: testAccParameterConfig_insecure(rName, "String", "atlantis"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "insecure_value", "atlantis"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
		},
	})
}

func TestAccSSMParameter_DataType_ec2Image(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_dataTypeEC2Image(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "data_type", "aws:ec2:image"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_DataType_ssmIntegration(t *testing.T) {
	ctx := //nosemgrep:ci.ssm-in-func-name
		acctest.Context(t)
	var param ssm.Parameter
	webhookName := sdkacctest.RandString(16)
	rName := fmt.Sprintf("/d9d01087-4a3f-49e0-b0b4-d568d7826553/ssm/integrations/webhook/%s", webhookName)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_dataTypeSSMIntegration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "data_type", "aws:ssm:integration"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Secure_key(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	randString := sdkacctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_secureKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/"+randString),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccSSMParameter_Secure_keyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var param ssm.Parameter
	randString := sdkacctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterConfig_secure(name, "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterConfig_secureKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(ctx, resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/"+randString),
				),
			},
		},
	})
}

func testAccCheckParameterRecreated(t *testing.T,
	before, after *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Name == *after.Name {
			t.Fatalf("Expected change of SSM Param Names, but both were %v", *before.Name)
		}
		return nil
	}
}

func testAccCheckParameterExists(ctx context.Context, n string, param *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Parameter ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
			WithDecryption: aws.Bool(true),
		}

		resp, err := conn.GetParametersWithContext(ctx, paramInput)
		if err != nil {
			return err
		}

		if len(resp.Parameters) == 0 {
			return fmt.Errorf("Expected AWS SSM Parameter to be created, but wasn't found")
		}

		*param = *resp.Parameters[0]

		return nil
	}
}

func testAccCheckParameterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_parameter" {
				continue
			}

			paramInput := &ssm.GetParametersInput{
				Names: []*string{
					aws.String(rs.Primary.Attributes["name"]),
				},
			}

			resp, err := conn.GetParametersWithContext(ctx, paramInput)

			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading SSM Parameter (%s): %w", rs.Primary.ID, err)
			}

			if resp == nil || len(resp.Parameters) == 0 {
				continue
			}

			return fmt.Errorf("Expected AWS SSM Parameter to be gone, but was still found")
		}

		return nil
	}
}

func testAccParameterConfig_basic(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
}
`, rName, pType, value)
}

func testAccParameterConfig_description(rName, description, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = %[1]q
  description = %[2]q
  type        = %[3]q
  value       = %[4]q
}
`, rName, description, pType, value)
}

func testAccParameterConfig_insecure(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name           = %[1]q
  type           = %[2]q
  insecure_value = %[3]q
}
`, rName, pType, value)
}

func testAccParameterConfig_tier(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  tier  = %[2]q
  type  = "String"
  value = "test2"
}
`, rName, tier)
}

func testAccParameterConfig_tierWithValue(rName, tier, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  tier  = %[2]q
  type  = "String"
  value = %[3]q
}
`, rName, tier, value)
}

func testAccParameterConfig_dataTypeEC2Image(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = %[1]q
  data_type = "aws:ec2:image"
  type      = "String"
  value     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
}
`, rName))
}

func testAccParameterConfig_dataTypeSSMIntegration(rName string) string { // nosemgrep:ci.ssm-in-func-name
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = %[1]q
  data_type = "aws:ssm:integration"
  type      = "SecureString"
  value     = "{\"description\": \"My first webhook integration for Automation.\", \"url\": \"https://example.com\"}"
}
`, rName))
}

func testAccParameterConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccParameterConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccParameterConfig_basicOverwrite(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "%[2]s"
  value       = "%[3]s"
  overwrite   = true
}
`, rName, pType, value)
}

func testAccParameterConfig_basicOverwriteNoDescription(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = "test_parameter-%[1]s"
  type      = "%[2]s"
  value     = "%[3]s"
  overwrite = true
}
`, rName, pType, value)
}

func testAccParameterConfig_overwriteTags1(rName string, overwrite bool, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = %[1]q
  overwrite = %[2]t
  type      = "String"
  value     = %[1]q
  tags = {
    %[3]q = %[4]q
  }
}
`, rName, overwrite, tagKey1, tagValue1)
}

func testAccParameterConfig_cascadeOverwrite(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test_upstream" {
  name      = "test_parameter_upstream-%[1]s"
  type      = "String"
  value     = "%[2]s"
  overwrite = true
}

resource "aws_ssm_parameter" "test_downstream" {
  name      = "test_parameter_downstream-%[1]s"
  type      = "String"
  value     = aws_ssm_parameter.test_upstream.version
  overwrite = true
}
`, rName, value)
}

func testAccParameterConfig_overwriteRemove_Setup(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "%[2]s"
  value       = "%[3]s"
  overwrite   = true
}
`, rName, pType, value)
}

func testAccParameterConfig_overwriteRemove_Remove(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "%[2]s"
  value       = "%[3]s"
}
`, rName, pType, value)
}

func testAccParameterConfig_secure(rName string, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_secure_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "SecureString"
  value       = "%[2]s"
}
`, rName, value)
}

func testAccParameterConfig_secureKey(rName string, value string, keyAlias string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_secure_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "SecureString"
  value       = "%[2]s"
  key_id      = "alias/%[3]s"
  depends_on  = [aws_kms_alias.test_alias]
}

resource "aws_kms_key" "test_key" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test_alias" {
  name          = "alias/%[3]s"
  target_key_id = aws_kms_key.test_key.id
}
`, rName, value, keyAlias)
}

func TestParameterShouldUpdate(t *testing.T) {
	t.Parallel()

	data := tfssm.ResourceParameter().TestResourceData()
	failure := false

	if !tfssm.ShouldUpdateParameter(data) {
		t.Logf("Existing resources should be overwritten if the values don't match!")
		failure = true
	}

	data.MarkNewResource()
	if tfssm.ShouldUpdateParameter(data) {
		t.Logf("New resources must never be overwritten, this will overwrite parameters created outside of the system")
		failure = true
	}

	data = tfssm.ResourceParameter().TestResourceData()
	data.Set("overwrite", true)
	if !tfssm.ShouldUpdateParameter(data) {
		t.Logf("Resources should always be overwritten if the user requests it")
		failure = true
	}

	data.Set("overwrite", false)
	if tfssm.ShouldUpdateParameter(data) {
		t.Logf("Resources should never be overwritten if the user requests it")
		failure = true
	}
	if failure {
		t.Fail()
	}
}
