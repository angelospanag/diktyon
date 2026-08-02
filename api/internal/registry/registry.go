package registry

import (
	"errors"
	"sync"
)

// ErrUnknownCountry is returned when Get is called with an unregistered country code.
var ErrUnknownCountry = errors.New("unknown country code")

// Registry maps ISO country codes (e.g. "uk", "gr") to their Provider implementations.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func New() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds or replaces the provider for the given country code.
func (r *Registry) Register(countryCode string, p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[countryCode] = p
}

// Get retrieves the provider for countryCode, or returns ErrUnknownCountry.
func (r *Registry) Get(countryCode string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[countryCode]
	if !ok {
		return nil, ErrUnknownCountry
	}
	return p, nil
}
