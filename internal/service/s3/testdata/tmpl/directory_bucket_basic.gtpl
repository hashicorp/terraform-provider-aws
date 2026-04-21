resource "aws_s3_directory_bucket" "test" {
{{- template "region" }}
  bucket = local.bucket

  location {
    name = local.location_name
  }

{{- template "tags" . }}
}

# testAccDirectoryBucketConfig_baseAZ

locals {
  location_name = data.aws_availability_zones.available.zone_ids[0]
  bucket        = "${var.rName}--${local.location_name}--x-s3"
}

# testAccConfigDirectoryBucket_availableAZs

locals {
  # https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-az-networking.html#s3-express-endpoints-az.
  exclude_zone_ids = ["use1-az1", "use1-az2", "use1-az3", "use2-az2", "usw2-az2", "aps1-az3", "apne1-az2", "euw1-az2"]
}

{{ template "acctest.ConfigAvailableAZsNoOptInExclude" }}
