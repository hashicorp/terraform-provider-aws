package elasticsearch

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainPolicyUpsert,
		Read:   resourceDomainPolicyRead,
		Update: resourceDomainPolicyUpsert,
		Delete: resourceDomainPolicyDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_policies": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceDomainPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	name := d.Get("domain_name").(string)

	ds, err := FindDomainByName(conn, d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Elasticsearch Domain Policy (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received Elasticsearch domain: %s", out)

	d.Set("access_policies", ds.AccessPolicies)

	return nil
}

func resourceDomainPolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	domainName := d.Get("domain_name").(string)
	_, err := conn.UpdateElasticsearchDomainConfig(&elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:     aws.String(domainName),
		AccessPolicies: aws.String(d.Get("access_policies").(string)),
	})
	if err != nil {
		return err
	}

	d.SetId("esd-policy-" + domainName)

	if err := waitForDomainUpdate(conn, d.Get("domain_name").(string)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain SAML Policy (%s) to be updated: %w", d.Id(), err)
	}

	return resourceDomainPolicyRead(d, meta)
}

func resourceDomainPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	_, err := conn.UpdateElasticsearchDomainConfig(&elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:     aws.String(d.Get("domain_name").(string)),
		AccessPolicies: aws.String(""),
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for Elasticsearch domain policy %q to be deleted", d.Get("domain_name").(string))

	if err := waitForDomainUpdate(conn, d.Get("domain_name").(string)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain SAML Policy (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
