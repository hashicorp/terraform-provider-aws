package elasticsearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainPolicyUpsert,
		ReadWithoutTimeout:   resourceDomainPolicyRead,
		UpdateWithoutTimeout: resourceDomainPolicyUpsert,
		DeleteWithoutTimeout: resourceDomainPolicyDelete,

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDomainPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn()

	ds, err := FindDomainByName(ctx, conn, d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	policies, err := verify.PolicyToSet(d.Get("access_policies").(string), aws.StringValue(ds.AccessPolicies))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	d.Set("access_policies", policies)

	return diags
}

func resourceDomainPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn()
	domainName := d.Get("domain_name").(string)

	policy, err := structure.NormalizeJsonString(d.Get("access_policies").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	_, err = conn.UpdateElasticsearchDomainConfigWithContext(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:     aws.String(domainName),
		AccessPolicies: aws.String(policy),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	d.SetId("esd-policy-" + domainName)

	if err := waitForDomainUpdate(ctx, conn, d.Get("domain_name").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain Policy (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceDomainPolicyRead(ctx, d, meta)...)
}

func resourceDomainPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn()

	_, err := conn.UpdateElasticsearchDomainConfigWithContext(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:     aws.String(d.Get("domain_name").(string)),
		AccessPolicies: aws.String(""),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Elasticsearch domain policy %q to be deleted", d.Get("domain_name").(string))

	if err := waitForDomainUpdate(ctx, conn, d.Get("domain_name").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain Policy (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
