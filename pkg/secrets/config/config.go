package config

import (
	"os"
	"strings"

	"github.com/cyberark/cyberark-secrets-provider-for-k8s/pkg/log"
	"github.com/cyberark/cyberark-secrets-provider-for-k8s/pkg/log/messages"
)

const (
	K8S                      = "k8s_secrets"
	SecretsDestinationEnvVar = "SECRETS_DESTINATION"
	CONJUR_MAP_KEY           = "conjur-map"
)

// Config defines the configuration parameters
// for the authentication requests
type Config struct {
	PodNamespace       string
	RequiredK8sSecrets []string
	StoreType          string
}

// New returns a new authenticator configuration object
func NewFromEnv() (*Config, error) {

	// Check that required environment variables are set
	for _, envvar := range []string{
		"MY_POD_NAMESPACE",
		"K8S_SECRETS",
	} {
		if os.Getenv(envvar) == "" {
			return nil, log.RecordedError(messages.CSPFK004E, envvar)
		}
	}

	// Load configuration from the environment
	podNamespace := os.Getenv("MY_POD_NAMESPACE")

	// Split the comma-separated list into an array
	requiredK8sSecrets := strings.Split(os.Getenv("K8S_SECRETS"), ",")

	var storeType string
	secretsDestinationValue := os.Getenv(SecretsDestinationEnvVar)
	if secretsDestinationValue == K8S {
		storeType = K8S
	} else if secretsDestinationValue == "" {
		// TODO: decide what to do in this case
		storeType = K8S
	} else {
		// In case SecretsDestinationEnvVar exists and is configured with incorrect value
		return nil, log.RecordedError(messages.CSPFK005E, SecretsDestinationEnvVar)
	}

	return &Config{
		PodNamespace:       podNamespace,
		RequiredK8sSecrets: requiredK8sSecrets,
		StoreType:          storeType,
	}, nil
}
