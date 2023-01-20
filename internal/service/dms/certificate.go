package dms

import (
	"context"
	"encoding/base64"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-]+$"), "must start with a letter, only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"certificate_wallet": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	certificateID := d.Get("certificate_id").(string)

	request := &dms.ImportCertificateInput{
		CertificateIdentifier: aws.String(certificateID),
		Tags:                  Tags(tags.IgnoreAWS()),
	}

	pem, pemSet := d.GetOk("certificate_pem")
	wallet, walletSet := d.GetOk("certificate_wallet")

	if !pemSet && !walletSet {
		return sdkdiag.AppendErrorf(diags, "Must set either certificate_pem or certificate_wallet for DMS Certificate (%s)", certificateID)
	}
	if pemSet && walletSet {
		return sdkdiag.AppendErrorf(diags, "Cannot set both certificate_pem and certificate_wallet for DMS Certificate (%s)", certificateID)
	}

	if pemSet {
		request.CertificatePem = aws.String(pem.(string))
	}
	if walletSet {
		certWallet, err := base64.StdEncoding.DecodeString(wallet.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Base64 decoding certificate_wallet for DMS Certificate (%s): %s", certificateID, err)
		}
		request.CertificateWallet = certWallet
	}

	_, err := conn.ImportCertificateWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS certificate (%s): %s", certificateID, err)
	}

	d.SetId(certificateID)
	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	response, err := conn.DescribeCertificatesWithContext(ctx, &dms.DescribeCertificatesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []*string{aws.String(d.Id())}, // Must use d.Id() to work with import.
			},
		},
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Certificate (%s): %s", d.Id(), err)
	}

	if response == nil || len(response.Certificates) == 0 || response.Certificates[0] == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading DMS Certificate (%s): not found", d.Id())
		}
		log.Printf("[WARN] DMS Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	resourceCertificateSetState(d, response.Certificates[0])

	tags, err := ListTags(ctx, conn, d.Get("certificate_arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DMS Certificate (%s): %s", d.Get("certificate_arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn()

	if d.HasChange("tags_all") {
		arn := d.Get("certificate_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Certificate (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn()

	request := &dms.DeleteCertificateInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
	}

	_, err := conn.DeleteCertificateWithContext(ctx, request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DMS Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceCertificateSetState(d *schema.ResourceData, cert *dms.Certificate) {
	d.SetId(aws.StringValue(cert.CertificateIdentifier))

	d.Set("certificate_id", cert.CertificateIdentifier)
	d.Set("certificate_arn", cert.CertificateArn)

	if aws.StringValue(cert.CertificatePem) != "" {
		d.Set("certificate_pem", cert.CertificatePem)
	}
	if cert.CertificateWallet != nil && len(cert.CertificateWallet) != 0 {
		d.Set("certificate_wallet", verify.Base64Encode(cert.CertificateWallet))
	}
}
