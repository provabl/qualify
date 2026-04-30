-- qualify Database Initialization
-- This script runs automatically when PostgreSQL container starts

-- Create extensions (migrations will handle schema)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Database is ready for migrations
