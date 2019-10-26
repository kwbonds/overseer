# Changelog

## [2019/10/26] cmaster11/overseer:release-1.11-4

* Added support for queue-bridge filters. Destination queues (`-dest-queue=overseer.results.email`) can now be filtered
by various test result tags: `-dest-queue=overseer.results.cronjobs[type=k8s-event,target=my-namespace/Job]`. More in 
the [Kubernetes example](example-kubernetes/overseer-bridge-queue.optional.yaml) and 
[source code](bridges/queue-bridge/filter.go).