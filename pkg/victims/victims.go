package victims

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"monkaos/pkg/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
	k "k8s.io/client-go/kubernetes"
)

type Strategy string

const (
	retryDelaySeconds                         = 1
	maxRetryCount                             = 10
	RandomPodRandomNamespaceStrategy Strategy = "RandomPodRandomNamespace"
)

// GetPodVictims gathers list of `count` pods, from all
// namespaces expect the ones specified in the exclude list.
func GetPodVictims(count int, excludedNamespaces []string, strategy Strategy) ([]v1.Pod, error) {
	clientset, err := kubernetes.NewClientset()
	if err != nil {
		return nil, err
	}

	var victims []v1.Pod

	switch strategy {
	case RandomPodRandomNamespaceStrategy:
		victims, err = GetRandomPods(clientset, count, excludedNamespaces)
		if err != nil {
			return nil, err
		}
	default:
		victims, err = GetRandomPods(clientset, count, excludedNamespaces)
		if err != nil {
			return nil, err
		}
	}

	return victims, nil
}

//nolint:funlen
func GetRandomPods(clientset k.Interface, count int, excludedNamespaces []string) ([]v1.Pod, error) {
	var randomPods []v1.Pod

	allNamespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces, err := getEligibleNamespaces(allNamespaces.Items, excludedNamespaces)
	if err != nil {
		return nil, err
	}

	i := 0
	retryCount := 0

	for i < count {
		randomIndex, err := getRandomInt(len(namespaces))
		if err != nil {
			return nil, err
		}

		if randomIndex < 1 {
			randomIndex++
		}

		// Take a random namespace.
		randomNamespace := namespaces[randomIndex-1]

		// Get all the pods from the selected namespace.
		pods, err := clientset.CoreV1().Pods(randomNamespace.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// If no pods are found in the selected namespace.
		if len(pods.Items) < 1 {
			time.Sleep(time.Duration(retryDelaySeconds) * time.Second)
			if retryCount >= maxRetryCount {
				return randomPods, nil
			}
			retryCount++

			continue
		}

		// Select a random pod from the selected namespace.
		// TODO: ensure the pod has not been already selected.
		randomIndex, err = getRandomInt(len(pods.Items))
		if err != nil {
			return nil, err
		}

		if randomIndex < 1 {
			randomIndex++
		}
		randomPod := pods.Items[randomIndex-1]
		randomPods = append(randomPods, randomPod)
		i++
	}

	return randomPods, nil
}

func getEligibleNamespaces(namespaces []v1.Namespace, excludeList []string) ([]v1.Namespace, error) {
	if len(namespaces) < 1 {
		return nil, fmt.Errorf("error: namespace list is empty")
	}

	//nolint:prealloc
	var eligibleNamespaces []v1.Namespace

	for _, namespace := range namespaces {

		// Skip excluded Namespaces
		if sets.NewString(excludeList...).Has(namespace.Name) {
			continue
		}
		eligibleNamespaces = append(eligibleNamespaces, namespace)
	}

	return eligibleNamespaces, nil
}

func getRandomInt(max int) (int, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return int(r.Int64()), nil
}
