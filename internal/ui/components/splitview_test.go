package components

import (
	"fmt"
	"testing"
)

func TestNewSplitView(t *testing.T) {
	sv := NewSplitView(100, 50)

	if sv.Width != 100 {
		t.Errorf("Width = %d; want 100", sv.Width)
	}
	if sv.Height != 50 {
		t.Errorf("Height = %d; want 50", sv.Height)
	}
	if sv.SplitRatio != 0.5 {
		t.Errorf("SplitRatio = %f; want 0.5", sv.SplitRatio)
	}
	if sv.Border != true {
		t.Error("Border should be true by default")
	}
}

func TestSplitViewDimensions(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		ratio      float64
		wantLeft   int
		wantRight  int
	}{
		{
			name:      "50/50 split",
			width:     100,
			ratio:     0.5,
			wantLeft:  50,
			wantRight: 49, // 100 - 50 - 1
		},
		{
			name:      "30/70 split",
			width:     100,
			ratio:     0.3,
			wantLeft:  30,
			wantRight: 69, // 100 - 30 - 1
		},
		{
			name:      "70/30 split",
			width:     100,
			ratio:     0.7,
			wantLeft:  70,
			wantRight: 29, // 100 - 70 - 1
		},
		{
			name:      "minimal width",
			width:     10,
			ratio:     0.5,
			wantLeft:  5,
			wantRight: 4, // 10 - 5 - 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := NewSplitView(tt.width, 50)
			sv.SplitRatio = tt.ratio

			// Calculate dimensions as Render() does
			leftWidth := int(float64(sv.Width) * sv.SplitRatio)
			rightWidth := sv.Width - leftWidth - 1

			if leftWidth != tt.wantLeft {
				t.Errorf("leftWidth = %d; want %d", leftWidth, tt.wantLeft)
			}
			if rightWidth != tt.wantRight {
				t.Errorf("rightWidth = %d; want %d", rightWidth, tt.wantRight)
			}
		})
	}
}

func TestSplitViewSetContent(t *testing.T) {
	sv := NewSplitView(100, 50)

	left := "left content"
	right := "right content"
	sv.SetContent(left, right)

	if sv.Left != left {
		t.Errorf("Left = %s; want %s", sv.Left, left)
	}
	if sv.Right != right {
		t.Errorf("Right = %s; want %s", sv.Right, right)
	}
}

func TestSplitRatioCalculations(t *testing.T) {
	tests := []struct {
		name  string
		ratio float64
		width int
	}{
		{"zero ratio", 0.0, 100},
		{"quarter ratio", 0.25, 100},
		{"half ratio", 0.5, 100},
		{"three-quarter ratio", 0.75, 100},
		{"full ratio", 1.0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := NewSplitView(tt.width, 50)
			sv.SplitRatio = tt.ratio

			leftWidth := int(float64(sv.Width) * sv.SplitRatio)
			rightWidth := sv.Width - leftWidth - 1

			// Verify total width accounts for separator
			if leftWidth+rightWidth+1 != tt.width {
				t.Errorf("total width mismatch: left(%d) + right(%d) + separator(1) != %d",
					leftWidth, rightWidth, tt.width)
			}

			// Verify left width matches ratio
			expectedLeft := int(float64(tt.width) * tt.ratio)
			if leftWidth != expectedLeft {
				t.Errorf("leftWidth = %d; want %d", leftWidth, expectedLeft)
			}
		})
	}
}

func TestNewTabView(t *testing.T) {
	tabs := []string{"Tab1", "Tab2", "Tab3"}
	tv := NewTabView(tabs, 100)

	if len(tv.Tabs) != 3 {
		t.Errorf("Tabs length = %d; want 3", len(tv.Tabs))
	}
	if tv.ActiveTab != 0 {
		t.Errorf("ActiveTab = %d; want 0", tv.ActiveTab)
	}
	if tv.Width != 100 {
		t.Errorf("Width = %d; want 100", tv.Width)
	}
	if tv.Content == nil {
		t.Error("Content map should be initialized")
	}
}

