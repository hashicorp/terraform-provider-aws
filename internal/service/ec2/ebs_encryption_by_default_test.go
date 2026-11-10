// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSEncryptionByDefault_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			acctest.CtBasic: testAccEBSEncryptionByDefault_basic,
			"importUpgrade": testAccEBSEncryptionByDefault_importUpgrade,
		},
		"DataSource": {
			acctest.CtBasic: testAccEBSEncryptionByDefaultDataSource_basic,
		},
		"List": {
			acctest.CtBasic:   testAccEC2EBSEncryptionByDefault_List_basic,
			"includeResource": testAccEC2EBSEncryptionByDefault_List_includeResource,
			"regionOverride":  testAccEC2EBSEncryptionByDefault_List_regionOverride,
		},
		"Identity": {
			"serial": testAccEC2EBSEBSEncryptionByDefault_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccEBSEncryptionByDefault_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ebs_encryption_by_default.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEncryptionByDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSEncryptionByDefaultConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSEncryptionByDefault(ctx, t, resourceName, false),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSEncryptionByDefaultConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSEncryptionByDefault(ctx, t, resourceName, true),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckEncryptionByDefaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		input := ec2.GetEbsEncryptionByDefaultInput{}
		response, err := conn.GetEbsEncryptionByDefault(ctx, &input)
		if err != nil {
			return err
		}

		if aws.ToBool(response.EbsEncryptionByDefault) != false {
			return fmt.Errorf("EBS encryption by default not disabled on resource removal")
		}

		return nil
	}
}

func testAccCheckEBSEncryptionByDefault(ctx context.Context, t *testing.T, n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		input := ec2.GetEbsEncryptionByDefaultInput{}
		response, err := conn.GetEbsEncryptionByDefault(ctx, &input)
		if err != nil {
			return err
		}

		if aws.ToBool(response.EbsEncryptionByDefault) != enabled {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", enabled)
		}

		return nil
	}
}

func testAccCheckEBSEncryptionByDefaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return testAccCheckEncryptionByDefaultDestroy(ctx, t)
}

func testAccCheckEBSEncryptionByDefaultExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return testAccCheckEBSEncryptionByDefault(ctx, t, resourceName, true)
}

func testAccEBSEncryptionByDefault_importUpgrade(t *testing.T) {
	testCases := map[string]struct {
		importID        string
		alternateRegion bool
	}{
		"generatedID": {},
		"default": {
			importID: "default",
		},
		"generatedIDRegionOverride": {
			alternateRegion: true,
		},
		"defaultRegionOverride": {
			importID:        "default",
			alternateRegion: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := acctest.Context(t)
			resourceName := "aws_ebs_encryption_by_default.test"
			region := acctest.Region()
			if tc.alternateRegion {
				region = acctest.AlternateRegion()
			}

			var originalID string

			destroyCtx := conns.NewResourceContext(
				ctx,
				names.EC2,
				"EBS Encryption By Default",
				"aws_ebs_encryption_by_default",
				region,
			)

			acctest.Test(ctx, t, resource.TestCase{
				PreCheck: func() {
					acctest.PreCheck(ctx, t)
					if tc.alternateRegion {
						acctest.PreCheckMultipleRegion(t, 2)
					}
				},
				ErrorCheck: acctest.ErrorCheck(t, names.EC2ServiceID),
				CheckDestroy: testAccCheckEncryptionByDefaultDestroy(
					destroyCtx, t,
				),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"aws": {
								Source:            "hashicorp/aws",
								VersionConstraint: "= 6.72.0",
							},
						},
						Config: testAccEBSEncryptionByDefaultConfig_importUpgrade(
							region, true,
						),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceName,
								names.AttrEnabled,
								acctest.CtTrue,
							),
							resource.TestCheckResourceAttr(
								resourceName,
								names.AttrRegion,
								region,
							),
							func(s *terraform.State) error {
								rs, ok := s.RootModule().Resources[resourceName]
								if !ok || rs.Primary == nil {
									return fmt.Errorf("resource %q not found", resourceName)
								}

								originalID = rs.Primary.ID
								if originalID == "" {
									return fmt.Errorf("resource %q has an empty ID", resourceName)
								}
								if inttypes.IsAWSRegion(originalID) {
									return fmt.Errorf(
										"expected a legacy generated ID, got region-form ID %q",
										originalID,
									)
								}

								return nil
							},
						),
					},
					{
						ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
						Config: testAccEBSEncryptionByDefaultConfig_importUpgrade(
							region, false,
						),
						ResourceName: resourceName,
						ImportState:  true,
						ImportStateIdFunc: func(_ *terraform.State) (string, error) {
							id := originalID
							if tc.importID != "" {
								id = tc.importID
							}
							if id == "" {
								return "", fmt.Errorf("original resource ID was not captured")
							}
							if tc.alternateRegion {
								id += "@" + region
							}

							return id, nil
						},
						ImportStateVerify:                    true,
						ImportStateVerifyIdentifierAttribute: names.AttrRegion,
						ImportStateVerifyIgnore:              []string{names.AttrID},
						ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
							func(states []*terraform.InstanceState) error {
								if len(states) != 1 {
									return fmt.Errorf(
										"expected one imported resource, got %d",
										len(states),
									)
								}
								if got := states[0].ID; got != region {
									return fmt.Errorf(
										"expected imported ID %q, got %q",
										region, got,
									)
								}

								return nil
							},
							acctest.ImportCheckResourceAttr(
								names.AttrID, region,
							),
							acctest.ImportCheckResourceAttr(
								names.AttrRegion, region,
							),
							acctest.ImportCheckResourceAttr(
								names.AttrEnabled, acctest.CtTrue,
							),
						),
					},
				},
			})
		})
	}
}

func testAccEBSEncryptionByDefaultConfig_importUpgrade(region string, preChange bool) string {
	versionConstraint := ""
	if preChange {
		versionConstraint = `version = "6.72.0"`
	}

	return fmt.Sprintf(`
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      %[1]s
    }
  }
}

resource "aws_ebs_encryption_by_default" "test" {
  region  = %[2]q
  enabled = true
}
`, versionConstraint, region)
}

func testAccEBSEncryptionByDefaultConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ebs_encryption_by_default" "test" {
  enabled = %[1]t
}
`, enabled)
}
