// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	// Applications must be swept after Entitlements: DeleteApplication will
	// fail (ConflictException) while child Entitlements still exist.
	awsv2.Register("aws_accountaccess_application", sweepApplications, "aws_accountaccess_entitlement")
	awsv2.Register("aws_accountaccess_entitlement", sweepEntitlements)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AccountAccessClient(ctx)
	var sweepResources []sweep.Sweepable
	var nextToken *string

	for {
		output, err := conn.ListApplications(ctx, &accountaccess.ListApplicationsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, app := range output.Applications {
			arn := aws.ToString(app.ApplicationArn)
			if arn == "" {
				continue
			}
			sweepResources = append(sweepResources, framework.NewSweepResource(
				newApplicationResource, client,
				framework.NewAttribute(names.AttrARN, arn),
				framework.NewAttribute(names.AttrID, arn),
			))
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return sweepResources, nil
}

// sweepEntitlements lists Entitlements for every Application in the caller
// account. Because ListEntitlements requires a filter (see questions.md, AAM
// team Q5), we filter on the caller's account ID — entitlements leaked by
// acceptance tests typically grant access to a role in this same account, so
// this catches the common leak shape. Entitlements pointing at other accounts
// are not swept; if you need that, run the sweeper with elevated cross-account
// list permissions and a different filter.
func sweepEntitlements(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AccountAccessClient(ctx)
	var sweepResources []sweep.Sweepable
	accountID := client.AccountID(ctx)

	// First page through all Applications.
	var nextToken *string
	for {
		appsOutput, err := conn.ListApplications(ctx, &accountaccess.ListApplicationsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, app := range appsOutput.Applications {
			applicationArn := aws.ToString(app.ApplicationArn)
			if applicationArn == "" {
				continue
			}

			// Per Application, list Entitlements filtered by the caller's
			// account.
			var entToken *string
			for {
				entOutput, err := conn.ListEntitlements(ctx, &accountaccess.ListEntitlementsInput{
					ApplicationArn: aws.String(applicationArn),
					Filter: &awstypes.EntitlementFilter{
						PrincipalRole: &awstypes.PrincipalRoleEntitlementFilter{
							Account: aws.String(accountID),
						},
					},
					NextToken: entToken,
				})
				if err != nil {
					return nil, err
				}

				for _, e := range entOutput.Entitlements {
					sweepResources = append(sweepResources, framework.NewSweepResource(
						newEntitlementResource, client,
						framework.NewAttribute("application_arn", applicationArn),
						framework.NewAttribute("entitlement_id", aws.ToString(e.EntitlementId)),
						framework.NewAttribute(names.AttrID, applicationArn+","+aws.ToString(e.EntitlementId)),
					))
				}

				if entOutput.NextToken == nil {
					break
				}
				entToken = entOutput.NextToken
			}
		}

		if appsOutput.NextToken == nil {
			break
		}
		nextToken = appsOutput.NextToken
	}

	return sweepResources, nil
}
