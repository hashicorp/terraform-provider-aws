// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata

import (
	"context"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	"github.com/aws/aws-sdk-go-v2/service/directoryservicedata"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_directoryservicedata_user", sweepUsers)
}

func sweepUsers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	dsConn := client.DSClient(ctx)
	directoryServiceDataConn := client.DirectoryServiceDataClient(ctx)
	var sweepResources []sweep.Sweepable
	const acctestUserPrefix = "tfacctest"

	directoryPages := directoryservice.NewDescribeDirectoriesPaginator(dsConn, &directoryservice.DescribeDirectoriesInput{})
	for directoryPages.HasMorePages() {
		page, err := directoryPages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, directory := range page.DirectoryDescriptions {
			directoryID := aws.ToString(directory.DirectoryId)

			input := directoryservicedata.ListUsersInput{
				DirectoryId: aws.String(directoryID),
			}

			userPages := directoryservicedata.NewListUsersPaginator(directoryServiceDataConn, &input)

			for userPages.HasMorePages() {
				page, err := userPages.NextPage(ctx)
				if awsv2.SkipSweepError(err) {
					break
				}
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, user := range page.Users {
					samAccountName := aws.ToString(user.SAMAccountName)
					if !strings.HasPrefix(samAccountName, acctestUserPrefix) {
						continue
					}
					sweepResources = append(sweepResources, framework.NewSweepResource(newUserResource, client,
						framework.NewAttribute("directory_id", directoryID), framework.NewAttribute("sam_account_name", aws.ToString(user.SAMAccountName))),
					)
				}
			}
		}
	}
	return sweepResources, nil
}
