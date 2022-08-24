Collecting debugging and troubleshooting info
----------------------

When having troubles with the GPU add-on, you may be asked by the support team to collect must-gather data for the add-on. The collection relies on the standard [OpenShift mechanism](https://docs.openshift.com/container-platform/4.10/support/gathering-cluster-data.html) for gathering troubleshooting data.

>**WARNING**: The collected data may include sensitive information such as secrets related to the add-on. Be careful when handing it over to other people.

# GPU Add-on

First, locate the GPU add-on `ClusterServiceVersion` (CSV) and note its output:

```sh
oc get csv -n <gpu-addon-namespace> -o custom-columns=NAME:.metadata.name --no-headers | grep nvidia-gpu-addon
```

The default GPU add-on namespace is `redhat-nvidia-gpu-addon`, so in most cases the command will look like

```sh
oc get csv -n redhat-nvidia-gpu-addon -o custom-columns=NAME:.metadata.name --no-headers | grep nvidia-gpu-addon
```

Next, find out the right must-gather container image by running:

```sh
oc get csv <gpu-addon-csv> -n <gpu-addon-namespace> -o jsonpath='{.spec.relatedImages[?(@.name == "must-gather")].image}'
```

Finally, gather the troubleshooting info:

```sh
oc adm must-gather --image=<must-gather-image> --dest-dir=<destination>
```

Let's put it all together:

```sh
namespace=redhat-nvidia-gpu-addon
csv=$(oc get csv -n $namespace -o custom-columns=NAME:.metadata.name --no-headers | grep nvidia-gpu-addon)
must_gather_image=$(oc get csv $csv -n $namespace -o jsonpath='{.spec.relatedImages[?(@.name == "must-gather")].image}')
oc adm must-gather --image="$must_gather_image" --dest-dir=gpu-addon-gather
```

# NVIDIA GPU Operator

For gathering NVIDIA GPU Operator debugging data, follow the following procedure:

1. Find out the NVIDIA GPU Operator container image:
```sh
oc get pods -A -lapp=gpu-operator -o=jsonpath='{.items[0].spec.containers[0].image}'
```

2. Invoke `oc adm must-gather` using this image:
```sh
oc adm must-gather --image=<gpu-operator-image> --dest-dir=<destination>
```

Or as a script:

```sh
operator_image=$(oc get pods -A -lapp=gpu-operator -o=jsonpath='{.items[0].spec.containers[0].image}')
oc adm must-gather --image="$operator_image" --dest-dir=nvidia-gpu-operator-gather
```