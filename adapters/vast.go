package adapters

import (
	"regexp"
	"strconv"
	"strings"
)

type VAST struct {
	Ads []Ad `xml:"Ad"`
}

type Ad struct {
	ID     string `xml:"id,attr"`
	InLine InLine `xml:"InLine"`
}

type InLine struct {
	Pricing   string    `xml:"Pricing"`
	Creatives Creatives `xml:"Creatives"`
}

type Creatives struct {
	Creative []Creative `xml:"Creative"`
}

type Creative struct {
	ID     string `xml:"id,attr"`
	Linear Linear `xml:"Linear"`
}

type Linear struct {
	Duration string `xml:"Duration"`
}

func ParseDuration(dur string) int {
	parts := strings.Split(dur, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		// ignore
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		// ignore
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		// ignore
	}

	return (hours * 60 * 60) + (minutes * 60) + seconds
}

var pricingHiderRe *regexp.Regexp

func init() {
	pricingHiderRe = regexp.MustCompile(`(<Pricing.*>)\s*.*\s*(</Pricing>)`)
}

func HidePricing(adm string) string {
	return pricingHiderRe.ReplaceAllString(adm, ``)
}
