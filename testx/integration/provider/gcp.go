package provider

/**
	Orphaned Resources
	- VMs:
		Describe instances with specified tag name:<cluster-name>
		Report/Print out instances found
		Describe volumes attached to the instance (using instance id)
		Report/Print out volumes found
		Delete attached volumes found
		Terminate instances found
	- Disks:
		Describe volumes with tag status:available
		Report/Print out volumes found
		Delete identified volumes
**/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	v1 "k8s.io/api/core/v1"

	providerDriver "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/apis"
	fake "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/fake"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
)

var (
	tagName   = "name"
	tagValue  = "istiodev"
	projectID = os.Getenv("projectID")
	filters   = [...]string{
		"status = RUNNING",
	}
	availableFilters = [...]string{
		"available = TRUE",
	}
)

func newSession(machineClass *v1alpha1.MachineClass, secret *v1.Secret) (*http.Client, *compute.Service) {
	var (
		providerSpec *api.GCPProviderSpec
		sPI          fake.PluginSPIImpl
	)

	err := json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Error occured while performing unmarshal %s", err.Error())
	}

	ctx, svc, err := sPI.NewComputeService(secret)
	if err != nil {
		log.Printf("Error occured while creating new session %s", err)
	}

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		fmt.Println("Could not get authenticated client: ", err)
	}

	return client, svc
}

func DescribeMachines(machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var machines []string
	var sPI fake.PluginSPIImpl
	driverprovider := providerDriver.NewGCPPlugin(&sPI)
	machineList, err := driverprovider.ListMachines(context.TODO(), &driver.ListMachinesRequest{
		MachineClass: machineClass,
		Secret:       &v1.Secret{Data: secretData},
	})
	if err != nil {
		return nil, err
	} else if len(machineList.MachineList) != 0 {
		fmt.Printf("\nAvailable Machines: ")
		for _, machine := range machineList.MachineList {
			machines = append(machines, machine)
		}
	}
	return machines, nil
}

// DescribeInstancesWithTag describes the instance with the specified tag
func DescribeInstancesWithTag(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	cli, computeService := newSession(machineClass, &v1.Secret{Data: secretData})

	var instances []string
	zoneListCall := computeService.Zones.List(projectID)
	zoneList, err := zoneListCall.Do()
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, zone := range zoneList.Items {
			instanceListCall := computeService.Instances.List(projectID, zone.Name)
			instanceListCall.Filter(strings.Join(filters[:], " "))
			instanceList, err := instanceListCall.Do()
			if err != nil {
				fmt.Println("Error", err)
			} else {
				for _, instance := range instanceList.Items {
					if instance.Labels[tagName] == tagValue {
						instances = append(instances, instance.Name)
						TerminateInstance(instance.Name, zone.Name, computeService)
					}
				}
			}
		}
	}
	return instances, nil
}

// TerminateInstance terminates the specified compute instance.
func TerminateInstance(instanceID string, zoneName string, computeService *compute.Service) error {
	result := computeService.Instances.Delete(projectID, zoneName, instanceID)

	fmt.Println(result)
	return nil
}

// DescribeAvailableVolumes describes volumes with the specified tag
func DescribeAvailableVolumes(tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	cli, computeService := newSession(machineClass, &v1.Secret{Data: secretData})

	var volumes []string
	zoneListCall := computeService.Zones.List(projectID)
	zoneList, err := zoneListCall.Do()
	if err != nil {
		fmt.Println("Error", err)
	} else {
		for _, zone := range zoneList.Items {
			volumesListCall := computeService.Disks.List(projectID, zone.Name)
			volumesListCall.Filter(strings.Join(filters[:], " "))
			volumesList, err := volumesListCall.Do()
			if err != nil {
				fmt.Println("Error", err)
			} else {
				for _, volume := range volumesList.Items {
					if volume.Labels[tagName] == tagValue {
						volumes = append(volumes, volume.Name)
						DeleteVolume(volume.Name, zone.Name, computeService)
					}
				}
			}
		}
	}
	return volumes, nil
}

// DeleteVolume deletes the specified volume
func DeleteVolume(VolumeID string, zoneName string, computeService *compute.Service) error {
	// TO-DO: deletes an available volume with the specified volume ID
	// If the command succeeds, no output is returned.
	result := computeService.Disks.Delete(projectID, zoneName, VolumeID)

	fmt.Println(result)
	return nil
}

// AdditionalResourcesCheck describes VPCs and network interfaces
func AdditionalResourcesCheck(tagName string, tagValue string) error {
	// TO-DO: Checks for Network interfaces and VPCs

	return nil
}
