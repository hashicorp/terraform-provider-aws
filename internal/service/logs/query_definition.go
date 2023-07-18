// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_query_definition")
func resourceQueryDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueryDefinitionPut,
		ReadWithoutTimeout:   resourceQueryDefinitionRead,
		UpdateWithoutTimeout: resourceQueryDefinitionPut,
		DeleteWithoutTimeout: resourceQueryDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceQueryDefinitionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^([^:*\/]+\/?)*[^:*\/]+$`), "cannot contain a colon or asterisk and cannot start or end with a slash"),
				),
			},
			"log_group_names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validLogGroupName,
				},
			},
			"query_definition_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query_string": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceQueryDefinitionPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	name := d.Get("name").(string)
	input := &cloudwatchlogs.PutQueryDefinitionInput{
		Name:        aws.String(name),
		QueryString: aws.String(d.Get("query_string").(string)),
	}

	if v, ok := d.GetOk("log_group_names"); ok && len(v.([]interface{})) > 0 {
		input.LogGroupNames = flex.ExpandStringList(v.([]interface{}))
	}

	if !d.IsNewResource() {
		input.QueryDefinitionId = aws.String(d.Id())
	}

	output, err := conn.PutQueryDefinitionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("putting CloudWatch Logs Query Definition (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.StringValue(output.QueryDefinitionId))
	}

	return resourceQueryDefinitionRead(ctx, d, meta)
}

func resourceQueryDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	result, err := FindQueryDefinitionByTwoPartKey(ctx, conn, d.Get("name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Query Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Query Definition (%s): %s", d.Id(), err)
	}

	d.Set("log_group_names", aws.StringValueSlice(result.LogGroupNames))
	d.Set("name", result.Name)
	d.Set("query_definition_id", result.QueryDefinitionId)
	d.Set("query_string", result.QueryString)

	return nil
}

func resourceQueryDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Query Definition: %s", d.Id())
	_, err := conn.DeleteQueryDefinitionWithContext(ctx, &cloudwatchlogs.DeleteQueryDefinitionInput{
		QueryDefinitionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Query Definition (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceQueryDefinitionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	arn, err := arn.Parse(d.Id())
	if err != nil {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	if arn.Service != cloudwatchlogs.ServiceName {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	matcher := regexp.MustCompile("^query-definition:(" + verify.UUIDRegexPattern + ")$")
	matches := matcher.FindStringSubmatch(arn.Resource)
	if len(matches) != 2 {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected a CloudWatch query definition ARN", d.Id())
	}

	d.SetId(matches[1])

	return []*schema.ResourceData{d}, nil
}

func FindQueryDefinitionByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, name, queryDefinitionID string) (*cloudwatchlogs.QueryDefinition, error) {
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}
	if name != "" {
		input.QueryDefinitionNamePrefix = aws.String(name)
	}
	var output *cloudwatchlogs.QueryDefinition

	err := describeQueryDefinitionsPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeQueryDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.QueryDefinitions {
			if aws.StringValue(v.QueryDefinitionId) == queryDefinitionID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, err
}
