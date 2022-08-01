package victims

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func newPod(name string, namespace string, status corev1.PodPhase) corev1.Pod {

	return corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: corev1.PodStatus{
			Phase: status,
		},
	}
}

func newNamespace(name string) corev1.Namespace {
	return corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newNamespaces(names []string) []corev1.Namespace {
	var namespaces []corev1.Namespace
	for _, v := range names {
		namespace := newNamespace(v)
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

func createNPods(namePrefix string, namespace string, n int, status corev1.PodPhase) []runtime.Object {
	var pods []runtime.Object
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), namespace, status)
		pods = append(pods, &pod)
	}

	return pods
}

func createNRunningPods(namePrefix string, namespace string, n int) []runtime.Object {
	return createNPods(namePrefix, namespace, n, corev1.PodRunning)
}

func getPodList(client client.Interface, namespace string) *corev1.PodList {
	podList, _ := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	return podList
}

func TestGetRandomPods(t *testing.T) {
	t.Parallel()

	pods := createNRunningPods("pod", "default", 2)
	namespace := newNamespace("default")
	client := fake.NewSimpleClientset(pods[0], pods[1], &namespace)
	excludedNamespaes := []string{"kube-system"}
	randomPods, _ := GetRandomPods(client, 2, excludedNamespaes)

	assert.Lenf(t, randomPods, 2, "Expected 2 items in podList, got %d", len(randomPods))
}

func TestGetEligibleNamespaces(t *testing.T) {
	t.Parallel()

	namespaces := newNamespaces([]string{"default", "kube-system"})
	excludedNamespaes := []string{"kube-system"}
	eligibleNamespaces, _ := getEligibleNamespaces(namespaces, excludedNamespaes)

	assert.Lenf(t, eligibleNamespaces, 1, "Expected 1 items in NS list, got %d", len(eligibleNamespaces))
}
