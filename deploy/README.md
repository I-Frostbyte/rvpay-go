# OCI Always Free deployment

These artifacts deploy the existing Deposits gRPC service on one ARM64
VM.Standard.A1.Flex instance. The Compose resource limits total 2 OCPUs and
1.75 GB RAM, leaving capacity within the Always Free allowance. PostgreSQL uses
a named Docker volume, which counts toward the VM's block storage allocation.

## Before deployment

1. Create an Oracle Cloud Free Tier account and select the Home Region
   carefully: Always Free resources stay there. Oracle requires a credit card
   for identity verification.
2. Create an ARM64 VM.Standard.A1.Flex instance with no more than 2 OCPUs and
   12 GB RAM. In the VCN Security List, allow inbound TCP 22 only from an
   administration IP and TCP 80/443 from the required public networks.
3. Copy this repository to `/opt/rvpay-go`, then run
   `sudo ./deploy-oci-bootstrap.sh`.
4. Copy `.env.example` to `.env` and replace its placeholders. Keep the file
   readable only by the deployment user.
5. Place your TLS certificate chain and private key at
   `certs/fullchain.pem` and `certs/privkey.pem`. Nginx runs unprivileged,
   redirects HTTP to HTTPS, and forwards TLS gRPC to the internal service.
6. Install the systemd unit with
   `sudo cp systemd/rvpay-go.service /etc/systemd/system/`, then run
   `sudo systemctl daemon-reload && sudo systemctl enable --now rvpay-go`.

## CI/CD secrets

The GitHub workflow requires `OCI_HOST`, `OCI_USER`, and
`OCI_SSH_PRIVATE_KEY`. Configure the OCI host to authenticate to GHCR before
the first deployment if the image package is private. The workflow pushes a
single `linux/arm64` image, pulls that immutable SHA-tagged image over SSH, and
restarts only the application stack.

## Migrations

`migration` is the only Compose service that runs `golang-migrate`; the
application is started with `RUN_MIGRATIONS=false`. This avoids migration races
when the service is scaled. The migration database URL expects a URL-safe
password; use percent encoding for reserved URL characters.
