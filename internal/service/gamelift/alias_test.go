// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	uAliasName := fmt.Sprintf("tf_acc_alias_upd_%s", rString)
	uDescription := fmt.Sprintf("tf test updated description %s", rString)
	uMessage := fmt.Sprintf("tf test updated message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", message),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, aliasName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckAliasExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", uMessage),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, uAliasName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, uDescription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccGameLiftAlias_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.Alias

	resourceName := "aws_gamelift_alias.test"
	aliasName := sdkacctest.RandomWithPrefix("tf-acc-alias")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basicTags1(aliasName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
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
				Config: testAccAliasConfig_basicTags2(aliasName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAliasConfig_basicTags1(aliasName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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

	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	fleetName := fmt.Sprintf("tf_acc_fleet_%s", rString)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_allFields(aliasName, description,
					fleetName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", acctest.Ct1),
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
	var conf gamelift.Alias

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasDisappears(ctx, &conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAliasDisappears(ctx context.Context, res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		input := &gamelift.DeleteAliasInput{AliasId: res.AliasId}

		_, err := conn.DeleteAliasWithContext(ctx, input)

		return err
	}
}

func testAccCheckAliasExists(ctx context.Context, n string, res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Alias ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		out, err := conn.DescribeAliasWithContext(ctx, &gamelift.DescribeAliasInput{
			AliasId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		a := out.Alias

		if *a.AliasId != rs.Primary.ID {
			return fmt.Errorf("GameLift Alias not found")
		}

		*res = *a

		return nil
	}
}

func testAccCheckAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_alias" {
				continue
			}

			_, err := conn.DescribeAliasWithContext(ctx, &gamelift.DescribeAliasInput{
				AliasId: aws.String(rs.Primary.ID),
			})
			if err == nil {
				return fmt.Errorf("GameLift Alias still exists")
			}

			if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
				return nil
			}

			return err
		}

		return nil
	}
}

func testAccAliasConfig_basic(aliasName, description, message string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    message = "%s"
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

func testAccAliasConfig_allFields(aliasName, description,
	fleetName, launchPath, params, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    fleet_id = aws_gamelift_fleet.test.id
    type     = "SIMPLE"
  }
}
%s
`, aliasName, description,
		testAccFleetConfig_basic(fleetName, launchPath, params, bucketName, key, roleArn))
}
