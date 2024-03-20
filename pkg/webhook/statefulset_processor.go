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

// statefulSetProcessorFactory implements the processFactory interface to mutate pod template spec in statefulsets.
type statefulSetProcessorFactory struct{}

func (pf *statefulSetProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "StatefulSet",
	}
}

func (pf *statefulSetProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &statefulSetCallbacks{},
		logger:           logger,
	}
}

func (pf *statefulSetProcessorFactory) isMutating() bool {
	return true
}

// statefulSetCallbacks implements podSpecCallbacks
type statefulSetCallbacks struct {
	statefulSet appsv1.StatefulSet
}

func (s *statefulSetCallbacks) getObject() runtime.Object {
	return &(s.statefulSet)
}

func (s *statefulSetCallbacks) getPodSpec() *corev1.PodSpec {
	return &(s.statefulSet.Spec.Template.Spec)
}

func (s *statefulSetCallbacks) getPodLabels() map[string]string {
	return s.statefulSet.Spec.Template.Labels
}
