package webhook

import (
	"fmt"
	"log"
	"strings"
	"time"

	"main.go/global"
	modelWebhook "main.go/model/webhook"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type RecommendationService struct{}

var recommendationGVR = schema.GroupVersionResource{
	Group:    "bcs.finops.io",
	Version:  "v1alpha1",
	Resource: "recommendations",
}

func (s *RecommendationService) InitK8s() {
	// 1. Initialize Config
	var config *rest.Config
	var err error

	k8sConfig := global.GVA_CONFIG.K8s
	if k8sConfig.KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", k8sConfig.KubeConfig)
	} else if k8sConfig.Host != "" && k8sConfig.BearerToken != "" {
		config = &rest.Config{
			Host:        k8sConfig.Host,
			BearerToken: k8sConfig.BearerToken,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// 2. Initialize Dynamic Client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create dynamic client: %v", err)
	}
	global.GVA_K8S_DYNAMIC = dynamicClient

	// 3. Initialize Dynamic Informer Factory
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 10*time.Minute)
	informer := factory.ForResource(recommendationGVR).Informer()

	// 4. Add Custom Indexer
	err = informer.AddIndexers(cache.Indexers{
		"targetWorkloadIndex": func(obj interface{}) ([]string, error) {
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return nil, nil
			}
			labels := u.GetLabels()
			targetNs := labels["bcs.finops.io/recommendation-target-namespace"]
			targetName := labels["bcs.finops.io/recommendation-target-name"]
			cluster, _, _ := unstructured.NestedString(u.Object, "spec", "cluster")

			if targetNs != "" && targetName != "" {
				return []string{fmt.Sprintf("%s/%s/%s", cluster, targetNs, targetName)}, nil
			}
			return nil, nil
		},
	})
	if err != nil {
		log.Fatalf("Failed to add indexer: %v", err)
	}

	global.GVA_K8S_INDEXER = informer.GetIndexer()

	// 5. Start Informer
	stopCh := make(chan struct{})
	// We do not close stopCh to keep the informer running

	log.Println("Starting Informer and waiting for cache sync...")
	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		log.Fatalf("Failed to sync cache for recommendations CR")
	}
	log.Println("Cache synced successfully. Webhook is ready.")
}

func (s *RecommendationService) MutatePod(pod *corev1.Pod, namespace string) ([]modelWebhook.JSONPatch, error) {
	var patches []modelWebhook.JSONPatch

	// 1. Get Workload Info
	workloadName, _ := s.getWorkloadInfo(pod)
	if workloadName == "" {
		return patches, nil
	}

	// 2. Get Recommendation from Cache
	recommendationMap := s.getRecommendationFromCache(namespace, workloadName)
	if len(recommendationMap) == 0 {
		return patches, nil
	}

	// 3. Generate Patches
	for i, container := range pod.Spec.Containers {
		targetRes, ok := recommendationMap[container.Name]
		if !ok || targetRes.CPU == "" {
			continue
		}

		// Fix Bug 2: Check limit
		targetQty, err := resource.ParseQuantity(targetRes.CPU)
		if err != nil {
			log.Printf("Failed to parse recommended CPU %s: %v", targetRes.CPU, err)
			continue
		}

		if container.Resources.Limits != nil {
			if limitQty, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				if targetQty.Cmp(limitQty) > 0 {
					log.Printf("Recommended CPU %s exceeds limit %s for container %s, skipping patch", targetRes.CPU, limitQty.String(), container.Name)
					continue
				}
			}
		}

		hasRequests := container.Resources.Requests != nil
		// Fix Bug 1: Correctly check map key
		hasCPU := false
		if hasRequests {
			_, hasCPU = container.Resources.Requests[corev1.ResourceCPU]
		}

		patchOp := "replace"
		if !hasRequests {
			patches = append(patches, modelWebhook.JSONPatch{
				Op:   "add",
				Path: fmt.Sprintf("/spec/containers/%d/resources/requests", i),
				Value: map[string]string{
					"cpu": targetRes.CPU,
				},
			})
			continue
		} else if !hasCPU {
			patchOp = "add"
		}

		patches = append(patches, modelWebhook.JSONPatch{
			Op:    patchOp,
			Path:  fmt.Sprintf("/spec/containers/%d/resources/requests/cpu", i),
			Value: targetRes.CPU,
		})

		log.Printf("[O(1) Cache Hit] Pod: %s, Container: %s, Set CPU: %s", pod.GenerateName, container.Name, targetRes.CPU)
	}

	return patches, nil
}

func (s *RecommendationService) getRecommendationFromCache(namespace, workloadName string) map[string]struct {
	CPU    string
	Memory string
} {
	if global.GVA_K8S_INDEXER == nil {
		return nil
	}

	targetCluster := global.GVA_CONFIG.System.ClusterId
	indexKey := fmt.Sprintf("%s/%s/%s", targetCluster, namespace, workloadName)

	objs, err := global.GVA_K8S_INDEXER.ByIndex("targetWorkloadIndex", indexKey)
	if err != nil || len(objs) == 0 {
		return nil
	}

	cr, ok := objs[0].(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	if cr == nil {
		return nil
	}

	status, found, err := unstructured.NestedMap(cr.Object, "status")
	if !found || err != nil {
		return nil
	}

	recommendedValStr, ok := status["recommendedValue"].(string)
	if !ok || recommendedValStr == "" {
		return nil
	}

	var recValue modelWebhook.RecommendedValue
	if err := yaml.Unmarshal([]byte(recommendedValStr), &recValue); err != nil {
		log.Printf("Failed to unmarshal recommendedValue: %v", err)
		return nil
	}

	result := make(map[string]struct {
		CPU    string
		Memory string
	})
	for _, c := range recValue.ResourceRequest.Containers {
		result[c.ContainerName] = struct {
			CPU    string
			Memory string
		}{
			CPU:    c.Target.CPU,
			Memory: c.Target.Memory,
		}
	}

	return result
}

func (s *RecommendationService) getWorkloadInfo(pod *corev1.Pod) (name string, kind string) {
	if len(pod.OwnerReferences) == 0 {
		return "", ""
	}
	owner := pod.OwnerReferences[0]

	if owner.Kind == "ReplicaSet" {
		lastDash := strings.LastIndex(owner.Name, "-")
		if lastDash > 0 {
			return owner.Name[:lastDash], "Deployment"
		}
	}
	return owner.Name, owner.Kind
}
