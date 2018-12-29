package pirri

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/vic999/pirrigo/data"
	"github.com/vic999/pirrigo/logging"
	"github.com/vic999/pirrigo/settings"
	"go.uber.org/zap"
	//	"time"
)

// TODO parameterize the inputs for date ranges and add selectors on stats page
func statsActivityByStation(rw http.ResponseWriter, req *http.Request) {
	type StatsChart struct {
		ReportType int
		Labels     []int
		Series     []string
		Data       [][]int
	}
	type RawResult struct {
		StationID int
		Secs      int
	}

	result := StatsChart{
		ReportType: 1,
		Labels:     []int{},
		Series:     []string{"Unscheduled", "Scheduled"},
	}

	var rawResult0 []RawResult
	var rawResult1 []RawResult

	seriesTracker := map[int]int{}
	tracker0 := 0
	tracker1 := 0

	// unscheduled
	sqlQuery0 := `SELECT DISTINCT station_id, SUM(duration) as secs
	            FROM station_histories
	            WHERE start_time >= (CURRENT_DATE - INTERVAL ? DAY)  AND schedule_id=0 AND station_id > 0
	            GROUP BY station_id
	            ORDER BY station_id ASC`
	// scheduled
	sqlQuery1 := `SELECT DISTINCT station_id, SUM(duration) as secs
	            FROM station_histories
	            WHERE start_time >= (CURRENT_DATE - INTERVAL ? DAY)  AND schedule_id>=1 AND station_id > 0
	            GROUP BY station_id
	            ORDER BY station_id ASC`

	data.Service().DB.Raw(sqlQuery0, 7).Scan(&rawResult0)
	data.Service().DB.Raw(sqlQuery1, 7).Scan(&rawResult1)
	result.Data = [][]int{[]int{}, []int{}}

	for _, i := range rawResult0 {
		result.Data[0] = append(result.Data[0], 0)
		result.Data[1] = append(result.Data[1], 0)

		if loc, ok := seriesTracker[i.StationID]; ok {
			result.Data[0][loc] += i.Secs / 60
		} else {
			seriesTracker[i.StationID] = tracker0
			result.Labels = append(result.Labels, i.StationID)
			result.Data[0][tracker0] += i.Secs / 60
			tracker0++
		}
	}
	for _, i := range rawResult1 {
		if loc, ok := seriesTracker[i.StationID]; ok {
			result.Data[1][loc] += i.Secs / 60
		} else {
			seriesTracker[i.StationID] = tracker1
			result.Labels = append(result.Labels, i.StationID)
			result.Data[1][tracker1] += i.Secs / 60
			tracker1++
		}
	}

	blob, err := json.Marshal(&result)
	if err != nil {
		logging.Service().LogError("Error while marshalling usage stats.",
			zap.String("error", err.Error()))
	}
	io.WriteString(rw, string(blob))
}

