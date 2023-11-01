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
	"zk-daemonset/internal/k8utils"
	"zk-daemonset/internal/models"
	"zk-daemonset/internal/storage"
)

var (
	ResourceDetailStore *storage.ResourceDetailStore
	ticker              *zktick.TickerTask
)

var detectLoggerTag = "Detector"

const (
	resourceSyncDuration = 2 * time.Minute
)

func Start(cfg config.AppConfigs) error {
	// initialize the store
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

	for _, pod := range podList.Items {
		err = storePodDetails(&pod)
		if err != nil {
			zklogger.Error(detectLoggerTag, "error %v\n", err)
		}
	}
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

	zklogger.Debug(detectLoggerTag, "handlePodEvent: for pod ", pod.Name)

	// 1. find pod IP to pod details for each Pod
	err := storePodDetails(pod)
	if err != nil {
		zklogger.Error(detectLoggerTag, "error %v\n", err)
	}

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
