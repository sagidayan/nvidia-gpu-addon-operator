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
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
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
				Name:      "test",
				Namespace: "test",
			},
		}
		gpuOperatorCsv := &operatorsv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gpu-operator-certified.v1.10.1",
				Namespace: gpuAddon.Namespace,
			},
			Spec: operatorsv1alpha1.ClusterServiceVersionSpec{
				RelatedImages: []operatorsv1alpha1.RelatedImage{
					{Name: "driver-image", Image: "nvcr.io/nvidia/driver@sha256somedigest"},
				},
			},
		}
		scheme := scheme.Scheme
		Expect(gpuv1.AddToScheme(scheme)).ShouldNot(HaveOccurred())
		Expect(operatorsv1alpha1.AddToScheme(scheme)).ShouldNot(HaveOccurred())

		var cp gpuv1.ClusterPolicy

		Context("with a GPUAddon.DriverVersion defined", func() {
			It("should create the ClusterPolicy with the GPUAddon.DriverVersion as the driver version", func() {
				gpuAddon.Spec = addonv1alpha1.GPUAddonSpec{
					DriverVersion: "515.48.07",
				}

				c := fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(gpuOperatorCsv).
					Build()

				cond, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cond).To(HaveLen(1))
				Expect(cond[0].Type).To(Equal(ClusterPolicyDeployedCondition))
				Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))

				err = c.Get(context.TODO(), types.NamespacedName{
					Name: common.GlobalConfig.ClusterPolicyName,
				}, &cp)
				Expect(err).ToNot(HaveOccurred())

				Expect(cp.Spec.Driver.Repository).To(Equal("nvcr.io/nvidia"))
				Expect(cp.Spec.Driver.Image).To(Equal("driver"))
				Expect(cp.Spec.Driver.Version).To(Equal("515.48.07"))
			})
		})

		Context("without a GPUAddon.DriverVersion defined", func() {
			It("should create the ClusterPolicy with the GPU operator CSV related driver-image", func() {
				gpuAddon.Spec = addonv1alpha1.GPUAddonSpec{
					DriverVersion: "",
				}

				c := fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(gpuOperatorCsv).
					Build()

				cond, err := rrec.Reconcile(context.TODO(), c, &gpuAddon)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(cond).To(HaveLen(1))
				Expect(cond[0].Type).To(Equal(ClusterPolicyDeployedCondition))
				Expect(cond[0].Status).To(Equal(metav1.ConditionTrue))

				err = c.Get(context.TODO(), types.NamespacedName{
					Name: common.GlobalConfig.ClusterPolicyName,
				}, &cp)
				Expect(err).ToNot(HaveOccurred())

				Expect(cp.Spec.Driver.Image).To(Equal("nvcr.io/nvidia/driver@sha256somedigest"))
			})
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
})
