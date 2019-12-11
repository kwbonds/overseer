package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

var percentageRegex = regexp.MustCompile(`^(\d+(?:.\d+)?)%$`)

// Parses 67.25% to 0.6725
func ParsePercentage(value string) (float32, error) {
	matches := percentageRegex.FindStringSubmatch(value)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid percentage, must be e.g. 14%%")
	}
	percentage100, err := strconv.ParseFloat(matches[1], 32)
	if err != nil {
		return 0, fmt.Errorf("invalid float %s", matches[1])
	}

	if percentage100 < 0 || percentage100 > 100 {
		return 0, fmt.Errorf("value must be > 0%% and <= 100%%")
	}

	percentage1 := float32(percentage100) / 100
	return percentage1, nil
}

// -- percentage Value
type PercentageValue float32

func NewPercentageValue(val float32, p *float32) *PercentageValue {
	*p = val
	return (*PercentageValue)(p)
}

func (i *PercentageValue) Set(s string) error {
	v, err := ParsePercentage(s)
	*i = PercentageValue(v)
	return err
}

func (i *PercentageValue) Get() interface{} { return float32(*i) }

func (i *PercentageValue) String() string { return fmt.Sprintf("%.2f%%", float32(*i)*100) }
