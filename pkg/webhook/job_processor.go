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
