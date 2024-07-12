// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_access_log_subscription", name="Access Log Subscription")
// @Tags(identifierAttribute="arn")
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
				DiffSuppressFunc: suppressEquivalentCloudWatchLogsLogGroupARN,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameAccessLogSubscription = "Access Log Subscription"
)

func resourceAccessLogSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateAccessLogSubscriptionInput{
		ClientToken:        aws.String(id.UniqueId()),
		DestinationArn:     aws.String(d.Get(names.AttrDestinationARN).(string)),
		ResourceIdentifier: aws.String(d.Get("resource_identifier").(string)),
		Tags:               getTagsIn(ctx),
	}

	out, err := conn.CreateAccessLogSubscription(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameAccessLogSubscription, d.Get(names.AttrDestinationARN).(string), err)
	}

	d.SetId(aws.ToString(out.Id))

	return append(diags, resourceAccessLogSubscriptionRead(ctx, d, meta)...)
}

func resourceAccessLogSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findAccessLogSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice AccessLogSubscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameAccessLogSubscription, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrDestinationARN, out.DestinationArn)
	d.Set(names.AttrResourceARN, out.ResourceArn)
	d.Set("resource_identifier", out.ResourceId)

	return diags
}

func resourceAccessLogSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceAccessLogSubscriptionRead(ctx, d, meta)
}

func resourceAccessLogSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice AccessLogSubscription %s", d.Id())
	_, err := conn.DeleteAccessLogSubscription(ctx, &vpclattice.DeleteAccessLogSubscriptionInput{
		AccessLogSubscriptionIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameAccessLogSubscription, d.Id(), err)
	}

	return diags
}

func findAccessLogSubscriptionByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetAccessLogSubscriptionOutput, error) {
	in := &vpclattice.GetAccessLogSubscriptionInput{
		AccessLogSubscriptionIdentifier: aws.String(id),
	}
	out, err := conn.GetAccessLogSubscription(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// suppressEquivalentCloudWatchLogsLogGroupARN provides custom difference suppression
// for strings that represent equal CloudWatch Logs log group ARNs.
func suppressEquivalentCloudWatchLogsLogGroupARN(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSuffix(old, ":*") == strings.TrimSuffix(new, ":*")
}
