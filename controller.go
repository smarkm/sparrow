package main

import (
	"fmt"
	"time"

	"github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/specs/clientset/versioned/scheme"
	splitcs "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	splitinformers "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/informers/externalversions/split/v1alpha3"
	listers "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/listers/split/v1alpha3"
	kcorev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

//SplitController Managed Split function
type SplitController struct {
	kclientset      kubernetes.Interface
	splitclinetsete splitcs.Interface

	deployLister appslisters.DeploymentLister
	deploySynced cache.InformerSynced

	splitLister listers.TrafficSplitLister
	splitSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

//NewSplitController instance Contoller
func NewSplitController(
	kclientset kubernetes.Interface,
	splitclientset splitcs.Interface,
	deployInformer informers.DeploymentInformer,
	splitInformer splitinformers.TrafficSplitInformer) *SplitController {

	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Create event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kclientset.CoreV1().Events("")})

	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, kcorev1.EventSource{Component: "TrafficSplit-Controller"})

	controller := &SplitController{
		kclientset:      kclientset,
		splitclinetsete: splitclientset,
		deployLister:    deployInformer.Lister(),
		deploySynced:    deployInformer.Informer().HasSynced,
		splitLister:     splitInformer.Lister(),
		splitSynced:     splitInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TrafficSplit"),
		recorder:        recorder,
	}
	klog.Info("Create split event handler")

	splitInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return controller
}
func (c *SplitController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Start Steward controller")

	klog.Info("Warting for informer cache to sync")
	if ok := cache.WaitForCacheSync(stopCh); !ok {
		return fmt.Errorf("failed wait for informer cache sync")
	}

	klog.Info("Start worker ....")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Worker started ...")
	<-stopCh
	klog.Info("Shutdown workers")
	return nil
}

func (c *SplitController) runWorker() {
	// for c.processNextWorkItem() {

	// }
}
