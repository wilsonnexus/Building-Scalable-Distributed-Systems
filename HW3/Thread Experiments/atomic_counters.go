package main

import (
	"fmt"
	"image/color"
	"sync"
	"sync/atomic"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	const goroutines = 50
	const iters = 1000
	const runs = 10

	expected := uint64(goroutines * iters)

	atomicResults := make([]uint64, runs)
	regularResults := make([]uint64, runs)

	for r := 0; r < runs; r++ {
		var atomicOps atomic.Uint64
		var regularOps uint64

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iters; j++ {
					atomicOps.Add(1)
					regularOps++ // intentional race
				}
			}()
		}

		wg.Wait()
		atomicResults[r] = atomicOps.Load()
		regularResults[r] = regularOps
	}

	// ---- Plot ----
	p := plot.New()
	p.Title.Text = "Atomic vs Regular Counter (race shows up in regular)"
	p.X.Label.Text = "Run"
	p.Y.Label.Text = "Count"

	// X labels: 1..runs
	labels := make([]string, runs)
	for i := range labels {
		labels[i] = fmt.Sprintf("%d", i+1)
	}
	p.NominalX(labels...)

	atomicVals := make(plotter.Values, runs)
	regularVals := make(plotter.Values, runs)
	for i := 0; i < runs; i++ {
		atomicVals[i] = float64(atomicResults[i])
		regularVals[i] = float64(regularResults[i])
	}

	// Bars
	w := vg.Points(14)

	atomicBars, _ := plotter.NewBarChart(atomicVals, w)
	atomicBars.Color = color.RGBA{50, 100, 255, 255}
	atomicBars.Offset = -w / 2

	regularBars, _ := plotter.NewBarChart(regularVals, w)
	regularBars.Color = color.RGBA{220, 50, 50, 255}
	regularBars.Offset = w / 2

	p.Add(atomicBars, regularBars)
	p.Legend.Add("atomic", atomicBars)
	p.Legend.Add("regular (race)", regularBars)

	// Expected line
	linePts := plotter.XYs{
		{X: -0.5, Y: float64(expected)},
		{X: float64(runs) - 0.5, Y: float64(expected)},
	}
	expLine, _ := plotter.NewLine(linePts)
	expLine.Color = color.RGBA{120, 120, 120, 255}
	p.Add(expLine)
	p.Legend.Add("expected", expLine)

	// Value labels ABOVE each bar (so you can explain numbers without printing)
	var labelPts plotter.XYs
	var labelTxt []string

	for i := 0; i < runs; i++ {
		// atomic label
		labelPts = append(labelPts, plotter.XY{X: float64(i) - 0.18, Y: float64(atomicResults[i])})
		labelTxt = append(labelTxt, fmt.Sprintf("%d", atomicResults[i]))

		// regular label
		labelPts = append(labelPts, plotter.XY{X: float64(i) + 0.18, Y: float64(regularResults[i])})
		labelTxt = append(labelTxt, fmt.Sprintf("%d", regularResults[i]))
	}

	valueLabels, _ := plotter.NewLabels(plotter.XYLabels{
		XYs:    labelPts,
		Labels: labelTxt,
	})
	p.Add(valueLabels)

	// Save image (size matters for readability)
	if err := p.Save(10*vg.Inch, 5*vg.Inch, "counter_comparison.png"); err != nil {
		panic(err)
	}

	fmt.Println("expected:", expected)
	fmt.Println("image saved as counter_comparison.png")
}
