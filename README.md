# Habitat Updater: A service for syncing package update with kubernetes managed habitat services
This is the repo for a brigade service that syncs package updates between builder and kubernetes.

In it's current state it is *ONLY FOR DEMO PURPOSES*.

## Usage
Clone this repository and then:

```
$ cd habitat-updater
$ hab pkg build .
$ hab pkg export docker results/habitat-habitat-updater-<release_version> --push-image --image-name habitat/habitat-updater
$ kubectl create -f habitat.yml
```