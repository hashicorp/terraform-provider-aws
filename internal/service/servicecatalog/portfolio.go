package servicecatalog

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePortfolio() *schema.Resource {
	return &schema.Resource{
		Create: resourcePortfolioCreate,
		Read:   resourcePortfolioRead,
		Update: resourcePortfolioUpdate,
		Delete: resourcePortfolioDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 2000),
			},
			"provider_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}
func resourcePortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	input := servicecatalog.CreatePortfolioInput{
		AcceptLanguage:   aws.String(AcceptLanguageEnglish),
		DisplayName:      aws.String(d.Get("name").(string)),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Tags:             Tags(tags.IgnoreAws()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Service Catalog Portfolio: %#v", input)
	resp, err := conn.CreatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Creating Service Catalog Portfolio failed: %s", err.Error())
	}
	d.SetId(aws.StringValue(resp.PortfolioDetail.Id))

	return resourcePortfolioRead(d, meta)
}

func resourcePortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := servicecatalog.DescribePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
	}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Reading Service Catalog Portfolio: %#v", input)
	resp, err := conn.DescribePortfolio(&input)
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Service Catalog Portfolio %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading ServiceCatalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	portfolioDetail := resp.PortfolioDetail
	if err := d.Set("created_time", portfolioDetail.CreatedTime.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}
	d.Set("arn", portfolioDetail.ARN)
	d.Set("description", portfolioDetail.Description)
	d.Set("name", portfolioDetail.DisplayName)
	d.Set("provider_name", portfolioDetail.ProviderName)

	tags := KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourcePortfolioUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn
	input := servicecatalog.UpdatePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
		Id:             aws.String(d.Id()),
	}

	if d.HasChange("name") {
		v, _ := d.GetOk("name")
		input.DisplayName = aws.String(v.(string))
	}

	if d.HasChange("accept_language") {
		v, _ := d.GetOk("accept_language")
		input.AcceptLanguage = aws.String(v.(string))
	}

	if d.HasChange("description") {
		v, _ := d.GetOk("description")
		input.Description = aws.String(v.(string))
	}

	if d.HasChange("provider_name") {
		v, _ := d.GetOk("provider_name")
		input.ProviderName = aws.String(v.(string))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		input.AddTags = Tags(tftags.New(n).IgnoreAws())
		input.RemoveTags = aws.StringSlice(tftags.New(o).IgnoreAws().Keys())
	}

	log.Printf("[DEBUG] Update Service Catalog Portfolio: %#v", input)
	_, err := conn.UpdatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Updating Service Catalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return resourcePortfolioRead(d, meta)
}

func resourcePortfolioDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn
	input := servicecatalog.DeletePortfolioInput{}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Delete Service Catalog Portfolio: %#v", input)
	_, err := conn.DeletePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
