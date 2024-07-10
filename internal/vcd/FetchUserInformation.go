package vcd

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func GetUserName(kubeclient kubernetes.Interface) string {
	var username string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get("csi-fptcloud-basic-auth", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "username" {
			username = string(v)
		}
	}
	return username
}

func GetPassword(kubeclient kubernetes.Interface) string {
	var password string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get("csi-fptcloud-basic-auth", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "password" {
			password = string(v)
		}
	}
	return password
}

// GetAccessToken gets access token of FPTCloud
func GetAccessToken(kubeclient kubernetes.Interface) string {
	var accessToken string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get("fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "access_token" {
			accessToken = string(v)
		}
	}
	return accessToken
}

// GetVPCId gets vpc_id of customer
func GetVPCId(kubeclient kubernetes.Interface) string {
	var vpcID string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get("fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "vpc_id" {
			vpcID = string(v)
		}
	}
	return vpcID
}

// GetClusterID gets cluster_id info of K8S cluster
func GetClusterID(kubeclient kubernetes.Interface) string {
	var clusterID string
	secret, err := kubeclient.CoreV1().Secrets("kube-system").Get("fke-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fke secret")
		klog.Fatalf("Failed to get information of fke secret: %v", err)
	}
	for k, v := range secret.Data {
		if k == "cluster_id" {
			clusterID = string(v)
		}
	}
	return clusterID
}

func GetCallBackURL(kubeclient kubernetes.Interface) string {
	var callbackURL string
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling configmap")
		klog.Fatalf("Failed to get information of autoscaling configmap: %v", err)
	}
	for k, v := range configmaps.Data {
		if k == "callback_url" {
			if err != nil {
				klog.Fatalf("Failed to convert string to integer: %v", err)
			}
			callbackURL = v
		}
	}
	return callbackURL
}