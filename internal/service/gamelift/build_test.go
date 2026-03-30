// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftBuild_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Build
	resourceName := "aws_gamelift_build.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basic(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2016"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"storage_location"},
			},
			{
				Config: testAccBuildConfig_basic(rNameUpdated, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2016"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Build

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_gamelift_build.test"

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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basicTags1(rName, bucketName, key, roleArn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"storage_location"},
			},
			{
				Config: testAccBuildConfig_basicTags2(rName, bucketName, key, roleArn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBuildConfig_basicTags1(rName, bucketName, key, roleArn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.Build

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_gamelift_build.test"

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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basic(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceBuild(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceBuild(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBuildExists(ctx context.Context, t *testing.T, n string, v *awstypes.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		output, err := tfgamelift.FindBuildByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.ToString(output.BuildId) != rs.Primary.ID {
			return fmt.Errorf("GameLift Build not found")
		}

		*v = *output

		return nil
	}
}

func testAccCheckBuildDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_build" {
				continue
			}

			_, err := tfgamelift.FindBuildByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GameLift Build %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

	input := &gamelift.ListBuildsInput{}

	_, err := conn.ListBuilds(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBuildConfig_basic(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2016"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }
}
`, buildName, bucketName, key, roleArn)
}

func testAccBuildConfig_basicTags1(buildName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2016"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
  }
}
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1)
}

func testAccBuildConfig_basicTags2(buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2016"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2)
}
