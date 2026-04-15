//go:build ignore
// =============================================================================
// LESSON 8.1: SCHEMA EVOLUTION — Taming Data Contracts at Scale
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Why raw JSON in Kafka leads to pain (no schema, no validation, bloated)
// - Serialization showdown: Avro vs Protobuf vs JSON Schema
// - Schema Registry: wire format, caching, how it works under the hood
// - Compatibility modes: BACKWARD, FORWARD, FULL, and TRANSITIVE variants
// - Evolution rules: which changes are safe, which are breaking
// - Migration strategies: when schemas must change incompatibly
//
// THE KEY INSIGHT:
// Producers and consumers are decoupled in Kafka. They deploy independently.
// Without schema enforcement, Producer v2 can silently break Consumer v1.
// Schema Registry + compatibility rules = a safety net that lets teams evolve
// independently without breaking each other.
//
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== SCHEMA EVOLUTION ===")
	fmt.Println()

	whyNotRawJSON()
	serializationFormats()
	schemaRegistryInternals()
	compatibilityModes()
	evolutionRules()
	migrationStrategies()
}

// =============================================================================
// PART 1: WHY NOT RAW JSON? — The pain of schema-less data
// =============================================================================
func whyNotRawJSON() {
	fmt.Println("--- WHY NOT RAW JSON ---")

	// Most teams start with JSON in Kafka. It works... until it doesn't.
	//
	// PROBLEMS WITH RAW JSON:
	//
	// 1. NO SCHEMA ENFORCEMENT
	//    Producer sends: {"user_id": 123, "name": "alice"}
	//    Next week: {"userId": 123, "name": "alice"}  ← field renamed!
	//    Consumer breaks silently — reads null for user_id.
	//
	// 2. FIELD TYPE CHANGES
	//    v1: {"amount": 100}       (integer)
	//    v2: {"amount": "100.50"}  (string)
	//    Consumer's JSON parser silently converts or throws at runtime.
	//
	// 3. SIZE
	//    JSON repeats field names in EVERY record.
	//    {"user_id":123,"event_type":"click","timestamp":1700000000}
	//    = 61 bytes for ~20 bytes of actual data (67% overhead)
	//    At 1M records/sec, that's 40 MB/sec wasted on field names.
	//
	// 4. PARSE SPEED
	//    JSON parsing: ~200 MB/sec (optimized)
	//    Avro decoding: ~1-2 GB/sec
	//    Protobuf decoding: ~1-2 GB/sec
	//    At scale, this difference matters.
	//
	// 5. NO EVOLUTION TRACKING
	//    Which version of the schema is this record?
	//    Who changed it? When? Is it compatible with the old version?
	//    With raw JSON: ¯\_(ツ)_/¯

	fmt.Println("  Raw JSON: no schema, bloated, slow, no versioning")
	fmt.Println("  At scale: 60%+ overhead, silent breakage, unmaintainable")
	fmt.Println()
}

