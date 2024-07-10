package vcd

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"gopkg.in/yaml.v2"
)

func GetVDC(kubeclient kubernetes.Interface) string {
	var VDC string
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("fptcloud-csi-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fptcloud csi configmap")
		klog.Fatalf("Failed to get information of fptcloud csi configmap: %v", err)
	}
	_, _, vdc := getHostOrgVDCFromConfigMap(configmaps.Data["vcloud-csi-config.yaml"])
	VDC = vdc
	return VDC
}

func GetORG(kubeclient kubernetes.Interface) string {
	var ORG string
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("fptcloud-csi-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fptcloud csi configmap")
		klog.Fatalf("Failed to get information of fptcloud csi configmap: %v", err)
	}

	//fmt.Printf("xnxx %v", reflect.TypeOf(configmaps.Data["vcloud-csi-config.yaml"]))
	_, org, _ := getHostOrgVDCFromConfigMap(configmaps.Data["vcloud-csi-config.yaml"])
	ORG = org
	return ORG
}

func GetHost(kubeclient kubernetes.Interface) string {
	var hosts string
	configmaps, err := kubeclient.CoreV1().ConfigMaps("kube-system").Get("fptcloud-csi-configmap", metav1.GetOptions{})
	if err != nil {
		fmt.Println("cannot get information from fptcloud csi configmap")
		klog.Fatalf("Failed to get information of fptcloud csi configmap: %v", err)
	}
	host, _, _ := getHostOrgVDCFromConfigMap(configmaps.Data["vcloud-csi-config.yaml"])
	hosts = host
	return hosts
}

type VCDConfig struct {
	VCD struct {
		Host     string `yaml:"host"`
		Org      string `yaml:"org"`
		VDC      string `yaml:"vdc"`
		VAppName string `yaml:"vAppName"`
	} `yaml:"vcd"`
}

//now the configmap is just a bunc of string
func getHostOrgVDCFromConfigMap(yamlContent string) (string, string, string) {
	// Unmarshal the YAML into the VCDConfig struct.
	var config VCDConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	// // Now you can access the values directly from the struct.
	// fmt.Printf("Host: %s\n", config.VCD.Host)
	// fmt.Printf("Org: %s\n", config.VCD.Org)
	// fmt.Printf("VDC: %s\n", config.VCD.VDC)
	// fmt.Printf("VAppName: %s\n", config.VCD.VAppName)

	return config.VCD.Host, config.VCD.Org, config.VCD.VDC
}

