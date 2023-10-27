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

package controllers

import (
	"context"
	"strconv"
	"time"

	ocpappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"

	proxyv1alpha1 "github.com/kenmoini/proxy-config-operator/api/v1alpha1"
)

// ProxyConfigReconciler reconciles a ProxyConfig object
type ProxyConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=proxy.k8s.kemo.dev,resources=proxyconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=proxy.k8s.kemo.dev,resources=proxyconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=proxy.k8s.kemo.dev,resources=proxyconfigs/finalizers,verbs=update

//+kubebuilder:rbac:groups=config.openshift.io,resources=proxies,verbs=get;list;watch
//+kubebuilder:rbac:groups=config.openshift.io,resources=proxies/status,verbs=get

//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=get;list;watch;update;patch

//

// ===========================================================================================
// INIT VARS
// ===========================================================================================
// Implement reconcile.Reconciler so the controller can reconcile objects
// var lggr = log.Log.WithName("proxy-config-controller")
// var SetLogLevel int
var lggr = ctrl.Log.WithName("proxy-config-controller")
var scanningInterval = 30

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ProxyConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
// func (r *ProxyConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
func (r *ProxyConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Get the current connection configuration
	currentConfig, _ := config.GetConfig()
	clusterEndpoint := currentConfig.Host
	apiPath := currentConfig.APIPath
	lggr.Info("Connected to: " + clusterEndpoint + " | API Path:" + apiPath)

	// Set up new Client
	cl, err := client.New(currentConfig, client.Options{Scheme: r.Scheme})
	if err != nil {
		lggr.Error(err, "Failed to create client")
		lggr.Info("Running reconciler again in " + strconv.Itoa(scanningInterval) + "s")
		time.Sleep(time.Second * time.Duration(scanningInterval))
		return ctrl.Result{}, err
	}

	// Fetch the proxyConfig instance that we're reconciling
	proxyConfig := &proxyv1alpha1.ProxyConfig{}
	err = r.Get(ctx, req.NamespacedName, proxyConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			lggr.Error(err, "proxyConfig resource not found on the cluster.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		lggr.Error(err, "Failed to get proxyConfig")
		return ctrl.Result{}, err
	}
	// Detect what type of proxySource we're using
	proxySource := SetDefaultString(DEFAULT_PROXY_SOURCE, proxyConfig.Spec.ProxySource)

	// Log out the proxyConfig metadata
	lggr.Info("proxyConfig found in '" + proxyConfig.ObjectMeta.Namespace + "/" + proxyConfig.ObjectMeta.Name + "', proxySource: " + proxySource)

	// Set up the proxy variables
	var httpProxy string
	var httpsProxy string
	var noProxy string
	var injectCACert bool

	// This can be overridden by the label on a workload
	//caCertConfigMapName := PROXY_CA_CERT_CONFIGMAP_DEFAULT_NAME

	// Switch based on proxySource types
	if proxySource == "openshift" {
		// Check to see if this is a SNO instance - just because
		IsOpenshiftSno, err := IsOpenshiftSno(cl, lggr)
		if err != nil {
			lggr.Error(err, "Failed to determine if this is a SNO instance")
		} else {
			lggr.Info("IsOpenshiftSno: " + strconv.FormatBool(IsOpenshiftSno))
		}

		// Get the OpenShift Cluster Proxy Configuration
		clusterProxyConfig, err := getOpenShiftClusterProxyConfiguration(cl, lggr)
		if err != nil {
			lggr.Error(err, "Failed to get OpenShift Cluster Proxy Configuration")
		} else {
			// Set the Proxy variables
			httpProxy = SetDefaultString("", clusterProxyConfig.Status.HTTPProxy)
			httpsProxy = SetDefaultString("", clusterProxyConfig.Status.HTTPSProxy)
			noProxy = SetDefaultString("", clusterProxyConfig.Status.NoProxy)

			// Check if there is a trustedCA defined in the OpenShift proxy config
			if clusterProxyConfig.Spec.TrustedCA.Name != "" && proxyConfig.Spec.InjectCACert {
				// We don't need to get the name of the CA Certificate ConfigMap since:
				// 1. The ConfigMap is in the openshift-config namespace
				// 2. The ConfigMap can be generated with the proper label
				// 3. We just need to know if we're injecting it into workloads at this point
				injectCACert = true
			}
		}
	} else {
		// Set the proxy variables
		httpProxy = SetDefaultString("", proxyConfig.Spec.Proxy.HTTPProxy)
		httpsProxy = SetDefaultString("", proxyConfig.Spec.Proxy.HTTPSProxy)
		noProxy = SetDefaultString("", proxyConfig.Spec.Proxy.NoProxy)
	}

	lggr.Info("httpProxy: " + httpProxy)
	lggr.Info("httpsProxy: " + httpsProxy)
	lggr.Info("noProxy: " + noProxy)
	proxyObj := proxyv1alpha1.Proxy{HTTPProxy: httpProxy, HTTPSProxy: httpsProxy, NoProxy: noProxy}
	lggr.Info("injectCACert: " + strconv.FormatBool(injectCACert))

	// Find the workloads that have the label to inject the proxy configuration
	listOpts := []client.ListOption{
		client.InNamespace(proxyConfig.ObjectMeta.Namespace),
		client.MatchingLabels(map[string]string{
			"proxy.k8s.kemo.dev/inject-proxy-env": "true",
		}),
	}

	deploymentList := &appsv1.DeploymentList{}
	deploymentConfigList := &ocpappsv1.DeploymentConfigList{}
	statefulSetList := &appsv1.StatefulSetList{}
	daemonSetList := &appsv1.DaemonSetList{}
	jobList := &batchv1.JobList{}
	cronJobList := &batchv1.CronJobList{}

	// Get the Deployments
	if err = cl.List(ctx, deploymentList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list Deployments in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(deploymentList.Items)) + " Deployments")

		for _, deployment := range deploymentList.Items {
			// Set the Proxy Secret Name
			proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, deployment.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
			// Create the Proxy Secret
			err = createWorkloadProxySecret(proxySecretName, deployment.ObjectMeta.Namespace, proxyObj, "Deployment", cl, ctx, lggr)
			if err != nil {
				lggr.Error(err, "Failed to create Proxy Secret")
			} else {
				// Loop through the containers and update the environmental variables
				for i := range deployment.Spec.Template.Spec.Containers {
					currentEnvVars := deployment.Spec.Template.Spec.Containers[i].Env
					updatedEnvVars := createWorkloadEnvVariables(currentEnvVars, proxySecretName, proxyObj)
					deployment.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
				}

				err = cl.Update(ctx, &deployment)
				if err != nil {
					lggr.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
				} else {
					lggr.Info("Updated Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
				}
			}
		}
	}

	// Get the DeploymentConfigs
	if err = cl.List(ctx, deploymentConfigList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list DeploymentConfigs in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(deploymentConfigList.Items)) + " DeploymentConfigs")

		for _, deploymentConfig := range deploymentConfigList.Items {
			// Set the Proxy Secret Name
			proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, deploymentConfig.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
			// Create the Proxy Secret
			err = createWorkloadProxySecret(proxySecretName, deploymentConfig.ObjectMeta.Namespace, proxyObj, "DeploymentConfig", cl, ctx, lggr)
			if err != nil {
				lggr.Error(err, "Failed to create Proxy Secret")
			} else {
				// Loop through the containers and update the environmental variables
				for i := range deploymentConfig.Spec.Template.Spec.Containers {
					currentEnvVars := deploymentConfig.Spec.Template.Spec.Containers[i].Env
					updatedEnvVars := createWorkloadEnvVariables(currentEnvVars, proxySecretName, proxyObj)
					deploymentConfig.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
				}

				err = cl.Update(ctx, &deploymentConfig)
				if err != nil {
					lggr.Error(err, "Failed to update DeploymentConfig", "DeploymentConfig.Namespace", deploymentConfig.Namespace, "DeploymentConfig.Name", deploymentConfig.Name)
				} else {
					lggr.Info("Updated DeploymentConfig", "DeploymentConfig.Namespace", deploymentConfig.Namespace, "DeploymentConfig.Name", deploymentConfig.Name)
				}
			}
		}
	}

	// Get the StatefulSets
	if err = cl.List(ctx, statefulSetList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list StatefulSets in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(statefulSetList.Items)) + " StatefulSets")

		for _, statefulSet := range statefulSetList.Items {
			// Set the Proxy Secret Name
			proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, statefulSet.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
			// Create the Proxy Secret
			err = createWorkloadProxySecret(proxySecretName, statefulSet.ObjectMeta.Namespace, proxyObj, "StatefulSet", cl, ctx, lggr)
			if err != nil {
				lggr.Error(err, "Failed to create Proxy Secret")
			} else {
				// Loop through the containers and update the environmental variables
				for i := range statefulSet.Spec.Template.Spec.Containers {
					currentEnvVars := statefulSet.Spec.Template.Spec.Containers[i].Env
					updatedEnvVars := createWorkloadEnvVariables(currentEnvVars, proxySecretName, proxyObj)
					statefulSet.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
				}

				err = cl.Update(ctx, &statefulSet)
				if err != nil {
					lggr.Error(err, "Failed to update StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
				} else {
					lggr.Info("Updated StatefulSet", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
				}
			}
		}
	}

	// Get the DaemonSets
	if err = cl.List(ctx, daemonSetList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list DaemonSets in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(daemonSetList.Items)) + " DaemonSets")

		for _, daemonSet := range daemonSetList.Items {
			// Set the Proxy Secret Name
			proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, daemonSet.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
			// Create the Proxy Secret
			err = createWorkloadProxySecret(proxySecretName, daemonSet.ObjectMeta.Namespace, proxyObj, "DaemonSet", cl, ctx, lggr)
			if err != nil {
				lggr.Error(err, "Failed to create Proxy Secret")
			} else {
				// Loop through the containers and update the environmental variables
				for i := range daemonSet.Spec.Template.Spec.Containers {
					currentEnvVars := daemonSet.Spec.Template.Spec.Containers[i].Env
					updatedEnvVars := createWorkloadEnvVariables(currentEnvVars, proxySecretName, proxyObj)
					daemonSet.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
				}

				err = cl.Update(ctx, &daemonSet)
				if err != nil {
					lggr.Error(err, "Failed to update DaemonSet", "DaemonSet.Namespace", daemonSet.Namespace, "DaemonSet.Name", daemonSet.Name)
				} else {
					lggr.Info("Updated DaemonSet", "DaemonSet.Namespace", daemonSet.Namespace, "DaemonSet.Name", daemonSet.Name)
				}
			}
		}
	}

	// Get the Jobs
	if err = cl.List(ctx, jobList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list Jobs in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(jobList.Items)) + " Jobs")
	}

	// Get the CronJobs
	if err = cl.List(ctx, cronJobList, listOpts...); err != nil {
		lggr.Error(err, "Failed to list CronJobs in "+proxyConfig.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	} else {
		lggr.Info("Found " + strconv.Itoa(len(cronJobList.Items)) + " CronJobs")

		for _, cronJob := range cronJobList.Items {
			// Set the Proxy Secret Name
			proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, cronJob.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
			// Create the Proxy Secret
			err = createWorkloadProxySecret(proxySecretName, cronJob.ObjectMeta.Namespace, proxyObj, "CronJob", cl, ctx, lggr)
			if err != nil {
				lggr.Error(err, "Failed to create Proxy Secret")
			} else {
				// Loop through the containers and update the environmental variables
				for i := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers {
					currentEnvVars := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Env
					updatedEnvVars := createWorkloadEnvVariables(currentEnvVars, proxySecretName, proxyObj)
					cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Env = updatedEnvVars
				}

				err = cl.Update(ctx, &cronJob)
				if err != nil {
					lggr.Error(err, "Failed to update CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
				} else {
					lggr.Info("Updated CronJob", "CronJob.Namespace", cronJob.Namespace, "CronJob.Name", cronJob.Name)
				}
			}
		}
	}

	// Create a ConfigMap
	//createOpenShiftCACertConfigMap(cl, ctx, lggr, caCertConfigMapName, proxyConfig.ObjectMeta.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProxyConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&proxyv1alpha1.ProxyConfig{}).
		Complete(r)
}
