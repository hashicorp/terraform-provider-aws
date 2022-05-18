package serverlessrepo

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceApplication() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceApplicationRead,

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required_capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_code_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServerlessRepoConn

	applicationID := d.Get("application_id").(string)
	semanticVersion := d.Get("semantic_version").(string)

	output, err := findApplication(conn, applicationID, semanticVersion)
	if err != nil {
		descriptor := applicationID
		if semanticVersion != "" {
			descriptor += fmt.Sprintf(", version %s", semanticVersion)
		}
		return fmt.Errorf("error getting Serverless Application Repository application (%s): %w", descriptor, err)
	}

	d.SetId(applicationID)
	d.Set("name", output.Name)
	d.Set("semantic_version", output.Version.SemanticVersion)
	d.Set("source_code_url", output.Version.SourceCodeUrl)
	d.Set("template_url", output.Version.TemplateUrl)
	if err = d.Set("required_capabilities", flex.FlattenStringSet(output.Version.RequiredCapabilities)); err != nil {
		return fmt.Errorf("failed to set required_capabilities: %w", err)
	}

	return nil
}
