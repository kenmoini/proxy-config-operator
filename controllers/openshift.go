package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsOpenshiftSno(c client.Client, log logr.Logger) (bool, error) {
	infra := &configv1.Infrastructure{}

	defaultInfraName := "cluster"
	err := c.Get(context.TODO(), types.NamespacedName{Name: defaultInfraName}, infra)
	if err != nil {
		return false, fmt.Errorf("getting resource Infrastructure (name: %s) succeeded but object was empty", defaultInfraName)
	}
	log.Info("OCP cluster infrastructure", "infra", infra.Status.ControlPlaneTopology)
	return infra.Status.ControlPlaneTopology == configv1.SingleReplicaTopologyMode, nil
}

func getOpenShiftClusterProxyConfiguration(cl client.Client, log logr.Logger) (configv1.Proxy, error) {
	// Get the cluster proxy config
	clusterProxyConfig := &configv1.Proxy{}

	if err := cl.Get(context.TODO(), OpenShiftProxy(), clusterProxyConfig); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			lggr.Error(err, "OpenShift Cluster Proxy resource not found on the cluster.")
			return configv1.Proxy{}, err
		}
		// Error reading the object - requeue the request.
		return configv1.Proxy{}, err
	} else {
		return *clusterProxyConfig, nil
	}
}

func createOpenShiftCACertConfigMap(cl client.Client, ctx context.Context, log logr.Logger, configMapName string, configMapNamespace string) {
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Labels: map[string]string{
				"config.openshift.io/inject-trusted-cabundle": "true",
			},
		},
	}

	err := cl.Create(ctx, &cm)
	if err != nil {
		log.Error(err, "Failed to create ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
	} else {
		log.Info("Created ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
	}
}
