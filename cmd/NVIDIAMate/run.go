package main

import (
	"app/internal/util"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gofrs/uuid"

	"github.com/gogf/gf/util/gconv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// # vf map
// # |mdev device id                       |name             | mdev_supported_type | gpu profile id |
// # |b04965e6-a9bb-591f-8f8a-1adcb2c8dc39 |GRID A100-3-20C  | nvidia-476          | 9              |
type MdeDeviceInfo struct {
	GpuBusID     string `json:"bus_id"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	GpuProfileID int    `json:"gpu_profile_id"`
}

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {

		configObject, err := util.YamlFileToObject("/root/.config/NVIDIAMate/NVIDIAMate.yaml")
		if err != nil {
			return err
		}

		zap.S().Debugf("configObject: %s.", gconv.String(configObject))

		mdevDevices := make([]*MdeDeviceInfo, 0)
		gpus := util.MustJsonPathQueryInObject(configObject, "$.gpus")
		for gpu_index, gpu := range gpus.([]interface{}) {

			zap.S().Debugf("[gpu-%03d] gpus: %s", gpu_index, gconv.String(gpu))

			bus_id := gconv.String(util.MustJsonPathQueryInObject(gpu, "$.bus_id"))
			zap.S().Debugf("[gpu-%03d] bus_id: %s", gpu_index, bus_id)

			// 开始解析pcie设备
			m := util.NamedStringSubMatch(
				regexp.MustCompile(
					// bus_id格式: domain:bus:slot.function
					// eg: 000000:00:00.0
					`(?P<domain>[0-9a-fA-F]{4,}):(?P<bus>[0-9a-fA-F]{2,}):(?P<slot>[0-9a-fA-F]{2,}).(?P<function>[0-9a-fA-F]{1,})`,
				),
				bus_id,
			)
			pci_device_id := fmt.Sprintf(
				// pcie格式: domain:bus:slot.function
				// 0000:00:00.0
				"%04x:%02x:%02x.%x",
				gconv.Int(util.IgnoreError(strconv.ParseInt(m["domain"], 16, 64))),
				gconv.Int(util.IgnoreError(strconv.ParseInt(m["bus"], 16, 64))),
				gconv.Int(util.IgnoreError(strconv.ParseInt(m["slot"], 16, 64))),
				gconv.Int(util.IgnoreError(strconv.ParseInt(m["function"], 16, 64))),
			)
			zap.S().Debugf("[gpu-%03d] pci_device_id: %s", gpu_index, pci_device_id)

			// /usr/lib/nvidia/sriov-manage -e ${pci_device_id}
			// eg. /usr/lib/nvidia/sriov-manage -e 0000:1e:00.0
			zap.S().Debugf("[gpu-%03d] COMMAND: /usr/lib/nvidia/sriov-manage -e %s", gpu_index, pci_device_id)
			if _, err := exec.Command(
				"/usr/lib/nvidia/sriov-manage",
				"-e",
				pci_device_id,
			).CombinedOutput(); err != nil {
				log.Fatalf("cmd.Run() failed with %s\n", err)
			}

			// grep -E '.+' /sys/bus/pci/devices/${pci_device_id}/virtfn0/mdev_supported_types/*/name
			// eg. grep -E '.+' /sys/bus/pci/devices/0000:1e:00.0/virtfn0/mdev_supported_types/*/name
			zap.S().Debugf("[gpu-%03d] COMMAND: grep -E '.+' /sys/bus/pci/devices/%s/virtfn0/mdev_supported_types/*/name", gpu_index, pci_device_id)

			vgpus := util.MustJsonPathQueryInObject(gpu, "$.vgpus")
			zap.S().Debugf("[gpu-%03d] vgpus: %s", gpu_index, gconv.String(vgpus))

			// nvidia-smi mig -dci; nvidia-smi mig -dgi; nvidia-smi mig -cgi 9,14,19,19 -C
			zap.S().Debugf("[gpu-%03d] COMMAND: nvidia-smi mig -dci; nvidia-smi mig -dgi; nvidia-smi mig -cgi 9,14,19,19 -C", gpu_index)
			_, _ = exec.Command("nvidia-smi", "mig", "-dci").CombinedOutput()
			_, _ = exec.Command("nvidia-smi", "mig", "-dgi").CombinedOutput()
			_, _ = exec.Command("nvidia-smi", "mig", "-cgi", "9,14,19,19", "-C").CombinedOutput()

			for vgpu_index, vgpu := range vgpus.([]interface{}) {

				mdev_id, err := util.JsonPathQueryInObject(gpu, "$.id")
				if err == nil {
					mdev_id = ""
				}

				mdevDevice := &MdeDeviceInfo{
					GpuBusID: gconv.String(bus_id),
					ID: util.GetAnyString(
						gconv.String(mdev_id),
						uuid.NewV5(uuid.NamespaceDNS, gconv.String(vgpu_index+1)).String(),
					),
					Name: gconv.String(util.MustJsonPathQueryInObject(vgpu, "$.name")),
					Type: gconv.String(util.MustJsonPathQueryInObject(vgpu, "$.type")),
				}

				// echo 'b04965e6-a9bb-591f-8f8a-1adcb2c8dc39' > /sys/bus/pci/devices/0000:1e:00.0/virtfn0/mdev_supported_types/nvidia-476/create
				cmd := fmt.Sprintf("echo '%s' > /sys/bus/pci/devices/%s/virtfn%d/mdev_supported_types/%s/create",
					mdevDevice.ID,
					pci_device_id,
					vgpu_index,
					mdevDevice.Type,
				)
				zap.S().Debugf("[gpu-%03d] COMMAND: %s", gpu_index, cmd)
				_, _ = exec.Command("bash", "-c", cmd).CombinedOutput()

				mdevDevices = append(mdevDevices, mdevDevice)
			}

			zap.S().Debugf("[gpu-%03d] COMMAND: mdevctl list", gpu_index)

		}

		zap.S().Infof("mdevDevices: %s", gconv.String(mdevDevices))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
