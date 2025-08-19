package arcregionswitch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_arcregionswitch_plan", name="Plan")
func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlanRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},

			"wait_for_health_checks": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Wait for Route53 health check IDs to be populated (takes ~4 minutes)",
			},

			"route53_health_checks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"health_check_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_approach": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workflow": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workflow_target_action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"workflow_target_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"workflow_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"step": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: dataSourceStepSchema(),
							},
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"primary_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_time_objective_minutes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"associated_alarms": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"alarm_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cross_account_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"trigger": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"conditions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"associated_alarm_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"condition": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"min_delay_minutes_between_executions": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"target_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	arn := d.Get("arn").(string)
	plan, err := FindPlanByARN(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("reading ARC Region Switch Plan (%s): %s", arn, err)
	}

	d.SetId(arn)
	d.Set("arn", plan.Arn)
	d.Set("name", plan.Name)
	d.Set("execution_role", plan.ExecutionRole)
	d.Set("recovery_approach", plan.RecoveryApproach)
	d.Set("regions", plan.Regions)
	d.Set("description", plan.Description)
	d.Set("primary_region", plan.PrimaryRegion)
	d.Set("recovery_time_objective_minutes", plan.RecoveryTimeObjectiveMinutes)
	d.Set("owner", plan.Owner)
	if plan.UpdatedAt != nil {
		d.Set("updated_at", plan.UpdatedAt.Format(time.RFC3339))
	}
	d.Set("version", plan.Version)

	if err := d.Set("workflow", flattenWorkflows(plan.Workflows)); err != nil {
		return diag.Errorf("setting workflow: %s", err)
	}

	if err := d.Set("associated_alarms", flattenAssociatedAlarms(plan.AssociatedAlarms)); err != nil {
		return diag.Errorf("setting associated_alarms: %s", err)
	}

	if err := d.Set("trigger", flattenTriggers(plan.Triggers)); err != nil {
		return diag.Errorf("setting trigger: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return diag.Errorf("listing tags for ARC Region Switch Plan (%s): %s", arn, err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	// Fetch Route53 health checks for this plan
	// Health check IDs are populated asynchronously ~4 minutes after plan creation
	if d.Get("wait_for_health_checks").(bool) {
		// Wait for health check IDs to be populated (takes ~4 minutes)
		timeout := d.Timeout(schema.TimeoutRead)
		err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
			healthChecks, err := listRoute53HealthChecks(ctx, conn, arn)
			if err != nil {
				return retry.NonRetryableError(err)
			}

			// Check if all health check IDs are populated
			for _, hc := range healthChecks {
				if aws.ToString(hc.HealthCheckId) == "" {
					return retry.RetryableError(fmt.Errorf("waiting for Route53 health check IDs to be populated"))
				}
			}

			if err := d.Set("route53_health_checks", flattenRoute53HealthChecks(healthChecks)); err != nil {
				return retry.NonRetryableError(fmt.Errorf("setting route53_health_checks: %s", err))
			}

			return nil
		})
		if err != nil {
			return diag.Errorf("waiting for Route53 health checks: %s", err)
		}
	} else {
		// Fetch health checks without waiting
		healthChecks, err := listRoute53HealthChecks(ctx, conn, arn)
		if err != nil {
			return diag.Errorf("listing Route53 health checks: %s", err)
		}
		if err := d.Set("route53_health_checks", flattenRoute53HealthChecks(healthChecks)); err != nil {
			return diag.Errorf("setting route53_health_checks: %s", err)
		}
	}

	return nil
}

func dataSourceStepSchema() map[string]*schema.Schema {
	stepSchema := stepSchema()

	// Convert all fields to computed and clear validation constraints
	for _, field := range stepSchema {
		field.Required = false
		field.Optional = false
		field.Computed = true
		field.ValidateFunc = nil
		field.MaxItems = 0
		field.MinItems = 0

		// Recursively handle nested resources
		if field.Elem != nil {
			if resource, ok := field.Elem.(*schema.Resource); ok {
				for _, nestedField := range resource.Schema {
					nestedField.Required = false
					nestedField.Optional = false
					nestedField.Computed = true
					nestedField.ValidateFunc = nil
					nestedField.MaxItems = 0
					nestedField.MinItems = 0
				}
			}
		}
	}

	return stepSchema
}

func listRoute53HealthChecks(ctx context.Context, conn *arcregionswitch.Client, planArn string) ([]types.Route53HealthCheck, error) {
	input := &arcregionswitch.ListRoute53HealthChecksInput{
		Arn: aws.String(planArn),
	}

	var healthChecks []types.Route53HealthCheck
	paginator := arcregionswitch.NewListRoute53HealthChecksPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		healthChecks = append(healthChecks, page.HealthChecks...)
	}

	return healthChecks, nil
}

func flattenRoute53HealthChecks(healthChecks []types.Route53HealthCheck) []interface{} {
	if len(healthChecks) == 0 {
		return nil
	}

	var result []interface{}
	for _, hc := range healthChecks {
		result = append(result, map[string]interface{}{
			"health_check_id": aws.ToString(hc.HealthCheckId),
			"hosted_zone_id":  aws.ToString(hc.HostedZoneId),
			"record_name":     aws.ToString(hc.RecordName),
			"region":          aws.ToString(hc.Region),
		})
	}
	return result
}
