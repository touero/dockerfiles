// Package chart renders modern-looking line charts as standalone SVG.
//
// It is dependency-free: the SVG uses system fonts (via a font stack) and
// works when embedded with <img> or rendered by a browser/GitHub.
package chart

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
	"time"
)

// Point is a single sample.
type Point struct {
	X time.Time
	Y float64
}

// Series is one line.
type Series struct {
	Name   string
	Points []Point
}

// Options controls a chart.
type Options struct {
	Title     string
	YLabel    string
	Series    []Series
	LogY      bool // logarithmic y axis (cumulative counts)
	ZeroBase  bool // force the y axis to start at 0 (per-day deltas)
	EndLabels bool // annotate the last value of each line
	Dark      bool // dark colour theme
	Width     float64
	Height    float64
}

type theme struct {
	bg     string
	title  string
	tick   string
	grid   string
	axis   string
	label  string
	colors []string
}

func themeFor(dark bool) theme {
	if dark {
		return theme{
			bg: "#0d1117", title: "#e6edf3", tick: "#768390", grid: "#1f242c",
			axis: "#30363d", label: "#adbac7",
			colors: []string{"#818cf8", "#fbbf24", "#34d399", "#f87171", "#a78bfa", "#22d3ee", "#f472b6", "#94a3b8"},
		}
	}
	return theme{
		bg: "#ffffff", title: "#111827", tick: "#9aa4b2", grid: "#eef1f5",
		axis: "#d6dbe2", label: "#5b6470",
		colors: []string{"#6366f1", "#f59e0b", "#10b981", "#ef4444", "#8b5cf6", "#06b6d4", "#ec4899", "#64748b"},
	}
}

const (
	marginLeft   = 78.0
	marginRight  = 98.0
	marginTop    = 70.0
	marginBottom = 54.0
)

