# configure-nad

## Overview

<!--mdtogo:Short-->

A function to configure [NetworkAttachmentDefinitions](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/deployments/multus-daemonset.yml#L13)(NAD),
it is used to create a NAD from ConfigMap data

<!--mdtogo-->

The function is used to simplify the creation and configuration of NADs, which are used to create secondary networks in
pods.

<!--mdtogo:Long-->

## Usage

The function reads ConfigMaps from the resourcesList, and looks for maps with annotation `configure-nad` 
(can be changed using FunctionConfig). The function then reads this COnfigMap, and create a NAD with a config field 
matching the specs defined in the ConfigMap data field. Currently the function only supports the sriov, and macvlan cnis
only, in the future, more cnis will be added.

### FunctionConfig

We use ConfigMap to configure the `configure-nad` function. The configurations
values are provided as key-value pairs using `data` field where there are two optional fields:

**identifierAnnotation**: The ConfigMap annotation to identify the target ConfigMap.

**resourceName**: An additional resource to link for this NAD.

Following is an example ConfigMap to configure a nad matching [sriov-net-a](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/examples/sriov-pod.yml#L8)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configure-nad-func-config
data:
  identifierAnnotation: "configure-nad"
  resourceName: "intel.com/sriov"
```

<!--mdtogo-->
