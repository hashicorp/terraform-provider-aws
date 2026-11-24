// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestRecordSetProtectionImportIDParse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		id      string
		want    map[string]any
		wantErr bool
	}{
		"valid": {
			id: "Z1D633PJN98FT9,www.example.com,A",
			want: map[string]any{
				"zone_id":     "Z1D633PJN98FT9",
				"record_name": "www.example.com",
				"record_type": "A",
			},
		},
		"wildcard with trailing dot": {
			id: "Z1D633PJN98FT9,*.example.com.,MX",
			want: map[string]any{
				"zone_id":     "Z1D633PJN98FT9",
				"record_name": "*.example.com.",
				"record_type": "MX",
			},
		},
		"two parts": {
			id:      "Z1D633PJN98FT9,www.example.com",
			wantErr: true,
		},
		"four parts": {
			id:      "Z1D633PJN98FT9,www.example.com,A,extra",
			wantErr: true,
		},
		"empty name": {
			id:      "Z1D633PJN98FT9,,A",
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, got, err := tfroute53.RecordSetProtectionImportID{}.Parse(testCase.id)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !maps.Equal(got, testCase.want) {
				t.Errorf("Parse(%q) = %v, want %v", testCase.id, got, testCase.want)
			}
		})
	}
}

func protectionBatch(codes ...string) *awstypes.InvalidResourceRecordSetProtectionBatch {
	v := &awstypes.InvalidResourceRecordSetProtectionBatch{
		Message: aws.String("batch failed"),
	}
	for _, code := range codes {
		v.Errors = append(v.Errors, awstypes.ProtectionError{
			Code:       aws.String(code),
			Message:    aws.String(code + " message"),
			RecordName: aws.String("www.example.com."),
			RecordType: awstypes.RRTypeA,
		})
	}
	return v
}

func TestIsProtectionNotFound(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err  error
		want bool
	}{
		"nil": {},
		"not found": {
			err:  protectionBatch("ProtectionNotFound"),
			want: true,
		},
		"wrapped not found": {
			err:  fmt.Errorf("deleting: %w", protectionBatch("ProtectionNotFound")),
			want: true,
		},
		"invalid input": {
			err: protectionBatch("InvalidInput"),
		},
		"unknown code": {
			err: protectionBatch("SomeFutureCode"),
		},
		"mixed codes": {
			err: protectionBatch("ProtectionNotFound", "InvalidInput"),
		},
		"no entries": {
			err: protectionBatch(),
		},
		"other error type": {
			err: &awstypes.NoSuchHostedZone{Message: aws.String("ProtectionNotFound")},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfroute53.IsProtectionNotFound(testCase.err); got != testCase.want {
				t.Errorf("IsProtectionNotFound() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestProtectionBatchError(t *testing.T) {
	t.Parallel()

	batch := protectionBatch("InvalidInput")
	err := tfroute53.ProtectionBatchError(batch)

	if !errors.Is(err, batch) {
		t.Errorf("ProtectionBatchError() does not wrap the batch error")
	}
	if want := "www.example.com. A: InvalidInput: InvalidInput message"; !strings.Contains(err.Error(), want) {
		t.Errorf("ProtectionBatchError() = %q, want it to contain %q", err, want)
	}

	other := errors.New("other")
	if got := tfroute53.ProtectionBatchError(other); got != other {
		t.Errorf("ProtectionBatchError(other) = %v, want it unchanged", got)
	}
}

// A protection on a name and type with no records reserves it.
func TestAccRoute53RecordSetProtection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSetProtection
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_record_set_protection.test"
	zoneResourceName := "aws_route53_zone.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", zoneResourceName, "zone_id"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrComment), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("record_name"), knownvalue.StringExact(recordName.String())),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("record_type"), knownvalue.StringExact(string(awstypes.RRTypeA))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "zone_id", "record_name", "record_type"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone_id",
			},
		},
	})
}

func TestAccRoute53RecordSetProtection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSetProtection
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_record_set_protection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfroute53.ResourceRecordSetProtection, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// There is no update API, so changing the comment replaces the protection.
func TestAccRoute53RecordSetProtection_comment(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSetProtection
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_record_set_protection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionConfig_comment(zoneName.String(), recordName.String(), "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrComment), knownvalue.StringExact("first")),
				},
			},
			{
				Config: testAccRecordSetProtectionConfig_comment(zoneName.String(), recordName.String(), "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrComment), knownvalue.StringExact("second")),
				},
			},
			{
				Config: testAccRecordSetProtectionConfig_basic(zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrComment), knownvalue.Null()),
				},
			},
		},
	})
}

