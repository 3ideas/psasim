package psalerts

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/3ideas/psasim/lib/csvutil"
)

type Alert struct {
	LicenseArea          string    `csv:"License Area"`
	Time                 time.Time // Computed from ALARM_TIME and ALARM_USECS
	Alias                string    // Computed from ALARM_COMPONENT_ALIAS
	AlarmTime            string    `csv:"ALARM_TIME"`
	AlarmUsecs           string    `csv:"ALARM_USECS"`
	ID                   string    `csv:"ID"`
	AlarmID              string    `csv:"ALARM_ID"`
	AlarmInitialTime     string    `csv:"ALARM_INITIAL_TIME"`
	AlarmInitialUsecs    string    `csv:"ALARM_INITIAL_USECS"`
	AlarmPriority        string    `csv:"ALARM_PRIORITY"`
	AlarmType            string    `csv:"ALARM_TYPE"`
	AlarmText            string    `csv:"ALARM_TEXT"`
	AlarmComponentAlias  string    `csv:"ALARM_COMPONENT_ALIAS"`
	AlarmDistrictZone    string    `csv:"ALARM_DISTRICT_ZONE"`
	AlarmSubstationAlias string    `csv:"ALARM_SUBSTATION_ALIAS"`
	AlarmSubstationName  string    `csv:"ALARM_SUBSTATION_NAME"`
	AlarmAckTime         string    `csv:"ALARM_ACK_TIME"`
	AlarmAckUsecs        string    `csv:"ALARM_ACK_USECS"`
	AlarmName            string    `csv:"ALARM_NAME"`
	AlarmBusbarNum       string    `csv:"ALARM_BUSBAR_NUM"`
	AlarmCircuitRef      string    `csv:"ALARM_CIRCUIT_REF"`
	AlarmCircuitName     string    `csv:"ALARM_CIRCUIT_NAME"`
	DeviceType           string    `csv:"DEVICE_TYPE"`
	Area                 string    `csv:"AREA"`
	OperatorAction       string    `csv:"OPERATOR_ACTION"`
	DataSourceID         string    `csv:"DATASOURCEID"`
	LocalDateTime        string    `csv:"LOCALDATETIME"`
	Supplementary        string    `csv:"SUPPLEMENTARY"`
	AlarmText2           string    `csv:"ALARM_TEXT2"`
	Descriptor           string    `csv:"DESCRIPTOR"`
	ComponentPathname    string    `csv:"COMPONENT_PATHNAME"`
	PrimaryBusbar        string    `csv:"PRIMARY_BUSBAR"`
	PrimaryFeeder        string    `csv:"PRIMARY_FEEDER"`
	RequeriedAt          string    `csv:"REQUERIED_AT"`
}

// PostProcess processes the alarm after CSV reading to set computed fields
func (a *Alert) PostProcess() error {
	// Parse the time string
	t, err := time.Parse("2006-01-02 15:04:05", a.AlarmTime)
	if err != nil {
		return fmt.Errorf("parsing alarm time: %s  Error: %w", a.AlarmTime, err)
	}

	// Convert microseconds to duration and add to time
	usecs := 0
	if a.AlarmUsecs != "" {
		fmt.Sscanf(a.AlarmUsecs, "%d", &usecs)
	}
	microseconds := time.Duration(usecs) * time.Microsecond

	// Set the computed time field
	a.Time = t.Add(microseconds)

	// Process the alias
	a.Alias = a.AlarmComponentAlias

	parts := strings.Split(a.AlarmComponentAlias, ".")
	if len(parts) == 4 {
		// Replace dots with forward slashes
		a.Alias = strings.ReplaceAll(a.AlarmComponentAlias, ".", "/")
	}

	// if name is of the form COCK2.275_CB.W10.SWDD.OPEN
	// then the alias is COCK2/275_CB/W10/SWDD
	if len(parts) == 5 {
		// Join the first 4 parts with forward slashes
		a.Alias = strings.Join(parts[:4], "/")
	}
	return nil
}

type PSAlerts struct {
	Alerts []*Alert
}

func ReadPSAlerts(filename string) (*PSAlerts, error) {
	alerts, err := csvutil.ReadItems[*Alert](filename)
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	parsedAlarms := make([]*Alert, 0, len(alerts))
	// Post-process each alarm to set computed fields
	for _, alert := range alerts {
		if alert.AlarmComponentAlias == "D06R0718C16D3-6" {
			fmt.Println(alert)
		}
		if alert.AlarmTime == "" {
			continue
		}
		if alert.AlarmSubstationName == "COMMS" {
			continue
		}
		if alert.AlarmName == "SWGR IED COMMS" || alert.AlarmName == "SWG IED COMMS" {
			continue
		}
		if err := alert.PostProcess(); err != nil {
			return nil, fmt.Errorf("post-processing alarm: %w", err)
		}
		parsedAlarms = append(parsedAlarms, alert)
	}

	return &PSAlerts{Alerts: parsedAlarms}, nil
}

func (p *PSAlerts) GetAlertCounts() map[string]int {
	alarmCounts := make(map[string]int)
	for _, alarm := range p.Alerts {
		alarmCounts[alarm.AlarmType]++
	}
	return alarmCounts
}

// PrintAlarmCounts prints the alarm counts to the console, sorted from highest to lowest count
func (p *PSAlerts) PrintAlarmCounts() {
	alarmCounts := p.GetAlertCounts()

	// Create a slice of key-value pairs for sorting
	type kv struct {
		Type  string
		Count int
	}

	pairs := make([]kv, 0, len(alarmCounts))
	for k, v := range alarmCounts {
		pairs = append(pairs, kv{k, v})
	}

	// Sort by count in descending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	// Print sorted results
	for _, pair := range pairs {
		fmt.Printf("%-25s: %d\n", pair.Type, pair.Count)
	}
}
