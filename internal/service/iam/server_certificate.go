// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_server_certificate", name="Server Certificate")
// @Tags(identifierAttribute="name", resourceType="ServerCertificate")
// @Testing(tagsTest=false)
func resourceServerCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServerCertificateCreate,
		ReadWithoutTimeout:   resourceServerCertificateRead,
		UpdateWithoutTimeout: resourceServerCertificateUpdate,
		DeleteWithoutTimeout: resourceServerCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceServerCertificateImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
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
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(0, 128-id.UniqueIDSuffixLength),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServerCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	sslCertName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &iam.UploadServerCertificateInput{
		CertificateBody:       aws.String(d.Get("certificate_body").(string)),
		PrivateKey:            aws.String(d.Get("private_key").(string)),
		ServerCertificateName: aws.String(sslCertName),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("certificate_chain"); ok {
		input.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path"); ok {
		input.Path = aws.String(v.(string))
	}

	output, err := conn.UploadServerCertificateWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.UploadServerCertificateWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Server Certificate (%s): %s", sslCertName, err)
	}

	d.SetId(aws.StringValue(output.ServerCertificateMetadata.ServerCertificateId))
	d.Set("name", sslCertName) // Required for resource Read.

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := serverCertificateCreateTags(ctx, conn, sslCertName, tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceServerCertificateRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Server Certificate (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServerCertificateRead(ctx, d, meta)...)
}

func resourceServerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	cert, err := findServerCertificateByName(ctx, conn, d.Get("name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Server Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Server Certificate (%s): %s", d.Id(), err)
	}

	metadata := cert.ServerCertificateMetadata
	d.SetId(aws.StringValue(metadata.ServerCertificateId))
	d.Set("arn", metadata.Arn)
	d.Set("certificate_body", cert.CertificateBody)
	d.Set("certificate_chain", cert.CertificateChain)
	if metadata.Expiration != nil {
		d.Set("expiration", aws.TimeValue(metadata.Expiration).Format(time.RFC3339))
	} else {
		d.Set("expiration", nil)
	}
	d.Set("name", metadata.ServerCertificateName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(metadata.ServerCertificateName)))
	d.Set("path", metadata.Path)
	if metadata.UploadDate != nil {
		d.Set("upload_date", aws.TimeValue(metadata.UploadDate).Format(time.RFC3339))
	} else {
		d.Set("upload_date", nil)
	}

	setTagsOut(ctx, cert.Tags)

	return diags
}

func resourceServerCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceServerCertificateRead(ctx, d, meta)...)
}

func resourceServerCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	log.Printf("[DEBUG] Deleting IAM Server Certificate: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 15*time.Minute, func() (interface{}, error) {
		return conn.DeleteServerCertificateWithContext(ctx, &iam.DeleteServerCertificateInput{
			ServerCertificateName: aws.String(d.Get("name").(string)),
		})
	}, iam.ErrCodeDeleteConflictException, "currently in use by arn")

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Server Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceServerCertificateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	// private_key can't be fetched from any API call
	return []*schema.ResourceData{d}, nil
}

func findServerCertificateByName(ctx context.Context, conn *iam.IAM, name string) (*iam.ServerCertificate, error) {
	input := &iam.GetServerCertificateInput{
		ServerCertificateName: aws.String(name),
	}

	output, err := conn.GetServerCertificateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServerCertificate == nil || output.ServerCertificate.ServerCertificateMetadata == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ServerCertificate, nil
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
		rawCert = aws.StringValue(cert)
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
