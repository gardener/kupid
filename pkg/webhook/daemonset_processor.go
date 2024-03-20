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

// daemonSetProcessorFactory implements the processFactory interface to mutate pod template spec in daemonsets.
type daemonSetProcessorFactory struct{}

func (pf *daemonSetProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "DaemonSet",
	}
}

func (pf *daemonSetProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &daemonSetCallbacks{},
		logger:           logger,
	}
}

func (pf *daemonSetProcessorFactory) isMutating() bool {
	return true
}

// daemonSetCallbacks implements podSpecCallbacks
type daemonSetCallbacks struct {
	daemonSet appsv1.DaemonSet
}

func (d *daemonSetCallbacks) getObject() runtime.Object {
	return &(d.daemonSet)
}

func (d *daemonSetCallbacks) getPodSpec() *corev1.PodSpec {
	return &(d.daemonSet.Spec.Template.Spec)
}

func (d *daemonSetCallbacks) getPodLabels() map[string]string {
	return d.daemonSet.Spec.Template.Labels
}
