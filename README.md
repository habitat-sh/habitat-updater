# Habitat Updater: A service for syncing package update with kubernetes managed habitat services

Habitat updater is a service that runs inside of a k8's cluster to watch for changes in pods running Habitat services and reconcile any updates from Habitat Builder. It does this by querying the k8's pods api for anything with a `habitat=true` label. Then it queries the supervisors running in those pods for what services they are running. Once it has a list of services, it asks builder for the most recent version of those packages in the stable channel. 

## Usage
Clone this repository and then:

```
$ cd habitat-updater
$ hab pkg build .
$ hab pkg export docker results/habitat-habitat-updater-<release_version> --push-image --image-name habitat/habitat-updater
$ kubectl create -f habitat.yml
```
