// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_directory_service_region", name="Region")
// @Tags
func resourceRegion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegionCreate,
		ReadWithoutTimeout:   resourceRegionRead,
		UpdateWithoutTimeout: resourceRegionUpdate,
		DeleteWithoutTimeout: resourceRegionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"desired_number_of_domain_controllers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(2),
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidRegionName,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID := d.Get("directory_id").(string)
	regionName := d.Get("region_name").(string)
	id := regionCreateResourceID(directoryID, regionName)
	input := &directoryservice.AddRegionInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
	}

	if v, ok := d.GetOk("vpc_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VPCSettings = expandDirectoryVpcSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.AddRegion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Region (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitRegionCreated(ctx, conn, directoryID, regionName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Region (%s) create: %s", d.Id(), err)
	}

	optFn := func(o *directoryservice.Options) {
		o.Region = regionName
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		if err := createTags(ctx, conn, directoryID, tags, optFn); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Directory Service Directory (%s) tags: %s", directoryID, err)
		}
	}

	if v, ok := d.GetOk("desired_number_of_domain_controllers"); ok {
		if err := updateNumberOfDomainControllers(ctx, conn, directoryID, v.(int), d.Timeout(schema.TimeoutCreate), optFn); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRegionRead(ctx, d, meta)...)
}

func resourceRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, regionName, err := regionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	region, err := findRegionByTwoPartKey(ctx, conn, directoryID, regionName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Region (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Region (%s): %s", d.Id(), err)
	}

	d.Set("desired_number_of_domain_controllers", region.DesiredNumberOfDomainControllers)
	d.Set("directory_id", region.DirectoryId)
	d.Set("region_name", region.RegionName)
	if region.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettings(region.VpcSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_settings: %s", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	optFn := func(o *directoryservice.Options) {
		o.Region = regionName
	}

	tags, err := listTags(ctx, conn, directoryID, optFn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Directory Service Directory (%s): %s", directoryID, err)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, regionName, err := regionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// The Region must be updated using a client in the region.
	optFn := func(o *directoryservice.Options) {
		o.Region = regionName
	}

	if d.HasChange("desired_number_of_domain_controllers") {
		if err := updateNumberOfDomainControllers(ctx, conn, directoryID, d.Get("desired_number_of_domain_controllers").(int), d.Timeout(schema.TimeoutUpdate), optFn); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)

		if err := updateTags(ctx, conn, directoryID, o, n, optFn); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Directory Service Directory (%s) tags: %s", directoryID, err)
		}
	}

	return append(diags, resourceRegionRead(ctx, d, meta)...)
}

func resourceRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, regionName, err := regionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// The Region must be removed using a client in the region.
	optFn := func(o *directoryservice.Options) {
		o.Region = regionName
	}

	_, err = conn.RemoveRegion(ctx, &directoryservice.RemoveRegionInput{
		DirectoryId: aws.String(directoryID),
	}, optFn)

	if errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Region (%s): %s", d.Id(), err)
	}

	if _, err := waitRegionDeleted(ctx, conn, directoryID, regionName, d.Timeout(schema.TimeoutDelete), optFn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Region (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const regionResourceIDSeparator = "," // nosemgrep:ci.ds-in-const-name,ci.ds-in-var-name

func regionCreateResourceID(directoryID, regionName string) string {
	parts := []string{directoryID, regionName}
	id := strings.Join(parts, regionResourceIDSeparator)

	return id
}

func regionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, regionResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DIRECTORY_ID%[2]sREGION_NAME", id, regionResourceIDSeparator)
}

func findRegion(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeRegionsInput, optFns ...func(*directoryservice.Options)) (*awstypes.RegionDescription, error) {
	output, err := findRegions(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRegions(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeRegionsInput, optFns ...func(*directoryservice.Options)) ([]awstypes.RegionDescription, error) {
	var output []awstypes.RegionDescription

	pages := directoryservice.NewDescribeRegionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RegionsDescription...)
	}

	return output, nil
}

func findRegionByTwoPartKey(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, optFns ...func(*directoryservice.Options)) (*awstypes.RegionDescription, error) {
	input := &directoryservice.DescribeRegionsInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
	}

	output, err := findRegion(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.DirectoryStageDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusRegion(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, optFns ...func(*directoryservice.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRegionByTwoPartKey(ctx, conn, directoryID, regionName, optFns...)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitRegionCreated(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, timeout time.Duration, optFns ...func(*directoryservice.Options)) (*awstypes.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectoryStageRequested, awstypes.DirectoryStageCreating, awstypes.DirectoryStageCreated),
		Target:  enum.Slice(awstypes.DirectoryStageActive),
		Refresh: statusRegion(ctx, conn, directoryID, regionName, optFns...),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDeleted(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, timeout time.Duration, optFns ...func(*directoryservice.Options)) (*awstypes.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectoryStageActive, awstypes.DirectoryStageDeleting),
		Target:  []string{},
		Refresh: statusRegion(ctx, conn, directoryID, regionName, optFns...),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RegionDescription); ok {
		return output, err
	}

	return nil, err
}
