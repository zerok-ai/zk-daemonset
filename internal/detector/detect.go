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
	"zerok-deamonset/internal/config"
	"zerok-deamonset/internal/inspectors"
	"zerok-deamonset/internal/k8utils"
	"zerok-deamonset/internal/models"
	"zerok-deamonset/internal/process"
	"zerok-deamonset/pkg/storage"
	"zerok-deamonset/pkg/zkclient/controller"

	types "zerok-deamonset/internal/models"
)

var (
	ImageStore *storage.ImageStore
)

func Start(cfg config.AppConfigs) error {

	// initialize the image store
	ImageStore = storage.GetNewImageStore(cfg)

	// initialize injector client
	injectorClient := &controller.InjectorClient{
		ContainerResults: []types.ContainerRuntime{},
	}

	// populate injectorClient
	ScanExistingPods(injectorClient)

	// watch pods as they come up for any new image data
	AddWatcherToPods(injectorClient)

	return nil
}

func ScanExistingPods(injectorClient *controller.InjectorClient) {

	fmt.Println("scanning existing pods")

	// 1. Pull data from ImageStore
	containerRuntimeMap := ImageStore.GetAllContainerRuntimes()

	// 2. Scan all pods for image data
	containerResultsFromPods := GetContainerResultsForAllPods(false)

	// 3. find diff and sync
	findDiffAndSync(containerRuntimeMap, containerResultsFromPods)
}

func AddWatcherToPods(injectorClient *controller.InjectorClient) {

	clientSet := k8utils.GetK8sClientSet()

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

		handlePodEvent(item.(*v1.Pod))
		queue.Done(item)
	}
}

// handlePodEvent handles pod events
func handlePodEvent(pod *v1.Pod) {

	//fmt.Printf("handling Pod event for %s/%s on node %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

	// 1. find language for each container
	containerResults := GetAllContainerRuntimes(pod)

	// 2. get existing value for containerruntime in Imagestore
	existingContainerRuntimeMap := map[string]*models.ContainerRuntime{}
	for _, containerResult := range containerResults {

		existingContainRuntime, err := ImageStore.GetContainerRuntime(containerResult.Image)
		if err != nil {
			continue
		}

		existingContainerRuntimeMap[containerResult.Image] = existingContainRuntime
	}

	// 3. find diff and sync
	findDiffAndSync(existingContainerRuntimeMap, containerResults)
}

func findDiffAndSync(existingContainerRuntimeMap map[string]*models.ContainerRuntime, containerResultsFromPods []models.ContainerRuntime) {

	// 1. Find the diff between the data in redis and the data from pods
	diffMapContainerRuntime := []models.ContainerRuntime{}
	for _, containerRuntime := range containerResultsFromPods {

		pushNewValue := false

		// get object from image store
		imgStoreContainerRuntime, ok := existingContainerRuntimeMap[containerRuntime.Image]
		if ok {
			// if present, compare if the values are different
			pushNewValue = !imgStoreContainerRuntime.Equals(containerRuntime)
		} else {
			// not found, push the containerRuntime
			pushNewValue = true
		}

		// if the containerRuntime is different push in the `diffMapContainerRuntime`
		if pushNewValue {
			diffMapContainerRuntime = append(diffMapContainerRuntime, containerRuntime)

			fmt.Printf("new containerRuntime = %v\n", containerRuntime)
		} else {
			fmt.Printf("containerRuntime value already exists for = %s\n", containerRuntime.Image)
		}
	}

	// 2. Add new data to ImageStore
	err := ImageStore.SetContainerRuntimes(diffMapContainerRuntime)
	if err != nil {
		return
	}
}

// watchPods watches for pod events of pods in current node and sends them to a workqueue
func watchPods(ctx context.Context, clientset *kubernetes.Clientset, queue workqueue.RateLimitingInterface) {

	// get the name of the current node
	node := k8utils.GetCurrentNodeName()

	// Create a field selector to watch for pods on a specific node
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": node})

	// create a pod watcher
	watcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, selector)

	// add event handlers to the watcher
	_, _controller := cache.NewInformer(watcher, &v1.Pod{}, time.Second*0, cache.ResourceEventHandlerFuncs{
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
	go _controller.Run(ctx.Done())
}

func GetContainerResultsForAllPods(allPods bool) []models.ContainerRuntime {
	podList := k8utils.GetPodsInCurrentNode(allPods)
	containerResults := []models.ContainerRuntime{}
	for _, pod := range podList.Items {
		temp := GetAllContainerRuntimes(&pod)
		containerResults = append(containerResults, temp...)
	}
	return containerResults
}

func GetAllContainerRuntimes(pod *v1.Pod) []models.ContainerRuntime {

	targetPodUID := string(pod.UID)
	targetContainers := pod.Status.ContainerStatuses

	// iterate through the containers
	var containerResults []models.ContainerRuntime
	for _, container := range targetContainers {

		processes, err := process.FindProcessInContainer(targetPodUID, container.Name)
		if err != nil {
			fmt.Println("caught error while getting processes ", processes)
			continue
		}
		languages := inspectors.DetectLanguageOfAllProcesses(processes)

		containerResults = append(containerResults, models.ContainerRuntime{
			Image:    container.Image,
			ImageID:  container.ImageID,
			Language: languages,
		})

	}
	return containerResults
}
