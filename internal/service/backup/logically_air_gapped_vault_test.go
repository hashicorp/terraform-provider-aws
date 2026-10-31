// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestLogicallyAirGappedVaultDelete(t *testing.T) {
	t.Parallel()

	type reply struct {
		method string
		path   string
		code   string
		body   string
	}

	deleted := reply{method: http.MethodDelete, path: "/backup-vaults/test"}
	present := reply{method: http.MethodGet, path: "/backup-vaults/test", body: `{"BackupVaultName":"test","VaultState":"FAILED","VaultType":"LOGICALLY_AIR_GAPPED_BACKUP_VAULT"}`}
	absent := reply{method: http.MethodGet, path: "/backup-vaults/test", code: "ResourceNotFoundException"}
	denied := reply{method: http.MethodGet, path: "/backup-vaults/test", code: "AccessDeniedException"}

	for _, tc := range []struct {
		name      string
		replies   []reply
		wantError bool
	}{
		{name: "waits_for_absence", replies: []reply{deleted, present, absent}},
		{name: "already_absent", replies: []reply{{method: http.MethodDelete, path: "/backup-vaults/test", code: "ResourceNotFoundException"}}},
		{name: "delete_access_denied", replies: []reply{{method: http.MethodDelete, path: "/backup-vaults/test", code: "AccessDeniedException"}}},
		{name: "delete_error", replies: []reply{{method: http.MethodDelete, path: "/backup-vaults/test", code: "InvalidParameterValueException"}}, wantError: true},
		{name: "describe_access_denied", replies: []reply{deleted, denied}},
		{name: "describe_error", replies: []reply{deleted, {method: http.MethodGet, path: "/backup-vaults/test", code: "InvalidParameterValueException"}}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				i := int(calls.Add(1)) - 1
				if i >= len(tc.replies) {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				reply := tc.replies[i]
				if request.Method != reply.method || request.URL.Path != reply.path {
					t.Errorf("request %d = %s %s, want %s %s", i, request.Method, request.URL.Path, reply.method, reply.path)
				}

				w.Header().Set("Content-Type", "application/json")
				body := reply.body
				if reply.code != "" {
					w.Header().Set("X-Amzn-Errortype", reply.code)
					if reply.code == "AccessDeniedException" {
						w.WriteHeader(http.StatusForbidden)
					} else {
						w.WriteHeader(http.StatusBadRequest)
					}
					body = `{"message":"scripted test error"}`
				}
				if body != "" {
					if _, err := fmt.Fprint(w, body); err != nil {
						t.Errorf("writing response: %s", err)
					}
				}
			}))
			defer ts.Close()

			ctx := t.Context()
			config := conns.Config{
				AccessKey: "test", SecretKey: "test", Region: endpoints.UsEast1RegionID,
				Endpoints:           map[string]string{names.Backup: ts.URL},
				SkipCredsValidation: true, SkipRequestingAccountId: true,
				SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{},
			}
			client := &conns.AWSClient{}
			client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.Backup: tfbackup.ServicePackage(ctx)})
			client, diags := config.ConfigureProvider(ctx, client)
			if diags.HasError() {
				t.Fatalf("configure client: %v", diags)
			}

			r, err := tfbackup.ResourceLogicallyAirGappedVault(ctx)
			if err != nil {
				t.Fatal(err)
			}
			configured := fwresource.ConfigureResponse{}
			r.Configure(ctx, fwresource.ConfigureRequest{ProviderData: client}, &configured)
			if configured.Diagnostics.HasError() {
				t.Fatalf("configure resource: %v", configured.Diagnostics)
			}

			schemaResponse := fwresource.SchemaResponse{}
			r.Schema(ctx, fwresource.SchemaRequest{}, &schemaResponse)
			schemaResponse.Schema.Attributes[names.AttrRegion] = schema.StringAttribute{Optional: true, Computed: true}
			objectType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
			values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
			for name, typ := range objectType.AttributeTypes {
				values[name] = tftypes.NewValue(typ, nil)
			}
			values[names.AttrID] = tftypes.NewValue(tftypes.String, "test")
			values[names.AttrRegion] = tftypes.NewValue(tftypes.String, endpoints.UsEast1RegionID)

			state := tfsdk.State{Schema: schemaResponse.Schema, Raw: tftypes.NewValue(objectType, values)}
			response := fwresource.DeleteResponse{State: state}
			r.Delete(ctx, fwresource.DeleteRequest{State: state}, &response)
			if response.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("diagnostics = %v, want error %t", response.Diagnostics, tc.wantError)
			}
			if got := int(calls.Load()); got != len(tc.replies) {
				t.Fatalf("API calls = %d, want %d", got, len(tc.replies))
			}
		})
	}
}

func TestAccBackupLogicallyAirGappedVault_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_retention_days"), knownvalue.Int64Exact(10)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_retention_days"), knownvalue.Int64Exact(7)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
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

func TestAccBackupLogicallyAirGappedVault_replacement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_retention(rName, 7),
				Check:  testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_retention_days"), knownvalue.Int64Exact(7)),
				},
			},
			{
				Config: testAccLogicallyAirGappedVaultConfig_retention(rName, 14),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("min_retention_days"), knownvalue.Int64Exact(14)),
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

func TestAccBackupLogicallyAirGappedVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbackup.ResourceLogicallyAirGappedVault, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBackupLogicallyAirGappedVault_encryptionKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BackupEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BackupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_encryptionKeyARN(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("encryption_key_arn"), kmsKeyResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
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

func TestAccBackupLogicallyAirGappedVault_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v backup.DescribeBackupVaultOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_backup_logically_air_gapped_vault.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogicallyAirGappedVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccLogicallyAirGappedVaultConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogicallyAirGappedVaultExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckLogicallyAirGappedVaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_backup_logically_air_gapped_vault" {
				continue
			}

			_, err := tfbackup.FindLogicallyAirGappedBackupVaultByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Backup Logically Air Gapped Vault %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLogicallyAirGappedVaultExists(ctx context.Context, t *testing.T, n string, v *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BackupClient(ctx)

		output, err := tfbackup.FindLogicallyAirGappedBackupVaultByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLogicallyAirGappedVaultConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 10
  min_retention_days = 7
}
`, rName)
}

func testAccLogicallyAirGappedVaultConfig_retention(rName string, minRetentionDays int) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  min_retention_days = %[2]d
  max_retention_days = 30
}
`, rName, minRetentionDays)
}

func testAccLogicallyAirGappedVaultConfig_encryptionKeyARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 10
  min_retention_days = 7
  encryption_key_arn = aws_kms_key.test.arn
}
`, rName)
}

func testAccLogicallyAirGappedVaultConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 7
  min_retention_days = 7

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccLogicallyAirGappedVaultConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_backup_logically_air_gapped_vault" "test" {
  name               = %[1]q
  max_retention_days = 7
  min_retention_days = 7

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
