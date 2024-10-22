package libraryinputresources

import (
	"context"
	"errors"
	"fmt"
	"github.com/openshift/library-go/pkg/manifestclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/jsonpath"
	"net/http"
	"os"
	"path"
)

func WriteRequiredInputResourcesFromMustGather(ctx context.Context, inputResources *InputResources, mustGatherDir, targetDir string) error {
	actualResources, err := GetRequiredInputResourcesFromMustGather(ctx, inputResources, mustGatherDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("unable to create %q: %w", targetDir, err)
	}

	errs := []error{}
	for _, currResource := range actualResources {
		if err := WriteResource(currResource, targetDir); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func GetRequiredInputResourcesFromMustGather(ctx context.Context, inputResources *InputResources, mustGatherDir string) ([]*Resource, error) {
	dynamicClient, err := NewDynamicClientFromMustGather(mustGatherDir)
	if err != nil {
		return nil, err
	}

	pertinentUnstructureds, err := GetRequiredInputResourcesForResourceList(ctx, inputResources.ApplyConfigurationResources, dynamicClient)
	if err != nil {
		return nil, err
	}

	return unstructuredToMustGatherFormat(pertinentUnstructureds)
}

func NewDynamicClientFromMustGather(mustGatherDir string) (dynamic.Interface, error) {
	roundTripper := manifestclient.NewRoundTripper(mustGatherDir)
	httpClient := &http.Client{
		Transport: roundTripper,
	}

	dynamicClient, err := dynamic.NewForConfigAndClient(&rest.Config{}, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failure creating dynamicClient for NewDynamicClientFromMustGather: %w", err)
	}

	return dynamicClient, nil
}

func GetRequiredInputResourcesForResourceList(ctx context.Context, resourceList ResourceList, dynamicClient dynamic.Interface) ([]*Resource, error) {
	instances := []*Resource{}
	errs := []error{}

	for _, currResource := range resourceList.ExactResources {
		resourceInstance, err := getExactResource(ctx, dynamicClient, currResource)
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		instances = append(instances, resourceInstance)
	}

	for i, currResourceRef := range resourceList.ResourceReference {
		refIdentifier := fmt.Sprintf("%d", i)
		fieldPathEvaluator := jsonpath.New(refIdentifier)
		fieldPathEvaluator.AllowMissingKeys(true)

		referringGVR := schema.GroupVersionResource{Group: currResourceRef.ReferringResource.Group, Version: currResourceRef.ReferringResource.Version, Resource: currResourceRef.ReferringResource.Resource}
		referringResourceInstance, err := dynamicClient.Resource(referringGVR).Namespace(currResourceRef.ReferringResource.Namespace).Get(ctx, currResourceRef.ReferringResource.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("failed reading referringResource [%v] %#v: %w", refIdentifier, currResourceRef.ReferringResource, err))
			continue
		}

		switch {
		case currResourceRef.ImplicitNamespacedReference != nil:
			err := fieldPathEvaluator.Parse("{" + currResourceRef.ImplicitNamespacedReference.NameJSONPath + "}")
			if err != nil {
				errs = append(errs, fmt.Errorf("error parsing [%v]: %q: %w", refIdentifier, currResourceRef.ImplicitNamespacedReference.NameJSONPath, err))
				continue
			}

			results, err := fieldPathEvaluator.FindResults(referringResourceInstance.UnstructuredContent())
			if err != nil {
				errs = append(errs, fmt.Errorf("unexpected error finding value for %v from %v with jsonPath: %w", refIdentifier, "TODO", err))
				continue
			}

			for _, currResultSlice := range results {
				for _, currResult := range currResultSlice {
					value := currResult.Interface()
					targetResourceName := fmt.Sprint(value)
					targetRef := ExactResource{
						DependsOnResourceTypeIdentifier: currResourceRef.ImplicitNamespacedReference.DependsOnResourceTypeIdentifier,
						Namespace:                       currResourceRef.ImplicitNamespacedReference.Namespace,
						Name:                            targetResourceName,
					}

					resourceInstance, err := getExactResource(ctx, dynamicClient, targetRef)
					if apierrors.IsNotFound(err) {
						continue
					}
					if err != nil {
						errs = append(errs, err)
						continue
					}

					instances = append(instances, resourceInstance)
				}
			}
		}
	}

	return instances, errors.Join(errs...)
}

func getExactResource(ctx context.Context, dynamicClient dynamic.Interface, resourceReference ExactResource) (*Resource, error) {
	gvr := schema.GroupVersionResource{Group: resourceReference.Group, Version: resourceReference.Version, Resource: resourceReference.Resource}
	unstructuredInstance, err := dynamicClient.Resource(gvr).Namespace(resourceReference.Namespace).Get(ctx, resourceReference.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed getting %v: %w", IdentifierForExactResourceRef(&resourceReference), err)
	}

	resourceInstance := &Resource{
		ResourceType: gvr,
		Content:      unstructuredInstance,
	}
	return resourceInstance, nil
}

func IdentifierForExactResourceRef(resourceReference *ExactResource) string {
	return fmt.Sprintf("%s.%s.%s/%s[%s]", resourceReference.Resource, resourceReference.Version, resourceReference.Group, resourceReference.Name, resourceReference.Namespace)
}

func unstructuredToMustGatherFormat(in []*Resource) ([]*Resource, error) {
	type mustGatherKeyType struct {
		gk        schema.GroupKind
		namespace string
	}

	versionsByGroupKind := map[schema.GroupKind]sets.Set[string]{}
	groupKindToResource := map[schema.GroupKind]schema.GroupVersionResource{}
	byGroupKind := map[mustGatherKeyType]*unstructured.UnstructuredList{}
	for _, curr := range in {
		gvk := curr.Content.GroupVersionKind()
		groupKind := curr.Content.GroupVersionKind().GroupKind()
		existingVersions, ok := versionsByGroupKind[groupKind]
		if !ok {
			existingVersions = sets.New[string]()
			versionsByGroupKind[groupKind] = existingVersions
		}
		existingVersions.Insert(gvk.Version)
		groupKindToResource[groupKind] = curr.ResourceType

		mustGatherKey := mustGatherKeyType{
			gk:        groupKind,
			namespace: curr.Content.GetNamespace(),
		}
		existing, ok := byGroupKind[mustGatherKey]
		if !ok {
			existing = &unstructured.UnstructuredList{
				Object: map[string]interface{}{},
			}
			listGVK := guessListKind(curr.Content)
			existing.GetObjectKind().SetGroupVersionKind(listGVK)
			byGroupKind[mustGatherKey] = existing
		}
		existing.Items = append(existing.Items, *curr.Content.DeepCopy())
	}

	errs := []error{}
	for groupKind, currVersions := range versionsByGroupKind {
		if len(currVersions) == 1 {
			continue
		}
		errs = append(errs, fmt.Errorf("groupKind=%v has multiple versions: %v, which prevents serialization", groupKind, sets.List(currVersions)))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	ret := []*Resource{}
	for mustGatherKey, list := range byGroupKind {
		namespacedString := "REPLACE_ME"
		if len(mustGatherKey.namespace) > 0 {
			namespacedString = "namespaces"
		} else {
			namespacedString = "cluster-scoped-resources"
		}

		groupString := mustGatherKey.gk.Group
		if len(groupString) == 0 {
			groupString = "core"
		}
		listAsUnstructured := &unstructured.Unstructured{Object: list.UnstructuredContent()}
		resourceType := groupKindToResource[mustGatherKey.gk]
		ret = append(ret, &Resource{
			Filename: path.Join(namespacedString, mustGatherKey.namespace, groupString, fmt.Sprintf("%s.yaml", resourceType.Resource)),
			Content:  listAsUnstructured,
		})
	}

	return ret, nil
}

func guessListKind(in *unstructured.Unstructured) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   in.GroupVersionKind().Group,
		Version: in.GroupVersionKind().Version,
		Kind:    in.GroupVersionKind().Kind + "List",
	}
}
