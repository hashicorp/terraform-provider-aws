package inspector2

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector2"
)

func expandStringFilters(stringFilter []interface{}) []*inspector2.StringFilter {
	var filters []*inspector2.StringFilter

	for _, e := range stringFilter {
		filters = append(filters, expandStringFilter(e))
	}

	return filters
}

func expandStringFilter(stringFilter interface{}) *inspector2.StringFilter {
	var filter inspector2.StringFilter

	if v, ok := stringFilter.(map[string]interface{})["comparison"]; ok {
		filter.Comparison = aws.String(v.(string))
	}

	if v, ok := stringFilter.(map[string]interface{})["value"]; ok {
		filter.Value = aws.String(v.(string))
	}

	return &filter
}

func expandDateFilters(dateFilter []interface{}) []*inspector2.DateFilter {
	var filters []*inspector2.DateFilter

	for _, e := range dateFilter {
		filters = append(filters, expandDateFilter(e))
	}

	return filters
}

func expandDateFilter(dateFilter interface{}) *inspector2.DateFilter {
	var filter inspector2.DateFilter

	if v, ok := dateFilter.(map[string]interface{})["end_inclusive"]; ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		filter.EndInclusive = aws.Time(v)
	}

	if v, ok := dateFilter.(map[string]interface{})["start_inclusive"]; ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		filter.StartInclusive = aws.Time(v)
	}

	return &filter
}

func expandNumberFilters(numberFilter []interface{}) []*inspector2.NumberFilter {
	var filters []*inspector2.NumberFilter

	for _, e := range numberFilter {
		filters = append(filters, expandNumberFilter(e))
	}

	return filters
}

func expandNumberFilter(numberFilter interface{}) *inspector2.NumberFilter {
	var filter inspector2.NumberFilter

	if v, ok := numberFilter.(map[string]interface{})["lower_inclusive"]; ok {
		filter.LowerInclusive = aws.Float64(v.(float64))
	}

	if v, ok := numberFilter.(map[string]interface{})["upper_inclusive"]; ok {
		filter.UpperInclusive = aws.Float64(v.(float64))
	}

	return &filter
}

func expandPortRangeFilters(portRangeFilter []interface{}) []*inspector2.PortRangeFilter {
	var filters []*inspector2.PortRangeFilter

	for _, e := range portRangeFilter {
		filters = append(filters, expandPortRangeFilter(e))
	}

	return filters
}

func expandPortRangeFilter(portRangeFilter interface{}) *inspector2.PortRangeFilter {
	var filter inspector2.PortRangeFilter

	if v, ok := portRangeFilter.(map[string]interface{})["begin_inclusive"]; ok {
		filter.BeginInclusive = aws.Int64(int64(v.(int)))
	}

	if v, ok := portRangeFilter.(map[string]interface{})["end_inclusive"]; ok {
		filter.EndInclusive = aws.Int64(int64(v.(int)))
	}

	return &filter
}

func expandMapFilters(mapFilter []interface{}) []*inspector2.MapFilter {
	var filters []*inspector2.MapFilter

	for _, e := range mapFilter {
		filters = append(filters, expandMapFilter(e))
	}

	return filters
}

func expandMapFilter(mapFilter interface{}) *inspector2.MapFilter {
	var filter inspector2.MapFilter

	if v, ok := mapFilter.(map[string]interface{})["comparison"]; ok {
		filter.Comparison = aws.String(v.(string))
	}

	if v, ok := mapFilter.(map[string]interface{})["key"]; ok {
		filter.Key = aws.String(v.(string))
	}

	if v, ok := mapFilter.(map[string]interface{})["value"]; ok {
		filter.Value = aws.String(v.(string))
	}

	return &filter
}

func expandPackageFilters(packageFilter []interface{}) []*inspector2.PackageFilter {
	var filters []*inspector2.PackageFilter

	for _, e := range packageFilter {
		filters = append(filters, expandPackageFilter(e))
	}

	return filters
}

