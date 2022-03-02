package cognitoidp

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUserPools() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserPoolsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserPoolsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	output, err := findUserPoolDescriptionTypes(conn)

	if err != nil {
		return fmt.Errorf("error reading Cognito User Pools: %w", err)
	}

	name := d.Get("name").(string)
	var arns, userPoolIDs []string

	for _, v := range output {
		if name != aws.StringValue(v.Name) {
			continue
		}

		userPoolID := aws.StringValue(v.Id)
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   cognitoidentityprovider.ServiceName,
			Region:    meta.(*conns.AWSClient).Region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  fmt.Sprintf("userpool/%s", userPoolID),
		}.String()

		userPoolIDs = append(userPoolIDs, userPoolID)
		arns = append(arns, arn)
	}

	d.SetId(name)
	d.Set("ids", userPoolIDs)
	d.Set("arns", arns)

	return nil
}

func findUserPoolDescriptionTypes(conn *cognitoidentityprovider.CognitoIdentityProvider) ([]*cognitoidentityprovider.UserPoolDescriptionType, error) {
	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(60),
	}
	var output []*cognitoidentityprovider.UserPoolDescriptionType

	err := conn.ListUserPoolsPages(input, func(page *cognitoidentityprovider.ListUserPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.UserPools {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
