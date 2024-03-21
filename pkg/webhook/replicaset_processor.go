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

// replicaSetProcessorFactory implements the processFactory interface to mutate pod template spec in replicasets.
type replicaSetProcessorFactory struct{}

func (pf *replicaSetProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "ReplicaSet",
	}
}

func (pf *replicaSetProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &replicaSetCallbacks{},
		logger:           logger,
	}
}

func (pf *replicaSetProcessorFactory) isMutating() bool {
	return true
}

// replicaSetCallbacks implements podSpecCallbacks
type replicaSetCallbacks struct {
	replicaSet appsv1.ReplicaSet
}

func (r *replicaSetCallbacks) getObject() runtime.Object {
	return &(r.replicaSet)
}

func (r *replicaSetCallbacks) getPodSpec() *corev1.PodSpec {
	return &(r.replicaSet.Spec.Template.Spec)
}

func (r *replicaSetCallbacks) getPodLabels() map[string]string {
	return r.replicaSet.Spec.Template.Labels
}
