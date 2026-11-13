// Package parallel provides a Testo plugin that marks tests as parallel by default.
//
// Embed [PluginParallel] in a custom Testo T. Use [WithScope] to select the
// execution levels and [WithSync] for synchronous exceptions.
//
// [WithBeforeEachPriority] and [WithBeforeAllPriority] independently control
// when the plugin pauses execution relative to other plugins' setup hooks.
// Both default to zero. Earlier priorities can defer resource allocation
// until execution resumes, but do not impose a global resource limit.
package parallel
