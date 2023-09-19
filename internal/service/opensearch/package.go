// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_opensearch_package")
func ResourcePackage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourcePackageCreate,
		ReadWithoutTimeout:   ResourcePackageRead,
		UpdateWithoutTimeout: ResourcePackageUpdate,
		DeleteWithoutTimeout: ResourcePackageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"available_package_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"package_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"package_source": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"s3_key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"package_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(opensearchservice.PackageType_Values(), false),
			},
		},
	}
}

func ResourcePackageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.CreatePackageInput{
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageName:        aws.String(d.Get("package_name").(string)),
		PackageType:        aws.String(d.Get("package_type").(string)),
	}

	if v, ok := d.GetOk("package_source"); ok {
		input.PackageSource = expandPackageSource(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreatePackageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Package: %s", err)
	}

	d.SetId(aws.StringValue(output.PackageDetails.PackageID))

	return append(diags, ResourcePackageRead(ctx, d, meta)...)
}

func ResourcePackageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.DescribePackagesInput{
		Filters: []*opensearchservice.DescribePackagesFilter{
			{
				Name:  aws.String("PackageID"),
				Value: []*string{aws.String(d.Id())},
			},
		},
	}

	output, err := conn.DescribePackagesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Package (%s): %s", d.Id(), err)
	}

	if output == nil || len(output.PackageDetailsList) == 0 {
		d.SetId("")
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Package (%s): not found", d.Id())
	}

	d.Set("package_id", output.PackageDetailsList[0].PackageID)
	d.Set("package_name", output.PackageDetailsList[0].PackageName)
	d.Set("package_description", output.PackageDetailsList[0].PackageDescription)
	d.Set("package_type", output.PackageDetailsList[0].PackageType)
	d.Set("available_package_version", output.PackageDetailsList[0].AvailablePackageVersion)

	return diags
}

func ResourcePackageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.UpdatePackageInput{
		PackageID:          aws.String(d.Id()),
		PackageDescription: aws.String(d.Get("package_description").(string)),
		PackageSource:      expandPackageSource(d.Get("package_source").([]interface{})[0].(map[string]interface{})),
	}

	_, err := conn.UpdatePackageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpenSearch Package (%s): %s", d.Id(), err)
	}

	return append(diags, ResourcePackageRead(ctx, d, meta)...)
}

func ResourcePackageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Package: %s", d.Id())
	_, err := conn.DeletePackageWithContext(ctx, &opensearchservice.DeletePackageInput{
		PackageID: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Package (%s): %s", d.Id(), err)
	}

	return diags
}

func expandPackageSource(v interface{}) *opensearchservice.PackageSource {
	if v == nil {
		return nil
	}

	return &opensearchservice.PackageSource{
		S3BucketName: aws.String(v.(map[string]interface{})["s3_bucket_name"].(string)),
		S3Key:        aws.String(v.(map[string]interface{})["s3_key"].(string)),
	}
}
