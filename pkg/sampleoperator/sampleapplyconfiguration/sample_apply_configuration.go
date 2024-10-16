package sampleapplyconfiguration

import (
	"context"
	"fmt"
	"os"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	applyoperatorv1 "github.com/openshift/client-go/operator/applyconfigurations/operator/v1"
	operatorclient "github.com/openshift/client-go/operator/clientset/versioned"
	operatorinformer "github.com/openshift/client-go/operator/informers/externalversions"
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
	"k8s.io/client-go/rest"
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
}

const componentName = "cluster-example-operator"

// CreateOperatorInputFromMOM this method is usually paired with CreateControllerInputFromControllerContext during the transition.
// This allows us to abstract the creation of clients from those things that depend on those clients so that initialization
// can happen as normal.
func CreateOperatorInputFromMOM(ctx context.Context, momInput libraryapplyconfiguration.ApplyConfigurationInput) (*exampleOperatorInput, error) {
	kubeClient, err := kubernetes.NewForConfigAndClient(&rest.Config{}, momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	configClient, err := configclient.NewForConfigAndClient(&rest.Config{}, momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	operatorClient, err := operatorclient.NewForConfigAndClient(&rest.Config{}, momInput.MutationTrackingClient.GetHTTPClient())
	if err != nil {
		return nil, err
	}

	authenticationOperatorClient, dynamicInformers, err := genericoperatorclient.NewOperatorClientWithClient(
		momInput.Clock,
		momInput.MutationTrackingClient.GetHTTPClient(),
		operatorv1.GroupVersion.WithResource("examples"),
		operatorv1.GroupVersion.WithKind("Example"),
		extractOperatorSpec,
		extractOperatorStatus,
	)
	if err != nil {
		return nil, err
	}

	eventRecorder := events.NewKubeRecorderWithOptions(
		kubeClient.CoreV1().Events("openshift-authentication-operator"),
		events.RecommendedClusterSingletonCorrelatorOptions(),
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
	}, nil
}

func CreateOperatorStarter(ctx context.Context, exampleOperatorInput *exampleOperatorInput) (libraryapplyconfiguration.OperatorStarter, error) {
	ret := &libraryapplyconfiguration.SimpleOperatorStarter{
		Informers: append([]libraryapplyconfiguration.SimplifiedInformerFactory{}, exampleOperatorInput.informers...),
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
		"example",
		exampleOperatorInput.exampleOperatorClient,
		kubeInformersForNamespaces,
		v1helpers.CachedSecretGetter(exampleOperatorInput.kubeClient.CoreV1(), kubeInformersForNamespaces),
		v1helpers.CachedConfigMapGetter(exampleOperatorInput.kubeClient.CoreV1(), kubeInformersForNamespaces),
		exampleOperatorInput.eventRecorder,
	)
	ret.ControllerRunFns = append(ret.ControllerRunFns, libraryapplyconfiguration.AdaptRunFn(resourceSyncer.Run))
	ret.ControllerRunOnceFns = append(ret.ControllerRunOnceFns, libraryapplyconfiguration.AdaptSyncFn(exampleOperatorInput.eventRecorder, resourceSyncer.Sync))

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
