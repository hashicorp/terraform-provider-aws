// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticsearch_domain_saml_options")
func ResourceDomainSAMLOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainSAMLOptionsPut,
		ReadWithoutTimeout:   resourceDomainSAMLOptionsRead,
		UpdateWithoutTimeout: resourceDomainSAMLOptionsPut,
		DeleteWithoutTimeout: resourceDomainSAMLOptionsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set(names.AttrDomainName, d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"saml_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"idp": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"metadata_content": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"master_backend_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"master_user_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"roles_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"session_timeout_minutes": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          60,
							ValidateFunc:     validation.IntBetween(1, 1440),
							DiffSuppressFunc: domainSamlOptionsDiffSupress,
						},
						"subject_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							DiffSuppressFunc: domainSamlOptionsDiffSupress,
						},
					},
				},
			},
		},
	}
}
func domainSamlOptionsDiffSupress(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.Get("saml_options").([]interface{}); ok && len(v) > 0 {
		if enabled, ok := v[0].(map[string]interface{})[names.AttrEnabled].(bool); ok && !enabled {
			return true
		}
	}
	return false
}

func resourceDomainSAMLOptionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn(ctx)

	ds, err := FindDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain SAML Options (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Received Elasticsearch domain: %s", ds)

	options := ds.AdvancedSecurityOptions.SAMLOptions

	if err := d.Set("saml_options", flattenESSAMLOptions(d, options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_options for Elasticsearch Configuration: %s", err)
	}

	return diags
}

func resourceDomainSAMLOptionsPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	config := elasticsearch.AdvancedSecurityOptionsInput{}
	config.SetSAMLOptions(expandESSAMLOptions(d.Get("saml_options").([]interface{})))

	log.Printf("[DEBUG] Updating Elasticsearch domain SAML Options %s", config)

	_, err := conn.UpdateElasticsearchDomainConfigWithContext(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	d.SetId(domainName)

	if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain SAML Options (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceDomainSAMLOptionsRead(ctx, d, meta)...)
}

func resourceDomainSAMLOptionsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchConn(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	config := elasticsearch.AdvancedSecurityOptionsInput{}
	config.SetSAMLOptions(nil)

	_, err := conn.UpdateElasticsearchDomainConfigWithContext(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Elasticsearch domain SAML Options %q to be deleted", d.Get(names.AttrDomainName).(string))

	if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain SAML Options (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
