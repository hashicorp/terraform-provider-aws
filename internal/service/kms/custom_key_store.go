package kms

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameCustomKeyStore = "Custom Key Store"
)

func ResourceCustomKeyStore() *schema.Resource {
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
			},
			"custom_key_store_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_key_store_type": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(kms.CustomKeyStoreType_Values(), false),
			},
			"key_store_password": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(7, 32)),
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"xks_proxy_authentication_credential": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"raw_secret_access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"xks_proxy_connectivity": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceCustomKeyStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	in := &kms.CreateCustomKeyStoreInput{
		CustomKeyStoreName: aws.String(d.Get("custom_key_store_name").(string)),
	}

	if v, ok := d.GetOk("custom_key_store_type"); ok {
		in.CustomKeyStoreType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloud_hsm_cluster_id"); ok {
		in.CloudHsmClusterId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key_store_password"); ok {
		in.KeyStorePassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("trust_anchor_certificate"); ok {
		in.TrustAnchorCertificate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_authentication_credential"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.XksProxyAuthenticationCredential = expandXksProxyAuthenticationCredential(v.([]interface{}))
	}

	if v, ok := d.GetOk("xks_proxy_connectivity"); ok {
		in.XksProxyConnectivity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_uri_endpoint"); ok {
		in.XksProxyUriEndpoint = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_uri_path"); ok {
		in.XksProxyUriPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("xks_proxy_vpc_endpoint_service_name"); ok {
		in.XksProxyVpcEndpointServiceName = aws.String(v.(string))
	}

	out, err := conn.CreateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.CustomKeyStoreId))

	return resourceCustomKeyStoreRead(ctx, d, meta)
}

func resourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	in := &kms.DescribeCustomKeyStoresInput{
		CustomKeyStoreId: aws.String(d.Id()),
	}
	out, err := FindCustomKeyStoreByID(ctx, conn, in)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS CustomKeyStore (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionReading, ResNameCustomKeyStore, d.Id(), err)
	}

	d.Set("custom_key_store_name", out.CustomKeyStoreName)
	d.Set("custom_key_store_type", out.CustomKeyStoreType)
	d.Set("cloud_hsm_cluster_id", out.CloudHsmClusterId)
	d.Set("trust_anchor_certificate", out.TrustAnchorCertificate)

	d.Set("xks_proxy_connectivity", out.XksProxyConfiguration.Connectivity)
	d.Set("xks_proxy_uri_endpoint", out.XksProxyConfiguration.UriEndpoint)
	d.Set("xks_proxy_uri_path", out.XksProxyConfiguration.UriPath)
	d.Set("xks_proxy_vpc_endpoint_service_name", out.XksProxyConfiguration.VpcEndpointServiceName)

	return nil
}

func resourceCustomKeyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	update := false

	in := &kms.UpdateCustomKeyStoreInput{
		CustomKeyStoreId:  aws.String(d.Id()),
		CloudHsmClusterId: aws.String(d.Get("cloud_hsm_cluster_id").(string)),
	}

	if d.HasChange("key_store_password") {
		in.KeyStorePassword = aws.String(d.Get("key_store_password").(string))
		update = true
	}

	if d.HasChange("custom_key_store_name") {
		in.NewCustomKeyStoreName = aws.String(d.Get("custom_key_store_name").(string))
		update = true
	}

	if d.HasChange("xks_proxy_authentication_credential") {
		in.XksProxyAuthenticationCredential = expandXksProxyAuthenticationCredential(d.Get("xks_proxy_authentication_credential").(*schema.Set).List())
		update = true
	}

	if d.HasChange("xks_proxy_connectivity") {
		in.XksProxyConnectivity = aws.String(d.Get("xks_proxy_connectivity").(string))
		update = true
	}

	if d.HasChange("xks_proxy_uri_endpoint") {
		in.XksProxyUriEndpoint = aws.String(d.Get("xks_proxy_uri_endpoint").(string))
		update = true
	}

	if d.HasChange("xks_proxy_uri_path") {
		in.XksProxyUriPath = aws.String(d.Get("xks_proxy_uri_path").(string))
		update = true
	}

	if d.HasChange("xks_proxy_vpc_endpoint_service_name") {
		in.XksProxyVpcEndpointServiceName = aws.String(d.Get("xks_proxy_vpc_endpoint_service_name").(string))
		update = true
	}

	if !update {
		return nil
	}

	_, err := conn.UpdateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionUpdating, ResNameCustomKeyStore, d.Id(), err)
	}

	return resourceCustomKeyStoreRead(ctx, d, meta)
}

func resourceCustomKeyStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	log.Printf("[INFO] Deleting KMS CustomKeyStore %s", d.Id())

	_, err := conn.DeleteCustomKeyStoreWithContext(ctx, &kms.DeleteCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil
	}

	return nil
}

func expandXksProxyAuthenticationCredential(l []interface{}) *kms.XksProxyAuthenticationCredentialType {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &kms.XksProxyAuthenticationCredentialType{}

	if v, ok := tfMap["access_key_id"].(string); ok {
		result.AccessKeyId = aws.String(v)
	}

	if v, ok := tfMap["raw_secret_access_key"].(string); ok {
		result.RawSecretAccessKey = aws.String(v)
	}

	return result
}

func flattenXksProxyAuthenticationCredential(config *kms.XksProxyAuthenticationCredentialType) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"access_key_id":         aws.StringValue(config.AccessKeyId),
		"raw_secret_access_key": aws.StringValue(config.RawSecretAccessKey),
	}

	return []interface{}{m}
}
