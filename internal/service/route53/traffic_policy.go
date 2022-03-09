package route53

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

func ResourceTrafficPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficPolicyCreate,
		ReadWithoutTimeout:   resourceTrafficPolicyRead,
		UpdateWithoutTimeout: resourceTrafficPolicyUpdate,
		DeleteWithoutTimeout: resourceTrafficPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected traffic-policy-id/traffic-policy-version", d.Id())
				}
				version, err := strconv.Atoi(idParts[1])
				if err != nil {
					return nil, fmt.Errorf("cannot convert to int: %s", idParts[1])
				}
				d.Set("version", version)
				d.SetId(idParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"document": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 102400),
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

func resourceTrafficPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.CreateTrafficPolicyInput{
		Document: aws.String(d.Get("document").(string)),
		Name:     aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		input.Comment = aws.String(v.(string))
	}

	var err error
	var output *route53.CreateTrafficPolicyOutput
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err = conn.CreateTrafficPolicyWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
				resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateTrafficPolicyWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("error creating Route53 traffic policy: %s", err)
	}

	d.SetId(aws.StringValue(output.TrafficPolicy.Id))

	return resourceTrafficPolicyRead(ctx, d, meta)
}

func resourceTrafficPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	object, err := FindTrafficPolicyById(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("error getting Route53 Traffic Policy %s from ListTrafficPolicies: %s", d.Get("name").(string), err)
	}

	if object == nil {
		log.Printf("[WARN] Route53 Traffic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	request := &route53.GetTrafficPolicyInput{
		Id:      aws.String(d.Id()),
		Version: object.LatestVersion,
	}

	response, err := conn.GetTrafficPolicy(request)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
		log.Printf("[WARN] Route53 Traffic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Route53 Traffic Policy %s, version %d: %s", d.Get("name").(string), d.Get("version").(int), err)
	}

	d.Set("comment", response.TrafficPolicy.Comment)
	d.Set("document", response.TrafficPolicy.Document)
	d.Set("name", response.TrafficPolicy.Name)
	d.Set("type", response.TrafficPolicy.Type)
	d.Set("version", response.TrafficPolicy.Version)

	return nil
}

func resourceTrafficPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.CreateTrafficPolicyVersionInput{
		Id:       aws.String(d.Id()),
		Document: aws.String(d.Get("document").(string)),
	}

	if d.HasChange("comment") {
		input.Comment = aws.String(d.Get("comment").(string))
	}

	_, err := conn.CreateTrafficPolicyVersionWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error updating Route53 Traffic Policy: %s. %s", d.Get("name").(string), err)
	}

	return resourceTrafficPolicyRead(ctx, d, meta)
}

func resourceTrafficPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	var trafficPolicies []*route53.TrafficPolicy
	var versionMarker *string

	for allPoliciesListed := false; !allPoliciesListed; {
		listRequest := &route53.ListTrafficPolicyVersionsInput{
			Id: aws.String(d.Id()),
		}
		if versionMarker != nil {
			listRequest.TrafficPolicyVersionMarker = versionMarker
		}

		listResponse, err := conn.ListTrafficPolicyVersionsWithContext(ctx, listRequest)
		if err != nil {
			return diag.Errorf("error listing Route 53 Traffic Policy versions: %v", err)
		}

		trafficPolicies = append(trafficPolicies, listResponse.TrafficPolicies...)

		if aws.BoolValue(listResponse.IsTruncated) {
			versionMarker = listResponse.TrafficPolicyVersionMarker
		} else {
			allPoliciesListed = true
		}
	}

	for _, trafficPolicy := range trafficPolicies {
		input := &route53.DeleteTrafficPolicyInput{
			Id:      trafficPolicy.Id,
			Version: trafficPolicy.Version,
		}

		_, err := conn.DeleteTrafficPolicyWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
				return nil
			}

			return diag.Errorf("error deleting Route53 Traffic Policy %s, version %d: %s", aws.StringValue(trafficPolicy.Id), aws.Int64Value(trafficPolicy.Version), err)
		}
	}
	return nil
}
