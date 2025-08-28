// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_access_log_subscription", name="Access Log Subscription")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceAccessLogSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessLogSubscriptionCreate,
		ReadWithoutTimeout:   resourceAccessLogSubscriptionRead,
		UpdateWithoutTimeout: resourceAccessLogSubscriptionUpdate,
		DeleteWithoutTimeout: resourceAccessLogSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidARN,
				DiffSuppressFunc: sdkv2.SuppressEquivalentCloudWatchLogsLogGroupARN,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_identifier": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentIDOrARN,
			},
			"service_network_log_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ServiceNetworkLogType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAccessLogSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	input := vpclattice.CreateAccessLogSubscriptionInput{
		ClientToken:        aws.String(sdkid.UniqueId()),
		DestinationArn:     aws.String(d.Get(names.AttrDestinationARN).(string)),
		ResourceIdentifier: aws.String(d.Get("resource_identifier").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("service_network_log_type"); ok {
		input.ServiceNetworkLogType = awstypes.ServiceNetworkLogType(v.(string))
	}

	output, err := conn.CreateAccessLogSubscription(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Access Log Subscription: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceAccessLogSubscriptionRead(ctx, d, meta)...)
}

func resourceAccessLogSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	output, err := findAccessLogSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Access Log Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Access Log Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDestinationARN, output.DestinationArn)
	d.Set(names.AttrResourceARN, output.ResourceArn)
	d.Set("resource_identifier", output.ResourceId)
	d.Set("service_network_log_type", output.ServiceNetworkLogType)

	return diags
}

func resourceAccessLogSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceAccessLogSubscriptionRead(ctx, d, meta)
}

func resourceAccessLogSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Access Log Subscription: %s", d.Id())
	input := vpclattice.DeleteAccessLogSubscriptionInput{
		AccessLogSubscriptionIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteAccessLogSubscription(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Access Log Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func findAccessLogSubscriptionByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetAccessLogSubscriptionOutput, error) {
	input := vpclattice.GetAccessLogSubscriptionInput{
		AccessLogSubscriptionIdentifier: aws.String(id),
	}
	output, err := findAccessLogSubscription(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findAccessLogSubscription(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetAccessLogSubscriptionInput) (*vpclattice.GetAccessLogSubscriptionOutput, error) {
	output, err := conn.GetAccessLogSubscription(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
