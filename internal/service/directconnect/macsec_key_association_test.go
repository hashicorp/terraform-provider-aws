// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectMacSecKeyAssociation_withCkn(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	resourceName := "aws_dx_macsec_key_association.test"
	ckn := testAccMacSecGenerateHex()
	cak := testAccMacSecGenerateHex()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMacSecKeyAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecKeyAssociationConfig_withCkn(ckn, cak, connectionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, ckn),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestMatchResourceAttr(resourceName, "ckn", regexache.MustCompile(ckn)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore the "cak" attribute as isn't returned by the API during read/refresh
				ImportStateVerifyIgnore: []string{"cak"},
			},
		},
	})
}

func TestAccDirectConnectMacSecKeyAssociation_writeOnly(t *testing.T) {
	for _, testCase := range []struct { //nolint:paralleltest // acctest.ParallelTest calls t.Parallel.
		name       string
		cakVersion int
		cknVersion int
	}{
		{name: "cak", cakVersion: 1},
		{name: "ckn", cknVersion: 1},
		{name: "both", cakVersion: 1, cknVersion: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := acctest.Context(t)
			connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
			resourceName := "aws_dx_macsec_key_association.test"
			cak, ckn := testAccMacSecGenerateHex(), testAccMacSecGenerateHex()
			cakUpdated, cknUpdated := cak, ckn
			cakVersion, cknVersion := testCase.cakVersion, testCase.cknVersion
			stateChecks := []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cak_wo"), knownvalue.Null()),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ckn_wo"), knownvalue.Null()),
			}
			importIgnore := []string{"cak"}
			if cakVersion > 0 {
				cakUpdated = testAccMacSecGenerateHex()
				stateChecks = append(stateChecks, statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cak"), knownvalue.Null()))
				importIgnore = append(importIgnore, "cak_wo_version")
			}
			if cknVersion > 0 {
				cknUpdated = testAccMacSecGenerateHex()
				stateChecks = append(stateChecks, statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ckn"), knownvalue.StringExact("")))
				// Import reads the CKN from AWS without the write-only configuration.
				importIgnore = append(importIgnore, "ckn", "ckn_wo_version")
			}

			steps := []resource.TestStep{
				{
					Config: testAccMacSecKeyAssociationConfig_writeOnly(ckn, cak, connectionID, cknVersion, cakVersion),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, ckn),
						resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					),
					ConfigStateChecks: stateChecks,
				},
				{
					Config: testAccMacSecKeyAssociationConfig_writeOnly(cknUpdated, cakUpdated, connectionID, cknVersion, cakVersion),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, ckn),
					),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						},
					},
					ConfigStateChecks: stateChecks,
				},
			}
			if cakVersion > 0 {
				cakVersion++
				steps = append(steps, resource.TestStep{
					Config: testAccMacSecKeyAssociationConfig_writeOnly(ckn, cakUpdated, connectionID, cknVersion, cakVersion),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, ckn),
						resource.TestCheckResourceAttr(resourceName, "cak_wo_version", "2"),
					),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
						},
					},
					ConfigStateChecks: stateChecks,
				})
			}
			if cknVersion > 0 {
				cknVersion++
				steps = append(steps, resource.TestStep{
					Config: testAccMacSecKeyAssociationConfig_writeOnly(cknUpdated, cakUpdated, connectionID, cknVersion, cakVersion),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, cknUpdated),
						resource.TestCheckResourceAttr(resourceName, "ckn_wo_version", "2"),
					),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
						},
					},
					ConfigStateChecks: stateChecks,
				})
			}
			steps = append(steps, resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: importIgnore,
			})

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_11_0),
				},
				CheckDestroy: testAccCheckMacSecKeyAssociationDestroy(ctx, t),
				Steps:        steps,
			})
		})
	}
}

func TestAccDirectConnectMacSecKeyAssociation_withSecret(t *testing.T) {
	ctx := acctest.Context(t)
	connectionID := acctest.SkipIfEnvVarNotSet(t, "DX_CONNECTION_ID")
	secretARN := acctest.SkipIfEnvVarNotSet(t, "SECRET_ARN")
	resourceName := "aws_dx_macsec_key_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMacSecKeyAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMacSecKeyAssociationConfig_withSecret(secretARN, connectionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMacSecKeyAssociationExists(ctx, t, resourceName, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, connectionID),
					resource.TestCheckResourceAttr(resourceName, "secret_arn", secretARN),
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

func testAccMacSecGenerateHex() string {
	s := make([]byte, 32)
	if _, err := rand.Read(s); err != nil {
		return ""
	}
	return hex.EncodeToString(s)
}

func testAccCheckMacSecKeyAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_macsec_key_association" {
				continue
			}

			_, err := tfdirectconnect.FindMacSecKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrConnectionID], rs.Primary.Attributes["secret_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect MACSec Key Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckMacSecKeyAssociationExists(ctx context.Context, t *testing.T, name, ckn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Direct Connect MACSec Key Association ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		key, err := tfdirectconnect.FindMacSecKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrConnectionID], rs.Primary.Attributes["secret_arn"])
		if err != nil {
			return err
		}
		if ckn != "" && (key.Ckn == nil || *key.Ckn != ckn) {
			return fmt.Errorf("Direct Connect MACSec Key Association CKN does not match the configured value")
		}

		return nil
	}
}

