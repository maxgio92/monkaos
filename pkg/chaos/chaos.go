package chaos

import (
	"context"
	"sync"
	"time"

	"monkaos/pkg/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k "k8s.io/client-go/kubernetes"
)

type Chaos struct {
	KillAt time.Time
	Victim v1.Pod
}

// New creates a new Chaos instance.
func New(killtime time.Time, pod v1.Pod) *Chaos {
	// TargetPodName will be populated at time of termination
	return &Chaos{
		KillAt: killtime,
		Victim: pod,
	}
}

// Schedule the execution of Chaos
func (c *Chaos) Run(ctx context.Context, results chan<- *Result, waitGroup *sync.WaitGroup, gracePeriodSeconds int64) {
	time.Sleep(c.durationToKill())
	c.execute(ctx, results, waitGroup, gracePeriodSeconds)
}

// timeToKill() calculates the duration from now until Chaos.killAt.
func (c *Chaos) durationToKill() time.Duration {
	return time.Until(c.KillAt)
}

// execute() executes the scheduled chaos.
func (c *Chaos) execute(ctx context.Context, resultCh chan<- *Result, waitGroup *sync.WaitGroup, gracePeriodSeconds int64) {
	defer waitGroup.Done()

	select {
	default:
		// critical zone: beginning.

		// Create kubernetes clientset
		clientset, err := kubernetes.NewClientset()
		if err != nil {
			resultCh <- NewResult(c, err)
			return
		}

		// Terminate and send error msg on failure
		err = c.terminate(ctx, clientset, gracePeriodSeconds)
		if err != nil {
			resultCh <- NewResult(c, err)
			return
		}

		// Send a success message.
		resultCh <- NewResult(c, nil)

		// critical zone: end.
	case <-ctx.Done():
		return
	}
}

// The termination type and value is processed here.
func (c *Chaos) terminate(ctx context.Context, clientset k.Interface, gracePeriodSeconds int64) error {
	return clientset.CoreV1().Pods(c.Victim.Namespace).Delete(ctx, c.Victim.Name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
}
