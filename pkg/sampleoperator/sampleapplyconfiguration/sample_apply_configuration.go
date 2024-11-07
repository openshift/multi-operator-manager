package sampleapplyconfiguration

import (
	"context"
	"fmt"
	"os"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	applyoperatorv1 "github.com/openshift/client-go/operator/applyconfigurations/operator/v1"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorinformer "github.com/openshift/client-go/operator/informers/externalversions"
	"github.com/openshift/library-go/pkg/manifestclient"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/genericoperatorclient"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"github.com/openshift/library-go/pkg/operator/status"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"github.com/openshift/multi-operator-manager/pkg/library/libraryapplyconfiguration"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func SampleRunApplyConfiguration(ctx context.Context, input libraryapplyconfiguration.ApplyConfigurationInput) (libraryapplyconfiguration.AllDesiredMutationsGetter, error) {
	authenticationOperatorInput, err := CreateOperatorInputFromMOM(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to configure operator input: %w", err)
	}
	operatorStarter, err := CreateOperatorStarter(ctx, authenticationOperatorInput)
	if err != nil {
		return nil, fmt.Errorf("unable to configure operators: %w", err)
	}
	var operatorRunError error
	if err := operatorStarter.RunOnce(ctx); err != nil {
		operatorRunError = fmt.Errorf("unable to run operators: %w", err)
	}

	return libraryapplyconfiguration.NewApplyConfigurationFromClient(input.MutationTrackingClient.GetMutations()), operatorRunError
}

type exampleOperatorInput struct {
	kubeClient            kubernetes.Interface
	configClient          configclient.Interface
	operatorClient        operatorclient.Interface
	exampleOperatorClient v1helpers.OperatorClient
	eventRecorder         events.Recorder

	informers []libraryapplyconfiguration.SimplifiedInformerFactory

	// controllers holds an optional list of controller names to run.
	// By default, all controllers are run.
	controllers []string
}

const componentName = "cluster-example-operator"

