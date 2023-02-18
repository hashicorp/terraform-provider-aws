package opensearch

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceEngineVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineListVersions,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"preferred_versions"},
			},
			"preferred_versions": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"version"},
			},
		},
	}
}

func dataSourceEngineListVersions(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn()

	input := &opensearchservice.ListVersionsInput{}

	var obtainedVersions []string
	err := conn.ListVersionsPagesWithContext(ctx, input, func(lvo *opensearchservice.ListVersionsOutput, lastPage bool) bool {
		for _, version := range lvo.Versions {
			if version == nil {
				continue
			}
			if *version == "" {
				continue
			}
			obtainedVersions = append(obtainedVersions, *version)
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch engine version: %s", err)
	}

	if requestVersion, ok := d.GetOk("version"); ok {
		for _, version := range obtainedVersions {
			if requestVersion.(string) == version {
				d.Set("version", version)
				d.SetId(version)
				return diags
			}
		}
	}
	if preferredVersions, ok := d.GetOk("preferred_versions"); ok {
		availableVersions := make(map[string]interface{})
		for _, version := range obtainedVersions {
			availableVersions[version] = nil
		}
		for _, preferredVersion := range preferredVersions.([]interface{}) {
			version := preferredVersion.(string)
			if _, ok := availableVersions[version]; ok {
				d.Set("version", version)
				d.SetId(version)
				return diags
			}
		}
	}

	sdkdiag.AppendErrorf(diags, "no OpenSearch version match the criterias")
	return diags
}
