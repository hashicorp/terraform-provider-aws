// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_workgroup", name="Workgroup")
// @Tags(identifierAttribute="arn")
func resourceWorkgroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkgroupCreate,
		ReadWithoutTimeout:   resourceWorkgroupRead,
		UpdateWithoutTimeout: resourceWorkgroupUpdate,
		DeleteWithoutTimeout: resourceWorkgroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"config_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_key": {
							Type: schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{
								// https://docs.aws.amazon.com/redshift-serverless/latest/APIReference/API_CreateWorkgroup.html#redshiftserverless-CreateWorkgroup-request-configParameters
								"auto_mv",
								"datestyle",
								"enable_case_sensitive_identifier", // "ValidationException: The parameter key enable_case_sensitivity_identifier isn't supported. Supported values: [[max_query_cpu_usage_percent, max_join_row_count, auto_mv, max_query_execution_time, max_query_queue_time, max_query_blocks_read, max_return_row_count, search_path, datestyle, max_query_cpu_time, max_io_skew, max_scan_row_count, query_group, enable_user_activity_logging, enable_case_sensitive_identifier, max_nested_loop_join_row_count, max_query_temp_blocks_to_disk, max_cpu_skew]]"
								"enable_user_activity_logging",
								"query_group",
								"search_path",
								// https://docs.aws.amazon.com/redshift/latest/dg/cm-c-wlm-query-monitoring-rules.html#cm-c-wlm-query-monitoring-metrics-serverless
								"max_query_cpu_time",
								"max_query_blocks_read",
								"max_scan_row_count",
								"max_query_execution_time",
								"max_query_queue_time",
								"max_query_cpu_usage_percent",
								"max_query_temp_blocks_to_disk",
								"max_join_row_count",
								"max_nested_loop_join_row_count",
								// default SSL parameters automatically added by AWS
								// https://docs.aws.amazon.com/redshift/latest/mgmt/connecting-ssl-support.html
								"require_ssl",
								"use_fips_ssl",
							}, false),
							Required: true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vpc_endpoint": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_interface": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrAvailabilityZone: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrNetworkInterfaceID: {
													Type:     schema.TypeString,
													Computed: true,
												},
												"private_ip_address": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrSubnetID: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrVPCEndpointID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrVPCID: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrMaxCapacity: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"price_performance_target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"level": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntInSlice([]int{1, 25, 50, 75, 100}),
						},
					},
				},
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"track_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_]+$`), "must be alphanumeric or underscore"),
				),
			},
			"workgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkgroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	name := d.Get("workgroup_name").(string)
	input := redshiftserverless.CreateWorkgroupInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		Tags:          getTagsIn(ctx),
		WorkgroupName: aws.String(name),
	}

	if v, ok := d.GetOk("base_capacity"); ok {
		input.BaseCapacity = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("config_parameter"); ok && v.(*schema.Set).Len() > 0 {
		input.ConfigParameters = expandConfigParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("enhanced_vpc_routing"); ok {
		input.EnhancedVpcRouting = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
		input.MaxCapacity = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrPort); ok {
		input.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("price_performance_target"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.PricePerformanceTarget = expandPerformanceTarget(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrPubliclyAccessible); ok {
		input.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSubnetIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("track_name"); ok {
		input.TrackName = aws.String(v.(string))
	}

	if input.BaseCapacity != nil && input.PricePerformanceTarget != nil && input.PricePerformanceTarget.Status == awstypes.PerformanceTargetStatusEnabled {
		return sdkdiag.AppendErrorf(diags, "base_capacity cannot be set when price_performance_target.enabled is true")
	}

	output, err := conn.CreateWorkgroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Workgroup (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Workgroup.WorkgroupName))

	if _, err := waitWorkgroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Workgroup (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	out, err := findWorkgroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Workgroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Workgroup (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.WorkgroupArn)
	d.Set("base_capacity", out.BaseCapacity)
	if err := d.Set("config_parameter", flattenConfigParameters(out.ConfigParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting config_parameter: %s", err)
	}
	if err := d.Set(names.AttrEndpoint, []any{flattenEndpoint(out.Endpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("enhanced_vpc_routing", out.EnhancedVpcRouting)
	d.Set(names.AttrMaxCapacity, out.MaxCapacity)
	d.Set("namespace_name", out.NamespaceName)
	d.Set(names.AttrPort, flattenEndpoint(out.Endpoint)[names.AttrPort])
	if err := d.Set("price_performance_target", flattenPerformanceTarget(out.PricePerformanceTarget)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting price_performance_target: %s", err)
	}
	d.Set(names.AttrPubliclyAccessible, out.PubliclyAccessible)
	d.Set(names.AttrSecurityGroupIDs, out.SecurityGroupIds)
	d.Set(names.AttrSubnetIDs, out.SubnetIds)
	d.Set("track_name", out.TrackName)
	d.Set("workgroup_id", out.WorkgroupId)
	d.Set("workgroup_name", out.WorkgroupName)

	return diags
}

func resourceWorkgroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	checkCapacityChange := func(key string) (bool, int, int) {
		o, n := d.GetChange(key)
		oldCapacity, newCapacity := o.(int), n.(int)
		hasCapacityChange := newCapacity != oldCapacity
		return hasCapacityChange, oldCapacity, newCapacity
	}

	// You can't update multiple workgroup parameters in one request.
	// This is particularly important when adjusting base_capacity and max_capacity due to their interdependencies:
	// - base_capacity cannot be increased to a value greater than the current max_capacity.
	// - max_capacity cannot be decreased to a value smaller than the current base_capacity.
	// The value 0 of max_capacity in the state signifies "not set".
	// Sending max_capacity value of -1 to AWS API removes max_capacity limit, but -1 cannot be used as max_capacity in the state,
	// because AWS API never returns -1 as the value of unset max_capacity. There would be a diff on subsequent apply,
	// resulting in errors due to the lack of AWS API idempotency.
	// Some validations, such as increasing base_capacity beyond an unchanged max_capacity, are deferred to the AWS API.

	hasBaseCapacityChange, _, newBaseCapacity := checkCapacityChange("base_capacity")
	hasMaxCapacityChange, oldMaxCapacity, newMaxCapacity := checkCapacityChange(names.AttrMaxCapacity)

	switch {
	case hasMaxCapacityChange && newMaxCapacity == 0:
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{MaxCapacity: aws.Int32(-1), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	case hasBaseCapacityChange && hasMaxCapacityChange && (oldMaxCapacity == 0 || newBaseCapacity <= oldMaxCapacity):
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{BaseCapacity: aws.Int32(int32(newBaseCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{MaxCapacity: aws.Int32(int32(newMaxCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	case hasBaseCapacityChange && hasMaxCapacityChange && newBaseCapacity > oldMaxCapacity:
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{MaxCapacity: aws.Int32(int32(newMaxCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{BaseCapacity: aws.Int32(int32(newBaseCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	case hasBaseCapacityChange:
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{BaseCapacity: aws.Int32(int32(newBaseCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	case hasMaxCapacityChange:
		if err :=
			updateWorkgroup(ctx, conn,
				&redshiftserverless.UpdateWorkgroupInput{MaxCapacity: aws.Int32(int32(newMaxCapacity)), WorkgroupName: aws.String(d.Id())},
				d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("price_performance_target") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			PricePerformanceTarget: expandPerformanceTarget(d.Get("price_performance_target").([]any)),
			WorkgroupName:          aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("config_parameter") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			ConfigParameters: expandConfigParameters(d.Get("config_parameter").(*schema.Set).List()),
			WorkgroupName:    aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("enhanced_vpc_routing") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			EnhancedVpcRouting: aws.Bool(d.Get("enhanced_vpc_routing").(bool)),
			WorkgroupName:      aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrPort) {
		input := &redshiftserverless.UpdateWorkgroupInput{
			Port:          aws.Int32(int32(d.Get(names.AttrPort).(int))),
			WorkgroupName: aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrPubliclyAccessible) {
		input := &redshiftserverless.UpdateWorkgroupInput{
			PubliclyAccessible: aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			WorkgroupName:      aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrSecurityGroupIDs) {
		input := &redshiftserverless.UpdateWorkgroupInput{
			SecurityGroupIds: flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroupIDs).(*schema.Set)),
			WorkgroupName:    aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrSubnetIDs) {
		input := &redshiftserverless.UpdateWorkgroupInput{
			SubnetIds:     flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
			WorkgroupName: aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("track_name") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			TrackName:     aws.String(d.Get("track_name").(string)),
			WorkgroupName: aws.String(d.Id()),
		}
		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Workgroup: %s", d.Id())
	const (
		retryTimeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ConflictException](ctx, retryTimeout,
		func(ctx context.Context) (any, error) {
			return conn.DeleteWorkgroup(ctx, &redshiftserverless.DeleteWorkgroupInput{
				WorkgroupName: aws.String(d.Id()),
			})
		},
		// "ConflictException: There is an operation running on the workgroup. Try deleting the workgroup again later."
		"operation running")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Workgroup (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkgroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Workgroup (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func updateWorkgroup(ctx context.Context, conn *redshiftserverless.Client, input *redshiftserverless.UpdateWorkgroupInput, timeout time.Duration) error {
	const (
		retryTimeout = 20 * time.Minute
	)
	_, err := tfresource.RetryWhen(ctx, retryTimeout,
		func(ctx context.Context) (any, error) {
			return conn.UpdateWorkgroup(ctx, input)
		}, func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "operation running") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Wait until at least one snapshot exists and retry") {
				return true, err
			}

			return false, err
		})

	name := aws.ToString(input.WorkgroupName)

	if err != nil {
		return fmt.Errorf("updating Redshift Serverless Workgroup (%s): %w", name, err)
	}

	if _, err := waitWorkgroupAvailable(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Redshift Serverless Workgroup (%s) update: %w", name, err)
	}

	return nil
}

func findWorkgroupByName(ctx context.Context, conn *redshiftserverless.Client, name string) (*awstypes.Workgroup, error) {
	input := &redshiftserverless.GetWorkgroupInput{
		WorkgroupName: aws.String(name),
	}

	output, err := conn.GetWorkgroup(ctx, input)

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

	return output.Workgroup, nil
}

func statusWorkgroup(ctx context.Context, conn *redshiftserverless.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findWorkgroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitWorkgroupAvailable(ctx context.Context, conn *redshiftserverless.Client, name string, wait time.Duration) (*awstypes.Workgroup, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkgroupStatusCreating, awstypes.WorkgroupStatusModifying),
		Target:  enum.Slice(awstypes.WorkgroupStatusAvailable),
		Refresh: statusWorkgroup(ctx, conn, name),
		Timeout: wait,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func waitWorkgroupDeleted(ctx context.Context, conn *redshiftserverless.Client, name string, wait time.Duration) (*awstypes.Workgroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkgroupStatusAvailable, awstypes.WorkgroupStatusModifying, awstypes.WorkgroupStatusDeleting),
		Target:  []string{},
		Refresh: statusWorkgroup(ctx, conn, name),
		Timeout: wait,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func expandPerformanceTarget(tfList []any) *awstypes.PerformanceTarget {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PerformanceTarget{}

	// bool is a nicer way to represent the enabled/disabled state but the API expects
	// a string enumeration.
	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		if v {
			apiObject.Status = awstypes.PerformanceTargetStatusEnabled
		} else {
			apiObject.Status = awstypes.PerformanceTargetStatusDisabled
		}
	}

	if v, ok := tfMap["level"].(int); ok && v != 0 {
		apiObject.Level = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenPerformanceTarget(apiObject *awstypes.PerformanceTarget) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Status == awstypes.PerformanceTargetStatusEnabled,
	}

	if v := apiObject.Level; v != nil {
		tfMap["level"] = aws.ToInt32(v)
	}

	return []any{tfMap}
}

func expandConfigParameter(tfMap map[string]any) awstypes.ConfigParameter {
	apiObject := awstypes.ConfigParameter{}

	if v, ok := tfMap["parameter_key"].(string); ok {
		apiObject.ParameterKey = aws.String(v)
	}

	if v, ok := tfMap["parameter_value"].(string); ok {
		apiObject.ParameterValue = aws.String(v)
	}

	return apiObject
}

func expandConfigParameters(tfList []any) []awstypes.ConfigParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ConfigParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandConfigParameter(tfMap))
	}

	return apiObjects
}

func flattenConfigParameter(apiObject awstypes.ConfigParameter) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.ParameterKey; v != nil {
		tfMap["parameter_key"] = aws.ToString(v)
	}

	if v := apiObject.ParameterValue; v != nil {
		tfMap["parameter_value"] = aws.ToString(v)
	}

	return tfMap
}

func flattenConfigParameters(apiObjects []awstypes.ConfigParameter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenConfigParameter(apiObject))
	}

	return tfList
}

func flattenEndpoint(apiObject *awstypes.Endpoint) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Address; v != nil {
		tfMap[names.AttrAddress] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	if v := apiObject.VpcEndpoints; v != nil {
		tfMap["vpc_endpoint"] = flattenVPCEndpoints(v)
	}

	return tfMap
}

func flattenVPCEndpoints(apiObjects []awstypes.VpcEndpoint) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVPCEndpoint(&apiObject))
	}

	return tfList
}

func flattenVPCEndpoint(apiObject *awstypes.VpcEndpoint) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap[names.AttrVPCEndpointID] = aws.ToString(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.ToString(v)
	}

	if v := apiObject.NetworkInterfaces; v != nil {
		tfMap["network_interface"] = flattenNetworkInterfaces(v)
	}

	return tfMap
}

func flattenNetworkInterfaces(apiObjects []awstypes.NetworkInterface) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject awstypes.NetworkInterface) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap[names.AttrAvailabilityZone] = aws.ToString(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.ToString(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(v)
	}

	return tfMap
}
