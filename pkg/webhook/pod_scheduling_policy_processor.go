// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/gardener/kupid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
	"github.com/go-logr/logr"
)

// podSchedulingPolicyProcessorFactory implements the processFactory interface to validate podschedulingpolicies.
type podSchedulingPolicyProcessorFactory struct{}

func (pf *podSchedulingPolicyProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   kupidv1alpha1.GroupVersion.Group,
		Version: kupidv1alpha1.GroupVersion.Version,
		Kind:    "PodSchedulingPolicy",
	}
}

func (pf *podSchedulingPolicyProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSchedulingPolicyConfigurationProcessor{
		podSchedulingPolicyConfigurationValidator: &podSchedulingPolicyConfigurationValidatorImpl{
			podSchedulingPolicyConfigurationValidatorCallbacks: &podSchedulingPolicyValidatorCallbacks{},
		},
	}
}

func (pf *podSchedulingPolicyProcessorFactory) isMutating() bool {
	return false
}

// podSchedulingPolicyValidatorCallbacks implements podSchedulingPolicyValidatorCallbacks interface
type podSchedulingPolicyValidatorCallbacks struct {
	psp kupidv1alpha1.PodSchedulingPolicy
}

func (p *podSchedulingPolicyValidatorCallbacks) isNamespaced() bool {
	return true
}

func (p *podSchedulingPolicyValidatorCallbacks) getObjectMeta() *metav1.ObjectMeta {
	return &(p.psp.ObjectMeta)
}

func (p *podSchedulingPolicyValidatorCallbacks) getPodSelector() *metav1.LabelSelector {
	return p.psp.Spec.PodSelector
}

func (p *podSchedulingPolicyValidatorCallbacks) getPodSchedulingPolicyConfiguration() common.PodSchedulingPolicyConfiguration {
	return &(p.psp)
}
