package main

import (
	"flag"
	nodeporthandler "github.com/code-ready/routes-controller/pkg/node-port-handler"
	routeshandler "github.com/code-ready/routes-controller/pkg/routes-handler"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var debug bool
	var kubeconfig string
	var master string

	// setup args
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.BoolVar(&debug, "debug", false, "Print debug info")
	flag.StringVar(&master, "master", "", "master url")
	flag.Parse()

	// setup logging
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	// build config
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return err
	}

	// setup informer stop channel
	stop := make(chan struct{})
	defer close(stop)

	// run node port handler
	nodePortClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	nodePortHandler := nodeporthandler.NodePortHandler(nodePortClientSet)
	go func() {
		nodePortHandler.Run(stop)
	}()

	// run routes handler
	routesClientSet, err := routeclientset.NewForConfig(config)
	if err != nil {
		return err
	}
	routePortHandler := routeshandler.RoutesHandler(routesClientSet)
	go func() {
		routePortHandler.Run(stop)
	}()

	return nil
}
