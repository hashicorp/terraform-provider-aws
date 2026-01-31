// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_group", name="Log Group")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(destroyTakesT=true)
// @Testing(existsTakesT=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.LogGroup")
// @Testing(idAttrDuplicates="name")
// @Testing(preIdentityVersion="v6.7.0")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.HasChange("log_group_class") {
						return false
					}
					if v, ok := d.GetOk("log_group_class"); ok {
						if awstypes.LogGroupClass(v.(string)) == awstypes.LogGroupClassDelivery {
							return true
						}
					}
					return false
				},
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"force_destroy": {
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
	input := cloudwatchlogs.CreateLogGroupInput{
		LogGroupClass: awstypes.LogGroupClass(d.Get("log_group_class").(string)),
		LogGroupName:  aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("deletion_protection_enabled"); ok {
		input.DeletionProtectionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateLogGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Log Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("retention_in_days"); ok && input.LogGroupClass != awstypes.LogGroupClassDelivery {
		input := cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(d.Id()),
			RetentionInDays: aws.Int32(int32(v.(int))),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.PutRetentionPolicy(ctx, &input)
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

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	resourceGroupFlatten(ctx, d, *lg)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	if d.HasChange("retention_in_days") {
		if v, ok := d.GetOk("retention_in_days"); ok {
			input := cloudwatchlogs.PutRetentionPolicyInput{
				LogGroupName:    aws.String(d.Id()),
				RetentionInDays: aws.Int32(int32(v.(int))),
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
				return conn.PutRetentionPolicy(ctx, &input)
			}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		} else {
			input := cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DeleteRetentionPolicy(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("deletion_protection_enabled") {
		var deletionProtectionEnabled bool
		if v, ok := d.GetOk("deletion_protection_enabled"); ok {
			deletionProtectionEnabled = v.(bool)
		} else {
			deletionProtectionEnabled = false
		}
		loggroup, err := findLogGroupByName(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", d.Id(), err)
		}
		input := cloudwatchlogs.PutLogGroupDeletionProtectionInput{
			LogGroupIdentifier:        loggroup.LogGroupArn,
			DeletionProtectionEnabled: aws.Bool(deletionProtectionEnabled),
		}

		_, err = conn.PutLogGroupDeletionProtection(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Logs Log Group (%s) deletion protection: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrKMSKeyID) {
		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input := cloudwatchlogs.AssociateKmsKeyInput{
				KmsKeyId:     aws.String(v.(string)),
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.AssociateKmsKey(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating CloudWatch Logs Log Group (%s) KMS key: %s", d.Id(), err)
			}
		} else {
			input := cloudwatchlogs.DisassociateKmsKeyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DisassociateKmsKey(ctx, &input)

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

	forceDestroy := d.Get("force_destroy").(bool)
	logGroupName := d.Id()

	log.Printf("[INFO] Deleting CloudWatch Logs Log Group: %s", logGroupName)

	if forceDestroy {
		// If retention policy is set, we need to clear it first to ensure proper deletion
		if v, ok := d.GetOk("retention_in_days"); ok && v.(int) > 0 {
			input := cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(logGroupName),
			}
			_, err := conn.DeleteRetentionPolicy(ctx, &input)
			if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return sdkdiag.AppendErrorf(diags, "removing retention policy on CloudWatch Logs Log Group (%s): %s", logGroupName, err)
			}
		}
	}

	input := cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	}
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.OperationAbortedException](ctx, 1*time.Minute, func(ctx context.Context) (any, error) {
		return conn.DeleteLogGroup(ctx, &input)
	}, "try again")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	if forceDestroy {
		// Wait for log group to be fully deleted before removing from state
		err = waitLogGroupDeleted(ctx, conn, logGroupName)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Logs Log Group (%s) deletion: %s", logGroupName, err)
		}
	}

	return diags
}

// waitLogGroupDeleted waits for a log group to be fully deleted
func waitLogGroupDeleted(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{"Deleted"},
		Refresh: statusLogGroup(ctx, conn, logGroupName),
		Timeout: 5 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// statusLogGroup returns a StateRefreshFunc that checks log group deletion status
func statusLogGroup(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName string) retry.StateRefreshFunc {
	return func(context.Context) (any, string, error) {
		input := cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(logGroupName),
		}

		output, err := conn.DescribeLogGroups(ctx, &input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, "Deleted", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, lg := range output.LogGroups {
			if aws.ToString(lg.LogGroupName) == logGroupName {
				return lg, "Deleting", nil
			}
		}

		return nil, "Deleted", nil
	}
}

func findLogGroupByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.LogGroup, error) {
	input := cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
	}

	return findLogGroup(ctx, conn, &input, func(v *awstypes.LogGroup) bool {
		return aws.ToString(v.LogGroupName) == name
	}, tfslices.WithReturnFirstMatch)
}

func findLogGroup(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup], optFns ...tfslices.FinderOptionsFunc) (*awstypes.LogGroup, error) {
	opts := tfslices.NewFinderOptions(optFns)
	var output []awstypes.LogGroup
	for value, err := range listLogGroups(ctx, conn, input, filter) {
		if err != nil {
			return nil, err
		}

		output = append(output, value)
		if opts.ReturnFirstMatch() {
			break
		}
	}

	return tfresource.AssertSingleValueResult(output)
}

func resourceGroupFlatten(_ context.Context, d *schema.ResourceData, lg awstypes.LogGroup) {
	d.Set(names.AttrARN, trimLogGroupARNWildcardSuffix(aws.ToString(lg.Arn)))
	d.Set("deletion_protection_enabled", lg.DeletionProtectionEnabled)
	d.Set(names.AttrKMSKeyID, lg.KmsKeyId)
	d.Set("log_group_class", lg.LogGroupClass)
	d.Set(names.AttrName, lg.LogGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lg.LogGroupName)))
	d.Set("retention_in_days", lg.RetentionInDays)
	// Support in-place update of non-refreshable attributes.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))
	d.Set("force_destroy", d.Get("force_destroy"))
}
