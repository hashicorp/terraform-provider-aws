// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_prometheus_scraper", name="Scraper")
// @Tags(identifierAttribute="arn")
func ResourceScraper() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScraperCreate,
		ReadWithoutTimeout:   resourceScraperRead,
		DeleteWithoutTimeout: resourceScraperDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_prometheus_workspace_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"scrape_configuration": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eks_cluster_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameScraper = "Scraper"
)

func resourceScraperCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("generating uuid for ClientToken for Prometheus Scraper %s", err)
	}

	scrapeConfig := d.Get("scrape_configuration").(string)

	in := &prometheusservice.CreateScraperInput{
		Source:      expandSource(d.Get("source").([]interface{})),
		Destination: expandDestination(d.Get("destination").([]interface{})),
		ScrapeConfiguration: &prometheusservice.ScrapeConfiguration{
			ConfigurationBlob: []byte(scrapeConfig),
		},
		ClientToken: aws.String(uuid),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alias"); ok {
		in.Alias = aws.String(v.(string))
	}

	out, err := conn.CreateScraperWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Managed Prometheus Scraper: %s", err)
	}

	d.SetId(aws.StringValue(out.ScraperId))

	if _, err := waitScraperCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Managed Prometheus Scraper (%s) create: %s", d.Id(), err)
	}

	return resourceScraperRead(ctx, d, meta)
}

func resourceScraperRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	scraper, err := FindScraperByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amazon Managed Prometheus Scraper Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Amazon Managed Prometheus Scraper Definition (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(scraper.ScraperId))
	d.Set("alias", aws.StringValue(scraper.Alias))
	d.Set("arn", aws.StringValue(scraper.Arn))
	d.Set("destination", flattenDestination(scraper.Destination))
	d.Set("scrape_configuration", string(scraper.ScrapeConfiguration.ConfigurationBlob))
	d.Set("source", flattenSource(scraper.Source))
	setTagsOut(ctx, scraper.Tags)

	return nil
}

func resourceScraperDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	log.Printf("[INFO] Deleting AMP Scraper %s", d.Id())

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("generating uuid for ClientToken for Prometheus Scraper (%s) %s", d.Id(), err)
	}

	input := &prometheusservice.DeleteScraperInput{
		ScraperId:   aws.String(d.Id()),
		ClientToken: aws.String(uuid),
	}

	_, err = conn.DeleteScraperWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Amazon Managed Prometheus Scraper (%s): %s", d.Id(), err)
	}

	if _, err := waitScraperDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Amazon Managed Prometheus Scraper (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func flattenStringSet(sx []*string) []interface{} {
	if len(sx) == 0 {
		return nil
	}

	var l []interface{}

	for _, s := range sx {
		if s == nil {
			continue
		}

		l = append(l, aws.StringValue(s))
	}

	return l
}

func flattenSource(source *prometheusservice.Source) []interface{} {
	if source == nil {
		return nil
	}

	var tfList []interface{}

	tfMap := map[string]interface{}{
		"eks_cluster_arn": aws.StringValue(source.EksConfiguration.ClusterArn),
		"subnet_ids":      flattenStringSet(source.EksConfiguration.SubnetIds),
	}
	if sg_ids := source.EksConfiguration.SecurityGroupIds; sg_ids != nil {
		tfMap["security_group_ids"] = flattenStringSet(sg_ids)
	}
	tfList = append(tfList, tfMap)

	return tfList
}

func expandSource(l []interface{}) *prometheusservice.Source {

	m := l[0].(map[string]interface{})

	eksConfig := &prometheusservice.EksConfiguration{
		ClusterArn: aws.String(m["eks_cluster_arn"].(string)),
		SubnetIds:  aws.StringSlice(flex.ExpandStringValueSet(m["subnet_ids"].(*schema.Set))),
	}

	if v, ok := m["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		eksConfig.SecurityGroupIds = aws.StringSlice(flex.ExpandStringValueSet(v))
	}

	return &prometheusservice.Source{EksConfiguration: eksConfig}
}

func flattenDestination(dest *prometheusservice.Destination) []interface{} {

	if dest == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"aws_prometheus_workspace_arn": aws.StringValue(dest.AmpConfiguration.WorkspaceArn),
	}
	tfList := []interface{}{tfMap}

	return tfList
}

func expandDestination(l []interface{}) *prometheusservice.Destination {

	m := l[0].(map[string]interface{})

	ampConfig := &prometheusservice.AmpConfiguration{
		WorkspaceArn: aws.String(m["aws_prometheus_workspace_arn"].(string)),
	}

	return &prometheusservice.Destination{AmpConfiguration: ampConfig}
}
