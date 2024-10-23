// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package simpledb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_simpledb_domain", &resource.Sweeper{
		Name: "aws_simpledb_domain",
		F:    sweepDomains,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := simpleDBConn(ctx, client)
	input := &simpledb.ListDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDomainsPagesWithContext(ctx, input, func(page *simpledb.ListDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainNames {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDomainResource, client,
				framework.NewAttribute(names.AttrID, aws.StringValue(v)),
			))
		}

		return !lastPage
	})

	if skipSweepError(err) {
		log.Printf("[WARN] Skipping SimpleDB Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SimpleDB Domains (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SimpleDB Domains (%s): %w", region, err)
	}

	return nil
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func skipSweepError(err error) bool {
	// Ignore missing API endpoints for AWS SDK for Go v1
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}

	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if tfawserr.ErrMessageContains(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: InvalidAction: Operation (ListPlatformApplications) is not supported in this region
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not valid") {
		return true
	}
	// For example from GovCloud SES.SetActiveReceiptRuleSet.
	if tfawserr.ErrMessageContains(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	// For example from us-west-2 Route53 key signing key
	if tfawserr.ErrMessageContains(err, "InvalidKeySigningKeyStatus", "cannot be deleted because") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// For example from us-gov-west-1 MemoryDB cluster
	if tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version") {
		return true
	}
	// For example from us-west-2 Route53 zone
	if tfawserr.ErrMessageContains(err, "KeySigningKeyInParentDSRecord", "Due to DNS lookup failure") {
		return true
	}
	// For example from us-gov-east-1 IoT domain configuration
	if tfawserr.ErrMessageContains(err, "UnauthorizedException", "API is not available in") {
		return true
	}
	// For example from us-gov-west-1 EventBridge archive
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "Operation is disabled in this region") {
		return true
	}
	// For example from us-east-1 SageMaker
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "The requested operation is not supported in the called region") {
		return true
	}
	// For example from us-west-2 ECR public repository
	if tfawserr.ErrMessageContains(err, "UnsupportedCommandException", "command is only supported in") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrCodeEquals(err, "UnsupportedOperation") {
		return true
	}
	// For example from us-west-1 EMR studio
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account is not whitelisted to use this feature") {
		return true
	}
	// For example from us-west-2 SageMaker device fleet
	if tfawserr.ErrMessageContains(err, "ValidationException", "We are retiring Amazon Sagemaker Edge") {
		return true
	}

	return false
}
