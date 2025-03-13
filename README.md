`in2un` - Read an [OpenShift insights archive](https://github.com/openshift/insights-operator/tree/master/docs/insights-archive-sample) and parse its content into [unstructured.Unstructured](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured) Golang types, or, print its logs if available. 

## Usage

While this is primarily used as a plugin for other tools, this repo contains code to build a standalone CLI.

~~~
$ in2un use /path/to/insights/archive

$ in2un get pods -n openshift-cluster-version
WARN[0000] Hint: use --api-version and --kind to override dummy values for missing fields in insights archives 
NAMESPACE                   NAME                                        AGE
openshift-cluster-version   cluster-version-operator-abc123-xyz45       1d

$ in2un logs -n openshift-cluster-version  cluster-version-operator-abc123-xyz45|head -1
I0313 15:23:46.179783       1 start.go:23] ClusterVersionOperator 4.16.0-202410011135.p0.g617769f.assembly.stream.el9-617769f

$ in2un get cm -n kube-system 
NAMESPACE     NAME                AGE
kube-system   cluster-config-v1   <unknown>
~~~

### Printing format

Printing options are limited to the default table output (namespace/name/age) or json/yaml format. Further object-specific pretty printing can be achieved using tools with richer printing capabilities (e.g. [koff](https://github.com/gmeghnag/koff)):

~~~
$ in2un get clusteroperator network
NAME      AGE
network   1d

$ in2un get clusteroperator network -o yaml|koff
NAME                                          VERSION   AVAILABLE   PROGRESSING   DEGRADED   SINCE
clusteroperator.config.openshift.io/network   4.16.40   True        False         False      1d
~~~


### Handling missing fields

This library processes data based on the available information. Some raw objects in an Insights archive may lack the `apiVersion` and `kind` fields, which are essential for parsing into Kubernetes unstructured types. To ensure proper parsing, the library injects DUMMY values for these fields. Users can override these dummy values using the `--api-version` and `--kind flags`, which is required by printing tools to generate the proper output table columns.

As an example, `Pod` resource files are lacking its apiVersion and Kind fields,
so dummy values are injected:

~~~
$ in2un get pods -n openshift-cluster-version -o yaml|head -6
WARN[0000] Hint: use --api-version and --kind to override dummy values for missing fields in insights archives 
apiVersion: v1
items:
- apiVersion: DUMMY
  kind: DUMMY
    metadata:
        annotations:
~~~

Without these fields correctly set, the default *table output* is used:

~~~
$ in2un get pods -n openshift-cluster-version
WARN[0000] Hint: use --api-version and --kind to override dummy values for missing fields in insights archives 
NAMESPACE                   NAME                                        AGE
openshift-cluster-version   cluster-version-operator-abc123-xyz45       1d
~~~

To ensure the proper output is printed, the missing fields can be overriden with the `--api-version` and `--kind` flags:

~~~
$ in2un get pods -n openshift-cluster-version -o yaml --api-version=v1 --kind=Pod|head -6
apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
    metadata:
        annotations:
~~~

When piped to a Kubernetes printer tool capable of handling yaml input:

~~~
$ omc insights get pods -n openshift-cluster-version -o yaml --api-version=v1 --kind=Pod|koff
NAME                                            READY   STATUS    RESTARTS   AGE
pod/cluster-version-operator-abc123-xyz45       1/1     Running   0          1d
~~~

## Building

Running `make bin` builds the CLI in `build/` directory.

See the [Makefile](Makefile) for further details.

