# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

output "collection_enpdoint" {
  value = aws_opensearchserverless_collection.collection.collection_endpoint
}

output "dashboard_endpoint" {
  value = aws_opensearchserverless_collection.collection.dashboard_endpoint
}
