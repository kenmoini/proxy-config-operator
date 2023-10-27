package controllers

import (
	"context"

	"github.com/go-logr/logr"
	proxyv1alpha1 "github.com/kenmoini/proxy-config-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//// LogWithLevel implements simple log levels
//func LogWithLevel(s string, level int, l logr.Logger) {
//	if SetLogLevel >= level {
//		l.Info(s)
//	}
//}

func createProxySecret(cl client.Client, ctx context.Context, log logr.Logger, secretName string, secretNamespace string, proxyConfig proxyv1alpha1.Proxy) error {
	secretCheck := corev1.Secret{}

	// Check to see if the secret already exists
	err := cl.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, &secretCheck)
	if err == nil {
		// Secret already exists
		// Check to see if it needs to be updated
		if string(secretCheck.Data["http_proxy"]) != proxyConfig.HTTPProxy || string(secretCheck.Data["https_proxy"]) != proxyConfig.HTTPSProxy || string(secretCheck.Data["no_proxy"]) != proxyConfig.NoProxy {
			secretCheck.Data = map[string][]byte{
				"http_proxy":  []byte(proxyConfig.HTTPProxy),
				"https_proxy": []byte(proxyConfig.HTTPSProxy),
				"no_proxy":    []byte(proxyConfig.NoProxy),
			}
			err = cl.Update(ctx, &secretCheck)
			if err != nil {
				log.Error(err, "Failed to update Secret", "Secret.Namespace", secretCheck.Namespace, "Secret.Name", secretCheck.Name)
				return err
			} else {
				log.Info("Reconciled Secret", "Secret.Namespace", secretCheck.Namespace, "Secret.Name", secretCheck.Name)
				return nil
			}
		} else {
			log.Info("Secret already exists and is up to date", "Secret.Namespace", secretCheck.Namespace, "Secret.Name", secretCheck.Name)
			return nil
		}
	} else if errors.IsNotFound(err) {
		// Secret doesn't exist, create it
		secret := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: secretNamespace,
			},
			Data: map[string][]byte{
				"http_proxy":  []byte(proxyConfig.HTTPProxy),
				"https_proxy": []byte(proxyConfig.HTTPSProxy),
				"no_proxy":    []byte(proxyConfig.NoProxy),
			},
		}

		err := cl.Create(ctx, &secret)
		if err != nil {
			log.Error(err, "Failed to create Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
			return err
		} else {
			log.Info("Created Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
			return nil
		}
	}
	return nil
}

func createWorkloadProxySecret(proxySecretName string, namespace string, proxyObj proxyv1alpha1.Proxy, workloadType string, cl client.Client, ctx context.Context, log logr.Logger) error {
	// Set the Proxy Secret Name
	//proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, pod.ObjectMeta.Labels[PROXY_INJECTION_SECRET_LABEL])
	//proxySecretName := SetDefaultString(PROXY_INJECTION_SECRET_DEFAULT_NAME, secretNameLabelOverride)

	// Create the Proxy Secret
	err := createProxySecret(cl, ctx, lggr, proxySecretName, namespace, proxyObj)
	if err != nil {
		lggr.Error(err, "Failed to create Proxy Secret for "+workloadType+" in "+namespace+" Secret Name "+proxySecretName)
		return err
	} else {
		lggr.Info("Created Proxy Secret for " + workloadType + " in " + namespace + " Secret Name " + proxySecretName)
		return nil
	}
}

func createOrUpdateEnvironmentVariable(envVars []corev1.EnvVar, envVarKey string, envVarValueFrom corev1.EnvVarSource) []corev1.EnvVar {
	// Loop through the envVars
	// If the envVarKey exists, update it
	// If the envVarKey doesn't exist, append it

	for i, e := range envVars {
		if e.Name == envVarKey {
			envVars[i] = corev1.EnvVar{Name: envVarKey, ValueFrom: &envVarValueFrom}
			return envVars
		}
	}
	envVar := corev1.EnvVar{Name: envVarKey, ValueFrom: &envVarValueFrom}
	envVars = append(envVars, envVar)
	return envVars
}

func createWorkloadEnvVariables(currentEnvVars []corev1.EnvVar, proxySecretName string, proxyObj proxyv1alpha1.Proxy) []corev1.EnvVar {

	// Add the proxy environmental variables
	if proxyObj.HTTPProxy != "" {
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "HTTP_PROXY", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "http_proxy"}})
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "http_proxy", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "http_proxy"}})
	}
	if proxyObj.HTTPSProxy != "" {
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "HTTPS_PROXY", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "https_proxy"}})
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "https_proxy", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "https_proxy"}})
	}
	if proxyObj.NoProxy != "" {
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "NO_PROXY", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "no_proxy"}})
		currentEnvVars = createOrUpdateEnvironmentVariable(currentEnvVars, "no_proxy", corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: proxySecretName}, Key: "no_proxy"}})
	}

	return currentEnvVars

}
