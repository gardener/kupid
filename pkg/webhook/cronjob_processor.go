// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/go-logr/logr"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// cronJobProcessorFactory implements the processFactory interface to mutate pod template spec in cronjobs.
type cronJobProcessorFactory struct{}

func (pf *cronJobProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   batchv1beta1.SchemeGroupVersion.Group,
		Version: batchv1beta1.SchemeGroupVersion.Version,
		Kind:    "CronJob",
	}
}

func (pf *cronJobProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &cronJobCallbacks{},
		logger:           logger,
	}
}

func (pf *cronJobProcessorFactory) isMutating() bool {
	return true
}

// cronJobCallbacks implements podSpecCallbacks
type cronJobCallbacks struct {
	cronJob batchv1beta1.CronJob
}

func (c *cronJobCallbacks) getObject() runtime.Object {
	return &(c.cronJob)
}

func (c *cronJobCallbacks) getPodSpec() *corev1.PodSpec {
	return &(c.cronJob.Spec.JobTemplate.Spec.Template.Spec)
}

func (c *cronJobCallbacks) getPodLabels() map[string]string {
	return c.cronJob.Spec.JobTemplate.Spec.Template.Labels
}
