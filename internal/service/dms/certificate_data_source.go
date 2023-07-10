// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_certificate")
func DataSourceCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCertificateRead,

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-]+$"), "must start with a letter, only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"certificate_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_pem": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"certificate_wallet": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"key_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"signing_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_from_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_to_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameCertificate = "Certificate Data Source"
)

func dataSourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DMSConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	certificateID := d.Get("certificate_id").(string)

	out, err := FindCertificateByID(ctx, conn, certificateID)

	if err != nil {
		create.DiagError(names.DMS, create.ErrActionReading, DSNameCertificate, d.Id(), err)
	}

	d.SetId(aws.StringValue(out.CertificateIdentifier))

	d.Set("certificate_id", out.CertificateIdentifier)
	d.Set("certificate_arn", out.CertificateArn)
	d.Set("certificate_pem", out.CertificatePem)

	if out.CertificateWallet != nil && len(out.CertificateWallet) != 0 {
		d.Set("certificate_wallet", verify.Base64Encode(out.CertificateWallet))
	}

	d.Set("key_length", out.KeyLength)
	d.Set("signing_algorithm", out.SigningAlgorithm)

	from_date := out.ValidFromDate.String()
	d.Set("valid_from_date", from_date)
	to_date := out.ValidToDate.String()
	d.Set("valid_to_date", to_date)

	tags, err := listTags(ctx, conn, aws.StringValue(out.CertificateArn))

	if err != nil {
		return create.DiagError(names.DMS, create.ErrActionReading, DSNameCertificate, d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.DMS, create.ErrActionSetting, DSNameCertificate, d.Id(), err)
	}

	return nil
}

func FindCertificateByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.Certificate, error) {
	input := &dms.DescribeCertificatesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("certificate-id"),
				Values: []*string{aws.String(id)},
			},
		},
	}
	response, err := conn.DescribeCertificatesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if response == nil || len(response.Certificates) == 0 || response.Certificates[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return response.Certificates[0], nil
}
