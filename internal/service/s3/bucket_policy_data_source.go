// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_s3_bucket_policy")
func DataSourceBucketPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	name := d.Get("bucket").(string)

	out, err := FindBucketPolicy(ctx, conn, name)
	if err != nil {
		return diag.Errorf("failed getting S3 bucket policy (%s): %s", name, err)
	}

	policy, err := structure.NormalizeJsonString(aws.StringValue(out.Policy))
	if err != nil {
		return diag.Errorf("policy (%s) is an invalid JSON: %s", policy, err)
	}

	d.SetId(name)
	d.Set("policy", policy)

	return nil
}

func FindBucketPolicy(ctx context.Context, conn *s3.S3, name string) (*s3.GetBucketPolicyOutput, error) {
	in := &s3.GetBucketPolicyInput{
		Bucket: aws.String(name),
	}
	log.Printf("[DEBUG] Reading S3 bucket policy: %s", in)

	out, err := conn.GetBucketPolicyWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
