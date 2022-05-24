package redshift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHSMClientCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceHSMClientCertificateCreate,
		Read:   resourceHSMClientCertificateRead,
		Update: resourceHSMClientCertificateUpdate,
		Delete: resourceHSMClientCertificateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceHSMClientCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	certIdentifier := d.Get("hsm_client_certificate_identifier").(string)

	input := redshift.CreateHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(certIdentifier),
	}

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateHsmClientCertificate(&input)
	if err != nil {
		return fmt.Errorf("error creating Redshift Hsm Client Certificate (%s): %s", certIdentifier, err)
	}

	d.SetId(aws.StringValue(out.HsmClientCertificate.HsmClientCertificateIdentifier))

	return resourceHSMClientCertificateRead(d, meta)
}

func resourceHSMClientCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindHSMClientCertificateByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Hsm Client Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Hsm Client Certificate (%s): %w", d.Id(), err)
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
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceHSMClientCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Hsm Client Certificate (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceHSMClientCertificateRead(d, meta)
}

func resourceHSMClientCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	deleteInput := redshift.DeleteHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift Hsm Client Certificate: %s", d.Id())
	_, err := conn.DeleteHsmClientCertificate(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmClientCertificateNotFoundFault) {
			return nil
		}
		return err
	}

	return err
}
