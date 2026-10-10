# NASP Key-Value Engine

A custom key-value storage engine developed as part of the course Advanced Algorithms and Data Structures. The project is implemented in Go and focuses on practical database-engine concepts such as write-ahead logging, in-memory memtables, LSM-tree organization, SSTable persistence, block management, caching, and efficient lookup and scan operations.

## Project overview

This project is a lightweight database-like engine designed to manage key-value pairs with durability, efficient reads, and scalable storage behavior. Instead of relying on a single monolithic storage mechanism, it combines multiple layers commonly found in modern storage systems:

- Memtables for fast writes
- Write-ahead log (WAL) for crash recovery
- LSM tree for organizing immutable sorted data
- SSTables for persistent storage
- Block manager for reading and writing data in fixed-size blocks
- In-memory cache for hot objects
- Dictionary and index structures for lookup acceleration
- Bloom filters and other probabilistic structures for efficient search
- CLI for direct interaction with the engine

The engine is intended as both an educational implementation and a research-oriented prototype for understanding how storage backends organize data and ensure reliability under real-world workloads.

## Why this project exists

Modern database engines must balance several competing goals:

- Write throughput must remain high even as data grows
- Reads should be fast, especially for frequently accessed keys
- Data must survive crashes and unexpected failures
- Large datasets should be stored efficiently on disk
- Range and prefix queries should be supported without losing performance

This project demonstrates these principles in a simplified but realistic architecture. It models a storage layer similar to systems used in log-structured databases, while keeping the code understandable and approachable for academic learning.

## Main architectural ideas

### 1. LSM-tree based design
The engine stores updates in an LSM (Log-Structured Merge) tree. New writes are first appended to an in-memory memtable. When the memtable reaches a threshold, it is flushed to disk as an SSTable. Over time, data is compacted and merged across levels, reducing the cost of old updates while preserving read efficiency.

This design is especially useful for write-heavy workloads, where appending and merging large sorted files is often faster than updating many random disk locations.

### 2. Write-ahead logging
Before data is considered durable, it is written to a WAL. This protects against crashes by allowing the engine to replay recent operations during recovery. The WAL is especially important for preserving data that has been accepted by the system but not yet fully flushed to the persistent tree structure.

### 3. SSTables and block storage
Persistent data is organized into SSTables and managed through a block manager. This allows the system to:

- Store data in fixed-size blocks
- Serialize structured records efficiently
- Reuse file blocks with indexing and dictionary metadata
- Support efficient reads and compaction operations

### 4. Cache and dictionary layers
Hot entries are cached in memory to avoid repeated disk access, while dictionary structures support faster record lookup and range navigation. These layers reduce latency without dramatically increasing implementation complexity.

### 5. Search optimization
The project includes several indexing and probabilistic data structures, such as:

- Bloom filters for quick negative lookups
- Hash-based structures for distributed key lookups
- Prefix and range scanning methods
- Summary information and metadata for faster data navigation

### 6. Access control and rate limiting
The implementation also includes token bucket based rate limiting to control access frequency and protect the system from excessive demand. This makes the engine more robust in a realistic deployment scenario.

## Features

- Key-value write, read, and delete operations
- In-memory memtable storage
- WAL-backed persistence
- LSM-tree organization and flushing
- SSTable creation and compaction
- Prefix scans and range scans
- Ordered iteration over keys
- Bloom filter support
- Block cache and dictionary-based lookup
- Token bucket rate limiting
- Command-line interface for direct interaction

## Project structure

- `main.go` – application entry point and initialization
- `cli/` – command-line interface for engine operations
- `structures/` – core storage engine implementation
- `utils/` – helper utilities and conversion logic
- `hash/` – hash-related functionality
- `data/` – persisted data and generated storage artifacts
- `wal/` – write-ahead log storage
- `config.json` – configuration for memory, cache, block size, memtable count, and LSM behavior

## Core components

### Memtables
The memtable is the active in-memory structure used for recent writes. It provides fast insertion and lookup while the data is still fresh. Once a threshold is reached, the memtable is converted into disk-friendly data and moved into the LSM pipeline.

### Write-Ahead Log
The WAL records database operations before full persistence. This provides a recovery mechanism after unexpected termination and supports consistent replay of recent writes.

### LSM tree and SSTables
The LSM tree organizes multiple levels of sorted data. Newer data is stored at higher levels and gradually merges into lower layers. This provides efficient ingestion while still allowing reads to be resolved through tiered search techniques.

### Block manager
The block manager reads and writes data in blocks, which is a common strategy in storage systems to reduce overhead and improve locality.

### Bloom filter
Bloom filters allow the engine to quickly determine whether a key is likely present before performing a more expensive lookup, which is especially useful in large persistent datasets.

## Example supported operations

The project includes a CLI that supports commands such as:

- `PUT key value`
- `GET key`
- `DELETE key`
- `PREFIX_SCAN prefix pageNumber pageSize`
- `RANGE_SCAN minKey maxKey pageNumber pageSize`
- `PREFIX_ITERATE prefix`
- `RANGE_ITERATE minKey maxKey`
- `BF_CREATE key expectedInsertions falsePositiveRate`
- `BF_PUT key element`
- `BF_CHECK key element`

These commands demonstrate the engine's ability to handle direct gets, deletes, range operations, and probabilistic membership checks.

## Configuration

The engine is configured through `config.json`, which controls parameters including:

- WAL segment size
- Memtable count and capacity
- SSTable behavior
- LSM maximum levels
- Cache size
- Block size
- Token bucket rate-limit settings

These values can be adjusted depending on the expected workload and performance goals.

## Execution

To run the project:

```bash
go run .
```

After startup, the CLI will prompt for commands and allow interaction with the database engine.

## Team

- Petar Popović SV17/2023
- Aleksa Dejanović SV24/2023
- Sara Stojkov SV38/2023
- Boško Vasilić SV48/2023

## Summary

NASP Key-Value Engine is a practical implementation of a modern storage engine architecture in Go. It demonstrates how a database can combine memtables, WAL durability, LSM-tree compaction, block-based persistent storage, and indexing techniques into a working, extensible system. The project is ideal for understanding the internals of high-performance key-value stores and serves as a strong learning foundation for advanced data-structure and systems design concepts.
