package main

import (
	"slices"
	"strconv"

	"github.com/jpicht/gira"
	"github.com/jpicht/gira/data"
)

type channel struct {
	Offset int
	Name   string
	UID    string
	Max    float64
}

type fixture struct {
	Name     string
	UID      string
	Offset   int
	Channels []channel
}

type onOff struct {
	channels []int
	uid      string
}

type channels struct {
	cfg      *gira.Config
	offset   int
	channels [512]channel
	onOff    []onOff
}

type channelValues [512]byte

func (c *channels) diff(a, b channelValues) *data.ValueBody {
	out := new(data.ValueBody)

	// tracks changed channel ID
	channels := make([]int, 0, 512)

	for i := 0; i < c.offset; i++ {
		if a[i] == b[i] {
			continue
		}
		channels = append(channels, i)
		out.Values = append(out.Values, data.Value{
			UID:   c.channels[i].UID,
			Value: strconv.Itoa(int(float64(b[i]) * 100.0 / 255.0)),
		})
	}

	// nothing to do, bail out early
	if len(out.Values) == 0 {
		return nil
	}

	if !c.cfg.AutoOnOff {
		return out
	}

	// calculate values of OnOff channels
	for _, ooi := range c.onOff {
		value := "0"
		changed := false
		for _, c := range ooi.channels {
			changed = changed || slices.Contains(channels, c)
			if b[c] > 0 {
				value = "1"
			}
		}

		if !changed {
			continue
		}

		out.Values = append(out.Values, data.Value{
			UID:   ooi.uid,
			Value: value,
		})
	}

	return out
}

func (c *channels) addChannel(fn data.Function, dp data.DataPoint) channel {
	next := channel{
		Offset: c.offset,
		Name:   fn.DisplayName + "/" + dp.Name,
		UID:    dp.UID,
	}
	c.channels[c.offset] = next
	c.offset++
	return next
}

func (c *channels) addFunction(fn data.Function) fixture {
	channels := make([]channel, len(fn.DataPoints))
	var onOff onOff

	for i, dp := range fn.DataPoints {
		if c.cfg.AutoOnOff && dp.Name == "OnOff" {
			onOff.uid = dp.UID
			continue
		}
		channels[i] = c.addChannel(fn, dp)
		onOff.channels = append(onOff.channels, channels[i].Offset)
	}

	c.onOff = append(c.onOff, onOff)

	fix := fixture{
		Name:     fn.DisplayName,
		UID:      fn.UID,
		Offset:   c.offset,
		Channels: channels,
	}

	return fix
}
