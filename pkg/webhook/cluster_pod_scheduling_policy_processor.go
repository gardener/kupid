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
	unversionedvalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
)

// clusterPodSchedulingPolicyProcessorFactory implements the processFactory interface to validate clusterpodschedulingpolicies.
type clusterPodSchedulingPolicyProcessorFactory struct{}

func (pf *clusterPodSchedulingPolicyProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   kupidv1alpha1.GroupVersion.Group,
		Version: kupidv1alpha1.GroupVersion.Version,
		Kind:    "ClusterPodSchedulingPolicy",
	}
}

func (pf *clusterPodSchedulingPolicyProcessorFactory) newProcessor() processor {
	var cpspValidator = &clusterPodSchedulingPolicyValidatorImpl{}
	cpspValidator.podSchedulingPolicyConfigurationValidator = &podSchedulingPolicyConfigurationValidatorImpl{
		podSchedulingPolicyConfigurationValidatorCallbacks: cpspValidator,
	}
	return &podSchedulingPolicyConfigurationProcessor{
		podSchedulingPolicyConfigurationValidator: cpspValidator,
	}
}

func (pf *clusterPodSchedulingPolicyProcessorFactory) isMutating() bool {
	return false
}

// clusterPodSchedulingPolicyValidatorImpl implements the podSchedulingPolicyConfigurationValidator interface by extending podSchedulingPolicyConfigurationValidatorImpl.
// To make use of the podSchedulingPolicyConfigurationValidatorImpl it also implements the podSchedulingPolicyValidatorCallbacks interface.
type clusterPodSchedulingPolicyValidatorImpl struct {
	cpsp kupidv1alpha1.ClusterPodSchedulingPolicy
	podSchedulingPolicyConfigurationValidator
}

func (p *clusterPodSchedulingPolicyValidatorImpl) validate() field.ErrorList {
	allErrs := p.podSchedulingPolicyConfigurationValidator.validate()

	nsSelector := p.cpsp.Spec.NamespaceSelector
	if nsSelector != nil {
		allErrs = append(allErrs, unversionedvalidation.ValidateLabelSelector(nsSelector, field.NewPath("spec").Child("namespaceSelector"))...)
	}

	return allErrs
}

func (p *clusterPodSchedulingPolicyValidatorImpl) isNamespaced() bool {
	return false
}

func (p *clusterPodSchedulingPolicyValidatorImpl) getObjectMeta() *metav1.ObjectMeta {
	return &(p.cpsp.ObjectMeta)
}

func (p *clusterPodSchedulingPolicyValidatorImpl) getPodSelector() *metav1.LabelSelector {
	return p.cpsp.Spec.PodSelector
}

func (p *clusterPodSchedulingPolicyValidatorImpl) getPodSchedulingPolicyConfiguration() common.PodSchedulingPolicyConfiguration {
	return &(p.cpsp)
}
