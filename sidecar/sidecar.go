package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	splitcs "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"
	splitinformers "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/informers/externalversions"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

//JSONPatch RT
type JSONPatch struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

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
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Errorf("could not read request body: %v", err)
		}
		if r.Header.Get("Content-Type") != "application/json" {

		}

		var admReview v1beta1.AdmissionReview
		var deploy appsv1.Deployment

		json.Unmarshal(body, &admReview)
		json.Unmarshal(admReview.Request.Object.Raw, &deploy)
		containers := deploy.Spec.Template.Spec.Containers
		klog.Infof("Resource: %s,API: %s,Object: %s", admReview.Request.Kind.Kind, admReview.APIVersion, containers)
		containers = append(containers, corev1.Container{Image: "busybox", Name: "sparrow-proxy", Command: []string{"busybox"}, Args: []string{"nc", "-l", "8888"}})
		klog.Infof("Resource: %s,API: %s,Object: %s", admReview.Request.Kind.Kind, admReview.APIVersion, containers)

		deploy.Spec.Template.Name = "nginx"

		admResp := v1beta1.AdmissionResponse{}
		pt := v1beta1.PatchTypeJSONPatch
		admResp.PatchType = &pt
		patchs := make([]JSONPatch, 0)
		patchs = append(patchs, JSONPatch{Op: "add", Path: "/spec/template/spec/containers", Value: containers})
		admResp.UID = admReview.Request.UID
		admResp.Patch, err = json.Marshal(patchs)

		if err != nil {
			klog.Errorf("Error: %s", err.Error())
		}
		admResp.Allowed = true
		admReview.Response = &admResp
		body, err = json.Marshal(admReview)
		if err != nil {
			klog.Error(err.Error())
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)
		klog.Infof("Execuate patch %s", err)
		// if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	fmt.Errorf("unsupported content type %s, only %s is supported", contentType, jsonContentType)
		// }
	})
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Execute validate start")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Errorf("could not read request body: %v", err)
		}
		if r.Header.Get("Content-Type") != "application/json" {

		}
		var admReview v1beta1.AdmissionReview
		json.Unmarshal(body, &admReview)
		var deploy appsv1.Deployment
		json.Unmarshal(admReview.Request.Object.Raw, &deploy)

		containers := deploy.Spec.Template.Spec.Containers
		klog.Info("Containers: %s", containers)
		ok := false
		for _, c := range containers {
			if strings.Contains(c.Image, "busybox") {
				ok = true
				break
			}
		}
		resp := &v1beta1.AdmissionResponse{}
		resp.UID = admReview.Request.UID
		resp.Allowed = ok
		if !ok {
			resp.Result = &v1.Status{Code: 403, Message: "not with valid sparrow-proxy "}
		}
		admReview.Response = resp
		body, _ = json.Marshal(admReview)
		w.Write(body)
		klog.Info("Execate validate end")
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
