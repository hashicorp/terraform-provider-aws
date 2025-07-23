// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_link", name="Link")
// @Tags(identifierAttribute="arn")
func resourceLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkCreate,
		ReadWithoutTimeout:   resourceLinkRead,
		UpdateWithoutTimeout: resourceLinkUpdate,
		DeleteWithoutTimeout: resourceLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				parsedARN, err := arn.Parse(d.Id())

				if err != nil {
					return nil, fmt.Errorf("parsing ARN (%s): %w", d.Id(), err)
				}

				// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_networkmanager.html#networkmanager-resources-for-iam-policies.
				resourceParts := strings.Split(parsedARN.Resource, "/")

				if actual, expected := len(resourceParts), 3; actual < expected {
					return nil, fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, d.Id(), actual)
				}

				d.SetId(resourceParts[2])
				d.Set("global_network_id", resourceParts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"download_speed": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"upload_speed": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrProviderName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
		},
	}
}

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	input := &networkmanager.CreateLinkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		SiteId:          aws.String(d.Get("site_id").(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("bandwidth"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Bandwidth = expandBandwidth(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrProviderName); ok {
		input.Provider = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Link: %#v", input)
	output, err := conn.CreateLink(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Link: %s", err)
	}

	d.SetId(aws.ToString(output.Link.LinkId))

	if _, err := waitLinkCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	link, err := findLinkByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Link %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Link (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, link.LinkArn)
	if link.Bandwidth != nil {
		if err := d.Set("bandwidth", []any{flattenBandwidth(link.Bandwidth)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting bandwidth: %s", err)
		}
	} else {
		d.Set("bandwidth", nil)
	}
	d.Set(names.AttrDescription, link.Description)
	d.Set("global_network_id", link.GlobalNetworkId)
	d.Set(names.AttrProviderName, link.Provider)
	d.Set("site_id", link.SiteId)
	d.Set(names.AttrType, link.Type)

	setTagsOut(ctx, link.Tags)

	return diags
}

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateLinkInput{
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			GlobalNetworkId: aws.String(globalNetworkID),
			LinkId:          aws.String(d.Id()),
			Provider:        aws.String(d.Get(names.AttrProviderName).(string)),
			Type:            aws.String(d.Get(names.AttrType).(string)),
		}

		if v, ok := d.GetOk("bandwidth"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Bandwidth = expandBandwidth(v.([]any)[0].(map[string]any))
		}

		log.Printf("[DEBUG] Updating Network Manager Link: %#v", input)
		_, err := conn.UpdateLink(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Link (%s): %s", d.Id(), err)
		}

		if _, err := waitLinkUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Link: %s", d.Id())
	_, err := conn.DeleteLink(ctx, &networkmanager.DeleteLinkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(d.Id()),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Link (%s): %s", d.Id(), err)
	}

	if _, err := waitLinkDeleted(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLink(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetLinksInput) (*awstypes.Link, error) {
	output, err := findLinks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func findLinks(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetLinksInput) ([]awstypes.Link, error) {
	var output []awstypes.Link

	pages := networkmanager.NewGetLinksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if globalNetworkIDNotFoundError(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Links...)
	}

	return output, nil
}

func findLinkByTwoPartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID string) (*awstypes.Link, error) {
	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkIds:         []string{linkID},
	}

	output, err := findLink(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.LinkId) != linkID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLinkState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findLinkByTwoPartKey(ctx, conn, globalNetworkID, linkID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitLinkCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID string, timeout time.Duration) (*awstypes.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LinkStatePending),
		Target:  enum.Slice(awstypes.LinkStateAvailable),
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Link); ok {
		return output, err
	}

	return nil, err
}

func waitLinkDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID string, timeout time.Duration) (*awstypes.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LinkStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Link); ok {
		return output, err
	}

	return nil, err
}

func waitLinkUpdated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID string, timeout time.Duration) (*awstypes.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LinkStateUpdating),
		Target:  enum.Slice(awstypes.LinkStateAvailable),
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Link); ok {
		return output, err
	}

	return nil, err
}

func expandBandwidth(tfMap map[string]any) *awstypes.Bandwidth {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Bandwidth{}

	if v, ok := tfMap["download_speed"].(int); ok && v != 0 {
		apiObject.DownloadSpeed = aws.Int32(int32(v))
	}

	if v, ok := tfMap["upload_speed"].(int); ok && v != 0 {
		apiObject.UploadSpeed = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenBandwidth(apiObject *awstypes.Bandwidth) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.DownloadSpeed; v != nil {
		tfMap["download_speed"] = aws.ToInt32(v)
	}

	if v := apiObject.UploadSpeed; v != nil {
		tfMap["upload_speed"] = aws.ToInt32(v)
	}

	return tfMap
}
