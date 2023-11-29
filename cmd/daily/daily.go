package daily

import (
	"fmt"
	"os"
	"time"
)

type Today struct {
	// this tightly couples Today to the database, but I don't care for now
	Row   int
	Date  string
	Start string
	End   string
}

func (t *Today) Update() {
	if t.Start == "" {
		t.Start = time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stdout, "[%s] Checked in at %s\n", t.Date, t.Start)
		return
	}
	if t.End == "" {
		t.End = time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stdout, "[%s] Checked in at %s\n", t.Date, t.End)
		return
	}
}
