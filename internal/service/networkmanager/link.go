// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_link", name="Link")
// @Tags(identifierAttribute="arn")
func ResourceLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkCreate,
		ReadWithoutTimeout:   resourceLinkRead,
		UpdateWithoutTimeout: resourceLinkUpdate,
		DeleteWithoutTimeout: resourceLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

		CustomizeDiff: verify.SetTagsDiff,

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

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	input := &networkmanager.CreateLinkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		SiteId:          aws.String(d.Get("site_id").(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("bandwidth"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Bandwidth = expandBandwidth(v.([]interface{})[0].(map[string]interface{}))
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

	log.Printf("[DEBUG] Creating Network Manager Link: %s", input)
	output, err := conn.CreateLinkWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Link: %s", err)
	}

	d.SetId(aws.StringValue(output.Link.LinkId))

	if _, err := waitLinkCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	link, err := FindLinkByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

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
		if err := d.Set("bandwidth", []interface{}{flattenBandwidth(link.Bandwidth)}); err != nil {
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

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateLinkInput{
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			GlobalNetworkId: aws.String(globalNetworkID),
			LinkId:          aws.String(d.Id()),
			Provider:        aws.String(d.Get(names.AttrProviderName).(string)),
			Type:            aws.String(d.Get(names.AttrType).(string)),
		}

		if v, ok := d.GetOk("bandwidth"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Bandwidth = expandBandwidth(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating Network Manager Link: %s", input)
		_, err := conn.UpdateLinkWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Link (%s): %s", d.Id(), err)
		}

		if _, err := waitLinkUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Link: %s", d.Id())
	_, err := conn.DeleteLinkWithContext(ctx, &networkmanager.DeleteLinkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(d.Id()),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindLink(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetLinksInput) (*networkmanager.Link, error) {
	output, err := FindLinks(ctx, conn, input)

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

func FindLinks(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetLinksInput) ([]*networkmanager.Link, error) {
	var output []*networkmanager.Link

	err := conn.GetLinksPagesWithContext(ctx, input, func(page *networkmanager.GetLinksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Links {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if globalNetworkIDNotFoundError(err) {
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

func FindLinkByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID string) (*networkmanager.Link, error) {
	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkIds:         aws.StringSlice([]string{linkID}),
	}

	output, err := FindLink(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.LinkId) != linkID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLinkState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLinkByTwoPartKey(ctx, conn, globalNetworkID, linkID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitLinkCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID string, timeout time.Duration) (*networkmanager.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.LinkStatePending},
		Target:  []string{networkmanager.LinkStateAvailable},
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Link); ok {
		return output, err
	}

	return nil, err
}

func waitLinkDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID string, timeout time.Duration) (*networkmanager.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.LinkStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Link); ok {
		return output, err
	}

	return nil, err
}

func waitLinkUpdated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID string, timeout time.Duration) (*networkmanager.Link, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.LinkStateUpdating},
		Target:  []string{networkmanager.LinkStateAvailable},
		Timeout: timeout,
		Refresh: statusLinkState(ctx, conn, globalNetworkID, linkID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Link); ok {
		return output, err
	}

	return nil, err
}

func expandBandwidth(tfMap map[string]interface{}) *networkmanager.Bandwidth {
	if tfMap == nil {
		return nil
	}

	apiObject := &networkmanager.Bandwidth{}

	if v, ok := tfMap["download_speed"].(int); ok && v != 0 {
		apiObject.DownloadSpeed = aws.Int64(int64(v))
	}

	if v, ok := tfMap["upload_speed"].(int); ok && v != 0 {
		apiObject.UploadSpeed = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenBandwidth(apiObject *networkmanager.Bandwidth) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DownloadSpeed; v != nil {
		tfMap["download_speed"] = aws.Int64Value(v)
	}

	if v := apiObject.UploadSpeed; v != nil {
		tfMap["upload_speed"] = aws.Int64Value(v)
	}

	return tfMap
}
