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
			State: resourceAwsSsoPermissionSetImport,
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
					validation.StringMatch(regexp.MustCompile(`^arn:aws(-[a-z]+)*:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws(-[a-z]+)*:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
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
				Default:      "PT1H",
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

			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
				Set: schema.HashString,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSsoPermissionSetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	permissionSetArn := d.Id()
	instanceArn, err := resourceAwsSsoPermissionSetParseID(permissionSetArn)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error parsing AWS Permission Set (%s) for import: %s", permissionSetArn, err)
	}

	ssoadminconn := meta.(*AWSClient).ssoadminconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	permissionSetResp, permissionSetErr := ssoadminconn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})

	if permissionSetErr != nil {
		return []*schema.ResourceData{}, permissionSetErr
	}

	permissionSet := permissionSetResp.PermissionSet

	log.Printf("[DEBUG] Getting Inline Policy for AWS SSO Permission Set")
	inlinePolicyResp, inlinePolicyErr := ssoadminconn.GetInlinePolicyForPermissionSet(&ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if inlinePolicyErr != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing Inline Policy for AWS SSO Permission Set (%s): %s", permissionSetArn, inlinePolicyErr)
	}

	log.Printf("[DEBUG] Getting Managed Policies for AWS SSO Permission Set")
	managedPoliciesResp, managedPoliciesErr := ssoadminconn.ListManagedPoliciesInPermissionSet(&ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if managedPoliciesErr != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing Managed Policies for AWS SSO Permission Set (%s): %s", permissionSetArn, managedPoliciesErr)
	}
	var managedPolicyArns []string
	for _, managedPolicy := range managedPoliciesResp.AttachedManagedPolicies {
		managedPolicyArns = append(managedPolicyArns, aws.StringValue(managedPolicy.Arn))
	}

	tags, err := keyvaluetags.SsoListTags(ssoadminconn, permissionSetArn, instanceArn)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error listing tags during AWS SSO Permission Set (%s) import: %s", permissionSetArn, err)
	}

	err = d.Set("instance_arn", instanceArn)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("arn", permissionSetArn)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("name", permissionSet.Name)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("description", permissionSet.Description)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("session_duration", permissionSet.SessionDuration)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("relay_state", permissionSet.RelayState)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("inline_policy", inlinePolicyResp.InlinePolicy)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	err = d.Set("managed_policy_arns", managedPolicyArns)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing AWS SSO Permission Set (%s) tags: %s", permissionSetArn, err)
	}
	d.SetId(permissionSetArn)

	return []*schema.ResourceData{d}, nil
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

	createPermissionSetResp, createPermissionerr := ssoadminconn.CreatePermissionSet(params)
	if createPermissionerr != nil {
		return fmt.Errorf("Error creating AWS SSO Permission Set: %s", createPermissionerr)
	}

        if createPermissionSetResp == nil || createPermissionSetResp.PermissionSet == nil {
            return fmt.Errorf("error creating AWS SSO Permission Set (%s): empty output, d.Get("name").(string))
        }
	permissionSetArn := createPermissionSetResp.PermissionSet.PermissionSetArn
	d.SetId(aws.StringValue(permissionSetArn))

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
	instanceArn := d.Get("instance_arn").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Reading AWS SSO Permission Set: %s", permissionSetArn)

	permissionSetResp, permissionerr := ssoadminconn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})

	if isAWSErr(permissionerr, ssoadmin.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] AWS SSO Permission Set (%s) not found, removing from state", permissionSetArn)
		d.SetId("")
		return nil
	}

	if permissionerr != nil {
		return fmt.Errorf("Error getting AWS SSO Permission Set: %s", permissionerr)
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
	var managedPolicyArns []string
	for _, managedPolicy := range managedPoliciesResp.AttachedManagedPolicies {
		managedPolicyArns = append(managedPolicyArns, aws.StringValue(managedPolicy.Arn))
	}

	tags, err := keyvaluetags.SsoListTags(ssoadminconn, permissionSetArn, instanceArn)
	if err != nil {
		return fmt.Errorf("Error listing tags for AWS SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	d.Set("arn", permissionSetArn)
	err = d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("instance_arn", instanceArn)
	if err != nil {
		return err
	}
	err = d.Set("name", permissionSet.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", permissionSet.Description)
	if err != nil {
		return err
	}
	err = d.Set("session_duration", permissionSet.SessionDuration)
	if err != nil {
		return err
	}
	err = d.Set("relay_state", permissionSet.RelayState)
	if err != nil {
		return err
	}
	err = d.Set("inline_policy", inlinePolicyResp.InlinePolicy)
	if err != nil {
		return err
	}
	if err = d.Set("managed_policy_arns", managedPolicyArns); err != nil {
		return err
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSsoPermissionSetUpdate(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn

	permissionSetArn := d.Id()
	instanceArn := d.Get("instance_arn").(string)

	log.Printf("[DEBUG] Updating AWS SSO Permission Set: %s", permissionSetArn)

	if d.HasChanges("description", "relay_state", "session_duration") {
		input := &ssoadmin.UpdatePermissionSetInput{
			PermissionSetArn: aws.String(permissionSetArn),
			InstanceArn:      aws.String(instanceArn),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("relay_state") {
			input.RelayState = aws.String(d.Get("relay_state").(string))
		}

		if d.HasChange("session_duration") {
			input.SessionDuration = aws.String(d.Get("session_duration").(string))
		}

		log.Printf("[DEBUG] Updating AWS SSO Permission Set: %s", input)
		_, permissionerr := ssoadminconn.UpdatePermissionSet(input)
		if permissionerr != nil {
			return fmt.Errorf("Error updating AWS SSO Permission Set: %s", permissionerr)
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

	if d.HasChange("managed_policy_arns") {
		o, n := d.GetChange("managed_policy_arns")

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
	if d.HasChanges("description", "relay_state", "session_duration", "inline_policy", "managed_policy_arns", "tags") {

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

		status := provisionResponse.PermissionSetProvisioningStatus

		_, waitErr := waitForPermissionSetProvisioning(ssoadminconn, instanceArn, aws.StringValue(status.RequestId), d.Timeout(schema.TimeoutUpdate))
		if waitErr != nil {
			return waitErr
		}
	}

	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetDelete(d *schema.ResourceData, meta interface{}) error {
	ssoadminconn := meta.(*AWSClient).ssoadminconn

	permissionSetArn := d.Id()
	instanceArn := d.Get("instance_arn").(string)

	log.Printf("[INFO] Deleting AWS SSO Permission Set: %s", permissionSetArn)

	params := &ssoadmin.DeletePermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := ssoadminconn.DeletePermissionSet(params)

	if err != nil {
		if isAWSErr(err, ssoadmin.ErrCodeResourceNotFoundException, "") {
			log.Printf("[DEBUG] AWS SSO Permission Set not found")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error deleting AWS SSO Permission Set (%s): %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

func attachPoliciesToPermissionSet(ssoadminconn *ssoadmin.SSOAdmin, d *schema.ResourceData, permissionSetArn *string, instanceArn *string) error {

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

	if v, ok := d.GetOk("managed_policy_arns"); ok {
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
	// id = arn:${Partition}:sso:::permissionSet/${InstanceID}/${PermissionSetID}
	idFormatErr := fmt.Errorf("Unexpected format of ARN (%s), expected arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}", id)
	permissionSetARN, err := arn.Parse(id)
	if err != nil {
		return "", idFormatErr
	}

	// We need:
	//  * The InstanceID portion of the permission set ARN resource (arn:aws:sso:::permissionSet/${InstanceId}/${PermissionSetId})
	// Split up the resource of the permission set ARN
	resourceParts := strings.Split(permissionSetARN.Resource, "/")
	if len(resourceParts) != 3 || resourceParts[0] != "permissionSet" || resourceParts[1] == "" || resourceParts[2] == "" {
		return "", idFormatErr
	}

	// resourceParts = ["permissionSet","ins-123456A", "ps-56789B"]
	instanceARN := &arn.ARN{
		Partition: permissionSetARN.Partition,
		Service:   permissionSetARN.Service,
		Resource:  fmt.Sprintf("instance/%s", resourceParts[1]),
	}

	return instanceARN.String(), nil
}

func waitForPermissionSetProvisioning(ssoadminconn *ssoadmin.SSOAdmin, instanceArn string, requestID string, timeout time.Duration) (*ssoadmin.PermissionSetProvisioningStatus, error) {

	stateConf := resource.StateChangeConf{
		Delay:      AWSSSOPermissionSetProvisioningRetryDelay,
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Timeout:    timeout,
		MinTimeout: AWSSSOPermissionSetProvisioningRetryMinTimeout,
		Refresh:    resourceAwsSsoPermissionSetProvisioningRefreshFunc(ssoadminconn, requestID, instanceArn),
	}
	status, err := stateConf.WaitForState()
	if err != nil {
		return nil, fmt.Errorf("Error waiting for AWS SSO Permission Set provisioning status: %s", err)
	}
	return status.(*ssoadmin.PermissionSetProvisioningStatus), nil
}

func resourceAwsSsoPermissionSetProvisioningRefreshFunc(ssoadminconn *ssoadmin.SSOAdmin, requestID, instanceArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		resp, err := ssoadminconn.DescribePermissionSetProvisioningStatus(input)
		if err != nil {
			return resp, "", fmt.Errorf("Error describing permission set provisioning status: %s", err)
		}
		status := resp.PermissionSetProvisioningStatus
		if aws.StringValue(status.Status) == ssoadmin.StatusValuesFailed {
			return resp, ssoadmin.StatusValuesFailed, fmt.Errorf("Failed to provision AWS SSO Permission Set (%s): %s", aws.StringValue(status.PermissionSetArn), aws.StringValue(status.FailureReason))
		}
		return status, aws.StringValue(status.Status), nil

	}
}
