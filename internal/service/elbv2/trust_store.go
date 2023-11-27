// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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

// @SDKResource("aws_lb_trust_store", name="Trust Store")
// @SDKResource("aws_alb_trust_store", name="Trust Store")
// @Tags(identifierAttribute="id")
func ResourceTrustStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustStoreCreate,
		ReadWithoutTimeout:   resourceTrustStoreRead,
		UpdateWithoutTimeout: resourceTrustStoreUpdate,
		DeleteWithoutTimeout: resourceTrustStoreDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificates_bundle_s3_bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"ca_certificates_bundle_s3_key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"ca_certificates_bundle_s3_object_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validNamePrefix,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrustStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get("name").(string)),
		create.WithConfiguredPrefix(d.Get("name_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	exist, err := FindTrustStoreByName(ctx, conn, name)

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store (%s): %s", name, err)
	}

	if exist != nil {
		return sdkdiag.AppendErrorf(diags, "ELBv2 Trust Store (%s) already exists", name)
	}

	input := &elbv2.CreateTrustStoreInput{
		Name:                         aws.String(name),
		Tags:                         getTagsIn(ctx),
		CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
		CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
	}

	if d.Get("ca_certificates_bundle_s3_object_version").(string) != "" {
		input.CaCertificatesBundleS3ObjectVersion = aws.String(d.Get("ca_certificates_bundle_s3_object_version").(string))
	}

	output, err := conn.CreateTrustStoreWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateTrustStoreWithContext(ctx, input)
	}

	// Tags are not supported on creation with some protocol types(i.e. GENEVE)
	// Retry creation without tags
	if input.Tags != nil && tfawserr.ErrMessageContains(err, ErrValidationError, TagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = conn.CreateTrustStoreWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Trust Store (%s): %s", name, err)
	}

	if len(output.TrustStores) == 0 {
		return sdkdiag.AppendErrorf(diags, "creating Trust Store: no trust stores returned in response")
	}

	d.SetId(aws.StringValue(output.TrustStores[0].TrustStoreArn))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTrustStoreByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Trust Store (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceTrustStoreRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Trust Store (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrustStoreRead(ctx, d, meta)...)
}

func resourceTrustStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTrustStoreByARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Trust Store %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store (%s): %s", d.Id(), err)
	}

	trustStore := outputRaw.(*elbv2.TrustStore)

	d.Set("name", trustStore.Name)
	d.Set("arn", trustStore.TrustStoreArn)

	return diags
}

func resourceTrustStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	if d.HasChanges("ca_certificates_bundle_s3_bucket", "ca_certificates_bundle_s3_key", "ca_certificates_bundle_s3_object_version", "tags") {
		var params = &elbv2.ModifyTrustStoreInput{
			TrustStoreArn:                aws.String(d.Id()),
			CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
			CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
		}

		if d.Get("ca_certificates_bundle_s3_object_version").(string) != "" {
			params.CaCertificatesBundleS3ObjectVersion = aws.String(d.Get("ca_certificates_bundle_s3_object_version").(string))
		}

		_, err := conn.ModifyTrustStoreWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Trust Store: %s", err)
		}
	}

	return append(diags, resourceTrustStoreRead(ctx, d, meta)...)
}

func resourceTrustStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	const (
		trustStoreDeleteTimeout = 2 * time.Minute
	)
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	err := waitForNoTrustStoreAssociations(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Trust Store Associations (%s) to be removed: %s", d.Get("name").(string), err)
	}

	input := &elbv2.DeleteTrustStoreInput{
		TrustStoreArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Trust Store (%s): %s", d.Id(), input)
	err = retry.RetryContext(ctx, trustStoreDeleteTimeout, func() *retry.RetryError {
		_, err := conn.DeleteTrustStoreWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, "TrustStoreInUse", "is currently in use by a listener") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteTrustStoreWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Trust Store: %s", err)
	}

	return diags
}

func FindTrustStoreByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.TrustStore, error) {
	input := &elbv2.DescribeTrustStoresInput{
		TrustStoreArns: aws.StringSlice([]string{arn}),
	}

	output, err := FindTrustStore(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrustStoreArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrustStoreByName(ctx context.Context, conn *elbv2.ELBV2, name string) (*elbv2.TrustStore, error) {
	input := &elbv2.DescribeTrustStoresInput{
		Names: aws.StringSlice([]string{name}),
	}

	output, err := FindTrustStore(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.Name) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrustStores(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoresInput) ([]*elbv2.TrustStore, error) {
	var output []*elbv2.TrustStore

	err := conn.DescribeTrustStoresPagesWithContext(ctx, input, func(page *elbv2.DescribeTrustStoresOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrustStores {
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

func FindTrustStore(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoresInput) (*elbv2.TrustStore, error) {
	output, err := FindTrustStores(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func waitForNoTrustStoreAssociations(ctx context.Context, conn *elbv2.ELBV2, arn string, timeout time.Duration) error {
	input := &elbv2.DescribeTrustStoreAssociationsInput{
		TrustStoreArn: aws.String(arn),
	}

	_, err := tfresource.RetryUntilEqual(ctx, timeout, 0, func() (int, error) {
		return GetRemainingTrustStoreAssociations(ctx, conn, input)
	})

	return err
}

func GetRemainingTrustStoreAssociations(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoreAssociationsInput) (int, error) {
	var output []*elbv2.TrustStoreAssociation

	err := conn.DescribeTrustStoreAssociationsPagesWithContext(ctx, input, func(page *elbv2.DescribeTrustStoreAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrustStoreAssociations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTrustStoreNotFoundException) {
		return -1, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return -1, err
	}

	return len(output), nil
}
