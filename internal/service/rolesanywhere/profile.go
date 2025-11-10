// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rolesanywhere_profile", name="Profile")
// @Tags(identifierAttribute="arn")
func resourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"require_instance_properties": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"session_policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &rolesanywhere.CreateProfileInput{
		Name:     aws.String(name),
		RoleArns: flex.ExpandStringValueSet(d.Get("role_arns").(*schema.Set)),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("duration_seconds"); ok {
		input.DurationSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrEnabled); ok {
		input.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok {
		input.ManagedPolicyArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("require_instance_properties"); ok {
		input.RequireInstanceProperties = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("session_policy"); ok {
		input.SessionPolicy = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating RolesAnywhere Profile: %#v", input)
	output, err := conn.CreateProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RolesAnywhere Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Profile.ProfileId))

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	profile, err := findProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RolesAnywhere Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RolesAnywhere Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, profile.ProfileArn)
	d.Set("duration_seconds", profile.DurationSeconds)
	d.Set(names.AttrEnabled, profile.Enabled)
	d.Set("managed_policy_arns", profile.ManagedPolicyArns)
	d.Set(names.AttrName, profile.Name)
	d.Set("require_instance_properties", profile.RequireInstanceProperties)
	d.Set("role_arns", profile.RoleArns)
	d.Set("session_policy", profile.SessionPolicy)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rolesanywhere.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("duration_seconds") {
			input.DurationSeconds = aws.Int32(int32(d.Get("duration_seconds").(int)))
		}

		if d.HasChange("managed_policy_arns") {
			input.ManagedPolicyArns = flex.ExpandStringValueSet(d.Get("managed_policy_arns").(*schema.Set))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("role_arns") {
			input.RoleArns = flex.ExpandStringValueSet(d.Get("role_arns").(*schema.Set))
		}

		if d.HasChange("session_policy") {
			input.SessionPolicy = aws.String(d.Get("session_policy").(string))
		}

		log.Printf("[DEBUG] Updating RolesAnywhere Profile (%s): %#v", d.Id(), input)
		_, err := conn.UpdateProfile(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RolesAnywhere Profile (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrEnabled) {
		_, n := d.GetChange(names.AttrEnabled)
		if n == true {
			err := enableProfile(ctx, d.Id(), meta)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		} else {
			err := disableProfile(ctx, d.Id(), meta)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	log.Printf("[DEBUG] Deleting RolesAnywhere Profile (%s)", d.Id())
	_, err := conn.DeleteProfile(ctx, &rolesanywhere.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RolesAnywhere Profile: (%s): %s", d.Id(), err)
	}

	return diags
}

func findProfileByID(ctx context.Context, conn *rolesanywhere.Client, id string) (*awstypes.ProfileDetail, error) {
	in := &rolesanywhere.GetProfileInput{
		ProfileId: aws.String(id),
	}

	out, err := conn.GetProfile(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Profile == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Profile, nil
}

func disableProfile(ctx context.Context, profileId string, meta any) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.DisableProfileInput{
		ProfileId: aws.String(profileId),
	}

	_, err := conn.DisableProfile(ctx, input)
	return err
}

func enableProfile(ctx context.Context, profileId string, meta any) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.EnableProfileInput{
		ProfileId: aws.String(profileId),
	}

	_, err := conn.EnableProfile(ctx, input)
	return err
}
