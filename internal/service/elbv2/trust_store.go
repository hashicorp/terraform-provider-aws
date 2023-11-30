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
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
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
	input := &elbv2.CreateTrustStoreInput{
		CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
		CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
		Name:                         aws.String(name),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ca_certificates_bundle_s3_object_version"); ok {
		input.CaCertificatesBundleS3ObjectVersion = aws.String(v.(string))
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

	d.SetId(aws.StringValue(output.TrustStores[0].TrustStoreArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
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

	trustStore, err := FindTrustStoreByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Trust Store %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store (%s): %s", d.Id(), err)
	}

	d.Set("arn", trustStore.TrustStoreArn)
	d.Set("name", trustStore.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(trustStore.Name)))

	return diags
}

func resourceTrustStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &elbv2.ModifyTrustStoreInput{
			CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
			CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
			TrustStoreArn:                aws.String(d.Id()),
		}

		if v, ok := d.GetOk("ca_certificates_bundle_s3_object_version"); ok {
			input.CaCertificatesBundleS3ObjectVersion = aws.String(v.(string))
		}

		_, err := conn.ModifyTrustStoreWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Trust Store (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrustStoreRead(ctx, d, meta)...)
}

func resourceTrustStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	if err := waitForNoTrustStoreAssociations(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBV2 Trust Store (%s) associations delete: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting ELBv2 Trust Store: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteTrustStoreWithContext(ctx, &elbv2.DeleteTrustStoreInput{
			TrustStoreArn: aws.String(d.Id()),
		})
	}, elbv2.ErrCodeTrustStoreInUseException, "is currently in use by a listener")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Trust Store (%s): %s", d.Id(), err)
	}

	return diags
}

func FindTrustStoreByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.TrustStore, error) {
	input := &elbv2.DescribeTrustStoresInput{
		TrustStoreArns: aws.StringSlice([]string{arn}),
	}
	output, err := findTrustStore(ctx, conn, input)

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

func findTrustStore(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoresInput) (*elbv2.TrustStore, error) {
	output, err := findTrustStores(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findTrustStores(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoresInput) ([]*elbv2.TrustStore, error) {
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

func findTrustStoreAssociations(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTrustStoreAssociationsInput) ([]*elbv2.TrustStoreAssociation, error) {
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

func waitForNoTrustStoreAssociations(ctx context.Context, conn *elbv2.ELBV2, arn string, timeout time.Duration) error {
	input := &elbv2.DescribeTrustStoreAssociationsInput{
		TrustStoreArn: aws.String(arn),
	}

	_, err := tfresource.RetryUntilEqual(ctx, timeout, 0, func() (int, error) {
		associations, err := findTrustStoreAssociations(ctx, conn, input)

		if err != nil {
			return 0, err
		}

		return len(associations), nil
	})

	return err
}
