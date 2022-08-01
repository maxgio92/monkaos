package victims

import (
	"context"
	"fmt"
	"math/rand"
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
func GetPodVictims(count int, excludedNamespaces []string, strategy Strategy) (eligibleVictims []v1.Pod, err error) {
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

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := 0
	retryCount := 0

	for i < count {
		randomIndex := r.Intn(len(namespaces))
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
		randomIndex = r.Intn(len(pods.Items))
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
