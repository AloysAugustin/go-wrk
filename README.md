# go-wrk

A variant of wrk, implemented in go, well suited to benchmark Kubernetes services.

This version does constant-rate load testing, and thanks to go's excellent concurrency support, does not suffer from coordinated omission, removing the need for hacks that attempt to correct it afterwards.

It also allows to do requests randomly against a set of URLs and not just one.
