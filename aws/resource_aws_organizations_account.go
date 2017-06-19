package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsOrganizationsAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationsAccountCreate,
		Read:   resourceAwsOrganizationsAccountRead,
		Delete: resourceAwsOrganizationsAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsOrganizationsAccountEmail,
			},
			"iam_user_access_to_billing": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{organizations.IAMUserAccessToBillingAllow, organizations.IAMUserAccessToBillingDeny}, true),
			},
			"role_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAwsOrganizationsAccountRoleName,
			},
		},
	}
}

func resourceAwsOrganizationsAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	// Create the account
	createOpts := &organizations.CreateAccountInput{
		AccountName: aws.String(d.Get("name").(string)),
		Email:       aws.String(d.Get("email").(string)),
	}
	if role, ok := d.GetOk("role_name"); ok {
		createOpts.RoleName = aws.String(role.(string))
	}

	if iam_user, ok := d.GetOk("iam_user_access_to_billing"); ok {
		createOpts.IamUserAccessToBilling = aws.String(iam_user.(string))
	}

	log.Printf("[DEBUG] Account create config: %#v", createOpts)

	var err error
	var resp *organizations.CreateAccountOutput
	err = resource.Retry(4*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreateAccount(createOpts)

		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return resource.NonRetryableError(err)
			}
			if isAWSErr(ec2err, organizations.ErrCodeFinalizingOrganizationException, "") {
				log.Printf("[DEBUG] Trying to create account again: %q", ec2err.Message())
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Error creating account: %s", err)
	}
	log.Printf("[DEBUG] Account create response: %#v", resp)

	requestId := *resp.CreateAccountStatus.Id

	// Wait for the account to become available
	log.Printf("[DEBUG] Waiting for account request (%s) to succeed", requestId)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{organizations.CreateAccountStateInProgress},
		Target:       []string{organizations.CreateAccountStateSucceeded},
		Refresh:      resourceAwsOrganizationsAccountStateRefreshFunc(conn, requestId),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}
	stateResp, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for account request (%s) to become available: %s",
			requestId, stateErr)
	}

	// Store the ID
	accountId := stateResp.(*organizations.CreateAccountStatus).AccountId
	d.SetId(*accountId)

	return resourceAwsOrganizationsAccountRead(d, meta)
}

func resourceAwsOrganizationsAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	describeOpts := &organizations.DescribeAccountInput{
		AccountId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeAccount(describeOpts)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeAccountNotFoundException, "") {
			log.Printf("[WARN] Account does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", resp.Account.Arn)
	d.Set("joined_method", resp.Account.JoinedMethod)
	d.Set("joined_timestamp", resp.Account.JoinedTimestamp)
	d.Set("name", resp.Account.Name)
	d.Set("status", resp.Account.Status)
	return nil
}

func resourceAwsOrganizationsAccountDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Organization member accounts must be deleted from the web console.")
	return nil
}

// resourceAwsOrganizationsAccountStateRefreshFunc returns a resource.StateRefreshFunc
// that is used to watch a CreateAccount request
func resourceAwsOrganizationsAccountStateRefreshFunc(conn *organizations.Organizations, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &organizations.DescribeCreateAccountStatusInput{
			CreateAccountRequestId: aws.String(id),
		}
		resp, err := conn.DescribeCreateAccountStatus(opts)
		if err != nil {
			if isAWSErr(err, organizations.ErrCodeCreateAccountStatusNotFoundException, "") {
				resp = nil
			} else {
				log.Printf("Error on OrganizationAccountStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our account yet. Return an empty state.
			return nil, "", nil
		}

		accountStatus := resp.CreateAccountStatus
		if *accountStatus.State == organizations.CreateAccountStateFailed {
			return nil, *accountStatus.State, fmt.Errorf(*accountStatus.FailureReason)
		}
		return accountStatus, *accountStatus.State, nil
	}
}

func validateAwsOrganizationsAccountEmail(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be a valid email address", value))
	}

	if len(value) < 6 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be less than 6 characters", value))
	}

	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 64 characters", value))
	}

	return
}

func validateAwsOrganizationsAccountRoleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[\w+=,.@-]{1,64}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must consist of uppercase letters, lowercase letters, digits with no spaces, and any of the following characters: =,.@-", value))
	}

	return
}
