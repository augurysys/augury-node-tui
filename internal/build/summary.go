package build

import (
	"context"

	"github.com/augurysys/augury-node-tui/internal/run"
)

type RowStatus string

const (
	RowStatusSuccess   RowStatus = "success"
	RowStatusFailure   RowStatus = "failure"
	RowStatusSkipped   RowStatus = "skipped"
	RowStatusCancelled RowStatus = "cancelled"
)

type SummaryRow struct {
	PlatformID string
	Status     RowStatus
}

type Summary struct {
	Rows []SummaryRow
}

func ExecuteSequential(ctx context.Context, specs []run.RunSpec) *Summary {
	s := &Summary{Rows: make([]SummaryRow, 0, len(specs))}
	for _, spec := range specs {
		result := run.Execute(ctx, spec)
		switch result.Status {
		case "success":
			s.Rows = append(s.Rows, SummaryRow{PlatformID: spec.Name, Status: RowStatusSuccess})
		case "cancelled":
			s.Rows = append(s.Rows, SummaryRow{PlatformID: spec.Name, Status: RowStatusCancelled})
			for _, rest := range specs[len(s.Rows):] {
				s.Rows = append(s.Rows, SummaryRow{PlatformID: rest.Name, Status: RowStatusSkipped})
			}
			return s
		default:
			s.Rows = append(s.Rows, SummaryRow{PlatformID: spec.Name, Status: RowStatusFailure})
		}
	}
	return s
}
