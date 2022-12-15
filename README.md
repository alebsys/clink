# clink
A simple tool for finding the network interface of a container on a host machine.

### Usage

```bash
clink --help
Usage of clink:
  -i, --container.id string   Container ID
  -n, --namespace string      Used namespace (default "k8s.io")
  -r, --runtime string        Used runtime (default "containerd")
  -v, --verbose               Verbose output (default false)
```

### Examples

```bash
# clink -i 3c6a72c0331794ec6aa2ce3b0d7be5de39624c8348bf5616e21b5c55bf21be68
cali588f0c831aa


# clink -v -i 3c6a72c0331794ec6aa2ce3b0d7be5de39624c8348bf5616e21b5c55bf21be68
+------------------------------------------------------------------+-------+-------------------+
| ID                                                               |   PID | NETWORK INTERFACE |
+------------------------------------------------------------------+-------+-------------------+
| 3c6a72c0331794ec6aa2ce3b0d7be5de39624c8348bf5616e21b5c55bf21be68 | 11914 | cali588f0c831aa   |
+------------------------------------------------------------------+-------+-------------------+
```

### Roadmap

* remove unnecessary dependencies
* search by incomplete container ID
