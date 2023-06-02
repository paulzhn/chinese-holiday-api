package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// nesting map
	// level 1 key: year
	// level 2 key: month
	// level 3 key: day
	// level 3 value: DateDetail
	dayOffMap = map[int]map[int]map[int]DateDetail{}
	mu        sync.Mutex
)

func stringToVerboseLevel(s string) VerboseLevel {
	switch s {
	case "0":
		return VerboseLevelLow
	case "1":
		return VerboseLevelMedium
	case "2":
		return VerboseLevelHigh
	default:
		return VerboseLevelLow
	}
}

func errHandler(w http.ResponseWriter, err error, verbose VerboseLevel) {
	if verbose == VerboseLevelHigh {
		resp := ReturnJson{
			Code:    -1,
			Message: err.Error(),
		}
		fmt.Fprintf(w, "%v", resp)
		return
	}
	fmt.Fprintf(w, "%v", err.Error())
}

func resultHandler(w http.ResponseWriter, dateDetail []DateDetail, verbose VerboseLevel) {
	if len(dateDetail) != 1 {
		resp := ReturnJson{
			Code:    0,
			Message: "success",
			Data:    dateDetail,
		}
		res, _ := json.Marshal(resp)
		w.Write(res)
		return
	}
	switch verbose {
	case VerboseLevelLow:
		dateType := dateDetail[0].DateType
		if dateType == Holiday || dateType == PlainOffDay {
			dateType = PlainOffDay
		} else {
			dateType = PlainWorkDay
		}
		fmt.Fprintf(w, "%v", dateType)
	case VerboseLevelMedium:
		fmt.Fprintf(w, "%v", dateDetail[0].DateType)
	case VerboseLevelHigh:
		resp := ReturnJson{
			Code:    0,
			Message: "success",
			Data:    dateDetail,
		}
		res, _ := json.Marshal(resp)
		w.Write(res)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	verbose := stringToVerboseLevel(r.URL.Query().Get("verbose"))
	date := r.URL.Query().Get("date")

	parsedDate, err := time.Parse("2006", date)
	if err == nil {
		dateDetail, err := judgeYear(parsedDate)
		if err != nil {
			errHandler(w, err, verbose)
			return
		}
		resultHandler(w, dateDetail, verbose)
		return
	}

	parsedDate, err = time.Parse("2006-01", date)
	if err == nil {
		dateDetail, err := judgeMonth(parsedDate)
		if err != nil {
			errHandler(w, err, verbose)
			return
		}
		resultHandler(w, dateDetail, verbose)
		return
	}

	parsedDate, err = time.Parse("2006-01-02", date)
	if err != nil {
		parsedDate = time.Now()
	}

	dateDetail, err := judgeDate(parsedDate)
	if err != nil {
		errHandler(w, err, verbose)
		return
	}
	resultHandler(w, dateDetail, verbose)
}

func judgeYear(date time.Time) ([]DateDetail, error) {
	res := []DateDetail{}
	for i := 1; i <= 12; i++ {
		nowRes, err := judgeMonth(time.Date(date.Year(), time.Month(i), 1, 0, 0, 0, 0, time.Local))
		if err != nil {
			return nil, err
		}
		res = append(res, nowRes...)
	}
	return res, nil
}

func judgeMonth(date time.Time) ([]DateDetail, error) {
	year := date.Year()
	month := int(date.Month())
	yearMap := dayOffMap[year]
	var err error
	if yearMap == nil {
		_, err = getMonthMap(year)
		yearMap = dayOffMap[year]
	}
	if yearMap == nil || err != nil {
		errMsg := fmt.Sprintf("无法获取%v年的数据", year)
		log.Println(errMsg)
		log.Println(err)
		return nil, fmt.Errorf(errMsg)
	}
	

	monthMap := yearMap[month]
	res := []DateDetail{}
	if monthMap == nil {
		return res, nil
	}

	for _, v := range monthMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Date < res[j].Date
	})
	return res, nil
}

