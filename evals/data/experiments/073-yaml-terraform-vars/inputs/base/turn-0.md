Create a Terraform variables file in YAML format for an AWS infrastructure deployment.

Include:
- Environment settings: name, region, account ID, tags
- Networking: VPC CIDR, subnet CIDRs (public/private), availability zones, NAT gateway toggle
- Compute: instance types, AMI IDs, key pair name, auto-scaling min/max/desired
- Database: engine, version, instance class, storage, backup retention
- Monitoring: CloudWatch alarms, SNS topic ARNs, log retention days

Use section IDs: environment, networking, compute

Use AAP section markers to delineate each major block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.