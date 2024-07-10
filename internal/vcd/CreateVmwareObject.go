package vcd

import (
    "fmt"
    "github.com/vmware/go-vcloud-director/v2/govcd"
	"errors"
	
)


func GetVMFromName(client *govcd.VCDClient, orgName string, vdcName string,  vappName string, vmName string) (*govcd.VM, error) {
	org, err := client.GetOrgByName(orgName)
	if err != nil {
		fmt.Printf("organization %s not found : %s\n", orgName, err)
		return nil, errors.New("error get org")
	}

	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		fmt.Printf("VDC %s not found : %s\n", vdcName, err)
		return nil, errors.New("error get vdc")
	}

	
	vApp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		fmt.Printf("VApp %s not found : %s\n", vappName, err)
		return nil, errors.New("error get vapp")
	}

	vm, err := vApp.GetVMByName(vmName, false)
	if err != nil {
		fmt.Printf("VMName %s not found : %s\n", vmName, err)
		return nil, errors.New("error get vm")
	}

	return vm, nil
}