// =============================================================================
// PART 2: SERIALIZATION FORMATS — Avro vs Protobuf vs JSON Schema
// =============================================================================
func serializationFormats() {
	fmt.Println("--- SERIALIZATION FORMATS ---")

	// ┌──────────────┬──────────────┬──────────────┬──────────────────┐
	// │ Feature      │ Avro         │ Protobuf     │ JSON Schema      │
	// ├──────────────┼──────────────┼──────────────┼──────────────────┤
	// │ Schema def   │ .avsc (JSON) │ .proto       │ JSON Schema spec │
	// │ Encoding     │ Binary       │ Binary       │ JSON text        │
	// │ Size         │ Smallest     │ Small        │ Largest          │
	// │ Speed        │ Fast         │ Fastest      │ Slow             │
	// │ Schema in    │ No (separate)│ No (separate)│ Yes (self-desc)  │
	// │  payload     │              │              │                  │
	// │ Evolution    │ Excellent    │ Good         │ Good             │
	// │ Ecosystem    │ Kafka-native │ gRPC-native  │ REST-native      │
	// │ Default vals │ Yes (rich)   │ Yes (zeros)  │ Yes              │
	// │ Union types  │ Yes          │ oneof        │ oneOf            │
	// │ Maps         │ Yes          │ Yes          │ Yes              │
	// │ Required     │ No* (removed)│ No* (proto3) │ Yes              │
	// │ Codegen      │ Optional     │ Required     │ Optional         │
	// └──────────────┴──────────────┴──────────────┴──────────────────┘
	//
	// WHEN TO USE WHAT:
	//
	// AVRO: Best for Kafka-centric architectures.
	//   - Schema stored separately in Schema Registry
	//   - Payload is pure data (no field names, no tags)
	//   - Excellent evolution: readers use their own schema to decode
	//   - Native support in Kafka Connect, KSQL, Kafka Streams
	//   - Default choice for most Kafka deployments
	//
	// PROTOBUF: Best when you also use gRPC or already have .proto files.
	//   - Slightly larger than Avro (field tags in payload)
	//   - Faster encoding/decoding
	//   - Strong typed codegen (Go, Java, Python, etc.)
	//   - Schema Registry supports Protobuf since Confluent Platform 5.5
	//   - Good choice for polyglot microservices
	//
	// JSON SCHEMA: Best for gradual migration from raw JSON.
	//   - Still JSON on the wire (human-readable)
	//   - Schema validation at producer/consumer
	//   - No size benefit, but adds type safety
	//   - Good stepping stone towards Avro/Protobuf

	fmt.Println("  Avro: smallest, Kafka-native, best evolution (DEFAULT CHOICE)")
	fmt.Println("  Protobuf: fast, great codegen, good with gRPC")
	fmt.Println("  JSON Schema: human-readable, good migration path from raw JSON")
	fmt.Println()
}

// =============================================================================
// PART 3: SCHEMA REGISTRY INTERNALS
// =============================================================================
func schemaRegistryInternals() {
	fmt.Println("--- SCHEMA REGISTRY INTERNALS ---")

	// Schema Registry is a separate service that stores and serves schemas.
	// It uses Kafka itself as its storage backend (_schemas topic).
	//
	// WIRE FORMAT:
	// ────────────
	// Every message produced with Schema Registry has this prefix:
	//
	// ┌──────────────────────────────────────────────────────────────┐
	// │  Byte 0     │ Bytes 1-4           │ Bytes 5+                │
	// │  Magic byte │ Schema ID (4 bytes) │ Avro/Protobuf payload   │
	// │  = 0x00     │ big-endian int32    │ binary encoded data     │
	// └──────────────────────────────────────────────────────────────┘
	//
	// Only 5 bytes of overhead! Much less than embedding the schema.
	//
	// The Schema ID is a globally unique identifier assigned by the Registry.
	// Consumer reads the ID → fetches schema from Registry → decodes payload.
	//
	// CACHING:
	// ────────
	// Producers and consumers cache schemas locally:
	// - Producer: schema → ID cache (avoid re-registration on every send)
	// - Consumer: ID → schema cache (avoid fetch on every message)
	//
	// After the first message, there are ZERO network calls to Schema Registry.
	// The Registry could go down and existing producers/consumers keep working!
	// (New schemas can't be registered until it's back.)
	//
	// SCHEMA SUBJECTS:
	// ────────────────
	// A "subject" is the scope for compatibility checking.
	// Default naming strategy: TopicNameStrategy
	//   Subject = "<topic-name>-key" or "<topic-name>-value"
	//   Example: "orders-value" → all schemas for the 'orders' topic value
	//
	// Alternative: RecordNameStrategy
	//   Subject = fully qualified record name (e.g., "com.example.OrderEvent")
	//   Allows multiple schema types in the same topic
	//   Used in event-driven architectures with polymorphic topics
	//
	// STORAGE:
	// ────────
	// Schemas are stored in the _schemas topic (compacted).
	// Key: {"subject":"orders-value","version":1}
	// Value: {"schema":"{...}", "schemaType":"AVRO", "id":42}
	// The Registry rebuilds its in-memory state from this topic on startup.

	fmt.Println("  Wire format: 1 magic byte + 4 byte schema ID + payload")
	fmt.Println("  Only 5 bytes overhead — schemas cached after first use")
	fmt.Println("  Registry backed by _schemas compacted topic")
	fmt.Println()
}

