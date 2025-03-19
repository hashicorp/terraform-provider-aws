// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_group", name="Log Group")
// @Tags(identifierAttribute="arn")
// @Testing(destroyTakesT=true)
// @Testing(existsTakesT=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.LogGroup")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_group_class": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LogGroupClass](),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validLogGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validLogGroupNamePrefix,
			},
			"retention_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653}),
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupClass: awstypes.LogGroupClass(d.Get("log_group_class").(string)),
		LogGroupName:  aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateLogGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Log Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("retention_in_days"); ok {
		input := &cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(d.Id()),
			RetentionInDays: aws.Int32(int32(v.(int))),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
			return conn.PutRetentionPolicy(ctx, input)
		}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	lg, err := findLogGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, trimLogGroupARNWildcardSuffix(aws.ToString(lg.Arn)))
	d.Set(names.AttrKMSKeyID, lg.KmsKeyId)
	d.Set("log_group_class", lg.LogGroupClass)
	d.Set(names.AttrName, lg.LogGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lg.LogGroupName)))
	d.Set("retention_in_days", lg.RetentionInDays)
	// Support in-place update of non-refreshable attribute.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	if d.HasChange("retention_in_days") {
		if v, ok := d.GetOk("retention_in_days"); ok {
			input := &cloudwatchlogs.PutRetentionPolicyInput{
				LogGroupName:    aws.String(d.Id()),
				RetentionInDays: aws.Int32(int32(v.(int))),
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
				return conn.PutRetentionPolicy(ctx, input)
			}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		} else {
			input := &cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DeleteRetentionPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrKMSKeyID) {
		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input := &cloudwatchlogs.AssociateKmsKeyInput{
				KmsKeyId:     aws.String(v.(string)),
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.AssociateKmsKey(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating CloudWatch Logs Log Group (%s) KMS key: %s", d.Id(), err)
			}
		} else {
			input := &cloudwatchlogs.DisassociateKmsKeyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DisassociateKmsKey(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating CloudWatch Logs Log Group (%s) KMS key: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining CloudWatch Logs Log Group: %s", d.Id())
		return diags
	}

	log.Printf("[INFO] Deleting CloudWatch Logs Log Group: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.OperationAbortedException](ctx, 1*time.Minute, func() (any, error) {
		return conn.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(d.Id()),
		})
	}, "try again")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findLogGroupByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.LogGroup, error) {
	input := cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
	}

	return findLogGroup(ctx, conn, &input, func(v *awstypes.LogGroup) bool {
		return aws.ToString(v.LogGroupName) == name
	})
}

func findLogGroup(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup]) (*awstypes.LogGroup, error) {
	output, err := findLogGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLogGroups(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup]) ([]awstypes.LogGroup, error) {
	var output []awstypes.LogGroup

	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.LogGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
