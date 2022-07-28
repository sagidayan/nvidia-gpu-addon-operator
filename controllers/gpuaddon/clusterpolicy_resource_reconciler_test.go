/*
Copyright 2022.

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

package gpuaddon

import (
	"context"

	gpuv1 "github.com/NVIDIA/gpu-operator/api/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/scheme"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	addonv1alpha1 "github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/api/v1alpha1"
	"github.com/rh-ecosystem-edge/nvidia-gpu-addon-operator/internal/common"
)

var _ = Describe("ClusterPolicy Resource Reconcile", Ordered, func() {
	Context("Reconcile", func() {
		common.ProcessConfig()
		rrec := &ClusterPolicyResourceReconciler{}
		gpuAddon := addonv1alpha1.GPUAddon{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		scheme := scheme.Scheme
		Expect(gpuv1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		var cp gpuv1.ClusterPolicy

		It("should create the ClusterPolicy", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects().
				Build()

			cond, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cond).To(HaveLen(1))
			Expect(cond[0].Type).To(Equal(ClusterPolicyDeployedCondition))
			Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))

			err = c.Get(context.TODO(), types.NamespacedName{
				Name: common.GlobalConfig.ClusterPolicyName,
			}, &cp)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cp.Spec.Driver.Version).To(BeEmpty())
		})
	})

	Context("Delete", func() {
		common.ProcessConfig()
		rrec := &ClusterPolicyResourceReconciler{}

		cp := &gpuv1.ClusterPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: common.GlobalConfig.ClusterPolicyName,
			},
		}

		scheme := scheme.Scheme
		Expect(gpuv1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		It("should delete the ClusterPolicy", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(cp).
				Build()

			deleted, err := rrec.Delete(context.TODO(), c)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(deleted).To(BeFalse())

			err = c.Get(context.TODO(), client.ObjectKey{
				Name: cp.Name,
			}, cp)
			Expect(err).Should(HaveOccurred())
			Expect(k8serrors.IsNotFound(err)).To(BeTrue())

			deleted, err = rrec.Delete(context.TODO(), c)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(deleted).To(BeTrue())
		})
	})

	Context("Driver version pinning", func() {
		common.ProcessConfig()
		pinnedVersion := "470.82.01"
		pinnedVersionUnsupported := "NotAVersion"
		rrec := &ClusterPolicyResourceReconciler{}
		gpuAddon := addonv1alpha1.GPUAddon{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: addonv1alpha1.GPUAddonSpec{
				DriverVersion: pinnedVersion,
			},
		}
		scheme := scheme.Scheme
		Expect(gpuv1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		var cp gpuv1.ClusterPolicy

		It("should create the ClusterPolicy with specified driver version", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects().
				Build()

			cond, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cond).To(HaveLen(1))
			Expect(cond[0].Type).To(Equal(ClusterPolicyDeployedCondition))
			Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))

			err = c.Get(context.TODO(), types.NamespacedName{
				Name: common.GlobalConfig.ClusterPolicyName,
			}, &cp)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(cp.Spec.Driver.Version).ToNot(BeEmpty())
			Expect(cp.Spec.Driver.Version).To(Equal(pinnedVersion))
		})

		It("should throw error for unsupported driver version", func() {
			c := fake.
				NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects().
				Build()

			gpuAddon.Spec.DriverVersion = pinnedVersionUnsupported
			err := c.Create(context.TODO(), &gpuAddon)
			Expect(err).Should(HaveOccurred())

			err = c.Get(context.TODO(), types.NamespacedName{
				Name: common.GlobalConfig.ClusterPolicyName,
			}, &cp)
			Expect(k8serrors.IsNotFound(err)).Should(BeTrue())
		})
	})
})
