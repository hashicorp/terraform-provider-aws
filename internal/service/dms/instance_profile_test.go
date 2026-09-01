// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSInstanceProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_instance_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, names.AttrName, regexache.MustCompile(`^ip-[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "subnet_group_identifier", "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCSecurityGroupIDs+".#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTagsAll+".%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dms", regexache.MustCompile(`instance-profile:.+$`)),
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

func TestAccDMSInstanceProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_instance_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceInstanceProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDMSInstanceProfile_full(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dms_instance_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_full(rName, "first instance profile", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first instance profile"),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_group_identifier", "aws_dms_replication_subnet_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCSecurityGroupIDs+".#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, names.AttrVPCSecurityGroupIDs+".*", "aws_security_group.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceProfileConfig_full(rName, "second instance profile", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second instance profile"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckInstanceProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_instance_profile" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

			_, err := tfdms.FindInstanceProfileByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Instance Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceProfileExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		ctx = conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		_, err := tfdms.FindInstanceProfileByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccInstanceProfileConfig_basic() string {
	return `
resource "aws_dms_instance_profile" "test" {}
`
}

func testAccInstanceProfileConfig_full(rName, description string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "testing"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_instance_profile" "test" {
  name                    = %[1]q
  description             = %[2]q
  network_type            = "IPV4"
  publicly_accessible     = %[3]t
  kms_key_arn             = aws_kms_key.test.arn
  subnet_group_identifier = aws_dms_replication_subnet_group.test.id
  vpc_security_group_ids  = [aws_security_group.test.id]
}
`, rName, description, publiclyAccessible))
}
