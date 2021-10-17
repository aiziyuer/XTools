package main

import (
	"app/internal/util"

	"github.com/gofrs/uuid"

	"github.com/gogf/gf/util/gconv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// # vf map
// # |mdev device id                       |name             | mdev_supported_type | gpu profile id |
// # |b04965e6-a9bb-591f-8f8a-1adcb2c8dc39 |GRID A100-3-20C  | nvidia-476          | 9              |
type MdeDeviceInfo struct {
	gpu_bus_id     string
	mdev_id        string
	mdev_name      string
	mdev_type      string
	gpu_profile_id int
}

var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {

		configObject, err := util.YamlFileToObject("configs/NVIDIAMate.yaml")
		if err != nil {
			return err
		}

		zap.S().Debugf("configObject: %s.", gconv.String(configObject))

		mdevDevices := make([]*MdeDeviceInfo, 0)
		gpus := util.MustJsonPathQueryInObject(configObject, "$.mig_vgpus")
		for gpu_index, gpu := range gpus.([]interface{}) {

			zap.S().Debugf("[gpu-%3d] mig_vgpu: %s", gpu_index, gconv.String(gpu))

			bus_id := util.MustJsonPathQueryInObject(gpu, "$.bus_id")
			zap.S().Debugf("[gpu-%3d] bus_id: %s", gpu_index, gconv.String(bus_id))

			vgpus := util.MustJsonPathQueryInObject(gpu, "$.partition")
			zap.S().Debugf("[gpu-%3d] partition: %s", gpu_index, gconv.String(vgpus))

			for vgpu_index, vgpu_name := range vgpus.([]interface{}) {

				mdevDevice := &MdeDeviceInfo{
					gpu_bus_id: gconv.String(bus_id),
					mdev_id:    uuid.NewV5(uuid.NamespaceDNS, gconv.String(vgpu_index+1)).String(),
					mdev_name:  gconv.String(vgpu_name),
				}

				mdevDevices = append(mdevDevices, mdevDevice)

			}

		}

		zap.S().Infof("mdevDevices: %s", gconv.String(mdevDevices))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
