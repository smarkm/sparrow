package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

	kInformerFactory.Start(stopCh)
	splitInformerFactory.Start(stopCh)
	http.HandleFunc("/inject", func(w http.ResponseWriter, r *http.Request) {
		klog.Info(r)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Errorf("could not read request body: %v", err)
		}
		klog.Info(string(body))
		// if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	fmt.Errorf("unsupported content type %s, only %s is supported", contentType, jsonContentType)
		// }
	})
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		klog.Info(req)
	})
	go func() {
		path := "/go/src/github.com/servicemeshinterface/sparrow/sidecar/"
		err := http.ListenAndServeTLS(":6001", path+"server.crt", path+"server.key", nil)
		if err != nil {
			klog.Error(err)
		}
	}()

	klog.Info("Controller started")
	<-stopCh
}
