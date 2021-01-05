package aws

import (
	"context"
	"fmt"
	"log"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.CreateAddonInput{
		AddonName:          aws.String(addonName),
		ClusterName:        aws.String(clusterName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.ResolveConflicts = aws.String(v.(string))
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
			//
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

	d.SetId(addonName)

	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating},
		Target:  []string{eks.AddonStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksAddonStatusContext(ctx, conn, addonName, clusterName),
	}
	// TODO: Handle unexpected sate CREATE_FAILED and retry?
	// Error: unexpected state 'CREATE_FAILED', wanted target 'ACTIVE'. last error: %!s(<nil>)
	// How to reproduce: taint eks addon resource and try to do terraform apply

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	log.Printf("[DEBUG] Reading EKS Addon: %s", input)
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

	d.SetId(addonName)
	d.Set("arn", addon.AddonArn)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)
	d.Set("status", addon.Status)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).String())
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).String())

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(addon.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags attribute: %s", err)
	}

	return nil
}

func resourceAwsEksAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.UpdateAddonInput{
		AddonName:          aws.String(addonName),
		ClusterName:        aws.String(clusterName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

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

		err = waitForUpdateEksAddonContext(ctx, conn, addonName, clusterName, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.Errorf("error waiting for EKS Addon (%s) version update (%s): %s", d.Id(), updateID, err)
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

		err = waitForUpdateEksAddonContext(ctx, conn, addonName, clusterName, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.Errorf("error waiting for EKS Addon (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.DeleteAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	log.Printf("[DEBUG] Deleting EKS Cluster Addon: %s", d.Id())
	_, err := conn.DeleteAddonWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error deleting EKS Addon (%s): %s", d.Id(), err)
	}

	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.AddonStatusActive,
			eks.AddonStatusDeleting,
		},
		Target:  []string{""},
		Timeout: d.Timeout(schema.TimeoutDefault),
		Refresh: refreshEksAddonStatusContext(ctx, conn, addonName, clusterName),
	}
	addon, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// Sometimes the EKS API returns the ResourceNotFound error in this form:
		// ResourceNotFoundException: No addon: vpc-cni found in cluster: test-eks-XmfX9Lur
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "No cluster found for name:") {
			return nil
		}
	}
	if addon != nil {
		diag.Errorf("resource not deleted")
	}

	return nil
}

func waitForUpdateEksAddonContext(ctx context.Context, conn *eks.EKS, addonName, clusterName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusUpdating},
		Target:  []string{eks.AddonStatusActive},
		Timeout: timeout,
		Refresh: refreshEksAddonStatusContext(ctx, conn, addonName, clusterName),
	}
	updateRaw, err := stateConf.WaitForStateContext(ctx)

	if err != nil {
		return err
	}

	update := updateRaw.(*eks.Update)

	if aws.StringValue(update.Status) == eks.AddonStatusActive {
		return nil
	}

	var detailedErrors []string
	for i, updateError := range update.Errors {
		detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s", i+1, aws.StringValue(updateError.ErrorCode), aws.StringValue(updateError.ErrorMessage)))
	}

	return fmt.Errorf("EKS Cluster (%s) Addon (%s) status (%s) not successful: Errors:\n%s", clusterName, addonName, aws.StringValue(update.Status), strings.Join(detailedErrors, "\n"))
}

func refreshEksAddonStatusContext(ctx context.Context, conn *eks.EKS, addonName, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			AddonName:   aws.String(addonName),
			ClusterName: aws.String(clusterName),
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
