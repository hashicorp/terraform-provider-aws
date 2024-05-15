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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftBuild_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.Build
	resourceName := "aws_gamelift_build.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basic(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
					testAccCheckBuildExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.Build

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_build.test"

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basicTags1(rName, bucketName, key, roleArn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckBuildExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBuildConfig_basicTags1(rName, bucketName, key, roleArn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.Build

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_build.test"

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBuildDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBuildConfig_basic(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgamelift.ResourceBuild(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgamelift.ResourceBuild(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBuildExists(ctx context.Context, n string, res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Build ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		build, err := tfgamelift.FindBuildByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if aws.StringValue(build.BuildId) != rs.Primary.ID {
			return fmt.Errorf("GameLift Build not found")
		}

		*res = *build

		return nil
	}
}

func testAccCheckBuildDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_build" {
				continue
			}

			build, err := tfgamelift.FindBuildByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if build != nil {
				return fmt.Errorf("GameLift Build (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

	input := &gamelift.ListBuildsInput{}

	_, err := conn.ListBuildsWithContext(ctx, input)

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
  name             = "%s"
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = "%s"
    key      = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}

func testAccBuildConfig_basicTags1(buildName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

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
  operating_system = "WINDOWS_2012"

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
