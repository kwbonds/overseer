# Example Kubernetes deployment

This folder contains a full Kubernetes deployment example.

It contains:
 
* An `overseer` Kubernetes namespace.
* A service account, which will let `overseer` observe services in the k8s cluster.
* An [`overseer-worker`](overseer-worker.yaml) deployment, to process tests to execute.
* An [`overseer-bridge-webhook-n17`](overseer-bridge-webhook-n17.yaml) deployment, to notify errors using [Notify17](https://notify17.net).
* An [`overseer-bridge-email`](overseer-bridge-email.yaml.optional) deployment, to notify errors using a standard email SMTP server.
* An [`overseer-bridge-queue`](overseer-bridge-queue.yaml.optional) deployment, to duplicate test results and send them to multiple destination ([read more](#multiple-destinations-eg-notify17-and-email)).
* A [`CronJob`](overseer-enqueue.yaml) that will periodically enqueue the tests you want to run.
* A simple [`Redis`](https://redis.io/) deployment, to hold test queues and results.

## Install

* Use the [Overseer recipe](https://notify17.net/recipes/overseer/) to create a notification template using [Notify17's dashboard](https://dash.notify17.net/#/notificationTemplates), 
and replace the `REPLACE_TEMPLATE_API_KEY` string in [`overseer-bridge-webhook-n17.yaml`](overseer-bridge-webhook-n17.yaml) with the template's API key.
* Run `kubectl apply -f .` to create all `overseer` resources.

### Multiple destinations (e.g. Notify17 AND email)

In the scenario where you want to send your notifications to multiple destinations (e.g. Notify17 AND email), you can use the [`overseer-bridge-queue`](overseer-bridge-queue.yaml.optional) deployment:

* Configure the `-destination-queues` argument to clone test results on how many queues you want (e.g. if you want to send an email, you can create a queue with name `overseer.results.email`).
* Configure the corresponding bridges to use the new queue names by having the `-redis-queue-key` argument match the previously configured one. (e.g. `overseer.results.email`).

A complete scenario can be:

* An **enqueue** cron job to queue the tests.
* A **worker** to process tests and write results (e.g. to standard Redis queue `overseer.results`).
* A **queue bridge** to clone results from `overseer.results` to `overseer.results.email` and `overseer.results.n17`.
* An **email bridge** to send emails using test results stored in `overseer.results.email` queue.
* A **webhook bridge** to send [Notify17](https://notify17.net) notifications, using test results stored in `overseer.results.n17` queue.

## Scripts

* [`enqueue.sh`](./enqueue.sh) can be used to manually enqueue a test:

    `./enqueue.sh "https://google.com must run ssl`
    
* [`test-cron.sh`](./test-cron.sh) can be used to manually trigger the enqueuing of all tests defined in [`overseer-enqueue.yaml`](./overseer-enqueue.yaml)'s `CronJob`.

## Delete

* Run `kubectl delete ns overseer` to destroy all `overseer` resources.