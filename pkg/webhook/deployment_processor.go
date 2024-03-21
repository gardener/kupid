// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// deploymentProcessorFactory implements the processFactory interface to mutate pod template spec in deployments.
type deploymentProcessorFactory struct{}

func (pf *deploymentProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "Deployment",
	}
}

func (pf *deploymentProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &deploymentCallbacks{},
		logger:           logger,
	}
}

func (pf *deploymentProcessorFactory) isMutating() bool {
	return true
}

// deploymentCallbacks implements podSpecCallbacks
type deploymentCallbacks struct {
	deployment appsv1.Deployment
}

func (d *deploymentCallbacks) getObject() runtime.Object {
	return &(d.deployment)
}

func (d *deploymentCallbacks) getPodSpec() *corev1.PodSpec {
	return &(d.deployment.Spec.Template.Spec)
}

func (d *deploymentCallbacks) getPodLabels() map[string]string {
	return d.deployment.Spec.Template.Labels
}
