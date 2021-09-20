package aws

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceServerCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCertificateCreate,
		Read:   resourceServerCertificateRead,
		Update: resourceServerCertificateUpdate,
		Delete: resourceServerCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIAMServerCertificateImport,
		},

		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNormalizeCertRemoval,
				StateFunc:        StateTrimSpace,
			},

			"certificate_chain": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressNormalizeCertRemoval,
				StateFunc:        StateTrimSpace,
			},

			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},

			"private_key": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Sensitive:        true,
				DiffSuppressFunc: suppressNormalizeCertRemoval,
				StateFunc:        StateTrimSpace,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 128),
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(0, 128-resource.UniqueIDSuffixLength),
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServerCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var sslCertName string
	if v, ok := d.GetOk("name"); ok {
		sslCertName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		sslCertName = resource.PrefixedUniqueId(v.(string))
	} else {
		sslCertName = resource.UniqueId()
	}

	createOpts := &iam.UploadServerCertificateInput{
		CertificateBody:       aws.String(d.Get("certificate_body").(string)),
		PrivateKey:            aws.String(d.Get("private_key").(string)),
		ServerCertificateName: aws.String(sslCertName),
		Tags:                  tags.IgnoreAws().IamTags(),
	}

	if v, ok := d.GetOk("certificate_chain"); ok {
		createOpts.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path"); ok {
		createOpts.Path = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IAM Server Certificate with opts: %s", createOpts)
	resp, err := conn.UploadServerCertificate(createOpts)
	if err != nil {
		return fmt.Errorf("error uploading server certificate: %w", err)
	}

	d.SetId(aws.StringValue(resp.ServerCertificateMetadata.ServerCertificateId))
	d.Set("name", sslCertName)

	return resourceServerCertificateRead(d, meta)
}

func resourceServerCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetServerCertificate(&iam.GetServerCertificateInput{
		ServerCertificateName: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM Server Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Server Certificate (%s): %w", d.Id(), err)
	}

	cert := resp.ServerCertificate
	metadata := cert.ServerCertificateMetadata
	d.SetId(aws.StringValue(metadata.ServerCertificateId))

	d.Set("certificate_body", cert.CertificateBody)
	d.Set("certificate_chain", cert.CertificateChain)
	d.Set("path", metadata.Path)
	d.Set("arn", metadata.Arn)
	if metadata.Expiration != nil {
		d.Set("expiration", aws.TimeValue(metadata.Expiration).Format(time.RFC3339))
	} else {
		d.Set("expiration", nil)
	}

	if metadata.UploadDate != nil {
		d.Set("upload_date", aws.TimeValue(metadata.UploadDate).Format(time.RFC3339))
	} else {
		d.Set("upload_date", nil)
	}

	tags := tftags.IamKeyValueTags(cert.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceServerCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := serverCertificateUpdateTags(conn, d.Get("name").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for IAM Server Certificate (%s): %w", d.Get("name").(string), err)
		}
	}

	return resourceServerCertificateRead(d, meta)
}

func resourceServerCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	log.Printf("[INFO] Deleting IAM Server Certificate: %s", d.Id())
	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteServerCertificate(&iam.DeleteServerCertificateInput{
			ServerCertificateName: aws.String(d.Get("name").(string)),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == iam.ErrCodeDeleteConflictException && strings.Contains(awsErr.Message(), "currently in use by arn") {
					currentlyInUseBy(awsErr.Message(), meta.(*conns.AWSClient).ELBConn)
					log.Printf("[WARN] Conflict deleting server certificate: %s, retrying", awsErr.Message())
					return resource.RetryableError(err)
				}
				if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
					return nil
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteServerCertificate(&iam.DeleteServerCertificateInput{
			ServerCertificateName: aws.String(d.Get("name").(string)),
		})
	}

	return err
}

func resourceAwsIAMServerCertificateImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	// private_key can't be fetched from any API call
	return []*schema.ResourceData{d}, nil
}

func currentlyInUseBy(awsErr string, conn *elb.ELB) {
	r := regexp.MustCompile(`currently in use by ([a-z0-9:-]+)\/([a-z0-9-]+)\.`)
	matches := r.FindStringSubmatch(awsErr)
	if len(matches) > 0 {
		lbName := matches[2]
		describeElbOpts := &elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []*string{aws.String(lbName)},
		}
		if _, err := conn.DescribeLoadBalancers(describeElbOpts); err != nil {
			if tfawserr.ErrMessageContains(err, "LoadBalancerNotFound", "") {
				log.Printf("[WARN] Load Balancer (%s) causing delete conflict not found", lbName)
			}
		}
	}
}

func normalizeCert(cert interface{}) string {
	if cert == nil || cert == (*string)(nil) {
		return ""
	}

	var rawCert string
	switch cert := cert.(type) {
	case string:
		rawCert = cert
	case *string:
		rawCert = *cert
	default:
		return ""
	}

	cleanVal := sha1.Sum(stripCR([]byte(strings.TrimSpace(rawCert))))
	return hex.EncodeToString(cleanVal[:])
}

// strip CRs from raw literals. Lifted from go/scanner/scanner.go
// See https://github.com/golang/go/blob/release-branch.go1.6/src/go/scanner/scanner.go#L479
func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

// Terraform AWS Provider version 3.0.0 removed state hash storage.
// This DiffSuppressFunc prevents the resource from triggering needless recreation.
func suppressNormalizeCertRemoval(k, old, new string, d *schema.ResourceData) bool {
	return normalizeCert(new) == old
}
