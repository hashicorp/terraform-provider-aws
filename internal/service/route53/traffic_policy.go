// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_traffic_policy", name="Traffic Policy")
func resourceTrafficPolicy() *schema.Resource {
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
				d.Set(names.AttrVersion, version)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrComment: {
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceTrafficPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53.CreateTrafficPolicyInput{
		Document: aws.String(d.Get("document").(string)),
		Name:     aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		input.Comment = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.NoSuchTrafficPolicy](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateTrafficPolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Traffic Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*route53.CreateTrafficPolicyOutput).TrafficPolicy.Id))

	return append(diags, resourceTrafficPolicyRead(ctx, d, meta)...)
}

func resourceTrafficPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	trafficPolicy, err := findTrafficPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Traffic Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Traffic Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrComment, trafficPolicy.Comment)
	d.Set("document", trafficPolicy.Document)
	d.Set(names.AttrName, trafficPolicy.Name)
	d.Set(names.AttrType, trafficPolicy.Type)
	d.Set(names.AttrVersion, trafficPolicy.Version)

	return diags
}

func resourceTrafficPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	input := &route53.UpdateTrafficPolicyCommentInput{
		Id:      aws.String(d.Id()),
		Version: aws.Int32(int32(d.Get(names.AttrVersion).(int))),
	}

	if d.HasChange(names.AttrComment) {
		input.Comment = aws.String(d.Get(names.AttrComment).(string))
	}

	_, err := conn.UpdateTrafficPolicyComment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Traffic Policy (%s) comment: %s", d.Id(), err)
	}

	return append(diags, resourceTrafficPolicyRead(ctx, d, meta)...)
}

func resourceTrafficPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	input := &route53.ListTrafficPolicyVersionsInput{
		Id: aws.String(d.Id()),
	}
	output, err := findTrafficPolicyVersions(ctx, conn, input)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Traffic Policy (%s) versions: %s", d.Id(), err)
	}

	for _, v := range output {
		version := aws.ToInt32(v.Version)

		log.Printf("[INFO] Deleting Route53 Traffic Policy (%s) version: %d", d.Id(), version)
		_, err := conn.DeleteTrafficPolicy(ctx, &route53.DeleteTrafficPolicyInput{
			Id:      aws.String(d.Id()),
			Version: aws.Int32(version),
		})

		if errs.IsA[*awstypes.NoSuchTrafficPolicy](err) {
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Route 53 Traffic Policy (%s) version (%d): %s", d.Id(), version, err)
		}
	}

	return diags
}

func findTrafficPolicyByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.TrafficPolicy, error) {
	inputLTP := &route53.ListTrafficPoliciesInput{}
	trafficPolicy, err := findTrafficPolicy(ctx, conn, inputLTP, func(v *awstypes.TrafficPolicySummary) bool {
		return aws.ToString(v.Id) == id
	})

	if err != nil {
		return nil, err
	}

	inputGTP := &route53.GetTrafficPolicyInput{
		Id:      aws.String(id),
		Version: trafficPolicy.LatestVersion,
	}

	output, err := conn.GetTrafficPolicy(ctx, inputGTP)

	if errs.IsA[*awstypes.NoSuchTrafficPolicy](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: inputLTP,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrafficPolicy == nil {
		return nil, tfresource.NewEmptyResultError(inputLTP)
	}

	return output.TrafficPolicy, nil
}

func findTrafficPolicy(ctx context.Context, conn *route53.Client, input *route53.ListTrafficPoliciesInput, filter tfslices.Predicate[*awstypes.TrafficPolicySummary]) (*awstypes.TrafficPolicySummary, error) {
	output, err := findTrafficPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrafficPolicies(ctx context.Context, conn *route53.Client, input *route53.ListTrafficPoliciesInput, filter tfslices.Predicate[*awstypes.TrafficPolicySummary]) ([]awstypes.TrafficPolicySummary, error) {
	var output []awstypes.TrafficPolicySummary

	err := listTrafficPoliciesPages(ctx, conn, input, func(page *route53.ListTrafficPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicySummaries {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findTrafficPolicyVersions(ctx context.Context, conn *route53.Client, input *route53.ListTrafficPolicyVersionsInput) ([]awstypes.TrafficPolicy, error) {
	var output []awstypes.TrafficPolicy

	err := listTrafficPolicyVersionsPages(ctx, conn, input, func(page *route53.ListTrafficPolicyVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.TrafficPolicies...)

		return !lastPage
	})

	if errs.IsA[*awstypes.NoSuchTrafficPolicy](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
