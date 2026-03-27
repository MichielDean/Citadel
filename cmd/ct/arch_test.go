package main

import (
	"strings"
	"testing"
)

// TestArchPixelMap_Has12RowsOf36Cols verifies the pixel map has exactly 12 rows of 36 columns.
//
// Given: the static archPixelMap constant
// When:  its dimensions are inspected
// Then:  it has archPillarH rows and archPillarW columns each
func TestArchPixelMap_Has12RowsOf36Cols(t *testing.T) {
	if len(archPixelMap) != archPillarH {
		t.Errorf("archPixelMap has %d rows, want %d", len(archPixelMap), archPillarH)
	}
	for r, row := range archPixelMap {
		if len(row) != archPillarW {
			t.Errorf("archPixelMap[%d] has %d cols, want %d", r, len(row), archPillarW)
		}
	}
}

// TestArchPixelMap_Rows0to3_AreBlank verifies rows 0–3 are entirely blank.
// These rows represent the space above the arch crown.
//
// Given: archPixelMap rows 0–3
// When:  each cell is inspected
// Then:  every cell is pxBlank
func TestArchPixelMap_Rows0to3_AreBlank(t *testing.T) {
	for r := 0; r < 4; r++ {
		for c, px := range archPixelMap[r] {
			if px != pxBlank {
				t.Errorf("archPixelMap[%d][%d] = %q, want pxBlank (space above crown)", r, c, px)
			}
		}
	}
}

// TestArchPixelMap_Row4_IsFullWidthCrown verifies row 4 is full-width fill.
// Row 4 represents the arch crown / road surface.
//
// Given: archPixelMap row 4
// When:  each cell is inspected
// Then:  all 36 cells are pxFill ('▒')
func TestArchPixelMap_Row4_IsFullWidthCrown(t *testing.T) {
	for c, px := range archPixelMap[4] {
		if px != pxFill {
			t.Errorf("archPixelMap[4][%d] = %q, want pxFill '▒' (arch crown)", c, px)
		}
	}
}

// TestArchPixelMap_Row5_ArchOpeningShape verifies row 5 encodes the arch opening.
// Expected layout: 8 blank + 1 edge + 19 fill + 8 blank = 36 cols.
//
// Given: archPixelMap row 5
// When:  each cell group is inspected
// Then:  columns 0–7 blank, column 8 edge, columns 9–27 fill, columns 28–35 blank
func TestArchPixelMap_Row5_ArchOpeningShape(t *testing.T) {
	row := archPixelMap[5]
	for c := 0; c < 8; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[5][%d] = %q, want pxBlank (leading indent)", c, row[c])
		}
	}
	if row[8] != pxEdge {
		t.Errorf("archPixelMap[5][8] = %q, want pxEdge '░' (arch edge)", row[8])
	}
	for c := 9; c <= 27; c++ {
		if row[c] != pxFill {
			t.Errorf("archPixelMap[5][%d] = %q, want pxFill '▒' (arch fill)", c, row[c])
		}
	}
	for c := 28; c <= 35; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[5][%d] = %q, want pxBlank (trailing)", c, row[c])
		}
	}
}

// TestArchPixelMap_Row6_ArchNarrowingShape verifies row 6 encodes a narrower arch opening.
// Expected layout: 12 blank + 1 edge + 11 fill + 12 blank = 36 cols.
//
// Given: archPixelMap row 6
// When:  each cell group is inspected
// Then:  columns 0–11 blank, column 12 edge, columns 13–23 fill, columns 24–35 blank
func TestArchPixelMap_Row6_ArchNarrowingShape(t *testing.T) {
	row := archPixelMap[6]
	for c := 0; c < 12; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[6][%d] = %q, want pxBlank", c, row[c])
		}
	}
	if row[12] != pxEdge {
		t.Errorf("archPixelMap[6][12] = %q, want pxEdge '░'", row[12])
	}
	for c := 13; c <= 23; c++ {
		if row[c] != pxFill {
			t.Errorf("archPixelMap[6][%d] = %q, want pxFill '▒'", c, row[c])
		}
	}
	for c := 24; c <= 35; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[6][%d] = %q, want pxBlank", c, row[c])
		}
	}
}

// TestArchPixelMap_Row7_ArchNarrowestShape verifies row 7 encodes the narrowest arch section.
// Expected layout: 13 blank + 1 edge + 9 fill + 13 blank = 36 cols.
//
// Given: archPixelMap row 7
// When:  each cell group is inspected
// Then:  columns 0–12 blank, column 13 edge, columns 14–22 fill, columns 23–35 blank
func TestArchPixelMap_Row7_ArchNarrowestShape(t *testing.T) {
	row := archPixelMap[7]
	for c := 0; c < 13; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[7][%d] = %q, want pxBlank", c, row[c])
		}
	}
	if row[13] != pxEdge {
		t.Errorf("archPixelMap[7][13] = %q, want pxEdge '░'", row[13])
	}
	for c := 14; c <= 22; c++ {
		if row[c] != pxFill {
			t.Errorf("archPixelMap[7][%d] = %q, want pxFill '▒'", c, row[c])
		}
	}
	for c := 23; c <= 35; c++ {
		if row[c] != pxBlank {
			t.Errorf("archPixelMap[7][%d] = %q, want pxBlank", c, row[c])
		}
	}
}

