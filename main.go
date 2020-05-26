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

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	appsv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
	"github.com/gardener/kupid/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const webhookName = "kupid"

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = kupidv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = batchv1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var certDir string
	var certName string
	var keyName string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&certDir, "cert-dir", "./certs", "The directory where the serving certs are kept.")
	flag.StringVar(&certName, "cert-name", "tls.crt", "The file name of the TLS serving certificate in the cert-dir.")
	flag.StringVar(&keyName, "cert-key", "tls.key", "The file name of the TLS serving certificate key in the cert-dir.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize webhook certificates
	func() {
		ws := mgr.GetWebhookServer()
		ws.CertDir = certDir
		ws.CertName = certName
		ws.KeyName = keyName
	}()

	if w, err := webhook.NewDefaultWebhook(); err != nil {
		setupLog.Error(err, "unable to create default webhook", "webhook", webhookName)
	} else if err := w.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup", "webhook", webhookName)
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
