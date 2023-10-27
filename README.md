# Proxy Configuration Operator

The Proxy Configuration Operator allows easy configuration of workloads to be used with Outbound Proxys.

## Problem

In OpenShift, the Cluster-wide Outbound Proxy Configuration is only available to operators and workloads that are developed to consume it when set.

For user workloads, you must define these environmental variables manually, which produces more labor.

## Solution

You can inject the cluster-wide additionalTrustBundle otherwise known as the trusted root CA system store via a blank ConfigMap with the label `config.openshift.io/inject-trusted-cabundle="true"` which you can then mount to a workload.  Doing the same for Outbound Proxy configuration would be ideal.