// =============================================================================
// PART 4: COMPATIBILITY MODES
// =============================================================================
func compatibilityModes() {
	fmt.Println("--- COMPATIBILITY MODES ---")

	// Compatibility = can Consumer with Schema v2 read data written with Schema v1?
	// (or vice versa)
	//
	// ┌──────────────────────────────────────────────────────────────────────┐
	// │  MODE             │ CHECK                │ USE CASE                  │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  BACKWARD         │ New schema can read   │ Consumers upgrade first  │
	// │                   │ data from old schema  │ then producers           │
	// │                   │                       │ DEFAULT MODE             │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  FORWARD          │ Old schema can read   │ Producers upgrade first  │
	// │                   │ data from new schema  │ then consumers           │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  FULL             │ Both BACKWARD and     │ Any upgrade order works  │
	// │                   │ FORWARD compatible    │ (most restrictive)       │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  BACKWARD_        │ BACKWARD with ALL     │ Can read ALL old data    │
	// │  TRANSITIVE       │ previous versions     │ (not just previous one)  │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  FORWARD_         │ FORWARD with ALL      │ All old consumers work   │
	// │  TRANSITIVE       │ previous versions     │                          │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  FULL_            │ Both BACKWARD and     │ Maximum safety           │
	// │  TRANSITIVE       │ FORWARD with ALL prev │ Recommended for critical │
	// ├──────────────────────────────────────────────────────────────────────┤
	// │  NONE             │ No checking           │ Dev/test only            │
	// │                   │                       │ NEVER in production      │
	// └──────────────────────────────────────────────────────────────────────┘
	//
	// BACKWARD (default and most common):
	// ───────────────────────────────────
	// "New consumers can read old data"
	// You upgrade consumers first (they can handle both old and new format),
	// then upgrade producers.
	//
	// In Avro, BACKWARD-compatible changes:
	//   ✓ Add a field WITH a default value
	//   ✓ Remove a field
	//   ✗ Add a field WITHOUT a default → reader can't decode old data
	//   ✗ Change field type → incompatible
	//
	// FULL_TRANSITIVE (most conservative):
	// ─────────────────────────────────────
	// Every version is both backward AND forward compatible with ALL versions.
	// Use for critical data where you might need to reprocess old topics.
	//
	// RECOMMENDATION:
	// ───────────────
	// Start with BACKWARD (default). Move to FULL_TRANSITIVE for critical topics.
	// NEVER use NONE in production.

	fmt.Println("  BACKWARD (default): new consumers can read old data")
	fmt.Println("  FULL_TRANSITIVE: maximum safety (both directions, all versions)")
	fmt.Println("  NONE: development only, never production")
	fmt.Println()
}

