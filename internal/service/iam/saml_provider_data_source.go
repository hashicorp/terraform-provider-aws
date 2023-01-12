package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSAMLProviderRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_metadata_document": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	output, err := FindSAMLProviderByARN(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("reading IAM SAML Provider (%s): %s", arn, err)
	}

	name, err := nameFromSAMLProviderARN(arn)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(arn)
	if output.CreateDate != nil {
		d.Set("create_date", aws.TimeValue(output.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}
	d.Set("name", name)
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.TimeValue(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
