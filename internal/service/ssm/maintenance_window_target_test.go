// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMMaintenanceWindowTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, string(awstypes.MaintenanceWindowResourceTypeInstance)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_noNameOrDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_noNameOrDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_validation(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "Bäd Name!@#$%^", "good description"),
				ExpectError: regexache.MustCompile(`Only alphanumeric characters, hyphens, dots & underscores allowed`),
			},
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "goodname", "bd"),
				ExpectError: regexache.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
			{
				Config:      testAccMaintenanceWindowTargetConfig_basic2(name, "goodname", "This description is tooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long"),
				ExpectError: regexache.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_update(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This resource is for test purpose only"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMaintenanceWindowTargetConfig_basicUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "owner_information", "something"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Updated"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "new-value"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This resource is for test purpose only - updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_resourceGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basicResourceGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "resource-groups:ResourceTypeFilters"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "AWS::EC2::Instance"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "resource-groups:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "resource-group-name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, string(awstypes.MaintenanceWindowResourceTypeResourceGroup)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceMaintenanceWindowTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMMaintenanceWindowTarget_Disappears_window(t *testing.T) {
	ctx := acctest.Context(t)
	var maint awstypes.MaintenanceWindowTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_maintenance_window_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMaintenanceWindowTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceWindowTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMaintenanceWindowTargetExists(ctx, resourceName, &maint),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceMaintenanceWindow(), "aws_ssm_maintenance_window.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMaintenanceWindowTargetExists(ctx context.Context, n string, v *awstypes.MaintenanceWindowTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		output, err := tfssm.FindMaintenanceWindowTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["window_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMaintenanceWindowTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_maintenance_window_target" {
				continue
			}

			_, err := tfssm.FindMaintenanceWindowTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["window_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Activation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMaintenanceWindowTargetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basicResourceGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "RESOURCE_GROUP"

  targets {
    key    = "resource-groups:ResourceTypeFilters"
    values = ["AWS::EC2::Instance"]
  }

  targets {
    key    = "resource-groups:Name"
    values = ["resource-group-name"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_noNameOrDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[1]q
  description       = "This resource is for test purpose only - updated"
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName)
}

func testAccMaintenanceWindowTargetConfig_basic2(rName, tName, tDesc string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[2]q
  description       = %[3]q
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName, tName, tDesc)
}

func testAccMaintenanceWindowTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}
