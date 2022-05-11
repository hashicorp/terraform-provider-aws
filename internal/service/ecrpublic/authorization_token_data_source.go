package ecrpublic

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAuthorizationToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAuthorizationTokenRead,

		Schema: map[string]*schema.Schema{
			"authorization_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthorizationTokenRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRPublicConn
	params := &ecrpublic.GetAuthorizationTokenInput{}

	out, err := conn.GetAuthorizationToken(params)

	if err != nil {
		return fmt.Errorf("error getting Public ECR authorization token: %w", err)
	}

	authorizationData := out.AuthorizationData
	authorizationToken := aws.StringValue(authorizationData.AuthorizationToken)
	expiresAt := aws.TimeValue(authorizationData.ExpiresAt).Format(time.RFC3339)
	authBytes, err := base64.URLEncoding.DecodeString(authorizationToken)
	if err != nil {
		return fmt.Errorf("error decoding Public ECR authorization token: %w", err)
	}

	basicAuthorization := strings.Split(string(authBytes), ":")
	if len(basicAuthorization) != 2 {
		return fmt.Errorf("unknown Public ECR authorization token format")
	}

	userName := basicAuthorization[0]
	password := basicAuthorization[1]
	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("authorization_token", authorizationToken)
	d.Set("expires_at", expiresAt)
	d.Set("user_name", userName)
	d.Set("password", password)

	return nil
}
