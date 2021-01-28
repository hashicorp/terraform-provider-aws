package aws

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func resourceAwsIamRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamRoleCreate,
		Read:   resourceAwsIamRoleRead,
		Update: resourceAwsIamRoleUpdate,
		Delete: resourceAwsIamRoleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsIamRoleImport,
		},
		CustomizeDiff: resourceAwsIamRoleInlineCustDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},

			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				),
			},

			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},

			"permissions_boundary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringDoesNotMatch(regexp.MustCompile("[“‘]"), "cannot contain specially formatted single or double quotes: [“‘]"),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), `must satisfy regular expression pattern: [\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*)`),
				),
			},

			"assume_role_policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},

			"force_detach_policies": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"max_session_duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(3600, 43200),
			},

			"tags": tagsSchema(),

			"inline_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     validateIAMPolicyJson,
							DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateIamRolePolicyName,
							// ConflictsWith: []string{"inline_policy.0.name_prefix"},
							// Not working: prevents two separate policies with
							// one having a name and the other a prefix
						},
						"name_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateIamRolePolicyNamePrefix,
						},
					},
				},
			},

			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceAwsIamRoleImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("force_detach_policies", false)
	return []*schema.ResourceData{d}, nil
}

func resourceAwsIamRoleCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	request := &iam.CreateRoleInput{
		Path:                     aws.String(d.Get("path").(string)),
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(d.Get("assume_role_policy").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_session_duration"); ok {
		request.MaxSessionDuration = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("permissions_boundary"); ok {
		request.PermissionsBoundary = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		request.Tags = keyvaluetags.New(v).IgnoreAws().IamTags()
	}

	var createResp *iam.CreateRoleOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		createResp, err = iamconn.CreateRole(request)
		// IAM users (referenced in Principal field of assume policy)
		// can take ~30 seconds to propagate in AWS
		if isAWSErr(err, "MalformedPolicyDocument", "Invalid principal in policy") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		createResp, err = iamconn.CreateRole(request)
	}
	if err != nil {
		return fmt.Errorf("Error creating IAM Role %s: %s", name, err)
	}

	if policyData, ok := d.GetOk("inline_policy"); ok {
		inlinePolicies := policyData.(*schema.Set).List()
		if err := addInlinePoliciesToRole(iamconn, inlinePolicies, name); err != nil {
			return fmt.Errorf("failed to add inline policies to IAM role %s, error: %s", name, err)
		}
	}

	if policies, ok := d.GetOk("managed_policy_arns"); ok {
		managedPolicies := expandStringList(policies.(*schema.Set).List())
		if err := addManagedPoliciesToRole(iamconn, managedPolicies, name); err != nil {
			return fmt.Errorf("failed to add managed policies to IAM role %s, error: %s", name, err)
		}
	}

	d.SetId(aws.StringValue(createResp.Role.RoleName))
	return resourceAwsIamRoleRead(d, meta)
}

func resourceAwsIamRoleRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &iam.GetRoleInput{
		RoleName: aws.String(d.Id()),
	}

	getResp, err := iamconn.GetRole(request)
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			log.Printf("[WARN] IAM Role %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM Role %s: %s", d.Id(), err)
	}

	if getResp == nil || getResp.Role == nil {
		log.Printf("[WARN] IAM Role %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	role := getResp.Role

	d.Set("arn", role.Arn)
	if err := d.Set("create_date", role.CreateDate.Format(time.RFC3339)); err != nil {
		return err
	}
	d.Set("description", role.Description)
	d.Set("max_session_duration", role.MaxSessionDuration)
	d.Set("name", role.RoleName)
	d.Set("path", role.Path)
	if role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", role.RoleId)

	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(role.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	assumRolePolicy, err := url.QueryUnescape(*role.AssumeRolePolicyDocument)
	if err != nil {
		return err
	}
	if err := d.Set("assume_role_policy", assumRolePolicy); err != nil {
		return err
	}

	inlinePolicies, err := readInlinePoliciesForRole(iamconn, *role.RoleName)
	if err != nil {
		return fmt.Errorf("failed to read inline policies for IAM role %s, error: %s", d.Id(), err)
	}
	if err = d.Set("inline_policy", inlinePolicies); err != nil {
		return fmt.Errorf("failed to set inline policies for IAM role %s, error: %s", d.Id(), err)
	}

	managedPolicies, err := readManagedPoliciesForRole(iamconn, *role.RoleName)
	if err != nil {
		return fmt.Errorf("failed to read managed policy list for IAM role %s, error: %s", *role.RoleName, err)
	}
	if err := d.Set("managed_policy_arns", managedPolicies); err != nil {
		return fmt.Errorf("failed to set managed policy list for IAM role (%s), error: %s", *role.RoleName, err)
	}

	return nil
}

func resourceAwsIamRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if d.HasChange("assume_role_policy") {
		assumeRolePolicyInput := &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(d.Id()),
			PolicyDocument: aws.String(d.Get("assume_role_policy").(string)),
		}
		_, err := iamconn.UpdateAssumeRolePolicy(assumeRolePolicyInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		roleDescriptionInput := &iam.UpdateRoleDescriptionInput{
			RoleName:    aws.String(d.Id()),
			Description: aws.String(d.Get("description").(string)),
		}
		_, err := iamconn.UpdateRoleDescription(roleDescriptionInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("max_session_duration") {
		roleMaxDurationInput := &iam.UpdateRoleInput{
			RoleName:           aws.String(d.Id()),
			MaxSessionDuration: aws.Int64(int64(d.Get("max_session_duration").(int))),
		}
		_, err := iamconn.UpdateRole(roleMaxDurationInput)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error Updating IAM Role (%s) Max Session Duration: %s", d.Id(), err)
		}
	}

	if d.HasChange("permissions_boundary") {
		permissionsBoundary := d.Get("permissions_boundary").(string)
		if permissionsBoundary != "" {
			input := &iam.PutRolePermissionsBoundaryInput{
				PermissionsBoundary: aws.String(permissionsBoundary),
				RoleName:            aws.String(d.Id()),
			}
			_, err := iamconn.PutRolePermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error updating IAM Role permissions boundary: %s", err)
			}
		} else {
			input := &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String(d.Id()),
			}
			_, err := iamconn.DeleteRolePermissionsBoundary(input)
			if err != nil {
				return fmt.Errorf("error deleting IAM Role permissions boundary: %s", err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.IamRoleUpdateTags(iamconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating IAM Role (%s) tags: %s", d.Id(), err)
		}
	}
	if d.HasChange("inline_policy") {
		roleName := d.Get("name").(string)
		o, n := d.GetChange("inline_policy")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		if err := removeInlinePoliciesFromRole(iamconn, remove, roleName); err != nil {
			return fmt.Errorf("failed to remove inline policies for IAM role %s, error: %s", roleName, err)
		}

		if err := addInlinePoliciesToRole(iamconn, add, roleName); err != nil {
			return fmt.Errorf("failed to add inline policies for IAM role %s, error: %s", roleName, err)
		}
	}

	if d.HasChange("managed_policy_arns") {
		roleName := d.Get("name").(string)

		o, n := d.GetChange("managed_policy_arns")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := expandStringList(os.Difference(ns).List())
		add := expandStringList(ns.Difference(os).List())

		if err := removeManagedPoliciesFromRole(iamconn, remove, roleName); err != nil {
			return fmt.Errorf("failed to detach managed policies for IAM role %s, error: %s", roleName, err)
		}

		if err := addManagedPoliciesToRole(iamconn, add, roleName); err != nil {
			return fmt.Errorf("failed to attach managed policies for IAM role %s, error: %s", roleName, err)
		}

	}

	return resourceAwsIamRoleRead(d, meta)
}

func resourceAwsIamRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	err := deleteAwsIamRole(conn, d.Id(), d.Get("force_detach_policies").(bool))
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting IAM Role (%s): %w", d.Id(), err)
	}

	return nil
}

func deleteAwsIamRole(conn *iam.IAM, rolename string, forceDetach bool) error {
	if err := deleteAwsIamRoleInstanceProfiles(conn, rolename); err != nil {
		return fmt.Errorf("unable to detach instance profiles: %w", err)
	}

	if forceDetach {
		if err := deleteAwsIamRolePolicyAttachments(conn, rolename); err != nil {
			return fmt.Errorf("unable to detach policies: %w", err)
		}

		if err := deleteAwsIamRolePolicies(conn, rolename); err != nil {
			return fmt.Errorf("unable to delete inline policies: %w", err)
	// Roles cannot be destroyed when attached to an existing Instance Profile
	resp, err := iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("failed to list profiles for IAM role %s, error: %s", d.Id(), err)
	}

	// Loop and remove this Role from any Profiles
	if len(resp.InstanceProfiles) > 0 {
		for _, i := range resp.InstanceProfiles {
			_, err := iamconn.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: i.InstanceProfileName,
				RoleName:            aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("failed to remove IAM role %s from instance profile, error: %s", d.Id(), err)
			}
		}
	}

	if d.Get("force_detach_policies").(bool) || d.Get("inline_policy.#").(int) > 0 {
		inlinePolicies, err := readInlinePoliciesForRole(iamconn, d.Id())
		if err != nil {
			return fmt.Errorf("failed to read inline policies for IAM role %s, error: %s", d.Id(), err)
		}

		if err := removeInlinePoliciesFromRole(iamconn, inlinePolicies, d.Id()); err != nil {
			return fmt.Errorf("failed to delete inline policies from IAM role %s, error: %s", d.Id(), err)
		}
	}

	if d.Get("force_detach_policies").(bool) || d.Get("managed_policy_arns.#").(int) > 0 {
		managedPolicies, err := readManagedPoliciesForRole(iamconn, d.Id())
		if err != nil {
			return fmt.Errorf("failed to read managed policies for IAM role %s, error: %s", d.Id(), err)
		}

		// convert []string to []*string
		managedPolicyPtrs := []*string{}
		for _, policy := range managedPolicies {
			newVar := policy // necessary to get a new pointer
			managedPolicyPtrs = append(managedPolicyPtrs, &newVar)
		}

		if err := removeManagedPoliciesFromRole(iamconn, managedPolicyPtrs, d.Id()); err != nil {
			return fmt.Errorf("failed to detach managed policies from IAM role %s, error: %s", d.Id(), err)
		}
	}

	deleteRoleInput := &iam.DeleteRoleInput{
		RoleName: aws.String(rolename),
	}
	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRole(deleteRoleInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeDeleteConflictException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRole(deleteRoleInput)
	}

	return err
}

func deleteAwsIamRoleInstanceProfiles(conn *iam.IAM, rolename string) error {
	resp, err := conn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(rolename),
	})
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil

}