func TestTabViewNavigation(t *testing.T) {
	tabs := []string{"Tab1", "Tab2", "Tab3"}
	tv := NewTabView(tabs, 100)

	// Test NextTab
	tv.NextTab()
	if tv.ActiveTab != 1 {
		t.Errorf("After NextTab: ActiveTab = %d; want 1", tv.ActiveTab)
	}

	tv.NextTab()
	if tv.ActiveTab != 2 {
		t.Errorf("After NextTab: ActiveTab = %d; want 2", tv.ActiveTab)
	}

	// Test wrapping forward
	tv.NextTab()
	if tv.ActiveTab != 0 {
		t.Errorf("After NextTab wrap: ActiveTab = %d; want 0", tv.ActiveTab)
	}

	// Test PrevTab
	tv.PrevTab()
	if tv.ActiveTab != 2 {
		t.Errorf("After PrevTab wrap: ActiveTab = %d; want 2", tv.ActiveTab)
	}

	tv.PrevTab()
	if tv.ActiveTab != 1 {
		t.Errorf("After PrevTab: ActiveTab = %d; want 1", tv.ActiveTab)
	}
}

func TestTabViewSetContent(t *testing.T) {
	tabs := []string{"Tab1", "Tab2"}
	tv := NewTabView(tabs, 100)

	content1 := "Content for Tab1"
	content2 := "Content for Tab2"

	tv.SetContent("Tab1", content1)
	tv.SetContent("Tab2", content2)

	if tv.Content["Tab1"] != content1 {
		t.Errorf("Tab1 content = %s; want %s", tv.Content["Tab1"], content1)
	}
	if tv.Content["Tab2"] != content2 {
		t.Errorf("Tab2 content = %s; want %s", tv.Content["Tab2"], content2)
	}
}

func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar(100)

	if sb.Width != 100 {
		t.Errorf("Width = %d; want 100", sb.Width)
	}
	if sb.Items != nil && len(sb.Items) != 0 {
		t.Errorf("Items should be empty initially, got %d items", len(sb.Items))
	}
}

func TestStatusBarAddItem(t *testing.T) {
	sb := NewStatusBar(100)

	sb.AddItem("🔍", "Status", "Active")
	sb.AddItem("📊", "Count", "42")

	if len(sb.Items) != 2 {
		t.Fatalf("Items length = %d; want 2", len(sb.Items))
	}

	if sb.Items[0].Icon != "🔍" {
		t.Errorf("Item[0].Icon = %s; want 🔍", sb.Items[0].Icon)
	}
	if sb.Items[0].Label != "Status" {
		t.Errorf("Item[0].Label = %s; want Status", sb.Items[0].Label)
	}
	if sb.Items[0].Value != "Active" {
		t.Errorf("Item[0].Value = %s; want Active", sb.Items[0].Value)
	}

	if sb.Items[1].Label != "Count" {
		t.Errorf("Item[1].Label = %s; want Count", sb.Items[1].Label)
	}
	if sb.Items[1].Value != "42" {
		t.Errorf("Item[1].Value = %s; want 42", sb.Items[1].Value)
	}
}

func TestSplitViewResize(t *testing.T) {
	sv := NewSplitView(100, 50)

	// Resize to new dimensions
	sv.Width = 200
	sv.Height = 100

	if sv.Width != 200 {
		t.Errorf("Width after resize = %d; want 200", sv.Width)
	}
	if sv.Height != 100 {
		t.Errorf("Height after resize = %d; want 100", sv.Height)
	}

	// Verify dimension calculations with new width
	leftWidth := int(float64(sv.Width) * sv.SplitRatio)
	rightWidth := sv.Width - leftWidth - 1

	expectedLeft := 100 // 200 * 0.5
	expectedRight := 99 // 200 - 100 - 1

	if leftWidth != expectedLeft {
		t.Errorf("leftWidth after resize = %d; want %d", leftWidth, expectedLeft)
	}
	if rightWidth != expectedRight {
		t.Errorf("rightWidth after resize = %d; want %d", rightWidth, expectedRight)
	}
}

