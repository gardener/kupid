// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// jobProcessorFactory implements the processFactory interface to mutate pod template spec in jobs.
type jobProcessorFactory struct{}

func (pf *jobProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   batchv1.SchemeGroupVersion.Group,
		Version: batchv1.SchemeGroupVersion.Version,
		Kind:    "Job",
	}
}

func (pf *jobProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &jobCallbacks{},
		logger:           logger,
	}
}

func (pf *jobProcessorFactory) isMutating() bool {
	return true
}

// jobCallbacks implements podSpecCallbacks
type jobCallbacks struct {
	job batchv1.Job
}

func (j *jobCallbacks) getObject() runtime.Object {
	return &(j.job)
}

func (j *jobCallbacks) getPodSpec() *corev1.PodSpec {
	return &(j.job.Spec.Template.Spec)
}

func (j *jobCallbacks) getPodLabels() map[string]string {
	return j.job.Spec.Template.Labels
}
