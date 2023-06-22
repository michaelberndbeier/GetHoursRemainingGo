package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func compareDays(i int, j int) bool {
	return i < j
}

func getDaysRemaining() int {

	layout := "02.01.2006"
	vacDayListPath := "vacationDays.csv"

	file, err := os.Open(vacDayListPath)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var vacationDays []time.Time

	for scanner.Scan() {
		line := scanner.Text()
		vacationDay, error := time.Parse(layout, line)

		if error != nil {
			fmt.Println(error)
			return 0
		}
		vacationDays = append(vacationDays, vacationDay)
	}

	sort.SliceStable(vacationDays, func(i, j int) bool {
		return vacationDays[i].Unix() < vacationDays[j].Unix()
	})

	now := time.Now()
	todayString := now.Format(layout)
	today, error := time.Parse(layout, todayString)
	if error != nil {
		fmt.Println(error)
		return 0
	}

	tomorrow := today.AddDate(0, 0, 1)
	maxVacDay := vacationDays[len(vacationDays)-1]

	fmt.Println(today)
	fmt.Println(maxVacDay)

	var daysUntil []time.Time

	nextDay := tomorrow
	for nextDay.Unix() <= maxVacDay.Unix() {
		switch nextDay.Weekday() {
		case time.Saturday:
		case time.Sunday:
			break
		default:
			if contains(vacationDays, nextDay) == false {
				daysUntil = append(daysUntil, nextDay)
			}
		}

		nextDay = nextDay.AddDate(0, 0, 1)
	}

	return len(daysUntil)
}

func web(w http.ResponseWriter, req *http.Request) {
	daysRemaining := getDaysRemaining()
	fmt.Fprintf(w, "%v", daysRemaining)
}

func lotr(w http.ResponseWriter, req *http.Request) {
	daysRemaining := getDaysRemaining()
	lotrExtendedAllTitlesLength := 715
	lotr := float32((daysRemaining * 8 * 60)) / float32(lotrExtendedAllTitlesLength)
	fmt.Fprintf(w, "Only %v LOTR watch-throughs!! (Extended obviously)", lotr)
}

func main() {
	http.HandleFunc("/web", web)
	http.HandleFunc("/lotr", lotr)

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
