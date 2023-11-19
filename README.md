# Ingress-Target-Updater
A Kubernetes CronJob to keep updated the external-dns target annotation with your public IPv4.

## How it works
This CronJob will check your public IPv4 and update the `external-dns.alpha.kubernetes.io/target` annotation of your Ingresses with it.

## How to use it
You just need to apply the kubernetes directory to your cluster and it will create the CronJob and the ServiceAccount with the required permissions. It will also create a namespace called `ingress-target-updater` where the kubernetes resources will be deployed.
