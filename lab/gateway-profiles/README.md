# Gateway Profile Images

Each gateway stub profile has its own Dockerfile so the image carries the
profile identity in its entrypoint. Compose supplies only lab-specific runtime
values such as MAC address, IP address, hostname, controller URL, SSH state, and
status paths.

Available gateway profile images:

- `ugw3/Dockerfile`: starts `unifi-stubd -profile ugw3`
- `uxg-lite/Dockerfile`: starts `unifi-stubd -profile uxg-lite`
- `uxgpro/Dockerfile`: starts `unifi-stubd -profile uxgpro`
- `ucg-fiber/Dockerfile`: starts `unifi-stubd -profile ucg-fiber`

Build through the controller lab:

```sh
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile ugw3 \
  up -d --build
```
