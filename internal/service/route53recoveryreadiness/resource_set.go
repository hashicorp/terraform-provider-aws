// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_resource_set", name="Resource Set")
// @Tags(identifierAttribute="arn")
func ResourceResourceSet() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceResourceSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	name := d.Get("resource_set_name").(string)
	input := &route53recoveryreadiness.CreateResourceSetInput{
		ResourceSetName: aws.String(name),
		ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
		Resources:       expandResourceSetResources(d.Get(names.AttrResources).([]interface{})),
	}

	output, err := conn.CreateResourceSetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Resource Set (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ResourceSetName))

	if err := createTags(ctx, conn, aws.StringValue(output.ResourceSetArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Resource Set (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceResourceSetRead(ctx, d, meta)...)
}

func resourceResourceSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	input := &route53recoveryreadiness.GetResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	}

	resp, err := conn.GetResourceSetWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Readiness Resource Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.ResourceSetArn)
	d.Set("resource_set_name", resp.ResourceSetName)
	d.Set("resource_set_type", resp.ResourceSetType)
	if err := d.Set(names.AttrResources, flattenResourceSetResources(resp.Resources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}

	return diags
}

func resourceResourceSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateResourceSetInput{
			ResourceSetName: aws.String(d.Id()),
			ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
			Resources:       expandResourceSetResources(d.Get(names.AttrResources).([]interface{})),
		}

		_, err := conn.UpdateResourceSetWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourceSetRead(ctx, d, meta)...)
}

func resourceResourceSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Resource Set: %s", d.Id())
	_, err := conn.DeleteResourceSetWithContext(ctx, &route53recoveryreadiness.DeleteResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Resource Set (%s): %s", d.Id(), err)
	}

	gcinput := &route53recoveryreadiness.GetResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.GetResourceSetWithContext(ctx, gcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Resource Set (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.GetResourceSetWithContext(ctx, gcinput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness Resource Set (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func expandResourceSetResources(rs []interface{}) []*route53recoveryreadiness.Resource {
	var resources []*route53recoveryreadiness.Resource

	for _, r := range rs {
		r := r.(map[string]interface{})
		resource := &route53recoveryreadiness.Resource{}
		if v, ok := r[names.AttrResourceARN]; ok && v.(string) != "" {
			resource.ResourceArn = aws.String(v.(string))
		}
		if v, ok := r["readiness_scopes"]; ok {
			resource.ReadinessScopes = flex.ExpandStringList(v.([]interface{}))
		}
		if v, ok := r["component_id"]; ok {
			resource.ComponentId = aws.String(v.(string))
		}
		if v, ok := r["dns_target_resource"]; ok {
			resource.DnsTargetResource = expandResourceSetDNSTargetResource(v.([]interface{}))
		}
		resources = append(resources, resource)
	}
	return resources
}

func flattenResourceSetResources(resources []*route53recoveryreadiness.Resource) []map[string]interface{} {
	rs := make([]map[string]interface{}, 0)
	for _, resource := range resources {
		r := map[string]interface{}{}
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

func expandResourceSetDNSTargetResource(dtrs []interface{}) *route53recoveryreadiness.DNSTargetResource {
	dtresource := &route53recoveryreadiness.DNSTargetResource{}
	for _, dtr := range dtrs {
		dtr := dtr.(map[string]interface{})
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
			dtresource.TargetResource = expandResourceSetTargetResource(v.([]interface{}))
		}
	}
	return dtresource
}

func flattenResourceSetDNSTargetResource(dtresource *route53recoveryreadiness.DNSTargetResource) []map[string]interface{} {
	if dtresource == nil {
		return nil
	}

	dtr := make(map[string]interface{})
	dtr[names.AttrDomainName] = dtresource.DomainName
	dtr["hosted_zone_arn"] = dtresource.HostedZoneArn
	dtr["record_set_id"] = dtresource.RecordSetId
	dtr["record_type"] = dtresource.RecordType
	dtr["target_resource"] = flattenResourceSetTargetResource(dtresource.TargetResource)
	result := []map[string]interface{}{dtr}
	return result
}

func expandResourceSetTargetResource(trs []interface{}) *route53recoveryreadiness.TargetResource {
	if len(trs) == 0 {
		return nil
	}
	tresource := &route53recoveryreadiness.TargetResource{}
	for _, tr := range trs {
		if tr == nil {
			return nil
		}
		tr := tr.(map[string]interface{})
		if v, ok := tr["nlb_resource"]; ok && len(v.([]interface{})) > 0 {
			tresource.NLBResource = expandResourceSetNLBResource(v.([]interface{}))
		}
		if v, ok := tr["r53_resource"]; ok && len(v.([]interface{})) > 0 {
			tresource.R53Resource = expandResourceSetR53ResourceRecord(v.([]interface{}))
		}
	}
	return tresource
}

func flattenResourceSetTargetResource(tresource *route53recoveryreadiness.TargetResource) []map[string]interface{} {
	if tresource == nil {
		return nil
	}

	tr := make(map[string]interface{})
	tr["nlb_resource"] = flattenResourceSetNLBResource(tresource.NLBResource)
	tr["r53_resource"] = flattenResourceSetR53ResourceRecord(tresource.R53Resource)
	result := []map[string]interface{}{tr}
	return result
}

func expandResourceSetNLBResource(nlbrs []interface{}) *route53recoveryreadiness.NLBResource {
	nlbresource := &route53recoveryreadiness.NLBResource{}
	for _, nlbr := range nlbrs {
		nlbr := nlbr.(map[string]interface{})
		if v, ok := nlbr[names.AttrARN]; ok && v.(string) != "" {
			nlbresource.Arn = aws.String(v.(string))
		}
	}
	return nlbresource
}

func flattenResourceSetNLBResource(nlbresource *route53recoveryreadiness.NLBResource) []map[string]interface{} {
	if nlbresource == nil {
		return nil
	}

	nlbr := make(map[string]interface{})
	nlbr[names.AttrARN] = nlbresource.Arn
	result := []map[string]interface{}{nlbr}
	return result
}

func expandResourceSetR53ResourceRecord(r53rs []interface{}) *route53recoveryreadiness.R53ResourceRecord {
	r53resource := &route53recoveryreadiness.R53ResourceRecord{}
	for _, r53r := range r53rs {
		r53r := r53r.(map[string]interface{})
		if v, ok := r53r[names.AttrDomainName]; ok && v.(string) != "" {
			r53resource.DomainName = aws.String(v.(string))
		}
		if v, ok := r53r["record_set_id"]; ok {
			r53resource.RecordSetId = aws.String(v.(string))
		}
	}
	return r53resource
}

func flattenResourceSetR53ResourceRecord(r53resource *route53recoveryreadiness.R53ResourceRecord) []map[string]interface{} {
	if r53resource == nil {
		return nil
	}

	r53r := make(map[string]interface{})
	r53r[names.AttrDomainName] = r53resource.DomainName
	r53r["record_set_id"] = r53resource.RecordSetId
	result := []map[string]interface{}{r53r}
	return result
}
