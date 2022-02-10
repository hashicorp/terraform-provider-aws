package route53domains

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegisteredDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegisteredDomainCreate,
		Read:   resourceRegisteredDomainRead,
		Update: resourceRegisteredDomainUpdate,
		Delete: resourceRegisteredDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegisteredDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	domainName := d.Get("domain_name").(string)
	domainDetail, err := FindDomainDetailByName(conn, domainName)

	if err != nil {
		return fmt.Errorf("error reading Route 53 Domains Domain (%s): %w", domainName, err)
	}

	d.SetId(aws.StringValue(domainDetail.DomainName))

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Route 53 Domains Domain (%s): %w", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{}))).IgnoreConfig(ignoreTagsConfig)
	oldTags := tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := UpdateTags(conn, d.Id(), oldTags, newTags); err != nil {
			return fmt.Errorf("error updating Route 53 Domains Domain (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceRegisteredDomainRead(d, meta)
}

func resourceRegisteredDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53DomainsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainDetail, err := FindDomainDetailByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Domains Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route 53 Domains Domain (%s): %w", d.Id(), err)
	}

	d.Set("domain_name", domainDetail.DomainName)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Route 53 Domains Domain (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRegisteredDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Route 53 Domains Domain (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceRegisteredDomainRead(d, meta)
}

func resourceRegisteredDomainDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Route 53 Domains Registered Domain (%s) not deleted, removing from state", d.Id())

	return nil
}

func FindDomainDetailByName(conn *route53domains.Route53Domains, name string) (*route53domains.GetDomainDetailOutput, error) {
	input := &route53domains.GetDomainDetailInput{
		DomainName: aws.String(name),
	}

	output, err := conn.GetDomainDetail(input)

	if tfawserr.ErrMessageContains(err, route53domains.ErrCodeInvalidInput, "not found") {
		return nil, &resource.NotFoundError{
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
