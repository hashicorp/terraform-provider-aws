// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccessanalyzer "github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAnalyzer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "analyzer_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "access-analyzer", fmt.Sprintf("analyzer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.TypeAccount)),
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

func testAccAnalyzer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfaccessanalyzer.ResourceAnalyzer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAnalyzer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
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
				Config: testAccAnalyzerConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAnalyzerConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAnalyzer_Type_Organization(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_typeOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "type", string(types.TypeOrganization)),
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

func testAccCheckAnalyzerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accessanalyzer_analyzer" {
				continue
			}

			_, err := tfaccessanalyzer.FindAnalyzerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Access Analyzer Analyzer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAnalyzerExists(ctx context.Context, n string, v *types.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Access Analyzer Analyzer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		output, err := tfaccessanalyzer.FindAnalyzerByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAnalyzerConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}
`, rName)
}

func testAccAnalyzerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAnalyzerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAnalyzerConfig_typeOrganization(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["access-analyzer.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_accessanalyzer_analyzer" "test" {
  depends_on = [aws_organizations_organization.test]

  analyzer_name = %[1]q
  type          = "ORGANIZATION"
}
`, rName)
}
