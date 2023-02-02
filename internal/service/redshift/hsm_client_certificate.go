package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHSMClientCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHSMClientCertificateCreate,
		ReadWithoutTimeout:   resourceHSMClientCertificateRead,
		UpdateWithoutTimeout: resourceHSMClientCertificateUpdate,
		DeleteWithoutTimeout: resourceHSMClientCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_client_certificate_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_client_certificate_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHSMClientCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	certIdentifier := d.Get("hsm_client_certificate_identifier").(string)

	input := redshift.CreateHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(certIdentifier),
	}

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateHsmClientCertificateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift HSM Client Certificate (%s): %s", certIdentifier, err)
	}

	d.SetId(aws.StringValue(out.HsmClientCertificate.HsmClientCertificateIdentifier))

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindHSMClientCertificateByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift HSM Client Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift HSM Client Certificate (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("hsmclientcertificate:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	d.Set("hsm_client_certificate_identifier", out.HsmClientCertificateIdentifier)
	d.Set("hsm_client_certificate_public_key", out.HsmClientCertificatePublicKey)

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceHSMClientCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift HSM Client Certificate (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	deleteInput := redshift.DeleteHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift HSM Client Certificate: %s", d.Id())
	_, err := conn.DeleteHsmClientCertificateWithContext(ctx, &deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmClientCertificateNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "updating Redshift HSM Client Certificate (%s) tags: %s", d.Get("arn").(string), err)
	}

	return diags
}
