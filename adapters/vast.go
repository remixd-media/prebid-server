package adapters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type VAST struct {
	Ads []Ad `xml:"Ad"`
}

type Ad struct {
	ID      string   `xml:"id,attr"`
	InLine  *InLine  `xml:"InLine"`
	Wrapper *Wrapper `xml:"Wrapper"`
}

func (a Ad) GetPricing() (float64, error) {
	if a.InLine != nil {
		return strconv.ParseFloat(a.InLine.Pricing, 64)
	}
	if a.Wrapper != nil {
		return strconv.ParseFloat(a.Wrapper.Pricing, 64)
	}

	return 0, fmt.Errorf("no inline or wrapper")
}

func (a Ad) GetCreativeId() string {
	if a.InLine != nil && len(a.InLine.Creatives.Creative) > 0 {
		return a.InLine.Creatives.Creative[0].ID
	}
	if a.Wrapper != nil && len(a.Wrapper.Creatives.Creative) > 0 {
		return a.Wrapper.Creatives.Creative[0].ID
	}
	return ""
}

type InLine struct {
	Pricing   string    `xml:"Pricing"`
	Creatives Creatives `xml:"Creatives"`
}

type Wrapper struct {
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

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		// ignore
	}

	return (hours * 60 * 60) + (minutes * 60) + int(seconds)
}

var pricingHiderRe *regexp.Regexp

func init() {
	pricingHiderRe = regexp.MustCompile(`(<Pricing.*>)\s*.*\s*(</Pricing>)`)
}

func HidePricing(adm string) string {
	return pricingHiderRe.ReplaceAllString(adm, ``)
}
