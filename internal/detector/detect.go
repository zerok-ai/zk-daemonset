package detector

import (
	"context"
	"fmt"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/inspectors"
	"zk-daemonset/internal/k8utils"
	"zk-daemonset/internal/models"
	"zk-daemonset/internal/storage"
)

var (
	ImageStore     *storage.ImageStore
	PodDetailStore *storage.PodDetailStore
	ticker         *zktick.TickerTask
)

func Start(cfg config.AppConfigs) error {
	// initialize the image store
	ImageStore = storage.GetNewImageStore(cfg)
	PodDetailStore = storage.GetNewPodDetailsStore(cfg)

	// scan existing pods for runtimes
	err := ScanExistingPods()
	if err != nil {
		log.Default().Printf("error in ScanExistingPods %v\n", err)
		return err
	}

	var duration = 10 * time.Minute
	ticker = zktick.GetNewTickerTask("scenario_sync", duration, periodicSync)
	ticker.Start()
	// watch pods as they come up for any new image data
	return AddWatcherToPods()
}

func periodicSync() {
	fmt.Printf("periodicSync: \n")
	err := ScanExistingPods()
	if err != nil {
		log.Default().Printf("error in ScanExistingPods %v\n", err)
	}
}

func ScanExistingPods() error {

	// Scan all pods for image data
	containerResultsFromPods, err := GetContainerResultsForAllPods()
	if err != nil {
		return err
	}

	// update the new results
	err = ImageStore.SetContainerRuntimes(containerResultsFromPods)
	if err != nil {
		return err
	}
	return err
}

func AddWatcherToPods() error {
	clientSet, err := k8utils.GetK8sClientSet()

	if err != nil {
		return err
	}

	// create a context and a work queue
	ctx := context.Background()
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// watch for pod events and send them to the workqueue
	watchPods(ctx, clientSet, queue)

	// process pod events from the workqueue
	for {
		item, shutdown := queue.Get()
		if shutdown {
			return nil
		}

		handlePodEvent(item.(*v1.Pod))
		queue.Done(item)
	}
}

// handlePodEvent handles pod events
func handlePodEvent(pod *v1.Pod) {

	fmt.Printf("\n\nhandlePodEvent: for pod %s\n", pod.Name)

	// 1. find language for each container from the Pod
	containerResults := GetAllContainerRuntimes(pod)

	// 2. find pod IP to pod details for each Pod
	podIp, podResults := GetPodDetails(pod)
	PodDetailStore.SetPodDetails(podIp, podResults)

	// 3. update the new results
	err := ImageStore.SetContainerRuntimes(containerResults)
	if err != nil {
		log.Default().Printf("error %v\n", err)
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

func GetContainerResultsForAllPods() ([]models.ContainerRuntime, error) {

	podList, err := k8utils.GetPodsInCurrentNode()
	if err != nil {
		return nil, err
	}
	containerResults := []models.ContainerRuntime{}

	fmt.Printf("Found %d pods on the node\n", len(podList.Items))

	for _, pod := range podList.Items {
		temp := GetAllContainerRuntimes(&pod)
		containerResults = append(containerResults, temp...)
	}
	return containerResults, nil
}

func GetAllContainerRuntimes(pod *v1.Pod) []models.ContainerRuntime {

	targetPodUID := string(pod.UID)
	targetContainers := pod.Status.ContainerStatuses

	// iterate through the containers
	containerResults := []models.ContainerRuntime{}
	for _, container := range targetContainers {

		processes, err := FindProcessInContainer(targetPodUID, container.Name)
		if err != nil {
			fmt.Println("error while getting processes of a container ", processes)
			continue
		}
		languages, processName := inspectors.DetectLanguageOfAllProcesses(processes)

		var cmdArray []string
		envMap := make(map[string]string)

		for _, process := range processes {
			cmdArray = append(cmdArray, process.CmdLine)
			for key, value := range process.EnvMap {
				envMap[key] = value
			}
		}

		if len(languages) > 0 {
			containerResults = append(containerResults, models.ContainerRuntime{
				Image:    container.Image,
				ImageID:  container.ImageID,
				Language: languages,
				Process:  processName,
				Cmd:      cmdArray,
				EnvMap:   envMap,
			})
		}

	}
	return containerResults
}
