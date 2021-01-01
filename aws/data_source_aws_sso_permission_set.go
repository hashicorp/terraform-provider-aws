package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSsoPermissionSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsoPermissionSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws(-[a-z]+)*:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws(-[a-z]+)*:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]+$`), "must match [\\w+=,.@-]"),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"session_duration": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"inline_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"managed_policy_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsSsoPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	instanceArn := d.Get("instance_arn").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Reading AWS SSO Permission Sets")

	var permissionSetArn string
	var permissionSet *ssoadmin.PermissionSet
	var permissionSetErr error

	req := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	}
	err := conn.ListPermissionSetsPages(req, func(page *ssoadmin.ListPermissionSetsOutput, lastPage bool) bool {
		if page != nil && len(page.PermissionSets) != 0 {
			for _, ps := range page.PermissionSets {
				permissionSetArn = aws.StringValue(ps)
				log.Printf("[DEBUG] Reading AWS SSO Permission Set: %v", permissionSetArn)
				var permissionSetResp *ssoadmin.DescribePermissionSetOutput
				permissionSetResp, permissionSetErr = conn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
					InstanceArn:      aws.String(instanceArn),
					PermissionSetArn: aws.String(permissionSetArn),
				})
				if permissionSetErr != nil {
					return false
				}
				if aws.StringValue(permissionSetResp.PermissionSet.Name) == name {
					permissionSet = permissionSetResp.PermissionSet
					return false
				}
			}
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error getting AWS SSO Permission Sets: %w", err)
	}

	if permissionSetErr != nil {
		return fmt.Errorf("Error getting AWS SSO Permission Set: %w", permissionSetErr)
	}

	if permissionSet == nil {
		return fmt.Errorf("AWS SSO Permission Set %s not found", name)
	}

	log.Printf("[DEBUG] Found AWS SSO Permission Set: %s", permissionSet)

	log.Printf("[DEBUG] Getting Inline Policy for AWS SSO Permission Set")
	inlinePolicyResp, inlinePolicyErr := conn.GetInlinePolicyForPermissionSet(&ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	})
	if inlinePolicyErr != nil {
		return fmt.Errorf("Error getting Inline Policy for AWS SSO Permission Set: %s", inlinePolicyErr)
	}

	log.Printf("[DEBUG] Getting Managed Policies for AWS SSO Permission Set")
	var managedPolicyArns []string
	managedPoliciesReq := &ssoadmin.ListManagedPoliciesInPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}
	managedPoliciesErr := conn.ListManagedPoliciesInPermissionSetPages(managedPoliciesReq, func(page *ssoadmin.ListManagedPoliciesInPermissionSetOutput, lastPage bool) bool {
		for _, managedPolicy := range page.AttachedManagedPolicies {
			managedPolicyArns = append(managedPolicyArns, aws.StringValue(managedPolicy.Arn))
		}
		return !lastPage
	})
	if managedPoliciesErr != nil {
		return fmt.Errorf("Error getting Managed Policies for AWS SSO Permission Set: %s", managedPoliciesErr)
	}

	tags, tagsErr := keyvaluetags.SsoListTags(conn, permissionSetArn, instanceArn)
	if tagsErr != nil {
		return fmt.Errorf("Error listing tags for AWS SSO Permission Set (%s): %s", permissionSetArn, tagsErr)
	}

	d.SetId(permissionSetArn)
	d.Set("arn", permissionSetArn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("instance_arn", instanceArn)
	d.Set("name", permissionSet.Name)
	d.Set("description", permissionSet.Description)
	d.Set("session_duration", permissionSet.SessionDuration)
	d.Set("relay_state", permissionSet.RelayState)
	d.Set("inline_policy", inlinePolicyResp.InlinePolicy)
	d.Set("managed_policy_arns", managedPolicyArns)
	tagsMapErr := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	if tagsMapErr != nil {
		return fmt.Errorf("Error setting tags: %s", tagsMapErr)
	}

	return nil
}
