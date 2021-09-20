package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53resolver/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallDomainListCreate,
		Read:   resourceFirewallDomainListRead,
		Update: resourceFirewallDomainListUpdate,
		Delete: resourceFirewallDomainListDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},

			"domains": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 255,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallDomainListCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53resolver.CreateFirewallDomainListInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-domain-list-")),
		Name:             aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().Route53resolverTags()
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall domain list: %#v", input)
	output, err := conn.CreateFirewallDomainList(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall domain list: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallDomainList.Id))
	d.Set("arn", output.FirewallDomainList.Arn)

	return resourceFirewallDomainListUpdate(d, meta)
}

func resourceFirewallDomainListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	firewallDomainList, err := finder.FindFirewallDomainListByID(conn, d.Id())

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver DNS Firewall domain list (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall domain list (%s): %w", d.Id(), err)
	}

	if firewallDomainList == nil {
		log.Printf("[WARN] Route 53 Resolver DNS Firewall domain list (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(firewallDomainList.Arn)
	d.Set("arn", arn)
	d.Set("name", firewallDomainList.Name)

	input := &route53resolver.ListFirewallDomainsInput{
		FirewallDomainListId: aws.String(d.Id()),
	}

	domains := []*string{}

	err = conn.ListFirewallDomainsPages(input, func(output *route53resolver.ListFirewallDomainsOutput, lastPage bool) bool {
		domains = append(domains, output.Domains...)
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Route 53 Resolver DNS Firewall domain list (%s) domains: %w", d.Id(), err)
	}

	d.Set("domains", flex.FlattenStringSet(domains))

	tags, err := tftags.Route53resolverListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver DNS Firewall domain list (%s): %w", arn, err)
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

func resourceFirewallDomainListUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	if d.HasChange("domains") {
		o, n := d.GetChange("domains")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		domains := ns
		operation := route53resolver.FirewallDomainUpdateOperationReplace

		if domains.Len() == 0 {
			domains = os
			operation = route53resolver.FirewallDomainUpdateOperationRemove
		}

		_, err := conn.UpdateFirewallDomains(&route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringSet(domains),
			Operation:            aws.String(operation),
		})

		if err != nil {
			return fmt.Errorf("error updating Route 53 Resolver DNS Firewall domain list (%s) domains: %w", d.Id(), err)
		}

		_, err = waiter.waitFirewallDomainListUpdated(conn, d.Id())

		if err != nil {
			return fmt.Errorf("error waiting for Route 53 Resolver DNS Firewall domain list (%s) domains to be updated: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.Route53resolverUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver DNS Firewall domain list (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceFirewallDomainListRead(d, meta)
}

func resourceFirewallDomainListDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	_, err := conn.DeleteFirewallDomainList(&route53resolver.DeleteFirewallDomainListInput{
		FirewallDomainListId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall domain list (%s): %w", d.Id(), err)
	}

	_, err = waiter.waitFirewallDomainListDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route 53 Resolver DNS Firewall domain list (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
