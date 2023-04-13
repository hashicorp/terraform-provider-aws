package events

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_event_endpoint", name="Endpoint")
func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate,
		Read:   resourceEndpointRead,
		Update: resourceEndpointUpdate,
		Delete: resourceEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
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

	log.Printf("[DEBUG] Creating EventBridge endpoint: %v", input)

	_, err := conn.CreateEndpoint(input)
	if err != nil {
		return fmt.Errorf("Creating EventBridge endpoint (%s) failed: %w", name, err)
	}

	d.SetId(name)

	pendingStates := []string{eventbridge.EndpointStateCreating}
	targetStates := []string{eventbridge.EndpointStateActive}
	_, err = waitEndpoint(conn, d.Id(), pendingStates, targetStates)
	if err != nil {
		return fmt.Errorf("error waiting for EventBridge endpoint (%s) to create: %w", d.Id(), err)
	}

	log.Printf("[INFO] EventBridge endpoint (%s) created", d.Id())

	return resourceEndpointRead(d, meta)
}

func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn()

	output, err := FindEndpointByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading EventBridge endpoint: %w", err)
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

	replicationConfig := make(map[string]interface{})
	isEnabled, err := EndpointReplicationEnabledFromState(aws.StringValue(output.ReplicationConfig.State))
	if err != nil {
		return fmt.Errorf("error getting Eventbridge endpoint replication state: %w", err)
	}

	replicationConfig["is_enabled"] = isEnabled
	d.Set("replication_config", []map[string]interface{}{replicationConfig})

	d.Set("routing_config", flatternRoutingConfig(output.RoutingConfig))
	return nil
}

func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
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

	log.Printf("[DEBUG] Updating EventBridge endpoint: %s", input)
	_, err := conn.UpdateEndpoint(input)
	if err != nil {
		return fmt.Errorf("error updating EventBridge endpoint (%s): %w", d.Id(), err)
	}

	pendingStates := []string{eventbridge.EndpointStateUpdating}
	targetStates := []string{eventbridge.EndpointStateActive}
	_, err = waitEndpoint(conn, d.Id(), pendingStates, targetStates)
	if err != nil {
		return fmt.Errorf("error waiting for EventBridge endpoint (%s) to create: %w", d.Id(), err)
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn()
	log.Printf("[INFO] Deleting EventBridge endpoint (%s)", d.Id())
	_, err := conn.DeleteEndpoint(&eventbridge.DeleteEndpointInput{
		Name: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge endpoint (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting EventBridge endpoint (%s): %w", d.Id(), err)
	}

	pendingStates := []string{eventbridge.EndpointStateDeleting}
	targetStates := []string{}
	_, err = waitEndpoint(conn, d.Id(), pendingStates, targetStates)
	if err != nil {
		return fmt.Errorf("error waiting for EventBridge endpoint (%s) to create: %w", d.Id(), err)
	}
	log.Printf("[INFO] EventBridge endpoint (%s) deleted", d.Id())

	return nil
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
