package diagram

import (
	"fmt"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

const MinDiagramWidth = 60

func PlatformFlow(platforms []platform.Platform) string {
	var b strings.Builder
	b.WriteString("  \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510     \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2510     \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510\n")
	b.WriteString("  \u2502 Platforms \u2502\u2500\u2500\u2500\u2502Build\u2502\u2500\u2500\u2500\u2502 Hydrate \u2502\n")
	b.WriteString("  \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518     \u2514\u2500\u2500\u2500\u2500\u2500\u2518     \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518\n")
	if len(platforms) > 0 {
		b.WriteString(fmt.Sprintf("  (%d platforms)\n", len(platforms)))
	}
	return b.String()
}

func CacheTopology(activeTab int) string {
	var b strings.Builder
	b.WriteString("  \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510     \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510\n")
	b.WriteString("  \u2502 build-unit \u2502\u2500\u2500\u2500\u2502 platform \u2502\n")
	b.WriteString("  \u2502   cache    \u2502\u2500\u2500\u2500\u2502  cache   \u2502\n")
	b.WriteString("  \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518     \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518\n")
	tabName := "build-unit"
	if activeTab == 1 {
		tabName = "platform"
	}
	b.WriteString(fmt.Sprintf("  Tab: %s\n", tabName))
	return b.String()
}

func ValidationPipeline() string {
	var b strings.Builder
	b.WriteString("  \u250c\u2500\u2500\u2500\u2510   \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510   \u250c\u2500\u2500\u2500\u2500\u2510   \u250c\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2510\n")
	b.WriteString("  \u2502all\u2502\u2500\u2500\u2502shellcheck\u2502\u2500\u2502bats\u2502\u2500\u2502parse-test\u2502\n")
	b.WriteString("  \u2514\u2500\u2500\u2500\u2518   \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518   \u2514\u2500\u2500\u2500\u2518   \u2514\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2518\n")
	return b.String()
}
