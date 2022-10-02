---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_gateway"
description: |-
  Terraform data source for managing an AWS App Mesh Virtual Gateway.
---

# Data Source: aws_appmesh_virtual_gateway

Terraform data source for managing an AWS App Mesh Virtual Gateway.

## Example Usage

### Basic Usage

```hcl
data "aws_appmesh_virtual_gateway" "example" {
  mesh_name = "mesh-gateway"
  name      = "example-mesh"
}
```

```hcl
data "aws_caller_identity" "current" {}

data "aws_appmesh_virtual_gateway" "test" {
  name       = "example.mesh.local"
  mesh_name  = "example-mesh"
  mesh_owner = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are required:


* `name` - (Required) Name of the virtual gateway.
* `mesh_name` - (Required) Name of the service mesh in which the virtual gateway exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:


* `arn` - ARN of the virtual gateway.
* `created_date` - Creation date of the virtual gateway.
* `last_updated_date` - Last update date of the virtual gateway.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual gateway specification
* `tags` - Map of tags.

### Spec

* `listener` - Listeners that the mesh endpoint is expected to receive inbound traffic from. You can specify one listener.
* `backend_defaults` - Defaults for backends.
* `logging` - Inbound and outbound access logging information for the virtual gateway.

### Backend_defaults

* `client_policy` - Default client policy for virtual gateway backends.

### Client_policy

* `tls` - Transport Layer Security (TLS) client policy.

### Tls

* `certificate` - Virtual gateway's client's Transport Layer Security (TLS) certificate.
* `enforce` - Whether the policy is enforced. Default is `true`.
* `ports` - One or more ports that the policy is enforced for.
* `validation` - TLS validation context.

### Certificate

* `file` - Local file certificate.
* `sds` - A [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate.

### File

* `certificate_chain` - Certificate chain for the certificate.
* `private_key` - Private key for a certificate stored on the file system of the mesh endpoint that the proxy is running on.

### Sds

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing Transport Layer Security (TLS) materials like a certificate or certificate chain.

### Validation

* `subject_alternative_names` - SANs for a virtual gateway's listener's Transport Layer Security (TLS) validation context.
* `trust` - TLS validation context trust.

### Subject_alternative_names

* `match` - Criteria for determining a SAN's match.

### Match

* `exact` - Values sent must match the specified values exactly.

### Trust

* `acm` - TLS validation context trust for an AWS Certificate Manager (ACM) certificate.
* `file` - TLS validation context trust for a local file certificate.
* `sds` - TLS validation context trust for a [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate.

### Acm

* `certificate_authority_arns` - One or more ACM ARNs.

### File

* `certificate_chain` - Certificate trust chain for a certificate stored on the file system of the mesh endpoint that the proxy is running on. Must be between 1 and 255 characters in length.

### Sds

* `secret_name` - Name of the secret for a virtual gateway's Transport Layer Security (TLS) Secret Discovery Service validation context trust.

### Listener

* `port_mapping` - Port mapping information for the listener.
* `connection_pool` - Connection pool information for the listener.
* `health_check` - Health check information for the listener.
* `tls` - Transport Layer Security (TLS) properties for the listener

### Logging

* `access_log` - Access log configuration for a virtual gateway.

### Access_log

* `file` - File object to send virtual gateway access logs to.

### File

* `path` - File path to write access logs to. You can use `/dev/stdout` to send access logs to standard out. Must be between 1 and 255 characters in length.

### Port_mapping

* `port` - Port used for the port mapping.
* `protocol` - Protocol used for the port mapping. Valid values are `http`, `http2`, `tcp` and `grpc`.

### Connection_pool

* `grpc` - Connection pool information for gRPC listeners.
* `http` - Connection pool information for HTTP listeners.
* `http2` - Connection pool information for HTTP2 listeners.

### Grpc

* `max_requests` - Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster. Minimum value of `1`.

### Http

* `max_connections` - Maximum number of outbound TCP connections Envoy can establish concurrently with all hosts in upstream cluster. Minimum value of `1`.
* `max_pending_requests` - Number of overflowing requests after `max_connections` Envoy will queue to upstream cluster. Minimum value of `1`.

### Http2

* `max_requests` - Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster. Minimum value of `1`.

### Health_check

* `healthy_threshold` - Number of consecutive successful health checks that must occur before declaring listener healthy.
* `interval_millis`- Time period in milliseconds between each health check execution.
* `protocol` - Protocol for the health check request. Valid values are `http`, `http2`, and `grpc`.
* `timeout_millis` - Amount of time to wait when receiving a response from the health check, in milliseconds.
* `unhealthy_threshold` - Number of consecutive failed health checks that must occur before declaring a virtual gateway unhealthy.
* `path` - Destination path for the health check request. This is only required if the specified protocol is `http` or `http2`.
* `port` - Destination port for the health check request. This port must match the port defined in the `port_mapping` for the listener.

### Tls

* `certificate` - Listener's TLS certificate.
* `mode`- Listener's TLS mode. Valid values: `DISABLED`, `PERMISSIVE`, `STRICT`.
* `validation`- Listener's Transport Layer Security (TLS) validation context.

### Certificate

* `acm` - An AWS Certificate Manager (ACM) certificate.
* `file` - Local file certificate.
* `sds` - A [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate.

### Acm

* `certificate_arn` - ARN for the certificate.

### File

* `certificate_chain` - Certificate chain for the certificate. Must be between 1 and 255 characters in length.
* `private_key` - Private key for a certificate stored on the file system of the mesh endpoint that the proxy is running on. Must be between 1 and 255 characters in length.

### Sds

* `secret_name` - Name of the secret requested from the Secret Discovery Service provider representing Transport Layer Security (TLS) materials like a certificate or certificate chain.

### Validation

* `subject_alternative_names` - SANs for a virtual gateway's listener's Transport Layer Security (TLS) validation context.
* `trust` - TLS validation context trust.

### Subject_alternative_names

* `match` - Criteria for determining a SAN's match.

### Match

* `exact` - Values sent must match the specified values exactly.

### Trust

* `file` - TLS validation context trust for a local file certificate.
* `sds` - TLS validation context trust for a [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret#secret-discovery-service-sds) certificate.

### File

* `certificate_chain` - Certificate trust chain for a certificate stored on the file system of the mesh endpoint that the proxy is running on. Must be between 1 and 255 characters in length.

### Sds

* `secret_name` - Name of the secret for a virtual gateway's Transport Layer Security (TLS) Secret Discovery Service validation context trust.
