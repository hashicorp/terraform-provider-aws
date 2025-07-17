// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_target_group", name="Target Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetGroupCreate,
		ReadWithoutTimeout:   resourceTargetGroupRead,
		UpdateWithoutTimeout: resourceTargetGroupUpdate,
		DeleteWithoutTimeout: resourceTargetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrHealthCheck: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"health_check_interval_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      30,
										ValidateFunc: validation.IntBetween(5, 300),
									},
									"health_check_timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      5,
										ValidateFunc: validation.IntBetween(1, 120),
									},
									"healthy_threshold_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      5,
										ValidateFunc: validation.IntBetween(2, 10),
									},
									"matcher": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrValue: {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "200",
												},
											},
										},
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
									},
									names.AttrPath: {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/",
									},
									names.AttrPort: {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsPortNumber,
									},
									names.AttrProtocol: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[types.TargetGroupProtocol](),
									},
									"protocol_version": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.HealthCheckProtocolVersionHttp1,
										StateFunc:        sdkv2.ToUpperSchemaStateFunc,
										ValidateDiagFunc: enum.Validate[types.HealthCheckProtocolVersion](),
									},
									"unhealthy_threshold_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      2,
										ValidateFunc: validation.IntBetween(2, 10),
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
						names.AttrIPAddressType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.IpAddressType](),
						},
						"lambda_event_structure_version": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.LambdaEventStructureVersion](),
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.TargetGroupProtocol](),
						},
						"protocol_version": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							StateFunc:        sdkv2.ToUpperSchemaStateFunc,
							ValidateDiagFunc: enum.Validate[types.TargetGroupProtocolVersion](),
						},
						"vpc_identifier": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.TargetGroupType](),
			},
		},
	}
}

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := vpclattice.CreateTargetGroupInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
		Type:        types.TargetGroupType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("config"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Config = expandTargetGroupConfig(v.([]any)[0].(map[string]any))
	}

	out, err := conn.CreateTargetGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Target Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitTargetGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPCLattice Target Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findTargetGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Lattice Target Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Target Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	if out.Config != nil {
		if err := d.Set("config", []any{flattenTargetGroupConfig(out.Config)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting config: %s", err)
		}
	} else {
		d.Set("config", nil)
	}
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrType, out.Type)

	return diags
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := vpclattice.UpdateTargetGroupInput{
			TargetGroupIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("config") {
			if v, ok := d.GetOk("config"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				config := expandTargetGroupConfig(v.([]any)[0].(map[string]any))

				if v := config.HealthCheck; v != nil {
					input.HealthCheck = v
				}
			}
		}

		if input.HealthCheck == nil {
			return diags
		}

		_, err := conn.UpdateTargetGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating VPCLattice Target Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPC Lattice Target Group: %s", d.Id())
	input := vpclattice.DeleteTargetGroupInput{
		TargetGroupIdentifier: aws.String(d.Id()),
	}

	// Draining the targets can take a moment, so we need to retry on conflict.
	_, err := tfresource.RetryWhenIsA[*types.ConflictException](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteTargetGroup(ctx, &input)
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Target Group (%s): %s", d.Id(), err)
	}

	if _, err := waitTargetGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPCLattice Target Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTargetGroupByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetTargetGroupOutput, error) {
	input := vpclattice.GetTargetGroupInput{
		TargetGroupIdentifier: aws.String(id),
	}

	return findTargetGroup(ctx, conn, &input)
}

