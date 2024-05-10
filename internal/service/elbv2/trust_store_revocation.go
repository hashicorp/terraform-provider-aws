// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_lb_trust_store_revocation", name="Trust Store Revocation")
func ResourceTrustStoreRevocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustStoreRevocationCreate,
		ReadWithoutTimeout:   resourceTrustStoreRevocationRead,
		DeleteWithoutTimeout: resourceTrustStoreRevocationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"revocation_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"revocations_s3_bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"revocations_s3_key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"revocations_s3_object_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"trust_store_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

const (
	trustStoreRevocationResourceIDPartCount = 2
)

func resourceTrustStoreRevocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	s3Bucket := d.Get("revocations_s3_bucket").(string)
	s3Key := d.Get("revocations_s3_key").(string)
	trustStoreARN := d.Get("trust_store_arn").(string)
	input := &elbv2.AddTrustStoreRevocationsInput{
		RevocationContents: []*elbv2.RevocationContent{{
			S3Bucket: aws.String(s3Bucket),
			S3Key:    aws.String(s3Key),
		}},
		TrustStoreArn: aws.String(trustStoreARN),
	}

	if v, ok := d.GetOk("revocations_s3_object_version"); ok {
		input.RevocationContents[0].S3ObjectVersion = aws.String(v.(string))
	}

	output, err := conn.AddTrustStoreRevocationsWithContext(ctx, input)

	if err != nil {
		sdkdiag.AppendErrorf(diags, "creating ELBv2 Trust Store (%s) Revocation (s3://%s/%s): %s", trustStoreARN, s3Bucket, s3Key, err)
	}

	revocationID := aws.Int64Value(output.TrustStoreRevocations[0].RevocationId)
	id := errs.Must(flex.FlattenResourceId([]string{trustStoreARN, strconv.FormatInt(revocationID, 10)}, trustStoreRevocationResourceIDPartCount, false))

	d.SetId(id)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindTrustStoreRevocationByTwoPartKey(ctx, conn, trustStoreARN, revocationID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Trust Store Revocation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTrustStoreRevocationRead(ctx, d, meta)...)
}

func resourceTrustStoreRevocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), trustStoreRevocationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	trustStoreARN := parts[0]
	revocationID := errs.Must(strconv.ParseInt(parts[1], 10, 64))
	revocation, err := FindTrustStoreRevocationByTwoPartKey(ctx, conn, trustStoreARN, revocationID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Trust Store Revocation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store Revocation (%s): %s", d.Id(), err)
	}

	d.Set("revocation_id", revocation.RevocationId)
	d.Set("trust_store_arn", revocation.TrustStoreArn)

	return diags
}

func resourceTrustStoreRevocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), trustStoreRevocationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	trustStoreARN := parts[0]
	revocationID := errs.Must(strconv.ParseInt(parts[1], 10, 64))

	log.Printf("[DEBUG] Deleting ELBv2 Trust Store Revocation: %s", d.Id())
	_, err = conn.RemoveTrustStoreRevocationsWithContext(ctx, &elbv2.RemoveTrustStoreRevocationsInput{
		RevocationIds: aws.Int64Slice([]int64{revocationID}),
		TrustStoreArn: aws.String(trustStoreARN),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Trust Store Revocation (%s): %s", d.Id(), err)
	}

	return diags
}

func FindTrustStoreRevocationByTwoPartKey(ctx context.Context, conn *elbv2.ELBV2, trustStoreARN string, revocationID int64) (*elbv2.DescribeTrustStoreRevocation, error) {
	input := &elbv2.DescribeTrustStoreRevocationsInput{
		RevocationIds: aws.Int64Slice([]int64{revocationID}),
		TrustStoreArn: aws.String(trustStoreARN),
	}
	output, err := findTrustStoreRevocation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrustStoreArn) != trustStoreARN || aws.Int64Value(output.RevocationId) != revocationID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTrustStoreRevocation(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoreRevocationsInput) (*elbv2.DescribeTrustStoreRevocation, error) {
	output, err := findTrustStoreRevocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findTrustStoreRevocations(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoreRevocationsInput) ([]*elbv2.DescribeTrustStoreRevocation, error) {
	var output []*elbv2.DescribeTrustStoreRevocation

	err := conn.DescribeTrustStoreRevocationsPagesWithContext(ctx, input, func(page *elbv2.DescribeTrustStoreRevocationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrustStoreRevocations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTrustStoreNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
