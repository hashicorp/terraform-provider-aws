package eks

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			},
			"service_account_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)
	id := AddonCreateResourceID(clusterName, addonName)

	input := &eks.CreateAddonInput{
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(resource.UniqueId()),
		ClusterName:        aws.String(clusterName),
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.AddonVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_account_role_arn"); ok {
		input.ServiceAccountRoleArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateAddonWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "CREATE_FAILED") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidParameterException, "does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateAddonWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating EKS Add-On (%s): %w", id, err))
	}

	d.SetId(id)

	_, err = waitAddonCreated(ctx, conn, clusterName, addonName)

	if err != nil {
		// Creating addon w/o setting resolve_conflicts to "OVERWRITE"
		// might result in a failed creation, if unmanaged version of addon is already deployed
		// and there are configuration conflicts:
		// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
		//
		// Addon resource is tainted after failed creation, thus will be deleted and created again.
		// Re-creating like this will resolve the error, but it will also purge any
		// configurations that were applied by the user (that were conflicting). This might we an unwanted
		// side effect and should be left for the user to decide how to handle it.
		return diag.FromErr(fmt.Errorf("unexpected EKS Add-On (%s) state returned during creation: %w\n[WARNING] Running terraform apply again will remove the kubernetes add-on and attempt to create it again effectively purging previous add-on configuration",
			d.Id(), err))
	}

	return resourceAddonRead(ctx, d, meta)
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	addon, err := FindAddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Add-On (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading EKS Add-On (%s): %w", d.Id(), err))
	}

	d.Set("addon_name", addon.AddonName)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("arn", addon.AddonArn)
	d.Set("cluster_name", addon.ClusterName)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).Format(time.RFC3339))
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)

	tags := KeyValueTags(addon.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("addon_version", "service_account_role_arn") {
		input := &eks.UpdateAddonInput{
			AddonName:          aws.String(addonName),
			ClientRequestToken: aws.String(resource.UniqueId()),
			ClusterName:        aws.String(clusterName),
		}

		if d.HasChange("addon_version") {
			input.AddonVersion = aws.String(d.Get("addon_version").(string))
		}

		if v, ok := d.GetOk("resolve_conflicts"); ok {
			input.ResolveConflicts = aws.String(v.(string))
		}

		// If service account role ARN is already provided, use it. Otherwise, the add-on uses
		// permissions assigned to the node IAM role.
		if d.HasChange("service_account_role_arn") || d.Get("service_account_role_arn").(string) != "" {
			input.ServiceAccountRoleArn = aws.String(d.Get("service_account_role_arn").(string))
		}

		output, err := conn.UpdateAddonWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating EKS Add-On (%s): %w", d.Id(), err))
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waitAddonUpdateSuccessful(ctx, conn, clusterName, addonName, updateID)

		if err != nil {
			if d.Get("resolve_conflicts") != eks.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return diag.FromErr(fmt.Errorf("error waiting for EKS Add-On (%s) update (%s): %w, consider setting attribute %q to %q",
					d.Id(), updateID, err, "resolve_conflicts", eks.ResolveConflictsOverwrite))
			}

			return diag.FromErr(fmt.Errorf("error waiting for EKS Add-On (%s) update (%s): %w", d.Id(), updateID, err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceAddonRead(ctx, d, meta)
}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, addonName, err := AddonParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
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
		return diag.FromErr(fmt.Errorf("error deleting EKS Add-On (%s): %w", d.Id(), err))
	}

	_, err = waitAddonDeleted(ctx, conn, clusterName, addonName)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for EKS Add-On (%s) to delete: %w", d.Id(), err))
	}

	return nil
}
