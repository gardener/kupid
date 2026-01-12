// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
	"github.com/gardener/kupid/pkg/webhook"
	kupidwebhook "github.com/gardener/kupid/pkg/webhook"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
)

var (
	scheme                 = runtime.NewScheme()
	setupLog               = ctrl.Log.WithName("setup")
	mutateOnCreateOrUpdate = []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
		admissionregistrationv1.Update,
	}
	mutateOnCreate = []admissionregistrationv1.OperationType{
		admissionregistrationv1.Create,
	}
)

const (
	webhookName     = "kupid"
	webhookFullName = "gardener-extension-" + webhookName

	flagWebhookPort           = "webhook-port"
	flagMetricsAddr           = "metrics-addr"
	flagHealthzAddr           = "healthz-addr"
	flagCertDir               = "cert-dir"
	flagRegisterWebhooks      = "register-webhooks"
	flagWebhookTimeoutSeconds = "webhook-timeout-seconds"
	flagWebhookFailurePolicy  = "webhook-failure-policy"
	flagSyncPeriod            = "sync-period"
	flagQPS                   = "qps"
	flagBurst                 = "burst"
	flagEnableLeaderElection  = "enable-leader-election"
	envNamespace              = "WEBHOOK_CONFIG_NAMESPACE"

	defaultWebhookPort           = 9443
	defaultWebhookTimeoutSeconds = 15
	defaultWebhookFailurePolicy  = string(admissionregistrationv1.Ignore)
	defaultMetricsAddr           = ":8081"
	defaultHealthzAddr           = ":8080"
	defaultSyncPeriod            = 1 * time.Hour
	defaultEnableLeaderElection  = true
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = kupidv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = batchv1beta1.AddToScheme(scheme)
}

