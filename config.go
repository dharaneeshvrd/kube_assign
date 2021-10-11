package main

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	SourceKubeConfFile string `json:"sourceKubeConfFile"`
	TargetKubeConfFile string `json:"targetKubeConfFile"`
	SourceNameSpace    string `json:"sourceNamespace"`
	TargetNameSpace    string `json:"targetNamespace"`
}

type KubeClient struct {
	SourceClient    *kubernetes.Clientset
	TargetClient    *kubernetes.Clientset
	SourceNameSpace string
	TargetNameSpace string
}

func (kube *KubeClient) initKubeClient(confFile string, clientType string) {
	/*
		var kubeconfigS string
		if home := homedir.HomeDir(); home != "" {
			kubeconfigS = filepath.Join(home, ".kube", "config")
		}
	*/

	kubeConfig, _ := clientcmd.BuildConfigFromFlags("", confFile)

	var err error
	switch clientType {
	case "source":
		kube.SourceClient, err = kubernetes.NewForConfig(kubeConfig)
	case "target":
		kube.TargetClient, err = kubernetes.NewForConfig(kubeConfig)
	default:
		fmt.Println("Invalid kube config client type")
	}

	if err != nil {
		panic(err.Error())
	}
}

//InitConfig initializing the config
func initConfig() Config {
	var config Config
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	return config
}