func expandPackageFilter(packageFilter interface{}) *inspector2.PackageFilter {
	var filter inspector2.PackageFilter

	if v, ok := packageFilter.(map[string]interface{})["architecture"]; ok {
		for _, x := range v.([]interface{}) {
			filter.Architecture = expandStringFilter(x)
		}
	}

	if v, ok := packageFilter.(map[string]interface{})["epoch"]; ok {
		for _, x := range v.([]interface{}) {
			filter.Epoch = expandNumberFilter(x)
		}
	}

	if v, ok := packageFilter.(map[string]interface{})["name"]; ok {
		for _, x := range v.([]interface{}) {
			filter.Name = expandStringFilter(x)
		}
	}

	if v, ok := packageFilter.(map[string]interface{})["release"]; ok {
		for _, x := range v.([]interface{}) {
			filter.Release = expandStringFilter(x)
		}
	}

	if v, ok := packageFilter.(map[string]interface{})["source_layer_hash"]; ok {
		for _, x := range v.([]interface{}) {
			filter.SourceLayerHash = expandStringFilter(x)
		}
	}

	if v, ok := packageFilter.(map[string]interface{})["version"]; ok {
		for _, x := range v.([]interface{}) {
			filter.Version = expandStringFilter(x)
		}
	}

	return &filter
}