// CreateOperatorInputFromMOM this method is usually paired with CreateControllerInputFromControllerContext during the transition.
// This allows us to abstract the creation of clients from those things that depend on those clients so that initialization
// can happen as normal.
func CreateOperatorInputFromMOM(ctx context.Context, momInput libraryapplyconfiguration.ApplyConfigurationInput) (*exampleOperatorInput, error) {
	kubeClient, err := kubernetes.NewForConfigAndClient(manifestclient.RecommendedRESTConfig(), momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	configClient, err := configclient.NewForConfigAndClient(manifestclient.RecommendedRESTConfig(), momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	operatorClient, err := operatorclient.NewForConfigAndClient(manifestclient.RecommendedRESTConfig(), momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	authenticationOperatorClient, dynamicInformers, err := genericoperatorclient.NewOperatorClientWithClient(
		momInput.Clock,
		momInput.MutationTrackingClient.GetHTTPClient(),
		operatorv1.GroupVersion.WithResource("authentications"),
		operatorv1.GroupVersion.WithKind("Authentication"),
		extractOperatorSpec,
		extractOperatorStatus,
	)
	if err != nil {
		return nil, err
	}

	// TODO figure out to do correlation of events after we fluff them up
	eventRecorder := events.NewRecorder(
		kubeClient.CoreV1().Events("openshift-example-operator"),
		componentName,
		&corev1.ObjectReference{
			Kind:      "Deployment",
			Namespace: "openshift-example-operator",
			Name:      "example-operator",
		},
	)

	return &exampleOperatorInput{
		kubeClient:            kubeClient,
		configClient:          configClient,
		operatorClient:        operatorClient,
		exampleOperatorClient: authenticationOperatorClient,
		eventRecorder:         eventRecorder,
		informers: []libraryapplyconfiguration.SimplifiedInformerFactory{
			libraryapplyconfiguration.DynamicInformerFactoryAdapter(dynamicInformers), // we don't share the dynamic informers, but we only want to start when requested
		},
		controllers: momInput.Controllers,
	}, nil
}

func CreateOperatorStarter(ctx context.Context, exampleOperatorInput *exampleOperatorInput) (libraryapplyconfiguration.OperatorStarter, error) {
	ret := &libraryapplyconfiguration.SimpleOperatorStarter{
		Informers:   append([]libraryapplyconfiguration.SimplifiedInformerFactory{}, exampleOperatorInput.informers...),
		Controllers: exampleOperatorInput.controllers,
	}

	// create informers. This one is common for control plane operators.
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(
		exampleOperatorInput.kubeClient,
		"default",
		"openshift-authentication",
		"openshift-config",
		"openshift-config-managed",
		"openshift-oauth-apiserver",
		"openshift-authentication-operator",
		"kube-system",
		"openshift-etcd",
	)
	ret.Informers = append(ret.Informers, libraryapplyconfiguration.GeneratedNamespacedInformerFactoryAdapter(kubeInformersForNamespaces))
	// resyncs in individual controller loops for this operator are driven by a duration based trigger independent of a resource resync.
	// this allows us to resync essentially never, but reach out to external systems on a polling basis around one minute.
	operatorConfigInformers := operatorinformer.NewSharedInformerFactory(exampleOperatorInput.operatorClient, 24*time.Hour)
	ret.Informers = append(ret.Informers, libraryapplyconfiguration.GeneratedInformerFactoryAdapter(operatorConfigInformers))

	versionRecorder := status.NewVersionGetter()
	clusterOperator, err := exampleOperatorInput.configClient.ConfigV1().ClusterOperators().Get(ctx, "example", metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	// perform version changes to the version getter prior to tying it up in the status controller
	// via change-notification channel so that it only updates operator version in status once
	// either of the workloads synces
	for _, version := range clusterOperator.Status.Versions {
		versionRecorder.SetVersion(version.Name, version.Version)
	}
	versionRecorder.SetVersion("operator", os.Getenv("OPERATOR_IMAGE_VERSION"))

	resourceSyncer := resourcesynccontroller.NewResourceSyncController(
		"sample-operator-example",
		exampleOperatorInput.exampleOperatorClient,
		kubeInformersForNamespaces,
		v1helpers.CachedSecretGetter(exampleOperatorInput.kubeClient.CoreV1(), kubeInformersForNamespaces),
		v1helpers.CachedConfigMapGetter(exampleOperatorInput.kubeClient.CoreV1(), kubeInformersForNamespaces),
		exampleOperatorInput.eventRecorder,
	)
	ret.ControllerRunFns = append(ret.ControllerRunFns, libraryapplyconfiguration.AdaptRunFn(resourceSyncer.Run))
	ret.ControllerNamedRunOnceFns = append(ret.ControllerNamedRunOnceFns,
		libraryapplyconfiguration.AdaptSyncFn(exampleOperatorInput.eventRecorder, resourceSyncer.Name(), resourceSyncer.Sync))

	// the good example was above, now just add a few resources for us to play with

	ret.ControllerNamedRunOnceFns = append(ret.ControllerNamedRunOnceFns, libraryapplyconfiguration.NewNamedRunOnce(
		"sample-operator-ingress-creator",
		func(ctx context.Context) error {
			exampleOperatorInput.configClient.ConfigV1().Ingresses().Create(ctx, &configv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}}, metav1.CreateOptions{})
			return nil
		},
	))

	// this ensures the configmapinformer is requested so that it will start.
	kubeInformersForNamespaces.ConfigMapLister().ConfigMaps("openshift-authentication")
	ret.ControllerNamedRunOnceFns = append(ret.ControllerNamedRunOnceFns, libraryapplyconfiguration.NewNamedRunOnce(
		"sample-operator-failure-generator",
		func(ctx context.Context) error {
			exampleOperatorInput.eventRecorder.Event("must", "event")
			_, err := kubeInformersForNamespaces.ConfigMapLister().ConfigMaps("openshift-authentication").Get("fail-check")
			if apierrors.IsNotFound(err) {
				fmt.Printf("forced-failure not required\n")
				return nil
			}
			if err != nil {
				fmt.Printf("failed to get configmap: %v\n", err)
				return err
			}
			fmt.Printf("forcing an error\n")
			return fmt.Errorf("fail the process")
		},
	))

	demoController := NewDemoController(
		"sample-operator",
		exampleOperatorInput.kubeClient,
		kubeInformersForNamespaces.InformersFor("openshift-authentication").Core().V1().ConfigMaps(),
		exampleOperatorInput.eventRecorder,
	)
	ret.ControllerNamedRunOnceFns = append(ret.ControllerNamedRunOnceFns,
		libraryapplyconfiguration.AdaptNamedController(exampleOperatorInput.eventRecorder, demoController))

	return ret, nil
}

func extractOperatorSpec(obj *unstructured.Unstructured, fieldManager string) (*applyoperatorv1.OperatorSpecApplyConfiguration, error) {
	castObj := &operatorv1.Authentication{} // need a real type to extract this from
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, castObj); err != nil {
		return nil, fmt.Errorf("unable to convert to Authentication: %w", err)
	}
	ret, err := applyoperatorv1.ExtractAuthentication(castObj, fieldManager)
	if err != nil {
		return nil, fmt.Errorf("unable to extract fields for %q: %w", fieldManager, err)
	}
	if ret.Spec == nil {
		return nil, nil
	}
	return &ret.Spec.OperatorSpecApplyConfiguration, nil
}

func extractOperatorStatus(obj *unstructured.Unstructured, fieldManager string) (*applyoperatorv1.OperatorStatusApplyConfiguration, error) {
	castObj := &operatorv1.Authentication{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, castObj); err != nil {
		return nil, fmt.Errorf("unable to convert to Authentication: %w", err)
	}
	ret, err := applyoperatorv1.ExtractAuthenticationStatus(castObj, fieldManager)
	if err != nil {
		return nil, fmt.Errorf("unable to extract fields for %q: %w", fieldManager, err)
	}

	if ret.Status == nil {
		return nil, nil
	}
	return &ret.Status.OperatorStatusApplyConfiguration, nil
}