// Render returns a complete SVG document.
func Render(o Options) string {
	if len(o.Series) == 0 {
		return ""
	}
	if o.Width == 0 {
		o.Width = 900
	}
	if o.Height == 0 {
		o.Height = 440
	}
	th := themeFor(o.Dark)

	x0, x1 := marginLeft, o.Width-marginRight
	y0, y1 := marginTop, o.Height-marginBottom

	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)
	for _, s := range o.Series {
		for _, p := range s.Points {
			minX = math.Min(minX, float64(p.X.Unix()))
			maxX = math.Max(maxX, float64(p.X.Unix()))
			minY = math.Min(minY, p.Y)
			maxY = math.Max(maxY, p.Y)
		}
	}
	if math.IsInf(minX, 1) {
		return ""
	}
	if maxX == minX {
		minX -= 12 * 3600
		maxX += 12 * 3600
	}
	if minY == maxY {
		if minY == 0 {
			maxY = 1
		} else {
			minY *= 0.9
			maxY *= 1.1
		}
	}

	switch {
	case o.LogY:
		minY = niceFloor(math.Max(minY, 1))
		maxY = niceCeil(maxY)
		if minY <= 0 {
			minY = 1
		}
		if maxY <= minY {
			maxY = minY * 10
		}
	case o.ZeroBase:
		minY = 0
		maxY = niceCeil(maxY)
	default:
		minY = niceFloor(minY)
		maxY = niceCeil(maxY)
		if maxY <= minY {
			maxY = minY + 1
		}
	}

	sx := func(t time.Time) float64 {
		return x0 + (float64(t.Unix())-minX)/(maxX-minX)*(x1-x0)
	}
	sy := func(v float64) float64 {
		if o.LogY {
			lo, hi := math.Log10(minY), math.Log10(maxY)
			return y1 - (math.Log10(math.Max(v, minY))-lo)/(hi-lo)*(y1-y0)
		}
		return y1 - (v-minY)/(maxY-minY)*(y1-y0)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g" viewBox="0 0 %g %g" role="img" aria-label="%s">`,
		o.Width, o.Height, o.Width, o.Height, html.EscapeString(o.Title))

	// Gradient defs (one per series).
	b.WriteString("<defs>")
	for i := range o.Series {
		c := th.colors[i%len(th.colors)]
		fmt.Fprintf(&b, `<linearGradient id="grad%d" gradientUnits="userSpaceOnUse" x1="0" y1="%s" x2="0" y2="%s">`+
			`<stop offset="0" stop-color="%s" stop-opacity="0.20"/>`+
			`<stop offset="1" stop-color="%s" stop-opacity="0"/></linearGradient>`,
			i, f(y0), f(y1), c, c)
	}
	b.WriteString("</defs>")

	fmt.Fprintf(&b, `<style>text{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif}`+
		`.t{font-size:16px;font-weight:600;fill:%s}`+
		`.y{font-size:11px;fill:%s}`+
		`.x{font-size:11px;fill:%s}`+
		`.l{font-size:12px;fill:%s}`+
		`.e{font-size:11px;font-weight:600}</style>`,
		th.title, th.tick, th.tick, th.label)

	fmt.Fprintf(&b, `<rect width="%g" height="%g" fill="%s"/>`, o.Width, o.Height, th.bg)

	// Title + y label.
	fmt.Fprintf(&b, `<text class="t" x="%s" y="36">%s</text>`, f(x0), html.EscapeString(o.Title))
	if o.YLabel != "" {
		cy := (y0 + y1) / 2
		fmt.Fprintf(&b, `<text class="y" x="20" y="%s" transform="rotate(-90 20 %s)" text-anchor="middle">%s</text>`,
			f(cy), f(cy), html.EscapeString(o.YLabel))
	}

	// Horizontal grid + y tick labels.
	for _, v := range yTicks(minY, maxY, o.LogY) {
		yy := sy(v)
		fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="1"/>`,
			f(x0), f(yy), f(x1), f(yy), th.grid)
		fmt.Fprintf(&b, `<text class="y" x="%s" y="%s" text-anchor="end">%s</text>`,
			f(x0-12), f(yy+4), humanize(v))
	}

	// Baseline + x tick labels.
	fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="1.5"/>`,
		f(x0), f(y1), f(x1), f(y1), th.axis)
	layout := "01-02"
	if maxX-minX > 370*86400 {
		layout = "2006-01"
	}
	for _, t := range xTicks(minX, maxX) {
		fmt.Fprintf(&b, `<text class="x" x="%s" y="%s" text-anchor="middle">%s</text>`,
			f(sx(t)), f(y1+24), t.Format(layout))
	}

	// Area fills first (back to front) so they never tint the lines.
	for i, s := range o.Series {
		if len(s.Points) < 2 {
			continue
		}
		xs := make([]float64, len(s.Points))
		ys := make([]float64, len(s.Points))
		for j, p := range s.Points {
			xs[j] = sx(p.X)
			ys[j] = sy(p.Y)
		}
		path := smoothPath(xs, ys)
		fmt.Fprintf(&b, `<path d="%s L %s %s L %s %s Z" fill="url(#grad%d)" stroke="none"/>`,
			path, f(xs[len(xs)-1]), f(y1), f(xs[0]), f(y1), i)
	}

	// Lines, markers and end labels on top.
	for i, s := range o.Series {
		col := th.colors[i%len(th.colors)]
		if len(s.Points) == 1 {
			px, py := sx(s.Points[0].X), sy(s.Points[0].Y)
			fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="3.5" fill="%s" stroke="%s" stroke-width="2"/>`,
				f(px), f(py), th.bg, col)
			if o.EndLabels {
				fmt.Fprintf(&b, `<text class="e" x="%s" y="%s" fill="%s">%s</text>`,
					f(px+10), f(py+4), col, comma(s.Points[0].Y))
			}
			continue
		}
		if len(s.Points) < 2 {
			continue
		}
		xs := make([]float64, len(s.Points))
		ys := make([]float64, len(s.Points))
		for j, p := range s.Points {
			xs[j] = sx(p.X)
			ys[j] = sy(p.Y)
		}
		path := smoothPath(xs, ys)

		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>`, path, col)

		// Point markers (only while they stay readable).
		if len(s.Points) <= 60 {
			for j := range xs {
				fmt.Fprintf(&b, `<circle cx="%s" cy="%s" r="3" fill="%s" stroke="%s" stroke-width="2"/>`,
					f(xs[j]), f(ys[j]), th.bg, col)
			}
		}
		// Value label at the end of the line.
		if o.EndLabels {
			last := s.Points[len(s.Points)-1]
			fmt.Fprintf(&b, `<text class="e" x="%s" y="%s" fill="%s">%s</text>`,
				f(x1+10), f(sy(last.Y)+4), col, comma(last.Y))
		}
	}

	// Legend, right aligned at the top.
	legend(&b, o.Series, th, x1, y0-34)

	b.WriteString("</svg>\n")
	return b.String()
}

