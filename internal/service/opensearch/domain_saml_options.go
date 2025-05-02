// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_domain_saml_options", name="Domain SAML Options")
func resourceDomainSAMLOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainSAMLOptionsPut,
		ReadWithoutTimeout:   resourceDomainSAMLOptionsRead,
		UpdateWithoutTimeout: resourceDomainSAMLOptionsPut,
		DeleteWithoutTimeout: resourceDomainSAMLOptionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set(names.AttrDomainName, d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
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
	if v, ok := d.Get("saml_options").([]any); ok && len(v) > 0 {
		if enabled, ok := v[0].(map[string]any)[names.AttrEnabled].(bool); ok && !enabled {
			return true
		}
	}
	return false
}

func resourceDomainSAMLOptionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	ds, err := findDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Domain SAML Options (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Received OpenSearch domain: %#v", ds)

	options := ds.AdvancedSecurityOptions.SAMLOptions

	if err := d.Set("saml_options", flattenESSAMLOptions(d, options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_options for OpenSearch Configuration: %s", err)
	}

	return diags
}

func resourceDomainSAMLOptionsPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	config := awstypes.AdvancedSecurityOptionsInput{}
	config.SAMLOptions = expandESSAMLOptions(d.Get("saml_options").([]any))

	log.Printf("[DEBUG] Updating OpenSearch domain SAML Options %#v", config)

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout, func() (any, error) {
		return conn.UpdateDomainConfig(ctx, &opensearch.UpdateDomainConfigInput{
			DomainName:              aws.String(domainName),
			AdvancedSecurityOptions: &config,
		})
	}, "A change/update is in progress. Please wait for it to complete before requesting another change.")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	d.SetId(domainName)

	if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain SAML Options (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceDomainSAMLOptionsRead(ctx, d, meta)...)
}

func resourceDomainSAMLOptionsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	config := awstypes.AdvancedSecurityOptionsInput{}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout, func() (any, error) {
		return conn.UpdateDomainConfig(ctx, &opensearch.UpdateDomainConfigInput{
			DomainName:              aws.String(domainName),
			AdvancedSecurityOptions: &config,
		})
	}, "A change/update is in progress. Please wait for it to complete before requesting another change.")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Domain SAML Options (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for OpenSearch domain SAML Options %q to be deleted", d.Get(names.AttrDomainName).(string))

	if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Domain SAML Options (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
