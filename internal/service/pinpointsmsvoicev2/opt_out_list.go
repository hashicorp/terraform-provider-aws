// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpointsmsvoicev2_opt_out_list", name="Opt-out List")
// @Tags(identifierAttribute="arn")
func resourceOptOutList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOptOutListCreate,
		ReadWithoutTimeout:   resourceOptOutListRead,
		UpdateWithoutTimeout: resourceOptOutListUpdate,
		DeleteWithoutTimeout: resourceOptOutListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOptOutListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &pinpointsmsvoicev2.CreateOptOutListInput{
		OptOutListName: aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreateOptOutList(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating End User Messaging Opt-out List (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.OptOutListName))

	return append(diags, resourceOptOutListRead(ctx, d, meta)...)
}

func resourceOptOutListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	out, err := findOptOutListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] End User Messaging Opt-out List (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading End User Messaging Opt-out List (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.OptOutListArn)
	d.Set(names.AttrName, out.OptOutListName)

	return diags
}

func resourceOptOutListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceOptOutListRead(ctx, d, meta)...)
}

func resourceOptOutListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	log.Printf("[INFO] Deleting End User Messaging Opt-out List: %s", d.Id())
	_, err := conn.DeleteOptOutList(ctx, &pinpointsmsvoicev2.DeleteOptOutListInput{
		OptOutListName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting End User Messaging Opt-out List (%s): %s", d.Id(), err)
	}

	return diags
}

func findOptOutListByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.OptOutListInformation, error) {
	input := &pinpointsmsvoicev2.DescribeOptOutListsInput{
		OptOutListNames: []string{id},
	}

	return findOptOutList(ctx, conn, input)
}

func findOptOutList(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeOptOutListsInput) (*awstypes.OptOutListInformation, error) {
	output, err := findOptOutLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOptOutLists(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeOptOutListsInput) ([]awstypes.OptOutListInformation, error) {
	var output []awstypes.OptOutListInformation

	pages := pinpointsmsvoicev2.NewDescribeOptOutListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.OptOutLists...)
	}

	return output, nil
}
