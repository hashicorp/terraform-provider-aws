// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSInstanceProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var instanceProfile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_instance_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName, &instanceProfile),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTags+".%", "0"),
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
	var instanceProfile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_instance_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, t, resourceName, &instanceProfile),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdms.ResourceInstanceProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSInstanceProfile_full(t *testing.T) {
	ctx := acctest.Context(t)
	var instanceProfile awstypes.InstanceProfile
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
					testAccCheckInstanceProfileExists(ctx, t, resourceName, &instanceProfile),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first instance profile"),
					resource.TestCheckResourceAttr(resourceName, "network_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_group_identifier", "aws_dms_replication_subnet_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCSecurityGroupIDs+".#", "1"),
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
					testAccCheckInstanceProfileExists(ctx, t, resourceName, &instanceProfile),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second instance profile"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, "false"),
				),
			},
		},
	})
}

func testAccCheckInstanceProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_instance_profile" {
				continue
			}

			_, err := tfdms.FindInstanceProfileByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.DMS, create.ErrActionCheckingDestroyed, tfdms.ResNameInstanceProfile, rs.Primary.ID, err)
			}

			return create.Error(names.DMS, create.ErrActionCheckingDestroyed, tfdms.ResNameInstanceProfile, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckInstanceProfileExists(ctx context.Context, t *testing.T, name string, v *awstypes.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, tfdms.ResNameInstanceProfile, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, tfdms.ResNameInstanceProfile, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DMSClient(ctx)

		out, err := tfdms.FindInstanceProfileByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.DMS, create.ErrActionCheckingExistence, tfdms.ResNameInstanceProfile, rs.Primary.ID, err)
		}

		*v = *out

		return nil
	}
}

func testAccInstanceProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_instance_profile" "test" {
  name = %[1]q
}
`, rName)
}

func testAccInstanceProfileConfig_full(rName, description string, publiclyAccessible bool) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
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
  subnet_group_identifier = aws_dms_replication_subnet_group.test.id
  vpc_security_group_ids  = [aws_security_group.test.id]
}
`, rName, description, publiclyAccessible))
}
