package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
)

type pipelineObjectTestCase struct {
	Attrs    map[string]interface{}
	Expected *datapipeline.PipelineObject
}

func TestBuildDefaultPipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"schedule_type":          "cron",
				"failure_and_rerun_mode": "CASCADE",
				"pipeline_log_uri":       "s3://bucket-name/key-name-prefix/",
				"role":                   "DataPipelineDefaultRole",
				"resource_role":          "DataPipelineDefaultResourceRole",
				"schedule":               "myDefaultSchedule",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("Default"),
				Name: aws.String("Default"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("Default"),
					},
					{
						Key:         aws.String("scheduleType"),
						StringValue: aws.String("cron"),
					},
					{
						Key:         aws.String("failureAndRerunMode"),
						StringValue: aws.String("CASCADE"),
					},
					{
						Key:         aws.String("pipelineLogUri"),
						StringValue: aws.String("s3://bucket-name/key-name-prefix/"),
					},
					{
						Key:         aws.String("role"),
						StringValue: aws.String("DataPipelineDefaultRole"),
					},
					{
						Key:         aws.String("resourceRole"),
						StringValue: aws.String("DataPipelineDefaultResourceRole"),
					},
					{
						Key:      aws.String("schedule"),
						RefValue: aws.String("myDefaultSchedule"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildDefaultPipelineObject(testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestBuildEc2ResourcePipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"id":                          "bar",
				"name":                        "boo",
				"associate_public_ip_address": true,
				"image_id":                    "i-xxxxxxxxxxxxxx",
				"instance_type":               "t2.micro",
				"max_active_instances":        10,
				"maximum_retries":             5,
				"security_group_ids": []string{
					"sg-12345678",
					"sg-23456789",
				},
				"subnet_id": "subnet-12345678",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("Ec2Resource"),
					},
					{
						Key:         aws.String("associatePublicIpAddress"),
						StringValue: aws.String("true"),
					},
					{
						Key:         aws.String("imageId"),
						StringValue: aws.String("i-xxxxxxxxxxxxxx"),
					},
					{
						Key:         aws.String("instanceType"),
						StringValue: aws.String("t2.micro"),
					},
					{
						Key:         aws.String("maxActiveInstances"),
						StringValue: aws.String("10"),
					},
					{
						Key:         aws.String("maximumRetries"),
						StringValue: aws.String("5"),
					},
					{
						Key:         aws.String("securityGroupIds"),
						StringValue: aws.String("sg-12345678"),
					},
					{
						Key:         aws.String("securityGroupIds"),
						StringValue: aws.String("sg-23456789"),
					},
					{
						Key:         aws.String("subnetId"),
						StringValue: aws.String("subnet-12345678"),
					},
				},
			},
		},
		{
			map[string]interface{}{
				"id":                          "bar",
				"name":                        "boo",
				"associate_public_ip_address": true,
				"availability_zone":           "ap-northeast-1a",
				"image_id":                    "i-xxxxxxxxxxxxxx",
				"instance_type":               "m4.large",
				"max_active_instances":        10,
				"maximum_retries":             5,
				"security_groups": []string{
					"default",
					"test-group",
				},
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("Ec2Resource"),
					},
					{
						Key:         aws.String("associatePublicIpAddress"),
						StringValue: aws.String("true"),
					},
					{
						Key:         aws.String("availabilityZone"),
						StringValue: aws.String("ap-northeast-1a"),
					},
					{
						Key:         aws.String("imageId"),
						StringValue: aws.String("i-xxxxxxxxxxxxxx"),
					},
					{
						Key:         aws.String("instanceType"),
						StringValue: aws.String("m4.large"),
					},
					{
						Key:         aws.String("maxActiveInstances"),
						StringValue: aws.String("10"),
					},
					{
						Key:         aws.String("maximumRetries"),
						StringValue: aws.String("5"),
					},
					{
						Key:         aws.String("securityGroups"),
						StringValue: aws.String("default"),
					},
					{
						Key:         aws.String("securityGroups"),
						StringValue: aws.String("test-group"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildCommonPipelineObject("Ec2Resource", testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestBuildSchedulePipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"id":              "bar",
				"name":            "boo",
				"period":          "1 hour",
				"start_date_time": "2019-01-01T00:00:00",
				"end_date_time":   "2019-09-01T00:00:00",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("Schedule"),
					},
					{
						Key:         aws.String("period"),
						StringValue: aws.String("1 hour"),
					},
					{
						Key:         aws.String("startDateTime"),
						StringValue: aws.String("2019-01-01T00:00:00"),
					},
					{
						Key:         aws.String("endDateTime"),
						StringValue: aws.String("2019-09-01T00:00:00"),
					},
				},
			},
		},
		{
			map[string]interface{}{
				"id":          "bar",
				"name":        "boo",
				"occurrences": 1,
				"period":      "1 Day",
				"start_at":    "FIRST_ACTIVATION_DATE_TIME",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("Schedule"),
					},
					{
						Key:         aws.String("period"),
						StringValue: aws.String("1 Day"),
					},
					{
						Key:         aws.String("startAt"),
						StringValue: aws.String("FIRST_ACTIVATION_DATE_TIME"),
					},
					{
						Key:         aws.String("occurrences"),
						StringValue: aws.String("1"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildCommonPipelineObject("Schedule", testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestBuildS3DataNodePipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"id":             "bar",
				"name":           "boo",
				"compression":    "none",
				"directory_path": "hogehoge",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("S3DataNode"),
					},
					{
						Key:         aws.String("compression"),
						StringValue: aws.String("none"),
					},
					{
						Key:         aws.String("directoryPath"),
						StringValue: aws.String("hogehoge"),
					},
				},
			},
		},
		{
			map[string]interface{}{
				"id":             "bar",
				"name":           "boo",
				"compression":    "gzip",
				"directory_path": "hogehoge",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("S3DataNode"),
					},
					{
						Key:         aws.String("compression"),
						StringValue: aws.String("gzip"),
					},
					{
						Key:         aws.String("directoryPath"),
						StringValue: aws.String("hogehoge"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildCommonPipelineObject("S3DataNode", testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestBuildSqlDataNodePipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"id":    "bar",
				"name":  "boo",
				"table": "test_table",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("SqlDataNode"),
					},
					{
						Key:         aws.String("table"),
						StringValue: aws.String("test_table"),
					},
				},
			},
		},
		{
			map[string]interface{}{
				"id":           "bar",
				"name":         "boo",
				"table":        "test_table",
				"database":     "hogehoge",
				"select_query": "select * from #{table} where eventTime >= '#{@scheduledStartTime.format('YYYY-MM-dd HH:mm:ss')}' and eventTime < '#{@scheduledEndTime.format('YYYY-MM-dd HH:mm:ss')}'",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("bar"),
				Name: aws.String("boo"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("SqlDataNode"),
					},
					{
						Key:      aws.String("database"),
						RefValue: aws.String("hogehoge"),
					},
					{
						Key:         aws.String("table"),
						StringValue: aws.String("test_table"),
					},
					{
						Key:         aws.String("selectQuery"),
						StringValue: aws.String("select * from #{table} where eventTime >= '#{@scheduledStartTime.format('YYYY-MM-dd HH:mm:ss')}' and eventTime < '#{@scheduledEndTime.format('YYYY-MM-dd HH:mm:ss')}'"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildCommonPipelineObject("SqlDataNode", testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestBuildRdsDatabasePipelineObject(t *testing.T) {
	testCases := []pipelineObjectTestCase{
		{
			map[string]interface{}{
				"id":              "MyRdsDatabase",
				"name":            "MyRdsDatabase",
				"username":        "user_name",
				"password":        "my_password",
				"rds_instance_id": "my_db_instance_identifier",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("MyRdsDatabase"),
				Name: aws.String("MyRdsDatabase"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("RdsDatabase"),
					},
					{
						Key:         aws.String("username"),
						StringValue: aws.String("user_name"),
					},
					{
						Key:         aws.String("*password"),
						StringValue: aws.String("my_password"),
					},
					{
						Key:         aws.String("rdsInstanceId"),
						StringValue: aws.String("my_db_instance_identifier"),
					},
				},
			},
		},
		{
			map[string]interface{}{
				"id":                  "MyRdsDatabase",
				"name":                "MyRdsDatabase",
				"username":            "user_name",
				"password":            "my_password",
				"rds_instance_id":     "my_db_instance_identifier",
				"jdbc_driver_jar_uri": "s3://example.com/test/jdbc.driver",
				"jdbc_properties":     "useUnicode=yes&characterEncoding=UTF-8",
			},
			&datapipeline.PipelineObject{
				Id:   aws.String("MyRdsDatabase"),
				Name: aws.String("MyRdsDatabase"),
				Fields: []*datapipeline.Field{
					{
						Key:         aws.String("type"),
						StringValue: aws.String("RdsDatabase"),
					},
					{
						Key:         aws.String("username"),
						StringValue: aws.String("user_name"),
					},
					{
						Key:         aws.String("*password"),
						StringValue: aws.String("my_password"),
					},
					{
						Key:         aws.String("rdsInstanceId"),
						StringValue: aws.String("my_db_instance_identifier"),
					},
					{
						Key:         aws.String("jdbcDriverJarUri"),
						StringValue: aws.String("s3://example.com/test/jdbc.driver"),
					},
					{
						Key:         aws.String("jdbcProperties"),
						StringValue: aws.String("useUnicode=yes&characterEncoding=UTF-8"),
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		result, err := buildCommonPipelineObject("RdsDatabase", testCase.Attrs)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf(
				"test case %d: got %#v, but want %#v",
				i, result, testCase.Expected,
			)
		}
	}
}

func TestFlattenDefaultPipelineObject(t *testing.T) {
	in := defaultPipelineObjectConf()
	dpo, err := buildDefaultPipelineObject(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := flattenDefaultPipelineObject(dpo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestFlattenCommonPipelineObject_copyActivity(t *testing.T) {
	in := copyActivityPipelineObjectConf()
	epo, err := buildCommonPipelineObject("CopyActivity", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestFlattenCommonPipelineObject_ec2Resource(t *testing.T) {
	in := ec2ResourcePipelineObjectConf()
	epo, err := buildCommonPipelineObject("Ec2Resource", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %+v, got %+v", in, out)
	}
}

func TestFlattenCommonPipelineObject_rdsDatabase(t *testing.T) {
	in := rdsDatabasePipelineObjectConf()
	epo, err := buildCommonPipelineObject("RdsDatabase", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestFlattenCommonPipelineObject_s3DataNode(t *testing.T) {
	in := s3DataNodePipelineObjectConf()
	epo, err := buildCommonPipelineObject("S3DataNode", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestFlattenCommonPipelineObject_sqlDataNode(t *testing.T) {
	in := sqlDataNodePipelineObjectConf()
	epo, err := buildCommonPipelineObject("SqlDataNode", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestFlattenCommonPipelineObject_schedule(t *testing.T) {
	in := schedulePipelineObjectConf()
	epo, err := buildCommonPipelineObject("Schedule", in)
	if err != nil {
		t.Fatal(err)
	}

	out, err := flattenCommonPipelineObject(epo)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %+v, got %+v", in, out)
	}
}

func defaultPipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"schedule_type":          "cron",
		"failure_and_rerun_mode": "CASCADE",
		"pipeline_log_uri":       "s3://example.com/test/",
		"role":                   "arn:aws:iam::123456789012:role/test-role",
		"resource_role":          "arn:aws:iam::123456789012:role/test-resource-role",
		"schedule":               "DefaultSchedule",
	}
}

func copyActivityPipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":                      "DefaultCopyActivity",
		"name":                    "DefaultCopyActivity",
		"schedule":                "DefaultSchedule",
		"runs_on":                 "DefaultEc2Resource",
		"attempt_timeout":         "1 hour",
		"depends_on":              "DefaultEc2Resource",
		"failure_and_rerun_mode":  "CASCADE",
		"input":                   "S3InputDataNode",
		"late_after_timeout":      "2 hours",
		"max_active_instances":    1,
		"maximum_retries":         5,
		"on_fail":                 "DefaultSnsAlarm",
		"on_late_action":          "DefaultSnsAlarm",
		"on_success":              "DefaultSnsAlarm",
		"output":                  "S3OutputDataNode",
		"parent":                  "DefaultEc2Resource",
		"pipeline_log_uri":        "s3://BucketName/Key/",
		"precondition":            "DefaultS3KeyExists",
		"report_progress_timeout": "10 minutes",
		"retry_delay":             "20 minutes",
		"schedule_type":           "cron",
	}
}

func ec2ResourcePipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":                          "DefaultEc2Resource",
		"name":                        "DefaultEc2Resource",
		"associate_public_ip_address": true,
		"attempt_timeout":             "1 hour",
		"availability_zone":           "ap-northeast-1a",
		"http_proxy":                  "DefaultHttpProxy",
		"image_id":                    "ami-012345678",
		"instance_type":               "t2.micro",
		"key_pair":                    "test-ssh-key",
		"max_active_instances":        5,
		"maximum_retries":             10,
		"on_fail":                     "DefaultSnsAlarm",
		"on_late_action":              "DefaultSnsAlarm",
		"on_success":                  "DefaultSnsAlarm",
		"pipeline_log_uri":            "s3://BucketName/Key/",
		"region":                      "ap-northeast-1",
		"schedule_type":               "ondemand",
		"security_group_ids": []string{
			"sg-0123456",
			"sg-1234567",
		},

		"security_groups": []string{
			"default",
			"test-sg",
		},
		"subnet_id":                     "subnet-01234567",
		"spot_bid_price":                0.05,
		"terminate_after":               "15 minutes",
		"use_on_demand_on_last_attempt": true,
	}
}

func rdsDatabasePipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":                  "MyRdsDatabase",
		"name":                "MyRdsDatabase",
		"username":            "user_name",
		"password":            "my_password",
		"rds_instance_id":     "my_db_instance_identifier",
		"database_name":       "database_name",
		"jdbc_properties":     "useUnicode=yes&characterEncoding=UTF-8",
		"region":              "us-east-1",
		"parent":              "DefaultEc2Resource",
		"jdbc_driver_jar_uri": "s3://BucketName/Key/",
	}
}

func s3DataNodePipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":                      "DefaultS3DataNode",
		"name":                    "DefaultS3DataNode",
		"compression":             "gzip",
		"data_format":             "DefaultCsvDataFormat",
		"depends_on":              "DefaultEc2Resource",
		"directory_path":          "s3://my-bucket/my-key-for-directory",
		"failure_and_rerun_mode":  "CASCADE",
		"file_path":               "s3://my-bucket/my-key-for-file",
		"late_after_timeout":      "2 hours",
		"manifest_file_path":      "s3://my-bucket/my-key-for-directory",
		"max_active_instances":    5,
		"maximum_retries":         10,
		"on_fail":                 "DefaultSnsAlarm",
		"on_late_action":          "DefaultSnsAlarm",
		"on_success":              "DefaultSnsAlarm",
		"parent":                  "DefaultEc2Resource",
		"pipeline_log_uri":        "s3://BucketName/Key/",
		"precondition":            "DefaultS3KeyExists",
		"report_progress_timeout": "10 minutes",
		"retry_delay":             "20 minutes",
		"runs_on":                 "DefaultEc2Resource",
		"s3_encryption_type":      "SERVER_SIDE_ENCRYPTION",
		"schedule_type":           "cron",
		"worker_group":            "test",
	}
}

func sqlDataNodePipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":                     "DefaultSqlDataNode",
		"name":                   "DefaultSqlDataNode",
		"table":                  "test",
		"create_table_sql":       "CREATE TABLE #{table}",
		"database":               "DefaultRdsDatabase",
		"depends_on":             "DefaultEc2Resource",
		"failure_and_rerun_mode": "none",
		"insert_query": `INSERT INTO #{table} (col_name1, col_name2)
		VALUES ("value1", "value2")`,
		"maximum_retries":  10,
		"on_fail":          "DefaultSnsAlarm",
		"on_late_action":   "DefaultSnsAlarm",
		"on_success":       "DefaultSnsAlarm",
		"parent":           "DefaultEc2Resource",
		"pipeline_log_uri": "s3://BucketName/Key/",
		"precondition":     "DefaultS3KeyExists",
		"retry_delay":      "20 minutes",
		"select_query":     "SELECT * FROM #{table}",
	}
}

func schedulePipelineObjectConf() map[string]interface{} {
	return map[string]interface{}{
		"id":              "DefaultSchedule",
		"name":            "DefaultSchedule",
		"period":          "1 hour",
		"start_at":        "FIRST_ACTIVATION_DATE_TIME",
		"start_date_time": "2019-01-01T00:00:00",
		"end_date_time":   "2019-09-01T00:00:00",
		"occurrences":     1,
		"parent":          "DefaultEc2Resource",
	}
}
