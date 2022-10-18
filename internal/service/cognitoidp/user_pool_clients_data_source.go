package cognitoidp

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUserPoolClients() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceuserPoolClientsRead,
		Schema: map[string]*schema.Schema{
			"client_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"client_names": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceuserPoolClientsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	userPoolID := d.Get("user_pool_id").(string)
	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolID),
	}

	var clientIDs []string
	var clientNames []string
	err := conn.ListUserPoolClientsPages(input, func(page *cognitoidentityprovider.ListUserPoolClientsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.UserPoolClients {
			if v == nil {
				continue
			}

			clientNames = append(clientNames, aws.StringValue(v.ClientName))
			clientIDs = append(clientIDs, aws.StringValue(v.ClientId))
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error getting user pool clients: %w", err)
	}

	d.SetId(userPoolID)
	d.Set("client_ids", clientIDs)
	d.Set("client_names", clientNames)

	return nil
}
