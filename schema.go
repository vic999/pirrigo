package main

import (
	"github.com/vic999/pirrigo/data"
	"github.com/vic999/pirrigo/pirri"
	"github.com/vic999/pirrigo/settings"
)

func migrateDataSchema() {
	d := data.Service()
	d.DB.AutoMigrate(
		&pirri.Station{},
		&pirri.DripNode{},
		&pirri.GpioPin{},
		&pirri.StationHistory{},
		&pirri.StationSchedule{},
		&settings.Settings{},
	)
}

//TODO: remove this later - it's for testing only.
func firstRunDBSetup() {
	d := data.Service()
	gpios := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28}
	for pin := range gpios {
		d.DB.Create(&pirri.GpioPin{
			GPIO:   pin,
			Notes:  "",
			Common: false,
		})
	}
}
