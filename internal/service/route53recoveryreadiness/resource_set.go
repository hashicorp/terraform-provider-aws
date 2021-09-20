package route53recoveryreadiness

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourceSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceSetCreate,
		Read:   resourceResourceSetRead,
		Update: resourceResourceSetUpdate,
		Delete: resourceResourceSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"resources": {
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
									"domain_name": {
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
															"arn": {
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
															"domain_name": {
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
						"resource_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceResourceSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53recoveryreadiness.CreateResourceSetInput{
		ResourceSetName: aws.String(d.Get("resource_set_name").(string)),
		ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
		Resources:       expandAwsRoute53RecoveryReadinessResourceSetResources(d.Get("resources").([]interface{})),
	}

	resp, err := conn.CreateResourceSet(input)
	if err != nil {
		return fmt.Errorf("error creating Route53 Recovery Readiness Resource Set: %w", err)
	}

	d.SetId(aws.StringValue(resp.ResourceSetName))

	if len(tags) > 0 {
		arn := aws.StringValue(resp.ResourceSetArn)
		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error adding Route53 Recovery Readiness Resource Set (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceResourceSetRead(d, meta)
}

func resourceResourceSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &route53recoveryreadiness.GetResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	}

	resp, err := conn.GetResourceSet(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53RecoveryReadiness Resource Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Route53 Recovery Readiness Resource Set: %s", err)
	}

	d.Set("arn", resp.ResourceSetArn)
	d.Set("resource_set_name", resp.ResourceSetName)
	d.Set("resource_set_type", resp.ResourceSetType)

	if err := d.Set("resources", flattenAwsRoute53RecoveryReadinessResourceSetResources(resp.Resources)); err != nil {
		return fmt.Errorf("Error setting AWS Route53 Recovery Readiness Resource Set resources: %s", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Recovery Readiness Resource Set (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceResourceSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn

	input := &route53recoveryreadiness.UpdateResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
		ResourceSetType: aws.String(d.Get("resource_set_type").(string)),
		Resources:       expandAwsRoute53RecoveryReadinessResourceSetResources(d.Get("resources").([]interface{})),
	}

	_, err := conn.UpdateResourceSet(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Readiness Resource Set: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)
		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Route53 Recovery Readiness Resource Set (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceResourceSetRead(d, meta)
}

func resourceResourceSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn

	input := &route53recoveryreadiness.DeleteResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	}
	_, err := conn.DeleteResourceSet(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Readiness Resource Set: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetResourceSetInput{
		ResourceSetName: aws.String(d.Id()),
	}
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetResourceSet(gcinput)
		if err != nil {
			if tfawserr.ErrMessageContains(err, route53recoveryreadiness.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Resource Set (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.GetResourceSet(gcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for Route 53 Recovery Readiness Resource Set (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandAwsRoute53RecoveryReadinessResourceSetResources(rs []interface{}) []*route53recoveryreadiness.Resource {
	var resources []*route53recoveryreadiness.Resource

	for _, r := range rs {
		r := r.(map[string]interface{})
		resource := &route53recoveryreadiness.Resource{}
		if v, ok := r["resource_arn"]; ok && v.(string) != "" {
			resource.ResourceArn = aws.String(v.(string))
		}
		if v, ok := r["readiness_scopes"]; ok {
			resource.ReadinessScopes = flex.ExpandStringList(v.([]interface{}))
		}
		if v, ok := r["component_id"]; ok {
			resource.ComponentId = aws.String(v.(string))
		}
		if v, ok := r["dns_target_resource"]; ok {
			resource.DnsTargetResource = expandAwsRoute53RecoveryReadinessResourceSetDnsTargetResource(v.([]interface{}))
		}
		resources = append(resources, resource)
	}
	return resources
}

func flattenAwsRoute53RecoveryReadinessResourceSetResources(resources []*route53recoveryreadiness.Resource) []map[string]interface{} {
	rs := make([]map[string]interface{}, 0)
	for _, resource := range resources {
		r := map[string]interface{}{}
		if v := resource.ResourceArn; v != nil {
			r["resource_arn"] = v
		}
		if v := resource.ReadinessScopes; v != nil {
			r["readiness_scopes"] = v
		}
		if v := resource.ComponentId; v != nil {
			r["component_id"] = v
		}
		if v := resource.DnsTargetResource; v != nil {
			r["dns_target_resource"] = flattenAwsRoute53RecoveryReadinessResourceSetDnsTargetResource(v)
		}
		rs = append(rs, r)
	}
	return rs
}

func expandAwsRoute53RecoveryReadinessResourceSetDnsTargetResource(dtrs []interface{}) *route53recoveryreadiness.DNSTargetResource {
	dtresource := &route53recoveryreadiness.DNSTargetResource{}
	for _, dtr := range dtrs {
		dtr := dtr.(map[string]interface{})
		if v, ok := dtr["domain_name"]; ok && v.(string) != "" {
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
			dtresource.TargetResource = expandAwsRoute53RecoveryReadinessResourceSetTargetResource(v.([]interface{}))
		}
	}
	return dtresource
}

func flattenAwsRoute53RecoveryReadinessResourceSetDnsTargetResource(dtresource *route53recoveryreadiness.DNSTargetResource) []map[string]interface{} {
	if dtresource == nil {
		return nil
	}

	dtr := make(map[string]interface{})
	dtr["domain_name"] = dtresource.DomainName
	dtr["hosted_zone_arn"] = dtresource.HostedZoneArn
	dtr["record_set_id"] = dtresource.RecordSetId
	dtr["record_type"] = dtresource.RecordType
	dtr["target_resource"] = flattenAwsRoute53RecoveryReadinessResourceSetTargetResource(dtresource.TargetResource)
	result := []map[string]interface{}{dtr}
	return result
}

func expandAwsRoute53RecoveryReadinessResourceSetTargetResource(trs []interface{}) *route53recoveryreadiness.TargetResource {
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
			tresource.NLBResource = expandAwsRoute53RecoveryReadinessResourceSetNLBResource(v.([]interface{}))
		}
		if v, ok := tr["r53_resource"]; ok && len(v.([]interface{})) > 0 {
			tresource.R53Resource = expandAwsRoute53RecoveryReadinessResourceSetR53ResourceRecord(v.([]interface{}))
		}
	}
	return tresource
}

func flattenAwsRoute53RecoveryReadinessResourceSetTargetResource(tresource *route53recoveryreadiness.TargetResource) []map[string]interface{} {
	if tresource == nil {
		return nil
	}

	tr := make(map[string]interface{})
	tr["nlb_resource"] = flattenAwsRoute53RecoveryReadinessResourceSetNLBResource(tresource.NLBResource)
	tr["r53_resource"] = flattenAwsRoute53RecoveryReadinessResourceSetR53ResourceRecord(tresource.R53Resource)
	result := []map[string]interface{}{tr}
	return result
}

func expandAwsRoute53RecoveryReadinessResourceSetNLBResource(nlbrs []interface{}) *route53recoveryreadiness.NLBResource {
	nlbresource := &route53recoveryreadiness.NLBResource{}
	for _, nlbr := range nlbrs {
		nlbr := nlbr.(map[string]interface{})
		if v, ok := nlbr["arn"]; ok && v.(string) != "" {
			nlbresource.Arn = aws.String(v.(string))
		}
	}
	return nlbresource
}

func flattenAwsRoute53RecoveryReadinessResourceSetNLBResource(nlbresource *route53recoveryreadiness.NLBResource) []map[string]interface{} {
	if nlbresource == nil {
		return nil
	}

	nlbr := make(map[string]interface{})
	nlbr["arn"] = nlbresource.Arn
	result := []map[string]interface{}{nlbr}
	return result
}

func expandAwsRoute53RecoveryReadinessResourceSetR53ResourceRecord(r53rs []interface{}) *route53recoveryreadiness.R53ResourceRecord {
	r53resource := &route53recoveryreadiness.R53ResourceRecord{}
	for _, r53r := range r53rs {
		r53r := r53r.(map[string]interface{})
		if v, ok := r53r["domain_name"]; ok && v.(string) != "" {
			r53resource.DomainName = aws.String(v.(string))
		}
		if v, ok := r53r["record_set_id"]; ok {
			r53resource.RecordSetId = aws.String(v.(string))
		}
	}
	return r53resource
}

func flattenAwsRoute53RecoveryReadinessResourceSetR53ResourceRecord(r53resource *route53recoveryreadiness.R53ResourceRecord) []map[string]interface{} {
	if r53resource == nil {
		return nil
	}

	r53r := make(map[string]interface{})
	r53r["domain_name"] = r53resource.DomainName
	r53r["record_set_id"] = r53resource.RecordSetId
	result := []map[string]interface{}{r53r}
	return result
}
