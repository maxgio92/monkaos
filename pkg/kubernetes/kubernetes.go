package kubernetes

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/discovery"
	k "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset creates, verifies and returns an instance of k8 clientset
func NewClientset() (*k.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeConfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
		log.Info("Program running from outside of the cluster")
	} else {
		log.Info("Program running inside the cluster, picking the in-cluster configuration")
	}
	clientset, err := k.NewForConfig(config)
	if err != nil {
		log.Errorf("failed to create clientset in NewForConfig: %v", err)

		return nil, err
	}

	if verifyClient(clientset) {
		return clientset, nil
	}

	return nil, fmt.Errorf("unable to verify client connectivity to Kubernetes apiserver")
}

// NewInClusterClient only creates an initialized instance of k8 clientset
func NewInClusterClient() (*k.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("failed to obtain config from InClusterConfig: %v", err)

		return nil, err
	}

	clientset, err := k.NewForConfig(config)
	if err != nil {
		log.Errorf("failed to create clientset in NewForConfig: %v", err)

		return nil, err
	}

	return clientset, nil
}

func verifyClient(client discovery.DiscoveryInterface) bool {
	_, err := client.ServerVersion()

	return err == nil
}
