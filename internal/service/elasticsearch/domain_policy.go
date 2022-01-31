package elasticsearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceDomainPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	ds, err := FindDomainByName(conn, d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Elasticsearch Domain Policy (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received Elasticsearch domain: %s", ds)

	policies, err := verify.PolicyToSet(d.Get("access_policies").(string), aws.StringValue(ds.AccessPolicies))

	if err != nil {
		return err
	}

	d.Set("access_policies", policies)

	return nil
}

func resourceDomainPolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	domainName := d.Get("domain_name").(string)

	policy, err := structure.NormalizeJsonString(d.Get("access_policies").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	_, err = conn.UpdateElasticsearchDomainConfig(&elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:     aws.String(domainName),
		AccessPolicies: aws.String(policy),
	})
	if err != nil {
		return err
	}

	d.SetId("esd-policy-" + domainName)

	if err := waitForDomainUpdate(conn, d.Get("domain_name").(string)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain Policy (%s) to be updated: %w", d.Id(), err)
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
		return fmt.Errorf("error waiting for Elasticsearch Domain Policy (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
