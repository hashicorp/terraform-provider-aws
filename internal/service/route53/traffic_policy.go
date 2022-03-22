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
				const idSeparator = "/"
				parts := strings.Split(d.Id(), idSeparator)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected TRAFFIC-POLICY-ID%[2]sTRAFFIC-POLICY-VERSION", d.Id(), idSeparator)
				}

				version, err := strconv.Atoi(parts[1])

				if err != nil {
					return nil, err
				}

				d.SetId(parts[0])
				d.Set("version", version)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
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

	name := d.Get("name").(string)
	input := &route53.CreateTrafficPolicyInput{
		Document: aws.String(d.Get("document").(string)),
		Name:     aws.String(name),
	}

	if v, ok := d.GetOk("comment"); ok {
		input.Comment = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating Route53 Traffic Policy: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEqualsContext(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateTrafficPolicyWithContext(ctx, input)
	}, route53.ErrCodeNoSuchTrafficPolicy)

	if err != nil {
		return diag.Errorf("error creating Route53 Traffic Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*route53.CreateTrafficPolicyOutput).TrafficPolicy.Id))

	return resourceTrafficPolicyRead(ctx, d, meta)
}

func resourceTrafficPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	trafficPolicy, err := FindTrafficPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Traffic Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Route53 Traffic Policy (%s): %s", d.Id(), err)
	}

	d.Set("comment", trafficPolicy.Comment)
	d.Set("document", trafficPolicy.Document)
	d.Set("name", trafficPolicy.Name)
	d.Set("type", trafficPolicy.Type)
	d.Set("version", trafficPolicy.Version)

	return nil
}

func resourceTrafficPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn

	input := &route53.UpdateTrafficPolicyCommentInput{
		Id:      aws.String(d.Id()),
		Version: aws.Int64(int64(d.Get("version").(int))),
	}

	if d.HasChange("comment") {
		input.Comment = aws.String(d.Get("comment").(string))
	}

	_, err := conn.UpdateTrafficPolicyCommentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Route53 Traffic Policy (%s): %s", d.Id(), err)
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
