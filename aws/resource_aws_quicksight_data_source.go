package aws

import (
	"fmt"
	//	"log"
	//	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsQuickSightDataSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQuickSightDataSourceCreate,
		//		Read:   resourceAwsQuickSightDataSourceRead,
		//		Update: resourceAwsQuickSightDataSourceUpdate,
		//		Delete: resourceAwsQuickSightDataSourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"aws_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"credentials": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credential_pair": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
					},
				},
			},

			"id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"parameters": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amazon_elasticsearch": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"athena": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"work_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"aurora": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"aurora_postgre_sql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// TODO: Extract common
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"aws_iot_analytics": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_set_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"jira": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"maria_db": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"mysql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"postgresql": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"presto": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:     schema.TypeString,
										Required: true,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						// The documentation is not clear on how to pass RDS parameters...
						"redshift": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"s3": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"manifest_file_location": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
												"key": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
											},
										},
									},
								},
							},
						},
						"service_now": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"site_base_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
						"snowflake": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"warehouse": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"spark": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Required: true,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"sql_server": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"teradata": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"host": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
									"port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"twitter": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_rows": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"query": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
					},
				},
			},

			"permissions": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"principal": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"ssl_properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disable_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"tags": tagsSchema(),

			// This will be inferred from the passed `parameters` value
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_connection_properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_connection_arn": {
							Type:         schema.TypeBool,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"creation_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsQuickSightDataSourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn

	awsAccountID := meta.(*AWSClient).accountid
	id := d.Get("id").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountID = v.(string)
	}

	params := &quicksight.CreateDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSourceId: aws.String(id),
	}

	if v := d.Get("credentials"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			credentials := v.(map[string]interface{})

			if v := credentials["credential_pair"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					credentialPairResource := v.(map[string]interface{})
					credentialPair := &quicksight.CredentialPair{}

					if v, ok := credentialPairResource["username"]; ok && v.(string) != "" {
						credentialPair.Username = aws.String(v.(string))
					}

					if v, ok := credentialPairResource["password"]; ok && v.(string) != "" {
						credentialPair.Password = aws.String(v.(string))
					}

					params.Credentials = &quicksight.DataSourceCredentials{
						CredentialPair: credentialPair,
					}
				}
			}
		}
	}

	if v := d.Get("parameters"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			dataSourceParams := v.(map[string]interface{})
			dataSourceParamsResource := &quicksight.DataSourceParameters{}

			if v := dataSourceParams["amazon_elasticsearch"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AmazonElasticsearchParameters{}

					if v, ok := psResource["domain"]; ok && v.(string) != "" {
						ps.Domain = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeAmazonElasticsearch)
					dataSourceParamsResource.AmazonElasticsearchParameters = ps
				}
			}

			if v := dataSourceParams["athena"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AthenaParameters{}

					if v, ok := psResource["work_group"]; ok && v.(string) != "" {
						ps.WorkGroup = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeAthena)
					dataSourceParamsResource.AthenaParameters = ps
				}
			}

			if v := dataSourceParams["aurora"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AuroraParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeAurora)
					dataSourceParamsResource.AuroraParameters = ps
				}
			}

			if v := dataSourceParams["aurora_postgre_sql"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AuroraPostgreSqlParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeAuroraPostgresql)
					dataSourceParamsResource.AuroraPostgreSqlParameters = ps
				}
			}

			if v := dataSourceParams["aws_iot_analytics"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.AwsIotAnalyticsParameters{}

					if v, ok := psResource["data_set_name"]; ok && v.(string) != "" {
						ps.DataSetName = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeAwsIotAnalytics)
					dataSourceParamsResource.AwsIotAnalyticsParameters = ps
				}
			}

			if v := dataSourceParams["jira"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.JiraParameters{}

					if v, ok := psResource["site_base_url"]; ok && v.(string) != "" {
						ps.SiteBaseUrl = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeJira)
					dataSourceParamsResource.JiraParameters = ps
				}
			}

			if v := dataSourceParams["maria_db"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.MariaDbParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeMariadb)
					dataSourceParamsResource.MariaDbParameters = ps
				}
			}

			if v := dataSourceParams["mysql"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.MySqlParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeMysql)
					dataSourceParamsResource.MySqlParameters = ps
				}
			}

			if v := dataSourceParams["postgresql"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.PostgreSqlParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypePostgresql)
					dataSourceParamsResource.PostgreSqlParameters = ps
				}
			}

			if v := dataSourceParams["presto"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.PrestoParameters{}

					if v, ok := psResource["catalog"]; ok && v.(string) != "" {
						ps.Catalog = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypePresto)
					dataSourceParamsResource.PrestoParameters = ps
				}
			}

			if v := dataSourceParams["redshift"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.RedshiftParameters{}

					if v, ok := psResource["cluster_id"]; ok && v.(string) != "" {
						ps.ClusterId = aws.String(v.(string))
					}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeRedshift)
					dataSourceParamsResource.RedshiftParameters = ps
				}
			}

			if v := dataSourceParams["s3"]; v != nil && v.(*schema.Set) != nil {
				s3 := v.(map[string]interface{})

				if v := s3["manifest_file_location"]; v != nil && v.(*schema.Set) != nil {
					for _, v := range (v.(*schema.Set)).List() {
						psResource := v.(map[string]interface{})
						ps := &quicksight.S3Parameters{
							ManifestFileLocation: &quicksight.ManifestFileLocation{},
						}

						if v, ok := psResource["bucket"]; ok && v.(string) != "" {
							ps.ManifestFileLocation.Bucket = aws.String(v.(string))
						}

						if v, ok := psResource["key"]; ok && v.(string) != "" {
							ps.ManifestFileLocation.Key = aws.String(v.(string))
						}

						params.Type = aws.String(quicksight.DataSourceTypeS3)
						dataSourceParamsResource.S3Parameters = ps
					}
				}
			}

			if v := dataSourceParams["service_now"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.ServiceNowParameters{}

					if v, ok := psResource["site_base_url"]; ok && v.(string) != "" {
						ps.SiteBaseUrl = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeServicenow)
					dataSourceParamsResource.ServiceNowParameters = ps
				}
			}

			if v := dataSourceParams["snowflake"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.SnowflakeParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["warehouse"]; ok && v.(string) != "" {
						ps.Warehouse = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeSnowflake)
					dataSourceParamsResource.SnowflakeParameters = ps
				}
			}

			if v := dataSourceParams["spark"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.SparkParameters{}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeSpark)
					dataSourceParamsResource.SparkParameters = ps
				}
			}

			if v := dataSourceParams["sql_server"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.SqlServerParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeSqlserver)
					dataSourceParamsResource.SqlServerParameters = ps
				}
			}

			if v := dataSourceParams["teradata"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.TeradataParameters{}

					if v, ok := psResource["database"]; ok && v.(string) != "" {
						ps.Database = aws.String(v.(string))
					}

					if v, ok := psResource["host"]; ok && v.(string) != "" {
						ps.Host = aws.String(v.(string))
					}

					if v, ok := psResource["port"]; ok && v.(int64) != 0 {
						ps.Port = aws.Int64(v.(int64))
					}

					params.Type = aws.String(quicksight.DataSourceTypeTeradata)
					dataSourceParamsResource.TeradataParameters = ps
				}
			}

			if v := dataSourceParams["twitter"]; v != nil && v.(*schema.Set) != nil {
				for _, v := range (v.(*schema.Set)).List() {
					psResource := v.(map[string]interface{})
					ps := &quicksight.TwitterParameters{}

					if v, ok := psResource["max_rows"]; ok && v.(int64) != 0 {
						ps.MaxRows = aws.Int64(v.(int64))
					}

					if v, ok := psResource["query"]; ok && v.(string) != "" {
						ps.Query = aws.String(v.(string))
					}

					params.Type = aws.String(quicksight.DataSourceTypeTwitter)
					dataSourceParamsResource.TwitterParameters = ps
				}
			}
		}
	}

	d.Set("type", params.Type)

	if v := d.Get("permissions"); v != nil {
		params.Permissions = make([]*quicksight.ResourcePermission, 0)

		for _, v := range v.(*schema.Set).List() {
			permissionResource := v.(map[string]interface{})
			permission := &quicksight.ResourcePermission{
				Actions:   expandStringSet(permissionResource["actions"].(*schema.Set)),
				Principal: aws.String(permissionResource["principal"].(string)),
			}

			params.Permissions = append(params.Permissions, permission)
		}
	}

	if v := d.Get("ssl_properties"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			sslProperties := v.(map[string]interface{})

			if v, present := sslProperties["disable_ssl"]; present {
				params.SslProperties = &quicksight.SslProperties{
					DisableSsl: aws.Bool(v.(bool)),
				}
			}
		}
	}

	if v, exists := d.GetOk("tags"); exists {
		tags := make([]*quicksight.Tag, 0)
		for k, v := range v.(map[string]interface{}) {
			if !tagIgnoredGeneric(k) {
				tags = append(tags, &quicksight.Tag{
					Key:   aws.String(k),
					Value: aws.String(v.(string)),
				})
			}
		}

		params.Tags = tags
	}

	if v := d.Get("vpc_connection_properties"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			vpcConnectionProperties := v.(map[string]interface{})

			if v := vpcConnectionProperties["vpc_connection_arn"]; v != nil && v.(string) != "" {
				params.VpcConnectionProperties = &quicksight.VpcConnectionProperties{
					VpcConnectionArn: aws.String(v.(string)),
				}
			}
		}
	}

	_, err := conn.CreateDataSource(params)
	if err != nil {
		return fmt.Errorf("Error creating QuickSight Data Source: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", awsAccountID, id))

	return nil
	//return resourceAwsQuickSightDataSourceRead(d, meta)
}

