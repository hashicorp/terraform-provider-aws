package auditmanager_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerFramework_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var framework types.Framework
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.AuditManagerEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_sets.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_sets.0.controls.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "auditmanager", regexp.MustCompile(`assessmentFramework/+.`)),
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

func TestAccAuditManagerFramework_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var framework types.Framework
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.AuditManagerEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					acctest.CheckFrameworkResourceDisappears(acctest.Provider, tfauditmanager.ResourceFramework, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerFramework_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var framework types.Framework
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.AuditManagerEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccFrameworkConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFrameworkConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAuditManagerFramework_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var framework types.Framework
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_framework.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.AuditManagerEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFrameworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFrameworkConfig_optional(rName, "text1", "text1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "compliance_type", "text1"),
					resource.TestCheckResourceAttr(resourceName, "description", "text1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFrameworkConfig_optional(rName, "text2", "text2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFrameworkExists(ctx, resourceName, &framework),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "compliance_type", "text2"),
					resource.TestCheckResourceAttr(resourceName, "description", "text2"),
				),
			},
		},
	})
}

func testAccCheckFrameworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_framework" {
				continue
			}

			_, err := tfauditmanager.FindFrameworkByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.AuditManager, create.ErrActionCheckingDestroyed, tfauditmanager.ResNameFramework, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFrameworkExists(ctx context.Context, name string, framework *types.Framework) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFramework, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFramework, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient()
		resp, err := tfauditmanager.FindFrameworkByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameFramework, rs.Primary.ID, err)
		}

		*framework = *resp

		return nil
	}
}

func testAccFrameworkConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}
`, rName)
}

func testAccFrameworkConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccFrameworkConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}
`, rName))
}

func testAccFrameworkConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccFrameworkConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFrameworkConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccFrameworkConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccFrameworkConfig_optional(rName, complianceType, description string) string {
	return acctest.ConfigCompose(
		testAccFrameworkConfigBase(rName),
		fmt.Sprintf(`
resource "aws_auditmanager_framework" "test" {
  name = %[1]q

  compliance_type = %[2]q
  description     = %[3]q

  control_sets {
    name = %[1]q
    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}
`, rName, complianceType, description))
}
