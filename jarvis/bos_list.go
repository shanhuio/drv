package jarvis

import (
	"strings"

	"shanhu.io/g/errcode"
)

type bosListEntry struct {
	name      string
	tags      []string
	local     bool
	remote    bool
	latest    bool
	running   bool
	available bool
}

func parseBosListEntry(line string) (*bosListEntry, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, errcode.InvalidArgf("empty line")
	}

	entry := &bosListEntry{name: fields[0]}
	entry.tags = append(entry.tags, fields[1:]...)
	for _, tag := range entry.tags {
		switch tag {
		case "local":
			entry.local = true
		case "remote":
			entry.remote = true
		case "running":
			entry.running = true
		case "available":
			entry.available = true
		case "latest":
			entry.latest = true
		}
	}
	return entry, nil
}

type bosList struct {
	entries []*bosListEntry
}

func (ls *bosList) find(v string) *bosListEntry {
	for _, entry := range ls.entries {
		if entry.name == v {
			return entry
		}
	}
	return nil
}

func (ls *bosList) running() *bosListEntry {
	for _, entry := range ls.entries {
		if entry.running {
			return entry
		}
	}
	return nil
}

func parseOSList(lines []string) (*bosList, error) {
	list := new(bosList)
	for _, line := range lines {
		entry, err := parseBosListEntry(line)
		if err != nil {
			return nil, errcode.Annotatef(err, "parse line: %q", line)
		}
		list.entries = append(list.entries, entry)
	}

	return list, nil
}
