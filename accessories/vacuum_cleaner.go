package accessory

import (
	acc "github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type VacuumCleanerAccessory struct {
	*acc.Accessory

	Fan     *Fan
	Battery *service.BatteryService
	Pause   *service.Switch
	Dock    *service.OccupancySensor
}

func NewCleaner(info acc.Info) *VacuumCleanerAccessory {
	cleaner := VacuumCleanerAccessory{}
	cleaner.Accessory = acc.New(info, acc.TypeOther)

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
