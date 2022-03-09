package route53

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceTrafficPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"traffic_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	object, err := FindTrafficPolicyById(ctx, conn, d.Get("traffic_policy_id").(string))
	if err != nil {
		return diag.Errorf("error getting Route53 Traffic Policy %s from ListTrafficPolicies: %s", d.Get("name").(string), err)
	}

	request := &route53.GetTrafficPolicyInput{
		Id:      object.Id,
		Version: object.LatestVersion,
	}

	response, err := conn.GetTrafficPolicy(request)

	if err != nil {
		return diag.Errorf("error getting Route53 Traffic Policy %s: %s", d.Get("name").(string), err)
	}

	d.Set("comment", response.TrafficPolicy.Comment)
	d.Set("document", response.TrafficPolicy.Document)
	d.Set("name", response.TrafficPolicy.Name)
	d.Set("type", response.TrafficPolicy.Type)
	d.Set("version", response.TrafficPolicy.Version)

	d.SetId(aws.StringValue(response.TrafficPolicy.Id))

	return nil
}
