package main

func main() {
	config := initConfig()
	var KubeClient KubeClient
	KubeClient.initKubeClient(config.SourceKubeConfFile, "source")
	KubeClient.initKubeClient(config.TargetKubeConfFile, "target")

	KubeClient.SourceNameSpace = config.SourceNameSpace
	KubeClient.TargetNameSpace = config.TargetNameSpace

	KubeClient.replicatePods()
}