//func resourceAwsQuickSightDataSourceRead(d *schema.ResourceData, meta interface{}) error {
//	conn := meta.(*AWSClient).quicksightconn
//
//	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
//	if err != nil {
//		return err
//	}
//
//	descOpts := &quicksight.DescribeDataSourceInput{
//		AwsAccountId: aws.String(awsAccountID),
//		Namespace:    aws.String(namespace),
//		DataSourceName:    aws.String(groupName),
//	}
//
//	resp, err := conn.DescribeDataSource(descOpts)
//	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
//		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
//		d.SetId("")
//		return nil
//	}
//	if err != nil {
//		return fmt.Errorf("Error describing QuickSight DataSource (%s): %s", d.Id(), err)
//	}
//
//	d.Set("arn", resp.DataSource.Arn)
//	d.Set("aws_account_id", awsAccountID)
//	d.Set("group_name", resp.DataSource.DataSourceName)
//	d.Set("description", resp.DataSource.Description)
//	d.Set("namespace", namespace)

//	return nil
//}

//
//func resourceAwsQuickSightDataSourceUpdate(d *schema.ResourceData, meta interface{}) error {
//	conn := meta.(*AWSClient).quicksightconn
//
//	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
//	if err != nil {
//		return err
//	}
//
//	updateOpts := &quicksight.UpdateDataSourceInput{
//		AwsAccountId: aws.String(awsAccountID),
//		Namespace:    aws.String(namespace),
//		DataSourceName:    aws.String(groupName),
//	}
//
//	if v, ok := d.GetOk("description"); ok {
//		updateOpts.Description = aws.String(v.(string))
//	}
//
//	_, err = conn.UpdateDataSource(updateOpts)
//	if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
//		log.Printf("[WARN] QuickSight DataSource %s is already gone", d.Id())
//		d.SetId("")
//		return nil
//	}
//	if err != nil {
//		return fmt.Errorf("Error updating QuickSight DataSource %s: %s", d.Id(), err)
//	}
//
//	return resourceAwsQuickSightDataSourceRead(d, meta)
//}
//
//func resourceAwsQuickSightDataSourceDelete(d *schema.ResourceData, meta interface{}) error {
//	conn := meta.(*AWSClient).quicksightconn
//
//	awsAccountID, namespace, groupName, err := resourceAwsQuickSightDataSourceParseID(d.Id())
//	if err != nil {
//		return err
//	}
//
//	deleteOpts := &quicksight.DeleteDataSourceInput{
//		AwsAccountId: aws.String(awsAccountID),
//		Namespace:    aws.String(namespace),
//		DataSourceName:    aws.String(groupName),
//	}
//
//	if _, err := conn.DeleteDataSource(deleteOpts); err != nil {
//		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
//			return nil
//		}
//		return fmt.Errorf("Error deleting QuickSight DataSource %s: %s", d.Id(), err)
//	}
//
//	return nil
//}

//func resourceAwsQuickSightDataSourceParseID(id string) (string, string, string, error) {
//	parts := strings.SplitN(id, "/", 3)
//	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
//		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
//	}
//	return parts[0], parts[1], parts[2], nil
//}
