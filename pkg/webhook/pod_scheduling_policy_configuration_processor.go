// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"context"

	"github.com/gardener/kupid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unversionedvalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// podSchedulingPolicyConfigurationProcessor implements processor interface to validate PodSchedulingPolicyConfiguration.
type podSchedulingPolicyConfigurationProcessor struct {
	podSchedulingPolicyConfigurationValidator
}

// podSchedulingPolicyConfigurationValidator defines the callback to valid
type podSchedulingPolicyConfigurationValidator interface {
	objectGetter
	validate() field.ErrorList
}

func (p *podSchedulingPolicyConfigurationProcessor) process(ctx context.Context, cacheReader, directReader client.Reader, namespace string) (mutated bool, err error) {
	allErrs := p.validate()
	if len(allErrs) > 0 {
		return false, allErrs.ToAggregate()
	}
	return false, nil
}

// podSchedulingPolicyConfigurationValidatorImpl implements the validation of PodSchedulingPolicyConfiguration using the supplied podSchedulingPolicyConfigurationCallbacks.
type podSchedulingPolicyConfigurationValidatorImpl struct {
	podSchedulingPolicyConfigurationValidatorCallbacks
}

// podSchedulingPolicyConfigurationValidatorCallbacks defines the callbacks necessary to implement the validation PodSchedulingPolicyConfiguration.
type podSchedulingPolicyConfigurationValidatorCallbacks interface {
	isNamespaced() bool
	getObjectMeta() *metav1.ObjectMeta
	getPodSelector() *metav1.LabelSelector
	getPodSchedulingPolicyConfiguration() common.PodSchedulingPolicyConfiguration
}

func (p *podSchedulingPolicyConfigurationValidatorImpl) getObject() runtime.Object {
	return p.getPodSchedulingPolicyConfiguration()
}

func (p *podSchedulingPolicyConfigurationValidatorImpl) validate() field.ErrorList {
	allErrs := common.ValidateObjectMeta(p.getObjectMeta(), p.isNamespaced(), common.ValidatePodName, field.NewPath("metadata"))

	specPath := field.NewPath("spec")

	podSelector := p.getPodSelector()
	if podSelector != nil {
		allErrs = append(allErrs, unversionedvalidation.ValidateLabelSelector(podSelector, specPath.Child("podSelector"))...)
	}

	allErrs = append(allErrs, p.validatePodSchedulingPolicyConfiguration(specPath)...)

	return allErrs
}

func (p *podSchedulingPolicyConfigurationValidatorImpl) validatePodSchedulingPolicyConfiguration(fldPath *field.Path) field.ErrorList {
	var (
		pspc    = p.getPodSchedulingPolicyConfiguration()
		allErrs = field.ErrorList{}
	)

	if pspc.GetNodeSelector() != nil {
		allErrs = append(allErrs, v1validation.ValidateLabels(pspc.GetNodeSelector(), fldPath.Child("nodeSelector"))...)
	}

	if pspc.GetNodeName() != "" {
		nodeName := pspc.GetNodeName()
		for _, msg := range common.ValidateNodeName(nodeName, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("nodeName"), nodeName, msg))
		}
	}

	if pspc.GetAffinity() != nil {
		allErrs = append(allErrs, common.ValidateAffinity(pspc.GetAffinity(), fldPath.Child("affinity"))...)
	}

	if pspc.GetTolerations() != nil {
		allErrs = append(allErrs, common.ValidateTolerations(pspc.GetTolerations(), fldPath.Child("tolerations"))...)
	}

	return allErrs
}
