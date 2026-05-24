package jarvis

import (
	"time"

	"shanhu.io/g/aries"
)

// DashboardSecurityLogsData encapsulates security logs entries
// for the dashbaord.
type DashboardSecurityLogsData struct {
	Entries []*LogEntry
}

func newDashboardSecurityLogsData(s *server, _ *aries.C) (
	*DashboardSecurityLogsData, error,
) {
	// TODO(h8liu): add pages
	entries, err := s.securityLogs.list(0)
	if err != nil {
		return nil, aries.AltInternal(err, "fail to fetch security logs")
	}

	for _, entry := range entries {
		entry.TSec = time.Unix(0, entry.T).Unix()
	}

	return &DashboardSecurityLogsData{
		Entries: entries,
	}, nil
}
