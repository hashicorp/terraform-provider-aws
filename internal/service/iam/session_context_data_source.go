package iam

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceSessionContext() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSessionContextRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"issuer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"issuer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSessionContextRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	arn := d.Get("arn").(string)

	d.SetId(arn)

	roleName := ""
	sessionName := ""
	var err error

	if roleName, sessionName = RoleNameSessionFromARN(arn); roleName == "" {
		d.Set("issuer_arn", arn)
		d.Set("issuer_id", "")
		d.Set("issuer_name", "")
		d.Set("session_name", "")

		return nil
	}

	var role *iam.Role

	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		role, err = FindRoleByName(conn, roleName)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		role, err = FindRoleByName(conn, roleName)
	}

	if err != nil {
		return fmt.Errorf("unable to get role (%s): %w", roleName, err)
	}

	if role == nil || role.Arn == nil {
		return fmt.Errorf("empty role returned (%s)", roleName)
	}

	d.Set("issuer_arn", role.Arn)
	d.Set("issuer_id", role.RoleId)
	d.Set("issuer_name", roleName)
	d.Set("session_name", sessionName)

	return nil
}

// RoleNameSessionFromARN returns the role and session names in an ARN if any.
// Otherwise, it returns empty strings.
func RoleNameSessionFromARN(rawARN string) (string, string) {
	parsedARN, err := arn.Parse(rawARN)

	if err != nil {
		return "", ""
	}

	reAssume := regexp.MustCompile(`^assumed-role/.{1,}/.{2,}`)

	if !reAssume.MatchString(parsedARN.Resource) || parsedARN.Service != "sts" {
		return "", ""
	}

	parts := strings.Split(parsedARN.Resource, "/")

	if len(parts) < 3 {
		return "", ""
	}

	return parts[len(parts)-2], parts[len(parts)-1]
}
