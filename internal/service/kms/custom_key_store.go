// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_kms_custom_key_store", name="Custom Key Store")
func resourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomKeyStoreCreate,
		ReadWithoutTimeout:   resourceCustomKeyStoreRead,
		UpdateWithoutTimeout: resourceCustomKeyStoreUpdate,
		DeleteWithoutTimeout: resourceCustomKeyStoreDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"custom_key_store_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_key_store_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CustomKeyStoreType](),
			},
			"key_store_password": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(7, 32)),
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"xks_proxy_authentication_credential": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"raw_secret_access_key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"xks_proxy_connectivity": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.XksProxyConnectivityType](),
			},
			"xks_proxy_uri_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"xks_proxy_uri_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"xks_proxy_vpc_endpoint_service_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCustomKeyStoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	name := d.Get("custom_key_store_name").(string)
	input := &kms.CreateCustomKeyStoreInput{
		CustomKeyStoreName: aws.String(name),
	}

	if v, ok := d.GetOk("custom_key_store_type"); ok {
		input.CustomKeyStoreType = awstypes.CustomKeyStoreType(v.(string))
	}

	if v, ok := d.GetOk("cloud_hsm_cluster_id"); ok {
		input.CloudHsmClusterId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key_store_password"); ok {
		input.KeyStorePassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("trust_anchor_certificate"); ok {
		input.TrustAnchorCertificate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_authentication_credential"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.XksProxyAuthenticationCredential = expandXksProxyAuthenticationCredential(v.([]any))
	}

	if v, ok := d.GetOk("xks_proxy_connectivity"); ok {
		input.XksProxyConnectivity = awstypes.XksProxyConnectivityType(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_uri_endpoint"); ok {
		input.XksProxyUriEndpoint = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_uri_path"); ok {
		input.XksProxyUriPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_vpc_endpoint_service_name"); ok {
		input.XksProxyVpcEndpointServiceName = aws.String(v.(string))
	}

	output, err := conn.CreateCustomKeyStore(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Custom Key Store (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CustomKeyStoreId))

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	output, err := findCustomKeyStoreByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Custom Key Store (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Custom Key Store (%s): %s", d.Id(), err)
	}

	d.Set("custom_key_store_name", output.CustomKeyStoreName)
	d.Set("cloud_hsm_cluster_id", output.CloudHsmClusterId)
	d.Set("custom_key_store_type", output.CustomKeyStoreType)
	d.Set("key_store_password", d.Get("key_store_password"))
	d.Set("trust_anchor_certificate", output.TrustAnchorCertificate)

	d.Set("xks_proxy_connectivity", output.XksProxyConfiguration.Connectivity)
	d.Set("xks_proxy_uri_endpoint", output.XksProxyConfiguration.UriEndpoint)
	d.Set("xks_proxy_uri_path", output.XksProxyConfiguration.UriPath)
	d.Set("xks_proxy_vpc_endpoint_service_name", output.XksProxyConfiguration.VpcEndpointServiceName)

	return diags
}

func resourceCustomKeyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	input := &kms.UpdateCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	}

	if d.HasChange("cloud_hsm_cluster_id") {
		input.CloudHsmClusterId = aws.String(d.Get("cloud_hsm_cluster_id").(string))
	}

	if d.HasChange("custom_key_store_name") {
		input.NewCustomKeyStoreName = aws.String(d.Get("custom_key_store_name").(string))
	}

	if d.HasChange("key_store_password") {
		input.KeyStorePassword = aws.String(d.Get("key_store_password").(string))
	}

	if d.HasChange("xks_proxy_authentication_credential") {
		input.XksProxyAuthenticationCredential = expandXksProxyAuthenticationCredential(d.Get("xks_proxy_authentication_credential").(*schema.Set).List())
	}

	if d.HasChange("xks_proxy_connectivity") {
		input.XksProxyConnectivity = awstypes.XksProxyConnectivityType(d.Get("xks_proxy_connectivity").(string))
	}

	if d.HasChange("xks_proxy_uri_endpoint") {
		input.XksProxyUriEndpoint = aws.String(d.Get("xks_proxy_uri_endpoint").(string))
	}

	if d.HasChange("xks_proxy_uri_path") {
		input.XksProxyUriPath = aws.String(d.Get("xks_proxy_uri_path").(string))
	}

	if d.HasChange("xks_proxy_vpc_endpoint_service_name") {
		input.XksProxyVpcEndpointServiceName = aws.String(d.Get("xks_proxy_vpc_endpoint_service_name").(string))
	}

	_, err := conn.UpdateCustomKeyStore(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating KMS Custom Key Store (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	log.Printf("[INFO] Deleting KMS Custom Key Store: %s", d.Id())
	_, err := conn.DeleteCustomKeyStore(ctx, &kms.DeleteCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	return diags
}

func findCustomKeyStoreByID(ctx context.Context, conn *kms.Client, id string) (*awstypes.CustomKeyStoresListEntry, error) {
	input := &kms.DescribeCustomKeyStoresInput{
		CustomKeyStoreId: aws.String(id),
	}

	return findCustomKeyStore(ctx, conn, input, tfslices.PredicateTrue[*awstypes.CustomKeyStoresListEntry]())
}

func findCustomKeyStore(ctx context.Context, conn *kms.Client, input *kms.DescribeCustomKeyStoresInput, filter tfslices.Predicate[*awstypes.CustomKeyStoresListEntry]) (*awstypes.CustomKeyStoresListEntry, error) {
	output, err := findCustomKeyStores(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCustomKeyStores(ctx context.Context, conn *kms.Client, input *kms.DescribeCustomKeyStoresInput, filter tfslices.Predicate[*awstypes.CustomKeyStoresListEntry]) ([]awstypes.CustomKeyStoresListEntry, error) {
	var output []awstypes.CustomKeyStoresListEntry

	pages := kms.NewDescribeCustomKeyStoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.CustomKeyStoreNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return output, err
		}

		for _, v := range page.CustomKeyStores {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandXksProxyAuthenticationCredential(tfList []any) *awstypes.XksProxyAuthenticationCredentialType {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.XksProxyAuthenticationCredentialType{}

	if v, ok := tfMap["access_key_id"].(string); ok {
		apiObject.AccessKeyId = aws.String(v)
	}

	if v, ok := tfMap["raw_secret_access_key"].(string); ok {
		apiObject.RawSecretAccessKey = aws.String(v)
	}

	return apiObject
}
