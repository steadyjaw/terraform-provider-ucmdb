
# Microfocus UCMDB Provider

The UCMDB Provider can be used to create, update or delete Configuration Item (CI) in [Micro Focus UCMDB](https://docs.microfocus.com/doc/Universal_CMDB/2020.08) using the REST API. Documentation regarding the Resources supported by the UCMDB Provider can be found in the navigation to the left.

## Getting Started

If you're new to the UCMDB, check out Micro Focus [documentation](https://docs.microfocus.com/doc/Universal_CMDB/2020.08)


## Configuring UCMDB Provider

```hcl
# Configure Terraform
terraform {
  required_providers {
    ucmdb = {
      source = "steadyjaw/ucmdb"
      version = "1.0.1"
    }
  }
}

# Configure Provider options
provider "ucmdb" {
  target_env = "CMS"
}

````

## Argument Reference

- `target_env` (String) Tag representing the target UCMDB environment. Possible values are: `CMS`, `APM`, `OPSB`. Based on the tag value the following Environmental Variables must be set:

    `UCMDB_<target_env>_ADDRESS` - UCMDB REST API url, e.g. `https://<fqdn>:<port>/rest-api`

    `UCMDB_<target_env>_API_USER` - UCMDB REST API username

    `UCMDB_<target_env>_API_PASSWORD` - UCMDB REST API password


