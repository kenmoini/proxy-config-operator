/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProxyConfigSpec defines the desired state of ProxyConfig
type ProxyConfigSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// ProxySource defines the source of the proxy configuration
	// Options include:
	// - "openshift" (default): Use the proxy configuration from the OpenShift cluster
	// - "custom": Use the proxy configuration defined in the ProxyConfig resource
	// +kubebuilder:validation:Enum=openshift;custom
	ProxySource string `json:"proxySource,omitempty"`

	// InjectCACert defines whether to inject the CA certificate into the workloads.
	// When proxySource is set to "openshift", it will use the cluster-wide additionalTrustBundle defined in
	//  the proxy.config.openshift.io/cluster resource and provided by the Cluster Network Operator.
	// When proxySource is set to "custom", it will use the CA certificate defined in the ProxyConfig resource.
	// +optional
	InjectCACert bool `json:"injectCACert,omitempty"`

	// Proxy defines the proxy configuration to use when ProxySource is set to "custom"
	// +optional
	Proxy Proxy `json:"proxy,omitempty"`
}

// Proxy defines the proxy configuration to use when ProxySource is set to "custom"
type Proxy struct {
	// HTTPProxy defines the HTTP proxy to use
	// +optional
	HTTPProxy string `json:"httpProxy,omitempty"`
	// HTTPSProxy defines the HTTPS proxy to use
	// +optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	// NoProxy defines the no proxy configuration to use
	// +optional
	NoProxy string `json:"noProxy,omitempty"`
	// CACert defines the CA certificate stored in a ConfigMap to use
	// +optional
	CAConfig CAConfig `json:"caConfig,omitempty"`
}

// CAConfig defines the CA configuration to use
type CAConfig struct {
	// Name defines the CA certificate stored in a ConfigMap to use
	// +optional
	Name string `json:"name,omitempty"`
	// Key defines the key of the CA certificate stored in a ConfigMap to use
	// +optional
	Key string `json:"key,omitempty"`
}

// ProxyConfigStatus defines the observed state of ProxyConfig
type ProxyConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProxyConfig is the Schema for the proxyconfigs API
type ProxyConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxyConfigSpec   `json:"spec,omitempty"`
	Status ProxyConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxyConfigList contains a list of ProxyConfig
type ProxyConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyConfig{}, &ProxyConfigList{})
}