func judgeDate(parsedDate time.Time) ([]DateDetail, error) {
	year := parsedDate.Year()
	month := parsedDate.Month()
	day := parsedDate.Day()

	monthMap, err := getMonthMap(year)
	if err != nil {
		return nil, err
	}

	dayMap, ok := monthMap[int(month)]
	if !ok {
		return []DateDetail{
			{
				Date:     parsedDate.Format("2006-01-02"),
				Name:     "",
				DateType: judgeWorkDayOrOff(parsedDate),
			},
		}, nil
	}

	dayDetail, ok := dayMap[day]
	if !ok {
		return []DateDetail{
			{
				Date:     parsedDate.Format("2006-01-02"),
				Name:     "",
				DateType: judgeWorkDayOrOff(parsedDate),
			},
		}, nil
	}
	return []DateDetail{dayDetail}, nil
}

func judgeWorkDayOrOff(d time.Time) DateTypeEnum {
	if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		return PlainOffDay
	} else {
		return PlainWorkDay
	}
}

func downloadJsonFile(year int) ([]byte, error) {
	// https://github.com/NateScarlet/holiday-cn
	url := fmt.Sprintf("https://raw.githubusercontent.com/NateScarlet/holiday-cn/master/%d.json", year)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		err = fmt.Errorf("所请求的年份数据不存在")
		return nil, err
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	return content, err
}

func getMonthMap(year int) (map[int]map[int]DateDetail, error) {
	// load from json
	mu.Lock()
	defer mu.Unlock()
	if s, ok := dayOffMap[year]; ok {
		return s, nil
	}
	content, err := downloadJsonFile(year)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	originJson := OriginJson{}
	err = json.Unmarshal(content, &originJson)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	holidayLastSituation := map[string]bool{}

	for _, v := range originJson.Days {
		dateComponentArr := strings.Split(v.Date, "-")
		if len(dateComponentArr) != 3 {
			log.Printf("invalidate date: %v", v.Date)
			err = fmt.Errorf("无效的日期")
			return nil, err
		}

		month, _ := strconv.Atoi(dateComponentArr[1])
		day, _ := strconv.Atoi(dateComponentArr[2])

		if dayOffMap[year] == nil {
			dayOffMap[year] = map[int]map[int]DateDetail{}
		}
		monthDayMap := dayOffMap[year]

		if monthDayMap[month] == nil {
			monthDayMap[month] = map[int]DateDetail{}
		}
		dayMap := monthDayMap[month]

		if _, ok := dayMap[day]; ok {
			log.Printf("duplicate date: %v", v.Date)
			err = fmt.Errorf("无效的日期")
			return nil, err
		}
		var dateType DateTypeEnum
		if !v.IsOffDay {
			if s, ok := holidayLastSituation[v.Name]; !ok || !s {
				dateType = WorkDayBeforeHoliday
			} else {
				dateType = WorkDayAfterHoliday
			}

		} else {
			dateType = Holiday
		}
		holidayLastSituation[v.Name] = v.IsOffDay

		dateDetail := DateDetail{
			Date:     v.Date,
			Name:     v.Name,
			DateType: dateType,
		}
		dayMap[day] = dateDetail
	}

	return dayOffMap[year], nil
}

type OriginJson struct {
	Year   int      `json:"year"`
	Papers []string `json:"papers"`
	Days   []struct {
		Name     string `json:"name"`
		Date     string `json:"date"`
		IsOffDay bool   `json:"isOffDay"`
	} `json:"days"`
}

type DateTypeEnum int

const (
	PlainWorkDay DateTypeEnum = iota
	PlainOffDay
	Holiday
	WorkDayBeforeHoliday
	WorkDayAfterHoliday
)

type DateDetail struct {
	Date     string       `json:"date"`
	Name     string       `json:"name"`
	DateType DateTypeEnum `json:"type"`
}

type ReturnJson struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    []DateDetail `json:"data"`
}

type VerboseLevel int

const (
	VerboseLevelLow VerboseLevel = iota
	VerboseLevelMedium
	VerboseLevelHigh
)
