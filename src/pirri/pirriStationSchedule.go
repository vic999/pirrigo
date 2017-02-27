package main

import (
	//	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

var lastTriggeredItem string

//StationSchedule describes a scheduled activation for a Station
type StationSchedule struct {
	ID        int       `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	StartDate time.Time `sql:"DEFAULT:current_timestamp" gorm:"not null"`
	EndDate   time.Time `sql:"DEFAULT:'2025-01-01 00:00:00'" gorm:"not null"`
	Sunday    bool      `sql:"DEFAULT:false" gorm:"not null"`
	Monday    bool      `sql:"DEFAULT:false" gorm:"not null"`
	Tuesday   bool      `sql:"DEFAULT:false" gorm:"not null"`
	Wednesday bool      `sql:"DEFAULT:false" gorm:"not null"`
	Thursday  bool      `sql:"DEFAULT:false" gorm:"not null"`
	Friday    bool      `sql:"DEFAULT:false" gorm:"not null"`
	Saturday  bool      `sql:"DEFAULT:false" gorm:"not null"`
	StationID int       `gorm:"not null"`
	StartTime int       `gorm:"not null"`
	Duration  int       `sql:"DEFAULT:0" gorm:"not null"`
	Repeating bool      `sql:"DEFAULT:false" gorm:"not null"`
}

func createNewStationSchedule() {
	nowTime := time.Now()
	startTime, ERR := strconv.Atoi(fmt.Sprintf("%02d%02d", nowTime.Hour(), nowTime.Minute()))
	failOnError(ERR, "Unable to create new station schedule entry.")
	sched := &StationSchedule{
		StationID: 23,
		StartTime: startTime,
		Duration:  300,
	}
	defer db.Close()

	gormDbConnect()
	db.Create(&sched)
}

func checkForTasks() {
	defer db.Close()
	scheds := []StationSchedule{}
	nowTime := time.Now()
	sqlFilter := fmt.Sprintf("(start_date <= NOW() AND end_date > NOW()) AND %s=true AND start_time=%s", nowTime.Weekday(), fmt.Sprintf("%02d%02d", nowTime.Hour(), nowTime.Minute()))

	gormDbConnect()
	db.Where(sqlFilter).Find(&scheds)
	sendFoundScheduleItems(scheds)
	if ERR != nil {
		panic(ERR.Error())
	}
}

func startTaskMonitor() {
	fmt.Println("Starting monitoring at interval:", SETTINGS.MonitorInterval, "seconds.")
	for !KILL {
		checkForTasks()
		time.Sleep(time.Duration(SETTINGS.MonitorInterval) * time.Second)
	}
	defer WG.Done()
}

func sendFoundScheduleItems(items []StationSchedule) {
	defer db.Close()
	for i := range items {
		task := Task{StationSchedule: items[i]}
		gormDbConnect()
		db.Where(Station{ID: task.StationSchedule.StationID}).Find(&task.Station)
		task.send()
	}
}