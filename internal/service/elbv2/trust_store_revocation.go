// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_lb_trust_store_revocation", name="Trust Store Revocation")
// @SDKResource("aws_alb_trust_store_revocation", name="Trust Store Revocation")
func ResourceTrustStoreRevocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustStoreRevocationCreate,
		ReadWithoutTimeout:   resourceTrustStoreRevocationRead,
		UpdateWithoutTimeout: nil,
		DeleteWithoutTimeout: resourceTrustStoreRevocationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceTrustStoreRevocationImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"trust_store_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
			"revocation_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceTrustStoreRevocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	trustStoreARN := d.Get("trust_store_arn").(string)

	_, err := FindTrustStoreByARN(ctx, conn, trustStoreARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Trust Store Revocations, Trust Store not found: %s: %s", trustStoreARN, err)
	}

	input := &elbv2.AddTrustStoreRevocationsInput{
		TrustStoreArn:      aws.String(trustStoreARN),
		RevocationContents: make([]*elbv2.RevocationContent, 1),
	}

	s3Bucket := d.Get("revocations_s3_bucket").(string)
	s3Key := d.Get("revocations_s3_key").(string)

	input.RevocationContents[0] = &elbv2.RevocationContent{
		S3Bucket: aws.String(s3Bucket),
		S3Key:    aws.String(s3Key),
	}

	if d.Get("revocations_s3_object_version").(string) != "" {
		input.RevocationContents[0].S3ObjectVersion = aws.String(d.Get("revocations_s3_object_version").(string))
	}

	output, err := conn.AddTrustStoreRevocationsWithContext(ctx, input)

	if err != nil {
		sdkdiag.AppendErrorf(diags, "creating ELBv2 Trust Store Revocations from %s s3://%s/%s: %s", trustStoreARN, s3Bucket, s3Key, err)
	}

	if len(output.TrustStoreRevocations) == 0 {
		return sdkdiag.AppendErrorf(diags, "creatingTrust Store Revocations: no revocations returned in response")
	}

	revocationID := aws.Int64Value(output.TrustStoreRevocations[0].RevocationId)
	d.SetId(TrustStoreRevocationCreateID(aws.StringValue(output.TrustStoreRevocations[0].TrustStoreArn), revocationID))

	d.Set("revocation_id", revocationID)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTrustStoreRevocation(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Trust Store Revocation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTrustStoreRevocationRead(ctx, d, meta)...)
}

func resourceTrustStoreRevocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTrustStoreRevocation(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Trust Store Revocation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store Revocation (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTrustStoreRevocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	const (
		trustStoreDeleteTimeout = 2 * time.Minute
	)

	parsed, err := parseTrustStoreRevocationID(d.Id())

	if err != nil {
		log.Printf("[DEBUG] error parsing Trust Store Revocation Id: %s", d.Id())
		return sdkdiag.AppendErrorf(diags, "deleting Trust Store: %s", err)
	}

	input := &elbv2.RemoveTrustStoreRevocationsInput{
		TrustStoreArn: aws.String(parsed.TrustStoreARN),
		RevocationIds: make([]*int64, 1),
	}
	input.RevocationIds[0] = aws.Int64(parsed.RevocationID)

	log.Printf("[DEBUG] Deleting Trust Store Revocation (%s): %s", d.Id(), input)
	err = retry.RetryContext(ctx, trustStoreDeleteTimeout, func() *retry.RetryError {
		_, err := conn.RemoveTrustStoreRevocationsWithContext(ctx, input)

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.RemoveTrustStoreRevocationsWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Trust Store Revocation: %s", err)
	}

	return diags
}

const trustStoreRevocationIDSeparator = "_"

func parseTrustStoreRevocationID(id string) (*TrustStoreRevocationID, error) {
	invalidIDError := func(msg string) error {
		return fmt.Errorf("unexpected format for ID (%q), expected TRUSTSTOREARN_REVOCATIONID: %s", id, msg)
	}

	parts := strings.Split(id, trustStoreRevocationIDSeparator)

	if len(parts) != 2 {
		return nil, invalidIDError("id should have two parts")
	}

	revocationID, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		return nil, invalidIDError("failed to parse revocationID")
	}

	result := &TrustStoreRevocationID{
		TrustStoreARN: parts[0],
		RevocationID:  revocationID,
	}

	return result, nil
}

func resourceTrustStoreRevocationImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parsed, err := parseTrustStoreRevocationID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("trust_store_arn", parsed.TrustStoreARN)
	d.Set("revocation_id", parsed.RevocationID)

	return []*schema.ResourceData{d}, nil
}

type TrustStoreRevocationID struct {
	TrustStoreARN string
	RevocationID  int64
}

func FindTrustStoreRevocation(ctx context.Context, conn *elbv2.ELBV2, id string) (*elbv2.DescribeTrustStoreRevocation, error) {
	parsed, err := parseTrustStoreRevocationID(id)
	if err != nil {
		log.Printf("[DEBUG] error parsing Trust Store Revocation Id: %s", id)
		return nil, err
	}

	input := &elbv2.DescribeTrustStoreRevocationsInput{
		TrustStoreArn: aws.String(parsed.TrustStoreARN),
	}
	var matched []*elbv2.DescribeTrustStoreRevocation

	err = conn.DescribeTrustStoreRevocationsPagesWithContext(ctx, input, func(page *elbv2.DescribeTrustStoreRevocationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrustStoreRevocations {
			if v != nil && aws.Int64Value(v.RevocationId) == parsed.RevocationID {
				matched = append(matched, v)
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

	count := len(matched)

	if count == 0 {
		return nil, nil
	}

	if count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return matched[0], nil
}

func TrustStoreRevocationCreateID(trustStoreARN string, revocationID int64) string {
	return fmt.Sprintf(
		"%s%s%d",
		trustStoreARN,
		trustStoreRevocationIDSeparator,
		revocationID,
	)
}
