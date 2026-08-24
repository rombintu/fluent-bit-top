package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"fbtop/internal/app"
	"fbtop/internal/shared"

	tea "charm.land/bubbletea/v2"
)

func main() {
	url := flag.String("u", "http://127.0.0.1:2020", "Fluent Bit URL")
	interval := flag.Duration("i", time.Second, "Poll interval")
	humanize := flag.Bool("H", true, "Humanize numbers")
	sortStr := flag.String("s", "id", "Sort column: id|records|bytes")
	sortDesc := flag.Bool("S", false, "Reverse sort order")
	kindStr := flag.String("k", "all", "Filter kind: input|filter|output|all")
	flag.Parse()

	sc := shared.SortID
	switch *sortStr {
	case "records":
		sc = shared.SortRecords
	case "bytes":
		sc = shared.SortBytes
	}

	k := shared.KindAll
	switch *kindStr {
	case "input":
		k = shared.KindInput
	case "filter":
		k = shared.KindFilter
	case "output":
		k = shared.KindOutput
	}

	m := app.New(*url, *interval, *humanize, sc, !*sortDesc, k)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
