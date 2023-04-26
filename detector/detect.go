package detector

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
	common "zerok.ai/deamonset/common"
	inspectors "zerok.ai/deamonset/inspectors"
	"zerok.ai/deamonset/process"
	utils "zerok.ai/deamonset/utils"
	zkclient "zerok.ai/deamonset/zkclient"
)

func ScanExistingPods(injectorClient *zkclient.InjectorClient) {
	containerResults := GetContainerResultsForAllPods(false)
	injectorClient.ContainerResults = append(injectorClient.ContainerResults, containerResults...)
	injectorClient.SyncDataWithInjector()
}

func AddWatcherToPods(injectorClient *zkclient.InjectorClient) {

	clientSet := utils.GetK8sClientSet()

	// create a context and a workqueue
	ctx := context.Background()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// watch for pod events and send them to the workqueue
	watchPods(ctx, clientSet, queue)

	// process pod events from the workqueue
	for {
		item, shutdown := queue.Get()
		if shutdown {
			return
		}
		handlePodEvent(item, injectorClient)
		queue.Done(item)
	}
}

// handlePodEvent handles pod events
func handlePodEvent(obj interface{}, injectorClient *zkclient.InjectorClient) {

	pod := obj.(*v1.Pod)
	fmt.Printf("handling Pod event for %s/%s on node %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

	// find language for each container
	containerResults := FindLang(pod, pod.Status.ContainerStatuses)

	injectorClient.ContainerResults = append(injectorClient.ContainerResults, containerResults...)
	injectorClient.SyncDataWithInjector()
}

// watchPods watches for pod events of pods in current node and sends them to a workqueue
func watchPods(ctx context.Context, clientset *kubernetes.Clientset, queue workqueue.RateLimitingInterface) {

	// get the name of the current node
	node := utils.GetCurrentNodeName()

	// Create a field selector to watch for pods on a specific node
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": node})

	// create a pod watcher
	watcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, selector)

	// add event handlers to the watcher
	_, controller := cache.NewInformer(watcher, &v1.Pod{}, time.Second*0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			queue.Add(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			queue.Add(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			// do nothing
		},
	})

	// run the pod watcher
	go controller.Run(ctx.Done())
}

func GetContainerResultsForAllPods(allPods bool) []common.ContainerRuntime {
	podList := utils.GetPodsInCurrentNode(allPods)
	fmt.Println("Pods are ", podList)
	containerResults := []common.ContainerRuntime{}
	for _, pod := range podList.Items {
		temp := FindLang(&pod, pod.Status.ContainerStatuses)
		containerResults = append(containerResults, temp...)
	}
	return containerResults
}

func FindLang(pod *v1.Pod, targetContainers []v1.ContainerStatus) []common.ContainerRuntime {
	targetPodUID := string(pod.UID)
	var containerResults []common.ContainerRuntime
	for _, container := range targetContainers {
		containerName := container.Name

		processes, err := process.FindProcessInContainer(targetPodUID, containerName)
		if err != nil {
			fmt.Println("caught error while getting processes ", processes)
			continue
		}
		processes = inspectors.DetectLanguage(processes)
		fmt.Println(processes)
		containerResults = append(containerResults, common.ContainerRuntime{
			ContainerName: containerName,
			Image:         container.Image,
			ImageID:       container.ImageID,
			Process:       processes,
			PodUID:        targetPodUID,
		})

	}
	fmt.Println(containerResults)
	return containerResults
}
