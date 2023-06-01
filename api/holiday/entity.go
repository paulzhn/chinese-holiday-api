package holiday

type OriginJson struct {
	Year   string   `json:"year"`
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
	Date     string `json:"date"`
	Name     string `json:"name"`
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