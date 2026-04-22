// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSCertificate_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccCertificate_basic,
		acctest.CtDisappears: testAccCertificate_disappears,
		"Identity":           testAccRDSCertificate_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Certificate
	resourceName := "aws_rds_certificate.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic("rds-ca-rsa4096-g1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_identifier", "rds-ca-rsa4096-g1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_basic("rds-ca-ecc384-g1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_identifier", "rds-ca-ecc384-g1"),
				),
			},
		},
	})
}

// func testAccRDSCertificate_Identity_basic(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	var v types.Certificate
// 	resourceName := "aws_rds_certificate.test"

// 	resource.Test(t, resource.TestCase{
// 		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
// 			tfversion.SkipBelow(tfversion.Version1_12_0),
// 		},
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCertificateConfig_basic("rds-ca-rsa4096-g1"),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccCheckCertificateExists(ctx, resourceName, &v),
// 				),
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrRegion), compare.ValuesSame()),
// 					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
// 					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
// 						names.AttrAccountID: tfknownvalue.AccountID(),
// 						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
// 					}),
// 				},
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateKind:   resource.ImportCommandWithID,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				ResourceName:    resourceName,
// 				ImportState:     true,
// 				ImportStateKind: resource.ImportBlockWithID,
// 				ImportPlanChecks: resource.ImportPlanChecks{
// 					PreApply: []plancheck.PlanCheck{
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.Region())),
// 					},
// 				},
// 			},
// 			{
// 				ResourceName:    resourceName,
// 				ImportState:     true,
// 				ImportStateKind: resource.ImportBlockWithResourceIdentity,
// 				ImportPlanChecks: resource.ImportPlanChecks{
// 					PreApply: []plancheck.PlanCheck{
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.Region())),
// 					},
// 				},
// 			},
// 		},
// 	})
// }

// func testAccRDSCertificate_Identity_regionOverride(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	resourceName := "aws_rds_certificate.test"

// 	resource.Test(t, resource.TestCase{
// 		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
// 			tfversion.SkipBelow(tfversion.Version1_12_0),
// 		},
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCertificateConfig_regionOverride(),
// 				ConfigStateChecks: []statecheck.StateCheck{
// 					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrRegion), compare.ValuesSame()),
// 					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
// 					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
// 						names.AttrAccountID: tfknownvalue.AccountID(),
// 						names.AttrRegion:    knownvalue.StringExact(acctest.AlternateRegion()),
// 					}),
// 				},
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateKind:   resource.ImportCommandWithID,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				ResourceName:    resourceName,
// 				ImportState:     true,
// 				ImportStateKind: resource.ImportBlockWithID,
// 				ImportPlanChecks: resource.ImportPlanChecks{
// 					PreApply: []plancheck.PlanCheck{
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.AlternateRegion())),
// 					},
// 				},
// 			},
// 			{
// 				ResourceName:    resourceName,
// 				ImportState:     true,
// 				ImportStateKind: resource.ImportBlockWithResourceIdentity,
// 				ImportPlanChecks: resource.ImportPlanChecks{
// 					PreApply: []plancheck.PlanCheck{
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
// 						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringExact(acctest.AlternateRegion())),
// 					},
// 				},
// 			},
// 		},
// 	})
// }

func testAccCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Certificate
	resourceName := "aws_rds_certificate.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic("rds-ca-rsa4096-g1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfrds.ResourceCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_certificate" {
				continue
			}

			_, err := tfrds.FindDefaultCertificate(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Default Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCertificateExists(ctx context.Context, t *testing.T, n string, v *types.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).RDSClient(ctx)

		output, err := tfrds.FindDefaultCertificate(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCertificateConfig_basic(certificateID string) string {
	return fmt.Sprintf(`
resource "aws_rds_certificate" "test" {
  certificate_identifier = %[1]q
}
`, certificateID)
}
