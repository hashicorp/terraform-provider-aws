// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	uAliasName := fmt.Sprintf("tf_acc_alias_upd_%s", rString)
	uDescription := fmt.Sprintf("tf test updated description %s", rString)
	uMessage := fmt.Sprintf("tf test updated message %s", rString)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", message),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, aliasName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig_basic(uAliasName, uDescription, uMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", uMessage),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, uAliasName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, uDescription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccGameLiftAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Alias

	resourceName := "aws_gamelift_alias.test"
	aliasName := acctest.RandomWithPrefix(t, "tf-acc-alias")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basicTags1(aliasName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
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
				Config: testAccAliasConfig_basicTags2(aliasName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAliasConfig_basicTags1(aliasName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftAlias_fleetRouting(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.Alias

	rString := sdkacctest.RandString(8)

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	fleetName := fmt.Sprintf("tf_acc_fleet_%s", rString)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if retry.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_allFields(aliasName, description, fleetName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_strategy.0.fleet_id"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "SIMPLE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, aliasName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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

func TestAccGameLiftAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAliasExists(ctx context.Context, t *testing.T, n string, v *awstypes.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		output, err := tfgamelift.FindAliasByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_alias" {
				continue
			}

			_, err := tfgamelift.FindAliasByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GameLift Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAliasConfig_basic(aliasName, description, message string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = %[2]q

  routing_strategy {
    message = %[3]q
    type    = "TERMINAL"
  }
}
`, aliasName, description, message)
}

func testAccAliasConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAliasConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAliasConfig_allFields(aliasName, description, fleetName, launchPath, params, bucketName, key, roleArn string) string {
	return acctest.ConfigCompose(testAccFleetConfig_basic(fleetName, launchPath, params, bucketName, key, roleArn), fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = %[2]q

  routing_strategy {
    fleet_id = aws_gamelift_fleet.test.id
    type     = "SIMPLE"
  }
}
`, aliasName, description))
}
