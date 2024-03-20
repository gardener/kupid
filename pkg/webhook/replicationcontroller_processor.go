// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// replicationControllerProcessorFactory implements the processFactory interface to mutate pod template spec in replicationcontrollers.
type replicationControllerProcessorFactory struct{}

func (pf *replicationControllerProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "ReplicationController",
	}
}

func (pf *replicationControllerProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &replicationControllerCallbacks{},
		logger:           logger,
	}
}

func (pf *replicationControllerProcessorFactory) isMutating() bool {
	return true
}

// replicationControllerCallbacks implements podSpecCallbacks
type replicationControllerCallbacks struct {
	replicationController corev1.ReplicationController
}

func (r *replicationControllerCallbacks) getObject() runtime.Object {
	return &(r.replicationController)
}

func (r *replicationControllerCallbacks) getPodSpec() *corev1.PodSpec {
	return &(r.replicationController.Spec.Template.Spec)
}

func (r *replicationControllerCallbacks) getPodLabels() map[string]string {
	return r.replicationController.Spec.Template.Labels
}
