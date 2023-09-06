// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// @SDKResource("aws_route53_traffic_policy")
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
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	name := d.Get("name").(string)
	input := &route53.CreateTrafficPolicyInput{
		Document: aws.String(d.Get("document").(string)),
		Name:     aws.String(name),
	}

	if v, ok := d.GetOk("comment"); ok {
		input.Comment = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating Route53 Traffic Policy: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateTrafficPolicyWithContext(ctx, input)
	}, route53.ErrCodeNoSuchTrafficPolicy)

	if err != nil {
		return diag.Errorf("creating Route53 Traffic Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*route53.CreateTrafficPolicyOutput).TrafficPolicy.Id))

	return resourceTrafficPolicyRead(ctx, d, meta)
}

func resourceTrafficPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	trafficPolicy, err := FindTrafficPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Traffic Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Traffic Policy (%s): %s", d.Id(), err)
	}

	d.Set("comment", trafficPolicy.Comment)
	d.Set("document", trafficPolicy.Document)
	d.Set("name", trafficPolicy.Name)
	d.Set("type", trafficPolicy.Type)
	d.Set("version", trafficPolicy.Version)

	return nil
}

func resourceTrafficPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	input := &route53.UpdateTrafficPolicyCommentInput{
		Id:      aws.String(d.Id()),
		Version: aws.Int64(int64(d.Get("version").(int))),
	}

	if d.HasChange("comment") {
		input.Comment = aws.String(d.Get("comment").(string))
	}

	log.Printf("[INFO] Updating Route53 Traffic Policy comment: %s", input)
	_, err := conn.UpdateTrafficPolicyCommentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Route53 Traffic Policy (%s) comment: %s", d.Id(), err)
	}

	return resourceTrafficPolicyRead(ctx, d, meta)
}

func resourceTrafficPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	input := &route53.ListTrafficPolicyVersionsInput{
		Id: aws.String(d.Id()),
	}
	var output []*route53.TrafficPolicy

	err := listTrafficPolicyVersionsPages(ctx, conn, input, func(page *route53.ListTrafficPolicyVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.TrafficPolicies...)

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("listing Route 53 Traffic Policy (%s) versions: %s", d.Id(), err)
	}

	for _, v := range output {
		version := aws.Int64Value(v.Version)

		log.Printf("[INFO] Delete Route53 Traffic Policy (%s) version: %d", d.Id(), version)
		_, err := conn.DeleteTrafficPolicyWithContext(ctx, &route53.DeleteTrafficPolicyInput{
			Id:      aws.String(d.Id()),
			Version: aws.Int64(version),
		})

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
			continue
		}

		if err != nil {
			return diag.Errorf("deleting Route 53 Traffic Policy (%s) version (%d): %s", d.Id(), version, err)
		}
	}

	return nil
}
