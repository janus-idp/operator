# DEVELOPER GUIDE --- WIP

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.



### Running on the cluster
1. Install Instances of Custom Resources:
```sh
kubectl apply -f config/samples/
```
2. Build and push your image to the location specified by `IMG`:
```sh
make docker-build docker-push IMG=<some-registry>/backstage-operator:tag
```
3. Deploy the controller to the cluster with the image specified by `IMG`:
```sh
make deploy IMG=<some-registry>/backstage-operator:tag
```
### Uninstall CRDs
To delete the CRDs from the cluster:
```sh
make uninstall
```
### Undeploy controller
UnDeploy the controller from the cluster:
```sh
make undeploy
```
## Build and Deploy with OLM
1. To build operator, bundle and catalog images:
```sh
make release-build
```
2. To push operator, bundle and catalog images to the registry:
```sh
make release-push
```
3. To deploy or update catalog source:
```sh
make catalog-update
```
4. To deloy the operator with OLM
```sh
make deploy-olm
```
4. To undeloy the operator with OLM
```sh
make undeploy-olm
```

### Test It Out
1. Install the CRDs into the cluster:
```sh
make install
```
2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):
```sh
make run
```
**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:
```sh
make manifests
```
**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