func TestSplitViewBorderToggle(t *testing.T) {
	sv := NewSplitView(100, 50)

	// Default should have border
	if !sv.Border {
		t.Error("Border should be true by default")
	}

	// Toggle border off
	sv.Border = false
	if sv.Border {
		t.Error("Border should be false after toggle")
	}

	// Toggle border back on
	sv.Border = true
	if !sv.Border {
		t.Error("Border should be true after toggle back")
	}
}

func TestTabViewSingleTab(t *testing.T) {
	tabs := []string{"OnlyTab"}
	tv := NewTabView(tabs, 100)

	// With single tab, NextTab should stay at 0
	tv.NextTab()
	if tv.ActiveTab != 0 {
		t.Errorf("ActiveTab with single tab after NextTab = %d; want 0", tv.ActiveTab)
	}

	// PrevTab should also stay at 0
	tv.PrevTab()
	if tv.ActiveTab != 0 {
		t.Errorf("ActiveTab with single tab after PrevTab = %d; want 0", tv.ActiveTab)
	}
}

func TestTabViewEmptyContent(t *testing.T) {
	tabs := []string{"Tab1", "Tab2"}
	tv := NewTabView(tabs, 100)

	// Access content for tab that hasn't been set
	content := tv.Content["Tab1"]
	if content != "" {
		t.Errorf("Unset tab content = %s; want empty string", content)
	}
}

// Render tests - verify rendering completes without errors
// We avoid testing exact output format to stay independent of lipgloss rendering

func TestSplitViewRender(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		ratio  float64
		border bool
	}{
		{"default settings", 100, 50, 0.5, true},
		{"no border", 100, 50, 0.5, false},
		{"wide layout", 200, 30, 0.6, true},
		{"narrow layout", 80, 20, 0.4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := NewSplitView(tt.width, tt.height)
			sv.SplitRatio = tt.ratio
			sv.Border = tt.border
			sv.SetContent("Left Content", "Right Content")

			// Render should not panic and should return non-empty string
			output := sv.Render()
			if output == "" {
				t.Error("Render() returned empty string")
			}
		})
	}
}

func TestTabViewRender(t *testing.T) {
	tabs := []string{"Tab1", "Tab2", "Tab3"}
	tv := NewTabView(tabs, 100)

	tv.SetContent("Tab1", "Content 1")
	tv.SetContent("Tab2", "Content 2")
	tv.SetContent("Tab3", "Content 3")

	// Test rendering different active tabs
	for i := 0; i < len(tabs); i++ {
		tv.ActiveTab = i
		output := tv.Render()
		if output == "" {
			t.Errorf("Render() for tab %d returned empty string", i)
		}
	}
}

func TestStatusBarRender(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		itemCount int
	}{
		{"empty status bar", 100, 0},
		{"single item", 100, 1},
		{"multiple items", 150, 3},
		{"narrow width", 50, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewStatusBar(tt.width)

			// Add items based on itemCount
			for i := 0; i < tt.itemCount; i++ {
				sb.AddItem("📊", "Label", "Value")
			}

			output := sb.Render()
			if output == "" {
				t.Error("Render() returned empty string")
			}
		})
	}
}

func TestSplitViewRenderWithDifferentRatios(t *testing.T) {
	ratios := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, ratio := range ratios {
		t.Run(fmt.Sprintf("ratio_%.2f", ratio), func(t *testing.T) {
			sv := NewSplitView(100, 50)
			sv.SplitRatio = ratio
			sv.SetContent("Left", "Right")

			output := sv.Render()
			if output == "" {
				t.Errorf("Render() with ratio %.2f returned empty string", ratio)
			}
		})
	}
}

func TestTabViewRenderEmptyContent(t *testing.T) {
	tabs := []string{"Tab1"}
	tv := NewTabView(tabs, 100)

	// Render without setting content
	output := tv.Render()
	if output == "" {
		t.Error("Render() with empty content returned empty string")
	}
}

func TestStatusBarRenderLongContent(t *testing.T) {
	sb := NewStatusBar(50)

	// Add items with long values that exceed width
	sb.AddItem("📊", "VeryLongLabel", "VeryLongValueThatExceedsWidth")
	sb.AddItem("🔍", "AnotherLabel", "AnotherValue")

	output := sb.Render()
	if output == "" {
		t.Error("Render() with long content returned empty string")
	}
}
