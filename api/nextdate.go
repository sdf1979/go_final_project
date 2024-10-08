package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	constVal "github.com/sdf1979/go_final_project/const"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	taskDate, err := time.Parse(constVal.FormatDate, date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}

	if repeat == "" {
		return "", fmt.Errorf("invalid repeat format: %s", "empty string")
	}

	rule := strings.Split(repeat, " ")
	if len(rule) == 0 {
		return "", fmt.Errorf("invalid repeat format: %s", repeat)
	}

	switch rule[0] {
	case "d":
		return nextDateDayRule(now, taskDate, rule)
	case "y":
		return nextDateYearRule(now, taskDate, rule)
	case "w":
		return nextDateWeekRule(now, taskDate, rule)
	case "m":
		return nextDateMonthRule(now, taskDate, rule)
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", rule[0])
	}
}

func nextDateDayRule(now time.Time, taskDate time.Time, rule []string) (string, error) {
	if len(rule) != 2 {
		return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
	}
	days, err := strconv.Atoi(rule[1])
	if err != nil {
		return "", fmt.Errorf("invalid days value: %w", err)
	}
	if days <= 0 || days > 400 {
		return "", fmt.Errorf("invalid days value: %d", days)
	}

	taskDate = taskDate.AddDate(0, 0, days)
	for taskDate.Before(now) || taskDate.Equal(now) {
		taskDate = taskDate.AddDate(0, 0, days)
	}

	return taskDate.Format(constVal.FormatDate), nil
}

func nextDateYearRule(now time.Time, taskDate time.Time, rule []string) (string, error) {
	if len(rule) != 1 {
		return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
	}
	taskDate = taskDate.AddDate(1, 0, 0)
	for taskDate.Before(now) || taskDate.Equal(now) {
		taskDate = taskDate.AddDate(1, 0, 0)
	}

	return taskDate.Format(constVal.FormatDate), nil
}

func nextDateWeekRule(now time.Time, taskDate time.Time, rule []string) (string, error) {
	if len(rule) != 2 {
		return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
	}
	daysTmp := strings.Split(rule[1], ",")
	days := make(map[int]bool)
	for _, dayStr := range daysTmp {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 7 {
			return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
		}
		days[day] = true
	}
	if taskDate.Before(now) {
		taskDate = now
	}
	taskDate = taskDate.AddDate(0, 0, 1)
	for {
		weekDay := int(taskDate.Weekday())
		if weekDay == 0 {
			weekDay = 7
		}
		_, ok := days[weekDay]
		if ok {
			break
		}
		taskDate = taskDate.AddDate(0, 0, 1)
	}
	return taskDate.Format(constVal.FormatDate), nil
}

func nextDateMonthRule(now time.Time, taskDate time.Time, rule []string) (string, error) {
	if len(rule) < 2 && len(rule) > 3 {
		return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
	}

	var days, lastDays []int
	daysTmp := strings.Split(rule[1], ",")
	for _, dayStr := range daysTmp {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < -2 || day > 31 || day == 0 {
			return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
		}
		if day > 0 {
			days = append(days, day)
		} else {
			lastDays = append(lastDays, day)
		}
	}

	months := make(map[int]bool)
	if len(rule) == 2 {
		months = map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true}
	} else if len(rule) == 3 {
		monthsStr := strings.Split(rule[2], ",")
		for _, monthStr := range monthsStr {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return "", fmt.Errorf("unsupported repeat rule: %s", strings.Join(rule, " "))
			}
			months[month] = true
		}
	}

	if taskDate.Before(now) {
		taskDate = now
	}
	taskDate = taskDate.AddDate(0, 0, 1)
	for {
		for {
			ok := months[int(taskDate.Month())]
			if ok {
				break
			}
			taskDate = lastDayMonth(taskDate).AddDate(0, 0, 1)
		}
		daysRule := daysMonthRule(days, lastDays, taskDate)
		ok := false
		curMonth := taskDate.Month()
		for {
			_, ok = daysRule[int(taskDate.Day())]
			if ok {
				break
			}
			taskDate = taskDate.AddDate(0, 0, 1)
			if curMonth != taskDate.Month() {
				break
			}
		}
		if ok {
			return taskDate.Format(constVal.FormatDate), nil
		}
	}
}

func daysMonthRule(days []int, lastDays []int, month time.Time) map[int]bool {
	daysResult := make(map[int]bool)
	for _, day := range days {
		daysResult[day] = true
	}

	lastDayMonth := lastDayMonth(month).Day()
	for _, day := range lastDays {
		daysResult[lastDayMonth+day+1] = true
	}

	return daysResult
}

func firstDayMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
}

func lastDayMonth(t time.Time) time.Time {
	return firstDayMonth(t).AddDate(0, 1, -1)
}