func findTargetGroup(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetTargetGroupInput) (*vpclattice.GetTargetGroupOutput, error) {
	output, err := conn.GetTargetGroup(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func statusTargetGroup(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTargetGroupByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTargetGroupCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetGroupStatusCreateInProgress),
		Target:                    enum.Slice(types.TargetGroupStatusActive),
		Refresh:                   statusTargetGroup(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetTargetGroupOutput); ok {
		if output.Status == types.TargetGroupStatusCreateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitTargetGroupDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TargetGroupStatusDeleteInProgress, types.TargetGroupStatusActive),
		Target:  []string{},
		Refresh: statusTargetGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetTargetGroupOutput); ok {
		if output.Status == types.TargetGroupStatusDeleteFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func flattenTargetGroupConfig(apiObject *types.TargetGroupConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrIPAddressType:          apiObject.IpAddressType,
		"lambda_event_structure_version": apiObject.LambdaEventStructureVersion,
		names.AttrProtocol:               apiObject.Protocol,
		"protocol_version":               apiObject.ProtocolVersion,
	}

	if v := apiObject.HealthCheck; v != nil {
		tfMap[names.AttrHealthCheck] = []any{flattenHealthCheckConfig(v)}
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.VpcIdentifier; v != nil {
		tfMap["vpc_identifier"] = aws.ToString(v)
	}

	return tfMap
}

func flattenHealthCheckConfig(apiObject *types.HealthCheckConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrProtocol: apiObject.Protocol,
		"protocol_version": apiObject.ProtocolVersion,
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.HealthCheckIntervalSeconds; v != nil {
		tfMap["health_check_interval_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.HealthCheckTimeoutSeconds; v != nil {
		tfMap["health_check_timeout_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.HealthyThresholdCount; v != nil {
		tfMap["healthy_threshold_count"] = aws.ToInt32(v)
	}

	if v := apiObject.Matcher; v != nil {
		tfMap["matcher"] = []any{flattenMatcherMemberHTTPCode(v.(*types.MatcherMemberHttpCode))}
	}

	if v := apiObject.Path; v != nil {
		tfMap[names.AttrPath] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.UnhealthyThresholdCount; v != nil {
		tfMap["unhealthy_threshold_count"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenMatcherMemberHTTPCode(apiObject *types.MatcherMemberHttpCode) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrValue: apiObject.Value,
	}

	return tfMap
}

func expandTargetGroupConfig(tfMap map[string]any) *types.TargetGroupConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetGroupConfig{}

	if v, ok := tfMap[names.AttrHealthCheck].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.HealthCheck = expandHealthCheckConfig(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrIPAddressType].(string); ok && v != "" {
		apiObject.IpAddressType = types.IpAddressType(v)
	}

	if v, ok := tfMap["lambda_event_structure_version"].(string); ok && v != "" {
		apiObject.LambdaEventStructureVersion = types.LambdaEventStructureVersion(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.TargetGroupProtocol(v)
	}

	if v, ok := tfMap["protocol_version"].(string); ok && v != "" {
		apiObject.ProtocolVersion = types.TargetGroupProtocolVersion(v)
	}

	if v, ok := tfMap["vpc_identifier"].(string); ok && v != "" {
		apiObject.VpcIdentifier = aws.String(v)
	}

	return apiObject
}

func expandHealthCheckConfig(tfMap map[string]any) *types.HealthCheckConfig {
	apiObject := &types.HealthCheckConfig{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["health_check_interval_seconds"].(int); ok && v != 0 {
		apiObject.HealthCheckIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["health_check_timeout_seconds"].(int); ok && v != 0 {
		apiObject.HealthCheckTimeoutSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["healthy_threshold_count"].(int); ok && v != 0 {
		apiObject.HealthyThresholdCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["matcher"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Matcher = expandMatcherMemberHTTPCode(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrPath].(string); ok && v != "" {
		apiObject.Path = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
		apiObject.Port = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		apiObject.Protocol = types.TargetGroupProtocol(v)
	}

	if v, ok := tfMap["protocol_version"].(string); ok && v != "" {
		apiObject.ProtocolVersion = types.HealthCheckProtocolVersion(v)
	}

	if v, ok := tfMap["unhealthy_threshold_count"].(int); ok && v != 0 {
		apiObject.UnhealthyThresholdCount = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMatcherMemberHTTPCode(tfMap map[string]any) types.Matcher {
	apiObject := &types.MatcherMemberHttpCode{}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = v
	}

	return apiObject
}
