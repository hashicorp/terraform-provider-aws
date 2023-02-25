package route53

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	conn := meta.(*conns.AWSClient).Route53Conn()

	name := d.Get("name").(string)
	input := &route53.CreateTrafficPolicyInstanceInput{
		HostedZoneId:         aws.String(d.Get("hosted_zone_id").(string)),
		Name:                 aws.String(name),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	log.Printf("[INFO] Creating Route53 Traffic Policy Instance: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateTrafficPolicyInstanceWithContext(ctx, input)
	}, route53.ErrCodeNoSuchTrafficPolicy)

	if err != nil {
		return diag.Errorf("error creating Route53 Traffic Policy Instance (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*route53.CreateTrafficPolicyInstanceOutput).TrafficPolicyInstance.Id))

	if _, err = waitTrafficPolicyInstanceStateCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for Route53 Traffic Policy Instance (%s) create: %s", d.Id(), err)
	}

	return resourceTrafficPolicyInstanceRead(ctx, d, meta)
}

func resourceTrafficPolicyInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn()

	trafficPolicyInstance, err := FindTrafficPolicyInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Traffic Policy Instance %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	d.Set("hosted_zone_id", trafficPolicyInstance.HostedZoneId)
	d.Set("name", strings.TrimSuffix(aws.StringValue(trafficPolicyInstance.Name), "."))
	d.Set("traffic_policy_id", trafficPolicyInstance.TrafficPolicyId)
	d.Set("traffic_policy_version", trafficPolicyInstance.TrafficPolicyVersion)
	d.Set("ttl", trafficPolicyInstance.TTL)

	return nil
}

func resourceTrafficPolicyInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn()

	input := &route53.UpdateTrafficPolicyInstanceInput{
		Id:                   aws.String(d.Id()),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	log.Printf("[INFO] Updating Route53 Traffic Policy Instance: %s", input)
	_, err := conn.UpdateTrafficPolicyInstanceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	if _, err = waitTrafficPolicyInstanceStateUpdated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for Route53 Traffic Policy Instance (%s) update: %s", d.Id(), err)
	}

	return resourceTrafficPolicyInstanceRead(ctx, d, meta)
}

func resourceTrafficPolicyInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn()

	log.Printf("[INFO] Delete Route53 Traffic Policy Instance: %s", d.Id())
	_, err := conn.DeleteTrafficPolicyInstanceWithContext(ctx, &route53.DeleteTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	if _, err = waitTrafficPolicyInstanceStateDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("error waiting for Route53 Traffic Policy Instance (%s) delete: %s", d.Id(), err)
	}

	return nil
}
