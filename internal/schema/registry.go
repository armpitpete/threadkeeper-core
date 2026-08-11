package schema

import (
	"fmt"
	"sync"

	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const Draft2020URL = "https://json-schema.org/draft/2020-12/schema"

type localLoader struct {
	resources map[string]any
}

func (l localLoader) Load(url string) (any, error) {
	v, ok := l.resources[url]
	if !ok {
		return nil, fmt.Errorf("network/schema loader disabled: resource %q is not registered locally", url)
	}
	return v, nil
}

type Registry struct {
	mu        sync.RWMutex
	resources map[string]any
	compiled  map[string]*jsonschema.Schema
}

func NewRegistry() *Registry {
	return &Registry{resources: make(map[string]any), compiled: make(map[string]*jsonschema.Schema)}
}

func (r *Registry) Add(id string, raw []byte) error {
	if id == "" {
		return fmt.Errorf("schema id is required")
	}
	v, err := strictjson.Decode(raw)
	if err != nil {
		return err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("schema %q root must be an object", id)
	}
	if got, _ := obj["$schema"].(string); got != Draft2020URL {
		return fmt.Errorf("SCHEMA_INVALID: schema %q must declare %q", id, Draft2020URL)
	}
	if got, _ := obj["$id"].(string); got != id {
		return fmt.Errorf("SCHEMA_INVALID: schema $id %q does not match registered id %q", got, id)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.resources[id]; exists {
		return fmt.Errorf("schema %q already registered", id)
	}
	r.resources[id] = v
	delete(r.compiled, id)
	return nil
}

func (r *Registry) Compile(id string) (*jsonschema.Schema, error) {
	r.mu.RLock()
	if sch := r.compiled[id]; sch != nil {
		r.mu.RUnlock()
		return sch, nil
	}
	resources := cloneResources(r.resources)
	r.mu.RUnlock()

	if _, ok := resources[id]; !ok {
		return nil, fmt.Errorf("UNKNOWN_SCHEMA: %q", id)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseLoader(localLoader{resources: resources})
	for resourceID, doc := range resources {
		if err := c.AddResource(resourceID, doc); err != nil {
			return nil, fmt.Errorf("add local schema resource %q: %w", resourceID, err)
		}
	}
	sch, err := c.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("SCHEMA_INVALID: compile %q: %w", id, err)
	}

	r.mu.Lock()
	if existing := r.compiled[id]; existing != nil {
		sch = existing
	} else {
		r.compiled[id] = sch
	}
	r.mu.Unlock()
	return sch, nil
}

func (r *Registry) Validate(id string, raw []byte) error {
	v, err := strictjson.Decode(raw)
	if err != nil {
		return err
	}
	sch, err := r.Compile(id)
	if err != nil {
		return err
	}
	if err := sch.Validate(v); err != nil {
		return fmt.Errorf("SCHEMA_VALIDATION_FAILED: %w", err)
	}
	return nil
}

func cloneResources(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
