// Package beatify provides a Testo plugin that prints colored per-test
// status lines and a suite summary.
package beatify

import (
	"flag"
	"time"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
)

var _ testoplugin.Plugin = (*PluginBeatify)(nil)

var (
	flagDisable      = flag.Bool("beatify.disable", false, "disable ANSI colors in beatify output")
	flagQuiet        = flag.Bool("beatify.quiet", false, "print only the final beatify summary")
	flagCoverageOnly = flag.Bool(
		"beatify.coverage-only",
		false,
		"print only pass-rate coverage: N%",
	)
)

// PluginBeatify prints colored per-test status lines and a suite summary.
type PluginBeatify struct {
	*testo.T

	stats   *runStats
	start   time.Time
	painter painter

	disable      bool
	quiet        bool
	coverageOnly bool
}

// Plugin implements [testoplugin.Plugin].
func (p *PluginBeatify) Plugin(
	parent testoplugin.Plugin,
	_ ...testoplugin.Option,
) testoplugin.Spec {
	return p.plugin(parent.(*PluginBeatify))
}

func (p *PluginBeatify) plugin(parent *PluginBeatify) testoplugin.Spec {
	p.inherit(parent)

	p.disable = p.disable || *flagDisable
	p.quiet = p.quiet || *flagQuiet
	p.coverageOnly = p.coverageOnly || *flagCoverageOnly
	p.painter = newPainter(!p.disable)

	if p.coverageOnly {
		silenceTestOutput()
	}

	return testoplugin.Spec{
		Hooks: testoplugin.Hooks{
			BeforeAll: testoplugin.Hook{
				Priority: testoplugin.TryFirst,
				Func:     p.beforeAll,
			},
			BeforeEach: testoplugin.Hook{
				Priority: testoplugin.TryFirst,
				Func:     p.beforeEach,
			},
			AfterEach: testoplugin.Hook{
				Priority: testoplugin.TryLast,
				Func:     p.afterEach,
			},
			AfterAll: testoplugin.Hook{
				Priority: testoplugin.TryLast,
				Func:     p.afterAll,
			},
		},
	}
}

func (p *PluginBeatify) inherit(parent *PluginBeatify) {
	if parent == nil {
		p.stats = &runStats{}

		return
	}

	p.stats = parent.stats
	p.disable = parent.disable
	p.quiet = parent.quiet
	p.coverageOnly = parent.coverageOnly
}

func (p *PluginBeatify) beforeAll() {
	p.Helper()
	p.stats.markStart(time.Now())
}

func (p *PluginBeatify) beforeEach() {
	p.Helper()

	p.start = time.Now()
}

func (p *PluginBeatify) afterEach() {
	p.Helper()

	result := testResult{
		Name:     p.Name(),
		Duration: time.Since(p.start),
		Status:   statusPass,
		Loc:      testLocation(p.T),
	}

	switch {
	case p.Failed():
		result.Status = statusFail
	case p.Skipped():
		result.Status = statusSkip
	}

	p.stats.add(result)

	if !p.coverageOnly {
		p.printTestLine(result)
	}
}

func (p *PluginBeatify) afterAll() {
	p.Helper()

	if p.coverageOnly {
		p.printCoverageOnly()

		return
	}

	p.printSummary()
}
