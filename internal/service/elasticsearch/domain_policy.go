// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	elasticsearch "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticsearch_domain_policy", name="Domain Policy")
func resourceDomainPolicy() *schema.Resource {
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
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	policy, err := structure.NormalizeJsonString(d.Get("access_policies").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	input := &elasticsearch.UpdateElasticsearchDomainConfigInput{
		AccessPolicies: aws.String(policy),
		DomainName:     aws.String(domainName),
	}

	_, err = tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout,
		func() (any, error) {
			return conn.UpdateElasticsearchDomainConfig(ctx, input)
		}, "A change/update is in progress")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId("esd-policy-" + domainName)
	}

	if _, err := waitDomainConfigUpdated(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
	}

	return append(diags, resourceDomainPolicyRead(ctx, d, meta)...)
}

func resourceDomainPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	ds, err := findDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	policies, err := verify.PolicyToSet(d.Get("access_policies").(string), aws.ToString(ds.AccessPolicies))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("access_policies", policies)

	return diags
}

func resourceDomainPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	log.Printf("[DEBUG] Deleting Elasticsearch Domain Policy: %s", d.Id())
	_, err := conn.UpdateElasticsearchDomainConfig(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		AccessPolicies: aws.String(""),
		DomainName:     aws.String(d.Get(names.AttrDomainName).(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain Policy (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainConfigUpdated(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
	}

	return diags
}
