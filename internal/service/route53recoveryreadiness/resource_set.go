// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_resource_set", name="Resource Set")
// @Tags(identifierAttribute="arn")
func resourceResourceSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceSetCreate,
		ReadWithoutTimeout:   resourceResourceSetRead,
		UpdateWithoutTimeout: resourceResourceSetUpdate,
		DeleteWithoutTimeout: resourceResourceSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_set_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResources: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_target_resource": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDomainName: {
										Type:     schema.TypeString,
										Required: true,
									},
									"hosted_zone_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"record_set_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"record_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"target_resource": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"nlb_resource": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrARN: {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"r53_resource": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrDomainName: {
																Type:     schema.TypeString,
																Optional: true,
															},
															"record_set_id": {
																Type:     schema.TypeString,
																Optional: true,
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
						"readiness_scopes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrResourceARN: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceResourceSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	name := d.Get("resource_set_name").(string)
	input := &route53recoveryreadiness.CreateResourceSetInput{
		ResourceSetName: aws.String(name),
		ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
		Resources:       expandResourceSetResources(d.Get(names.AttrResources).([]any)),
	}

	output, err := conn.CreateResourceSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Resource Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ResourceSetName))

	if err := createTags(ctx, conn, aws.ToString(output.ResourceSetArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Resource Set (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceResourceSetRead(ctx, d, meta)...)
}

func resourceResourceSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	output, err := findResourceSetByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Readiness Resource Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ResourceSetArn)
	d.Set("resource_set_name", output.ResourceSetName)
	d.Set("resource_set_type", output.ResourceSetType)
	if err := d.Set(names.AttrResources, flattenResourceSetResources(output.Resources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}

	return diags
}

func resourceResourceSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateResourceSetInput{
			ResourceSetName: aws.String(d.Id()),
			ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
			Resources:       expandResourceSetResources(d.Get(names.AttrResources).([]any)),
		}

		_, err := conn.UpdateResourceSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourceSetRead(ctx, d, meta)...)
}

func resourceResourceSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Resource Set: %s", d.Id())
	_, err := conn.DeleteResourceSet(ctx, &route53recoveryreadiness.DeleteResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := findResourceSetByName(ctx, conn, d.Id())
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Resource Set (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = findResourceSetByName(ctx, conn, d.Id())
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness Resource Set (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func findResourceSetByName(ctx context.Context, conn *route53recoveryreadiness.Client, name string) (*route53recoveryreadiness.GetResourceSetOutput, error) {
	input := &route53recoveryreadiness.GetResourceSetInput{
		ResourceSetName: aws.String(name),
	}

	output, err := conn.GetResourceSet(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandResourceSetResources(rs []any) []awstypes.Resource {
	var resources []awstypes.Resource

	for _, r := range rs {
		r := r.(map[string]any)
		resource := awstypes.Resource{}
		if v, ok := r[names.AttrResourceARN]; ok && v.(string) != "" {
			resource.ResourceArn = aws.String(v.(string))
		}
		if v, ok := r["readiness_scopes"]; ok {
			resource.ReadinessScopes = flex.ExpandStringValueList(v.([]any))
		}
		if v, ok := r["component_id"]; ok {
			resource.ComponentId = aws.String(v.(string))
		}
		if v, ok := r["dns_target_resource"]; ok {
			resource.DnsTargetResource = expandResourceSetDNSTargetResource(v.([]any))
		}
		resources = append(resources, resource)
	}
	return resources
}

func flattenResourceSetResources(resources []awstypes.Resource) []map[string]any {
	rs := make([]map[string]any, 0)
	for _, resource := range resources {
		r := map[string]any{}
		if v := resource.ResourceArn; v != nil {
			r[names.AttrResourceARN] = v
		}
		if v := resource.ReadinessScopes; v != nil {
			r["readiness_scopes"] = v
		}
		if v := resource.ComponentId; v != nil {
			r["component_id"] = v
		}
		if v := resource.DnsTargetResource; v != nil {
			r["dns_target_resource"] = flattenResourceSetDNSTargetResource(v)
		}
		rs = append(rs, r)
	}
	return rs
}

func expandResourceSetDNSTargetResource(dtrs []any) *awstypes.DNSTargetResource {
	dtresource := &awstypes.DNSTargetResource{}
	for _, dtr := range dtrs {
		dtr := dtr.(map[string]any)
		if v, ok := dtr[names.AttrDomainName]; ok && v.(string) != "" {
			dtresource.DomainName = aws.String(v.(string))
		}
		if v, ok := dtr["hosted_zone_arn"]; ok {
			dtresource.HostedZoneArn = aws.String(v.(string))
		}
		if v, ok := dtr["record_set_id"]; ok {
			dtresource.RecordSetId = aws.String(v.(string))
		}
		if v, ok := dtr["record_type"]; ok {
			dtresource.RecordType = aws.String(v.(string))
		}
		if v, ok := dtr["target_resource"]; ok {
			dtresource.TargetResource = expandResourceSetTargetResource(v.([]any))
		}
	}
	return dtresource
}

func flattenResourceSetDNSTargetResource(dtresource *awstypes.DNSTargetResource) []map[string]any {
	if dtresource == nil {
		return nil
	}

	dtr := make(map[string]any)
	dtr[names.AttrDomainName] = dtresource.DomainName
	dtr["hosted_zone_arn"] = dtresource.HostedZoneArn
	dtr["record_set_id"] = dtresource.RecordSetId
	dtr["record_type"] = dtresource.RecordType
	dtr["target_resource"] = flattenResourceSetTargetResource(dtresource.TargetResource)
	result := []map[string]any{dtr}
	return result
}

func expandResourceSetTargetResource(trs []any) *awstypes.TargetResource {
	if len(trs) == 0 {
		return nil
	}
	tresource := &awstypes.TargetResource{}
	for _, tr := range trs {
		if tr == nil {
			return nil
		}
		tr := tr.(map[string]any)
		if v, ok := tr["nlb_resource"]; ok && len(v.([]any)) > 0 {
			tresource.NLBResource = expandResourceSetNLBResource(v.([]any))
		}
		if v, ok := tr["r53_resource"]; ok && len(v.([]any)) > 0 {
			tresource.R53Resource = expandResourceSetR53ResourceRecord(v.([]any))
		}
	}
	return tresource
}

func flattenResourceSetTargetResource(tresource *awstypes.TargetResource) []map[string]any {
	if tresource == nil {
		return nil
	}

	tr := make(map[string]any)
	tr["nlb_resource"] = flattenResourceSetNLBResource(tresource.NLBResource)
	tr["r53_resource"] = flattenResourceSetR53ResourceRecord(tresource.R53Resource)
	result := []map[string]any{tr}
	return result
}

func expandResourceSetNLBResource(nlbrs []any) *awstypes.NLBResource {
	nlbresource := &awstypes.NLBResource{}
	for _, nlbr := range nlbrs {
		nlbr := nlbr.(map[string]any)
		if v, ok := nlbr[names.AttrARN]; ok && v.(string) != "" {
			nlbresource.Arn = aws.String(v.(string))
		}
	}
	return nlbresource
}

func flattenResourceSetNLBResource(nlbresource *awstypes.NLBResource) []map[string]any {
	if nlbresource == nil {
		return nil
	}

	nlbr := make(map[string]any)
	nlbr[names.AttrARN] = nlbresource.Arn
	result := []map[string]any{nlbr}
	return result
}

func expandResourceSetR53ResourceRecord(r53rs []any) *awstypes.R53ResourceRecord {
	r53resource := &awstypes.R53ResourceRecord{}
	for _, r53r := range r53rs {
		r53r := r53r.(map[string]any)
		if v, ok := r53r[names.AttrDomainName]; ok && v.(string) != "" {
			r53resource.DomainName = aws.String(v.(string))
		}
		if v, ok := r53r["record_set_id"]; ok {
			r53resource.RecordSetId = aws.String(v.(string))
		}
	}
	return r53resource
}

func flattenResourceSetR53ResourceRecord(r53resource *awstypes.R53ResourceRecord) []map[string]any {
	if r53resource == nil {
		return nil
	}

	r53r := make(map[string]any)
	r53r[names.AttrDomainName] = r53resource.DomainName
	r53r["record_set_id"] = r53resource.RecordSetId
	result := []map[string]any{r53r}
	return result
}
