package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strconv"
)

func main() {
	if err := _main(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

type definition struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	RTM    bool   `json:"rtm"`
	Events bool   `json:"events"`
}

func _main() error {
	f, err := os.Open("event-types.json")
	if err != nil {
		return err
	}
	defer f.Close()

	var list []definition
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return err
	}

	if err := writeRTMEvents(list); err != nil {
		return err
	}
	if err := writeEventEvents(list); err != nil {
		return err
	}

	return nil
}

func writeEventEvents(list []definition) error {
	var buf bytes.Buffer
	buf.WriteString("// This file is auto-generated. DO NOT EDIT")
	buf.WriteString("\n\npackage events")
	buf.WriteString("\n\n// These constants match the event types generated by Slack")
	buf.WriteString("\nconst (")
	for _, data := range list {
		if !data.Events {
			continue
		}
		fmt.Fprintf(&buf, "%s = %s\n", data.Name, strconv.Quote(data.Value))
	}
	buf.WriteString("\n)")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	dst, err := os.Create("events/event_types_gen.go")
	if err != nil {
		return err
	}
	defer dst.Close()
	dst.Write(src)
	return nil
}

func writeRTMEvents(list []definition) error {
	var buf bytes.Buffer
	buf.WriteString("// This file is auto-generated. DO NOT EDIT")
	buf.WriteString("\n\npackage rtm")
	buf.WriteString("\n\n// These constants match the event types generated by Slack")
	buf.WriteString("\nconst (")
	for _, data := range list {
		if !data.RTM {
			continue
		}
		fmt.Fprintf(&buf, "%sKey = %s\n", data.Name, strconv.Quote(data.Value))
	}
	buf.WriteString("\n)")

	buf.WriteString("\nconst (")
	buf.WriteString("\nInvalidEventType EventType = iota")
	buf.WriteString("\nClientConnectingEventType // internal")
	buf.WriteString("\nClientDisconnectedEventType // internal")
	for _, data := range list {
		if !data.RTM {
			continue
		}
		fmt.Fprintf(&buf, "\n%s", data.Name)
	}
	buf.WriteString("\n)")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	dst, err := os.Create("rtm/event_types.go")
	if err != nil {
		return err
	}
	defer dst.Close()
	dst.Write(src)
	return nil
}