// TestArchPixelMap_Rows8to11_PierBodyShape verifies all pier body rows have the same shape.
// Expected layout: 15 blank + 1 edge + 5 fill + 15 blank = 36 cols.
//
// Given: archPixelMap rows 8–11
// When:  each row's cells are inspected
// Then:  columns 0–14 blank, column 15 edge, columns 16–20 fill, columns 21–35 blank
func TestArchPixelMap_Rows8to11_PierBodyShape(t *testing.T) {
	for r := 8; r <= 11; r++ {
		row := archPixelMap[r]
		for c := 0; c < 15; c++ {
			if row[c] != pxBlank {
				t.Errorf("archPixelMap[%d][%d] = %q, want pxBlank (pier indent)", r, c, row[c])
			}
		}
		if row[15] != pxEdge {
			t.Errorf("archPixelMap[%d][15] = %q, want pxEdge '░' (pier edge)", r, row[15])
		}
		for c := 16; c <= 20; c++ {
			if row[c] != pxFill {
				t.Errorf("archPixelMap[%d][%d] = %q, want pxFill '▒' (pier fill)", r, c, row[c])
			}
		}
		for c := 21; c <= 35; c++ {
			if row[c] != pxBlank {
				t.Errorf("archPixelMap[%d][%d] = %q, want pxBlank (pier trailing)", r, c, row[c])
			}
		}
	}
}

// TestRenderArchPillarRow_Row4_ContainsFillChar verifies that rendering row 4 produces output
// containing fill characters ('▒').
//
// Given: row 4 of archPixelMap (full-width crown)
// When:  renderArchPillarRow is called with active=false
// Then:  the output contains '▒'
func TestRenderArchPillarRow_Row4_ContainsFillChar(t *testing.T) {
	got := renderArchPillarRow(4, false)
	if !strings.Contains(got, "▒") {
		t.Errorf("renderArchPillarRow(4, false): expected '▒' in output, got %q", got)
	}
}

// TestRenderArchPillarRow_Row4_Active_ContainsFillChar verifies that rendering row 4 with
// active=true also produces fill characters.
//
// Given: row 4 of archPixelMap
// When:  renderArchPillarRow is called with active=true
// Then:  the output contains '▒'
func TestRenderArchPillarRow_Row4_Active_ContainsFillChar(t *testing.T) {
	got := renderArchPillarRow(4, true)
	if !strings.Contains(got, "▒") {
		t.Errorf("renderArchPillarRow(4, true): expected '▒' in output, got %q", got)
	}
}

// TestRenderArchPillarRow_Row5_ContainsEdgeChar verifies that rendering row 5 produces output
// containing an edge character ('░').
//
// Given: row 5 of archPixelMap (arch opening with one edge pixel at col 8)
// When:  renderArchPillarRow is called
// Then:  the output contains '░'
func TestRenderArchPillarRow_Row5_ContainsEdgeChar(t *testing.T) {
	got := renderArchPillarRow(5, false)
	if !strings.Contains(got, "░") {
		t.Errorf("renderArchPillarRow(5, false): expected '░' in output, got %q", got)
	}
}

// TestRenderDroughtPillarRow_Row4_ContainsFillChar verifies that the drought renderer
// produces fill characters for row 4.
//
// Given: row 4 of archPixelMap (full-width crown)
// When:  renderDroughtPillarRow is called
// Then:  the output contains '▒'
func TestRenderDroughtPillarRow_Row4_ContainsFillChar(t *testing.T) {
	got := renderDroughtPillarRow(4)
	if !strings.Contains(got, "▒") {
		t.Errorf("renderDroughtPillarRow(4): expected '▒' in output, got %q", got)
	}
}

// TestRenderDroughtPillarRow_Rows0to3_ProduceNoVisibleChars verifies that drought rendering
// of blank rows produces only whitespace.
//
// Given: rows 0–3 of archPixelMap (blank above arch crown)
// When:  renderDroughtPillarRow is called for each
// Then:  the output contains no non-space characters
func TestRenderDroughtPillarRow_Rows0to3_ProduceNoVisibleChars(t *testing.T) {
	for r := 0; r < 4; r++ {
		got := renderDroughtPillarRow(r)
		if strings.TrimSpace(got) != "" {
			t.Errorf("renderDroughtPillarRow(%d): expected blank output, got %q", r, got)
		}
	}
}

// TestRenderArchPillarRow_Rows0to3_ProduceNoVisibleChars verifies that rendering blank
// rows above the crown produces only whitespace (with background color applied).
//
// Given: rows 0–3 of archPixelMap
// When:  renderArchPillarRow is called for each
// Then:  the output contains no non-space characters after stripping ANSI codes
//
// Note: archRoleBackground sets a black terminal background, so lipgloss emits ANSI
// escape codes in color-enabled environments (e.g. CLICOLOR_FORCE=1). strings.TrimSpace
// cannot strip escape codes, so we use stripANSI first.
func TestRenderArchPillarRow_Rows0to3_ProduceNoVisibleChars(t *testing.T) {
	for r := 0; r < 4; r++ {
		got := stripANSI(renderArchPillarRow(r, false))
		if strings.TrimSpace(got) != "" {
			t.Errorf("renderArchPillarRow(%d, false): expected blank output, got %q", r, got)
		}
	}
}
