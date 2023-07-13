// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_addon", name="Add-On")
// @Tags(identifierAttribute="arn")
func ResourceAddon() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAddonCreate,
		ReadWithoutTimeout:   resourceAddonRead,
		UpdateWithoutTimeout: resourceAddonUpdate,
		DeleteWithoutTimeout: resourceAddonDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					// Regular expression taken from: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
					validation.StringMatch(regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`), "must follow semantic version format"),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"configuration_values": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preserve": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"resolve_conflicts": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(eks.ResolveConflicts_Values(), false),
				Deprecated:   `The "resolve_conflicts" attribute can't be set to "PRESERVE" on initial resource creation. Use "resolve_conflicts_on_create" and/or "resolve_conflicts_on_update" instead`,
			},
			"resolve_conflicts_on_create": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					eks.ResolveConflictsNone,
					eks.ResolveConflictsOverwrite,
				}, false),
				ConflictsWith: []string{"resolve_conflicts"},
			},
			"resolve_conflicts_on_update": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(eks.ResolveConflicts_Values(), false),
				ConflictsWith: []string{"resolve_conflicts"},
			},
			"service_account_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)
	id := AddonCreateResourceID(clusterName, addonName)
	input := &eks.CreateAddonInput{
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		ClusterName:        aws.String(clusterName),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.AddonVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration_values"); ok {
		input.ConfigurationValues = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	} else if v, ok := d.GetOk("resolve_conflicts_on_create"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_account_role_arn"); ok {
		input.ServiceAccountRoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateAddonWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "CREATE_FAILED") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "does not exist") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Add-On (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitAddonCreated(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutCreate)); err != nil {
		// Creating addon w/o setting resolve_conflicts to "OVERWRITE"
		// might result in a failed creation, if unmanaged version of addon is already deployed
		// and there are configuration conflicts:
		// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
		//
		// Addon resource is tainted after failed creation, thus will be deleted and created again.
		// Re-creating like this will resolve the error, but it will also purge any
		// configurations that were applied by the user (that were conflicting). This might we an unwanted
		// side effect and should be left for the user to decide how to handle it.
		diags = sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) create: %s", d.Id(), err)
		return sdkdiag.AppendWarningf(diags, "Running terraform apply again will remove the kubernetes add-on and attempt to create it again effectively purging previous add-on configuration")
	}

	return append(diags, resourceAddonRead(ctx, d, meta)...)
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	addon, err := FindAddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Add-On (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Add-On (%s): %s", d.Id(), err)
	}

	d.Set("addon_name", addon.AddonName)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("arn", addon.AddonArn)
	d.Set("cluster_name", addon.ClusterName)
	d.Set("configuration_values", addon.ConfigurationValues)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).Format(time.RFC3339))
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)

	setTagsOut(ctx, addon.Tags)

	return diags
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("addon_version", "service_account_role_arn", "configuration_values") {
		input := &eks.UpdateAddonInput{
			AddonName:          aws.String(addonName),
			ClientRequestToken: aws.String(sdkid.UniqueId()),
			ClusterName:        aws.String(clusterName),
		}

		if d.HasChange("addon_version") {
			input.AddonVersion = aws.String(d.Get("addon_version").(string))
		}

		if d.HasChange("configuration_values") {
			input.ConfigurationValues = aws.String(d.Get("configuration_values").(string))
		}

		var conflictResolutionAttr, conflictResolution string

		if v, ok := d.GetOk("resolve_conflicts"); ok {
			conflictResolutionAttr = "resolve_conflicts"
			conflictResolution = v.(string)
			input.ResolveConflicts = aws.String(v.(string))
		} else if v, ok := d.GetOk("resolve_conflicts_on_update"); ok {
			conflictResolutionAttr = "resolve_conflicts_on_update"
			conflictResolution = v.(string)
			input.ResolveConflicts = aws.String(v.(string))
		}

		// If service account role ARN is already provided, use it. Otherwise, the add-on uses
		// permissions assigned to the node IAM role.
		if d.HasChange("service_account_role_arn") || d.Get("service_account_role_arn").(string) != "" {
			input.ServiceAccountRoleArn = aws.String(d.Get("service_account_role_arn").(string))
		}

		output, err := conn.UpdateAddonWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EKS Add-On (%s): %s", d.Id(), err)
		}

		updateID := aws.StringValue(output.Update.Id)
		if _, err := waitAddonUpdateSuccessful(ctx, conn, clusterName, addonName, updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			if conflictResolution != eks.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) update (%s): %s. Consider setting attribute %q to %q", d.Id(), updateID, err, conflictResolutionAttr, eks.ResolveConflictsOverwrite)
			}

			return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) update (%s): %s", d.Id(), updateID, err)
		}
	}

	return append(diags, resourceAddonRead(ctx, d, meta)...)
}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &eks.DeleteAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	if v, ok := d.GetOk("preserve"); ok {
		input.Preserve = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Deleting EKS Add-On: %s", d.Id())
	_, err = conn.DeleteAddonWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Add-On (%s): %s", d.Id(), err)
	}

	if _, err := waitAddonDeleted(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) delete: %s", d.Id(), err)
	}

	return diags
}
