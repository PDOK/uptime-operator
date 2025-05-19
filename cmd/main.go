/*
Copyright 2024 pdok.nl.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto/tls"
	"flag"
	"os"

	"github.com/PDOK/uptime-operator/internal/service"
	"github.com/PDOK/uptime-operator/internal/service/providers"
	"github.com/PDOK/uptime-operator/internal/util"
	"github.com/peterbourgon/ff"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/PDOK/uptime-operator/internal/controller"
	traefikio "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(traefikio.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

//nolint:funlen
func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var namespaces util.SliceFlag
	var slackChannel string
	var slackWebhookURL string
	var enableDeletes bool
	var uptimeProvider string
	var pingdomAPIToken string
	var pingdomAlertUserIDs util.SliceFlag
	var pingdomAlertIntegrationIDs util.SliceFlag
	var betterstackAPIToken string

	// Default kubebuilder
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers.")
	flag.BoolVar(&enableDeletes, "enable-deletes", false,
		"Allow the operator to delete checks from the uptime provider when ingress routes are removed.")

	// General uptime-operator
	flag.Var(&namespaces, "namespace", "Namespace(s) to watch for changes. "+
		"Specify this flag multiple times for each namespace to watch. When not provided all namespaces will be watched.")
	flag.StringVar(&slackChannel, "slack-channel", "",
		"The Slack Channel ID for posting updates when uptime checks are mutated.")
	flag.StringVar(&slackWebhookURL, "slack-webhook-url", "",
		"The webhook URL required to post messages to the given Slack channel.")
	flag.StringVar(&uptimeProvider, "uptime-provider", "mock",
		"Name of the (SaaS) uptime monitoring provider to use.")

	// Pingdom specific
	flag.StringVar(&pingdomAPIToken, "pingdom-api-token", "",
		"The API token to authenticate with Pingdom. Only applies when 'uptime-provider' is 'pingdom'")
	flag.Var(&pingdomAlertUserIDs, "pingdom-alert-user-ids",
		"One or more IDs of Pingdom users to alert. Only applies when 'uptime-provider' is 'pingdom'")
	flag.Var(&pingdomAlertIntegrationIDs, "pingdom-alert-integration-ids",
		"One or more IDs of Pingdom integrations (like slack channels) to alert. Only applies when 'uptime-provider' is 'pingdom'")

	// Better Stack specific
	flag.StringVar(&betterstackAPIToken, "betterstack-api-token", "",
		"The API token to authenticate with Better Stack. Only applies when 'uptime-provider' is 'betterstack'")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if err := ff.Parse(flag.CommandLine, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		setupLog.Error(err, "unable to parse flags")
		os.Exit(1)
	}

	mgr, err := createManager(enableHTTP2, metricsAddr, secureMetrics, probeAddr, enableLeaderElection, namespaces)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var uptimeProviderSettings any
	uptimeProviderID := service.UptimeProviderID(uptimeProvider)

	// Optional provider specific flag handling
	if uptimeProviderID == service.ProviderPingdom {
		alertUserIDs, err := util.StringsToInts(pingdomAlertUserIDs)
		if err != nil {
			setupLog.Error(err, "Unable to parse 'pingdom-alert-user-ids' flag")
			os.Exit(1)
		}
		alertIntegrationIDs, err := util.StringsToInts(pingdomAlertIntegrationIDs)
		if err != nil {
			setupLog.Error(err, "Unable to parse 'pingdom-alert-integration-ids' flag")
			os.Exit(1)
		}
		uptimeProviderSettings = providers.PingdomSettings{
			APIToken:       pingdomAPIToken,
			UserIDs:        alertUserIDs,
			IntegrationIDs: alertIntegrationIDs,
		}
	} else if uptimeProviderID == service.ProviderBetterStack {
		uptimeProviderSettings = providers.BetterStackSettings{
			APIToken: betterstackAPIToken,
		}
	}

	// Setup controller
	if err = (&controller.IngressRouteReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		UptimeCheckService: service.New(
			service.WithProviderAndSettings(uptimeProviderID, uptimeProviderSettings),
			service.WithSlack(slackWebhookURL, slackChannel),
			service.WithDeletes(enableDeletes),
		),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IngressRoute")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func createManager(enableHTTP2 bool, metricsAddr string, secureMetrics bool, probeAddr string,
	enableLeaderElection bool, namespaces util.SliceFlag) (manager.Manager, error) {
	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancelation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	var tlsOpts []func(*tls.Config)
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	managerOpts := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ce3b4f94.pdok.nl",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	}

	if len(namespaces) > 0 {
		namespacesToWatch := make(map[string]cache.Config)
		for _, namespace := range namespaces {
			namespacesToWatch[namespace] = cache.Config{}
		}
		managerOpts.Cache.DefaultNamespaces = namespacesToWatch
	}

	return ctrl.NewManager(ctrl.GetConfigOrDie(), managerOpts)
}
