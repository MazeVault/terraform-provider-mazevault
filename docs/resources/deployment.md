# mazevault_deployment

Agent deployment bootstrap package (OS type, mode, agent server URL, auto-update).

`gateway_url` is the server URL the installed agents connect to. Leave it unset for agents that can reach the primary directly. For a datacenter without a network path to the primary, point it at that datacenter's gateway (`public_url` of the gateway, configured with `MAZEVAULT_GATEWAY_PUBLIC_URL`): the gateway relays the agent API — registration, heartbeat, discovery and the deployment WebSocket — to the primary, which remains the only control plane.

## Example Usage

```hcl
# See the [Terraform Registry documentation](https://registry.terraform.io/providers/MazeVault/mazevault/latest/docs/resources/deployment) for a complete example.
```

## Argument Reference

Refer to the [Terraform Registry](https://registry.terraform.io/providers/MazeVault/mazevault/latest/docs/resources/deployment) for the full list of supported arguments and attributes.

## Import

Most `mazevault_deployment` resources can be imported by UUID:

```
terraform import mazevault_deployment.example <uuid>
```
