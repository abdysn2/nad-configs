# configure-nad

## Overview

<!--mdtogo:Short-->

A function to configre [NetworkAttachmentDefinitions](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/deployments/multus-daemonset.yml#L13)(NAD),
it is used to configure the config field and the resource annotation

<!--mdtogo-->

The function is used to simplify the creation and configuration of NADs, which are used to create secondary networks in
pods.

<!--mdtogo:Long-->

## Usage

The function is used to set two things, the `spec.config` field of the NAD, and the resource annotation for requesting
additional resources (e.g. SRIOV VFs).

To use the function, you will need to provide the NAD without the `spec` and `annotation` field, and then provide the
function with the config and additional resource name. The function would then set the `spec.config` to match the input
config and set the `k8s.v1.cni.cncf.io/resourceName` annotation to match the resource name.

### FunctionConfig

We use ConfigMap to configure the `configure-nad` function. The configurations
values are provided as key-value pairs using `data` field where there are three fields:

**config**: The nad config to set. 

**resourceName**: an optional additional resource to link for this NAD.

Following is an example ConfigMap to configure a nad matching [sriov-net-a](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/examples/sriov-pod.yml#L8)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configure-nad-func-config
data:
  config: '{
  "type": "sriov",
  "vlan": 1000,
  "ipam": {
    "type": "host-local",
    "subnet": "10.56.217.0/24",
    "rangeStart": "10.56.217.171",
    "rangeEnd": "10.56.217.181",
    "routes": [{
      "dst": "0.0.0.0/0"
    }],
    "gateway": "10.56.217.1"
  }
}'
  resourceName: "intel.com/sriov"
```

<!--mdtogo-->