// =============================================================================
// PART 5: EVOLUTION RULES — What's safe, what breaks
// =============================================================================
func evolutionRules() {
	fmt.Println("--- EVOLUTION RULES ---")

	// AVRO EVOLUTION RULES:
	// ─────────────────────
	// ┌─────────────────────────────┬──────────┬─────────┬──────────┐
	// │ Change                      │ BACKWARD │ FORWARD │ FULL     │
	// ├─────────────────────────────┼──────────┼─────────┼──────────┤
	// │ Add field with default      │ ✓        │ ✓       │ ✓        │
	// │ Add field without default   │ ✗        │ ✓       │ ✗        │
	// │ Remove field with default   │ ✓        │ ✗       │ ✗        │
	// │ Remove field w/o default    │ ✓        │ ✗       │ ✗        │
	// │ Rename field                │ ✗        │ ✗       │ ✗        │
	// │ Change type (int→long)      │ ✓*       │ ✓*      │ ✓*       │
	// │ Change type (string→int)    │ ✗        │ ✗       │ ✗        │
	// │ Add enum value              │ ✗ (!)    │ ✓       │ ✗        │
	// │ Remove enum value           │ ✓        │ ✗       │ ✗        │
	// └─────────────────────────────┴──────────┴─────────┴──────────┘
	// *type promotion only (int→long, float→double)
	//
	// KEY RULE: "Add field with default" is ALWAYS safe.
	//
	// GOTCHA: Adding enum value is NOT backward compatible!
	// Old consumers don't know about the new value and can't decode it.
	// Solution: use strings instead of enums for evolving value sets.
	//
	// PROTOBUF EVOLUTION RULES:
	// ─────────────────────────
	// ┌─────────────────────────────┬──────────┬─────────┬──────────┐
	// │ Change                      │ BACKWARD │ FORWARD │ FULL     │
	// ├─────────────────────────────┼──────────┼─────────┼──────────┤
	// │ Add optional field          │ ✓        │ ✓       │ ✓        │
	// │ Remove optional field       │ ✓        │ ✓       │ ✓        │
	// │ Add required field (proto2) │ ✗        │ ✗       │ ✗        │
	// │ Change field number         │ ✗        │ ✗       │ ✗        │
	// │ Change field type           │ ✗*       │ ✗*      │ ✗*       │
	// │ Rename field                │ ✓**      │ ✓**     │ ✓**      │
	// │ Add enum value              │ ✓***     │ ✓       │ ✓***     │
	// └─────────────────────────────┴──────────┴─────────┴──────────┘
	// * except compatible wire types (int32/int64/uint32/...)
	// ** Protobuf uses field numbers, not names, on the wire
	// *** If old consumer uses proto3 (unknown enum → 0/default)
	//
	// Proto3 is inherently more evolution-friendly than Avro because:
	// - All fields are optional (no required)
	// - Field IDs are stable (renaming is free)
	// - Unknown fields are preserved (forward compatible by default)

	fmt.Println("  Avro golden rule: add fields WITH defaults")
	fmt.Println("  Avro gotcha: adding enum values is NOT backward-compatible")
	fmt.Println("  Protobuf: field numbers are key, renaming is free")
	fmt.Println()
}

