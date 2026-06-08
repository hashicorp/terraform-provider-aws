# Example: Route53 + Target Group + ALB Listener Integration

This example demonstrates how to create a Route53 DNS record, a Target Group, and associate them with an ALB Listener using Terraform. It's a complete setup to expose your application behind an Application Load Balancer with a custom domain.

## Components

- **Route53 Record** – Points your domain (e.g., `app.example.com`) to the ALB
- **Target Group** – Forwards traffic to registered targets (EC2 instances)
- **ALB Listener** – Listens for HTTP/HTTPS traffic on the ALB

## 🛠️ Getting Started

1. **Clone this repository** or copy the example into your Terraform project.

2. **Modify `terraform.tfvars`**
   Update the variables such as:
   - `alb_listener_arn`
   - `vpc_id`
   - `domain_name`
   - `hosted_zone_id`
   - `targets` (IP or instance IDs)

3. **Initialize Terraform**

   ```hcl
   terraform init
   ```

4. **Format File**

   ```hcl
   terraform fmt
   ```

5. **Validate Terraform**

   ```hcl
   terraform validate
   ```

6. **Review the execution plan**

   ```hcl
   terraform plan
   ```

7. **Apply the configuration**

   ```hcl
   terraform apply
   ```

8. **Verify**
   - The ALB listener should forward traffic to the registered targets.
   - A Route53 DNS record should be created pointing to the ALB.

---
