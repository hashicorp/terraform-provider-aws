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

// @SDKResource("aws_networkmanager_site", name="Site")
// @Tags(identifierAttribute="arn")
func ResourceSite() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSiteCreate,
		ReadWithoutTimeout:   resourceSiteRead,
		UpdateWithoutTimeout: resourceSiteUpdate,
		DeleteWithoutTimeout: resourceSiteDelete,

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
			names.AttrLocation: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"latitude": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"longitude": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSiteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	input := &networkmanager.CreateSiteInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrLocation); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Location = expandLocation(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Network Manager Site: %s", input)
	output, err := conn.CreateSiteWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Site: %s", err)
	}

	d.SetId(aws.StringValue(output.Site.SiteId))

	if _, err := waitSiteCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSiteRead(ctx, d, meta)...)
}

func resourceSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	site, err := FindSiteByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Site %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Site (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, site.SiteArn)
	d.Set(names.AttrDescription, site.Description)
	d.Set("global_network_id", site.GlobalNetworkId)
	if site.Location != nil {
		if err := d.Set(names.AttrLocation, []interface{}{flattenLocation(site.Location)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting location: %s", err)
		}
	} else {
		d.Set(names.AttrLocation, nil)
	}

	setTagsOut(ctx, site.Tags)

	return diags
}

func resourceSiteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateSiteInput{
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			GlobalNetworkId: aws.String(globalNetworkID),
			SiteId:          aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrLocation); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Location = expandLocation(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating Network Manager Site: %s", input)
		_, err := conn.UpdateSiteWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Site (%s): %s", d.Id(), err)
		}

		if _, err := waitSiteUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSiteRead(ctx, d, meta)...)
}

func resourceSiteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Site: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, siteValidationExceptionTimeout,
		func() (interface{}, error) {
			return conn.DeleteSiteWithContext(ctx, &networkmanager.DeleteSiteInput{
				GlobalNetworkId: aws.String(globalNetworkID),
				SiteId:          aws.String(d.Id()),
			})
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, networkmanager.ErrCodeValidationException, "cannot be deleted due to existing association") {
				return true, err
			}

			return false, err
		},
	)

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Site (%s): %s", d.Id(), err)
	}

	if _, err := waitSiteDeleted(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindSite(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetSitesInput) (*networkmanager.Site, error) {
	output, err := FindSites(ctx, conn, input)

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

func FindSites(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetSitesInput) ([]*networkmanager.Site, error) {
	var output []*networkmanager.Site

	err := conn.GetSitesPagesWithContext(ctx, input, func(page *networkmanager.GetSitesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Sites {
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

func FindSiteByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, siteID string) (*networkmanager.Site, error) {
	input := &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		SiteIds:         aws.StringSlice([]string{siteID}),
	}

	output, err := FindSite(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.SiteId) != siteID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusSiteState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, siteID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSiteByTwoPartKey(ctx, conn, globalNetworkID, siteID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitSiteCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, siteID string, timeout time.Duration) (*networkmanager.Site, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.SiteStatePending},
		Target:  []string{networkmanager.SiteStateAvailable},
		Timeout: timeout,
		Refresh: statusSiteState(ctx, conn, globalNetworkID, siteID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Site); ok {
		return output, err
	}

	return nil, err
}

func waitSiteDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, siteID string, timeout time.Duration) (*networkmanager.Site, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.SiteStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusSiteState(ctx, conn, globalNetworkID, siteID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Site); ok {
		return output, err
	}

	return nil, err
}

func waitSiteUpdated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, siteID string, timeout time.Duration) (*networkmanager.Site, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.SiteStateUpdating},
		Target:  []string{networkmanager.SiteStateAvailable},
		Timeout: timeout,
		Refresh: statusSiteState(ctx, conn, globalNetworkID, siteID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Site); ok {
		return output, err
	}

	return nil, err
}

const (
	siteValidationExceptionTimeout = 2 * time.Minute
)

func expandLocation(tfMap map[string]interface{}) *networkmanager.Location {
	if tfMap == nil {
		return nil
	}

	apiObject := &networkmanager.Location{}

	if v, ok := tfMap[names.AttrAddress].(string); ok {
		apiObject.Address = aws.String(v)
	}

	if v, ok := tfMap["latitude"].(string); ok {
		apiObject.Latitude = aws.String(v)
	}

	if v, ok := tfMap["longitude"].(string); ok {
		apiObject.Longitude = aws.String(v)
	}

	return apiObject
}

func flattenLocation(apiObject *networkmanager.Location) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Address; v != nil {
		tfMap[names.AttrAddress] = aws.StringValue(v)
	}

	if v := apiObject.Latitude; v != nil {
		tfMap["latitude"] = aws.StringValue(v)
	}

	if v := apiObject.Longitude; v != nil {
		tfMap["longitude"] = aws.StringValue(v)
	}

	return tfMap
}
