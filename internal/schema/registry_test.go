package schema

import "testing"

const testSchemaID = "urn:threadkeeper:test:event-v1"
const testSchema = `{
  "$schema":"https://json-schema.org/draft/2020-12/schema",
  "$id":"urn:threadkeeper:test:event-v1",
  "type":"object",
  "required":["schema_version","event_id"],
  "properties":{
    "schema_version":{"const":"1"},
    "event_id":{"type":"string","minLength":1}
  },
  "additionalProperties":false
}`

func TestDraft2020LocalValidation(t *testing.T) {
	r := NewRegistry()
	if err := r.Add(testSchemaID, []byte(testSchema)); err != nil { t.Fatal(err) }
	if err := r.Validate(testSchemaID, []byte(`{"schema_version":"1","event_id":"E1"}`)); err != nil { t.Fatal(err) }
	if err := r.Validate(testSchemaID, []byte(`{"schema_version":"1"}`)); err == nil { t.Fatal("expected schema validation failure") }
}

func TestRejectsWrongDialect(t *testing.T) {
	r := NewRegistry()
	bad := []byte(`{"$schema":"http://json-schema.org/draft-07/schema","$id":"urn:x","type":"object"}`)
	if err := r.Add("urn:x", bad); err == nil { t.Fatal("expected wrong dialect rejection") }
}

func TestMissingSchemaFailsClosed(t *testing.T) {
	r := NewRegistry()
	if err := r.Validate("urn:missing", []byte(`{}`)); err == nil { t.Fatal("expected unknown schema error") }
}
