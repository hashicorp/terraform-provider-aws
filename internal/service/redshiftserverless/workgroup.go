// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_workgroup", name="Workgroup")
// @Tags(identifierAttribute="arn")
func ResourceWorkgroup() *schema.Resource {
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
			"arn": {
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
			"endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
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
												"availability_zone": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"network_interface_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"private_ip_address": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"vpc_endpoint_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"vpc_id": {
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
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkgroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	name := d.Get("workgroup_name").(string)
	input := redshiftserverless.CreateWorkgroupInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		Tags:          getTagsIn(ctx),
		WorkgroupName: aws.String(name),
	}

	if v, ok := d.GetOk("base_capacity"); ok {
		input.BaseCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("config_parameter"); ok && v.(*schema.Set).Len() > 0 {
		input.ConfigParameters = expandConfigParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("enhanced_vpc_routing"); ok {
		input.EnhancedVpcRouting = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		input.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateWorkgroupWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Workgroup (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Workgroup.WorkgroupName))

	if _, err := waitWorkgroupAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Serverless Workgroup (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	out, err := FindWorkgroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Workgroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Workgroup (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(out.WorkgroupArn)
	d.Set("arn", arn)
	d.Set("base_capacity", out.BaseCapacity)
	if err := d.Set("config_parameter", flattenConfigParameters(out.ConfigParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting config_parameter: %s", err)
	}
	if err := d.Set("endpoint", []interface{}{flattenEndpoint(out.Endpoint)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("enhanced_vpc_routing", out.EnhancedVpcRouting)
	d.Set("namespace_name", out.NamespaceName)
	d.Set("publicly_accessible", out.PubliclyAccessible)
	d.Set("security_group_ids", flex.FlattenStringSet(out.SecurityGroupIds))
	d.Set("subnet_ids", flex.FlattenStringSet(out.SubnetIds))
	d.Set("workgroup_id", out.WorkgroupId)
	d.Set("workgroup_name", out.WorkgroupName)

	return diags
}

func resourceWorkgroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	// You can't update multiple parameters in one request.

	if d.HasChange("base_capacity") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			BaseCapacity:  aws.Int64(int64(d.Get("base_capacity").(int))),
			WorkgroupName: aws.String(d.Id()),
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

	if d.HasChange("publicly_accessible") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			PubliclyAccessible: aws.Bool(d.Get("publicly_accessible").(bool)),
			WorkgroupName:      aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("security_group_ids") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
			WorkgroupName:    aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("subnet_ids") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			SubnetIds:     flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
			WorkgroupName: aws.String(d.Id()),
		}

		if err := updateWorkgroup(ctx, conn, input, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceWorkgroupRead(ctx, d, meta)...)
}

func resourceWorkgroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Workgroup: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 10*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteWorkgroupWithContext(ctx, &redshiftserverless.DeleteWorkgroupInput{
				WorkgroupName: aws.String(d.Id()),
			})
		},
		// "ConflictException: There is an operation running on the workgroup. Try deleting the workgroup again later."
		redshiftserverless.ErrCodeConflictException, "operation running")

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
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

func updateWorkgroup(ctx context.Context, conn *redshiftserverless.RedshiftServerless, input *redshiftserverless.UpdateWorkgroupInput, timeout time.Duration) error {
	name := aws.StringValue(input.WorkgroupName)
	_, err := conn.UpdateWorkgroupWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Redshift Serverless Workgroup (%s): %w", name, err)
	}

	if _, err := waitWorkgroupAvailable(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Redshift Serverless Workgroup (%s) update: %w", name, err)
	}

	return nil
}

func FindWorkgroupByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Workgroup, error) {
	input := &redshiftserverless.GetWorkgroupInput{
		WorkgroupName: aws.String(name),
	}

	output, err := conn.GetWorkgroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
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

func statusWorkgroup(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkgroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitWorkgroupAvailable(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string, wait time.Duration) (*redshiftserverless.Workgroup, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{redshiftserverless.WorkgroupStatusCreating, redshiftserverless.WorkgroupStatusModifying},
		Target:  []string{redshiftserverless.WorkgroupStatusAvailable},
		Refresh: statusWorkgroup(ctx, conn, name),
		Timeout: wait,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func waitWorkgroupDeleted(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string, wait time.Duration) (*redshiftserverless.Workgroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{redshiftserverless.WorkgroupStatusAvailable, redshiftserverless.WorkgroupStatusModifying, redshiftserverless.WorkgroupStatusDeleting},
		Target:  []string{},
		Refresh: statusWorkgroup(ctx, conn, name),
		Timeout: wait,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Workgroup); ok {
		return output, err
	}

	return nil, err
}

func expandConfigParameter(tfMap map[string]interface{}) *redshiftserverless.ConfigParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshiftserverless.ConfigParameter{}

	if v, ok := tfMap["parameter_key"].(string); ok {
		apiObject.ParameterKey = aws.String(v)
	}

	if v, ok := tfMap["parameter_value"].(string); ok {
		apiObject.ParameterValue = aws.String(v)
	}

	return apiObject
}

func expandConfigParameters(tfList []interface{}) []*redshiftserverless.ConfigParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*redshiftserverless.ConfigParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandConfigParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConfigParameter(apiObject *redshiftserverless.ConfigParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ParameterKey; v != nil {
		tfMap["parameter_key"] = aws.StringValue(v)
	}

	if v := apiObject.ParameterValue; v != nil {
		tfMap["parameter_value"] = aws.StringValue(v)
	}
	return tfMap
}

func flattenConfigParameters(apiObjects []*redshiftserverless.ConfigParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenConfigParameter(apiObject))
	}

	return tfList
}

func flattenEndpoint(apiObject *redshiftserverless.Endpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if v := apiObject.Address; v != nil {
		tfMap["address"] = aws.StringValue(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap["port"] = aws.Int64Value(v)
	}

	if v := apiObject.VpcEndpoints; v != nil {
		tfMap["vpc_endpoint"] = flattenVPCEndpoints(v)
	}

	return tfMap
}

func flattenVPCEndpoints(apiObjects []*redshiftserverless.VpcEndpoint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenVPCEndpoint(apiObject))
	}

	return tfList
}

func flattenVPCEndpoint(apiObject *redshiftserverless.VpcEndpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap["vpc_endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaces; v != nil {
		tfMap["network_interface"] = flattenNetworkInterfaces(v)
	}
	return tfMap
}

func flattenNetworkInterfaces(apiObjects []*redshiftserverless.NetworkInterface) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject *redshiftserverless.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}
	return tfMap
}
