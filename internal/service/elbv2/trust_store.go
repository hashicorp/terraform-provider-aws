// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lb_trust_store", name="Trust Store")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types;types.TrustStore")
// @Testing(importIgnore="ca_certificates_bundle_s3_bucket;ca_certificates_bundle_s3_key")
func resourceTrustStore() *schema.Resource {
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
			names.AttrARN: {
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
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validNamePrefix,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrustStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)
	partition := meta.(*conns.AWSClient).Partition

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &elasticloadbalancingv2.CreateTrustStoreInput{
		CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
		CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
		Name:                         aws.String(name),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ca_certificates_bundle_s3_object_version"); ok {
		input.CaCertificatesBundleS3ObjectVersion = aws.String(v.(string))
	}

	output, err := conn.CreateTrustStore(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateTrustStore(ctx, input)
	}

	// Tags are not supported on creation with some protocol types(i.e. GENEVE)
	// Retry creation without tags
	if input.Tags != nil && tfawserr.ErrMessageContains(err, errCodeValidationError, tagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = conn.CreateTrustStore(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Trust Store (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TrustStores[0].TrustStoreArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findTrustStoreByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Trust Store (%s) create: %s", d.Id(), err)
	}

	if _, err := waitTrustStoreActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Trust Store (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	trustStore, err := findTrustStoreByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Trust Store %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Trust Store (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, trustStore.TrustStoreArn)
	d.Set(names.AttrName, trustStore.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(trustStore.Name)))

	return diags
}

func resourceTrustStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &elasticloadbalancingv2.ModifyTrustStoreInput{
			CaCertificatesBundleS3Bucket: aws.String(d.Get("ca_certificates_bundle_s3_bucket").(string)),
			CaCertificatesBundleS3Key:    aws.String(d.Get("ca_certificates_bundle_s3_key").(string)),
			TrustStoreArn:                aws.String(d.Id()),
		}

		if v, ok := d.GetOk("ca_certificates_bundle_s3_object_version"); ok {
			input.CaCertificatesBundleS3ObjectVersion = aws.String(v.(string))
		}

		_, err := conn.ModifyTrustStore(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Trust Store (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrustStoreRead(ctx, d, meta)...)
}

func resourceTrustStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	if err := waitForNoTrustStoreAssociations(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBV2 Trust Store (%s) associations delete: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting ELBv2 Trust Store: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.TrustStoreInUseException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteTrustStore(ctx, &elasticloadbalancingv2.DeleteTrustStoreInput{
			TrustStoreArn: aws.String(d.Id()),
		})
	}, "is currently in use by a listener")

	if errs.IsA[*awstypes.TrustStoreNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Trust Store (%s): %s", d.Id(), err)
	}

	return diags
}

func findTrustStoreByARN(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (*awstypes.TrustStore, error) {
	input := &elasticloadbalancingv2.DescribeTrustStoresInput{
		TrustStoreArns: []string{arn},
	}
	output, err := findTrustStore(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TrustStoreArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTrustStore(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTrustStoresInput) (*awstypes.TrustStore, error) {
	output, err := findTrustStores(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrustStores(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTrustStoresInput) ([]awstypes.TrustStore, error) {
	var output []awstypes.TrustStore

	pages := elasticloadbalancingv2.NewDescribeTrustStoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.TrustStoreNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrustStores...)
	}

	return output, nil
}

func statusTrustStore(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTrustStoreByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTrustStoreActive(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string, timeout time.Duration) (*awstypes.TrustStore, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.TrustStoreStatusCreating),
		Target:     enum.Slice(awstypes.TrustStoreStatusActive),
		Refresh:    statusTrustStore(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrustStore); ok {
		return output, err
	}

	return nil, err
}

func findTrustStoreAssociations(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTrustStoreAssociationsInput) ([]awstypes.TrustStoreAssociation, error) {
	var output []awstypes.TrustStoreAssociation

	pages := elasticloadbalancingv2.NewDescribeTrustStoreAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.TrustStoreNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrustStoreAssociations...)
	}

	return output, nil
}

func waitForNoTrustStoreAssociations(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string, timeout time.Duration) error {
	input := &elasticloadbalancingv2.DescribeTrustStoreAssociationsInput{
		TrustStoreArn: aws.String(arn),
	}

	_, err := tfresource.RetryUntilEqual(ctx, timeout, 0, func() (int, error) {
		associations, err := findTrustStoreAssociations(ctx, conn, input)

		if tfresource.NotFound(err) {
			return 0, nil
		}

		if err != nil {
			return 0, err
		}

		return len(associations), nil
	})

	return err
}
