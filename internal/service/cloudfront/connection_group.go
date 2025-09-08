// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	ret "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_connection_group", name="Connection Group")
// @Tags(identifierAttribute="arn")
func resourceConnectionGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionGroupCreate,
		ReadWithoutTimeout:   resourceConnectionGroupRead,
		UpdateWithoutTimeout: resourceConnectionGroupUpdate,
		DeleteWithoutTimeout: resourceConnectionGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"anycast_ip_list_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"ipv6_enabled"},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_enabled": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"anycast_ip_list_id"},
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait_for_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := cloudfront.CreateConnectionGroupInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("anycast_ip_list_id"); ok {
		input.AnycastIpListId = aws.String(v.(string))
	}

	if d.GetRawConfig().GetAttr("ipv6_enabled").IsKnown() {
		input.Ipv6Enabled = aws.Bool(d.Get("ipv6_enabled").(bool))
	}
	input.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{Items: tags}
	}

	output, err := conn.CreateConnectionGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Connection Group (%s): %s", *input.Name, err)
	}

	d.SetId(aws.ToString(output.ConnectionGroup.Id))

	if d.Get("wait_for_deployment").(bool) {
		if _, err := waitConnectionGroupDeployed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Connection Group (%s) deploy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectionGroupRead(ctx, d, meta)...)
}

func resourceConnectionGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findConnectionGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && ret.NotFound(err) {
		log.Printf("[WARN] CloudFront Connection Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Connection Group (%s): %s", d.Id(), err)
	}

	connectionGroup := output.ConnectionGroup
	d.Set("anycast_ip_list_id", connectionGroup.AnycastIpListId)
	d.Set(names.AttrARN, connectionGroup.Arn)
	d.Set(names.AttrEnabled, connectionGroup.Enabled)
	d.Set("etag", output.ETag)
	d.Set("ipv6_enabled", connectionGroup.Ipv6Enabled)
	d.Set("is_default", connectionGroup.IsDefault)
	d.Set("last_modified_time", connectionGroup.LastModifiedTime.String())
	d.Set(names.AttrName, connectionGroup.Name)
	d.Set("routing_endpoint", connectionGroup.RoutingEndpoint)
	d.Set(names.AttrStatus, connectionGroup.Status)

	return diags
}

func resourceConnectionGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := cloudfront.UpdateConnectionGroupInput{
			Id:      aws.String(d.Id()),
			IfMatch: aws.String(d.Get("etag").(string)),
		}

		if v, ok := d.GetOk("anycast_ip_list_id"); ok {
			input.AnycastIpListId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("ipv6_enabled"); ok {
			input.Ipv6Enabled = aws.Bool(v.(bool))
		} else if d.HasChange("ipv6_enabled") {
			input.Ipv6Enabled = aws.Bool(d.Get("ipv6_enabled").(bool))
		}

		input.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

		_, err := conn.UpdateConnectionGroup(ctx, &input)

		// Refresh our ETag if it is out of date and attempt update again.
		if errs.IsA[*awstypes.PreconditionFailed](err) {
			var etag string
			etag, err = connectionGroupETag(ctx, conn, d.Id())

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input.IfMatch = aws.String(etag)

			_, err = conn.UpdateConnectionGroup(ctx, &input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFront Connection Group: %s", err)
		}

		if d.Get("wait_for_deployment").(bool) {
			if _, err := waitConnectionGroupDeployed(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Connection Group (%s) deploy: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceConnectionGroupRead(ctx, d, meta)...)
}

func resourceConnectionGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.Get(names.AttrARN).(string) == "" {
		diags = append(diags, resourceConnectionGroupRead(ctx, d, meta)...)
	}

	if err := disableConnectionGroup(ctx, conn, d.Id()); err != nil {
		if ret.NotFound(err) {
			return diags
		}
		return sdkdiag.AppendFromErr(diags, err)
	}

	err := deleteConnectionGroup(ctx, conn, d.Id())

	if err == nil || ret.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
		return diags
	}
	// Disable connection group if it is not yet disabled and attempt deletion again.
	// Here we update via the deployed configuration to ensure we are not submitting an out of date
	// configuration from the Terraform configuration, should other changes have occurred manually.
	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableConnectionGroup(ctx, conn, d.Id()); err != nil {
			if ret.NotFound(err) {
				return diags
			}

			return sdkdiag.AppendFromErr(diags, err)
		}

		const (
			timeout = 30 * time.Second
		)
		_, err = tfresource.RetryWhenIsA[any, *awstypes.ResourceNotDisabled](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteConnectionGroup(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		const (
			timeout = 10 * time.Second
		)
		_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteConnectionGroup(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableConnectionGroup(ctx, conn, d.Id()); err != nil {
			if ret.NotFound(err) {
				return diags
			}

			return sdkdiag.AppendFromErr(diags, err)
		}

		err = deleteConnectionGroup(ctx, conn, d.Id())
	}

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func findConnectionGroupByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	input := &cloudfront.GetConnectionGroupInput{
		Identifier: aws.String(id),
	}

	output, err := conn.GetConnectionGroup(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectionGroup == nil || output.ConnectionGroup.Name == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func disableConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findConnectionGroupByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Connection Group (%s): %w", id, err)
	}

	if aws.ToString(output.ConnectionGroup.Status) == connectionGroupStatusInProgress {
		output, err = waitConnectionGroupDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Connection Group (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.ConnectionGroup.Enabled) {
		return nil
	}

	input := cloudfront.UpdateConnectionGroupInput{
		Id:      aws.String(id),
		IfMatch: output.ETag,
	}

	input.Enabled = aws.Bool(false)

	_, err = conn.UpdateConnectionGroup(ctx, &input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Connection Group (%s): %w", id, err)
	}

	if _, err := waitConnectionGroupDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Connection Group (%s) deploy: %w", id, err)
	}

	return nil
}

func deleteConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := connectionGroupETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := cloudfront.DeleteConnectionGroupInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteConnectionGroup(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Connection Group (%s): %w", id, err)
	}

	if _, err := waitConnectionGroupDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Connection Group (%s) delete: %w", id, err)
	}

	return nil
}

func waitConnectionGroupDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{connectionGroupStatusInProgress},
		Target:     []string{connectionGroupStatusDeployed},
		Refresh:    statusConnectionGroup(ctx, conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetConnectionGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionGroupDeleted(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{connectionGroupStatusInProgress, connectionGroupStatusDeployed},
		Target:     []string{},
		Refresh:    statusConnectionGroup(ctx, conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetConnectionGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func statusConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectionGroupByID(ctx, conn, id)

		if ret.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.ConnectionGroup.Status), nil
	}
}

func connectionGroupETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findConnectionGroupByID(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Connection Group (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}
