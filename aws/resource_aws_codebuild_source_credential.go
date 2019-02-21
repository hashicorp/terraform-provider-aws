package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsCodeBuildSourceCredential() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildSourceCredentialCreate,
		Read:   resourceAwsCodeBuildSourceCredentialRead,
		Delete: resourceAwsCodeBuildSourceCredentialDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codebuild.AuthTypeBasicAuth,
					codebuild.AuthTypePersonalAccessToken,
				}, false),
			},
			"server_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					codebuild.ServerTypeGithub,
					codebuild.ServerTypeBitbucket,
					codebuild.ServerTypeGithubEnterprise,
				}, false),
			},
			"token": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsCodeBuildSourceCredentialCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	authType := d.Get("auth_type").(string)

	createOpts := &codebuild.ImportSourceCredentialsInput{
		AuthType:   aws.String(authType),
		ServerType: aws.String(d.Get("server_type").(string)),
		Token:      aws.String(d.Get("token").(string)),
	}

	if attr, ok := d.GetOk("user_name"); ok && attr.(string) != "" && authType == codebuild.AuthTypeBasicAuth {
		createOpts.Username = aws.String(attr.(string))
	}

	resp, err := conn.ImportSourceCredentials(createOpts)
	if err != nil {
		return fmt.Errorf("Error importing source credentials: %s", err)
	}

	d.SetId(aws.StringValue(resp.Arn))
	d.Set("arn", resp.Arn)

	return nil
}

func resourceAwsCodeBuildSourceCredentialRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	resp, err := conn.ListSourceCredentials(&codebuild.ListSourceCredentialsInput{})
	if err != nil {
		return fmt.Errorf("Error list source credentials: %s", err)
	}

	if len(resp.SourceCredentialsInfos) == 0 {
		log.Printf("[WARN] Source Credentials(%s) is already deleted", d.Id())
		d.SetId("")
		return nil
	}

	resourceNotFound := true
	for _, sourceCredentialsInfo := range resp.SourceCredentialsInfos {
		if d.Id() == aws.StringValue(sourceCredentialsInfo.Arn) {
			d.Set("auth_type", sourceCredentialsInfo.AuthType)
			d.Set("server_type", sourceCredentialsInfo.ServerType)
			resourceNotFound = false
		}
	}

	if resourceNotFound {
		log.Printf("[WARN] Source Credentials(%s) is already deleted", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsCodeBuildSourceCredentialDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	deleteOpts := &codebuild.DeleteSourceCredentialsInput{
		Arn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSourceCredentials(deleteOpts); err != nil {
		if !isAWSErr(err, codebuild.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Source Credentials(%s): %s", d.Id(), err)
	}

	return nil
}
