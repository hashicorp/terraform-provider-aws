package elasticache

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserGroupAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserGroupAssociationCreate,
		Read:   resourceUserGroupAssociationRead,
		Delete: resourceUserGroupAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserGroupAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:  aws.String(d.Get("user_group_id").(string)),
		UserIdsToAdd: aws.StringSlice([]string{d.Get("user_id").(string)}),
	}

	id := userGroupAssociationID(d.Get("user_group_id").(string), d.Get("user_id").(string))

	_, err := tfresource.RetryWhenNotFound(30*time.Second, func() (interface{}, error) {
		return conn.ModifyUserGroup(input)
	})

	if err != nil {
		return fmt.Errorf("creating ElastiCache User Group Association (%q): %w", id, err)
	}

	d.SetId(id)

	stateConf := &resource.StateChangeConf{
		Pending:        []string{"modifying", ""},
		Target:         []string{"active"},
		Refresh:        resourceUserGroupStateRefreshFunc(d.Get("user_group_id").(string), conn),
		Timeout:        d.Timeout(schema.TimeoutCreate),
		MinTimeout:     2 * time.Second,
		NotFoundChecks: 5,
		Delay:          10 * time.Second,
	}

	log.Printf("[INFO] Waiting for ElastiCache User Group (%s) to be available", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("creating ElastiCache User Group Association (%q): %w", d.Id(), err)
	}

	return resourceUserGroupAssociationRead(d, meta)
}

func resourceUserGroupAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	groupID, userID, err := UserGroupAssociationParseID(d.Id())
	if err != nil {
		return fmt.Errorf("reading ElastiCache User Group Association (%s): %w", d.Id(), err)
	}

	output, err := FindUserGroupByID(conn, groupID)
	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault)) {
		d.SetId("")
		log.Printf("[DEBUG] ElastiCache User Group Association (%s) not found", d.Id())
		return nil
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
		return fmt.Errorf("describing ElastiCache User Group (%s): %w", d.Id(), err)
	}

	gotUserID := ""
	for _, v := range output.UserIds {
		if aws.StringValue(v) == userID {
			gotUserID = aws.StringValue(v)
			break
		}
	}

	if !d.IsNewResource() && gotUserID == "" {
		d.SetId("")
		log.Printf("[DEBUG] ElastiCache User Group Association (%s) not found", d.Id())
		return nil
	}

	if gotUserID == "" {
		return fmt.Errorf("reading ElastiCache User Group Association, user ID (%s) not associated with user group (%s)", userID, groupID)
	}

	d.Set("user_id", gotUserID)
	d.Set("user_group_id", groupID)

	return nil
}

func resourceUserGroupAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:     aws.String(d.Get("user_group_id").(string)),
		UserIdsToRemove: aws.StringSlice([]string{d.Get("user_id").(string)}),
	}

	_, err := conn.ModifyUserGroup(input)
	if err != nil && !tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "not a member") {
		return fmt.Errorf("deleting ElastiCache User Group Association (%q): %w", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"active"},
		Refresh:    resourceUserGroupStateRefreshFunc(d.Get("user_group_id").(string), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	log.Printf("[INFO] Waiting for ElastiCache User Group (%s) to be available", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("waiting for ElastiCache User Group Association delete (%q): %w", d.Id(), err)
	}

	return nil
}

func userGroupAssociationID(userGroupID, userID string) string {
	parts := []string{userGroupID, userID}
	id := strings.Join(parts, ",")
	return id
}

func UserGroupAssociationParseID(id string) (string, string, error) {
	parts := strings.Split(id, ",")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ElastiCache User Group Association ID (%q), expected '<user group ID>,<user ID>'", id)
}
