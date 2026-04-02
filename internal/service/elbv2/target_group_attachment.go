// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package elbv2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb_target_group_attachment", name="Target Group Attachment")
// @SDKResource("aws_lb_target_group_attachment", name="Target Group Attachment")
// @IdentityAttribute("target_group_arn")
// @IdentityAttribute("target_id")
// @IdentityAttribute("port", valueType="int", optional="true", testNotNull="true")
// @IdentityAttribute("availability_zone", optional="true")
// @IdentityAttribute("quic_server_id", optional="true")
// @MutableIdentity
// @ImportIDHandler("targetGroupAttachmentImportID")
// @Testing(preIdentityVersion="v6.33.0")
func resourceTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentCreate,
		ReadWithoutTimeout:   resourceAttachmentRead,
		DeleteWithoutTimeout: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"quic_server_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"target_group_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("quic_server_id"); ok {
		input.Targets[0].QuicServerId = aws.String(v.(string))
	}

	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.InvalidTargetException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.RegisterTargets(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering ELBv2 Target Group (%s) target: %s", targetGroupARN, err)
	}

	d.SetId(targetGroupAttachmentImportID{}.Create(d))

	return diags
}

// resourceAttachmentRead requires all of the fields in order to describe the correct
// target, so there is no work to do beyond ensuring that the target and group still exist.
func resourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("quic_server_id"); ok {
		input.Targets[0].QuicServerId = aws.String(v.(string))
	}

	target, err := findTargetHealthDescription(ctx, conn, &input)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ELBv2 Target Group Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group Attachment (%s): %s", d.Id(), err)
	}

	if target == nil || target.Target == nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group Attachment (%s): target not found", d.Id())
	}

	d.Set("target_group_arn", targetGroupARN)
	d.Set("target_id", target.Target.Id)

	if rawConfig := d.GetRawConfig(); rawConfig.IsKnown() && !rawConfig.IsNull() {
		if rawPort := rawConfig.GetAttr(names.AttrPort); rawPort.IsKnown() && !rawPort.IsNull() {
			d.Set(names.AttrPort, target.Target.Port)
		}
		if rawAZ := rawConfig.GetAttr(names.AttrAvailabilityZone); rawAZ.IsKnown() && !rawAZ.IsNull() {
			d.Set(names.AttrAvailabilityZone, target.Target.AvailabilityZone)
		}
	}

	if v := aws.ToString(target.Target.QuicServerId); v != "" {
		d.Set("quic_server_id", v)
	}

	return diags
}

func resourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroupARN := d.Get("target_group_arn").(string)
	input := elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupARN),
		Targets: []awstypes.TargetDescription{{
			Id: aws.String(d.Get("target_id").(string)),
		}},
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.Targets[0].AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Targets[0].Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("quic_server_id"); ok {
		input.Targets[0].QuicServerId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting ELBv2 Target Group Attachment: %s", d.Id())
	_, err := conn.DeregisterTargets(ctx, &input)

	if errs.IsA[*awstypes.LoadBalancerNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering ELBv2 Target Group (%s) target: %s", targetGroupARN, err)
	}

	return diags
}

func findTargetHealthDescription(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetHealthInput) (*awstypes.TargetHealthDescription, error) {
	output, err := findTargetHealthDescriptions(ctx, conn, input, func(v *awstypes.TargetHealthDescription) bool {
		// This will catch targets being removed by hand (draining as we plan) or that have been removed for a while
		// without trying to re-create ones that are just not in use. For example, a target can be `unused` if the
		// target group isnt assigned to anything, a scenario where we don't want to continuously recreate the resource.
		if v := v.TargetHealth; v != nil {
			switch v.Reason {
			case awstypes.TargetHealthReasonEnumDeregistrationInProgress, awstypes.TargetHealthReasonEnumNotRegistered:
				return false
			default:
				return true
			}
		}

		return false
	})

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTargetHealthDescriptions(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetHealthInput, filter tfslices.Predicate[*awstypes.TargetHealthDescription]) ([]awstypes.TargetHealthDescription, error) {
	var targetHealthDescriptions []awstypes.TargetHealthDescription

	output, err := conn.DescribeTargetHealth(ctx, input)

	if errs.IsA[*awstypes.InvalidTargetException](err) || errs.IsA[*awstypes.TargetGroupNotFoundException](err) {
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

	for _, v := range output.TargetHealthDescriptions {
		if filter(&v) {
			targetHealthDescriptions = append(targetHealthDescriptions, v)
		}
	}

	return targetHealthDescriptions, nil
}

const targetGroupAttachmentResourceIDSeparator = ","

var _ inttypes.SDKv2ImportID = targetGroupAttachmentImportID{}

type targetGroupAttachmentImportID struct{}

func (targetGroupAttachmentImportID) Create(d *schema.ResourceData) string {
	parts := []string{
		d.Get("target_group_arn").(string),
		d.Get("target_id").(string),
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		parts = append(parts, strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		if len(parts) == 2 {
			parts = append(parts, "") // placeholder for port when only AZ is set
		}
		parts = append(parts, v.(string))
	}

	return strings.Join(parts, targetGroupAttachmentResourceIDSeparator)
}

func (targetGroupAttachmentImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, targetGroupAttachmentResourceIDSeparator)
	if len(parts) < 2 || len(parts) > 4 {
		return id, nil, fmt.Errorf("unexpected format for ID (%s), expected TARGET_GROUP_ARN,TARGET_ID[,PORT][,AVAILABILITY_ZONE]", id)
	}

	results := map[string]any{
		"target_group_arn": parts[0],
		"target_id":        parts[1],
	}

	if len(parts) >= 3 && parts[2] != "" {
		port, err := strconv.Atoi(parts[2])
		if err != nil {
			return id, nil, fmt.Errorf("parsing port: %w", err)
		}
		results[names.AttrPort] = port
	}

	if len(parts) >= 4 && parts[3] != "" {
		results[names.AttrAvailabilityZone] = parts[3]
	}

	return id, results, nil
}
