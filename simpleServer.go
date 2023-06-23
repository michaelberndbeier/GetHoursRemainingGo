package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type WorkDay struct {
	Date    string
	Weekday string
}

type WorkDays struct {
	Items []WorkDay
}

type MarqueeText struct {
	What         string
	HowManyTimes float32
}

type DataForIndex struct {
	Days                     WorkDays
	Hours                    int
	MarqueeTexts             []MarqueeText
	HoursUntilNextEmployment int
	NumOfWorkDays            int
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getLayout() string {
	return "02.01.2006"
}

var vacationDaysCached = getVacationDaysFromFile()

func getVacationDaysFromFile() []time.Time {
	vacDayListPath := "vacationDays.csv"

	file, err := os.Open(vacDayListPath)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var vacationDays []time.Time

	for scanner.Scan() {
		line := scanner.Text()
		vacationDay, _ := time.Parse(getLayout(), line)

		vacationDays = append(vacationDays, vacationDay)
	}

	sort.SliceStable(vacationDays, func(i, j int) bool {
		return vacationDays[i].Unix() < vacationDays[j].Unix()
	})

	return vacationDays

}

func getVacationDays() []time.Time {

	now := time.Now()
	todayString := now.Format(getLayout())
	today, _ := time.Parse(getLayout(), todayString)

	if vacationDaysCached[0].Unix() != today.Unix() {
		vacationDaysCached = getVacationDaysFromFile()
	}

	return vacationDaysCached
}

var daysRemainingCached []time.Time

func getDaysRemaining() []time.Time {
	var daysRemaining []time.Time
	now := time.Now()
	todayString := now.Format(getLayout())
	today, _ := time.Parse(getLayout(), todayString)

	if len(daysRemaining) > 0 {
		if daysRemaining[0].Unix() == today.Unix() {
			return daysRemainingCached
		}
	}

	vacDays := getVacationDays()
	maxVacDay := vacDays[len(vacDays)-1]

	nextDay := today
	for nextDay.Unix() <= maxVacDay.Unix() {
		daysRemaining = append(daysRemaining, nextDay)

		nextDay = nextDay.AddDate(0, 0, 1)
	}

	daysRemainingCached = daysRemaining

	return daysRemainingCached
}

var workDaysRemainingCached []time.Time

func getWorkDaysRemaining() []time.Time {
	var daysRemaining []time.Time

	now := time.Now()
	todayString := now.Format(getLayout())
	today, _ := time.Parse(getLayout(), todayString)

	if len(workDaysRemainingCached) > 0 {
		if workDaysRemainingCached[0].Unix() == today.Unix() {
			return workDaysRemainingCached
		}
	}

	vacationDays := getVacationDays()

	maxVacDay := vacationDays[len(vacationDays)-1]

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

	workDaysRemainingCached = daysRemaining

	return daysRemaining
}

func contains(vacationDays []time.Time, day time.Time) bool {
	for _, e := range vacationDays {
		if e.Unix() == day.Unix() {
			return true
		}
	}

	return false
}

func getNumberOfRemainingDays() int {
	return len(getWorkDaysRemaining())
}

func numDays(w http.ResponseWriter, req *http.Request) {
	daysRemaining := getNumberOfRemainingDays()
	fmt.Fprintf(w, "%v", daysRemaining)
}

func getMinutesBasedText(what string, minutesToDoItOnce float32) MarqueeText {
	var retVal = MarqueeText{}

	retVal.HowManyTimes = float32((getHoursRemaining(getWorkDaysRemaining()) * 60)) / float32(minutesToDoItOnce)
	retVal.What = what

	return retVal
}

func days(w http.ResponseWriter, req *http.Request) {
	var daysRemaining = getWorkDaysRemaining()
	var daysRemainingJson, _ = json.Marshal(daysRemaining)

	fmt.Fprintf(w, "%s", daysRemainingJson)
}

func getWorkDays() WorkDays {
	var daysRemaining = getWorkDaysRemaining()

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

var activitiesCached = readActivitiesCSV()

func getRatioTexts() []MarqueeText {

	retVals := []MarqueeText{}

	for _, e := range activitiesCached {
		retVals = append(retVals, getMinutesBasedText(e.what, e.duration))
	}

	return retVals
}

func changeSite(w http.ResponseWriter, req *http.Request) {
	data := DataForIndex{}
	data.Days = getWorkDays()
	data.Hours = getHoursRemaining(getWorkDaysRemaining())
	data.HoursUntilNextEmployment = getHoursUntilNextEmployment()
	data.MarqueeTexts = getRatioTexts()
	data.NumOfWorkDays = len(data.Days.Items)

	tmpl, _ := template.ParseFiles("./index.html")
	tmpl.Execute(w, data)
}

func getHoursUntilNextEmployment() int {
	var daysLeft = getDaysRemaining()
	return getHoursRemaining(daysLeft)
}

func getHoursRemaining(daysLeft []time.Time) int {
	var layout = getLayout()
	var now = time.Now()
	var today, _ = time.Parse(layout, time.Now().Format(layout))

	var hoursAlreadyWorkedToday = 0
	if today.Unix() == daysLeft[0].Unix() {
		var thisHour = now.Hour()

		var calcedHoursSinceStartOfWorkday = thisHour - 6
		if calcedHoursSinceStartOfWorkday > 0 && calcedHoursSinceStartOfWorkday < 11 {
			hoursAlreadyWorkedToday = calcedHoursSinceStartOfWorkday
		}
	}

	var numDaysRemaining = len(daysLeft)
	var hoursForDaysRemaining = numDaysRemaining * 8.0
	var calcedHoursRemaining = hoursForDaysRemaining - hoursAlreadyWorkedToday

	return calcedHoursRemaining
}
func hours(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%v Hours left", getHoursRemaining(getWorkDaysRemaining()))
}

func main() {

	http.HandleFunc("/change", changeSite)

	http.ListenAndServe(":8090", nil)
}

type ActivityItem struct {
	what     string
	duration float32
}

func readActivitiesCSV() []ActivityItem {

	var activitiesCSV = "./activities.csv"
	file, err := os.Open(activitiesCSV)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var activityItems = []ActivityItem{}

	for scanner.Scan() {
		line := scanner.Text()
		var splitted = strings.Split(line, ";")

		if len(splitted) == 2 {
			var what = splitted[0]
			var durationText = splitted[1]
			var duration, err = strconv.ParseFloat(durationText, 32)
			if err != nil {
				fmt.Println(err)
			}

			activityItem := ActivityItem{
				what:     what,
				duration: float32(duration),
			}
			activityItems = append(activityItems, activityItem)

		}

	}

	return activityItems

}
