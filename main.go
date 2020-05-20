package main

import (
	"flag"
	"time"

	splitcs "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	splitinformers "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/informers/externalversions"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	klog.InitFlags(nil)
	var masterUrl string
	var kubeconfigPath string
	flag.StringVar(&masterUrl, "url", "", "master url")
	flag.StringVar(&kubeconfigPath, "config", "", "kuberconfig")
	flag.Parse()

	stopCh := signals.SetupSignalHandler()
	cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		klog.Fatal("Error load kubeconfig: %s", err.Error())
	}

	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatal("Error build k8s client: %s", err.Error())
	}

	splitClient, err := splitcs.NewForConfig(cfg)
	if err != nil {
		klog.Fatal("Error build split client: %s", err.Error())
	}
	kInformerFactory := informers.NewSharedInformerFactory(kclient, time.Second*30)
	splitInformerFactory := splitinformers.NewSharedInformerFactory(splitClient, time.Second*1)

	//new smi controller here
	splitcontroller := NewSplitController(kclient, splitClient,
		kInformerFactory.Apps().V1().Deployments(),
		splitInformerFactory.Split().V1alpha3().TrafficSplits())

	kInformerFactory.Start(stopCh)
	splitInformerFactory.Start(stopCh)

	if err = splitcontroller.Run(2, stopCh); err != nil {
		klog.Fatal("Error run split controller: %s", err.Error())
	}

	<-stopCh
}
