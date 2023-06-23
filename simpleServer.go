package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"
)

type WorkDay struct {
	Date    string
	Weekday string
}

type WorkDays struct {
	Items []WorkDay
}

type DataForIndex struct {
	Days  WorkDays
	LOTR  string
	Hours int
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func compareDays(i int, j int) bool {
	return i < j
}

func getLayout() string {
	return "02.01.2006"
}
func getVacationDays() []time.Time {
	vacDayListPath := "vacationDays.csv"

	file, err := os.Open(vacDayListPath)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var vacationDays []time.Time

	for scanner.Scan() {
		line := scanner.Text()
		vacationDay, error := time.Parse(getLayout(), line)

		if error != nil {
			fmt.Println(error)
			return vacationDays
		}
		vacationDays = append(vacationDays, vacationDay)
	}

	sort.SliceStable(vacationDays, func(i, j int) bool {
		return vacationDays[i].Unix() < vacationDays[j].Unix()
	})

	return vacationDays
}

func getDaysRemaning() []time.Time {
	var daysRemaining []time.Time

	now := time.Now()
	todayString := now.Format(getLayout())
	today, error := time.Parse(getLayout(), todayString)
	if error != nil {
		fmt.Println(error)
		return daysRemaining
	}
	vacationDays := getVacationDays()

	// tomorrow := today.AddDate(0, 0, 1)
	maxVacDay := vacationDays[len(vacationDays)-1]

	fmt.Println(today)
	fmt.Println(maxVacDay)

	nextDay := today
	for nextDay.Unix() <= maxVacDay.Unix() {
		switch nextDay.Weekday() {
		case time.Saturday:
		case time.Sunday:
			break
		default:
			if contains(vacationDays, nextDay) == false {
				daysRemaining = append(daysRemaining, nextDay)
			}
		}

		nextDay = nextDay.AddDate(0, 0, 1)
	}
	return daysRemaining
}

func getNumberOfRemainingDays() int {
	return len(getDaysRemaning())
}

func numDays(w http.ResponseWriter, req *http.Request) {
	daysRemaining := getNumberOfRemainingDays()
	fmt.Fprintf(w, "%v", daysRemaining)
}

func getLOTR() string {

	lotrExtendedAllTitlesLength := 715
	lotr := float32((getHoursRemaining() * 60)) / float32(lotrExtendedAllTitlesLength)
	text := fmt.Sprintf("Only %v LOTR watch-throughs!! (Extended obviously)", lotr)

	return text
}

func lotr(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, getLOTR())
}

func days(w http.ResponseWriter, req *http.Request) {
	var daysRemaining = getDaysRemaning()
	var daysRemainingJson, _ = json.Marshal(daysRemaining)

	fmt.Fprintf(w, "%s", daysRemainingJson)
}

func getWorkDays() WorkDays {
	var daysRemaining = getDaysRemaning()

	workDays := WorkDays{}
	layout := getLayout()

	for i := range daysRemaining {
		day := daysRemaining[i]
		workDay := WorkDay{
			Date:    day.Format(layout),
			Weekday: day.Weekday().String(),
		}

		workDays.Items = append(workDays.Items, workDay)
	}
	return workDays
}

func htmlTableDays(w http.ResponseWriter, req *http.Request) {
	data := DataForIndex{}
	data.Days = getWorkDays()
	data.LOTR = getLOTR()
	data.Hours = getHoursRemaining()

	tmpl, _ := template.ParseFiles("./index.html")
	tmpl.Execute(w, data)
}

func getHoursRemaining() int {
	var daysRemaining = getDaysRemaning()

	var layout = getLayout()
	var now = time.Now()
	var today, _ = time.Parse(layout, time.Now().Format(layout))

	var hoursAlreadyWorkedToday = 0
	if today.Unix() == daysRemaining[0].Unix() {
		var thisHour = now.Hour()

		var calcedHoursSinceStartOfWorkday = thisHour - 6
		if calcedHoursSinceStartOfWorkday > 0 && calcedHoursSinceStartOfWorkday < 11 {
			hoursAlreadyWorkedToday = calcedHoursSinceStartOfWorkday
		}
	}

	var numDaysRemaining = len(daysRemaining)
	var hoursForDaysRemaining = numDaysRemaining * 8.0
	var calcedHoursRemaining = hoursForDaysRemaining - hoursAlreadyWorkedToday

	return calcedHoursRemaining
}
func hours(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%v Hours left", getHoursRemaining())
}

func main() {
	http.HandleFunc("/numDays", numDays)
	http.HandleFunc("/lotr", lotr)
	http.HandleFunc("/days", days)
	http.HandleFunc("/hours", hours)
	http.HandleFunc("/table", htmlTableDays)

	http.ListenAndServe(":8090", nil)
}

func contains(days []time.Time, day time.Time) bool {
	for _, a := range days {
		if a.Unix() == day.Unix() {
			return true
		}
	}
	return false
}
