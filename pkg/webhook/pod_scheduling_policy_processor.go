// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"github.com/gardener/kupid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
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

func (pf *podSchedulingPolicyProcessorFactory) newProcessor() processor {
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
