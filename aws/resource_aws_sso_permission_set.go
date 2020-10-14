package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	AWSSSOPermissionSetCreateTimeout               = 5 * time.Minute
	AWSSSOPermissionSetUpdateTimeout               = 10 * time.Minute
	AWSSSOPermissionSetDeleteTimeout               = 5 * time.Minute
	AWSSSOPermissionSetProvisioningRetryDelay      = 5 * time.Second
	AWSSSOPermissionSetProvisioningRetryMinTimeout = 3 * time.Second
)

func resourceAwsSsoPermissionSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoPermissionSetCreate,
		Read:   resourceAwsSsoPermissionSetRead,
		Update: resourceAwsSsoPermissionSetUpdate,
		Delete: resourceAwsSsoPermissionSetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				instanceArn, err := resourceAwsSsoPermissionSetParseID(d.Id())
				if err != nil {
					return nil, fmt.Errorf("Error parsing AWS Permission Set ID %s: %s", d.Id(), err)
				}

				d.Set("instance_arn", instanceArn)
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSSSOPermissionSetCreateTimeout),
			Update: schema.DefaultTimeout(AWSSSOPermissionSetUpdateTimeout),
			Delete: schema.DefaultTimeout(AWSSSOPermissionSetDeleteTimeout),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioning_created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioning_failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioning_request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioning_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]+$`), "must match [\\w+=,.@-]"),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 700),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*$`), "must match [\\p{L}\\p{M}\\p{Z}\\p{S}\\p{N}\\p{P}]"),
				),
			},

			"session_duration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"relay_state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 240),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9&$@#\\\/%?=~\-_'"|!:,.;*+\[\]\(\)\{\} ]+$`), "must match [a-zA-Z0-9&$@#\\\\\\/%?=~\\-_'\"|!:,.;*+\\[\\]\\(\\)\\{\\} ]"),
				),
			},

			"inline_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"managed_policies": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSsoPermissionSetCreate(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn

	log.Printf("[INFO] Creating AWS SSO Permission Set")

	instanceArn := aws.String(d.Get("instance_arn").(string))

	params := &ssoadmin.CreatePermissionSetInput{
		InstanceArn: instanceArn,
		Name:        aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("relay_state"); ok {
		params.RelayState = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_duration"); ok {
		params.SessionDuration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SsoTags()
	}

	createPermissionSetResp, createPermissionSetErr := ssoadminconn.CreatePermissionSet(params)
	if createPermissionSetErr != nil {
		return fmt.Errorf("Error creating AWS SSO Permission Set: %s", createPermissionSetErr)
	}

	permissionSetArn := createPermissionSetResp.PermissionSet.PermissionSetArn
	d.SetId(*permissionSetArn)

	if attachPoliciesErr := attachPoliciesToPermissionSet(ssoadminconn, d, permissionSetArn, instanceArn); attachPoliciesErr != nil {
		return attachPoliciesErr
	}

	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var permissionSet *ssoadmin.PermissionSet
	permissionSetArn := d.Id()
	instanceArn, err := resourceAwsSsoPermissionSetParseID(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing AWS Permission Set ID %s: %s", d.Id(), err)
	}

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Reading AWS SSO Permission Set: %s", permissionSetArn)

	permissionSetResp, permissionSetErr := ssoadminconn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})

	if permissionSetErr != nil {
		return fmt.Errorf("Error getting AWS SSO Permission Set: %s", permissionSetErr)
	}
	if aws.StringValue(permissionSetResp.PermissionSet.Name) == name {
		permissionSet = permissionSetResp.PermissionSet
	}

	if permissionSet == nil {
		log.Printf("[WARN] AWS SSO Permission Set %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Found AWS SSO Permission Set: %s", permissionSet)

	log.Printf("[DEBUG] Getting Inline Policy for AWS SSO Permission Set")
	inlinePolicyResp, inlinePolicyErr := ssoadminconn.GetInlinePolicyForPermissionSet(&ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if inlinePolicyErr != nil {
		return fmt.Errorf("Error getting Inline Policy for AWS SSO Permission Set: %s", inlinePolicyErr)
	}

	log.Printf("[DEBUG] Getting Managed Policies for AWS SSO Permission Set")
	managedPoliciesResp, managedPoliciesErr := ssoadminconn.ListManagedPoliciesInPermissionSet(&ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if managedPoliciesErr != nil {
		return fmt.Errorf("Error getting Managed Policies for AWS SSO Permission Set: %s", managedPoliciesErr)
	}
	managedPoliciesSet := &schema.Set{
		F: permissionSetManagedPoliciesHash,
	}
	for _, managedPolicy := range managedPoliciesResp.AttachedManagedPolicies {
		managedPoliciesSet.Add(map[string]interface{}{
			"arn":  aws.StringValue(managedPolicy.Arn),
			"name": aws.StringValue(managedPolicy.Name),
		})
	}

	tags, err := keyvaluetags.SsoListTags(ssoadminconn, permissionSetArn, instanceArn)
	if err != nil {
		return fmt.Errorf("Error listing tags for AWS SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	d.Set("arn", permissionSetArn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("instance_arn", instanceArn)
	d.Set("name", permissionSet.Name)
	d.Set("description", permissionSet.Description)
	d.Set("session_duration", permissionSet.SessionDuration)
	d.Set("relay_state", permissionSet.RelayState)
	d.Set("inline_policy", inlinePolicyResp.InlinePolicy)
	d.Set("managed_policies", managedPoliciesSet)
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSsoPermissionSetUpdate(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn

	permissionSetArn := d.Id()
	instanceArn, err := resourceAwsSsoPermissionSetParseID(permissionSetArn)
	if err != nil {
		return fmt.Errorf("Error parsing AWS Permission Set ID %s: %s", permissionSetArn, err)
	}

	log.Printf("[DEBUG] Updating AWS SSO Permission Set: %s", permissionSetArn)

	if d.HasChanges("description", "relay_state", "session_duration") {
		input := &ssoadmin.UpdatePermissionSetInput{
			PermissionSetArn: aws.String(permissionSetArn),
			InstanceArn:      aws.String(instanceArn),
			Description:      aws.String(d.Get("description").(string)),
			RelayState:       aws.String(d.Get("relay_state").(string)),
			SessionDuration:  aws.String(d.Get("session_duration").(string)),
		}

		log.Printf("[DEBUG] Updating AWS SSO Permission Set: %s", input)
		_, permissionSetErr := ssoadminconn.UpdatePermissionSet(input)
		if permissionSetErr != nil {
			return fmt.Errorf("Error updating AWS SSO Permission Set: %s", permissionSetErr)
		}
	}

	if d.HasChange("tags") {
		oldTags, newTags := d.GetChange("tags")
		if updateTagsErr := keyvaluetags.SsoUpdateTags(ssoadminconn, d.Get("arn").(string), d.Get("instance_arn").(string), oldTags, newTags); updateTagsErr != nil {
			return fmt.Errorf("Error updating tags: %s", updateTagsErr)
		}
	}

	if v, ok := d.GetOk("inline_policy"); ok {
		log.Printf("[DEBUG] AWS SSO Permission Set %s updating IAM inline policy", permissionSetArn)

		inlinePolicy := aws.String(v.(string))

		updateInput := &ssoadmin.PutInlinePolicyToPermissionSetInput{
			InlinePolicy:     inlinePolicy,
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
		}

		_, inlinePolicyErr := ssoadminconn.PutInlinePolicyToPermissionSet(updateInput)
		if inlinePolicyErr != nil {
			return fmt.Errorf("Error attaching IAM inline policy to AWS SSO Permission Set: %s", inlinePolicyErr)
		}
	} else if d.HasChange("inline_policy") {
		deleteInput := &ssoadmin.DeleteInlinePolicyFromPermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
		}

		_, inlinePolicyErr := ssoadminconn.DeleteInlinePolicyFromPermissionSet(deleteInput)
		if inlinePolicyErr != nil {
			return fmt.Errorf("Error deleting IAM inline policy from AWS SSO Permission Set: %s", inlinePolicyErr)
		}
	}

	if d.HasChange("managed_policies") {
		o, n := d.GetChange("managed_policies")

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			input := &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
				InstanceArn:      aws.String(instanceArn),
				ManagedPolicyArn: aws.String(v.(string)),
				PermissionSetArn: aws.String(permissionSetArn),
			}

			_, managedPoliciesErr := ssoadminconn.DetachManagedPolicyFromPermissionSet(input)
			if managedPoliciesErr != nil {
				return fmt.Errorf("Error detaching Managed Policy from AWS SSO Permission Set: %s", managedPoliciesErr)
			}
		}

		additionList := ns.Difference(os)
		for _, v := range additionList.List() {
			input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
				InstanceArn:      aws.String(instanceArn),
				ManagedPolicyArn: aws.String(v.(string)),
				PermissionSetArn: aws.String(permissionSetArn),
			}

			_, managedPoliciesErr := ssoadminconn.AttachManagedPolicyToPermissionSet(input)
			if managedPoliciesErr != nil {
				return fmt.Errorf("Error attaching Managed Policy to AWS SSO Permission Set: %s", managedPoliciesErr)
			}
		}
	}

	// Reprovision if anything has changed
	if d.HasChanges("description", "relay_state", "session_duration", "inline_policy", "managed_policies", "tags") {

		// Auto provision all accounts
		targetType := ssoadmin.ProvisionTargetTypeAllProvisionedAccounts
		provisionInput := &ssoadmin.ProvisionPermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
			TargetType:       aws.String(targetType),
		}

		log.Printf("[INFO] Provisioning AWS SSO Permission Set")
		provisionResponse, err := ssoadminconn.ProvisionPermissionSet(provisionInput)
		if err != nil {
			return fmt.Errorf("Error provisioning AWS SSO Permission Set (%s): %w", d.Id(), err)
		}

		if provisionResponse != nil && provisionResponse.PermissionSetProvisioningStatus != nil {
			status := provisionResponse.PermissionSetProvisioningStatus

			if status.CreatedDate != nil {
				d.Set("created_date", status.CreatedDate.Format(time.RFC3339))
			}

			wait := resource.StateChangeConf{
				Delay:      AWSSSOPermissionSetProvisioningRetryDelay,
				Pending:    []string{ssoadmin.StatusValuesInProgress},
				Target:     []string{ssoadmin.StatusValuesSucceeded},
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				MinTimeout: AWSSSOPermissionSetProvisioningRetryMinTimeout,
				Refresh:    resourceAwsSsoPermissionSetProvisioningRefreshFunc(ssoadminconn, aws.StringValue(status.RequestId), instanceArn),
			}

			if _, err := wait.WaitForState(); err != nil {
				return fmt.Errorf("Error waiting for AWS SSO Permission Set (%s) provisioning: %w", d.Id(), err)
			}
		}
	}

	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetDelete(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn

	permissionSetArn := d.Id()
	instanceArn, parseErr := resourceAwsSsoPermissionSetParseID(permissionSetArn)
	if parseErr != nil {
		return fmt.Errorf("Error parsing AWS Permission Set ID %s: %s", permissionSetArn, parseErr)
	}

	log.Printf("[INFO] Deleting AWS SSO Permission Set: %s", permissionSetArn)

	params := &ssoadmin.DeletePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := ssoadminconn.DeletePermissionSet(params)
	if err != nil {
		return fmt.Errorf("Error deleting AWS SSO Permission Set (%s): %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func attachPoliciesToPermissionSet(ssoadminconn *ssoadmin.SSOAdmin, d *schema.ResourceData, instanceArn *string, permissionSetArn *string) error {

	if v, ok := d.GetOk("inline_policy"); ok {
		log.Printf("[INFO] Attaching IAM inline policy to AWS SSO Permission Set")

		inlinePolicy := aws.String(v.(string))

		input := &ssoadmin.PutInlinePolicyToPermissionSetInput{
			InlinePolicy:     inlinePolicy,
			InstanceArn:      instanceArn,
			PermissionSetArn: permissionSetArn,
		}

		_, inlinePolicyErr := ssoadminconn.PutInlinePolicyToPermissionSet(input)
		if inlinePolicyErr != nil {
			return fmt.Errorf("Error attaching IAM inline policy to AWS SSO Permission Set: %s", inlinePolicyErr)
		}
	}

	if v, ok := d.GetOk("managed_policies"); ok {
		log.Printf("[INFO] Attaching Managed Policies to AWS SSO Permission Set")

		managedPolicies := expandStringSet(v.(*schema.Set))

		for _, managedPolicyArn := range managedPolicies {

			input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
				InstanceArn:      instanceArn,
				ManagedPolicyArn: managedPolicyArn,
				PermissionSetArn: permissionSetArn,
			}

			_, managedPoliciesErr := ssoadminconn.AttachManagedPolicyToPermissionSet(input)
			if managedPoliciesErr != nil {
				return fmt.Errorf("Error attaching Managed Policy to AWS SSO Permission Set: %s", managedPoliciesErr)
			}
		}
	}

	return nil
}

func resourceAwsSsoPermissionSetParseID(id string) (string, error) {
	// id = arn:aws:sso:::permissionSet/${InstanceID}/${PermissionSetID}
	idFormatErr := fmt.Errorf("Unexpected format of AWS Permission Set ID (%s), expected format arn:aws:sso:::permissionSet/ins-123456A/ps-56789B", id)
	permissionSetARN, err := arn.Parse(id)
	if err != nil {
		return "", idFormatErr
	}

	// We need:
	//  * The InstanceID portion of the permission set ARN resource (arn:aws:sso:::permissionSet/ins-123456A/ps-56789B)
	// Split up the resource of the permission set ARN
	resourceParts := strings.SplitN(permissionSetARN.Resource, "/", 3)
	if len(resourceParts) != 3 {
		return "", idFormatErr
	}
	// resourceParts = ["permissionSet","ins-123456A", "ps-56789B"]
	instanceARN := &arn.ARN{
		AccountID: permissionSetARN.AccountID,
		Partition: permissionSetARN.Partition,
		Region:    permissionSetARN.Region,
		Service:   "instance",
		Resource:  resourceParts[1],
	}

	return instanceARN.String(), nil
}

func resourceAwsSsoPermissionSetProvisioningRefreshFunc(ssoadminconn *ssoadmin.SSOAdmin, instanceArn, requestID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		return resourceAwsSsoPermissionSetProvisioningWait(ssoadminconn, input)
	}
}

func resourceAwsSsoPermissionSetProvisioningWait(ssoadminconn *ssoadmin.SSOAdmin, input *ssoadmin.DescribePermissionSetProvisioningStatusInput) (result interface{}, state string, err error) {

	resp, err := ssoadminconn.DescribePermissionSetProvisioningStatus(input)

	if aws.StringValue(resp.PermissionSetProvisioningStatus.Status) == ssoadmin.StatusValuesFailed {
		return nil, ssoadmin.StatusValuesFailed, fmt.Errorf("Failed to provision AWS SSO Permission Set (%s): %s", aws.StringValue(resp.PermissionSetProvisioningStatus.PermissionSetArn), aws.StringValue(resp.PermissionSetProvisioningStatus.FailureReason))
	}

	if err != nil {
		return nil, *resp.PermissionSetProvisioningStatus.Status, err
	}
	return true, *resp.PermissionSetProvisioningStatus.Status, nil
}
