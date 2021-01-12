package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEksAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsEksAddonCreate,
		ReadContext:   resourceAwsEksAddonRead,
		UpdateContext: resourceAwsEksAddonUpdate,
		DeleteContext: resourceAwsEksAddonDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"service_account_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"resolve_conflicts": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(eks.ResolveConflicts_Values(), false),
			},
			"status": {
				Type:     schema.TypeString,
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
			"tags": tagsSchemaComputed(),
		},
	}
}

func resourceAwsEksAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName := d.Get("cluster_name").(string)
	addonName := d.Get("addon_name").(string)

	input := &eks.CreateAddonInput{
		ClusterName:        aws.String(clusterName),
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.AddonVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_account_role_arn"); ok {
		input.ServiceAccountRoleArn = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	}

	log.Printf("[DEBUG] Creating EKS Addon %s", addonName)

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateAddonWithContext(ctx, input)
		if err != nil {
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "CREATE_FAILED") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "does not exist") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateAddonWithContext(ctx, input)
	}
	if err != nil {
		return diag.Errorf("error creating EKS Addon (%s): %s", addonName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", clusterName, addonName))

	// TODO: refactor and move all waiters and status getters to waiter package for easier maintenance
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating},
		Target:  []string{eks.AddonStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksAddonStatusContext(ctx, conn, addonName, clusterName),
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		if d.Get("resolve_conflicts") == "" {
			// Creating addon w/o setting resolve_conflicts to "OVERWRITE"
			// might result in a failed creation, if unmanaged version of addon is already deployed
			// and there are configuration conflicts:
			// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
			//
			// Removing the failed to create Addon and re-applying also resolves this this issue.
			// Handling this Error and recreating the Addon might result in unwanted side-effect
			// of loosing unmanaged VPC-CNI plugin configuration. Thus it should be left to
			return diag.Errorf("unexpected EKS Addon (%s) state returned during creation: %s, [WARNING] running terraform apply again will remove the EKS Addon and attempt to create it again",
				d.Id(), err)
		}
		return diag.FromErr(err)
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.DescribeAddonInput{
		ClusterName: aws.String(clusterName),
		AddonName:   aws.String(addonName),
	}

	log.Printf("[DEBUG] Reading EKS Addon: %s", d.Id())
	output, err := conn.DescribeAddonWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] EKS Addon (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading EKS Addon (%s): %s", d.Id(), err)
	}

	addon := output.Addon
	if addon == nil {
		log.Printf("[WARN] EKS Addon (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cluster_name", addon.ClusterName)
	d.Set("addon_name", addon.AddonName)
	d.Set("arn", addon.AddonArn)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)
	d.Set("status", addon.Status)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).Format(time.RFC3339))

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(addon.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags attribute: %s", err)
	}

	return nil
}

func resourceAwsEksAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.UpdateAddonInput{
		ClusterName:        aws.String(clusterName),
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if d.HasChange("resolve_conflicts") {
		if v, ok := d.GetOk("resolve_conflicts"); ok {
			d.Set("resolve_conflicts", aws.String(v.(string)))
		}
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	// TODO: addon_version and service_account_role_arn attribute changes can be done in one API call
	// refactor to reduce amount of API calls
	if d.HasChange("addon_version") {
		input.AddonVersion = aws.String(d.Get("addon_version").(string))
		log.Printf("[DEBUG] Updating EKS Addon (%s) version: %s", d.Id(), input)
		output, err := conn.UpdateAddonWithContext(ctx, input)
		if err != nil {
			return diag.Errorf("error updating EKS Addon (%s) version: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return diag.Errorf("error determining EKS Addon (%s) version update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForUpdateEksAddonContext(ctx, conn, clusterName, addonName, updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			if d.Get("resolve_conflicts") != eks.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return diag.Errorf("error waiting for EKS Addon (%s) update (%s): %s, try setting attribute %q to %q",
					d.Id(), updateID, err, "resolve_conflicts", eks.ResolveConflictsOverwrite)
			}
			return diag.Errorf("error waiting for EKS Addon (%s) update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChange("service_account_role_arn") {
		input.ServiceAccountRoleArn = aws.String(d.Get("service_account_role_arn").(string))
		log.Printf("[DEBUG] Updating EKS Addon (%s) service account role: %s", d.Id(), input)
		output, err := conn.UpdateAddonWithContext(ctx, input)
		if err != nil {
			return diag.Errorf("error updating EKS Addon (%s) version: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return diag.Errorf("error determining EKS Addon (%s) version update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForUpdateEksAddonContext(ctx, conn, clusterName, addonName, updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			if d.Get("resolve_conflicts") != eks.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return diag.Errorf("error waiting for EKS Addon (%s) update (%s): %s, try setting attribute %q to %q",
					d.Id(), updateID, err, "resolve_conflicts", eks.ResolveConflictsOverwrite)
			}
			return diag.Errorf("error waiting for EKS Addon (%s) update (%s): %s", d.Id(), updateID, err)
		}
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.DeleteAddonInput{
		ClusterName: aws.String(clusterName),
		AddonName:   aws.String(addonName),
	}

	log.Printf("[DEBUG] Deleting EKS Cluster Addon: %s", d.Id())
	_, err = conn.DeleteAddonWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error deleting EKS Addon (%s): %s", d.Id(), err)
	}

	err = waitForDeleteEksAddonDeleteContext(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.Errorf("error waiting for EKS Addon (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func waitForDeleteEksAddonDeleteContext(ctx context.Context, conn *eks.EKS, clusterName, addonName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.AddonStatusActive,
			eks.AddonStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshEksAddonStatusContext(ctx, conn, addonName, clusterName),
	}
	addon, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// EKS API returns the ResourceNotFound error in this form:
		// ResourceNotFoundException: No addon: vpc-cni found in cluster: tf-acc-test-533189557170672934
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}
	}
	if addon != nil {
		return fmt.Errorf("resource not deleted")
	}

	return nil
}

func waitForUpdateEksAddonContext(ctx context.Context, conn *eks.EKS, clusterName, addonName, updateID string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target: []string{
			eks.UpdateStatusCancelled,
			eks.UpdateStatusFailed,
			eks.UpdateStatusSuccessful,
		},
		Timeout: timeout,
		Refresh: refreshEksUpdateStatusContext(ctx, conn, clusterName, addonName, updateID),
	}
	updateRaw, err := stateConf.WaitForStateContext(ctx)

	if err != nil {
		return err
	}

	update := updateRaw.(*eks.Update)

	if aws.StringValue(update.Status) == eks.UpdateStatusSuccessful {
		return nil
	}

	var detailedErrors []string
	for i, updateError := range update.Errors {
		detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s", i+1, aws.StringValue(updateError.ErrorCode), aws.StringValue(updateError.ErrorMessage)))
	}

	return fmt.Errorf("EKS Addon update (%s) not successful (%s): Errors:\n%s", updateID, aws.StringValue(update.Status), strings.Join(detailedErrors, "\n"))
}

func refreshEksUpdateStatusContext(ctx context.Context, conn *eks.EKS, clusterName, addonName, updateID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeUpdateInput{
			Name:      aws.String(clusterName),
			AddonName: aws.String(addonName),
			UpdateId:  aws.String(updateID),
		}

		output, err := conn.DescribeUpdateWithContext(ctx, input)
		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Update == nil {
			return nil, "", fmt.Errorf("EKS Addon update (%s) missing", updateID)
		}

		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}

func refreshEksAddonStatusContext(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if err != nil {
			return 42, "", err
		}
		addon := output.Addon
		if addon == nil {
			return addon, "", fmt.Errorf("EKS Cluster Addon (%s) missing", clusterName)
		}
		return addon, aws.StringValue(addon.Status), nil
	}
}

func resourceAwsEksAddonParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected cluster-name:addon-name", id)
	}

	return parts[0], parts[1], nil
}
