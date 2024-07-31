/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"crypto/tls"
	apiv1 "k8s.io/api/core/v1"
	// "repair/clusterstate"
	// "k8s.io/autoscaler/cluster-autoscaler/metrics"
	// "repair/utils/errors"
	// "k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	// schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// isVirtualNode determines if the node is created by virtual kubelet
func isVirtualNode(node *apiv1.Node) bool {
	return node.ObjectMeta.Labels["type"] == "virtual-kubelet"
}

func hasHardInterPodAffinity(affinity *apiv1.Affinity) bool {
	if affinity == nil {
		return false
	}
	if affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	if affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	return false
}

// GetNodeCoresAndMemory extracts cpu and memory resources out of Node object
func GetNodeCoresAndMemory(node *apiv1.Node) (int64, int64) {
	cores := getNodeResource(node, apiv1.ResourceCPU)
	memory := getNodeResource(node, apiv1.ResourceMemory)
	return cores, memory
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

// GetOldestCreateTime returns oldest creation time out of the pods in the set
func GetOldestCreateTime(pods []*apiv1.Pod) time.Time {
	oldest := time.Now()
	for _, pod := range pods {
		if oldest.After(pod.CreationTimestamp.Time) {
			oldest = pod.CreationTimestamp.Time
		}
	}
	return oldest
}

// GetMinSizeNodeGroup gets min size group
func GetMinSizeNodeGroup(kubeclient kube_client.Interface) int {
	var minSizeNodeGroup int
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling configmap")
		klog.Fatalf("Failed to get information of autoscaling configmap: %v", err)
	}
	for k, v := range configmaps.Data {
		if k == "min_node_group_size" {
			value, err := strconv.Atoi(v)
			if err != nil {
				klog.Fatalf("Failed to convert string to integer: %v", err)
			}
			minSizeNodeGroup = value
		}
	}
	return minSizeNodeGroup
}

// GetMaxSizeNodeGroup gets max size group
func GetMaxSizeNodeGroup(kubeclient kube_client.Interface) int {
	var maxSizeNodeGroup int
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("autoscaling-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from autoscaling configmap")
		klog.Fatalf("Failed to get information of autoscaling configmap: %v", err)
	}
	for k, v := range configmaps.Data {
		if k == "max_node_group_size" {
			value, err := strconv.Atoi(v)
			if err != nil {
				klog.Fatalf("Failed to convert string to integer: %v", err)
			}
			maxSizeNodeGroup = value
		}
	}
	return maxSizeNodeGroup
}

