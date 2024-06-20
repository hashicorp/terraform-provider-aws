// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMDocument_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_format", "JSON"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ssm", fmt.Sprintf("document/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, "document_version"),
					resource.TestCheckResourceAttr(resourceName, "version_name", ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
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

func TestAccSSMDocument_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccSSMDocument_Target_type(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_basicTargetType(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_basicTargetType(rName, "/AWS::EC2::Instance"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "/AWS::EC2::Instance"),
				),
			},
		},
	})
}

func TestAccSSMDocument_versionName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_basicVersionName(rName, "release-1.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "version_name", "release-1.0.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_basicVersionName(rName, "release-1.0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "version_name", "release-1.0.1"),
				),
			},
		},
	})
}

func TestAccSSMDocument_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_20(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "2.0"),
					resource.TestCheckResourceAttr(resourceName, "document_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct1),
					resource.TestCheckOutput("default_version", acctest.Ct1),
					resource.TestCheckOutput("document_version", acctest.Ct1),
					resource.TestCheckOutput("hash", "1a200df3fefa0e7f8814829781d6295e616474945a239a956561876b4c820cde"),
					resource.TestCheckOutput("latest_version", acctest.Ct1),
					resource.TestCheckOutput("parameter_len", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_20Updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "latest_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_version", acctest.Ct2),
					resource.TestCheckOutput("default_version", acctest.Ct2),
					resource.TestCheckOutput("document_version", acctest.Ct2),
					resource.TestCheckOutput("hash", "214c51d87f98ae07b868a63cd866955578c1ef41c3ab8c36f80039dfd9565f53"),
					resource.TestCheckOutput("latest_version", acctest.Ct2),
					resource.TestCheckOutput("parameter_len", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccSSMDocument_Permission_public(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_publicPermission(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", "all"),
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

func TestAccSSMDocument_Permission_private(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"
	ids := "123456789012"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_privatePermission(rName, ids),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
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

func TestAccSSMDocument_Permission_batching(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"
	ids := "123456789012,123456789013,123456789014,123456789015,123456789016,123456789017,123456789018,123456789019,123456789020,123456789021,123456789022,123456789023,123456789024,123456789025,123456789026,123456789027,123456789028,123456789029,123456789030,123456789031,123456789032"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_privatePermission(rName, ids),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
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

func TestAccSSMDocument_Permission_change(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"
	idsInitial := "123456789012,123456789013"
	idsRemove := "123456789012"
	idsAdd := "123456789012,123456789014"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_privatePermission(rName, idsInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsInitial),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_privatePermission(rName, idsRemove),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsRemove),
				),
			},
			{
				Config: testAccDocumentConfig_privatePermission(rName, idsAdd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "permissions.type", "Share"),
					resource.TestCheckResourceAttr(resourceName, "permissions.account_ids", idsAdd),
				),
			},
		},
	})
}

func TestAccSSMDocument_params(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_param(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.0.name", "commands"),
					resource.TestCheckResourceAttr(resourceName, "parameter.0.type", "StringList"),
					resource.TestCheckResourceAttr(resourceName, "parameter.1.name", "workingDirectory"),
					resource.TestCheckResourceAttr(resourceName, "parameter.1.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.name", "executionTimeout"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2.type", "String"),
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

func TestAccSSMDocument_automation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_typeAutomation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Automation"),
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

func TestAccSSMDocument_package(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rInt1 := sdkacctest.RandInt()
	rInt2 := sdkacctest.RandInt()
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_typePackage(rName, rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Package"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"attachments_source"}, // This doesn't work because the API doesn't provide attachments info directly
			},
			{
				Config: testAccDocumentConfig_typePackage(rName, rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Package"),
				),
			},
		},
	})
}

func TestAccSSMDocument_SchemaVersion_1(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_schemaVersion1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_schemaVersion1Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schema_version", "1.0"),
				),
			},
		},
	})
}

func TestAccSSMDocument_session(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_typeSession(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_type", "Session"),
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

func TestAccSSMDocument_DocumentFormat_yaml(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"
	content1 := `
---
schemaVersion: '2.2'
description: Sample document
mainSteps:
- action: aws:runPowerShellScript
  name: runPowerShellScript
  inputs:
    runCommand:
      - hostname
`
	content2 := `
---
schemaVersion: '2.2'
description: Sample document
mainSteps:
- action: aws:runPowerShellScript
  name: runPowerShellScript
  inputs:
    runCommand:
      - Get-Process
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_formatYAML(rName, content1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content1+"\n"),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDocumentConfig_formatYAML(rName, content2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, content2+"\n"),
					resource.TestCheckResourceAttr(resourceName, "document_format", "YAML"),
				),
			},
		},
	})
}

func TestAccSSMDocument_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDocumentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocumentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceDocument(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDocumentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		_, err := tfssm.FindDocumentByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDocumentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_document" {
				continue
			}

			_, err := tfssm.FindDocumentByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Document %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

/*
Based on examples from here: https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/create-ssm-doc.html
*/

func testAccDocumentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}
`, rName)
}

func testAccDocumentConfig_basicTargetType(rName, typ string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"
  target_type   = %[2]q

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {},
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process"
        ]
      }
    }
  ]
}
DOC
}
`, rName, typ)
}

func testAccDocumentConfig_basicVersionName(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"
  version_name  = %[2]q

  content = <<DOC
    {
       "schemaVersion": "2.0",
       "description": "Sample version 2.0 document %[2]s",
       "parameters": {

       },
       "mainSteps": [
          {
             "action": "aws:runPowerShellScript",
             "name": "runPowerShellScript",
             "inputs": {
                "runCommand": [
                   "Get-Process"
                ]
             }
          }
       ]
    }
DOC
}
`, rName, version)
}

