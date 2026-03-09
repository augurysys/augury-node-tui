package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

const metricsBarBlocks = 5

// MetricsBar displays real-time resource usage (CPU, Memory, Disk) with a hot process indicator.
type MetricsBar struct {
	CPU        float64 // 0.0 to 1.0
	Memory     float64
	Disk       float64
	HotProcess string // Most active process name
	Width      int
}

// Render produces the styled metrics bar output.
// Format: CPU: ████░ 82%  MEM: ███░░ 65%  DISK: ██░░░ 48%  Hot: gcc (3 threads)
// Progress bars use sequential color mapping: <50% dim, 50-80% yellow, >80% red.
func (m MetricsBar) Render() string {
	palette := styles.DefaultPalette()
	typo := styles.DefaultTypography()

	// Clamp values to 0-1 for display
	cpu := clamp01(m.CPU)
	mem := clamp01(m.Memory)
	disk := clamp01(m.Disk)

	cpuPart := renderMetricPart("CPU", cpu, palette, typo)
	memPart := renderMetricPart("MEM", mem, palette, typo)
	diskPart := renderMetricPart("DISK", disk, palette, typo)

	hotLabel := typo.Body.Render("Hot: ")
	hotValue := typo.Dim.Render(m.HotProcess)
	if m.HotProcess == "" {
		hotValue = typo.Dim.Render("-")
	}
	hotPart := hotLabel + hotValue

	line := cpuPart + "  " + memPart + "  " + diskPart + "  " + hotPart

	if m.Width > 0 {
		line = truncateLineToWidth(line, m.Width)
	}
	return line
}

// FetchMetrics updates the MetricsBar fields with current system metrics using gopsutil.
func (m *MetricsBar) FetchMetrics() error {
	// Fetch CPU (average across all cores, 100ms sample)
	cpuPercents, err := cpu.Percent(100*time.Millisecond, false)
	if err == nil && len(cpuPercents) > 0 {
		m.CPU = clamp01(cpuPercents[0] / 100.0)
	}

	// Fetch Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		m.Memory = clamp01(vmStat.UsedPercent / 100.0)
	}

	// Fetch Disk (root partition "/")
	diskStat, err := disk.Usage("/")
	if err == nil {
		m.Disk = clamp01(diskStat.UsedPercent / 100.0)
	}

	// Fetch hot process (most CPU-intensive)
	procs, err := process.Processes()
	if err == nil {
		var maxCPU float64
		var hotProc *process.Process
		for _, p := range procs {
			cpuPct, err := p.CPUPercent()
			if err == nil && cpuPct > maxCPU {
				maxCPU = cpuPct
				hotProc = p
			}
		}
		if hotProc != nil {
			name, _ := hotProc.Name()
			numThreads, _ := hotProc.NumThreads()
			m.HotProcess = fmt.Sprintf("%s (%d threads)", name, numThreads)
		}
	}

	return nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func renderMetricPart(label string, value float64, palette styles.Palette, typo styles.Typography) string {
	pct := value * 100

	barW := metricsBarBlocks
	filledBlocks := int(float64(barW) * value)
	if filledBlocks < 0 {
		filledBlocks = 0
	}
	if filledBlocks > barW {
		filledBlocks = barW
	}

	filled := strings.Repeat("█", filledBlocks)
	unfilled := strings.Repeat("░", barW-filledBlocks)

	// Sequential color: <50% dim, 50-80% yellow, >80% red
	var barColor string
	if pct < 50 {
		barColor = palette.Overlay0
	} else if pct < 80 {
		barColor = palette.Warning
	} else {
		barColor = palette.Error
	}

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor))
	labelPart := typo.Body.Render(label + ": ")
	barPart := barStyle.Render(filled + unfilled)
	pctPart := typo.Body.Render(fmt.Sprintf(" %.0f%%", pct))

	return labelPart + barPart + pctPart
}
