// plot_results.go
package main

import (
	"fmt"
	"image/color"
	"log"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func mean(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func main() {
	// Times in milliseconds (ms)
	mutex := []float64{11.7941, 15.2073, 16.1811}
	rwmutex := []float64{16.0711, 21.1915, 23.9709}
	syncMap := []float64{9.2804, 11.6263, 12.6783}

	methods := []string{"Mutex", "RWMutex", "sync.Map"}

	// Bar values per run (each slice is one run across methods)
	run1 := plotter.Values{mutex[0], rwmutex[0], syncMap[0]}
	run2 := plotter.Values{mutex[1], rwmutex[1], syncMap[1]}
	run3 := plotter.Values{mutex[2], rwmutex[2], syncMap[2]}

	p := plot.New()
	p.Title.Text = "Collections: Time per Run (lower is better)"
	p.Y.Label.Text = "Time (ms)"
	p.NominalX(methods...)

	// Bar width in "category units"
	w := vg.Points(18)

	b1, err := plotter.NewBarChart(run1, w)
	if err != nil {
		log.Fatal(err)
	}
	b2, err := plotter.NewBarChart(run2, w)
	if err != nil {
		log.Fatal(err)
	}
	b3, err := plotter.NewBarChart(run3, w)
	if err != nil {
		log.Fatal(err)
	}

	// Offset bars so they show side-by-side per method
	b1.Offset = -w
	b2.Offset = 0
	b3.Offset = w

	// Colors (just to distinguish runs)
	b1.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255} // steel-ish
	b2.Color = color.RGBA{R: 60, G: 179, B: 113, A: 255} // green-ish
	b3.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}  // red-ish

	p.Add(b1, b2, b3)

	p.Legend.Add(fmt.Sprintf("run 1  (means: Mutex %.2f, RW %.2f, sync.Map %.2f)",
		mean(mutex), mean(rwmutex), mean(syncMap)), b1)
	p.Legend.Add("run 2", b2)
	p.Legend.Add("run 3", b3)

	// Make the output readable
	p.Legend.Top = true
	p.Add(plotter.NewGrid())

	if err := p.Save(7*vg.Inch, 4*vg.Inch, "collections_results.png"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("saved: collections_results.png")
}
