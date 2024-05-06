// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	tfsts "github.com/hashicorp/terraform-provider-aws/internal/service/sts"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Prerequisite: the current account must be either:
// * not a member of an organization
// * a delegated Security Lake administrator account
func TestAccSecurityLake_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AWSLogSource": {
			"basic":         testAccAWSLogSource_basic,
			"disappears":    testAccAWSLogSource_disappears,
			"multiple":      testAccAWSLogSource_multiple,
			"multiRegion":   testAccAWSLogSource_multiRegion,
			"sourceVersion": testAccAWSLogSource_sourceVersion,
		},
		"CustomLogSource": {
			"basic":         testAccCustomLogSource_basic,
			"disappears":    testAccCustomLogSource_disappears,
			"eventClasses":  testAccCustomLogSource_eventClasses,
			"multiple":      testAccCustomLogSource_multiple,
			"sourceVersion": testAccCustomLogSource_sourceVersion,
		},
		"DataLake": {
			"basic":           testAccDataLake_basic,
			"disappears":      testAccDataLake_disappears,
			names.AttrTags:    testAccDataLake_tags,
			"lifecycle":       testAccDataLake_lifeCycle,
			"lifecycleUpdate": testAccDataLake_lifeCycleUpdate,
			"replication":     testAccDataLake_replication,
		},
		"Subscriber": {
			"accessType":      testAccSubscriber_accessType,
			"basic":           testAccSubscriber_basic,
			"customLogs":      testAccSubscriber_customLogSource,
			"disappears":      testAccSubscriber_disappears,
			"multipleSources": testAccSubscriber_multipleSources,
			names.AttrTags:    testAccSubscriber_tags,
			"updated":         testAccSubscriber_update,
			"migrateSource":   testAccSubscriber_migrate_source,
		},
		"SubscriberNotification": {
			"basic":      testAccSubscriberNotification_basic,
			"https":      testAccSubscriberNotification_https,
			"disappears": testAccSubscriberNotification_disappears,
			"update":     testAccSubscriberNotification_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

// testAccPreCheck validates that the current account is either
// * not a member of an organization
// * a member of an organization and is the delegated Security Lake administrator account
func testAccPreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	awsClient := acctest.Provider.Meta().(*conns.AWSClient)

	organization, err := tforganizations.FindOrganization(ctx, awsClient.OrganizationsConn(ctx))

	// Not a member of an organization
	if tfresource.NotFound(err) {
		return
	}

	callerIdentity, err := tfsts.FindCallerIdentity(ctx, awsClient.STSClient(ctx))

	if err != nil {
		t.Fatalf("getting current identity: %s", err)
	}

	if aws.StringValue(organization.MasterAccountId) == aws.StringValue(callerIdentity.Account) {
		t.Skip("this AWS account must not be the management account of an AWS Organization")
	}

	_, err = tfsecuritylake.FindDataLakes(ctx, awsClient.SecurityLakeClient(ctx), &securitylake.ListDataLakesInput{}, slices.PredicateTrue[*awstypes.DataLakeResource]())

	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "must be a delegated Security Lake administrator account") {
		t.Skip("this AWS account must be a delegate Security Lake administrator account")
	}

	if err != nil {
		t.Fatalf("finding data lakes: %s", err)
	}
}
