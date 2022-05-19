package organizations

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountCreate,
		Read:   resourceAccountRead,
		Update: resourceAccountUpdate,
		Delete: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"close_on_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"create_govcloud": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email": {
				ForceNew: true,
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(6, 64),
					validation.StringMatch(regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`), "must be a valid email address"),
				),
			},
			"govcloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_user_access_to_billing": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{organizations.IAMUserAccessToBillingAllow, organizations.IAMUserAccessToBillingDeny}, true),
			},
			"joined_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"parent_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_MoveAccount.html#organizations-MoveAccount-request-DestinationParentId"),
			},
			"role_name": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]{1,64}$`), "must consist of uppercase letters, lowercase letters, digits with no spaces, and any of the following characters"),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var iamUserAccessToBilling *string

	if v, ok := d.GetOk("iam_user_access_to_billing"); ok {
		iamUserAccessToBilling = aws.String(v.(string))
	}

	var roleName *string

	if v, ok := d.GetOk("role_name"); ok {
		roleName = aws.String(v.(string))
	}

	s, err := createAccount(
		conn,
		d.Get("name").(string),
		d.Get("email").(string),
		iamUserAccessToBilling,
		roleName,
		Tags(tags.IgnoreAWS()),
		d.Get("create_govcloud").(bool),
	)

	if err != nil {
		return fmt.Errorf("error creating AWS Organizations Account (%s): %w", d.Get("name").(string), err)
	}

	output, err := waitAccountCreated(conn, aws.StringValue(s.Id))

	if err != nil {
		return fmt.Errorf("error waiting for AWS Organizations Account (%s) create: %w", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(output.AccountId))

	if v, ok := d.GetOk("parent_id"); ok {
		oldParentAccountID, err := findParentAccountID(conn, d.Id())

		if err != nil {
			return fmt.Errorf("error reading AWS Organizations Account (%s) parent: %w", d.Id(), err)
		}

		if newParentAccountID := v.(string); newParentAccountID != oldParentAccountID {
			input := &organizations.MoveAccountInput{
				AccountId:           aws.String(d.Id()),
				DestinationParentId: aws.String(newParentAccountID),
				SourceParentId:      aws.String(oldParentAccountID),
			}

			log.Printf("[DEBUG] Moving AWS Organizations Account: %s", input)
			if _, err := conn.MoveAccount(input); err != nil {
				return fmt.Errorf("error moving AWS Organizations Account (%s): %w", d.Id(), err)
			}
		}
	}

	return resourceAccountRead(d, meta)
}

func resourceAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	account, err := FindAccountByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AWS Organizations Account does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading AWS Organizations Account (%s): %w", d.Id(), err)
	}

	parentAccountID, err := findParentAccountID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading AWS Organizations Account (%s) parent: %w", d.Id(), err)
	}

	d.Set("arn", account.Arn)
	d.Set("email", account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", aws.TimeValue(account.JoinedTimestamp).Format(time.RFC3339))
	d.Set("name", account.Name)
	d.Set("parent_id", parentAccountID)
	d.Set("status", account.Status)

	s, err := findCreateAccountStatusByID(conn, d.Id())

	if err != nil {
		return names.Error(names.Organizations, "finding", "Create Account Status", d.Id(), err)
	}

	d.Set("govcloud_id", s.GovCloudAccountId)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for AWS Organizations Account (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	if d.HasChange("parent_id") {
		o, n := d.GetChange("parent_id")

		input := &organizations.MoveAccountInput{
			AccountId:           aws.String(d.Id()),
			SourceParentId:      aws.String(o.(string)),
			DestinationParentId: aws.String(n.(string)),
		}

		if _, err := conn.MoveAccount(input); err != nil {
			return fmt.Errorf("error moving AWS Organizations Account (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating AWS Organizations Account (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAccountRead(d, meta)
}

func resourceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	close := d.Get("close_on_deletion").(bool)
	var err error

	if close {
		log.Printf("[DEBUG] Closing AWS Organizations Account: %s", d.Id())
		_, err = conn.CloseAccount(&organizations.CloseAccountInput{
			AccountId: aws.String(d.Id()),
		})
	} else {
		log.Printf("[DEBUG] Removing AWS Organizations Account from organization: %s", d.Id())
		_, err = conn.RemoveAccountFromOrganization(&organizations.RemoveAccountFromOrganizationInput{
			AccountId: aws.String(d.Id()),
		})
	}

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAccountNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting AWS Organizations Account (%s): %w", d.Id(), err)
	}

	if close {
		if _, err := waitAccountDeleted(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for AWS Organizations Account (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}

func createAccount(conn *organizations.Organizations, name, email string, iamUserAccessToBilling, roleName *string, tags []*organizations.Tag, govCloud bool) (*organizations.CreateAccountStatus, error) {
	if govCloud {
		input := &organizations.CreateGovCloudAccountInput{
			AccountName: aws.String(name),
			Email:       aws.String(email),
		}

		if iamUserAccessToBilling != nil {
			input.IamUserAccessToBilling = iamUserAccessToBilling
		}

		if roleName != nil {
			input.RoleName = roleName
		}

		if len(tags) > 0 {
			input.Tags = tags
		}

		log.Printf("[DEBUG] Creating AWS Organizations Account with GovCloud Account: %s", input)
		outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(4*time.Minute,
			func() (interface{}, error) {
				return conn.CreateGovCloudAccount(input)
			},
			organizations.ErrCodeFinalizingOrganizationException,
		)

		if err != nil {
			return nil, err
		}

		return outputRaw.(*organizations.CreateGovCloudAccountOutput).CreateAccountStatus, nil
	}

	input := &organizations.CreateAccountInput{
		AccountName: aws.String(name),
		Email:       aws.String(email),
	}

	if iamUserAccessToBilling != nil {
		input.IamUserAccessToBilling = iamUserAccessToBilling
	}

	if roleName != nil {
		input.RoleName = roleName
	}

	if len(tags) > 0 {
		input.Tags = tags
	}

	log.Printf("[DEBUG] Creating AWS Organizations Account: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(4*time.Minute,
		func() (interface{}, error) {
			return conn.CreateAccount(input)
		},
		organizations.ErrCodeFinalizingOrganizationException,
	)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*organizations.CreateAccountOutput).CreateAccountStatus, nil
}

func findParentAccountID(conn *organizations.Organizations, id string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(id),
	}
	var output []*organizations.Parent

	err := conn.ListParentsPages(input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		output = append(output, page.Parents...)

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if len(output) == 0 || output[0] == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	if count := len(output); count > 1 {
		return "", tfresource.NewTooManyResultsError(count, input)
	}

	return aws.StringValue(output[0].Id), nil
}

func findCreateAccountStatusByID(conn *organizations.Organizations, id string) (*organizations.CreateAccountStatus, error) {
	input := &organizations.DescribeCreateAccountStatusInput{
		CreateAccountRequestId: aws.String(id),
	}

	output, err := conn.DescribeCreateAccountStatus(input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeCreateAccountStatusNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CreateAccountStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CreateAccountStatus, nil
}

func statusCreateAccountState(conn *organizations.Organizations, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCreateAccountStatusByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitAccountCreated(conn *organizations.Organizations, id string) (*organizations.CreateAccountStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{organizations.CreateAccountStateInProgress},
		Target:       []string{organizations.CreateAccountStateSucceeded},
		Refresh:      statusCreateAccountState(conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*organizations.CreateAccountStatus); ok {
		if state := aws.StringValue(output.State); state == organizations.CreateAccountStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func statusAccountStatus(conn *organizations.Organizations, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAccountByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitAccountDeleted(conn *organizations.Organizations, id string) (*organizations.Account, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{organizations.AccountStatusPendingClosure},
		Target:       []string{},
		Refresh:      statusAccountStatus(conn, id),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*organizations.Account); ok {
		return output, err
	}

	return nil, err
}
