package sagemaker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCodeRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceCodeRepositoryCreate,
		Read:   resourceCodeRepositoryRead,
		Update: resourceCodeRepositoryUpdate,
		Delete: resourceCodeRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code_repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"git_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_url": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"branch": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCodeRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("code_repository_name").(string)

	input := &sagemaker.CreateCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
		GitConfig:          expandCodeRepositoryGitConfig(d.Get("git_config").([]interface{})),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] sagemaker code repository create config: %#v", *input)
	_, err := conn.CreateCodeRepository(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker code repository: %w", err)
	}

	d.SetId(name)

	return resourceCodeRepositoryRead(d, meta)
}

func resourceCodeRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	codeRepository, err := FindCodeRepositoryByName(conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "Cannot find CodeRepository") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker code repository (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker code repository (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(codeRepository.CodeRepositoryArn)
	d.Set("code_repository_name", codeRepository.CodeRepositoryName)
	d.Set("arn", arn)

	if err := d.Set("git_config", flattenCodeRepositoryGitConfig(codeRepository.GitConfig)); err != nil {
		return fmt.Errorf("error setting git_config for sagemaker code repository (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Code Repository (%s): %w", d.Id(), err)
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

func resourceCodeRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Code Repository (%s) tags: %w", d.Id(), err)
		}
	}

	if d.HasChange("git_config") {
		input := &sagemaker.UpdateCodeRepositoryInput{
			CodeRepositoryName: aws.String(d.Id()),
			GitConfig:          expandCodeRepositoryUpdateGitConfig(d.Get("git_config").([]interface{})),
		}

		log.Printf("[DEBUG] sagemaker code repository update config: %#v", *input)
		_, err := conn.UpdateCodeRepository(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker code repository: %w", err)
		}
	}

	return resourceCodeRepositoryRead(d, meta)
}

func resourceCodeRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteCodeRepositoryInput{
		CodeRepositoryName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteCodeRepository(input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "Cannot find CodeRepository") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker code repository (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCodeRepositoryGitConfig(l []interface{}) *sagemaker.GitConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfig{
		RepositoryUrl: aws.String(m["repository_url"].(string)),
	}

	if v, ok := m["branch"].(string); ok && v != "" {
		config.Branch = aws.String(v)
	}

	if v, ok := m["secret_arn"].(string); ok && v != "" {
		config.SecretArn = aws.String(v)
	}

	return config
}

func flattenCodeRepositoryGitConfig(config *sagemaker.GitConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"repository_url": aws.StringValue(config.RepositoryUrl),
	}

	if config.Branch != nil {
		m["branch"] = aws.StringValue(config.Branch)
	}

	if config.SecretArn != nil {
		m["secret_arn"] = aws.StringValue(config.SecretArn)
	}

	return []map[string]interface{}{m}
}

func expandCodeRepositoryUpdateGitConfig(l []interface{}) *sagemaker.GitConfigForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfigForUpdate{
		SecretArn: aws.String(m["secret_arn"].(string)),
	}

	return config
}
