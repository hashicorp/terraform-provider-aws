package aws

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDbEngineVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDbEngineVersionRead,

		Schema: map[string]*schema.Schema{
			//selection criteria
			"engine": {
				Type:     schema.TypeString,
				Required: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			//Computed values returned
			"db_engine_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_engine_version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exportable_log_types": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supported_engine_modes": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supported_feature_names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supports_log_exports_to_cloudwatch_logs": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_read_replica": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDbEngineVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	req := &rds.DescribeDBEngineVersionsInput{
		Engine: aws.String(d.Get("engine").(string)),
	}
	if v, ok := d.GetOk("engine_version"); ok {
		req.EngineVersion = aws.String(v.(string))
	}

	resp, err := conn.DescribeDBEngineVersions(req)
	if err != nil {
		return err
	}

	if len(resp.DBEngineVersions) < 1 {
		return errors.New("Your query returned no results. Please change your search criteria and try again.")
	}

	var version *rds.DBEngineVersion
	if len(resp.DBEngineVersions) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] aws_db_engine_version - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			version = mostRecentDbEngineVersion(resp.DBEngineVersions)
		} else {
			return errors.New("Your query returned more than one result. Please try a more specific search criteria.")
		}
	} else {
		version = resp.DBEngineVersions[0]
	}

	d.SetId(aws.StringValue(version.EngineVersion))
	d.Set("db_engine_description", version.DBEngineDescription)
	d.Set("db_engine_version_description", version.DBEngineVersionDescription)
	d.Set("db_parameter_group_family", version.DBParameterGroupFamily)
	d.Set("engine_version", version.EngineVersion)
	d.Set("exportable_log_types", version.ExportableLogTypes)
	d.Set("supported_engine_modes", version.SupportedEngineModes)
	d.Set("supported_feature_names", version.SupportedFeatureNames)
	d.Set("supports_log_exports_to_cloudwatch_logs", version.SupportsLogExportsToCloudwatchLogs)
	d.Set("supports_read_replica", version.SupportsReadReplica)

	return nil
}

func mostRecentDbEngineVersion(versions []*rds.DBEngineVersion) *rds.DBEngineVersion {
	sortedVersions := versions
	sort.Slice(sortedVersions, func(i, j int) bool {
		iEngineVersion, err := gversion.NewVersion(aws.StringValue(sortedVersions[i].EngineVersion))
		if err != nil {
			panic(fmt.Sprintf("error converting (%s) to go-version: %s", aws.StringValue(sortedVersions[i].EngineVersion), err))
		}

		jEngineVersion, err := gversion.NewVersion(aws.StringValue(sortedVersions[j].EngineVersion))
		if err != nil {
			panic(fmt.Sprintf("error converting (%s) to go-version: %s", aws.StringValue(sortedVersions[j].EngineVersion), err))
		}

		return iEngineVersion.GreaterThan(jEngineVersion)
	})
	return sortedVersions[0]
}