// Route 53 returns names lowercased, fully qualified and with "*" escaped.
// The configured spelling must survive refresh without a diff.
func TestAccRoute53RecordSetProtection_nameNormalization(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSetProtection
	zoneName := acctest.RandomDomain(t)
	recordName := "*." + strings.ToUpper(zoneName.RandomSubdomain(t).String()) + "."
	resourceName := "aws_route53_record_set_protection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionConfig_basic(zoneName.String(), recordName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("record_name"), knownvalue.StringExact(recordName)),
				},
			},
			{
				Config: testAccRecordSetProtectionConfig_basic(zoneName.String(), recordName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "zone_id", "record_name", "record_type"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "zone_id",
			},
		},
	})
}

// One protection covers every record with its name and type, whatever the set
// identifier, and Route 53 rejects changes to any of them.
func TestAccRoute53RecordSetProtection_weightedRecords(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResourceRecordSetProtection
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)
	resourceName := "aws_route53_record_set_protection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRecordSetProtectionConfig_weightedRecords(zoneName.String(), recordName.String(), 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecordSetProtectionExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config:      testAccRecordSetProtectionConfig_weightedRecords(zoneName.String(), recordName.String(), 60),
				ExpectError: regexache.MustCompile(`is protected and cannot be changed`),
			},
		},
	})
}

// An existing protection is reported, not adopted.
func TestAccRoute53RecordSetProtection_alreadyExists(t *testing.T) {
	ctx := acctest.Context(t)
	zoneName := acctest.RandomDomain(t)
	recordName := zoneName.RandomSubdomain(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecordSetProtectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRecordSetProtectionConfig_alreadyExists(zoneName.String(), recordName.String()),
				ExpectError: regexache.MustCompile(`ProtectionAlreadyExists`),
			},
		},
	})
}

func testAccCheckRecordSetProtectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_record_set_protection" {
				continue
			}

			zoneID, recordName, recordType := rs.Primary.Attributes["zone_id"], rs.Primary.Attributes["record_name"], rs.Primary.Attributes["record_type"]
			_, err := tfroute53.FindRecordSetProtectionByThreePartKey(ctx, conn, zoneID, recordName, recordType)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Record Set Protection %s,%s,%s still exists", zoneID, recordName, recordType)
		}

		return nil
	}
}

func testAccCheckRecordSetProtectionExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceRecordSetProtection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		output, err := tfroute53.FindRecordSetProtectionByThreePartKey(ctx, conn, rs.Primary.Attributes["zone_id"], rs.Primary.Attributes["record_name"], rs.Primary.Attributes["record_type"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRecordSetProtectionConfig_base(zoneName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[1]q
}
`, zoneName)
}

func testAccRecordSetProtectionConfig_basic(zoneName, recordName string) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), fmt.Sprintf(`
resource "aws_route53_record_set_protection" "test" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = %[1]q
  record_type = "A"
}
`, recordName))
}

func testAccRecordSetProtectionConfig_comment(zoneName, recordName, comment string) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), fmt.Sprintf(`
resource "aws_route53_record_set_protection" "test" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = %[1]q
  record_type = "A"
  comment     = %[2]q
}
`, recordName, comment))
}

func testAccRecordSetProtectionConfig_weightedRecords(zoneName, recordName string, ttl int) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), fmt.Sprintf(`
resource "aws_route53_record" "test" {
  count = 2

  zone_id        = aws_route53_zone.test.zone_id
  name           = %[1]q
  type           = "A"
  ttl            = %[2]d
  records        = ["192.0.2.${count.index + 1}"]
  set_identifier = "test-${count.index}"

  weighted_routing_policy {
    weight = 1
  }
}

# Referencing every record makes Terraform remove the protection first on destroy.
resource "aws_route53_record_set_protection" "test" {
  zone_id     = aws_route53_zone.test.zone_id
  record_name = aws_route53_record.test[0].name
  record_type = aws_route53_record.test[0].type
  comment     = "acceptance test"

  depends_on = [aws_route53_record.test]
}
`, recordName, ttl))
}

func testAccRecordSetProtectionConfig_alreadyExists(zoneName, recordName string) string {
	return acctest.ConfigCompose(testAccRecordSetProtectionConfig_base(zoneName), fmt.Sprintf(`
resource "aws_route53_record_set_protection" "test" {
  count = 2

  zone_id     = aws_route53_zone.test.zone_id
  record_name = %[1]q
  record_type = "A"
}
`, recordName))
}
