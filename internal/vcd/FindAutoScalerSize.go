package vcd

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strconv"
)

func GetMaxSize(kubeclient kubernetes.Interface) int {
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling-configmap")
		klog.Fatalf("Failed to get information of autoscaling-configmap: %v", err)
	}
	maxSize := configmaps.Data["max_node_group_size"]
	maxSizeInteger, _ := strconv.Atoi(maxSize)
	return maxSizeInteger
}

func GetMinSize(kubeclient kubernetes.Interface) int {
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling-configmap")
		klog.Fatalf("Failed to get information of autoscaling-configmap: %v", err)
	}
	maxSize := configmaps.Data["min_node_group_size"]
	minSizeInteger, _ := strconv.Atoi(maxSize)
	return minSizeInteger
}
