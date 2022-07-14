package deploy

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceUpdate,
		Delete: resourceAppDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")

				if len(idParts) == 2 {
					return []*schema.ResourceData{d}, nil
				}

				applicationName := d.Id()
				conn := meta.(*conns.AWSClient).DeployConn

				input := &codedeploy.GetApplicationInput{
					ApplicationName: aws.String(applicationName),
				}

				log.Printf("[DEBUG] Reading CodeDeploy Application: %s", input)
				output, err := conn.GetApplication(input)

				if err != nil {
					return []*schema.ResourceData{}, err
				}

				if output == nil || output.Application == nil {
					return []*schema.ResourceData{}, fmt.Errorf("error reading CodeDeploy Application (%s): empty response", applicationName)
				}

				d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.Application.ApplicationId), applicationName))
				d.Set("name", applicationName)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"compute_platform": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codedeploy.ComputePlatform_Values(), false),
				Default:      codedeploy.ComputePlatformServer,
			},
			"github_account_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"linked_to_github": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	application := d.Get("name").(string)
	computePlatform := d.Get("compute_platform").(string)
	log.Printf("[DEBUG] Creating CodeDeploy application %s", application)

	resp, err := conn.CreateApplication(&codedeploy.CreateApplicationInput{
		ApplicationName: aws.String(application),
		ComputePlatform: aws.String(computePlatform),
		Tags:            Tags(tags.IgnoreAWS()),
	})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] CodeDeploy application %s created", *resp.ApplicationId)

	// Despite giving the application a unique ID, AWS doesn't actually use
	// it in API calls. Use it and the app name to identify the resource in
	// the state file. This allows us to reliably detect both when the TF
	// config file changes and when the user deletes the app without removing
	// it first from the TF config.
	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(resp.ApplicationId), application))

	return resourceAppRead(d, meta)
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	application := resourceAppParseID(d.Id())
	name := d.Get("name").(string)
	if name != "" && application != name {
		application = name
	}
	log.Printf("[DEBUG] Reading CodeDeploy application %s", application)
	resp, err := conn.GetApplication(&codedeploy.GetApplicationInput{
		ApplicationName: aws.String(application),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codedeploy.ErrCodeApplicationDoesNotExistException) {
		log.Printf("[WARN] CodeDeploy Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("finding CodeDeploy Application (%s): %w", d.Id(), err)
	}

	app := resp.Application
	appName := aws.StringValue(app.ApplicationName)

	if !strings.Contains(d.Id(), appName) {
		d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(app.ApplicationId), appName))
	}

	appArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("application:%s", appName),
	}.String()

	d.Set("arn", appArn)
	d.Set("application_id", app.ApplicationId)
	d.Set("compute_platform", app.ComputePlatform)
	d.Set("name", appName)
	d.Set("github_account_name", app.GitHubAccountName)
	d.Set("linked_to_github", app.LinkedToGitHub)

	tags, err := ListTags(conn, appArn)

	if err != nil {
		return fmt.Errorf("error listing tags for CodeDeploy application (%s): %w", d.Id(), err)
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

func resourceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn

	if d.HasChange("name") {
		o, n := d.GetChange("name")

		_, err := conn.UpdateApplication(&codedeploy.UpdateApplicationInput{
			ApplicationName:    aws.String(o.(string)),
			NewApplicationName: aws.String(n.(string)),
		})

		if err != nil {
			return fmt.Errorf("error updating CodeDeploy Application (%s) name: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating CodeDeploy Application (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAppRead(d, meta)
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn

	_, err := conn.DeleteApplication(&codedeploy.DeleteApplicationInput{
		ApplicationName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, codedeploy.ErrCodeApplicationDoesNotExistException) {
			return nil
		}

		log.Printf("[ERROR] Error deleting CodeDeploy application: %s", err)
		return err
	}

	return nil
}

func resourceAppParseID(id string) string {
	parts := strings.SplitN(id, ":", 2)
	// We currently omit the application ID as it is not currently used anywhere
	return parts[1]
}
