package detector

import (
	"context"
	"fmt"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	zktick "github.com/zerok-ai/zk-utils-go/ticker"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/inspectors"
	"zk-daemonset/internal/k8utils"
	"zk-daemonset/internal/models"
	"zk-daemonset/internal/storage"
)

var (
	ImageStore          *storage.ImageStore
	ResourceDetailStore *storage.ResourceDetailStore
	ticker              *zktick.TickerTask
)

var detectLoggerTag = "Detector"

const (
	resourceSyncDuration = 2 * time.Minute
)

func Start(cfg config.AppConfigs) error {
	// initialize the image store
	ImageStore = storage.GetNewImageStore(cfg)
	ResourceDetailStore = storage.GetNewPodDetailsStore(cfg)

	// scan existing pods for runtimes
	err := ScanExistingPods()
	if err != nil {
		zklogger.Error(detectLoggerTag, "error in ScanExistingPods %v\n", err)
		return err
	}

	ticker = zktick.GetNewTickerTask("resource_sync", resourceSyncDuration, periodicSync)
	ticker.Start()

	go func() {
		zklogger.Debug(detectLoggerTag, "Adding watcher to services.")
		err := AddWatcherToServices()
		if err != nil {
			zklogger.Error(detectLoggerTag, "error in AddWatcherToServices %v\n", err)
		}
	}()

	// watch pods as they come up for any new image data
	return AddWatcherToPods()
}

func periodicSync() {
	zklogger.Debug(detectLoggerTag, "periodicSync: ")
	err := ScanExistingPods()
	if err != nil {
		zklogger.Error(detectLoggerTag, "error in ScanExistingPods %v\n", err)
	}
	err = scanExistingServices()
	if err != nil {
		zklogger.Error(detectLoggerTag, "error in scanExistingServices %v\n", err)
	}
}

func scanExistingServices() error {
	services, err := k8utils.GetServicesInCluster()
	if err != nil {
		return err
	}
	for _, service := range services.Items {
		err := storeServiceDetails(&service)
		if err != nil {
			zklogger.Error(detectLoggerTag, "error %v\n", err)
		}
	}
	return nil
}

func ScanExistingPods() error {

	podList, err := k8utils.GetPodsInCurrentNode()
	if err != nil {
		return err
	}
	//// Scan all pods for image data
	//containerResultsFromPods, err := GetContainerResultsForAllPods(podList)
	//if err != nil {
	//	return err
	//}

	for _, pod := range podList.Items {
		err = storePodDetails(&pod)
		if err != nil {
			zklogger.Error(detectLoggerTag, "error %v\n", err)
		}
	}

	// update the new results
	//err = ImageStore.SetContainerRuntimes(containerResultsFromPods)
	//if err != nil {
	//	return err
	//}
	return err
}

func AddWatcherToServices() error {
	clientSet, err := k8utils.GetK8sClientSet()
	if err != nil {
		return err
	}

	// Create a SharedInformerFactory for services.
	factory := informers.NewSharedInformerFactory(clientSet, 0)

	// Create a Service informer.
	serviceInformer := factory.Core().V1().Services()

	// Add event handlers to the informer.
	_, err = serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := obj.(*v1.Service)
			zklogger.Debug(detectLoggerTag, "Add service event received for name %v.", service.Name)
			handleServiceEvent(service)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newService := newObj.(*v1.Service)
			zklogger.Debug(detectLoggerTag, "Update service event received for name %v.", newService.Name)
			handleServiceEvent(newService)
		},
		DeleteFunc: func(obj interface{}) {
			service := obj.(*v1.Service)
			zklogger.Debug(detectLoggerTag, "Delete service event received for name %v.", service.Name)
			//Do Nothing.
		},
	})
	if err != nil {
		return err
	}

	// Start the informer.
	stopCh := make(chan struct{})
	defer close(stopCh)

	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// Run the informer until a signal is received.
	<-wait.NeverStop
	return nil
}

func AddWatcherToPods() error {
	clientSet, err := k8utils.GetK8sClientSet()
	if err != nil {
		return err
	}

	// Create a SharedInformerFactory for pods.
	factory := informers.NewSharedInformerFactory(clientSet, 0)

	// Create a Pod informer.
	podInformer := factory.Core().V1().Pods()

	// Add event handlers to the informer.
	_, err = podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			zklogger.Debug(detectLoggerTag, "Add pod event received for name %v.", pod.Name)
			handlePodEvent(pod) // Adjust this function to handle pod events
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*v1.Pod)
			zklogger.Debug(detectLoggerTag, "Update pod event received for name %v.", newPod.Name)
			handlePodEvent(newPod) // Adjust this function to handle pod events
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			zklogger.Debug(detectLoggerTag, "Delete pod event received for name %v.", pod.Name)
			// Do Nothing or handle deletion event if necessary.
		},
	})
	if err != nil {
		return err
	}

	// Start the informer.
	stopCh := make(chan struct{})
	defer close(stopCh)

	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// Run the informer until a signal is received.
	<-wait.NeverStop
	return nil
}

// handleServiceEvent handles service events
func handleServiceEvent(service *v1.Service) {

	zklogger.Debug(detectLoggerTag, "handleServiceEvent: for service %s ", service.Name)
	err := storeServiceDetails(service)
	if err != nil {
		zklogger.Error(detectLoggerTag, "error %v\n", err)
	}
}

// handlePodEvent handles pod events
func handlePodEvent(pod *v1.Pod) {

	zklogger.Debug(detectLoggerTag, "handlePodEvent: for pod %s", pod.Name)

	// 1. find language for each container from the Pod
	//containerResults := GetAllContainerRuntimes(pod)

	// 2. find pod IP to pod details for each Pod
	err := storePodDetails(pod)
	if err != nil {
		zklogger.Error(detectLoggerTag, "error %v\n", err)
	}

	// 3. update the new results
	//err = ImageStore.SetContainerRuntimes(containerResults)
	//if err != nil {
	//	zklogger.Error(detectLoggerTag, "error %v\n", err)
	//	return
	//}
}

func storePodDetails(pod *v1.Pod) error {
	podIp, podResults := GetPodDetails(pod)
	return ResourceDetailStore.SetPodDetails(podIp, podResults)
}

func storeServiceDetails(s *v1.Service) error {
	serviceDetails := models.ServiceDetails{}
	serviceDetails.Metadata = models.ServiceMetadata{
		ServiceName: fmt.Sprintf("%v/%v", s.ObjectMeta.Namespace, s.ObjectMeta.Name),
	}
	serviceIp := s.Spec.ClusterIP
	return ResourceDetailStore.SetServiceDetails(serviceIp, serviceDetails)
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

func GetContainerResultsForAllPods(podList *v1.PodList) ([]models.ContainerRuntime, error) {

	containerResults := []models.ContainerRuntime{}

	zklogger.Debug(detectLoggerTag, "Found %d pods on the node ", len(podList.Items))

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
			zklogger.Error(detectLoggerTag, "error while getting processes of a container ", processes)
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
