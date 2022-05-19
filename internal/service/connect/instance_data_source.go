package connect

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstanceRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_resolve_best_voices_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"identity_management_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"instance_alias", "instance_id"},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"instance_id", "instance_alias"},
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// "use_custom_tts_voices_enabled": {
			// 	Type:     schema.TypeBool,
			// 	Computed: true,
			// },
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	var matchedInstance *connect.Instance

	if v, ok := d.GetOk("instance_id"); ok {
		instanceId := v.(string)

		input := connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceId),
		}

		log.Printf("[DEBUG] Reading Connect Instance by instance_id: %s", input)

		output, err := conn.DescribeInstance(&input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting Connect Instance by instance_id (%s): %w", instanceId, err))
		}

		if output == nil {
			return diag.FromErr(fmt.Errorf("error getting Connect Instance by instance_id (%s): empty output", instanceId))
		}

		matchedInstance = output.Instance

	} else if v, ok := d.GetOk("instance_alias"); ok {
		instanceAlias := v.(string)

		instanceSummary, err := dataSourceGetInstanceSummaryByInstanceAlias(ctx, conn, instanceAlias)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Instance Summary by instance_alias (%s): %w", instanceAlias, err))
		}

		if instanceSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Instance Summary by instance_alias (%s): not found", instanceAlias))
		}

		matchedInstance = &connect.Instance{
			Arn:                    instanceSummary.Arn,
			CreatedTime:            instanceSummary.CreatedTime,
			Id:                     instanceSummary.Id,
			IdentityManagementType: instanceSummary.IdentityManagementType,
			InboundCallsEnabled:    instanceSummary.InboundCallsEnabled,
			InstanceAlias:          instanceSummary.InstanceAlias,
			InstanceStatus:         instanceSummary.InstanceStatus,
			OutboundCallsEnabled:   instanceSummary.OutboundCallsEnabled,
			ServiceRole:            instanceSummary.ServiceRole,
		}
	}

	if matchedInstance == nil {
		return diag.FromErr(fmt.Errorf("no Connect Instance found for query, try adjusting your search criteria"))
	}

	d.SetId(aws.StringValue(matchedInstance.Id))

	d.Set("arn", matchedInstance.Arn)
	d.Set("created_time", matchedInstance.CreatedTime.Format(time.RFC3339))
	d.Set("identity_management_type", matchedInstance.IdentityManagementType)
	d.Set("inbound_calls_enabled", matchedInstance.InboundCallsEnabled)
	d.Set("instance_alias", matchedInstance.InstanceAlias)
	d.Set("outbound_calls_enabled", matchedInstance.OutboundCallsEnabled)
	d.Set("service_role", matchedInstance.ServiceRole)
	d.Set("status", matchedInstance.InstanceStatus)

	for att := range InstanceAttributeMapping() {
		value, err := dataSourceInstanceReadAttribute(ctx, conn, d.Id(), att)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading Connect Instance (%s) attribute (%s): %w", d.Id(), att, err))
		}
		d.Set(InstanceAttributeMapping()[att], value)
	}

	return nil
}

func dataSourceGetInstanceSummaryByInstanceAlias(ctx context.Context, conn *connect.Connect, instanceAlias string) (*connect.InstanceSummary, error) {
	var result *connect.InstanceSummary

	input := &connect.ListInstancesInput{
		MaxResults: aws.Int64(ListInstancesMaxResults),
	}

	err := conn.ListInstancesPagesWithContext(ctx, input, func(page *connect.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, is := range page.InstanceSummaryList {
			if is == nil {
				continue
			}

			if aws.StringValue(is.InstanceAlias) == instanceAlias {
				result = is
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func dataSourceInstanceReadAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string) (bool, error) {
	input := &connect.DescribeInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
	}

	out, err := conn.DescribeInstanceAttributeWithContext(ctx, input)

	if err != nil {
		return false, err
	}

	result, parseErr := strconv.ParseBool(*out.Attribute.Value)
	return result, parseErr
}
