package ssm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// This resource affects regional defaults, so it needs to be serialized
func TestAccSSMDefaultPatchBaseline_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                testAccSSMDefaultPatchBaseline_basic,
		"disappears":           testAccSSMDefaultPatchBaseline_disappears,
		"otherOperatingSystem": testAccSSMDefaultPatchBaseline_otherOperatingSystem,
		"patchBaselineARN":     testAccSSMDefaultPatchBaseline_patchBaselineARN,
		"systemDefault":        testAccSSMDefaultPatchBaseline_systemDefault,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccSSMDefaultPatchBaseline_basic(t *testing.T) {
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_disappears(t *testing.T) {
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(resourceName, &defaultpatchbaseline),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourceDefaultPatchBaseline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_patchBaselineARN(t *testing.T) {
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_patchBaselineARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_otherOperatingSystem(t *testing.T) {
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_operatingSystem(rName, types.OperatingSystemAmazonLinux2022),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_systemDefault(t *testing.T) {
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineDataSourceName := "data.aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_systemDefault(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", baselineDataSourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDefaultPatchBaselineDestroy(s *terraform.State) error {
	tfssm.SSMClientV2.Init(acctest.Provider.Meta().(*conns.AWSClient).Config, func(c aws.Config) *ssm.Client {
		return ssm.NewFromConfig(c)
	})
	conn := tfssm.SSMClientV2.Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_default_patch_baseline" {
			continue
		}

		defaultOSPatchBaseline, err := tfssm.FindDefaultDefaultPatchBaselineIDForOS(ctx, conn, types.OperatingSystem(rs.Primary.ID))
		if err != nil {
			return err
		}

		// If the resource has been deleted, the default patch baseline will be the AWS-provided patch baseline for the OS
		out, err := tfssm.FindDefaultPatchBaseline(ctx, conn, types.OperatingSystem(rs.Primary.ID))
		if tfresource.NotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		if aws.ToString(out.BaselineId) == defaultOSPatchBaseline {
			return nil
		}

		return create.Error(names.SSM, create.ErrActionCheckingDestroyed, tfssm.ResNameDefaultPatchBaseline, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckDefaultPatchBaselineExists(name string, defaultpatchbaseline *ssm.GetDefaultPatchBaselineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSM, create.ErrActionCheckingExistence, tfssm.ResNameDefaultPatchBaseline, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSM, create.ErrActionCheckingExistence, tfssm.ResNameDefaultPatchBaseline, name, errors.New("not set"))
		}

		tfssm.SSMClientV2.Init(acctest.Provider.Meta().(*conns.AWSClient).Config, func(c aws.Config) *ssm.Client {
			return ssm.NewFromConfig(c)
		})
		conn := tfssm.SSMClientV2.Client()
		ctx := context.Background()

		resp, err := tfssm.FindDefaultPatchBaseline(ctx, conn, types.OperatingSystem(rs.Primary.ID))
		if err != nil {
			return create.Error(names.SSM, create.ErrActionCheckingExistence, tfssm.ResNameDefaultPatchBaseline, rs.Primary.ID, err)
		}

		*defaultpatchbaseline = *resp

		return nil
	}
}

func testAccDefaultPatchBaselineImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["baseline_id"], nil
	}
}

func testAccDefaultPatchBaselineConfig_basic(rName string) string {
	return testAccDefaultPatchBaselineConfig_operatingSystem(rName, types.OperatingSystemWindows)
}

func testAccDefaultPatchBaselineConfig_operatingSystem(rName string, os types.OperatingSystem) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id = aws_ssm_patch_baseline.test.id
}

resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = %[2]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName, os)
}

func testAccDefaultPatchBaselineConfig_patchBaselineARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id = aws_ssm_patch_baseline.test.arn
}

resource "aws_ssm_patch_baseline" "test" {
  name = %[1]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName)
}

func testAccDefaultPatchBaselineConfig_systemDefault() string {
	return `
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id = data.aws_ssm_patch_baseline.test.id
}

data "aws_ssm_patch_baseline" "test" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
`
}