// GetCallBackURL gets callback url of portal
func GetCallBackURL(kubeclient kube_client.Interface) string {
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

// GetAccessToken gets access token of FPTCloud
func GetAccessToken(kubeclient kube_client.Interface) string {
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
func GetVPCId(kubeclient kube_client.Interface) string {
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
func GetClusterID(kubeclient kube_client.Interface) string {
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

type Cluster struct {
	Total int `json:"total"`
	Data  []struct {
		ID                  string `json:"id"`
		ClusterSlug         string `json:"cluster_slug"`
		ClusterID           string `json:"cluster_id"`
		VpcID               string `json:"vpc_id"`
		EdgeGatewayID       string `json:"edge_gateway_id"`
		NetworkID           string `json:"network_id"`
		CreatedAt           string `json:"created_at"`
		UpdatedAt           string `json:"updated_at"`
		Name                string `json:"name"`
		Status              string `json:"status"`
		WorkerNodeCount     string `json:"worker_node_count"`
		MasterNodeCount     string `json:"master_node_count"`
		KubernetesVersion   string `json:"kubernetes_version"`
		IsDeleted           string `json:"is_deleted"`
		AwxJobID            string `json:"awx_job_id"`
		AwxParams           string `json:"awx_params"`
		NfsDiskSize         string `json:"nfs_disk_size"`
		NfsStatus           string `json:"nfs_status"`
		IsRunning           string `json:"is_running"`
		ErrorMessage        string `json:"error_message"`
		Templates           string `json:"templates"`
		LoadBalancerSize    string `json:"load_balancer_size"`
		ProcessingMess      string `json:"processing_mess"`
		ClusterType         string `json:"cluster_type"`
		DistributedFirewall string `json:"distributed_firewall"`
	} `json:"data"`
}

// GetIDCluster gets ID of cluster
func GetIDCluster(domainAPI string, vpcID string, accessToken string, clusterID string) string {
	var id string
	var k8sCluster Cluster
	url := domainAPI + "/api/v1/xplat/fke/vpc/" + vpcID + "/kubernetes?page=1&page_size=25"
	token := accessToken
	client := &http.Client{		
		Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    },}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	//log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	dataBody := []byte(body)
	error := json.Unmarshal(dataBody, &k8sCluster)
	if error != nil {
		// if error is not nil
		// print error
		fmt.Println(error)
	}

	//fmt.Println(k8sCluster.Data[0])
	for _, cluster := range k8sCluster.Data {
		//fmt.Println(cluster)
		if cluster.ClusterID == clusterID {
			id = cluster.ID
		}
	}

	defer resp.Body.Close()
	//fmt.Println("ID is: ", id)
	return id
}

// CheckStatusCluster checks if status cluster is Succeeded
func CheckStatusCluster(domainAPI string, vpcID string, accessToken string, clusterID string) bool {
	var isSucceeded bool = false
	var k8sCluster Cluster
	url := domainAPI + "/api/v1/xplat/fke/vpc/" + vpcID + "/kubernetes?page=1&page_size=25"
	token := accessToken
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	dataBody := []byte(body)
	error := json.Unmarshal(dataBody, &k8sCluster)
	if error != nil {
		// if error is not nil
		// print error
		fmt.Println(error)
	}

	//fmt.Println(k8sCluster.Data[0])
	for _, cluster := range k8sCluster.Data {
		//fmt.Println(cluster)
		if cluster.ClusterID == clusterID {
			klog.V(1).Infof("status of cluster is: %v", cluster.Status)
			if cluster.Status == "SCALING" {
				isSucceeded = false
			} else if cluster.Status == "SUCCEEDED" {
				isSucceeded = true
			}
		}
	}

	defer resp.Body.Close()
	// klog.V(1).Infof("isSucceed is: %v", isSucceeded)
	return isSucceeded
}

// CheckErrorStatusCluster Checks if status cluster is Error
func CheckErrorStatusCluster(domainAPI string, vpcID string, accessToken string, clusterID string) bool {
	var isError bool = false
	var k8sCluster Cluster
	url := domainAPI + "/api/v1/xplat/fke/vpc/" + vpcID + "/kubernetes?page=1&page_size=25"
	token := accessToken
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	//log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	dataBody := []byte(body)
	error := json.Unmarshal(dataBody, &k8sCluster)
	if error != nil {
		// if error is not nil
		// print error
		fmt.Println(error)
	}

	//fmt.Println(k8sCluster.Data[0])
	for _, cluster := range k8sCluster.Data {
		//fmt.Println(cluster)
		if cluster.ClusterID == clusterID {
			if cluster.Status == "ERROR" {
				isError = true
			}
		}
	}

	defer resp.Body.Close()
	//fmt.Println("isSucceed is: ", isSucceeded)
	return isError
}


type ScaleUpBodyRequest struct {
	ClusterID      string   `json:"cluster_id"`
	WorkerCount    string   `json:"worker_count"`
	ScaleType      string   `json:"scale_type"`
	WorkerNameList []string `json:"worker_name_list"`
}

// PerformScaleUp performs scale up
func PerformScaleUp(domainAPI string, vpcID string, accessToken string, workerCount int, idCluster string, clusterIDPortal string) (string, error) {
	url := domainAPI + "/api/v1/xplat/fke/vpc/" + vpcID + "/cluster/" + idCluster + "/scale-cluster"
	workerNameList := make([]string, 0)
	bodyRequest := ScaleUpBodyRequest{
		ClusterID:      clusterIDPortal,
		ScaleType:      "up",
		WorkerCount:    strconv.Itoa(workerCount),
		WorkerNameList: workerNameList,
	}
	postBody, _ := json.Marshal(bodyRequest)
	responseBody := bytes.NewBuffer(postBody)
	var bearer = "Bearer " + accessToken
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, responseBody)
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err.Error(), err
	}
	defer resp.Body.Close()
	log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
		return err.Error(), err
	}
	log.Println(string([]byte(body)))
	return string(body), nil
	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// fmt.Println("response Body:", string(body))
}

type ScaleDownBodyRequest struct {
	ClusterID      string   `json:"cluster_id"`
	WorkerCount    string   `json:"worker_count"`
	ScaleType      string   `json:"scale_type"`
	WorkerNameList []string `json:"worker_name_list"`
}

// PerformScaleDown performs scale down
func PerformScaleDown(domainAPI, vpcID, token, idCluster, clusterIDPortal string, workerNodeNameList []string) {
	workerNameList := make([]string, 0)
	workerNameList = append(workerNameList, workerNodeNameList...)
	bodyRequest := ScaleDownBodyRequest{
		ClusterID:      clusterIDPortal,
		WorkerCount:    strconv.Itoa(len(workerNodeNameList)),
		ScaleType:      "down",
		WorkerNameList: workerNameList,
	}
	url := domainAPI + "/api/v1/xplat/fke/vpc/" + vpcID + "/cluster/" + idCluster + "/scale-cluster"
	postBody, _ := json.Marshal(bodyRequest)
	responseBody := bytes.NewBuffer(postBody)
	var bearer = "Bearer " + token
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, responseBody)
	req.Header.Add("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	// log.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	log.Println(string([]byte(body)))
	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	//fmt.Println("response Body:", string(body))
}

// GetDomainApiConformEnv gets url conform environment
func GetDomainApiConformEnv(callbackURL string) string {
	s := strings.Split(callbackURL, "/api")
	return s[0]
}

func Logger() logr.Logger {
	logger := zap.NewExample()
	defer logger.Sync() // Flushes buffer, if any
	// Wrap the Zap logger with logr
	log := zapr.NewLogger(logger)
	return log
}
