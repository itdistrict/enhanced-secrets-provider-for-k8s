package k8s

import (
	"context"
	"regexp"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cyberark/secrets-provider-for-k8s/pkg/log/messages"
)

type RetrieveK8sSecretFunc func(namespace string, secretName string) (*v1.Secret, error)
type UpdateK8sSecretFunc func(namespace string, secretName string, originalK8sSecret *v1.Secret, stringDataEntriesMap map[string][]byte) error
type RetrieveSecretsListFunc func(regex []string, mapKey string, podNamespace string) ([]string, error)


func RetrieveK8sSecret(namespace string, secretName string) (*v1.Secret, error) {
	// get K8s client object
	kubeClient, _ := configK8sClient()
	log.Info(messages.CSPFK005I, secretName, namespace)
	k8sSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		// Error messages returned from K8s should be printed only in debug mode
		log.Debug(messages.CSPFK004D, err.Error())
		return nil, log.RecordedError(messages.CSPFK020E)
	}

	return k8sSecret, nil
}

func RetrieveSecretsList(regex []string,mapKey string, podNamespace string) ([]string, error) {
	// get K8s client object
	kubeClient, _ := configK8sClient()
	secretsMap := []string{}
	log.Info(messages.CSPFK023I, regex)
	namespaces, err := kubeClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// Error messages returned from K8s should be printed only in debug mode
		log.Debug(messages.CSPFK010D, err.Error())
		list,_ := GetSecretsByNamespace(podNamespace,regex,mapKey)
		secretsMap = append(secretsMap,list...)
		return secretsMap,nil
	}
	for _, namespace := range namespaces.Items {
        list,_ := GetSecretsByNamespace(namespace.Name,regex,mapKey)
		secretsMap = append(secretsMap,list...)
    }
	if (len(secretsMap) == 0){
		log.Debug(messages.CSPFK011D)
	}
	return secretsMap, nil
}

func GetSecretsByNamespace(namespace string, regex []string, mapKey string)([]string, error){
	kubeClient, _ := configK8sClient()
	secretsMap := []string{}
	k8sSecretList, _ := kubeClient.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	for _, secret := range k8sSecretList.Items {
		log.Debug("SecretName: '%s'",secret)
		k8sSecret, err := RetrieveK8sSecret(namespace,secret.Name)
		if err == nil {
			conjurMapKey := mapKey
			_ , entryExists := k8sSecret.Data[conjurMapKey]
			if entryExists {
				if (regex[0] == ""){
					secretsMap = append(secretsMap,secret.Name)
				}else{
					for _, reg := range regex{
						res1, _ := regexp.MatchString(reg, secret.Name)
						if res1{
							secretsMap = append(secretsMap,secret.Name)
							break
						}
					}
				}
			}
		}
	}
	return secretsMap,nil
}

func UpdateK8sSecret(namespace string, secretName string, originalK8sSecret *v1.Secret, stringDataEntriesMap map[string][]byte) error {
	// get K8s client object
	kubeClient, _ := configK8sClient()

	for secretName, secretValue := range stringDataEntriesMap {
		originalK8sSecret.Data[secretName] = secretValue
	}

	log.Info(messages.CSPFK006I, secretName, namespace)
	_, err := kubeClient.CoreV1().Secrets(namespace).Update(context.Background(), originalK8sSecret, metav1.UpdateOptions{})
	// Clear secret from memory
	stringDataEntriesMap = nil
	originalK8sSecret = nil
	if err != nil {
		// Error messages returned from K8s should be printed only in debug mode
		log.Debug(messages.CSPFK005D, err.Error())
		return log.RecordedError(messages.CSPFK022E)
	}

	return nil
}

func configK8sClient() (*kubernetes.Clientset, error) {
	// Create the Kubernetes client
	log.Info(messages.CSPFK004I)
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		// Error messages returned from K8s should be printed only in debug mode
		log.Debug(messages.CSPFK002D, err.Error())
		return nil, log.RecordedError(messages.CSPFK019E)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		// Error messages returned from K8s should be printed only in debug mode
		log.Debug(messages.CSPFK003D, err.Error())
		return nil, log.RecordedError(messages.CSPFK018E)
	}
	// return a K8s client
	return kubeClient, err
}
