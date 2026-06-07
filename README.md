# Terraform Provider for Cockpit

The Cockpit Terraform Provider allows you to manage Cockpit resources using Infrastructure as Code (IaC). Cockpit is a virtualization management platform for managing multiple Vapor environments (utilizing KVM, QEMU, libvirt, etc.).

## Requirements

- Terraform >= 1.0
- Go >= 1.25 (to build from source)

## Configuration

To use the provider, configure it in your Terraform files. Cockpit requires service account credentials (client ID and client secret) for authentication.

### Provider Schema

```hcl
provider "cockpit" {
  host          = "http://localhost:7771/api/v1"
  client_id     = "your-service-account-client-id"
  client_secret = "your-service-account-client-secret"
}
```

### Environment Variables

Alternatively, you can set configuration values via environment variables:

- `COCKPIT_HOST` (e.g. `http://localhost:7771/api/v1`)
- `COCKPIT_CLIENT_ID`
- `COCKPIT_CLIENT_SECRET`

---

## Resources

### cockpit_host

Manages a Vapor host connected to Cockpit.

#### Example Usage

```hcl
resource "cockpit_host" "example" {
  address   = "10.0.0.15"
  port      = 7770
  api_token = "vapor-api-token-here"
  parent_id = "datacenter-or-cluster-uuid"
}
```

#### Argument Reference

- `address` (String, Required) - Vapor host IP address or hostname.
- `port` (Number, Optional) - Vapor API port (defaults to 7770).
- `api_token` (String, Required, Sensitive) - Vapor API token.
- `parent_id` (String, Required) - Inventory entity ID (Datacenter or Cluster UUID) to place host under.

---

### cockpit_virtual_machine

Manages a Virtual Machine in Cockpit.

#### Example Usage

```hcl
resource "cockpit_virtual_machine" "example" {
  host_id   = "host-uuid-here"
  name      = "terraform-vm"
  memory    = 2048
  vcpus     = 2
  autostart = true

  os = {
    type    = "linux"
    variant = "ubuntu22.04"
  }

  storage = {
    disks = [
      {
        size = 20
        pool = "default"
      }
    ]
  }

  networks = [
    {
      type   = "network"
      source = "default"
    }
  ]
}
```

#### Argument Reference

- `host_id` (String, Required) - Vapor host ID to run the VM on.
- `name` (String, Required) - Name of the virtual machine.
- `memory` (Number, Required) - RAM size in MB.
- `vcpus` (Number, Required) - Number of virtual CPUs.
- `cluster_id` (String, Optional) - Target Cluster ID.
- `datacenter_id` (String, Optional) - Target Datacenter ID.
- `os` (Block, Optional) - OS configuration.
- `storage` (Block, Optional) - Storage configuration (disks/iso).
- `networks` (Block List, Optional) - Network interfaces.
- `autostart` (Boolean, Optional) - Automatically start the VM on boot (defaults to false).

---

### cockpit_kubernetes_cluster

Manages a Kubernetes cluster provisioned via Cockpit's Auto-Provisioner.

#### Example Usage

```hcl
resource "cockpit_kubernetes_cluster" "k8s" {
  name           = "k8s-prod"
  version        = "v1.36.1+k3s1"
  network_id     = "network-switch-uuid"
  ssh_public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQ..."
  image_path     = "/var/lib/libvirt/images/ubuntu-k8s-template.qcow2"

  control_plane = {
    cpu             = 2
    memory_mb       = 2048
    storage_pool_id = "default"
    storage_gb      = 20
    desired_count   = 1
  }

  workers = {
    cpu             = 2
    memory_mb       = 2048
    storage_pool_id = "default"
    storage_gb      = 20
    desired_count   = 3
  }

  csi_tiers = {
    local_path_enabled = true
    longhorn_enabled   = false
  }
}
```

#### Argument Reference

- `name` (String, Required) - Cluster name.
- `version` (String, Required) - K3s version tag.
- `network_id` (String, Required) - Network switch ID (UUID).
- `ssh_public_key` (String, Required) - SSH Public key configured on cluster nodes.
- `image_path` (String, Optional) - Custom template or golden image path.
- `control_plane` (Block, Required) - Control plane node pool spec.
- `workers` (Block, Required) - Worker node pool spec.
- `csi_tiers` (Block, Optional) - Enabled storage CSI drivers.

---

### cockpit_datastore

Manages a Datastore (storage pool) on a Vapor host connected to Cockpit.

#### Example Usage

```hcl
resource "cockpit_datastore" "example" {
  name      = "nfs-images"
  host_id   = "host-uuid-here"
  type      = "netfs"
  path      = "/var/lib/libvirt/images/nfs-images"
  source    = "10.0.0.5:/exports/images"
  autostart = true
  scope     = "shared"
}
```

#### Argument Reference

- `name` (String, Required) - Name of the storage pool / datastore.
- `host_id` (String, Required) - Target Vapor Host ID.
- `type` (String, Required) - Type of storage pool (e.g. `dir`, `netfs`, `logical`, `fs`).
- `path` (String, Optional) - Target mount path for directory/netfs pool.
- `source` (String, Optional) - Source path or host directory (e.g., NFS server export source).
- `target` (String, Optional) - Target storage path.
- `autostart` (Boolean, Optional) - Whether to start the storage pool automatically on host boot (defaults to true).
- `scope` (String, Optional) - Scope of the datastore (`local` or `shared`, defaults to `local`).

---

### cockpit_switch

Manages a Switch (virtual network) on a Vapor host connected to Cockpit.

#### Example Usage

```hcl
resource "cockpit_switch" "example" {
  name       = "vnet-nat-example"
  host_id    = "host-uuid-here"
  mode       = "nat"
  ip_address = "192.168.100.1"
  netmask    = "255.255.255.0"
  dhcp_start = "192.168.100.10"
  dhcp_end   = "192.168.100.100"
  autostart  = true
}
```

#### Argument Reference

- `name` (String, Required) - Name of the network switch.
- `host_id` (String, Required) - Target Vapor Host ID.
- `mode` (String, Required) - Virtual network mode (`nat`, `bridge`, `route`, `isolated`).
- `bridge` (String, Optional) - Bridge device name.
- `ip_address` (String, Optional) - IP address of the bridge interface.
- `netmask` (String, Optional) - Netmask of the bridge interface network.
- `dhcp_start` (String, Optional) - DHCP start IP address pool.
- `dhcp_end` (String, Optional) - DHCP end IP address pool.
- `autostart` (Boolean, Optional) - Whether to start the switch automatically on host boot (defaults to true).
- `domain` (String, Optional) - DNS Domain name config for DHCP clients.

---

## Developing the Provider

To compile the provider locally, run:

```bash
go build -o terraform-provider-cockpit
```
