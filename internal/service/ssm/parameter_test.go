package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccSSMParameter_basic(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("parameter/%s", name)),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "tier", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "data_type", "text"),
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
	var parameter1, parameter2, parameter3 ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter1),
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
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter2),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter3),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
		},
	})
}

func TestAccSSMParameter_Tier_intelligentTieringToStandard(t *testing.T) {
	var parameter ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter),
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
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierStandard),
				),
			},
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter),
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
	var parameter1, parameter2 ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter1),
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
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierAdvanced),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter1),
					resource.TestCheckResourceAttr(resourceName, "tier", ssm.ParameterTierAdvanced),
				),
			},
			{
				Config: testAccParameterTierConfig(rName, ssm.ParameterTierIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &parameter2),
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

func TestAccSSMParameter_disappears(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceParameter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMParameter_overwrite(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicOverwriteConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterBasicOverwriteConfig(name, "String", "test3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "test3"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12213
func TestAccSSMParameter_overwriteCascade(t *testing.T) {
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterCascadeOverwriteConfig(name, "test1"),
			},
			{
				Config: testAccParameterCascadeOverwriteConfig(name, "test2"),
			},
			{
				Config:             testAccParameterCascadeOverwriteConfig(name, "test2"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18550
func TestAccSSMParameter_overwriteWithTags(t *testing.T) {
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterOverwriteWithTags1Config(rName, true, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
func TestAccSSMParameter_noOverwriteWithTags(t *testing.T) {
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterOverwriteWithTags1Config(rName, false, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
func TestAccSSMParameter_updateToOverwriteWithTags(t *testing.T) {
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
				Config: testAccParameterOverwriteWithTags1Config(rName, true, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
		},
	})
}

func TestAccSSMParameter_tags(t *testing.T) {
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
				Config: testAccParameterBasicTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccParameterBasicTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSSMParameter_updateType(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(name, "SecureString", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
		},
	})
}

func TestAccSSMParameter_updateDescription(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicOverwriteConfig(name, "String", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterBasicOverwriteWithoutDescriptionConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccSSMParameter_changeNameForcesNew(t *testing.T) {
	var beforeParam, afterParam ssm.Parameter
	before := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	after := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(before, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &beforeParam),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccParameterBasicConfig(after, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &afterParam),
					testAccCheckParameterRecreated(t, &beforeParam, &afterParam),
				),
			},
		},
	})
}

func TestAccSSMParameter_fullPath(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("/path/%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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

func TestAccSSMParameter_secure(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), sdkacctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterBasicConfig(name, "SecureString", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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

func TestAccSSMParameter_DataType_ec2Image(t *testing.T) {
	var param ssm.Parameter
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterDataTypeEC2ImageConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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

func TestAccSSMParameter_secureWithKey(t *testing.T) {
	var param ssm.Parameter
	randString := sdkacctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.secret_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterSecureWithKeyConfig(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
	var param ssm.Parameter
	randString := sdkacctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.secret_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterSecureConfig(name, "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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
				Config: testAccParameterSecureWithKeyConfig(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterExists(resourceName, &param),
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

func testAccCheckParameterExists(n string, param *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Parameter ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
			WithDecryption: aws.Bool(true),
		}

		resp, err := conn.GetParameters(paramInput)
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

func testAccCheckParameterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_parameter" {
			continue
		}

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
		}

		resp, err := conn.GetParameters(paramInput)

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

func testAccParameterBasicConfig(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
}
`, rName, pType, value)
}

func testAccParameterTierConfig(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  tier  = %[2]q
  type  = "String"
  value = "test2"
}
`, rName, tier)
}

func testAccParameterDataTypeEC2ImageConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = %[1]q
  data_type = "aws:ec2:image"
  type      = "String"
  value     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
}
`, rName))
}

func testAccParameterBasicTags1Config(rName, tagKey1, tagValue1 string) string {
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

func testAccParameterBasicTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccParameterBasicOverwriteConfig(rName, pType, value string) string {
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

func testAccParameterBasicOverwriteWithoutDescriptionConfig(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = "test_parameter-%[1]s"
  type      = "%[2]s"
  value     = "%[3]s"
  overwrite = true
}
`, rName, pType, value)
}

func testAccParameterOverwriteWithTags1Config(rName string, overwrite bool, tagKey1, tagValue1 string) string {
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

func testAccParameterCascadeOverwriteConfig(rName, value string) string {
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

func testAccParameterSecureConfig(rName string, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_test" {
  name        = "test_secure_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "SecureString"
  value       = "%[2]s"
}
`, rName, value)
}

func testAccParameterSecureWithKeyConfig(rName string, value string, keyAlias string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_test" {
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
