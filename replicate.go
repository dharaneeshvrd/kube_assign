package main

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kube KubeClient) replicateConfMap(configMaps []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error in replicateConfMap: %v\n", r)
		}
	}()

	fmt.Printf("Num of Config Maps to replicate: %d\n", len(configMaps))
	for confMapIndex := range configMaps {
		confMap, err := kube.SourceClient.CoreV1().ConfigMaps(kube.SourceNameSpace).Get(context.TODO(), configMaps[confMapIndex], metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error while retrieving ConfigMap %s: %v\n", configMaps[confMapIndex], err)
		}
		var confMapN v1.ConfigMap
		confMapN.Data = confMap.Data
		confMapN.Name = confMap.Name
		confMapN.Labels = confMap.Labels
		confMapN.Namespace = kube.TargetNameSpace

		_, err = kube.TargetClient.CoreV1().ConfigMaps(kube.TargetNameSpace).Create(context.TODO(), &confMapN, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Error while creating ConfigMap %s in target kube cluster: %v\n", configMaps[confMapIndex], err)
		}

		fmt.Printf("Replicated Config Map: %v\n", confMapN.Name)
	}
}

func (kube KubeClient) replicateSecrets(secrets []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error in replicateSecrets: %v\n", r)
		}
	}()

	fmt.Printf("Num of Secrets to replicate: %d\n", len(secrets))
	for secretIndex := range secrets {
		secret, err := kube.SourceClient.CoreV1().Secrets(kube.SourceNameSpace).Get(context.TODO(), secrets[secretIndex], metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error while retrieving ConfigMap %s: %v\n", secrets[secretIndex], err)
		}
		var secretN v1.Secret
		secretN.Data = secret.Data
		secretN.Name = secret.Name
		secretN.Labels = secret.Labels
		secretN.Namespace = kube.TargetNameSpace

		_, err = kube.TargetClient.CoreV1().Secrets(kube.TargetNameSpace).Create(context.TODO(), &secretN, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Error while creating ConfigMap %s in target kube cluster: %v\n", secrets[secretIndex], err)
		}

		fmt.Printf("Replicated Secret: %v\n", secret.Name)
	}
}

func (kube KubeClient) replicatePvc(pvcL []string) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error in replicatePvc: %v\n", r)
		}
	}()

	fmt.Printf("Num of PVC's to replicate: %d\n", len(pvcL))
	for pvcIndex := range pvcL {
		pvc, err := kube.SourceClient.CoreV1().PersistentVolumeClaims(kube.SourceNameSpace).Get(context.TODO(), pvcL[pvcIndex], metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error while retrieving PVC %s: %v\n", pvcL[pvcIndex], err)
		}

		var pvcN v1.PersistentVolumeClaim
		pvcN.Spec = pvc.Spec
		pvcN.Name = pvc.Name
		pvcN.Labels = pvc.Labels
		pvcN.Namespace = kube.TargetNameSpace

		pvcSelector := pvc.Spec.Selector
		pvcSelectorLabels := pvcSelector.MatchLabels
		pvcSelectorLabelsS, _ := json.Marshal(pvcSelectorLabels)
		pvL, err := kube.SourceClient.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{LabelSelector: string(pvcSelectorLabelsS)})
		if err != nil {
			fmt.Printf("Error retrieving PV with Labels %v : %v\n", pvcSelectorLabelsS, err)
		}

		for pvIndex := range pvL.Items {
			pv := pvL.Items[pvIndex]
			pv.Namespace = kube.TargetNameSpace
			kube.TargetClient.CoreV1().PersistentVolumes().Create(context.TODO(), &pv, metav1.CreateOptions{})
			fmt.Printf("Created PV: %v\n", pv.Name)
		}

		kube.TargetClient.CoreV1().PersistentVolumeClaims(kube.TargetNameSpace).Create(context.TODO(), &pvcN, metav1.CreateOptions{})
		fmt.Printf("Created PVC: %v\n", pvcN.Name)
	}
}

func (kube KubeClient) replicatePods() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Error in replicatePods: %v\n", r)
		}
	}()

	fmt.Println("Replication Started ...")

	pods, err := kube.SourceClient.CoreV1().Pods(kube.SourceNameSpace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Printf("Number of Pods present in the cluster: %d\n", len(pods.Items))

	for podIndex := range pods.Items {
		pod := pods.Items[podIndex]

		if string(pod.Status.Phase) != "Running" {
			continue
		}

		fmt.Printf("Replicating %v\n", pod.Name)
		var newPod v1.Pod
		configMapToReplicate := make([]string, 0)
		secretsToReplicate := make([]string, 0)
		pvcToReplicate := make([]string, 0)

		for conIndex := range pod.Spec.Containers {
			for envIndex := range pod.Spec.Containers[conIndex].Env {
				envVar := pod.Spec.Containers[conIndex].Env[envIndex]
				if envVar.ValueFrom != nil {
					if envVar.ValueFrom.ConfigMapKeyRef != nil {
						configMapToReplicate = append(configMapToReplicate, envVar.ValueFrom.ConfigMapKeyRef.Name)
					}
					if envVar.ValueFrom.SecretKeyRef != nil {
						secretsToReplicate = append(secretsToReplicate, envVar.ValueFrom.ConfigMapKeyRef.Name)
					}
				}
			}

			for envFromIndex := range pod.Spec.Containers[conIndex].EnvFrom {
				envFrom := pod.Spec.Containers[conIndex].EnvFrom[envFromIndex]
				if envFrom.ConfigMapRef != nil {
					configMapToReplicate = append(configMapToReplicate, envFrom.ConfigMapRef.Name)
				}
				if envFrom.SecretRef != nil {
					secretsToReplicate = append(secretsToReplicate, envFrom.SecretRef.Name)
				}
			}
		}

		for volIndex := range pod.Spec.Volumes {
			vol := pod.Spec.Volumes[volIndex]
			if vol.ConfigMap != nil {
				configMapToReplicate = append(configMapToReplicate, vol.ConfigMap.Name)
			}

			if vol.PersistentVolumeClaim != nil {
				pvcToReplicate = append(pvcToReplicate, vol.PersistentVolumeClaim.ClaimName)
			}
		}

		fmt.Println("Replicating Config Map ...")
		kube.replicateConfMap(configMapToReplicate)
		fmt.Println("Config Map replication completed")
		fmt.Println("Replicating Secrets ...")
		kube.replicateSecrets(secretsToReplicate)
		fmt.Println("Secret replication completed")
		fmt.Println("Replicating PVC & PV ...")
		kube.replicatePvc(pvcToReplicate)
		fmt.Println("PVC & PV replication completed")

		newPod.Name = pod.Name
		newPod.Labels = pod.Labels
		newPod.Spec = pod.Spec
		newPod.Namespace = kube.TargetNameSpace

		_, err := kube.TargetClient.CoreV1().Pods(kube.TargetNameSpace).Create(context.TODO(), &newPod, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Error while replicating %v: %v\n", pod.Name, err)
		}
		fmt.Println("Replicated ...")
	}
}
