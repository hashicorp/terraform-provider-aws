package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"strings"
)

func resourceAwsCodeArtifactRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactRepositoryCreate,
		Read:   resourceAwsCodeArtifactRepositoryRead,
		Update: resourceAwsCodeArtifactRepositoryUpdate,
		Delete: resourceAwsCodeArtifactRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"upstreams": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"administrator_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeArtifactRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Repository")

	params := &codeartifact.CreateRepositoryInput{
		Repository: aws.String(d.Get("repository").(string)),
		Domain:     aws.String(d.Get("domain").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("upstreams"); ok {
		params.Upstreams = expandCodeArtifactUpstreams(v.([]interface{}))
	}

	res, err := conn.CreateRepository(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Repository: %w", err)
	}

	repo := res.Repository
	d.SetId(fmt.Sprintf("%s:%s:%s", aws.StringValue(repo.DomainOwner),
		aws.StringValue(repo.DomainName), aws.StringValue(repo.Name)))

	return resourceAwsCodeArtifactRepositoryRead(d, meta)
}

func resourceAwsCodeArtifactRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Updating CodeArtifact Repository")

	params := &codeartifact.UpdateRepositoryInput{
		Repository:  aws.String(d.Get("repository").(string)),
		Domain:      aws.String(d.Get("domain").(string)),
		DomainOwner: aws.String(d.Get("domain_owner").(string)),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			params.Description = aws.String(v.(string))
		}
	}

	if d.HasChange("upstreams") {
		if v, ok := d.GetOk("upstreams"); ok {
			params.Upstreams = expandCodeArtifactUpstreams(v.([]interface{}))
		}
	}

	_, err := conn.UpdateRepository(params)
	if err != nil {
		return fmt.Errorf("error updating CodeArtifact Repository: %w", err)
	}

	return resourceAwsCodeArtifactRepositoryRead(d, meta)
}

func resourceAwsCodeArtifactRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn

	log.Printf("[DEBUG] Reading CodeArtifact Repository: %s", d.Id())

	owner, domain, repo, err := decodeCodeArtifactRepositoryID(d.Id())
	if err != nil {
		return err
	}
	sm, err := conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Repository %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading CodeArtifact Repository (%s): %w", d.Id(), err)
	}

	d.Set("repository", sm.Repository.Name)
	d.Set("arn", sm.Repository.Arn)
	d.Set("domain_owner", sm.Repository.DomainOwner)
	d.Set("domain", sm.Repository.DomainName)
	d.Set("administrator_account", sm.Repository.AdministratorAccount)
	d.Set("description", sm.Repository.Description)

	if sm.Repository.Upstreams != nil {
		if err := d.Set("upstreams", flattenCodeArtifactUpstreams(sm.Repository.Upstreams)); err != nil {
			return fmt.Errorf("[WARN] Error setting upstreams: %s", err)
		}
	}

	if sm.Repository.ExternalConnections != nil {
		if err := d.Set("external_connections", flattenCodeArtifactUpstreams(sm.Repository.Upstreams)); err != nil {
			return fmt.Errorf("[WARN] Error setting upstreams: %s", err)
		}
	}

	return nil
}

func resourceAwsCodeArtifactRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Repository: %s", d.Id())

	owner, domain, repo, err := decodeCodeArtifactRepositoryID(d.Id())
	if err != nil {
		return err
	}
	input := &codeartifact.DeleteRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	}

	_, err = conn.DeleteRepository(input)

	if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Repository (%s): %w", d.Id(), err)
	}

	return nil
}

func expandCodeArtifactUpstreams(l []interface{}) []*codeartifact.UpstreamRepository {
	upstreams := []*codeartifact.UpstreamRepository{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		upstream := &codeartifact.UpstreamRepository{
			RepositoryName: aws.String(m["repository_name"].(string)),
		}

		upstreams = append(upstreams, upstream)
	}

	return upstreams
}

func flattenCodeArtifactUpstreams(upstreams []*codeartifact.UpstreamRepositoryInfo) []interface{} {
	if len(upstreams) == 0 {
		return nil
	}

	var ls []interface{}

	for _, upstream := range upstreams {
		m := map[string]interface{}{
			"repository_name": aws.StringValue(upstream.RepositoryName),
		}

		ls = append(ls, m)
	}

	return ls
}

func decodeCodeArtifactRepositoryID(id string) (string, string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format DomainOwner:DomainName:RepositoryName, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}
