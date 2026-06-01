package stream

import "time"

type Progress struct {
	Table       string
	RowsWritten int64
	TotalRows   int64
	Elapsed     time.Duration
	RowsPerSec  float64
	ETA         time.Duration
}

type ProgressFunc func(Progress)

func calcProgress(table string, written, total int64, elapsed time.Duration) Progress {
	p := Progress{
		Table:       table,
		RowsWritten: written,
		TotalRows:   total,
		Elapsed:     elapsed,
	}

	if elapsed.Seconds() > 0 {
		p.RowsPerSec = float64(written) / elapsed.Seconds()
		if p.RowsPerSec > 0 && written < total {
			remaining := float64(total-written) / p.RowsPerSec
			p.ETA = time.Duration(remaining * float64(time.Second))
		}
	}

	return p
}
