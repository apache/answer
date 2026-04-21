# Vector Search & Semantic Search Design

## Architecture Layers

```
┌──────────────────────────────────────────────┐
│  AI Chat / MCP Tool ("semantic_search")      │  ← Controller layer
├──────────────────────────────────────────────┤
│  EmbeddingService                            │  ← Service layer (thin facade)
├──────────────────────────────────────────────┤
│  plugin.VectorSearch interface               │  ← Plugin abstraction
├──────────────────────────────────────────────┤
│  pgvector / elasticsearch / weaviate / ...   │  ← Plugin implementations
└──────────────────────────────────────────────┘
```

## Plugin Interface (`plugin/vector_search.go`)

- `RegisterSyncer(ctx, syncer)` — core provides a syncer for bulk data pull
- `SearchSimilar(ctx, query, topK)` — returns `[]VectorSearchResult{ObjectID, ObjectType, Metadata, Score}`
- `UpdateContent(ctx, content)` — upserts a document with embedding
- `DeleteContent(ctx, objectID)` — removes a document
- `ConfigReceiver(config)` / `ConfigFields()` — plugin config lifecycle

`GenerateEmbedding()` is the shared embedding utility used by plugins.

## Content Syncing (`vector_search_sync/syncer.go`)

Core implements `VectorSearchSyncer` with:

- `GetQuestionsPage(page, pageSize)`
- `GetAnswersPage(page, pageSize)`

Each indexed document aggregates question/answer/comment text. Metadata stores deshortened IDs for reconstruction at query time.

Sync is triggered from `RegisterSyncer()` (startup + config update flow).

## Startup & Activation Flow

```
initPluginData():
  1. Load plugin status from DB
  2. Call ConfigReceiver for configured plugins
     -> parse config always
     -> if active: run heavy init (probe embedding + connect/schema checks)
     -> if inactive: skip heavy init (IsEnabled guard)
  3. Call RegisterSyncer for vector search plugins
     -> if active/initialized: trigger full sync
     -> if inactive/uninitialized: skip sync
```

On admin config save:
1. `ConfigReceiver`
2. `UpdatePluginConfig` -> `RegisterSyncer` -> full sync

### Current Behavior Summary

- **Active plugin on startup**: does one probe embedding call, then full sync to vector storage.
- **Inactive plugin on startup**: parses config only; no probe embedding and no sync.
- **Config save for active plugin**: re-runs init path and full sync.

## Semantic Search Query Flow

```
User query -> MCP tool "semantic_search"
  -> EmbeddingService.SearchSimilar(query, topK)
  -> plugin.SearchSimilar() returns scored IDs + metadata
  -> handler fetches full DB content (question/answers/comments)
  -> returns structured semantic search response
```

## Follow-up: Real-Time Sync Gap

### Comparison with Search Plugin

| Aspect | Search Plugin | VectorSearch Plugin |
|--------|---------------|---------------------|
| Bulk sync | Yes | Yes |
| Real-time sync | Yes (create/update/delete hooks) | No |
| Trigger | Event-driven + startup/config | Startup/config only |
| Consistency | Near real-time | Eventually consistent |

### Current Gap

`UpdateContent()` / `DeleteContent()` exist in `plugin.VectorSearch`, but are not called from question/answer service events. So after initial sync, content changes are not reflected until next full re-sync.

### Options

1. **Manual (current)**
   - Re-sync only on plugin config save/update
   - Simple, but stale results between syncs

2. **Real-time**
   - Add event hooks to call vector search update/delete
   - Can be async (goroutine / queue) to avoid write-path latency
   - Higher embedding API call volume

3. **Scheduled (cron)**
   - Periodic bulk sync via cron expression
   - Good for off-peak syncing
   - Delayed freshness until next run

### Proposed Common Setting (All Vector Search Plugins)

Add a **single common sync policy setting** at the vector-search framework level (not per plugin implementation), so all vector search plugins behave consistently.

#### New common config fields

- `sync_mode`: enum
  - `manual`
  - `realtime`
  - `scheduled`
- `sync_crontab`: string (required only when `sync_mode=scheduled`)

#### Behavior by mode

- `manual`
  - No runtime incremental sync.
  - Full sync only when plugin config is saved/updated (current behavior).

- `realtime`
  - Hook question/answer create/update/delete events in core service layer.
  - Trigger `VectorSearch.UpdateContent` / `VectorSearch.DeleteContent` for the active plugin.
  - Recommend async dispatch (queue/worker) to avoid request-path latency.

- `scheduled`
  - Core scheduler triggers periodic full sync (`RegisterSyncer`/sync routine) by `sync_crontab`.
  - Intended for off-peak indexing windows.

#### Design constraints

- This setting should be **core-owned and plugin-agnostic**.
- Individual vector plugins should not implement their own sync-mode semantics.
- Core decides when to call plugin sync/update APIs; plugins only execute storage/index operations.
- Switching modes should apply immediately after config save.
- Inactive plugins remain parse-only (no heavy init/sync).

#### Why common instead of per-plugin

- Consistent admin UX and operational behavior.
- Prevents drift where plugins implement different sync semantics.
- Easier future enhancements (retry, dead-letter queue, backfill controls, observability) in one place.