// legend draws a horizontal legend ending at right.
func legend(b *strings.Builder, series []Series, th theme, right float64, y float64) {
	const swatch, swatchGap, itemGap = 16.0, 7.0, 20.0
	widths := make([]float64, len(series))
	total := 0.0
	for i, s := range series {
		widths[i] = swatch + swatchGap + float64(len([]rune(s.Name)))*6.8
		total += widths[i]
	}
	total += itemGap * float64(len(series)-1)

	x := right - total
	for i, s := range series {
		col := th.colors[i%len(th.colors)]
		fmt.Fprintf(b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="3" stroke-linecap="round"/>`,
			f(x), f(y), f(x+swatch), f(y), col)
		fmt.Fprintf(b, `<text class="l" x="%s" y="%s">%s</text>`,
			f(x+swatch+swatchGap), f(y+4), html.EscapeString(s.Name))
		x += widths[i] + itemGap
	}
}

// smoothPath builds a monotone cubic (Fritsch-Carlson) path so the curve never
// overshoots the data.
func smoothPath(x, y []float64) string {
	n := len(x)
	var b strings.Builder
	fmt.Fprintf(&b, "M %s %s", f(x[0]), f(y[0]))
	if n < 3 {
		for i := 1; i < n; i++ {
			fmt.Fprintf(&b, " L %s %s", f(x[i]), f(y[i]))
		}
		return b.String()
	}

	m := make([]float64, n-1)
	for i := 0; i < n-1; i++ {
		dx := x[i+1] - x[i]
		if dx == 0 {
			m[i] = 0
		} else {
			m[i] = (y[i+1] - y[i]) / dx
		}
	}
	t := make([]float64, n)
	t[0] = m[0]
	t[n-1] = m[n-2]
	for i := 1; i < n-1; i++ {
		if m[i-1]*m[i] <= 0 {
			t[i] = 0
		} else {
			t[i] = (m[i-1] + m[i]) / 2
		}
	}
	for i := 0; i < n-1; i++ {
		if m[i] == 0 {
			t[i], t[i+1] = 0, 0
			continue
		}
		alpha := t[i] / m[i]
		beta := t[i+1] / m[i]
		if s := alpha*alpha + beta*beta; s > 9 {
			tau := 3 / math.Sqrt(s)
			t[i] = tau * alpha * m[i]
			t[i+1] = tau * beta * m[i]
		}
	}
	for i := 0; i < n-1; i++ {
		dx := x[i+1] - x[i]
		fmt.Fprintf(&b, " C %s %s %s %s %s %s",
			f(x[i]+dx/3), f(y[i]+t[i]*dx/3),
			f(x[i+1]-dx/3), f(y[i+1]-t[i+1]*dx/3),
			f(x[i+1]), f(y[i+1]))
	}
	return b.String()
}

func yTicks(min, max float64, logScale bool) []float64 {
	var ticks []float64
	if logScale {
		for v := math.Pow(10, math.Floor(math.Log10(min))); v <= max*1.0001; v *= 10 {
			for _, m := range []float64{1, 2, 5} {
				if val := v * m; val >= min*0.999 && val <= max*1.001 {
					ticks = append(ticks, val)
				}
			}
		}
		return ticks
	}
	step := tickStep((max - min) / 4)
	for v := min; v <= max+step*0.5; v += step {
		ticks = append(ticks, v)
	}
	return ticks
}

func xTicks(min, max float64) []time.Time {
	days := (max - min) / 86400
	var step int
	switch {
	case days <= 7:
		step = 1
	case days <= 16:
		step = 2
	case days <= 40:
		step = 7
	case days <= 100:
		step = 14
	case days <= 220:
		step = 30
	case days <= 420:
		step = 60
	default:
		step = 120
	}
	first := time.Unix(int64(min), 0).UTC()
	day := time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC)
	var ticks []time.Time
	for t := day.AddDate(0, 0, step); float64(t.Unix()) <= max; t = t.AddDate(0, 0, step) {
		ticks = append(ticks, t)
	}
	return ticks
}

// tickStep returns a "nice" tick step from {1,2,2.5,5,10} * 10^n.
func tickStep(x float64) float64 {
	if x <= 0 {
		return 1
	}
	e := math.Pow(10, math.Floor(math.Log10(x)))
	switch r := x / e; {
	case r < 1.5:
		return e
	case r < 2.25:
		return 2 * e
	case r < 3.5:
		return 2.5 * e
	case r < 7.5:
		return 5 * e
	default:
		return 10 * e
	}
}

// niceCeil rounds x up to a "nice" value from {1,1.5,2,2.5,3,4,5,6,8,10}*10^n.
func niceCeil(x float64) float64 {
	if x <= 0 {
		return 1
	}
	e := math.Pow(10, math.Floor(math.Log10(x)))
	set := []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10}
	r := x / e
	for _, s := range set {
		if r <= s {
			return s * e
		}
	}
	return 10 * e
}

// niceFloor rounds x down to a "nice" value.
func niceFloor(x float64) float64 {
	if x <= 0 {
		return 0
	}
	e := math.Pow(10, math.Floor(math.Log10(x)))
	set := []float64{10, 8, 6, 5, 4, 3, 2.5, 2, 1.5, 1}
	r := x / e
	for _, s := range set {
		if r >= s {
			return s * e
		}
	}
	return 0
}

func humanize(v float64) string {
	a := math.Abs(v)
	switch {
	case a >= 1e9:
		return trimSig(v/1e9) + "B"
	case a >= 1e6:
		return trimSig(v/1e6) + "M"
	case a >= 1e3:
		return trimSig(v/1e3) + "K"
	default:
		return trimSig(v)
	}
}

func trimSig(v float64) string {
	return strconv.FormatFloat(v, 'g', 3, 64)
}

func comma(v float64) string {
	s := strconv.FormatInt(int64(math.Round(v)), 10)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(s[i])
	}
	return sign + out.String()
}

func f(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
