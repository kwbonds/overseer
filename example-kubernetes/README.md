# Example Kubernetes deployment

This folder contains a full Kubernetes deployment example.

It will generate:
 
* An `overseer` Kubernetes namespace.
* A service account, which will let `overseer` observe services in the k8s cluster.
* An [`overseer-worker`](overseer-worker.yaml) deployment, to process tests to execute.
* An [`overseer-bridge-webhook-n17`](overseer-bridge-webhook-n17.yaml) deployment, to notify errors using [Notify17](https://notify17.net).
* A [`CronJob`](overseer-enqueue.yaml) that will periodically enqueue the tests you want to run.
* A simple [`Redis`](https://redis.io/) deployment, to hold test queues and results.

## Install

* Create a [notification template](https://notify17.net/docs/templates/) using [Notify17's dashboard](https://dash.notify17.net/#/notificationTemplates), 
and replace the `REPLACE_TEMPLATE_API_KEY` string in [`overseer-bridge-webhook-n17.yaml`](overseer-bridge-webhook-n17.yaml) with the template's API key.
* Run `kubectl apply -f .` to create all `overseer` resources.

## Scripts

* [`enqueue.sh`](./enqueue.sh) can be used to manually enqueue a test:

    `./enqueue.sh "https://google.com must run ssl`
    
* [`test-cron.sh`](./test-cron.sh) can be used to manually trigger the enqueuing of all tests defined in [`overseer-enqueue.yaml`](./overseer-enqueue.yaml)'s `CronJob`.

## Delete

* Run `kubectl delete -f .` to destroy all `overseer` resources.