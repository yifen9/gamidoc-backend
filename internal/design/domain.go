package design

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

const SectionCount = 7

const (
	PathExperienceFirst = "A"
	PathMechanicsFirst  = "B"
)

const (
	SectionStatusNotStarted = "not_started"
	SectionStatusInProgress = "in_progress"
	SectionStatusComplete   = "complete"
)

const (
	ReportVersionStandard = "standard"
	ReportVersionEnhanced = "enhanced"
)

var ErrInvalidSectionNumber = errors.New("invalid section number")
var ErrInvalidSectionData = errors.New("invalid section data")
var ErrInvalidPath = errors.New("invalid path")
var ErrPathAlreadyChosen = errors.New("path already chosen")
var ErrPathNotChosen = errors.New("path not chosen")
var ErrSectionLocked = errors.New("section locked")
var ErrDashboardLocked = errors.New("dashboard locked")
var ErrInvalidReportVersion = errors.New("invalid report version")

var sectionNames = map[int]string{
	1: "Context",
	2: "Experience Timeline",
	3: "Personification & Dynamics",
	4: "Gameful Core",
	5: "Technology",
	6: "Impacts & Benefits",
	7: "Evaluation & Feedback",
}

func SectionName(number int) string {
	return sectionNames[number]
}

func Order(path string) []int {
	switch path {
	case PathExperienceFirst:
		return []int{1, 2, 3, 4, 5, 6, 7}
	case PathMechanicsFirst:
		return []int{1, 4, 5, 2, 3, 6, 7}
	default:
		return nil
	}
}

type SectionState struct {
	Content  json.RawMessage `json:"content,omitempty"`
	Complete bool            `json:"complete"`
	Visited  bool            `json:"visited"`
}

type Status struct {
	Spark         string                  `json:"spark,omitempty"`
	Path          string                  `json:"path,omitempty"`
	Cursor        int                     `json:"cursor"`
	FirstPassDone bool                    `json:"firstPassDone"`
	Sections      map[string]SectionState `json:"sections"`
}

func NewInitialStatus() Status {
	return Status{
		Cursor:   0,
		Sections: map[string]SectionState{},
	}
}

func (s Status) Section(number int) SectionState {
	return s.Sections[SectionKey(number)]
}

func (s Status) HasContent() bool {
	for _, state := range s.Sections {
		if len(state.Content) > 0 {
			return true
		}
	}
	return false
}

func SectionKey(number int) string {
	return strconv.Itoa(number)
}

func SectionStatus(state SectionState) string {
	if len(state.Content) == 0 {
		return SectionStatusNotStarted
	}
	if state.Complete {
		return SectionStatusComplete
	}
	return SectionStatusInProgress
}

func SectionPercent(state SectionState) int {
	switch SectionStatus(state) {
	case SectionStatusComplete:
		return 100
	case SectionStatusInProgress:
		return 50
	default:
		return 0
	}
}

type DashboardSection struct {
	SectionNumber int    `json:"sectionNumber"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Percent       int    `json:"percent"`
}

type Dashboard struct {
	Sections       []DashboardSection `json:"sections"`
	OverallPercent int                `json:"overallPercent"`
	FirstPassDone  bool               `json:"firstPassDone"`
	Path           string             `json:"path,omitempty"`
}

type Report struct {
	ID        string    `json:"reportId"`
	ProjectID string    `json:"-"`
	Version   string    `json:"version"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}
