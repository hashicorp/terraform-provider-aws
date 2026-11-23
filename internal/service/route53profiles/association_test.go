// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"
	profileName := "aws_route53profiles_profile.test"
	vpcName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					resource.TestCheckResourceAttrPair(resourceName, "profile_id", profileName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, vpcName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusMessage),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccRoute53ProfilesAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"
	var v awstypes.ProfileAssociation

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53ProfilesAssociation_priority(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_priority(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "200"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssociationConfig_priority(rName, 300),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ProfileStatusComplete)),
				),
			},
			{
				Config: testAccAssociationConfig_name(rName, rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "300"),
				),
			},
		},
	})
}

func TestAccRoute53ProfilesAssociation_nameUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_name(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccAssociationConfig_name(rName, rNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
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

func TestAccRoute53ProfilesAssociation_multipleProfiles(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association1, association2 awstypes.ProfileAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_route53profiles_association.test1"
	resourceName2 := "aws_route53profiles_association.test2"
	vpcName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53Profiles)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_multipleProfiles(rName, 100, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName1, &association1),
					testAccCheckAssociationExists(ctx, t, resourceName2, &association2),
					resource.TestCheckResourceAttrPair(resourceName1, names.AttrResourceID, vpcName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrResourceID, vpcName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPriority, "100"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPriority, "200"),
				),
			},
			{
				Config: testAccAssociationConfig_multipleProfiles(rName, 100, 300),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName1, &association1),
					testAccCheckAssociationExists(ctx, t, resourceName2, &association2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPriority, "100"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPriority, "300"),
				),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53ProfilesAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, t, resourceName, &association),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfroute53profiles.Route53ProfileAssocation, resourceName),
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

func testAccCheckAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_association" {
				continue
			}

			_, err := tfroute53profiles.FindAssociationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssociationExists(ctx context.Context, t *testing.T, name string, association *awstypes.ProfileAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).Route53ProfilesClient(ctx)
		resp, err := tfroute53profiles.FindAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, rs.Primary.ID, err)
		}

		*association = *resp

		return nil
	}
}

func testAccAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {
  name        = %[1]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id
}
`, rName)
}

func testAccAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {

  name        = %[1]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {
  name        = %[1]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAssociationConfig_priority(rName string, priority int) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {
  name        = %[1]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id
  priority    = %[2]d
}
`, rName, priority)
}

func testAccAssociationConfig_name(rName, associationName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {
  name        = %[2]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id
}
`, rName, associationName)
}

func testAccAssociationConfig_multipleProfiles(rName string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test1" {
  name = "%[1]s-1"
}

resource "aws_route53profiles_profile" "test2" {
  name = "%[1]s-2"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test1" {
  name        = "%[1]s-1"
  profile_id  = aws_route53profiles_profile.test1.id
  resource_id = aws_vpc.test.id
  priority    = %[2]d
}

resource "aws_route53profiles_association" "test2" {
  name        = "%[1]s-2"
  profile_id  = aws_route53profiles_profile.test2.id
  resource_id = aws_vpc.test.id
  priority    = %[3]d
}
`, rName, priority1, priority2)
}
