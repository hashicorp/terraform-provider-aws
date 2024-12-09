// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_target_group", name="Target Group")
// @Tags(identifierAttribute="arn")
func ResourceTargetGroup() *schema.Resource {
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
										Type:     schema.TypeString,
										Optional: true,
										Default:  types.HealthCheckProtocolVersionHttp1,
										StateFunc: func(v interface{}) string {
											return strings.ToUpper(v.(string))
										},
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
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
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.TargetGroupType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameTargetGroup = "Target Group"
)

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	in := &vpclattice.CreateTargetGroupInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
		Type:        types.TargetGroupType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Config = expandTargetGroupConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateTargetGroup(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameService, name, err)
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitTargetGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForCreation, ResNameTargetGroup, d.Id(), err)
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := FindTargetGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VpcLattice Target Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameTargetGroup, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	if out.Config != nil {
		if err := d.Set("config", []interface{}{flattenTargetGroupConfig(out.Config)}); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionSetting, ResNameTargetGroup, d.Id(), err)
		}
	} else {
		d.Set("config", nil)
	}
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrType, out.Type)

	return diags
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &vpclattice.UpdateTargetGroupInput{
			TargetGroupIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("config") {
			if v, ok := d.GetOk("config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				config := expandTargetGroupConfig(v.([]interface{})[0].(map[string]interface{}))

				if v := config.HealthCheck; v != nil {
					in.HealthCheck = v
				}
			}
		}

		if in.HealthCheck == nil {
			return diags
		}

		out, err := conn.UpdateTargetGroup(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionUpdating, ResNameTargetGroup, d.Id(), err)
		}

		if _, err := waitTargetGroupUpdated(ctx, conn, aws.ToString(out.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForUpdate, ResNameTargetGroup, d.Id(), err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VpcLattice TargetGroup: %s", d.Id())
	_, err := conn.DeleteTargetGroup(ctx, &vpclattice.DeleteTargetGroupInput{
		TargetGroupIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameTargetGroup, d.Id(), err)
	}

	if _, err := waitTargetGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameTargetGroup, d.Id(), err)
	}

	return diags
}

func waitTargetGroupCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.CreateTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetGroupStatusCreateInProgress),
		Target:                    enum.Slice(types.TargetGroupStatusActive),
		Refresh:                   statusTargetGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.CreateTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTargetGroupUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.UpdateTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetGroupStatusCreateInProgress),
		Target:                    enum.Slice(types.TargetGroupStatusActive),
		Refresh:                   statusTargetGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.UpdateTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTargetGroupDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.DeleteTargetGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TargetGroupStatusDeleteInProgress, types.TargetGroupStatusActive),
		Target:  []string{},
		Refresh: statusTargetGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.DeleteTargetGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func statusTargetGroup(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindTargetGroupByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindTargetGroupByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetTargetGroupOutput, error) {
	in := &vpclattice.GetTargetGroupInput{
		TargetGroupIdentifier: aws.String(id),
	}
	out, err := conn.GetTargetGroup(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenTargetGroupConfig(apiObject *types.TargetGroupConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrIPAddressType:          apiObject.IpAddressType,
		"lambda_event_structure_version": apiObject.LambdaEventStructureVersion,
		names.AttrProtocol:               apiObject.Protocol,
		"protocol_version":               apiObject.ProtocolVersion,
	}

	if v := apiObject.HealthCheck; v != nil {
		tfMap[names.AttrHealthCheck] = []interface{}{flattenHealthCheckConfig(v)}
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.VpcIdentifier; v != nil {
		tfMap["vpc_identifier"] = aws.ToString(v)
	}

	return tfMap
}

func flattenHealthCheckConfig(apiObject *types.HealthCheckConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
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
		tfMap["matcher"] = []interface{}{flattenMatcherMemberHTTPCode(v.(*types.MatcherMemberHttpCode))}
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

func flattenMatcherMemberHTTPCode(apiObject *types.MatcherMemberHttpCode) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrValue: apiObject.Value,
	}

	return tfMap
}

func expandTargetGroupConfig(tfMap map[string]interface{}) *types.TargetGroupConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetGroupConfig{}

	if v, ok := tfMap[names.AttrHealthCheck].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.HealthCheck = expandHealthCheckConfig(v[0].(map[string]interface{}))
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

func expandHealthCheckConfig(tfMap map[string]interface{}) *types.HealthCheckConfig {
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

	if v, ok := tfMap["matcher"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Matcher = expandMatcherMemberHTTPCode(v[0].(map[string]interface{}))
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

func expandMatcherMemberHTTPCode(tfMap map[string]interface{}) types.Matcher {
	apiObject := &types.MatcherMemberHttpCode{}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}