func resourceAwsIamRoleInlineCustDiff(diff *schema.ResourceDiff, v interface{}) error {

	// Avoids diffs resulting when inline policies are configured without either
	// name or name prefix, or with a name prefix. In these cases, Terraform
	// generates some or all of the name. Without a customized diff function,
	// comparing the config to the state will always generate a diff since the
	// config has no information about the policy's generated name.
	if diff.HasChange("inline_policy") {

		o, n := diff.GetChange("inline_policy")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		_, err := conn.RemoveRoleFromInstanceProfile(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// a single empty inline_policy in the config produces a diff with
		// inline_policy.# = 0 and subattributes all blank
		if len(os.List()) == 0 && len(ns.List()) == 1 {
			data := (ns.List())[0].(map[string]interface{})
			if data["name"].(string) == "" && data["name_prefix"].(string) == "" && data["policy"].(string) == "" {
				if err := diff.Clear("inline_policy"); err != nil {
					return fmt.Errorf("failed to clear diff for IAM role %s, error: %s", diff.Id(), err)
				}
			}
		}

		// if there's no old or new set, nothing to do - can't match up
		// equivalents between the lists
		if len(os.List()) > 0 && len(ns.List()) > 0 {

			// fast O(n) comparison in case of thousands of policies

			// current state lookup map:
			// key: inline policy doc hash
			// value: string slice with policy names (slice in case of dupes)
			statePolicies := make(map[int]interface{})
			for _, policy := range os.List() {
				data := policy.(map[string]interface{})
				name := data["name"].(string)

				// condition probably not needed, will have been assigned name
				if name != "" {
					docHash := hashcode.String(data["policy"].(string))
					if _, ok := statePolicies[docHash]; !ok {
						statePolicies[docHash] = []string{name}
					} else {
						statePolicies[docHash] = append(statePolicies[docHash].([]string), name)
					}
				}
			}

			// construct actual changes by going through incoming config changes
			configSet := make([]interface{}, 0)
			for _, policy := range ns.List() {
				appended := false
				data := policy.(map[string]interface{})
				namePrefix := data["name_prefix"].(string)
				name := data["name"].(string)

				if namePrefix != "" || (namePrefix == "" && name == "") {
					docHash := hashcode.String(data["policy"].(string))
					if namesFromState, ok := statePolicies[docHash]; ok {
						for i, nameFromState := range namesFromState.([]string) {
							if (namePrefix == "" && name == "") || strings.HasPrefix(nameFromState, namePrefix) {
								// match - we want the state value
								pair := make(map[string]interface{})
								pair["name"] = nameFromState
								pair["policy"] = data["policy"]
								configSet = append(configSet, pair)
								appended = true

								// remove - in case of duplicate policies
								stateSlice := namesFromState.([]string)
								stateSlice = append(stateSlice[:i], stateSlice[i+1:]...)
								if len(stateSlice) == 0 {
									delete(statePolicies, docHash)
								} else {
									statePolicies[docHash] = stateSlice
								}
								break
							}
						}
					}
				}

				if !appended {
					pair := make(map[string]interface{})
					pair["name"] = name
					pair["name_prefix"] = namePrefix
					pair["policy"] = data["policy"]
					configSet = append(configSet, pair)
				}
			}
			if err := diff.SetNew("inline_policy", configSet); err != nil {
				return fmt.Errorf("failed to set new inline policies for IAM role %s, error: %s", diff.Id(), err)
			}
		}
	}

	return nil
}

func readInlinePoliciesForRole(iamconn *iam.IAM, roleName string) ([]interface{}, error) {
	inlinePolicies := make([]interface{}, 0)
	var marker *string
	for {
		resp, err := iamconn.ListRolePolicies(&iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
			Marker:   marker,
		})

		if err != nil {
			return nil, err
		}

		for _, policyName := range resp.PolicyNames {
			policyResp, err := iamconn.GetRolePolicy(&iam.GetRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: policyName,
			})
			if err != nil {
				return nil, err
			}

			json, err := url.QueryUnescape(*policyResp.PolicyDocument)
			if err != nil {
				return nil, err
			}
			pair := make(map[string]interface{})
			pair["name"] = *policyName
			pair["policy"] = json
			inlinePolicies = append(inlinePolicies, pair)
		}

		if !*resp.IsTruncated {
			break
		}
		marker = resp.Marker
	}

	return inlinePolicies, nil
}