// =============================================================================
// PART 6: MIGRATION STRATEGIES — When you need breaking changes
// =============================================================================
func migrationStrategies() {
	fmt.Println("--- MIGRATION STRATEGIES ---")

	// Sometimes you NEED an incompatible schema change.
	// Rename a field, change a type, restructure the message.
	// Here's how to do it WITHOUT breaking consumers.
	//
	// STRATEGY 1: NEW TOPIC (cleanest, recommended)
	// ──────────────────────────────────────────────
	// 1. Create topic "orders-v2" with new schema
	// 2. Migrate producers to write to orders-v2
	// 3. Run bridge: consume from "orders", transform, produce to "orders-v2"
	// 4. Migrate consumers from "orders" to "orders-v2"
	// 5. Decommission "orders" after retention expires
	//
	// Pros: Clean separation, can run both in parallel, easy rollback.
	// Cons: Temporary duplicate storage, bridge to maintain.
	//
	// STRATEGY 2: DUAL-FORMAT (gradual migration)
	// ────────────────────────────────────────────
	// 1. Add ALL new fields to the existing schema (with defaults)
	// 2. Producers start writing both old and new field formats
	// 3. Consumers read new fields if present, fall back to old fields
	// 4. Once all consumers updated: producers stop writing old fields
	// 5. Deprecate old fields (but never remove — keep defaults)
	//
	// Pros: No new topic, no bridge, gradual.
	// Cons: Schema grows (accumulated deprecated fields), messier.
	//
	// STRATEGY 3: ENVELOPE PATTERN
	// ────────────────────────────
	// Wrap the payload in an envelope that includes version info:
	//
	// {
	//   "schema_version": 2,
	//   "payload_type": "com.example.OrderV2",
	//   "payload": <binary data>
	// }
	//
	// Consumer checks schema_version → uses appropriate decoder.
	// Pros: Explicit versioning, any change is possible.
	// Cons: More complex consumer logic, loses Schema Registry benefits.
	//
	// RECOMMENDATION:
	// ───────────────
	// 1. Try to make the change backward-compatible (add with default)
	// 2. If impossible: NEW TOPIC strategy is cleanest
	// 3. Use DUAL-FORMAT for small incremental changes
	// 4. ENVELOPE only for polymorphic topics with many event types

	fmt.Println("  Breaking change? New topic + bridge (cleanest)")
	fmt.Println("  Small change? Dual-format with deprecated fields")
	fmt.Println("  Best practice: evolve with defaults, avoid breaks entirely")
	fmt.Println()
}


























































































































































































































































































































































}	fmt.Println("  Always register schemas in Schema Registry from day one")	fmt.Println("  Breaking changes: new topic + migration consumer")	fmt.Println("  JSON → Avro: new topic (recommended) or dual-format detection")	// 5. Decommission old topic	// 4. Switch consumers to new topic	// 3. Write a migration consumer: reads old topic → transforms → writes new topic	// 2. Create a new topic with the new schema	// 1. Temporarily set compatibility to NONE (or use a new subject)	// ────────────────────────────────────	// HOW TO DO A BREAKING SCHEMA CHANGE:	//	// 5. Eventually all data is Avro, remove JSON handling	//    0x00 = Avro (new format), otherwise = JSON (old format)	// 4. Consumer detects format by checking first byte:	// 3. Producer starts writing Avro (with schema ID prefix)	// 2. Register Avro schema as the first version	// 1. Configure Schema Registry subject for the topic	// APPROACH 2: Same Topic, Gradual Migration	//	// 5. After retention period: delete old topic	// 4. When all consumers migrated: stop writing to old topic	// 3. Migrate consumers one by one to new topic	// 2. Dual-write: producer writes to both old and new topics	// 1. Create new topic "events-v2" with Avro serialization	// APPROACH 1: New Topic (recommended)	// ──────────────────────────────────────	// HOW TO MIGRATE FROM RAW JSON TO AVRO:	fmt.Println("--- MIGRATION STRATEGIES ---")func migrationStrategies() {// =============================================================================// PART 6: MIGRATION STRATEGIES// =============================================================================}	fmt.Println()	fmt.Println("  Never: change types, reuse field numbers, remove non-default fields")	fmt.Println("  Protobuf: add/remove fields freely (uses field numbers)")	fmt.Println("  Avro: add fields with defaults, remove fields with defaults")	// Then in V3 you can safely remove it.	//   {"name": "name", "type": "string", "default": ""}	// Fix: First add a default to "name" in V2.5:	// V3: REMOVE name (has no default!) → BREAKING ❌	//	//   null is the first type → default: null → backward compatible.	//   Avro requires defaults to match the FIRST type in a union.	// Why "email" uses union ["null", "string"]:	//	// ]}	//   {"name": "email", "type": ["null", "string"], "default": null}	//   {"name": "name", "type": "string"},	//   {"name": "id", "type": "long"},	// {"type": "record", "name": "User", "fields": [	// V2: ADD email with default → BACKWARD COMPATIBLE ✅	//	// ]}	//   {"name": "name", "type": "string"}	//   {"name": "id", "type": "long"},	// V1: {"type": "record", "name": "User", "fields": [	//	// PRACTICAL SCHEMA EVOLUTION EXAMPLE (Avro):	//	// ❌ Reuse a deleted field number (use `reserved` keyword!)	// ❌ Change a field TYPE to an incompatible type	// ❌ Change a field NUMBER	// BREAKING CHANGES:	//	// ✅ Rename a field (Protobuf uses field numbers, not names)	// ✅ Remove a field (old field number is reserved)	// ✅ Add a new field (unknown fields are ignored by old consumers)	// SAFE CHANGES:	// ────────────────	// PROTOBUF RULES:	//	// ❌ Rename a field (different name = different field)	// ❌ Change a field's type (int → string)	// ❌ Remove a field WITHOUT a default value	// ❌ Add a field WITHOUT a default value (old data lacks it)	// BREAKING CHANGES:	//	// ✅ Change field order (Avro uses names, not position)	// ✅ Remove a field that has a default value	// ✅ Add a field with a default value	// SAFE CHANGES (backward compatible):	// ───────────	// AVRO RULES:	fmt.Println("--- EVOLUTION RULES ---")func evolutionRules() {// =============================================================================// PART 5: EVOLUTION RULES — What changes break what// =============================================================================}	fmt.Println()	fmt.Println("  FULL_TRANSITIVE: safest (compatible with ALL versions)")	fmt.Println("  FULL: both directions (deploy in any order)")	fmt.Println("  FORWARD: old consumer reads new data (deploy producers first)")	fmt.Println("  BACKWARD: new consumer reads old data (deploy consumers first)")	// NONE: dangerous. Use only during development or one-time migrations.	//	//   Use for: when producer team deploys first and consumer adapts.	// FORWARD: less common. Producers first, then consumers.	//	//   Use for: typical microservices where consumer team deploys independently.	// BACKWARD: most common default. Consumers first, then producers.	//	//   Use for: critical topics, shared across many teams.	// FULL_TRANSITIVE: maximum safety. Every schema works with every other.	// ─────────────	// WHICH TO USE:	//	// └──────────────────────────────────────────────────────────────────┘	// │           FULL_TRANSITIVE = compatible with ALL versions (1..N-1) │	// │  Example: FULL = compatible with version N-1 only                 │	// │  not just the latest. More strict.                                │	// │  Same as above but checked against ALL previous versions,         │	// │  BACKWARD_TRANSITIVE / FORWARD_TRANSITIVE / FULL_TRANSITIVE:     │	// │                                                                  │	// │  Use for: development topics, one-time migration topics.          │	// │  No compatibility checking. Breaking changes allowed.             │	// │  NONE:                                                            │	// │                                                                  │	// │  Deploy: ANY ORDER.                                               │	// │  Most restrictive but safest.                                     │	// │  Allowed: add/remove OPTIONAL fields with defaults only.          │	// │  BACKWARD + FORWARD simultaneously.                               │	// │  "Both old and new can read each other's data."                   │	// │  FULL:                                                            │	// │                                                                  │	// │  Deploy: UPDATE PRODUCERS FIRST, then consumers.                  │	// │  Allowed: add fields with defaults, remove optional fields.       │	// │  Old schema can READ data written with new schema.                │	// │  "Old consumers CAN read new producers' data."                    │	// │  FORWARD:                                                         │	// │                                                                  │	// │  Deploy: UPDATE CONSUMERS FIRST, then producers.                  │	// │  Allowed: add optional fields, remove fields with defaults.       │	// │  New schema can READ data written with old schema.                │	// │  "New consumers CAN read old producers' data."                    │	// │  BACKWARD (default):                                              │	// │                                                                  │	// │  COMPATIBILITY MODES:                                             │	// ┌──────────────────────────────────────────────────────────────────┐	fmt.Println("--- COMPATIBILITY MODES ---")func compatibilityModes() {// =============================================================================// PART 4: COMPATIBILITY MODES// =============================================================================}	fmt.Println()	fmt.Println("  Caching: after warmup, no Registry calls needed (survives downtime)")	fmt.Println("  Wire format: 5 bytes overhead (magic + schema ID)")	fmt.Println("  Schema Registry: stores schemas, assigns IDs, checks compatibility")	// - Karapace (Aiven, drop-in replacement)	// - AWS Glue Schema Registry	// - Apicurio Registry (Red Hat, open source)	// ALTERNATIVES TO CONFLUENT SCHEMA REGISTRY:	//	// - Schema subjects: typically <topic>-key, <topic>-value	// - REST API: GET/POST/DELETE schemas	// - Can run multiple instances (one is leader via Kafka)	// - Stores schemas in a Kafka topic (_schemas, compacted)	// - Stateless REST service	// ───────────────────────	// SCHEMA REGISTRY ITSELF:	//	// └──────────────────────────────────────────────────────────────┘	// │  (as long as schemas are cached). But NEW schemas need it.    │	// │  Registry downtime doesn't affect running producers/consumers│	// │  After warmup, NO Registry calls are needed.                  │	// │  CACHING: Schema lookups are cached locally.                  │	// │                                                              │	// │  5. Deserialize payload using the schema                      │	// │  4. If not: GET schema from Registry by ID                    │	// │  3. Check local cache: do I have schema for this ID?         │	// │  2. Extract schema ID from first 5 bytes                     │	// │  1. Read record from Kafka                                    │	// │  CONSUMER FLOW:                                               │	// │                                                              │	// │  6. Send to Kafka                                             │	// │  5. Prepend 0x00 + schema ID to the serialized bytes         │	// │  4. Serialize data using the schema                           │	// │  3. If not: POST schema to Registry → get schema ID          │	// │  2. Check local cache: do I have schema ID for this schema?  │	// │  1. Producer has data to serialize                             │	// │  PRODUCER FLOW:                                               │	// │                                                              │	// │  The schema ID points to the schema in Schema Registry.      │	// │  Total overhead: 5 bytes per record.                         │	// │                                                              │	// │   (1 byte)           (big-endian int32)   (serialized data)  │	// │  [magic byte: 0x00] [schema ID: 4 bytes] [payload: N bytes] │	// │                                                              │	// │  CONFLUENT WIRE FORMAT (5 bytes overhead per record):         │	// ┌──────────────────────────────────────────────────────────────┐	//	// WIRE FORMAT (how data travels from producer to consumer):	//	// 4. Provides schemas to consumers for deserialization	// 3. Checks compatibility when new schemas are registered	// 2. Assigns each schema version a unique SCHEMA ID (integer)	// 1. Stores schemas (Avro, Protobuf, JSON Schema)	// Schema Registry is a separate service that:	fmt.Println("--- SCHEMA REGISTRY ---")func schemaRegistry() {// =============================================================================// PART 3: SCHEMA REGISTRY// =============================================================================}	fmt.Println()	fmt.Println("  Pick one format per domain. Don't mix within a topic.")	fmt.Println("  Protobuf: gRPC ecosystem, required code gen, excellent evolution")	fmt.Println("  Avro: smallest, Kafka-native, best for data pipelines")	// - Don't mix formats within a topic	// - JSON Schema if you're migrating from raw JSON	// - Protobuf if you're a microservices team with gRPC	// - Avro if you're primarily a Kafka/data team	// ──────────────────	// MY RECOMMENDATION:	//	// - gRPC compatibility if you also have gRPC services	// - Field numbers ensure forward/backward compatibility	// - Required code gen (not always desirable for dynamic systems)	// - Great tooling (protoc, code gen for every language)	// ────────────────────────────────────	// PROTOBUF — The cloud-native choice:	//	// - Confluent uses Avro as the default for Kafka	// - Reader's schema can differ from writer's schema (schema evolution!)	// - Data is pure values, no field names → very compact	// - Schema is stored SEPARATELY from data (in Schema Registry)	// ────────────────────────────────	// AVRO — The Kafka-native choice:	//	// └──────────────┴───────────┴───────────┴──────────────┴───────────────┘	// │ Human read   │ No (bin)  │ No (bin)  │ Yes          │ Yes           │	// │              │ Kafka     │ Cloud     │              │               │	// │ Ecosystem    │ Hadoop/   │ gRPC/     │ REST/Web     │ Everything    │	// │ Null handling│ Union type│ Optional  │ nullable     │ Implicit      │	// │ Default vals │ Yes       │ Yes       │ Yes          │ None          │	// │ Code gen     │ Optional  │ Required  │ Optional     │ None          │	// │ Evolution    │ Excellent │ Good      │ Good         │ None          │	// │ Schema       │ .avsc     │ .proto    │ JSON Schema  │ None          │	// │ Speed        │ Fast      │ Fast      │ Slow         │ Slow          │	// │ Size         │ Smallest  │ Small     │ Large        │ Largest       │	// ├──────────────┼───────────┼───────────┼──────────────┼───────────────┤	// │ Feature      │ Avro      │ Protobuf  │ JSON Schema  │ Raw JSON      │	// ┌──────────────┬───────────┬───────────┬──────────────┬───────────────┐	fmt.Println("--- SERIALIZATION FORMATS ---")func serializationFormats() {// =============================================================================// PART 2: SERIALIZATION FORMATS// =============================================================================}	fmt.Println()	fmt.Println("  JSON OK for: low throughput, logging, prototyping")	fmt.Println("  Avro: 40% smaller, 5-10x faster, schema evolution built in")	fmt.Println("  JSON: no schema, big, slow to parse")	// - When you use JSON Schema with Schema Registry (gets you schemas)	// - Rapid prototyping (switch to Avro/Protobuf later)	// - Human-readable logs or debug topics	// - Low throughput topics (< 1000 records/sec)	// ────────────────	// WHEN JSON IS OK:	//	// 5-10x faster deserialization with binary formats.	// Avro/Protobuf → direct binary decoding → no parsing step	// JSON → string parsing → type checking → struct mapping	// ────────────────────────	// PROBLEM 3: PARSING SPEED	//	// At 1M records/sec: 40 MB/sec saved.	// Avro equivalent: 20 bytes (field names stored in schema, not in data).	//	//   → 60 bytes (field names: 24 bytes = 40% overhead!)	// {"user_id": 123, "name": "Alice", "email": "alice@example.com"}	// JSON repeats field names in EVERY record.	// ───────────────	// PROBLEM 2: SIZE	//	// Consumer breaks. In production. At 3 AM.	//	// {"user_id": "123", "name": "Alice"}      ← v3 (someone changed int to string)	// {"userId": 123, "name": "Alice"}         ← v2 (someone renamed the field)	// {"user_id": 123, "name": "Alice"}       ← v1	// JSON has no schema. Any producer can send any shape of data.	// ─────────────────────	// PROBLEM 1: NO SCHEMA	//	// And it's a TRAP at scale.	// Most teams start with JSON. It's human-readable, flexible, easy.	fmt.Println("--- WHY NOT RAW JSON ---")func whyNotJson() {// =============================================================================// PART 1: WHY NOT RAW JSON// =============================================================================}	migrationStrategies()	evolutionRules()	compatibilityModes()	schemaRegistry()	serializationFormats()	whyNotJson()	fmt.Println()	fmt.Println("=== SCHEMA EVOLUTION ===")func main() {import "fmt"package main// =============================================================================//// consumers in production. Schema Registry + compatibility checks prevent this.// You WILL change your message format. Without schema management, you'll break// In a microservices world, producers and consumers are deployed independently.// THE KEY INSIGHT://// - Practical migration strategies// - Schema evolution rules: what changes break what// - Compatibility modes: BACKWARD, FORWARD, FULL, NONE// - Schema Registry: the brain behind schema management// - Avro, Protobuf, JSON Schema: tradeoffs at scale// - Why JSON is a trap for production Kafka// WHAT YOU'LL LEARN://// =============================================================================// LESSON 8.1: SCHEMA EVOLUTION — Never Break Your Consumers// =============================================================================