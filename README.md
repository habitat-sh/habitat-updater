# Habitat Updater: A service for syncing package update with kubernetes managed habitat services

Habitat updater is a service that runs inside of a k8's cluster to watch for changes in pods running Habitat services and reconcile any updates from Habitat Builder. It does this by querying the k8's pods api for anything with a `habitat=true` label. Then it queries the supervisors running in those pods for what services they are running. Once it has a list of services, it asks builder for the most recent version of those packages in the stable channel. 

## Building
Clone this repository and then:

```
$ cd habitat-updater
$ hab pkg build .
$ hab pkg export docker results/habitat-habitat-updater-<release_version> --push-image --image-name habitat/habitat-updater
```
## Deploying

### GKE

Follow the [instructions](https://cloud.google.com/kubernetes-engine/docs/quickstart) for getting k8's setup on GKE.
You will also need the gcloud tools locally to create and manage your kubernetes config.

GKE runs with RBAC enabled so we need to create a service account. Unfortunately this means we can't deploy our update service with the Habitat Operator.

** Note: You need to run `kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account)` in order to create service accounts in GKE.

```
kubectl apply -f kubernetes/rbac/rbac.yml
kubectl apply -f kubernetes/rbac/updater.yml
```

### Minikube

In environments without RBAC enabled (like minikube) we can leverage the Habitat Operator to deploy and manage our update service.

```
$ kubectl apply -f kubernetes/habitat-operator.yml
```
