-- Migration: Add indexes for duplicate prevention performance optimization
-- This migration adds database indexes to improve the performance of geospatial
-- and name-based duplicate detection queries.

-- Index for geospatial queries (latitude, longitude)
-- This composite index enables efficient bounding box queries for proximity searches
CREATE INDEX IF NOT EXISTS idx_spots_location 
ON spots (latitude, longitude) 
WHERE deleted_at IS NULL;

-- Index for name-based searches combined with status filtering
-- This helps with fuzzy name matching queries and status-aware duplicate detection
CREATE INDEX IF NOT EXISTS idx_spots_name_status 
ON spots (name, status) 
WHERE deleted_at IS NULL;

-- Index for status and source filtering
-- Optimizes queries that filter by verification status and data source
CREATE INDEX IF NOT EXISTS idx_spots_status_source 
ON spots (status, source) 
WHERE deleted_at IS NULL;

-- Partial index for active spots (non-deleted)
-- Improves performance for queries that exclude soft-deleted records
CREATE INDEX IF NOT EXISTS idx_spots_active 
ON spots (id, created_at) 
WHERE deleted_at IS NULL;

-- Index for PlaceID lookups (already exists as unique index, but ensuring it's optimal)
-- The existing uniqueIndex on place_id should handle this, but we ensure it exists
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_spots_place_id ON spots (place_id);

-- Index for last_seen timestamp queries
-- Useful for tracking when spots were last discovered by the scraper
CREATE INDEX IF NOT EXISTS idx_spots_last_seen 
ON spots (last_seen DESC) 
WHERE deleted_at IS NULL AND last_seen IS NOT NULL;

-- Composite index for user-specific queries
-- Optimizes queries filtering by user and source
CREATE INDEX IF NOT EXISTS idx_spots_user_source 
ON spots (user_id, source) 
WHERE deleted_at IS NULL AND user_id IS NOT NULL;

-- Comments explaining the indexes:
-- 1. idx_spots_location: Primary index for geospatial duplicate detection
--    Enables efficient bounding box queries before applying Haversine distance
-- 2. idx_spots_name_status: Supports name-based duplicate detection with status filtering
--    Helps when searching for similar names among verified/pending spots
-- 3. idx_spots_status_source: General performance improvement for status/source filtering
--    Used in various queries throughout the application
-- 4. idx_spots_active: Optimizes queries on active (non-deleted) spots
--    Most queries should exclude soft-deleted records
-- 5. idx_spots_last_seen: Supports scraper-related queries and spot freshness tracking
-- 6. idx_spots_user_source: Optimizes user-specific spot queries