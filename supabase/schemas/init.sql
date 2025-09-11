-- Core schema initialization
-- Extensions and types

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable moddatetime extension for automatic updated_at triggers
CREATE EXTENSION IF NOT EXISTS moddatetime;

-- Thread status enum
CREATE TYPE thread_status AS ENUM ('pending', 'scraping', 'completed', 'failed');
