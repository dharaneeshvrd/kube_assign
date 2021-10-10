package main

import "fmt"

func main() {
	config := initConfig()
	var KubeClient KubeClient
	KubeClient.initKubeClient(config.SourceKubeConfFile, "source")
	KubeClient.initKubeClient(config.TargetKubeConfFile, "target")
	fmt.Println(config.SourceNameSpace, config.TargetNameSpace)
	KubeClient.SourceNameSpace = config.SourceNameSpace
	KubeClient.TargetNameSpace = config.TargetNameSpace

	KubeClient.replicatePods()
}
