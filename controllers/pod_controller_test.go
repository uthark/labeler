// +kubebuilder:docs-gen:collapse=Apache License

package controllers

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

const (
	PodName      = "test-pod"
	PodNamespace = "default"
)

var _ = Describe("PodController.Reconcile()", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	AfterEach(func() {
		Expect(k8sClient.Delete(context.Background(), buildPod(PodNamespace, PodName))).Should(Succeed())
	})

	Context("Namespace doesn't have a label", func() {
		It("Should not copy labels from namespace", func() {
			ctx := context.Background()
			pod := buildPod(PodNamespace, PodName)
			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())

			podLookupKey := types.NamespacedName{Name: PodName, Namespace: PodNamespace}
			createdPod := &v1.Pod{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.

			Expect(createdPod.Labels["team"]).Should(BeEmpty())
		})
	})

	Context("Namespace has a label", func() {
		It("Should copy labels from namespace", func() {

			ctx := context.Background()
			defaultNs := &v1.Namespace{}
			k8sClient.Get(ctx, client.ObjectKey{Name: PodNamespace}, defaultNs)
			metav1.SetMetaDataAnnotation(&defaultNs.ObjectMeta, "labeler.uthark.dev/enabled", "true")
			metav1.SetMetaDataLabel(&defaultNs.ObjectMeta, "team", "foo")

			Expect(k8sClient.Update(ctx, defaultNs)).Should(Succeed())

			//ctx := context.Background()
			pod := buildPod(PodNamespace, PodName)
			Expect(k8sClient.Create(ctx, pod)).Should(Succeed())

			podLookupKey := types.NamespacedName{Name: PodName, Namespace: PodNamespace}
			createdPod := &v1.Pod{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Eventually(func() string {
				err := k8sClient.Get(ctx, podLookupKey, createdPod)
				if err != nil {
					return ""
				}
				return createdPod.Labels["team"]
			}, timeout, interval).Should(Equal("foo"))

		})
	})

})

func buildPod(namespace, name string) *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "default",
					Image: "k8s.gcr.io/pause:3.5.0",
				},
			},
		},
	}
}
