package plugin

import "sync"

var (
	mu      sync.RWMutex
	plugins []LinePlugin
)

// Register adds a plugin to the global registry.
// Typically called from a plugin package's init() function.
func Register(p LinePlugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins = append(plugins, p)
}

// All returns all registered plugins.
func All() []LinePlugin {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]LinePlugin, len(plugins))
	copy(out, plugins)
	return out
}

// FindHandlers returns all plugins that can handle the given line.
func FindHandlers(line string) []LinePlugin {
	mu.RLock()
	defer mu.RUnlock()

	var handlers []LinePlugin
	for _, p := range plugins {
		if p.CanHandle(line) {
			handlers = append(handlers, p)
		}
	}
	return handlers
}
