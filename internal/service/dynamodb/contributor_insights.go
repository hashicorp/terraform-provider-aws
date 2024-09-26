// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_contributor_insights", name="Contributor Insights")
func resourceContributorInsights() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContributorInsightsCreate,
		ReadWithoutTimeout:   resourceContributorInsightsRead,
		DeleteWithoutTimeout: resourceContributorInsightsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"index_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceContributorInsightsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName := d.Get(names.AttrTableName).(string)
	input := &dynamodb.UpdateContributorInsightsInput{
		ContributorInsightsAction: awstypes.ContributorInsightsActionEnable,
		TableName:                 aws.String(tableName),
	}

	var indexName string
	if v, ok := d.GetOk("index_name"); ok {
		indexName = v.(string)
		input.IndexName = aws.String(indexName)
	}

	_, err := conn.UpdateContributorInsights(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Contributor Insights for table (%s): %s", tableName, err)
	}

	d.SetId(contributorInsightsCreateResourceID(tableName, indexName, meta.(*conns.AWSClient).AccountID))

	if _, err := waitContributorInsightsCreated(ctx, conn, tableName, indexName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Contributor Insights (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceContributorInsightsRead(ctx, d, meta)...)
}

func resourceContributorInsightsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName, indexName, err := contributorInsightsParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findContributorInsightsByTwoPartKey(ctx, conn, tableName, indexName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB Contributor Insights (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Contributor Insights (%s): %s", d.Id(), err)
	}

	d.Set("index_name", output.IndexName)
	d.Set(names.AttrTableName, output.TableName)

	return diags
}

func resourceContributorInsightsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	tableName, indexName, err := contributorInsightsParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &dynamodb.UpdateContributorInsightsInput{
		ContributorInsightsAction: awstypes.ContributorInsightsActionDisable,
		TableName:                 aws.String(tableName),
	}

	if indexName != "" {
		input.IndexName = aws.String(indexName)
	}

	log.Printf("[INFO] Deleting DynamoDB Contributor Insights: %s", d.Id())
	_, err = conn.UpdateContributorInsights(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Contributor Insights (%s): %s", d.Id(), err)
	}

	if _, err := waitContributorInsightsDeleted(ctx, conn, tableName, indexName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Contributor Insights (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const contributorInsightsResourceIDSeparator = "/"

func contributorInsightsCreateResourceID(tableName, indexName, accountID string) string {
	return fmt.Sprintf("name:%s/index:%s/%s", tableName, indexName, accountID)
}

func contributorInsightsParseResourceID(id string) (string, string, error) {
	idParts := strings.Split(id, contributorInsightsResourceIDSeparator)
	if len(idParts) != 3 || idParts[0] == "" || idParts[2] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected table_name%[2]sindex_name%[2]saccount_id", id, contributorInsightsResourceIDSeparator)
	}

	tableName := strings.TrimPrefix(idParts[0], "name:")
	indexName := strings.TrimPrefix(idParts[1], "index:")

	return tableName, indexName, nil
}

func findContributorInsightsByTwoPartKey(ctx context.Context, conn *dynamodb.Client, tableName, indexName string) (*dynamodb.DescribeContributorInsightsOutput, error) {
	input := &dynamodb.DescribeContributorInsightsInput{
		TableName: aws.String(tableName),
	}
	if indexName != "" {
		input.IndexName = aws.String(indexName)
	}

	output, err := findContributorInsights(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.ContributorInsightsStatus; status == awstypes.ContributorInsightsStatusDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findContributorInsights(ctx context.Context, conn *dynamodb.Client, input *dynamodb.DescribeContributorInsightsInput) (*dynamodb.DescribeContributorInsightsOutput, error) {
	output, err := conn.DescribeContributorInsights(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusContributorInsights(ctx context.Context, conn *dynamodb.Client, tableName, indexName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findContributorInsightsByTwoPartKey(ctx, conn, tableName, indexName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.ContributorInsightsStatus), nil
	}
}

func waitContributorInsightsCreated(ctx context.Context, conn *dynamodb.Client, tableName, indexName string, timeout time.Duration) (*dynamodb.DescribeContributorInsightsOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ContributorInsightsStatusEnabling),
		Target:  enum.Slice(awstypes.ContributorInsightsStatusEnabled),
		Timeout: timeout,
		Refresh: statusContributorInsights(ctx, conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.DescribeContributorInsightsOutput); ok {
		if status, failureException := output.ContributorInsightsStatus, output.FailureException; status == awstypes.ContributorInsightsStatusFailed && failureException != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(failureException.ExceptionName), aws.ToString(failureException.ExceptionDescription)))
		}

		return output, err
	}

	return nil, err
}

func waitContributorInsightsDeleted(ctx context.Context, conn *dynamodb.Client, tableName, indexName string, timeout time.Duration) (*dynamodb.DescribeContributorInsightsOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ContributorInsightsStatusDisabling),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusContributorInsights(ctx, conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.DescribeContributorInsightsOutput); ok {
		if status, failureException := output.ContributorInsightsStatus, output.FailureException; status == awstypes.ContributorInsightsStatusFailed && failureException != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(failureException.ExceptionName), aws.ToString(failureException.ExceptionDescription)))
		}

		return output, err
	}

	return nil, err
}