func readManagedPoliciesForRole(iamconn *iam.IAM, roleName string) ([]string, error) {
	var managedPolicyList []string
	var marker *string
	for {
		resp, err := iamconn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(roleName),
			Marker:   marker,
		})
		if err != nil {
			return nil, err
		}
		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, parn := range managedPolicies {
		input := &iam.DetachRolePolicyInput{
			PolicyArn: parn,
			RoleName:  aws.String(rolename),

		for _, ap := range resp.AttachedPolicies {
			managedPolicyList = append(managedPolicyList, *ap.PolicyArn)
		}

		if !*resp.IsTruncated {
			break
		}
		marker = resp.Marker
	}
	return managedPolicyList, nil
}

func addInlinePoliciesToRole(iamconn *iam.IAM, inlinePolicies []interface{}, roleName string) error {

	if len(inlinePolicies) == 1 {
		// check for special case: one empty inline policy
		data := inlinePolicies[0].(map[string]interface{})
		if data["name"].(string) == "" && data["name_prefix"].(string) == "" && data["policy"].(string) == "" {
			return nil
		}
	}

		_, err = conn.DetachRolePolicy(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
	for _, policy := range inlinePolicies {
		data := policy.(map[string]interface{})

		policyDoc := data["policy"].(string)
		if policyDoc == "" {
			return fmt.Errorf("policy is required")
		}

		var policyName string

		if v, ok := data["name"]; ok && v.(string) != "" {
			policyName = v.(string)
		} else if v, ok := data["name_prefix"]; ok && v.(string) != "" {
			policyName = resource.PrefixedUniqueId(v.(string))
		} else {
			policyName = resource.UniqueId()
		}

		_, err := iamconn.PutRolePolicy(&iam.PutRolePolicyInput{
			PolicyName:     aws.String(policyName),
			RoleName:       aws.String(roleName),
			PolicyDocument: aws.String(policyDoc),
		})

		if err != nil {
			return fmt.Errorf("failed to add inline policy to IAM role %s, error: %s", roleName, err)
		}
	}

	return nil
}

func removeInlinePoliciesFromRole(iamconn *iam.IAM, inlinePolicies []interface{}, roleName string) error {

	err := conn.ListRolePoliciesPages(input, func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
	})
	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}
	if err != nil {
		return err
	}
		policyName := ""
		if ok {
			policyName = data["name"].(string)
		} else {
			dataS := policy.(map[string]string)
			policyName = dataS["name"]
		}

		_, err := iamconn.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			PolicyName: aws.String(policyName),
			RoleName:   aws.String(roleName),
		})

				log.Printf("[WARN] Inline role policy (%s) was already removed from role (%s)", policyName, roleName)
				continue
			}
			return fmt.Errorf("failed to delete inline policy of IAM role %s, error: %s", roleName, err)
		}
	}
	return nil
}

		_, err := conn.DeleteRolePolicy(input)
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		if err != nil {
			return err
func addManagedPoliciesToRole(iamconn *iam.IAM, managedPolicies []*string, roleName string) error {
	for _, arn := range managedPolicies {
		_, err := iamconn.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String(*arn),
			RoleName:  aws.String(roleName),
		})

		if err != nil {
			return fmt.Errorf("failed to attach managed policy to IAM role %s, error: %s", roleName, err)
		}
	}
	return nil
}

func removeManagedPoliciesFromRole(iamconn *iam.IAM, managedPolicies []*string, roleName string) error {
	for _, arn := range managedPolicies {
		_, err := iamconn.DetachRolePolicy(&iam.DetachRolePolicyInput{
			PolicyArn: aws.String(*arn),
			RoleName:  aws.String(roleName),
		})

		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				log.Printf("[WARN] Managed role policy (%s) was already detached from role (%s)", *arn, roleName)
				continue
			}
			return fmt.Errorf("failed to detach managed policy of IAM role %s, error: %s", roleName, err)
		}
	}
	return nil
}
