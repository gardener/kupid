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
