package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUserGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupAssociationCreate,
		ReadWithoutTimeout:   resourceUserGroupAssociationRead,
		DeleteWithoutTimeout: resourceUserGroupAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceUserGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:  aws.String(d.Get("user_group_id").(string)),
		UserIdsToAdd: aws.StringSlice([]string{d.Get("user_id").(string)}),
	}

	id := userGroupAssociationID(d.Get("user_group_id").(string), d.Get("user_id").(string))

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 10*time.Minute, func() (interface{}, error) {
		return tfresource.RetryWhenNotFound(ctx, 30*time.Second, func() (interface{}, error) {
			return conn.ModifyUserGroupWithContext(ctx, input)
		})
	}, elasticache.ErrCodeInvalidUserGroupStateFault)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User Group Association (%q): %s", id, err)
	}

	d.SetId(id)

	stateConf := &resource.StateChangeConf{
		Pending:        []string{"modifying", ""},
		Target:         []string{"active"},
		Refresh:        resourceUserGroupStateRefreshFunc(ctx, d.Get("user_group_id").(string), conn),
		Timeout:        d.Timeout(schema.TimeoutCreate),
		MinTimeout:     2 * time.Second,
		NotFoundChecks: 5,
		Delay:          10 * time.Second,
	}

	log.Printf("[INFO] Waiting for ElastiCache User Group (%s) to be available", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache User Group Association (%q): %s", d.Id(), err)
	}

	return append(diags, resourceUserGroupAssociationRead(ctx, d, meta)...)
}

func resourceUserGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	groupID, userID, err := UserGroupAssociationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache User Group Association (%s): %s", d.Id(), err)
	}

	output, err := FindUserGroupByID(ctx, conn, groupID)
	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault)) {
		d.SetId("")
		log.Printf("[DEBUG] ElastiCache User Group Association (%s) not found", d.Id())
		return diags
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
		return sdkdiag.AppendErrorf(diags, "describing ElastiCache User Group (%s): %s", d.Id(), err)
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
		return diags
	}

	if gotUserID == "" {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache User Group Association, user ID (%s) not associated with user group (%s)", userID, groupID)
	}

	d.Set("user_id", gotUserID)
	d.Set("user_group_id", groupID)

	return diags
}

func resourceUserGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	input := &elasticache.ModifyUserGroupInput{
		UserGroupId:     aws.String(d.Get("user_group_id").(string)),
		UserIdsToRemove: aws.StringSlice([]string{d.Get("user_id").(string)}),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 10*time.Minute, func() (interface{}, error) {
		return conn.ModifyUserGroupWithContext(ctx, input)
	}, elasticache.ErrCodeInvalidUserGroupStateFault)

	if err != nil && !tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "not a member") {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache User Group Association (%q): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"active"},
		Refresh:    resourceUserGroupStateRefreshFunc(ctx, d.Get("user_group_id").(string), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	log.Printf("[INFO] Waiting for ElastiCache User Group (%s) to be available", d.Id())
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ElastiCache User Group Association delete (%q): %s", d.Id(), err)
	}

	return diags
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
