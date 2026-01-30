// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	gluetypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// Prerequisite: the current account must be either:
// * not a member of an organization
// * a delegated Security Lake administrator account
func TestAccSecurityLake_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AWSLogSource": {
			acctest.CtBasic:      testAccAWSLogSource_basic,
			acctest.CtDisappears: testAccAWSLogSource_disappears,
			"multiple":           testAccAWSLogSource_multiple,
			"multiRegion":        testAccAWSLogSource_multiRegion,
			"sourceVersion":      testAccAWSLogSource_sourceVersion,
		},
		"CustomLogSource": {
			acctest.CtBasic:      testAccCustomLogSource_basic,
			acctest.CtDisappears: testAccCustomLogSource_disappears,
			"eventClasses":       testAccCustomLogSource_eventClasses,
			"multiple":           testAccCustomLogSource_multiple,
			"sourceVersion":      testAccCustomLogSource_sourceVersion,
		},
		"DataLake": {
			acctest.CtBasic:      testAccDataLake_basic,
			acctest.CtDisappears: testAccDataLake_disappears,
			"tags":               testAccDataLake_tags,
			"lifecycle":          testAccDataLake_lifeCycle,
			"metaStoreUpdate":    testAccDataLake_metaStoreUpdate,
			"replication":        testAccDataLake_replication,
			"Identity":           testAccDataLake_IdentitySerial,
		},
		"Subscriber": {
			"accessType":         testAccSubscriber_accessType,
			acctest.CtBasic:      testAccSubscriber_basic,
			"customLogs":         testAccSubscriber_customLogSource,
			acctest.CtDisappears: testAccSubscriber_disappears,
			"multipleSources":    testAccSubscriber_multipleSources,
			"tags":               testAccSubscriber_tags,
			"updated":            testAccSubscriber_update,
			"migrateSource":      testAccSubscriber_migrate_source,
		},
		"SubscriberNotification": {
			acctest.CtDisappears: testAccSubscriberNotification_disappears,
			"https_basic":        testAccSubscriberNotification_https_basic,
			"update":             testAccSubscriberNotification_update,
			"sqs_basic":          testAccSubscriberNotification_sqs_basic,
			"apiKeyNameOnly":     testAccSubscriberNotification_https_apiKeyNameOnly,
			"apiKey":             testAccSubscriberNotification_https_apiKey,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

// testAccPreCheck validates that the current account is either
// * not a member of an organization
// * a member of an organization and is the delegated Security Lake administrator account
func testAccPreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	acctest.PreCheckOrganizationMemberAccount(ctx, t)

	_, err := tfsecuritylake.FindDataLakes(ctx, acctest.ProviderMeta(ctx, t).SecurityLakeClient(ctx), &securitylake.ListDataLakesInput{}, tfslices.PredicateTrue[*awstypes.DataLakeResource]())

	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "must be a delegated Security Lake administrator account") {
		t.Skip("this AWS account must be a delegated Security Lake administrator account")
	}

	if err != nil {
		t.Fatalf("finding data lakes: %s", err)
	}
}

func testAccDeleteGlueDatabases(ctx context.Context, t *testing.T, regions ...string) {
	t.Helper()

	if os.Getenv(resource.EnvTfAcc) == "" {
		return
	}

	// The Context returned by t.Context() is canceled when the test ends.
	ctx = context.WithoutCancel(ctx)

	for _, region := range regions {
		deleteGlueDatabase(ctx, t, region)
	}
}

func deleteGlueDatabase(ctx context.Context, t *testing.T, region string) {
	t.Helper()

	// e.g. "amazon_security_lake_glue_db_us-east-1"
	databaseName := "amazon_security_lake_glue_db_" + strings.ReplaceAll(region, "-", "_")
	input := glue.DeleteDatabaseInput{
		Name: aws.String(databaseName),
	}
	_, err := acctest.ProviderMeta(ctx, t).GlueClient(ctx).DeleteDatabase(ctx, &input, func(o *glue.Options) {
		o.Region = region
	})
	if errs.IsA[*gluetypes.EntityNotFoundException](err) {
		return
	}
	if err != nil {
		t.Errorf("deleting Glue Database (%s) in %q: %s", databaseName, region, err)
	}
}