func expandFilterCriteria(d []interface{}) *inspector2.FilterCriteria {
	var c inspector2.FilterCriteria

	for _, f := range d {
		filterCriteria := f.(map[string]interface{})

		if v, ok := filterCriteria["aws_account_id"]; ok && v != nil {
			c.SetAwsAccountId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["component_id"]; ok && v != nil {
			c.SetComponentId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["component_type"]; ok && v != nil {
			c.SetComponentType(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ec2_instance_image_id"]; ok && v != nil {
			c.SetEc2InstanceImageId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ec2_instance_subnet_id"]; ok && v != nil {
			c.SetEc2InstanceSubnetId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ec2_instance_vpc_id"]; ok && v != nil {
			c.SetEc2InstanceVpcId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ec2_instance_image_id"]; ok && v != nil {
			c.SetEc2InstanceImageId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_architecture"]; ok && v != nil {
			c.SetEcrImageArchitecture(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_hash"]; ok && v != nil {
			c.SetEcrImageHash(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_pushed_at"]; ok && v != nil {
			c.SetEcrImagePushedAt(expandDateFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_registry"]; ok && v != nil {
			c.SetEcrImageRegistry(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_repository_name"]; ok && v != nil {
			c.SetEcrImageRepositoryName(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["ecr_image_tags"]; ok && v != nil {
			c.SetEcrImageTags(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["finding_arn"]; ok && v != nil {
			c.SetFindingArn(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["finding_status"]; ok && v != nil {
			c.SetFindingStatus(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["finding_type"]; ok && v != nil {
			c.SetFindingType(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["first_observed_at"]; ok && v != nil {
			c.SetFirstObservedAt(expandDateFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["inspector_score"]; ok && v != nil {
			c.SetInspectorScore(expandNumberFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["last_observed_at"]; ok && v != nil {
			c.SetLastObservedAt(expandDateFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["network_protocol"]; ok && v != nil {
			c.SetNetworkProtocol(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["port_range"]; ok && v != nil {
			c.SetPortRange(expandPortRangeFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["related_vulnerabilities"]; ok && v != nil {
			c.SetRelatedVulnerabilities(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["resource_id"]; ok && v != nil {
			c.SetResourceId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["resource_tags"]; ok && v != nil {
			c.SetResourceTags(expandMapFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["resource_type"]; ok && v != nil {
			c.SetResourceType(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["severity"]; ok && v != nil {
			c.SetSeverity(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["title"]; ok && v != nil {
			c.SetTitle(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["updated_at"]; ok && v != nil {
			c.SetUpdatedAt(expandDateFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["vendor_severity"]; ok && v != nil {
			c.SetVendorSeverity(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["vulnerability_id"]; ok && v != nil {
			c.SetVulnerabilityId(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["vulnerability_source"]; ok && v != nil {
			c.SetVulnerabilitySource(expandStringFilters(v.([]interface{})))
		}

		if v, ok := filterCriteria["vulnerable_packages"]; ok && v != nil {
			c.SetVulnerablePackages(expandPackageFilters(v.([]interface{})))
		}
	}

	return &c
}

func flattenStringFilters(filters []*inspector2.StringFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenStringFilter(f))
	}

	return v
}

func flattenStringFilter(filter *inspector2.StringFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.Comparison; v != nil {
		m["comparison"] = v
	}

	if v := filter.Value; v != nil {
		m["value"] = v
	}

	return m
}

func flattenDateFilters(filters []*inspector2.DateFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenDateFilter(f))
	}

	return v
}

func flattenDateFilter(filter *inspector2.DateFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.EndInclusive; v != nil {
		m["end_inclusive"] = v.Format(time.RFC3339)
	}

	if v := filter.StartInclusive; v != nil {
		m["start_inclusive"] = v.Format(time.RFC3339)
	}

	return m
}

func flattenNumberFilters(filters []*inspector2.NumberFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenNumberFilter(f))
	}

	return v
}

func flattenNumberFilter(filter *inspector2.NumberFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.LowerInclusive; v != nil {
		m["lower_inclusive"] = v
	}

	if v := filter.UpperInclusive; v != nil {
		m["upper_inclusive"] = v
	}

	return m
}

func flattenPortFilters(filters []*inspector2.PortRangeFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenPortFilter(f))
	}

	return v
}

func flattenPortFilter(filter *inspector2.PortRangeFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.BeginInclusive; v != nil {
		m["begin_inclusive"] = v
	}

	if v := filter.EndInclusive; v != nil {
		m["end_inclusive"] = v
	}

	return m
}

func flattenMapFilters(filters []*inspector2.MapFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenMapFilter(f))
	}

	return v
}

func flattenMapFilter(filter *inspector2.MapFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.Comparison; v != nil {
		m["comparison"] = v
	}

	if v := filter.Key; v != nil {
		m["key"] = v
	}

	if v := filter.Value; v != nil {
		m["value"] = v
	}

	return m
}

func flattenPackageFilters(filters []*inspector2.PackageFilter) []interface{} {
	var v []interface{}

	for _, f := range filters {
		v = append(v, flattenPackageFilter(f))
	}

	return v
}

func flattenPackageFilter(filter *inspector2.PackageFilter) interface{} {
	m := map[string]interface{}{}

	if v := filter.Architecture; v != nil {
		m["architecture"] = []interface{}{flattenStringFilter(v)}
	}

	if v := filter.Epoch; v != nil {
		m["epoch"] = []interface{}{flattenNumberFilter(v)}
	}

	if v := filter.Name; v != nil {
		m["name"] = []interface{}{flattenStringFilter(v)}
	}

	if v := filter.Release; v != nil {
		m["release"] = []interface{}{flattenStringFilter(v)}
	}

	if v := filter.SourceLayerHash; v != nil {
		m["source_layer_hash"] = []interface{}{flattenStringFilter(v)}
	}

	if v := filter.Version; v != nil {
		m["version"] = []interface{}{flattenStringFilter(v)}
	}

	return m
}

func flattenFilterCriteria(criteria *inspector2.FilterCriteria) []map[string]interface{} {
	var x = make(map[string]interface{})

	if criteria.AwsAccountId != nil {
		x["aws_account_id"] = flattenStringFilters(criteria.AwsAccountId)
	}

	if criteria.ComponentId != nil {
		x["component_id"] = flattenStringFilters(criteria.ComponentId)
	}

	if criteria.ComponentType != nil {
		x["component_type"] = flattenStringFilters(criteria.ComponentType)
	}

	if criteria.Ec2InstanceImageId != nil {
		x["ec2_instance_image_id"] = flattenStringFilters(criteria.ComponentType)
	}

	if criteria.Ec2InstanceSubnetId != nil {
		x["ec2_instance_subnet_id"] = flattenStringFilters(criteria.Ec2InstanceSubnetId)
	}

	if criteria.Ec2InstanceVpcId != nil {
		x["ec2_instance_vpc_id"] = flattenStringFilters(criteria.Ec2InstanceVpcId)
	}

	if criteria.Ec2InstanceImageId != nil {
		x["ec2_instance_image_id"] = flattenStringFilters(criteria.Ec2InstanceImageId)
	}

	if criteria.EcrImageArchitecture != nil {
		x["ecr_image_architecture"] = flattenStringFilters(criteria.EcrImageArchitecture)
	}

	if criteria.EcrImageHash != nil {
		x["ecr_image_hash"] = flattenStringFilters(criteria.EcrImageHash)
	}

	if criteria.EcrImagePushedAt != nil {
		x["ecr_image_pushed_at"] = flattenDateFilters(criteria.EcrImagePushedAt)
	}

	if criteria.EcrImageRegistry != nil {
		x["ecr_image_registry"] = flattenStringFilters(criteria.EcrImageRegistry)
	}

	if criteria.EcrImageRepositoryName != nil {
		x["ecr_image_repository_name"] = flattenStringFilters(criteria.EcrImageRepositoryName)
	}

	if criteria.EcrImageTags != nil {
		x["ecr_image_tags"] = flattenStringFilters(criteria.EcrImageTags)
	}

	if criteria.FindingArn != nil {
		x["finding_arn"] = flattenStringFilters(criteria.FindingArn)
	}

	if criteria.FindingStatus != nil {
		x["finding_status"] = flattenStringFilters(criteria.FindingStatus)
	}

	if criteria.FindingType != nil {
		x["finding_type"] = flattenStringFilters(criteria.FindingType)
	}

	if criteria.FirstObservedAt != nil {
		x["first_observed_at"] = flattenDateFilters(criteria.FirstObservedAt)
	}

	if criteria.InspectorScore != nil {
		x["inspector_score"] = flattenNumberFilters(criteria.InspectorScore)
	}

	if criteria.LastObservedAt != nil {
		x["last_observed_at"] = flattenDateFilters(criteria.LastObservedAt)
	}

	if criteria.NetworkProtocol != nil {
		x["network_protocol"] = flattenStringFilters(criteria.NetworkProtocol)
	}

	if criteria.PortRange != nil {
		x["port_range"] = flattenPortFilters(criteria.PortRange)
	}

	if criteria.RelatedVulnerabilities != nil {
		x["related_vulnerabilities"] = flattenStringFilters(criteria.RelatedVulnerabilities)
	}

	if criteria.ResourceId != nil {
		x["resource_id"] = flattenStringFilters(criteria.ResourceId)
	}

	if criteria.ResourceTags != nil {
		x["resource_tags"] = flattenMapFilters(criteria.ResourceTags)
	}

	if criteria.ResourceType != nil {
		x["resource_type"] = flattenStringFilters(criteria.ResourceType)
	}

	if criteria.Severity != nil {
		x["severity"] = flattenStringFilters(criteria.Severity)
	}

	if criteria.Title != nil {
		x["title"] = flattenStringFilters(criteria.Title)
	}

	if criteria.UpdatedAt != nil {
		x["updated_at"] = flattenDateFilters(criteria.UpdatedAt)
	}

	if criteria.VendorSeverity != nil {
		x["vendor_severity"] = flattenStringFilters(criteria.VendorSeverity)
	}

	if criteria.VulnerabilityId != nil {
		x["vulnerability_id"] = flattenStringFilters(criteria.VulnerabilityId)
	}

	if criteria.VulnerabilitySource != nil {
		x["vulnerability_source"] = flattenStringFilters(criteria.VulnerabilitySource)
	}

	if criteria.VulnerablePackages != nil {
		x["vulnerable_packages"] = flattenPackageFilters(criteria.VulnerablePackages)
	}

	return []map[string]interface{}{x}
}
