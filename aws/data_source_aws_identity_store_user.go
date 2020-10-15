package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsIdentityStoreUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIdentityStoreUserRead,

		Schema: map[string]*schema.Schema{
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},

			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"user_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_id"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}]+$`), "must match [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]"),
				),
			},
		},
	}
}

func dataSourceAwsIdentityStoreUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).identitystoreconn

	identityStoreID := d.Get("identity_store_id").(string)
	userID := d.Get("user_id").(string)
	userName := d.Get("user_name").(string)

	if userID != "" {
		log.Printf("[DEBUG] Reading AWS Identity Store User")
		resp, err := conn.DescribeUser(&identitystore.DescribeUserInput{
			IdentityStoreId: aws.String(identityStoreID),
			UserId:          aws.String(userID),
		})
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == identitystore.ErrCodeResourceNotFoundException {
				log.Printf("[DEBUG] AWS Identity Store User not found with the id %v", userID)
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error getting AWS Identity Store User: %s", err)
		}
		d.SetId(userID)
		d.Set("user_name", resp.UserName)
	} else if userName != "" {
		log.Printf("[DEBUG] Reading AWS Identity Store Users")
		req := &identitystore.ListUsersInput{
			IdentityStoreId: aws.String(identityStoreID),
			Filters: []*identitystore.Filter{
				{
					AttributePath:  aws.String("UserName"),
					AttributeValue: aws.String(userName),
				},
			},
		}
		users := []*identitystore.User{}
		err := conn.ListUsersPages(req, func(page *identitystore.ListUsersOutput, lastPage bool) bool {
			if page != nil && page.Users != nil && len(page.Users) != 0 {
				users = append(users, page.Users...)
			}
			return !lastPage
		})
		if err != nil {
			return fmt.Errorf("Error getting AWS Identity Store Users: %s", err)
		}
		if len(users) == 0 {
			log.Printf("[DEBUG] No AWS Identity Store Users found")
			d.SetId("")
			return nil
		}
		if len(users) > 1 {
			return fmt.Errorf("Found multiple AWS Identity Store Users with the UserName %v. Not sure which one to use. %s", userName, users)
		}
		user := users[0]
		d.SetId(aws.StringValue(user.UserId))
		d.Set("user_id", user.UserId)
	} else {
		return fmt.Errorf("One of user_id or user_name is required")
	}

	return nil
}
