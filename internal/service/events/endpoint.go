package events

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"event_buses": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"replication_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
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
												"health_check_arn": {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	name := d.Get("name").(string)
	input := &eventbridge.CreateEndpointInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_buses"); ok {
		busARNs := v.([]interface{})

		var buses []*eventbridge.EndpointEventBus
		for _, arn := range busARNs {
			buses = append(buses, &eventbridge.EndpointEventBus{
				EventBusArn: aws.String(arn.(string)),
			})
		}
		input.EventBuses = buses
	}

	if v, ok := d.GetOk("replication_config"); ok {
		config := v.([]interface{})
		for _, c := range config {
			param := c.(map[string]interface{})
			if val, ok := param["is_enabled"]; ok {
				isEnabled := val.(bool)
				input.ReplicationConfig = &eventbridge.ReplicationConfig{
					State: aws.String(EndpointReplicationStateFromEnabled(isEnabled)),
				}
			}
		}
	}

	if v, ok := d.GetOk("routing_config"); ok {
		routingConfig := v.([]interface{})
		input.RoutingConfig = expandRoutingConfig(routingConfig)
	}

	_, err := conn.CreateEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Global Endpoint (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitEndpointCreated(ctx, conn, d.Id(), 2*time.Minute); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	output, err := FindEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Global Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("endpoint_url", output.EndpointUrl)
	d.Set("role_arn", output.RoleArn)

	var eventBuses []string
	for _, endpointBus := range output.EventBuses {
		eventBuses = append(eventBuses, aws.StringValue(endpointBus.EventBusArn))
	}
	d.Set("event_buses", eventBuses)

	// TODO
	// replicationConfig := make(map[string]interface{})
	// isEnabled, err := EndpointReplicationEnabledFromState(aws.StringValue(output.ReplicationConfig.State))
	// if err != nil {
	// 	return fmt.Errorf("error getting Eventbridge endpoint replication state: %w", err)
	// }

	// replicationConfig["is_enabled"] = isEnabled
	// d.Set("replication_config", []map[string]interface{}{replicationConfig})

	if err := d.Set("routing_config", flatternRoutingConfig(output.RoutingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routing_config: %s", err)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

	input := &eventbridge.UpdateEndpointInput{
		Name: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_buses"); ok {
		busARNs := v.([]string)

		var buses []*eventbridge.EndpointEventBus
		for _, arn := range busARNs {
			buses = append(buses, &eventbridge.EndpointEventBus{
				EventBusArn: aws.String(arn),
			})
		}
		input.EventBuses = buses
	}

	if v, ok := d.GetOk("replication_config"); ok {
		config := v.([]interface{})
		for _, c := range config {
			param := c.(map[string]interface{})
			if val, ok := param["is_enabled"]; ok {
				isEnabled := val.(bool)
				input.ReplicationConfig = &eventbridge.ReplicationConfig{
					State: aws.String(EndpointReplicationStateFromEnabled(isEnabled)),
				}
			}
		}
	}

	if v, ok := d.GetOk("routing_config"); ok {
		routingConfig := v.([]interface{})
		input.RoutingConfig = expandRoutingConfig(routingConfig)
	}

	_, err := conn.UpdateEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Global Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointUpdated(ctx, conn, d.Id(), 2*time.Minute); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Global Endpoint (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn()

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

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), 2*time.Minute); err != nil {
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

	outputRaw, err := stateConf.WaitForState()

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

	outputRaw, err := stateConf.WaitForState()

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

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*eventbridge.DescribeEndpointOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func expandRoutingConfig(routingConfig []interface{}) *eventbridge.RoutingConfig {
	failover := eventbridge.FailoverConfig{}
	for _, c := range routingConfig {
		routingConfigParam := c.(map[string]interface{})
		if val, ok := routingConfigParam["failover_config"]; ok {
			failoverConfig := val.([]interface{})
			failover.Primary = expandRoutingConfigFailoverPrimary(failoverConfig)
			failover.Secondary = expandRoutingConfigFailoverSecondary(failoverConfig)
		}
	}

	return &eventbridge.RoutingConfig{
		FailoverConfig: &failover,
	}
}

func expandRoutingConfigFailoverPrimary(failoverConfig []interface{}) *eventbridge.Primary {
	failoverPrimary := eventbridge.Primary{}
	for _, fc := range failoverConfig {
		failoverConfigParam := fc.(map[string]interface{})
		if primaryVal, ok := failoverConfigParam["primary"]; ok {
			primaryConfig := primaryVal.([]interface{})
			for _, pc := range primaryConfig {
				primaryConfigParam := pc.(map[string]interface{})
				if val, ok := primaryConfigParam["health_check_arn"]; ok {
					failoverPrimary.HealthCheck = aws.String(val.(string))
				}
			}
		}
	}
	return &failoverPrimary

}
func expandRoutingConfigFailoverSecondary(failoverConfig []interface{}) *eventbridge.Secondary {
	failoverSecondary := eventbridge.Secondary{}
	for _, fc := range failoverConfig {
		failoverConfigParam := fc.(map[string]interface{})
		if secondayVal, ok := failoverConfigParam["secondary"]; ok {
			secondaryConfig := secondayVal.([]interface{})
			for _, sc := range secondaryConfig {
				secondaryConfigParam := sc.(map[string]interface{})
				if val, ok := secondaryConfigParam["route"]; ok {
					failoverSecondary.Route = aws.String(val.(string))
				}
			}
		}
	}
	return &failoverSecondary
}

func flatternRoutingConfig(config *eventbridge.RoutingConfig) []map[string]interface{} {
	routingConfig := make(map[string]interface{})

	if config.FailoverConfig != nil {
		failoverConfig := make(map[string]interface{})

		if config.FailoverConfig.Primary != nil {
			primaryConfig := make(map[string]interface{})
			if config.FailoverConfig.Primary.HealthCheck != nil {
				primaryConfig["health_check_arn"] = aws.StringValue(config.FailoverConfig.Primary.HealthCheck)
			}
			failoverConfig["primary"] = []map[string]interface{}{primaryConfig}
		}

		if config.FailoverConfig.Secondary != nil {
			secondayConfig := make(map[string]interface{})
			if config.FailoverConfig.Secondary.Route != nil {
				secondayConfig["route"] = aws.StringValue(config.FailoverConfig.Secondary.Route)
			}
			failoverConfig["secondary"] = []map[string]interface{}{secondayConfig}

		}
		routingConfig["failover_config"] = failoverConfig
	}

	result := []map[string]interface{}{routingConfig}
	return result
}
