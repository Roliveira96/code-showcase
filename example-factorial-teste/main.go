package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Period representa um período de agendamento.
type Period struct {
	Id           int    `json:"id"`
	Default      bool   `json:"default"`
	ScheduleId   int    `json:"schedule_id"`
	StartDate    string `json:"start_date"`
	StartDay     int    `json:"start_day"`
	EndDate      string `json:"end_date"`
	EndDay       int    `json:"end_day"`
	StartMonth   int    `json:"start_month"`
	EndMonth     int    `json:"end_month"`
	ScheduleType string `json:"schedule_type"`
}

// Schedule representa um agendamento.
type Schedule struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	CompanyId   int       `json:"company_id"`
	ArchivedAt  *string   `json:"archived_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Periods     []Period  `json:"periods"`
	Employees   []int     `json:"employees"`
	EmployeeIds []int     `json:"employee_ids"`
}

// Meta representa os metadados da resposta da API.
type Meta struct {
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	Total           int    `json:"total"`
	Limit           int    `json:"limit"`
}

type ScheduleResponse struct {
	Data []Schedule `json:"data"`
	Meta Meta       `json:"meta"`
}

type WorkScheduleResponse struct {
	Data []WorkSchedule `json:"data"`
	Meta Meta           `json:"meta"`
}

type WorkSchedule struct {
	ID                int    `json:"id"`
	Weekday           string `json:"weekday"`
	StartAt           string `json:"start_at"`
	EndAt             string `json:"end_at"`
	FormattedEndAt    string `json:"formatted_end_at"`
	FormattedStartAt  string `json:"formatted_start_at"`
	DurationInSeconds int    `json:"duration_in_seconds"`
	OverlapPeriodID   int    `json:"overlap_period_id"`
}

type TimeEntry struct {
	ID                               int       `json:"id"`
	EmployeeID                       int       `json:"employee_id"`
	Date                             string    `json:"date"`
	ReferenceDate                    string    `json:"reference_date"`
	ClockIn                          string    `json:"clock_in"`
	ClockOut                         *string   `json:"clock_out"`
	InSource                         string    `json:"in_source"`
	OutSource                        *string   `json:"out_source"`
	Observations                     *string   `json:"observations"`
	LocationType                     *string   `json:"location_type"`
	HalfDay                          *bool     `json:"half_day"`
	InLocationLatitude               *float64  `json:"in_location_latitude"`
	InLocationLongitude              *float64  `json:"in_location_longitude"`
	InLocationAccuracy               *float64  `json:"in_location_accuracy"`
	OutLocationLatitude              *float64  `json:"out_location_latitude"`
	OutLocationLongitude             *float64  `json:"out_location_longitude"`
	OutLocationAccuracy              *float64  `json:"out_location_accuracy"`
	Workable                         bool      `json:"workable"`
	CreatedAt                        time.Time `json:"created_at"`
	WorkplaceID                      int       `json:"workplace_id"`
	TimeSettingsBreakConfigurationID *int      `json:"time_settings_break_configuration_id"`
	CompanyID                        int       `json:"company_id"`
	UpdatedAt                        time.Time `json:"updated_at"`
	Minutes                          int       `json:"minutes"`
	ClockInWithSeconds               string    `json:"clock_in_with_seconds"`
}

type TimeEntryResponse struct {
	Data []TimeEntry `json:"data"`
	Meta Meta        `json:"meta"`
}

const (
	uriBase   = "https://api.factorialhr.com/api/2025-04-01/"
	apiKey    = "94e54fd2859b99b993f9c839ea7574307115a0376ed3432c58ee7483c54af118"
	idUsuario = 2288637
)

func main() {
	s := getAllSchedules()
	s = getSchedulesUsuarioPertence(idUsuario, s)

	for _, schedule := range s {
		fmt.Println(schedule.Name)
		wcfg := getDayConfigurationSchedule(schedule.Id)

		for _, workSchedule := range wcfg {
			fmt.Println(workSchedule.Weekday)
			fmt.Printf("%s - %s\n", workSchedule.FormattedStartAt, workSchedule.FormattedEndAt)
		}
	}

	times := getShifOfUser(idUsuario, "2025-04-11", "2025-04-11")
	fmt.Println("Dados do Fulano sem nome")
	mintotal := 0

	for _, entry := range times {
		exp := "Folga"

		if entry.Workable {
			exp = "Trabalhando"
		}
		if entry.ClockOut != nil {
			fmt.Printf("%s Iniciou: %s -  Parou: %s. Total: %s \n", exp, entry.ClockIn, *entry.ClockOut, formatDuration(entry.Minutes))
		} else {
			fmt.Printf("%s Iniciou: %s.  Total: %s  \n", exp, entry.ClockIn, formatDuration(entry.Minutes))
		}

		mintotal += entry.Minutes
	}
	fmt.Printf("Total de horas trabalhada: %s\n", formatDuration(mintotal))
}

func getDayConfigurationSchedule(idSchedule int) []WorkSchedule {
	var workSchedule []WorkSchedule

	url := uriBase + "/resources/work_schedule/day_configurations?schedule_id=" + strconv.Itoa(idSchedule)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-api-key", apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var result WorkScheduleResponse
	err := json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return workSchedule
	}

	return result.Data
}

func getSchedulesUsuarioPertence(idusuario int, schedules []Schedule) []Schedule {
	var newSchedules []Schedule

	for _, schedule := range schedules {
		if contains(schedule.Employees, idusuario) {
			newSchedules = append(newSchedules, schedule)
		}
	}

	return newSchedules
}

func getAllSchedules() []Schedule {

	schedules := []Schedule{}

	url := uriBase + "resources/work_schedule/schedules?with_employee_ids=true&with_periods=true"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-api-key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return schedules
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return schedules
	}

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Erro: Status code %d, resposta: %s\n", res.StatusCode, string(body))
		return schedules
	}

	var result ScheduleResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return schedules
	}

	return result.Data
}

func contains(s []int, id int) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func getShifOfUser(id int, start, end string) []TimeEntry {
	var times []TimeEntry

	url := fmt.Sprintf("/resources/attendance/shifts?employee_ids[]=%d&start_on=%s&end_on=%s&sort_created_at_asc=true", id, start, end)

	url = uriBase + url

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-api-key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição:", err)
		return times
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return times
	}

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Erro: Status code %d, resposta: %s\n", res.StatusCode, string(body))
		return times
	}

	defer res.Body.Close()

	var result TimeEntryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Erro ao decodificar o JSON:", err)
		return times
	}

	return result.Data
}

func formatDuration(minutes int) string {
	hours := minutes / 60
	remainingMinutes := minutes % 60

	hoursStr := strconv.Itoa(hours)
	minutesStr := strconv.Itoa(remainingMinutes)

	if remainingMinutes < 10 {
		minutesStr = "0" + minutesStr
	}

	return hoursStr + ":" + minutesStr
}
