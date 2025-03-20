// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_security_profile", name="Security Profile")
// @Tags(identifierAttribute="arn")
func resourceSecurityProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityProfileCreate,
		ReadWithoutTimeout:   resourceSecurityProfileRead,
		UpdateWithoutTimeout: resourceSecurityProfileUpdate,
		DeleteWithoutTimeout: resourceSecurityProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPermissions: {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 500,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
			"security_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSecurityProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	securityProfileName := d.Get(names.AttrName).(string)
	input := &connect.CreateSecurityProfileInput{
		InstanceId:          aws.String(instanceID),
		SecurityProfileName: aws.String(securityProfileName),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPermissions); ok && v.(*schema.Set).Len() > 0 {
		input.Permissions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateSecurityProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Security Profile (%s): %s", securityProfileName, err)
	}

	id := securityProfileCreateResourceID(instanceID, aws.ToString(output.SecurityProfileId))
	d.SetId(id)

	return append(diags, resourceSecurityProfileRead(ctx, d, meta)...)
}

func resourceSecurityProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, securityProfileID, err := securityProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	securityProfile, err := findSecurityProfileByTwoPartKey(ctx, conn, instanceID, securityProfileID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Security Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Security Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, securityProfile.Arn)
	d.Set(names.AttrDescription, securityProfile.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, securityProfile.SecurityProfileName)
	d.Set("organization_resource_id", securityProfile.OrganizationResourceId)
	d.Set("security_profile_id", securityProfile.Id)

	permissions, err := findSecurityProfilePermissionsByTwoPartKey(ctx, conn, instanceID, securityProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Security Profile (%s) permissions: %s", d.Id(), err)
	}

	d.Set(names.AttrPermissions, permissions)

	setTagsOut(ctx, securityProfile.Tags)

	return diags
}

func resourceSecurityProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, securityProfileID, err := securityProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &connect.UpdateSecurityProfileInput{
			InstanceId:        aws.String(instanceID),
			SecurityProfileId: aws.String(securityProfileID),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrPermissions) {
			input.Permissions = flex.ExpandStringValueSet(d.Get(names.AttrPermissions).(*schema.Set))
		}

		_, err = conn.UpdateSecurityProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Security Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSecurityProfileRead(ctx, d, meta)...)
}

func resourceSecurityProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, securityProfileID, err := securityProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Security Profile: %s", d.Id())
	input := connect.DeleteSecurityProfileInput{
		InstanceId:        aws.String(instanceID),
		SecurityProfileId: aws.String(securityProfileID),
	}
	_, err = conn.DeleteSecurityProfile(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Security Profile (%s): %s", d.Id(), err)
	}

	return diags
}

const securityProfileResourceIDSeparator = ":"

func securityProfileCreateResourceID(instanceID, securityProfileID string) string {
	parts := []string{instanceID, securityProfileID}
	id := strings.Join(parts, securityProfileResourceIDSeparator)

	return id
}

func securityProfileParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, securityProfileResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]ssecurityProfileID", id, securityProfileResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findSecurityProfileByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, securityProfileID string) (*awstypes.SecurityProfile, error) {
	input := &connect.DescribeSecurityProfileInput{
		InstanceId:        aws.String(instanceID),
		SecurityProfileId: aws.String(securityProfileID),
	}

	return findSecurityProfile(ctx, conn, input)
}

func findSecurityProfile(ctx context.Context, conn *connect.Client, input *connect.DescribeSecurityProfileInput) (*awstypes.SecurityProfile, error) {
	output, err := conn.DescribeSecurityProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SecurityProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SecurityProfile, nil
}

func findSecurityProfilePermissionsByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, securityProfileID string) ([]string, error) {
	const maxResults = 60
	input := &connect.ListSecurityProfilePermissionsInput{
		InstanceId:        aws.String(instanceID),
		MaxResults:        aws.Int32(maxResults),
		SecurityProfileId: aws.String(securityProfileID),
	}

	return findSecurityProfilePermissions(ctx, conn, input)
}

func findSecurityProfilePermissions(ctx context.Context, conn *connect.Client, input *connect.ListSecurityProfilePermissionsInput) ([]string, error) {
	var output []string

	pages := connect.NewListSecurityProfilePermissionsPaginator(conn, input)
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

		output = append(output, page.Permissions...)
	}

	return output, nil
}