func statsActivityByDayOfWeek(rw http.ResponseWriter, req *http.Request) {
	type StatsChart struct {
		ReportType int
		Labels     []string
		Series     []string
		Data       [][]int
	}

	result := StatsChart{
		ReportType: 2,
		Labels:     []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		Series:     []string{"Total", "Scheduled", "Unscheduled"},
	}

	type RawResult struct {
		Day  int
		Secs int
	}

	var rawResults0 []RawResult
	var rawResults1 []RawResult
	var rawResults2 []RawResult

	sqlQuery0 := fmt.Sprintf(`SELECT DISTINCT DAYOFWEEK((start_time + INTERVAL ? HOUR)) as day, SUM(duration) as secs
            FROM station_histories
            WHERE start_time >= (CURRENT_DATE - INTERVAL ? DAY)
            GROUP BY day
            ORDER BY day ASC`)
	sqlQuery1 := fmt.Sprintf(`SELECT DISTINCT DAYOFWEEK((start_time + INTERVAL ? HOUR)) as day, SUM(duration) as secs
            FROM station_histories
            WHERE start_time >= (CURRENT_DATE - INTERVAL ? DAY) AND schedule_id > 0
            GROUP BY day
            ORDER BY day ASC`)
	sqlQuery2 := fmt.Sprintf(`SELECT DISTINCT DAYOFWEEK((start_time + INTERVAL ? HOUR)) as day, SUM(duration) as secs
            FROM station_histories
            WHERE start_time >= (CURRENT_DATE - INTERVAL ? DAY) AND schedule_id = 0
            GROUP BY day
            ORDER BY day ASC`)

	data.Service().DB.Raw(sqlQuery0, settings.Service().Pirri.UtcOffset, 7).Scan(&rawResults0)
	data.Service().DB.Raw(sqlQuery1, settings.Service().Pirri.UtcOffset, 7).Scan(&rawResults1)
	data.Service().DB.Raw(sqlQuery2, settings.Service().Pirri.UtcOffset, 7).Scan(&rawResults2)

	result.Data = [][]int{
		[]int{0, 0, 0, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0, 0},
		[]int{0, 0, 0, 0, 0, 0, 0},
	}

	for _, v := range rawResults0 {
		result.Data[0][v.Day-1] = v.Secs / 60
	}
	for _, v := range rawResults1 {
		result.Data[1][v.Day-1] = v.Secs / 60
	}
	for _, v := range rawResults2 {
		result.Data[2][v.Day-1] = v.Secs / 60
	}

	blob, err := json.Marshal(&result)
	if err != nil {
		logging.Service().LogError("Error while marshalling usage stats.", zap.String("error", err.Error()))
	}
	io.WriteString(rw, string(blob))
}

func statsActivityPerStationByDOW(rw http.ResponseWriter, req *http.Request) {
	type StatsChart struct {
		ReportType int
		Labels     []string
		Series     []string
		Data       [][]float32
	}

	type RawResult struct {
		Day  int
		Mins float32
	}

	result := StatsChart{
		ReportType: 2,
		Labels:     []string{},
		Series:     []string{},
	}

	blob, err := json.Marshal(&result)
	if err != nil {
		logging.Service().LogError("Error while marshalling usage stats.", zap.String("error", err.Error()))
	}
	io.WriteString(rw, string(blob))
}

func statsStationActivity(rw http.ResponseWriter, req *http.Request) {
	type StatsChart struct {
		ReportType int
		Labels     []string
		Series     []int
		Data       [][]int
	}

	type ChartData struct {
		ID      int
		Hour    int
		RunSecs int
	}

	var chartData []ChartData
	result := StatsChart{
		ReportType: 4,
		Labels: []string{"00:00", "01:00", "02:00", "03:00",
			"04:00", "05:00", "06:00", "07:00", "08:00",
			"09:00", "10:00", "11:00", "12:00", "13:00",
			"14:00", "15:00", "16:00", "17:00", "18:00",
			"19:00", "20:00", "21:00", "22:00", "23:00"},
		Series: []int{},
	}
	result.Data = [][]int{}

	sqlStr := fmt.Sprintf(`SELECT stations.id, 
					  HOUR(start_time + INTERVAL %s HOUR) as hour, 
					  (duration) as run_secs
				FROM station_histories
				JOIN stations ON stations.id = station_histories.station_id
				WHERE start_time >= (CURRENT_DATE - INTERVAL 7 DAY) 
					AND stations.id > 0
				ORDER BY station_id ASC`, strconv.Itoa(settings.Service().Pirri.UtcOffset))

	seriesTracker := map[int]int{}

	data.Service().DB.Raw(sqlStr).Scan(&chartData)

	for n, i := range chartData {
		if n == 0 || i.ID != result.Series[len(result.Series)-1] {
			seriesTracker[i.ID] = len(result.Series)
			result.Data = append(result.Data, []int{
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0})
			result.Series = append(result.Series, i.ID)
		}
		result.Data[seriesTracker[i.ID]][i.Hour] += i.RunSecs / 60
	}

	blob, err := json.Marshal(&result)
	if err != nil {
		logging.Service().LogError("Error while marshalling usage stats.", zap.String("error", err.Error()))
	}

	io.WriteString(rw, string(blob))
}
