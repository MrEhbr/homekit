package accessories

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type VacuumCleanerAccessory struct {
	*accessory.Accessory

	Fan     *Fan
	Battery *service.BatteryService
	Pause   *service.Switch
	Dock    *service.OccupancySensor
}

func NewCleaner(info accessory.Info) *VacuumCleanerAccessory {
	cleaner := VacuumCleanerAccessory{}
	cleaner.Accessory = accessory.New(info, accessory.TypeOther)

	cleaner.Fan = NewFan()
	cleaner.AddService(cleaner.Fan.Service)

	cleaner.Battery = service.NewBatteryService()
	cleaner.AddService(cleaner.Battery.Service)

	cleaner.Pause = service.NewSwitch()
	cleaner.AddService(cleaner.Pause.Service)

	cleaner.Dock = service.NewOccupancySensor()
	cleaner.AddService(cleaner.Dock.Service)

	return &cleaner
}