func main() {
	var (
		webhookPort           int
		metricsAddr           string
		healthzAddr           string
		certDir               string
		registerWebhooks      bool
		webhookTimeoutSeconds int
		webhookFailurePolicy  string
		syncPeriod            time.Duration
		qps                   float64
		burst                 int
		enableLeaderElection  bool
		namespace             string
		logLevel              = uberzap.LevelFlag("v", zapcore.InfoLevel, "Logging level")
	)

	flag.IntVar(&webhookPort, flagWebhookPort, defaultWebhookPort, "The port for the webhook server to listen on.")
	flag.StringVar(&metricsAddr, flagMetricsAddr, defaultMetricsAddr, "The address the metric endpoint binds to.")
	flag.StringVar(&healthzAddr, flagHealthzAddr, defaultHealthzAddr, "The address the healthz endpoint binds to.")
	flag.StringVar(&certDir, flagCertDir, "./certs", "The directory where the serving certs are kept.")
	flag.BoolVar(&registerWebhooks, flagRegisterWebhooks, false, "If enabled registers the webhook configurations automatically. The webhook is assumed to be reachable by a service with the name 'gardener-extension-kupid' within the same namespace. If necessary this will also generate TLS certificates for the webhook as well as the webhook configurations. The generated certificates will be published by creating a secret with the name 'gardener-extension-webhook-cert'. If the secret already exists then it is reused and no new certificates are generated.")
	flag.IntVar(&webhookTimeoutSeconds, flagWebhookTimeoutSeconds, defaultWebhookTimeoutSeconds, "If webhooks are registered automatically then they are configured with this timeout.")
	flag.StringVar(&webhookFailurePolicy, flagWebhookFailurePolicy, defaultWebhookFailurePolicy, "If webhooks are enabled, this flag sets the failure policy set for the webhook configurations deployed by Kupid. Allowed values are `Ignore` and `Fail`.")
	flag.DurationVar(&syncPeriod, flagSyncPeriod, defaultSyncPeriod, "SyncPeriod determines the minimum frequency at which watched resources are reconciled. A lower period will correct entropy more quickly, but reduce responsiveness to change if there are many watched resources. Change this value only if you know what you are doing.")
	flag.Float64Var(&qps, flagQPS, float64(rest.DefaultQPS), "Throttling QPS configuration for the client to host apiserver.")
	flag.IntVar(&burst, flagBurst, rest.DefaultBurst, "Throttling burst configuration for the client to host apiserver.")
	flag.BoolVar(&enableLeaderElection, flagEnableLeaderElection, defaultEnableLeaderElection, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager at any given time.")

	flag.Parse()

	// validations
	if webhookFailurePolicy != string(admissionregistrationv1.Ignore) && webhookFailurePolicy != string(admissionregistrationv1.Fail) {
		setupLog.Error(fmt.Errorf("provided failure policy %s is invalid; allowed values are `Ignore` and `Fail`", webhookFailurePolicy), "flag validation failed")
		os.Exit(1)
	}

	level := uberzap.NewAtomicLevelAt(*logLevel)
	ctrl.SetLogger(zap.New(buildLoggerOpts(level)...))

	namespace = os.Getenv(envNamespace)

	setupLog.Info("Running with",
		flagWebhookPort, webhookPort,
		flagMetricsAddr, metricsAddr,
		flagCertDir, certDir,
		flagRegisterWebhooks, registerWebhooks,
		flagWebhookTimeoutSeconds, webhookTimeoutSeconds,
		flagWebhookFailurePolicy, webhookFailurePolicy,
		flagSyncPeriod, syncPeriod,
		flagQPS, qps,
		flagBurst, burst,
		envNamespace, namespace)

	config := ctrl.GetConfigOrDie()

	config.QPS = float32(qps)
	config.Burst = burst

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         metricsAddr,
		Port:                       webhookPort,
		HealthProbeBindAddress:     healthzAddr,
		SyncPeriod:                 &syncPeriod,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "kupid-leader-election",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize webhook certificates
	func() {
		ws := mgr.GetWebhookServer()
		ws.CertDir = certDir
	}()

	if registerWebhooks {
		// #nosec G115 (CWE-190) -- webhookTimeout is controlled via spec and will not overflow and default value is set to 15 seconds
		if err := doRegisterWebhooks(mgr, certDir, namespace, int32(webhookTimeoutSeconds), admissionregistrationv1.FailurePolicyType(webhookFailurePolicy)); err != nil {
			setupLog.Error(err, "Error registering webhooks. Aborting startup...")
			os.Exit(1)
		}
	}

	if w, err := webhook.NewDefaultWebhook(); err != nil {
		setupLog.Error(err, "unable to create default webhook", "webhook", webhookName)
	} else if err := w.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup", "webhook", webhookName)
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func doRegisterWebhooks(mgr manager.Manager, certDir, namespace string, timeoutSeconds int32, webhookFailurePolicy admissionregistrationv1.FailurePolicyType) error {
	k8sClient, err := getClient(mgr)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	setupLog.Info("Registering TLS certificates if necessary.")

	caBundle, err := kupidwebhook.GenerateCertificates(
		ctx,
		mgr,
		certDir,
		namespace,
		webhookName,
		extensionswebhook.ModeService,
		"",
	)
	if err != nil {
		return err
	}
	clientConfig := buildWebhookClientConfig(namespace, caBundle)

	setupLog.Info("Registering webhooks if necessary.")

	namespaceObj := &corev1.Namespace{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: namespace}, namespaceObj); err != nil {
		return fmt.Errorf("failed to get namespace %s: %w", namespace, err)
	}
	namespaceUID := namespaceObj.UID

	for _, f := range []webhookConfigGeneratorFn{
		newValidatingWebhookConfig,
		newMutatingWebhookConfig,
	} {
		obj, mutateFn := f(clientConfig, timeoutSeconds, webhookFailurePolicy, namespace, namespaceUID)
		if _, err := controllerutil.CreateOrUpdate(ctx, k8sClient, obj, mutateFn); err != nil {
			return err
		}
	}

	return nil
}

func buildWebhookClientConfig(namespace string, caBundle []byte) admissionregistrationv1.WebhookClientConfig {
	var path = webhook.WebhookPath
	return admissionregistrationv1.WebhookClientConfig{
		CABundle: caBundle,
		Service: &admissionregistrationv1.ServiceReference{
			Namespace: namespace,
			Name:      webhookFullName,
			Path:      &path,
		},
	}
}

func buildRuleWithOperations(gv schema.GroupVersion, resources []string, operations []admissionregistrationv1.OperationType) admissionregistrationv1.RuleWithOperations {
	return admissionregistrationv1.RuleWithOperations{
		Operations: operations,
		Rule: admissionregistrationv1.Rule{
			APIGroups:   []string{gv.Group},
			APIVersions: []string{gv.Version},
			Resources:   resources,
		},
	}
}

type webhookConfigGeneratorFn func(clientConfig admissionregistrationv1.WebhookClientConfig, timeoutSeconds int32, webhookFailurePolicy admissionregistrationv1.FailurePolicyType, extensionNamespace string, namespaceUID types.UID) (client.Object, controllerutil.MutateFn)

func newValidatingWebhookConfig(clientConfig admissionregistrationv1.WebhookClientConfig, timeoutSeconds int32, webhookFailurePolicy admissionregistrationv1.FailurePolicyType, extensionNamespace string, namespaceUID types.UID) (client.Object, controllerutil.MutateFn) {
	var (
		exact = admissionregistrationv1.Exact
		none  = admissionregistrationv1.SideEffectClassNone
	)

	obj := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookFullName,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "v1", Kind: "Namespace", Name: extensionNamespace, UID: namespaceUID},
			},
		},
	}

	return obj, func() error {
		obj.Webhooks = []admissionregistrationv1.ValidatingWebhook{
			{
				Name:         "validate." + kupidv1alpha1.GroupVersion.Group,
				ClientConfig: clientConfig,
				Rules: []admissionregistrationv1.RuleWithOperations{
					buildRuleWithOperations(
						kupidv1alpha1.GroupVersion,
						[]string{
							"clusterpodschedulingpolicies",
							"podschedulingpolicies",
						},
						mutateOnCreateOrUpdate,
					),
				},
				FailurePolicy:  &webhookFailurePolicy,
				MatchPolicy:    &exact,
				SideEffects:    &none,
				TimeoutSeconds: &timeoutSeconds,
				AdmissionReviewVersions: []string{
					"v1",
				},
			},
		}
		return nil
	}
}

