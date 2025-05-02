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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticsearch_domain_saml_options", name="Domain SAML Options")
func resourceDomainSAMLOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainSAMLOptionsPut,
		ReadWithoutTimeout:   resourceDomainSAMLOptionsRead,
		UpdateWithoutTimeout: resourceDomainSAMLOptionsPut,
		DeleteWithoutTimeout: resourceDomainSAMLOptionsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDomainSAMLOptionsPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := &elasticsearch.UpdateElasticsearchDomainConfigInput{
		AdvancedSecurityOptions: &awstypes.AdvancedSecurityOptionsInput{
			SAMLOptions: expandESSAMLOptions(d.Get("saml_options").([]any)),
		},
		DomainName: aws.String(domainName),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout,
		func() (any, error) {
			return conn.UpdateElasticsearchDomainConfig(ctx, input)
		}, "A change/update is in progress")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(domainName)
	}

	if _, err := waitDomainConfigUpdated(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
	}

	return append(diags, resourceDomainSAMLOptionsRead(ctx, d, meta)...)
}

func resourceDomainSAMLOptionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	output, err := findDomainSAMLOptionByDomainName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain SAML Options (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDomainName, d.Id())
	if err := d.Set("saml_options", flattenSAMLOptionsOutput(d, output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_options: %s", err)
	}

	return diags
}

func resourceDomainSAMLOptionsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	log.Printf("[DEBUG] Deleting Elasticsearch Domain SAML Options: %s", d.Id())
	_, err := conn.UpdateElasticsearchDomainConfig(ctx, &elasticsearch.UpdateElasticsearchDomainConfigInput{
		AdvancedSecurityOptions: &awstypes.AdvancedSecurityOptionsInput{},
		DomainName:              aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Domain is being deleted") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainConfigUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
	}

	return diags
}

func domainSamlOptionsDiffSupress(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.Get("saml_options").([]any); ok && len(v) > 0 {
		if enabled, ok := v[0].(map[string]any)[names.AttrEnabled].(bool); ok && !enabled {
			return true
		}
	}
	return false
}

func findDomainSAMLOptionByDomainName(ctx context.Context, conn *elasticsearch.Client, name string) (*awstypes.SAMLOptionsOutput, error) {
	output, err := findDomainByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.AdvancedSecurityOptions == nil || output.AdvancedSecurityOptions.SAMLOptions == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.AdvancedSecurityOptions.SAMLOptions, nil
}
