// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_event_endpoint", name="Global Endpoint")
func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"name": {
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
						"state": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      eventbridge.ReplicationStateEnabled,
							ValidateFunc: validation.StringInSlice(eventbridge.ReplicationState_Values(), false),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"role_arn": {
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
												"health_check": {
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
	const (
		timeout = 2 * time.Minute
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	name := d.Get("name").(string)
	input := &eventbridge.CreateEndpointInput{
		EventBuses:    expandEndpointEventBuses(d.Get("event_bus").([]interface{})),
		RoutingConfig: expandRoutingConfig(d.Get("routing_config").([]interface{})[0].(map[string]interface{})),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("replication_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ReplicationConfig = expandReplicationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateEndpointWithContext(ctx, input)
	}, "ValidationException", "cannot be assumed by principal")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Global Endpoint (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitEndpointCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	output, err := FindEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Global Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("description", output.Description)
	d.Set("endpoint_url", output.EndpointUrl)
	if err := d.Set("event_bus", flattenEndpointEventBuses(output.EventBuses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_bus: %s", err)
	}
	d.Set("name", output.Name)
	if output.ReplicationConfig != nil {
		if err := d.Set("replication_config", []interface{}{flattenReplicationConfig(output.ReplicationConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting replication_config: %s", err)
		}
	} else {
		d.Set("replication_config", nil)
	}
	d.Set("role_arn", output.RoleArn)
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
	const (
		timeout = 2 * time.Minute
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := &eventbridge.UpdateEndpointInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("event_bus") {
		input.EventBuses = expandEndpointEventBuses(d.Get("event_bus").([]interface{}))
	}

	if d.HasChange("replication_config") {
		if v, ok := d.GetOk("replication_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ReplicationConfig = expandReplicationConfig(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if d.HasChange("role_arn") {
		input.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	if d.HasChange("routing_config") {
		input.RoutingConfig = expandRoutingConfig(d.Get("routing_config").([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.UpdateEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointUpdated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		timeout = 2 * time.Minute
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	log.Printf("[INFO] Deleting EventBridge Global Endpoint: %s", d.Id())
	_, err := conn.DeleteEndpointWithContext(ctx, &eventbridge.DeleteEndpointInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindEndpointByName(ctx context.Context, conn *eventbridge.EventBridge, name string) (*eventbridge.DescribeEndpointOutput, error) {
	input := &eventbridge.DescribeEndpointInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeEndpointWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
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

func statusEndpointState(ctx context.Context, conn *eventbridge.EventBridge, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitEndpointCreated(ctx context.Context, conn *eventbridge.EventBridge, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventbridge.EndpointStateCreating},
		Target:  []string{eventbridge.EndpointStateActive},
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitEndpointUpdated(ctx context.Context, conn *eventbridge.EventBridge, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventbridge.EndpointStateUpdating},
		Target:  []string{eventbridge.EndpointStateActive},
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitEndpointDeleted(ctx context.Context, conn *eventbridge.EventBridge, name string, timeout time.Duration) (*eventbridge.DescribeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventbridge.EndpointStateDeleting},
		Target:  []string{},
		Refresh: statusEndpointState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func expandEndpointEventBus(tfMap map[string]interface{}) *eventbridge.EndpointEventBus {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.EndpointEventBus{}

	if v, ok := tfMap["event_bus_arn"].(string); ok && v != "" {
		apiObject.EventBusArn = aws.String(v)
	}

	return apiObject
}

func expandEndpointEventBuses(tfList []interface{}) []*eventbridge.EndpointEventBus {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*eventbridge.EndpointEventBus

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndpointEventBus(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandReplicationConfig(tfMap map[string]interface{}) *eventbridge.ReplicationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.ReplicationConfig{}

	if v, ok := tfMap["state"].(string); ok && v != "" {
		apiObject.State = aws.String(v)
	}

	return apiObject
}

func expandRoutingConfig(tfMap map[string]interface{}) *eventbridge.RoutingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.RoutingConfig{}

	if v, ok := tfMap["failover_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.FailoverConfig = expandFailoverConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFailoverConfig(tfMap map[string]interface{}) *eventbridge.FailoverConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.FailoverConfig{}

	if v, ok := tfMap["primary"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Primary = expandPrimary(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["secondary"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Secondary = expandSecondary(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrimary(tfMap map[string]interface{}) *eventbridge.Primary {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.Primary{}

	if v, ok := tfMap["health_check"].(string); ok && v != "" {
		apiObject.HealthCheck = aws.String(v)
	}

	return apiObject
}

func expandSecondary(tfMap map[string]interface{}) *eventbridge.Secondary {
	if tfMap == nil {
		return nil
	}

	apiObject := &eventbridge.Secondary{}

	if v, ok := tfMap["route"].(string); ok && v != "" {
		apiObject.Route = aws.String(v)
	}

	return apiObject
}

func flattenEndpointEventBus(apiObject *eventbridge.EndpointEventBus) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EventBusArn; v != nil {
		tfMap["event_bus_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEndpointEventBuses(apiObjects []*eventbridge.EndpointEventBus) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEndpointEventBus(apiObject))
	}

	return tfList
}

func flattenReplicationConfig(apiObject *eventbridge.ReplicationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.State; v != nil {
		tfMap["state"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenRoutingConfig(apiObject *eventbridge.RoutingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FailoverConfig; v != nil {
		tfMap["failover_config"] = []interface{}{flattenFailoverConfig(v)}
	}

	return tfMap
}

func flattenFailoverConfig(apiObject *eventbridge.FailoverConfig) map[string]interface{} {
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

func flattenPrimary(apiObject *eventbridge.Primary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HealthCheck; v != nil {
		tfMap["health_check"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSecondary(apiObject *eventbridge.Secondary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Route; v != nil {
		tfMap["route"] = aws.StringValue(v)
	}

	return tfMap
}