func testAccMacSecKeyAssociationConfig_writeOnly(ckn, cak, connectionID string, cknVersion, cakVersion int) string {
	cknConfig := fmt.Sprintf("ckn = %q", ckn)
	if cknVersion > 0 {
		cknConfig = fmt.Sprintf("ckn_wo = %q\nckn_wo_version = %d", ckn, cknVersion)
	}
	cakConfig := fmt.Sprintf("cak = %q", cak)
	if cakVersion > 0 {
		cakConfig = fmt.Sprintf("cak_wo = %q\ncak_wo_version = %d", cak, cakVersion)
	}

	return fmt.Sprintf(`
resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[1]q
  %[2]s
  %[3]s
}
`, connectionID, cknConfig, cakConfig)
}

func testAccMacSecKeyAssociationConfig_withCkn(ckn, cak, connectionID string) string {
	return fmt.Sprintf(`
resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[3]q
  ckn           = %[1]q
  cak           = %[2]q
}
`, ckn, cak, connectionID)
}

// Can only be used with an EXISTING secrets created by previous association - cannot create secrets from scratch.
func testAccMacSecKeyAssociationConfig_withSecret(secretARN, connectionID string) string {
	return fmt.Sprintf(`
data "aws_secretsmanager_secret" "test" {
  arn = %[1]q
}

resource "aws_dx_macsec_key_association" "test" {
  connection_id = %[2]q
  secret_arn    = data.aws_secretsmanager_secret.test.arn
}
`, secretARN, connectionID)
}

func TestMacSecKeyAssociationSchema(t *testing.T) {
	t.Parallel()

	key := strings.Repeat("a", 64)
	secretARN := "arn:aws:secretsmanager:us-west-2:123456789012:secret:example" //lintignore:AWSAT003,AWSAT005
	testCases := map[string]struct {
		config      map[string]any
		expectError bool
	}{
		"regular": {
			config: map[string]any{"cak": key, "ckn": key},
		},
		"secret": {
			config: map[string]any{"secret_arn": secretARN},
		},
		"write-only": {
			config: map[string]any{"cak_wo": key, "cak_wo_version": 1, "ckn_wo": key, "ckn_wo_version": 1},
		},
		"write-only CAK": {
			config: map[string]any{"cak_wo": key, "cak_wo_version": 1, "ckn": key},
		},
		"write-only CKN": {
			config: map[string]any{"cak": key, "ckn_wo": key, "ckn_wo_version": 1},
		},
		"missing CAK": {
			config: map[string]any{"ckn_wo": key, "ckn_wo_version": 1}, expectError: true,
		},
		"missing CKN": {
			config: map[string]any{"cak_wo": key, "cak_wo_version": 1}, expectError: true,
		},
		"missing CAK version": {
			config: map[string]any{"cak_wo": key, "ckn": key}, expectError: true,
		},
		"missing CKN version": {
			config: map[string]any{"cak": key, "ckn_wo": key}, expectError: true,
		},
		"CAK version without value": {
			config: map[string]any{"cak_wo_version": 1, "secret_arn": secretARN}, expectError: true,
		},
		"CKN version without value": {
			config: map[string]any{"ckn_wo_version": 1, "secret_arn": secretARN}, expectError: true,
		},
		"conflicting CAK": {
			config: map[string]any{"cak": key, "cak_wo": key, "cak_wo_version": 1, "ckn": key}, expectError: true,
		},
		"conflicting CKN": {
			config: map[string]any{"cak": key, "ckn": key, "ckn_wo": key, "ckn_wo_version": 1}, expectError: true,
		},
		"CAK with secret": {
			config: map[string]any{"cak_wo": key, "cak_wo_version": 1, "secret_arn": secretARN}, expectError: true,
		},
		"CKN with secret": {
			config: map[string]any{"ckn_wo": key, "ckn_wo_version": 1, "secret_arn": secretARN}, expectError: true,
		},
		"invalid CAK": {
			config: map[string]any{"cak_wo": "invalid", "cak_wo_version": 1, "ckn": key}, expectError: true,
		},
		"invalid CKN": {
			config: map[string]any{"cak": key, "ckn_wo": "invalid", "ckn_wo_version": 1}, expectError: true,
		},
		"zero CAK version": {
			config: map[string]any{"cak_wo": key, "cak_wo_version": 0, "ckn": key}, expectError: true,
		},
		"zero CKN version": {
			config: map[string]any{"cak": key, "ckn_wo": key, "ckn_wo_version": 0}, expectError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := tfdirectconnect.ResourceMacSecKeyAssociation()
			if err := r.InternalValidate(nil, true); err != nil {
				t.Fatalf("invalid resource schema: %s", err)
			}
			testCase.config[names.AttrConnectionID] = "dxcon-example"
			diags := r.Validate(terraformsdk.NewResourceConfigRaw(testCase.config))
			if diags.HasError() != testCase.expectError {
				t.Fatalf("expected error: %t, got diagnostics: %v", testCase.expectError, diags)
			}
		})
	}
}
