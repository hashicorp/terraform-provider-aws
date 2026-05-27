// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_lifecycle_hook", name="Lifecycle Hook")
// @IdentityAttribute("autoscaling_group_name")
// @IdentityAttribute("name")
// @ImportIDHandler("lifecycleHookImportID")
// @Testing(importStateIdFunc=testAccLifecycleHookImportStateIDFunc)
// @Testing(preIdentityVersion="v6.40.0")
func resourceLifecycleHook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLifecycleHookPut,
		ReadWithoutTimeout:   resourceLifecycleHookRead,
		UpdateWithoutTimeout: resourceLifecycleHookPut,
		DeleteWithoutTimeout: resourceLifecycleHookDelete,

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_result": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[lifecycleHookDefaultResult](),
			},
			"heartbeat_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(30, 7200),
			},
			"lifecycle_transition": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[lifecycleHookLifecycleTransition](),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`[A-Za-z0-9\-_\/]+`),
						`no spaces or special characters except "-", "_", and "/"`),
				),
			},
			"notification_metadata": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notification_target_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceLifecycleHookPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := autoscaling.PutLifecycleHookInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		LifecycleHookName:    aws.String(name),
	}

	if v, ok := d.GetOk("default_result"); ok {
		input.DefaultResult = aws.String(v.(string))
	}

	if v, ok := d.GetOk("heartbeat_timeout"); ok {
		input.HeartbeatTimeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("lifecycle_transition"); ok {
		input.LifecycleTransition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_metadata"); ok {
		input.NotificationMetadata = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_target_arn"); ok {
		input.NotificationTargetARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleARN = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 5*time.Minute,
		func(ctx context.Context) (any, error) {
			return conn.PutLifecycleHook(ctx, &input)
		},
		errCodeValidationError, "Unable to publish test message to notification target")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Auto Scaling Lifecycle Hook (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceLifecycleHookRead(ctx, d, meta)...)
}

func resourceLifecycleHookRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	p, err := findLifecycleHookByTwoPartKey(ctx, conn, d.Get("autoscaling_group_name").(string), d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Lifecycle Hook %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Lifecycle Hook (%s): %s", d.Id(), err)
	}

	d.Set("default_result", p.DefaultResult)
	d.Set("heartbeat_timeout", p.HeartbeatTimeout)
	d.Set("lifecycle_transition", p.LifecycleTransition)
	d.Set(names.AttrName, p.LifecycleHookName)
	d.Set("notification_metadata", p.NotificationMetadata)
	d.Set("notification_target_arn", p.NotificationTargetARN)
	d.Set(names.AttrRoleARN, p.RoleARN)

	return diags
}

func resourceLifecycleHookDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	log.Printf("[INFO] Deleting Auto Scaling Lifecycle Hook: %s", d.Id())
	input := autoscaling.DeleteLifecycleHookInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		LifecycleHookName:    aws.String(d.Id()),
	}
	_, err := conn.DeleteLifecycleHook(ctx, &input)

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "No Lifecycle Hook found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Lifecycle Hook (%s): %s", d.Id(), err)
	}

	return diags
}

func findLifecycleHook(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeLifecycleHooksInput) (*awstypes.LifecycleHook, error) {
	output, err := findLifecycleHooks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLifecycleHooks(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeLifecycleHooksInput) ([]awstypes.LifecycleHook, error) {
	output, err := conn.DescribeLifecycleHooks(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.LifecycleHooks, nil
}

func findLifecycleHookByTwoPartKey(ctx context.Context, conn *autoscaling.Client, asgName, hookName string) (*awstypes.LifecycleHook, error) {
	input := autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: aws.String(asgName),
		LifecycleHookNames:   []string{hookName},
	}

	return findLifecycleHook(ctx, conn, &input)
}

const lifecycleHookImportIDSeparator = "/"

func lifecycleHookParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, lifecycleHookImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected auto-scaling-group-name%[2]slifecycle-hook-name", id, lifecycleHookImportIDSeparator)
}

var (
	_ inttypes.SDKv2ImportID = lifecycleHookImportID{}
)

type lifecycleHookImportID struct{}

func (lifecycleHookImportID) Parse(id string) (string, map[string]any, error) {
	asgName, hookName, err := lifecycleHookParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"autoscaling_group_name": asgName,
		names.AttrName:           hookName,
	}

	return hookName, result, nil
}

func (lifecycleHookImportID) Create(d *schema.ResourceData) string {
	return d.Get(names.AttrName).(string)
}
