package route53

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceTrafficPolicyInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficPolicyInstanceCreate,
		ReadWithoutTimeout:   resourceTrafficPolicyInstanceRead,
		UpdateWithoutTimeout: resourceTrafficPolicyInstanceUpdate,
		DeleteWithoutTimeout: resourceTrafficPolicyInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"hosted_zone_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
				StateFunc: func(v interface{}) string {
					value := strings.TrimSuffix(v.(string), ".")
					return strings.ToLower(value)
				},
			},
			"traffic_policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
			},
			"traffic_policy_version": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 1000),
			},
			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtMost(2147483647),
			},
		},
	}
}

func resourceTrafficPolicyInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.CreateTrafficPolicyInstanceInput{
		HostedZoneId:         aws.String(d.Get("hosted_zone_id").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	var err error
	var output *route53.CreateTrafficPolicyInstanceOutput
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err = conn.CreateTrafficPolicyInstanceWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateTrafficPolicyInstanceWithContext(ctx, input)
	}
	if err != nil {
		return diag.Errorf("error creating Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	if _, err = waitTrafficPolicyInstanceStateApplied(ctx, conn, aws.StringValue(output.TrafficPolicyInstance.Id)); err != nil {
		return diag.Errorf("error waiting for Route53 Traffic Policy Instance (%s) to be Applied: %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.TrafficPolicyInstance.Id))

	return resourceTrafficPolicyInstanceRead(ctx, d, meta)
}

func resourceTrafficPolicyInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.GetTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetTrafficPolicyInstanceWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
		log.Printf("[WARN] Route53 Traffic Policy Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	d.Set("hosted_zone_id", output.TrafficPolicyInstance.HostedZoneId)
	d.Set("message", output.TrafficPolicyInstance.Message)
	d.Set("name", strings.TrimSuffix(aws.StringValue(output.TrafficPolicyInstance.Name), "."))
	d.Set("traffic_policy_id", output.TrafficPolicyInstance.TrafficPolicyId)
	d.Set("traffic_policy_version", output.TrafficPolicyInstance.TrafficPolicyVersion)
	d.Set("ttl", output.TrafficPolicyInstance.TTL)

	return nil
}

func resourceTrafficPolicyInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.UpdateTrafficPolicyInstanceInput{
		Id:                   aws.String(d.Id()),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	_, err := conn.UpdateTrafficPolicyInstanceWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error updating Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	return resourceTrafficPolicyInstanceRead(ctx, d, meta)
}

func resourceTrafficPolicyInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.DeleteTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrafficPolicyInstanceWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
			return nil
		}
		return diag.Errorf("error deleting Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	if _, err = waitTrafficPolicyInstanceStateDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
			return nil
		}
		return diag.Errorf("error waiting for Route53 Traffic Policy Instance (%s) to be Deleted: %s", d.Id(), err)
	}

	return nil
}
