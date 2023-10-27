package controllers

import (
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

const (
	// PROXY_INJECTION_LABEL is the label used to determine whether to inject the proxy configuration into the workloads
	// This can be applied to Deployments, DeploymentConfigs, StatefulSets, DaemonSets, Jobs, CronJobs, and Pods
	// +optional
	PROXY_INJECTION_LABEL = "proxy.k8s.kemo.dev/inject-proxy-env"

	// PROXY_INJECTION_SECRET_LABEL is the label used to determine which Secret to use for the proxy configuration
	// Defaults to "proxy-config" which it will generate and maintain
	// +optional
	PROXY_INJECTION_SECRET_LABEL = "proxy.k8s.kemo.dev/proxy-secret-name"

	// PROXY_INJECTION_SECRET_DEFAULT_NAME is the default name of the Secret to use for the proxy configuration
	// Defaults to "proxy-config"
	// +optional
	PROXY_INJECTION_SECRET_DEFAULT_NAME = "proxy-config"

	// PROXY_CA_CERT_INJECTION_LABEL is the label used to determine whether to inject the CA certificate into the workloads
	// This can be applied to Deployments, DeploymentConfigs, StatefulSets, DaemonSets, Jobs, CronJobs, and Pods.
	// When set to "true", it will inject the CA certificate into the workloads via a ConfigMap volume mount.
	// +optional
	PROXY_CA_CERT_INJECTION_LABEL = "proxy.k8s.kemo.dev/inject-ca-cert"

	// PROXY_CA_CERT_CONFIGMAP_LABEL is the label used to determine which ConfigMap to use for the CA certificate
	// Defaults to "proxy-ca-cert" which it will generate and maintain
	// +optional
	PROXY_CA_CERT_CONFIGMAP_LABEL = "proxy.k8s.kemo.dev/ca-cert-configmap-name"

	// PROXY_CA_CERT_CONFIGMAP_KEY_LABEL is the label used to determine which key of the ConfigMap to use for the CA certificate
	// Defaults to "ca-bundle.crt" which it will generate and maintain
	// +optional
	PROXY_CA_CERT_CONFIGMAP_KEY_LABEL = "proxy.k8s.kemo.dev/ca-cert-configmap-key"

	// PROXY_CA_CERT_CONFIGMAP_DEFAULT_NAME is the default name of the ConfigMap to use for the CA certificate
	// Defaults to "proxy-ca-cert"
	// +optional
	PROXY_CA_CERT_CONFIGMAP_DEFAULT_NAME = "proxy-ca-cert"

	// PROXY_CA_CERT_CONFIGMAP_DEFAULT_KEY is the default key of the ConfigMap to use for the CA certificate
	// Defaults to "ca-bundle.crt"
	// +optional
	PROXY_CA_CERT_CONFIGMAP_DEFAULT_KEY = "ca-bundle.crt"

	// PROXY_CA_CERT_MOUNT_PATH_LABEL is the label used to determine which mount path to use for the CA certificate
	// Defaults to "/etc/pki/ca-trust/extracted/pem"
	// +optional
	PROXY_CA_CERT_MOUNT_PATH_LABEL = "proxy.k8s.kemo.dev/ca-cert-mount-path"

	// PROXY_CA_CERT_MOUNT_PATH is the default mount path to use for the CA certificate
	// Defaults to "/etc/pki/ca-trust/extracted/pem"
	// +optional
	PROXY_CA_CERT_MOUNT_PATH = "/etc/pki/ca-trust/extracted/pem"

	DEFAULT_PROXY_SOURCE = "openshift"
)

// OpenShiftProxy returns the namespaced name "cluster" in the
// default namespace.
func OpenShiftProxy() types.NamespacedName {
	return types.NamespacedName{
		Name: "cluster",
	}
}

// SetDefaultInt64 will return either the default int64 or an overriden value
func SetDefaultInt64(defaultVal int64, overrideVal int64) int64 {
	if overrideVal == 0 {
		return defaultVal
	}
	return overrideVal
}

// SetDefaultInt32 will return either the default int32 or an overriden value
func SetDefaultInt32(defaultVal int32, overrideVal int32) int32 {
	iString := strings.TrimSpace(I32ToString(overrideVal))
	if overrideVal == 0 {
		return defaultVal
	}
	if len(iString) > 0 {
		return overrideVal
	}
	return defaultVal
}

// SetDefaultInt will return either the default int or an overriden value
func SetDefaultInt(defaultVal int, overrideVal int) int {
	if overrideVal == 0 {
		return defaultVal
	}
	if len(strings.TrimSpace(strconv.Itoa(overrideVal))) > 0 {
		return overrideVal
	}
	return defaultVal
}

// SetDefaultString will return either the default string or an overriden value
func SetDefaultString(defaultVal string, overrideVal string) string {
	if len(strings.TrimSpace(overrideVal)) > 0 {
		return overrideVal
	}
	return defaultVal
}

// SetDefaultBool will return either the default bool or an overriden value
func SetDefaultBool(defaultVal *bool, overrideVal *bool) *bool {
	if overrideVal != nil {
		return overrideVal
	}
	return defaultVal
}

// I32ToString will convert an int32 to a string for length comparison
func I32ToString(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

// ContainsString checks if a string is present in a slice
func ContainsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
