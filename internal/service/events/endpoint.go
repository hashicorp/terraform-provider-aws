// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_endpoint", name="Global Endpoint")
func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event_bus": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 2,
				MinItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_bus_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{1,64}$`), "Maximum of 64 characters consisting of numbers, lower/upper case letters, .,-,_."),
			},
			"replication_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrState: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ReplicationStateEnabled,
							ValidateDiagFunc: enum.Validate[types.ReplicationState](),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"routing_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failover_config": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrHealthCheck: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"secondary": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"route": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidRegionName,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &eventbridge.CreateEndpointInput{
		EventBuses:    expandEndpointEventBuses(d.Get("event_bus").([]interface{})),
		Name:          aws.String(name),
		RoutingConfig: expandRoutingConfig(d.Get("routing_config").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("replication_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ReplicationConfig = expandReplicationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateEndpoint(ctx, input)
	}, errCodeValidationException, "cannot be assumed by principal")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Global Endpoint (%s): %s", name, err)
	}

	d.SetId(name)

	const (
		timeout = 2 * time.Minute
	)
	if _, err := waitEndpointCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	output, err := findEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Global Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("endpoint_url", output.EndpointUrl)
	if err := d.Set("event_bus", flattenEndpointEventBuses(output.EventBuses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_bus: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	if output.ReplicationConfig != nil {
		if err := d.Set("replication_config", []interface{}{flattenReplicationConfig(output.ReplicationConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting replication_config: %s", err)
		}
	} else {
		d.Set("replication_config", nil)
	}
	d.Set(names.AttrRoleARN, output.RoleArn)
	if output.RoutingConfig != nil {
		if err := d.Set("routing_config", []interface{}{flattenRoutingConfig(output.RoutingConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting routing_config: %s", err)
		}
	} else {
		d.Set("routing_config", nil)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.UpdateEndpointInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("event_bus") {
		input.EventBuses = expandEndpointEventBuses(d.Get("event_bus").([]interface{}))
	}

	if d.HasChange("replication_config") {
		if v, ok := d.GetOk("replication_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ReplicationConfig = expandReplicationConfig(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if d.HasChange(names.AttrRoleARN) {
		input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
	}

	if d.HasChange("routing_config") {
		input.RoutingConfig = expandRoutingConfig(d.Get("routing_config").([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	const (
		timeout = 2 * time.Minute
	)
	if _, err := waitEndpointUpdated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Global Endpoint: %s", d.Id())
	_, err := conn.DeleteEndpoint(ctx, &eventbridge.DeleteEndpointInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	const (
		timeout = 2 * time.Minute
	)
	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEndpointByName(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeEndpointOutput, error) {
	input := &eventbridge.DescribeEndpointInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeEndpoint(ctx, input)

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

func statusEndpointState(ctx context.Context, conn *eventbridge.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEndpointByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitEndpointCreated(ctx context.Context, conn *eventbridge.Client, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EndpointStateCreating),
		Target:  enum.Slice(types.EndpointStateActive),
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitEndpointUpdated(ctx context.Context, conn *eventbridge.Client, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EndpointStateUpdating),
		Target:  enum.Slice(types.EndpointStateActive),
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitEndpointDeleted(ctx context.Context, conn *eventbridge.Client, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EndpointStateDeleting),
		Target:  []string{},
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func expandEndpointEventBus(tfMap map[string]interface{}) *types.EndpointEventBus {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EndpointEventBus{}

	if v, ok := tfMap["event_bus_arn"].(string); ok && v != "" {
		apiObject.EventBusArn = aws.String(v)
	}

	return apiObject
}

func expandEndpointEventBuses(tfList []interface{}) []types.EndpointEventBus {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EndpointEventBus

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndpointEventBus(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandReplicationConfig(tfMap map[string]interface{}) *types.ReplicationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ReplicationConfig{}

	if v, ok := tfMap[names.AttrState].(string); ok && v != "" {
		apiObject.State = types.ReplicationState(v)
	}

	return apiObject
}

func expandRoutingConfig(tfMap map[string]interface{}) *types.RoutingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RoutingConfig{}

	if v, ok := tfMap["failover_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.FailoverConfig = expandFailoverConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFailoverConfig(tfMap map[string]interface{}) *types.FailoverConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.FailoverConfig{}

	if v, ok := tfMap["primary"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Primary = expandPrimary(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["secondary"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Secondary = expandSecondary(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrimary(tfMap map[string]interface{}) *types.Primary {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Primary{}

	if v, ok := tfMap[names.AttrHealthCheck].(string); ok && v != "" {
		apiObject.HealthCheck = aws.String(v)
	}

	return apiObject
}

func expandSecondary(tfMap map[string]interface{}) *types.Secondary {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Secondary{}

	if v, ok := tfMap["route"].(string); ok && v != "" {
		apiObject.Route = aws.String(v)
	}

	return apiObject
}

func flattenEndpointEventBus(apiObject *types.EndpointEventBus) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventBusArn; v != nil {
		tfMap["event_bus_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEndpointEventBuses(apiObjects []types.EndpointEventBus) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEndpointEventBus(&apiObject))
	}

	return tfList
}

func flattenReplicationConfig(apiObject *types.ReplicationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrState: apiObject.State,
	}

	return tfMap
}

func flattenRoutingConfig(apiObject *types.RoutingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailoverConfig; v != nil {
		tfMap["failover_config"] = []interface{}{flattenFailoverConfig(v)}
	}

	return tfMap
}

func flattenFailoverConfig(apiObject *types.FailoverConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Primary; v != nil {
		tfMap["primary"] = []interface{}{flattenPrimary(v)}
	}

	if v := apiObject.Secondary; v != nil {
		tfMap["secondary"] = []interface{}{flattenSecondary(v)}
	}

	return tfMap
}

func flattenPrimary(apiObject *types.Primary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HealthCheck; v != nil {
		tfMap[names.AttrHealthCheck] = aws.ToString(v)
	}

	return tfMap
}

func flattenSecondary(apiObject *types.Secondary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Route; v != nil {
		tfMap["route"] = aws.ToString(v)
	}

	return tfMap
}
