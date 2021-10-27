# Proxmox generate resource usage CSV report

Based on https://pve.proxmox.com/wiki/Proxmox_VE_API it will loop thru the API, extract all virtuals and group them by first part of the name ( splitted by `-` ) and then extract their CPU share and RAM MB usage.

## Setup

1. [Install golang](https://golang.org/doc/install)
2. [Install direnv](https://direnv.net/#basic-installation)
3. Setup credentials in `.envrc` based on `.envrc_example`
4. Allow direnv `direnv allow`

## Run

    go run main.go
