p, err := plot.New()
    if err != nil {
        panic(err)
    }
    p.Title.Text = "Points Example"
    p.X.Label.Text = "X"
    p.Y.Label.Text = "Y"
    // Draw a grid behind the data
    p.Add(plotter.NewGrid())

    // Make a scatter plotter and set its style.
    s, err := plotter.NewScatter(scatterData)
    if err != nil {
        panic(err)
    }
    s.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}

    // Make a line plotter and set its style.
    l, err := plotter.NewLine(lineData)
    if err != nil {
        panic(err)
    }
    l.LineStyle.Width = vg.Points(1)
    l.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
    l.LineStyle.Color = color.RGBA{B: 255, A: 255}

    // Make a line plotter with points and set its style.
    lpLine, lpPoints, err := plotter.NewLinePoints(linePointsData)
    if err != nil {
        panic(err)
    }
    lpLine.Color = color.RGBA{G: 255, A: 255}
    lpPoints.Shape = draw.PyramidGlyph{}
    lpPoints.Color = color.RGBA{R: 255, A: 255}

    // Add the plotters to the plot, with a legend
    // entry for each
    p.Add(s, l, lpLine, lpPoints)
    p.Legend.Add("scatter", s)
    p.Legend.Add("line", l)
    p.Legend.Add("line points", lpLine, lpPoints)

    // Save the plot to a PNG file.
    if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
        panic(err)
    }