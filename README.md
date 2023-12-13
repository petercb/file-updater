# file-updater

file-updater is a small untility that reads a list of files from a yaml config
file and downloads them from the Source location to the Dest location on the
local system.

In it's default behaviour it watches the config file for changes and responds by
re-processing the config file each time.

The primary use case for this is as a Kubernetes sidecar container to update
some Minecraft server plugins whenever a configMap is updated.

## Example config

```yaml
---
files:
  - source: https://example.com/stuff.txt
    dest: /tmp/stuff.txt
```
