// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccM2Application_basic_Content(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var application m2.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic_Content(rName, "bluage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "m2", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "current_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "definition.0.content"),
					resource.TestCheckNoResourceAttr(resourceName, "definition.0.s3_location"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "bluage"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrRoleARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccM2Application_basic_S3Location(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var application m2.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic_S3Location(rName, "bluage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "m2", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "current_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "definition.0.content"),
					resource.TestMatchResourceAttr(resourceName, "definition.0.s3_location", regexache.MustCompile(`s3://[-a-z0-9]+/definition.json`)),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "bluage"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrRoleARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("definition").AtSliceIndex(0).AtMapKey(names.AttrContent), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("definition").AtSliceIndex(0).AtMapKey("s3_location"), knownvalue.StringRegexp(regexache.MustCompile(`s3://[-a-z0-9]+/definition.json`))),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"definition"},
			},
		},
	})
}

func TestAccM2Application_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var application m2.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccApplicationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic_Content(rName, "bluage"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfm2.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccM2Application_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"
	var application m2.GetApplicationOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "m2", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "current_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "definition.0.content"),
					resource.TestCheckNoResourceAttr(resourceName, "definition.0.s3_location"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "testing"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "bluage"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccM2Application_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var application m2.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_versioned(rName, "bluage", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "current_version", acctest.Ct1),
				),
			},
			{
				Config: testAccApplicationConfig_versioned(rName, "bluage", 2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "current_version", acctest.Ct2),
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

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_application" {
				continue
			}

			_, err := tfm2.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Mainframe Modernization Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, n string, v *m2.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		output, err := tfm2.FindApplicationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

	input := &m2.ListApplicationsInput{}
	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccApplicationConfig_basic_Content(rName, engineType string) string {
	return testAccApplicationConfig_versioned(rName, engineType, 1, 1)
}

func testAccApplicationConfig_versioned(rName, engineType string, version, versions int) string {
	return fmt.Sprintf(`
resource "aws_m2_application" "test" {
  name        = %[1]q
  engine_type = %[2]q
  definition {
    content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = %[3]d })
  }

  depends_on = [aws_s3_object.test]
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  count = %[4]d

  bucket = aws_s3_bucket.test.id
  key    = "v${count.index + 1}/PlanetsDemo-v${count.index + 1}.zip"
  source = "test-fixtures/PlanetsDemo-v1.zip"
}
`, rName, engineType, version, versions)
}

func testAccApplicationConfig_basic_S3Location(rName, engineType string) string {
	return testAccApplicationConfig_S3Location_versioned(rName, engineType, 1, 1)
}

func testAccApplicationConfig_S3Location_versioned(rName, engineType string, version, versions int) string {
	return fmt.Sprintf(`
resource "aws_m2_application" "test" {
  name        = %[1]q
  engine_type = %[2]q
  definition {
    s3_location = "s3://${aws_s3_object.definition.bucket}/${aws_s3_object.definition.key}"
  }

  depends_on = [
    aws_s3_object.application,
    aws_s3_object.definition,
  ]
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "application" {
  count = %[4]d

  bucket = aws_s3_bucket.test.id
  key    = "v${count.index + 1}/PlanetsDemo-v${count.index + 1}.zip"
  source = "test-fixtures/PlanetsDemo-v1.zip"
}

resource "aws_s3_object" "definition" {
  bucket  = aws_s3_bucket.test.id
  key     = "definition.json"
  content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = %[3]d })
}
`, rName, engineType, version, versions)
}

func testAccApplicationConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_m2_application" "test" {
  name        = %[1]q
  engine_type = "bluage"
  description = "testing"
  kms_key_id  = aws_kms_key.test.arn
  role_arn    = aws_iam_role.test.arn
  definition {
    content = templatefile("test-fixtures/application-definition.json", { s3_bucket = aws_s3_bucket.test.id, version = "v1" })
  }

  depends_on = [aws_s3_object.test, aws_iam_role_policy.test]
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "v1/PlanetsDemo-v1.zip"
  source = "test-fixtures/PlanetsDemo-v1.zip"
}

resource "aws_kms_key" "test" {
  description = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "m2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:DescribeSecret",
          "secretsmanager:GetSecretValue",
          "kms:Decrypt",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
`, rName)
}
