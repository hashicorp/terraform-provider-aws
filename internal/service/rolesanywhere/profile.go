// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rolesanywhere_profile", name="Profile")
// @Tags(identifierAttribute="arn")
func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"enabled": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"require_instance_properties": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arns": {
				Type:     schema.TypeSet,
				Required: true,
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
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	name := d.Get("name").(string)
	input := &rolesanywhere.CreateProfileInput{
		Name:     aws.String(name),
		RoleArns: expandStringList(d.Get("role_arns").(*schema.Set).List()),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("duration_seconds"); ok {
		input.DurationSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("enabled"); ok {
		input.Enabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("managed_policy_arns"); ok {
		input.ManagedPolicyArns = expandStringList(v.(*schema.Set).List())
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
		return diag.Errorf("creating RolesAnywhere Profile (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Profile.ProfileId))

	return resourceProfileRead(ctx, d, meta)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	profile, err := FindProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RolesAnywhere Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading RolesAnywhere Profile (%s): %s", d.Id(), err)
	}

	d.Set("arn", profile.ProfileArn)
	d.Set("duration_seconds", profile.DurationSeconds)
	d.Set("enabled", profile.Enabled)
	d.Set("managed_policy_arns", profile.ManagedPolicyArns)
	d.Set("name", profile.Name)
	d.Set("require_instance_properties", profile.RequireInstanceProperties)
	d.Set("role_arns", profile.RoleArns)
	d.Set("session_policy", profile.SessionPolicy)

	return nil
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	if d.HasChangesExcept("enabled", "tags_all") {
		input := &rolesanywhere.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("duration_seconds") {
			input.DurationSeconds = aws.Int32(int32(d.Get("duration_seconds").(int)))
		}

		if d.HasChange("managed_policy_arns") {
			input.ManagedPolicyArns = expandStringList(d.Get("managed_policy_arns").(*schema.Set).List())
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arns") {
			input.RoleArns = expandStringList(d.Get("role_arns").(*schema.Set).List())
		}

		if d.HasChange("session_policy") {
			input.SessionPolicy = aws.String(d.Get("session_policy").(string))
		}

		log.Printf("[DEBUG] Updating RolesAnywhere Profile (%s): %#v", d.Id(), input)
		_, err := conn.UpdateProfile(ctx, input)
		if err != nil {
			return diag.Errorf("updating RolesAnywhere Profile (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("enabled") {
		_, n := d.GetChange("enabled")
		if n == true {
			err := enableProfile(ctx, d.Id(), meta)
			if err != nil {
				diag.Errorf("enabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		} else {
			err := disableProfile(ctx, d.Id(), meta)
			if err != nil {
				diag.Errorf("disabling RolesAnywhere Profile (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceProfileRead(ctx, d, meta)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	log.Printf("[DEBUG] Deleting RolesAnywhere Profile (%s)", d.Id())
	_, err := conn.DeleteProfile(ctx, &rolesanywhere.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting RolesAnywhere Profile: (%s): %s", d.Id(), err)
	}

	return nil
}

func disableProfile(ctx context.Context, profileId string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.DisableProfileInput{
		ProfileId: aws.String(profileId),
	}

	_, err := conn.DisableProfile(ctx, input)
	return err
}

func enableProfile(ctx context.Context, profileId string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.EnableProfileInput{
		ProfileId: aws.String(profileId),
	}

	_, err := conn.EnableProfile(ctx, input)
	return err
}