func newMutatingWebhookConfig(clientConfig admissionregistrationv1.WebhookClientConfig, timeoutSeconds int32, webhookFailurePolicy admissionregistrationv1.FailurePolicyType, extensionNamespace string, namespaceUID types.UID) (client.Object, controllerutil.MutateFn) {
	var (
		equivalent = admissionregistrationv1.Equivalent
		none       = admissionregistrationv1.SideEffectClassNone
		ifNeeded   = admissionregistrationv1.IfNeededReinvocationPolicy
	)

	obj := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookFullName,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "v1", Kind: "Namespace", Name: extensionNamespace, UID: namespaceUID},
			},
		},
	}

	return obj, func() error {
		obj.Webhooks = []admissionregistrationv1.MutatingWebhook{
			{
				Name: "mutate." + kupidv1alpha1.GroupVersion.Group,
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "role",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"kube-system"},
						},
					},
				},
				ClientConfig: clientConfig,
				Rules: []admissionregistrationv1.RuleWithOperations{
					buildRuleWithOperations(
						appsv1.SchemeGroupVersion,
						[]string{
							"daemonsets",
							"deployments",
							"statefulsets",
						},
						mutateOnCreateOrUpdate,
					),
					buildRuleWithOperations(
						batchv1.SchemeGroupVersion,
						[]string{
							"jobs",
						},
						mutateOnCreate,
					),
					buildRuleWithOperations(
						batchv1beta1.SchemeGroupVersion,
						[]string{
							"cronjobs",
						},
						mutateOnCreateOrUpdate,
					),
				},
				FailurePolicy:      &webhookFailurePolicy,
				MatchPolicy:        &equivalent,
				SideEffects:        &none,
				TimeoutSeconds:     &timeoutSeconds,
				ReinvocationPolicy: &ifNeeded,
				AdmissionReviewVersions: []string{
					"v1",
				},
			},
		}
		return nil
	}
}

func getClient(mgr manager.Manager) (client.Client, error) {
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		return nil, err
	}

	return client.New(mgr.GetConfig(), client.Options{Scheme: s})
}

func buildLoggerOpts(level uberzap.AtomicLevel) []zap.Opts {
	var opts []zap.Opts
	opts = append(opts, zap.UseDevMode(false))
	opts = append(opts, zap.JSONEncoder(func(encoderConfig *zapcore.EncoderConfig) {
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	}))
	opts = append(opts, zap.Level(&level))
	return opts
}
