// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubAppVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app_version.test"
	appResourceName := "aws_resiliencehub_app.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppVersionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppVersionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "app_arn", appResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "app_version"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrIdentifier),
					resource.TestCheckResourceAttrSet(resourceName, "app_template_body"),
					resource.TestCheckResourceAttr(resourceName, "source_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "terraform_source.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "version_name"),
					resource.TestCheckNoResourceAttr(resourceName, "import_strategy"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAppVersionImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "app_version",
				// These are create-time inputs that are not returned by the
				// published version, and identifier is not retrievable via
				// DescribeAppVersion.
				ImportStateVerifyIgnore: []string{"app_template_body", "source_arns", "terraform_source", "import_strategy", "version_name", names.AttrIdentifier},
			},
		},
	})
}

func TestAccResilienceHubAppVersion_versionName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app_version.test"
	var identifier1, identifier2 string

	// A new published version is created on any change, but it always belongs
	// to the same application.
	expectSameAppARN := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppVersionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppVersionExists(ctx, t, resourceName),
					testAccCheckAppVersionStoreIdentifier(resourceName, &identifier1),
					resource.TestCheckNoResourceAttr(resourceName, "version_name"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectSameAppARN.AddStateValue(resourceName, tfjsonpath.New("app_arn")),
				},
			},
			{
				Config: testAccAppVersionConfig_versionName(rName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppVersionExists(ctx, t, resourceName),
					testAccCheckAppVersionStoreIdentifier(resourceName, &identifier2),
					testAccCheckAppVersionIdentifierChanged(&identifier1, &identifier2),
					resource.TestCheckResourceAttr(resourceName, "version_name", "updated"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectSameAppARN.AddStateValue(resourceName, tfjsonpath.New("app_arn")),
				},
			},
		},
	})
}

func TestAccResilienceHubAppVersion_appTemplateBody(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app_version.test"
	var identifier1, identifier2 string

	expectSameAppARN := statecheck.CompareValue(compare.ValuesSame())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppVersionConfig_template(rName, "queue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppVersionExists(ctx, t, resourceName),
					testAccCheckAppVersionStoreIdentifier(resourceName, &identifier1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectSameAppARN.AddStateValue(resourceName, tfjsonpath.New("app_arn")),
				},
			},
			{
				Config: testAccAppVersionConfig_template(rName, "messaging"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppVersionExists(ctx, t, resourceName),
					testAccCheckAppVersionStoreIdentifier(resourceName, &identifier2),
					testAccCheckAppVersionIdentifierChanged(&identifier1, &identifier2),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					expectSameAppARN.AddStateValue(resourceName, tfjsonpath.New("app_arn")),
				},
			},
		},
	})
}

func testAccCheckAppVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehub_app_version" {
				continue
			}

			_, err := conn.DescribeAppVersion(ctx, &resiliencehub.DescribeAppVersionInput{
				AppArn:     aws.String(rs.Primary.Attributes["app_arn"]),
				AppVersion: aws.String(rs.Primary.Attributes["app_version"]),
			})
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ResilienceHub App Version %s/%s still exists", rs.Primary.Attributes["app_arn"], rs.Primary.Attributes["app_version"])
		}

		return nil
	}
}

func testAccCheckAppVersionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubClient(ctx)

		_, err := conn.DescribeAppVersion(ctx, &resiliencehub.DescribeAppVersionInput{
			AppArn:     aws.String(rs.Primary.Attributes["app_arn"]),
			AppVersion: aws.String(rs.Primary.Attributes["app_version"]),
		})

		return err
	}
}

func testAccCheckAppVersionStoreIdentifier(n string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		*v = rs.Primary.Attributes[names.AttrIdentifier]

		return nil
	}
}

func testAccCheckAppVersionIdentifierChanged(v1, v2 *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v1 == *v2 {
			return fmt.Errorf("expected a new application version to be published, but identifier did not change (%s)", *v1)
		}

		return nil
	}
}

func testAccAppVersionImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_arn"], rs.Primary.Attributes["app_version"]), nil
	}
}

func testAccAppVersionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    Resources = {
      Queue = {
        Type = "AWS::SQS::Queue"
      }
    }
  })
}

resource "aws_resiliencehub_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppVersionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAppVersionConfig_base(rName), `
resource "aws_resiliencehub_app_version" "test" {
  app_arn     = aws_resiliencehub_app.test.arn
  source_arns = [aws_cloudformation_stack.test.id]

  app_template_body = jsonencode({
    resources = [{
      logicalResourceId = {
        identifier       = "Queue"
        logicalStackName = aws_cloudformation_stack.test.name
      }
      type = "AWS::SQS::Queue"
      name = "Queue"
    }]
    appComponents = [{
      name          = "appcommon"
      type          = "AWS::ResilienceHub::AppCommonAppComponent"
      resourceNames = []
      }, {
      name          = "queue"
      type          = "AWS::ResilienceHub::QueueAppComponent"
      resourceNames = ["Queue"]
    }]
    excludedResources = {
      logicalResourceIds = []
    }
    version = 2
  })
}
`)
}

func testAccAppVersionConfig_versionName(rName, versionName string) string {
	return acctest.ConfigCompose(testAccAppVersionConfig_base(rName), fmt.Sprintf(`
resource "aws_resiliencehub_app_version" "test" {
  app_arn      = aws_resiliencehub_app.test.arn
  source_arns  = [aws_cloudformation_stack.test.id]
  version_name = %[1]q

  app_template_body = jsonencode({
    resources = [{
      logicalResourceId = {
        identifier       = "Queue"
        logicalStackName = aws_cloudformation_stack.test.name
      }
      type = "AWS::SQS::Queue"
      name = "Queue"
    }]
    appComponents = [{
      name          = "appcommon"
      type          = "AWS::ResilienceHub::AppCommonAppComponent"
      resourceNames = []
      }, {
      name          = "queue"
      type          = "AWS::ResilienceHub::QueueAppComponent"
      resourceNames = ["Queue"]
    }]
    excludedResources = {
      logicalResourceIds = []
    }
    version = 2
  })
}
`, versionName))
}

func testAccAppVersionConfig_template(rName, componentName string) string {
	return acctest.ConfigCompose(testAccAppVersionConfig_base(rName), fmt.Sprintf(`
resource "aws_resiliencehub_app_version" "test" {
  app_arn     = aws_resiliencehub_app.test.arn
  source_arns = [aws_cloudformation_stack.test.id]

  app_template_body = jsonencode({
    resources = [{
      logicalResourceId = {
        identifier       = "Queue"
        logicalStackName = aws_cloudformation_stack.test.name
      }
      type = "AWS::SQS::Queue"
      name = "Queue"
    }]
    appComponents = [{
      name          = "appcommon"
      type          = "AWS::ResilienceHub::AppCommonAppComponent"
      resourceNames = []
      }, {
      name          = %[1]q
      type          = "AWS::ResilienceHub::QueueAppComponent"
      resourceNames = ["Queue"]
    }]
    excludedResources = {
      logicalResourceIds = []
    }
    version = 2
  })
}
`, componentName))
}