func testAccDocumentConfig_20(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {},
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process"
        ]
      }
    }
  ]
}
DOC
}

output "default_version" {
  value = aws_ssm_document.test.default_version
}

output "document_version" {
  value = aws_ssm_document.test.document_version
}

output "hash" {
  value = aws_ssm_document.test.hash
}

output "latest_version" {
  value = aws_ssm_document.test.latest_version
}

output "parameter_len" {
  value = length(aws_ssm_document.test.parameter)
}
`, rName)
}

func testAccDocumentConfig_20Updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "2.0",
  "description": "Sample version 2.0 document v2",
  "parameters": {
    "processOptions": {
      "type": "String",
      "default": "-Verbose",
      "description": "(Optional) Get-Process command options."
    }
  },
  "mainSteps": [
    {
      "action": "aws:runPowerShellScript",
      "name": "runPowerShellScript",
      "inputs": {
        "runCommand": [
          "Get-Process {{processOptions}}"
        ]
      }
    }
  ]
}
DOC
}

output "default_version" {
  value = aws_ssm_document.test.default_version
}

output "document_version" {
  value = aws_ssm_document.test.document_version
}

output "hash" {
  value = aws_ssm_document.test.hash
}

output "latest_version" {
  value = aws_ssm_document.test.latest_version
}

output "parameter_len" {
  value = length(aws_ssm_document.test.parameter)
}
`, rName)
}

func testAccDocumentConfig_publicPermission(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  permissions = {
    type        = "Share"
    account_ids = "all"
  }

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}
`, rName)
}

func testAccDocumentConfig_privatePermission(rName, ids string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  permissions = {
    type        = "Share"
    account_ids = %[2]q
  }

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC
}
`, rName, ids)
}

func testAccDocumentConfig_param(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Run a PowerShell script or specify the paths to scripts to run.",
  "parameters": {
    "commands": {
      "type": "StringList",
      "description": "(Required) Specify the commands to run or the paths to existing scripts on the instance.",
      "minItems": 1,
      "displayType": "textarea"
    },
    "workingDirectory": {
      "type": "String",
      "default": "",
      "description": "(Optional) The path to the working directory on your instance.",
      "maxChars": 4096
    },
    "executionTimeout": {
      "type": "String",
      "default": "3600",
      "description": "(Optional) The time in seconds for a command to be completed before it is considered to have failed. Default is 3600 (1 hour). Maximum is 28800 (8 hours).",
      "allowedPattern": "([1-9][0-9]{0,3})|(1[0-9]{1,4})|(2[0-7][0-9]{1,3})|(28[0-7][0-9]{1,2})|(28800)"
    }
  },
  "runtimeConfig": {
    "aws:runPowerShellScript": {
      "properties": [
        {
          "id": "0.aws:runPowerShellScript",
          "runCommand": "{{ commands }}",
          "workingDirectory": "{{ workingDirectory }}",
          "timeoutSeconds": "{{ executionTimeout }}"
        }
      ]
    }
  }
}
DOC
}
`, rName)
}

func testAccDocumentConfig_typeAutomation(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_iam_instance_profile" "ssm_profile" {
  name = %[1]q
  role = aws_iam_role.ssm_role.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "ssm_role" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC
}
`, rName))
}

func testAccDocumentConfig_typePackage(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-object-test-bucket-%[2]d"
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "test.zip"
  source       = "test-fixtures/ssm-doc-acc-test.zip"
  content_type = "binary/octet-stream"
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Package"

  attachments_source {
    key    = "SourceUrl"
    values = ["s3://${aws_s3_object.test.bucket}"]
  }

  content = <<DOC
{
  "description": "Systems Manager Package Document Test",
  "schemaVersion": "2.0",
  "version": "0.1",
  "assumeRole": "${aws_iam_role.test.arn}",
  "files": {
    "test.zip": {
      "checksums": {
        "sha256": "${filesha256("test-fixtures/ssm-doc-acc-test.zip")}"
      }
    }
  },
  "packages": {
    "amazon": {
      "_any": {
        "x86_64": {
          "file": "${aws_s3_object.test.key}"
        }
      }
    }
  }
}
DOC

  depends_on = [aws_iam_role_policy.test]
}
`, rName, rInt)
}

func testAccDocumentConfig_typeSession(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC
}
`, rName)
}

func testAccDocumentConfig_formatYAML(rName, content string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  document_format = "YAML"
  document_type   = "Command"
  name            = %[1]q

  content = <<DOC
%[2]s
DOC
}
`, rName, content)
}

func testAccDocumentConfig_schemaVersion1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC
}
`, rName)
}

func testAccDocumentConfig_schemaVersion1Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Session"

  content = <<DOC
{
  "schemaVersion": "1.0",
  "description": "Document to hold regional settings for Session Manager",
  "sessionType": "Standard_Stream",
  "inputs": {
    "s3BucketName": "test",
    "s3KeyPrefix": "test",
    "s3EncryptionEnabled": true,
    "cloudWatchLogGroupName": "/logs/sessions-updated",
    "cloudWatchEncryptionEnabled": false
  }
}
DOC
}
`, rName)
}
