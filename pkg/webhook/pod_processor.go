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

// podProcessorFactory implements the processFactory interface to mutate pod spec in pods.
type podProcessorFactory struct{}

func (pf *podProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "Pod",
	}
}

func (pf *podProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &podCallbacks{},
		logger:           logger,
	}
}

func (pf *podProcessorFactory) isMutating() bool {
	return true
}

// podCallbacks implements podSpecCallbacks
type podCallbacks struct {
	pod corev1.Pod
}

func (p *podCallbacks) getObject() runtime.Object {
	return &(p.pod)
}

func (p *podCallbacks) getPodSpec() *corev1.PodSpec {
	return &(p.pod.Spec)
}

func (p *podCallbacks) getPodLabels() map[string]string {
	return p.pod.Labels
}
