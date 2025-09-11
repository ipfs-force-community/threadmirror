# Coding guidelines

> These guidelines are enforced by Cursor IDE for this Go backend project with **SQL First** architecture.

## Project Architecture Standards (SQL First)

### 1. API Development Workflow (OpenAPI-First)
- **Always start with OpenAPI specification**: Create/update `api/v1/openapi.yaml`
  - Use comprehensive constraints: `required` fields, enum definitions, type restrictions
  - Include detailed `summary` and `description` for all endpoints
  - Follow OpenAPI 3.0+ standards
  - Reuse first: reference existing schemas/params/responses via `$ref`
  - Paginated responses must include the following property:
  ```yaml
  meta:
    $ref: "#/components/schemas/PaginationMeta"
  ```
- **Code generation(MUST)**: Run `make generate` to auto-generate:
  - `internal/api/v1/types_gen.go` (request/response types)
  - `internal/api/v1/gin_gen.go` (handler methods declared in `ServerInterface`, interfaces and route registration codes)
- **Implementation**: Create the API-handler files under `internal/api/v1/xxx.go`. Add methods to `V1Handler` so it implements `ServerInterface` in those files, DO NOT create a new Handler structure. `V1Handler` should depend only on the structs defined in the `service` package.

### 2. Directory Structure (SQL First Architecture)
```
internal/
├── api/
│   ├── v1/           # API handlers and generated types
│   └── v1/util.go    # API-specific helpers
├── sqlc/             # SQLC generated code and queries
│   ├── db.go         # Generated database interface and types
│   ├── models.go     # Generated database models
│   └── *.sql.go      # Generated query implementations
└── service/          # Business logic layer (direct SQLC usage)
pkg/util/             # Shared utility functions
i18n/                 # Internationalization configuration
api/v1/
└── openapi.yaml      # OpenAPI specification
sql/
└── queries/          # SQL query files for SQLC
supabase/
├── schemas/          # Declarative schema files
│   ├── init.sql      # Core schema definitions (extensions, types)
│   ├── *.sql         # Table definitions (each table as independent SQL file)
│   ├── function/     # Database functions (each function as independent SQL file)
│   └── trigger/      # Trigger definitions (each trigger as independent SQL file)
├── migrations/       # Generated migration files (auto-generated from schemas)
└── seed/
    └── functions/    # Database functions by module
sqlc.yaml             # SQLC configuration
```

### 3. SQL First Dependency Architecture
- **Service layer**: Direct dependency on `*sql.DB`, use `db.Queries()` to access SQLC generated code
- **No Repository layer**: Eliminated for simplified architecture
- **Database functions**: Handle complex multi-table operations atomically
- **Triggers**: Automatic field updates (e.g., counters, timestamps)
- **Dependency injection**: Use constructor functions with database dependency
- **Naming Convention**: 
  - Service names: PascalCase (e.g., `PostService`)
  - Database functions: snake_case (e.g., `follow_user`)
  - Trigger functions: snake_case with descriptive names (e.g., `update_user_follow_counts`)

### 4. Database Usage Standards (SQL First with SQLC)

#### 4.1 Simple Queries (SQLC)
- **Location**: All simple query files must be in `sql/queries/` directory
- **Principle**: Use for any operations that can be accomplished with a single SQL statement (including complex joins)
- **Format**: Use SQLC query annotations with named parameters:
  ```sql
  -- name: GetUser :one
  SELECT * FROM users WHERE id = @user_id;
  
  -- name: ListPosts :many
  SELECT * FROM posts ORDER BY created_at DESC LIMIT @limit_ OFFSET @offset_;
  
  -- name: CreatePost :one
  INSERT INTO posts (title, content, user_id) 
  VALUES (@title, @content, @user_id) RETURNING *;
  
  -- name: UpdatePost :exec
  UPDATE posts SET title = @title, content = @content WHERE id = @post_id;
  
  -- name: DeletePost :exec
  DELETE FROM posts WHERE id = @post_id;
  ```
- **Query Types**:
  - `:one` - Returns single row
  - `:many` - Returns multiple rows
  - `:exec` - Execute only, returns error
  - `:execresult` - Execute, returns sql.Result
  - `:copyfrom` - Bulk insert (PostgreSQL)
  - `:batchexec`, `:batchone`, `:batchmany` - Batch operations

#### 4.2 Database Functions (Complex Operations)
- **Purpose**: Handle atomic operations involving multiple tables or complex business logic
- **Location**: `supabase/schemas/function/` directory (each function as independent SQL file)
- **Function Principles**:
  - **Single Responsibility**: Each function should handle one specific business operation
  - **Atomic Operations**: Use for operations that must succeed or fail as a unit
  - **Avoid JSON Returns**: Prefer `void` return type for operations that only need success/failure indication
  - **Return Simple Types**: When data must be returned, use simple types (INTEGER, UUID, TEXT) instead of JSON
  
- **Example - Good Practice (void return)**:
  ```sql
  create or replace function public.follow_user(
      p_follower_id uuid,
      p_following_id uuid
  )
  returns void
  language plpgsql
  security invoker
  as $$
  begin
      -- Validation and business logic
      if p_follower_id = p_following_id then
          raise exception 'Cannot follow yourself';
      end if;
      
      -- Insert follow relationship (triggers handle count updates)
      insert into public.user_follows (follower_id, following_id)
      values (p_follower_id, p_following_id);
  end;
  $$;
  ```

- **Example - Good Practice (simple return)**:
  ```sql
  create or replace function public.create_user_profile(
      p_user_id uuid,
      p_display_id text,
      p_nickname text
  )
  returns uuid
  language plpgsql
  security invoker
  as $$
  declare
      v_profile_id uuid;
  begin
      insert into public.user_profiles (id, display_id, nickname)
      values (p_user_id, p_display_id, p_nickname)
      returning id into v_profile_id;
      
      return v_profile_id;
  end;
  $$;
  ```

#### 4.3 Database Triggers (Automatic Updates)
- **Purpose**: Automatically maintain data consistency and update derived fields
- **Location**: `supabase/schemas/trigger/` directory (each trigger as independent SQL file)
- **Trigger Principles**:
  - **Minimal Logic**: Keep trigger functions as simple as possible
  - **Single Purpose**: Each trigger should handle one specific update operation
  - **Avoid Unnecessary TG_OP Checks**: Only check `TG_OP` when genuinely needed for reuse
  - **Clear Documentation**: Document multi-purpose triggers clearly

#### 4.4 SQLC Configuration Standards
- **Configuration file**: `sqlc.yaml` in project root
- **Named Parameters Support**: SQLC automatically generates Go structs for named parameters when using `@param_name` syntax in queries
- **Nullable parameters Support**: sqlc infers the nullability of any specified parameters, and often does exactly what you want. If you want finer control over the nullability of your parameters, you may use `sqlc.narg()` (nullable arg) to override the default behavior. Using `sqlc.narg` tells sqlc to ignore whatever nullability it has inferred and generate a nullable parameter instead. There is no nullable equivalent of the `@` syntax.
  - Here is an example that uses a single query to allow updating an author’s name, bio or both.
    ```sql
    -- name: UpdateAuthor :one
    UPDATE author
    SET
    name = coalesce(sqlc.narg('name'), name),
    bio = coalesce(sqlc.narg('bio'), bio)
    WHERE id = sqlc.arg('id')
    RETURNING *;
    ```
    The following code is generated:
    ```go
    type UpdateAuthorParams struct {
      Name sql.NullString
      Bio  sql.NullString
      ID   int64
    }
    ```
- **Parameter Validation**: SQLC validates named parameter usage at compile time, ensuring all parameters are properly defined

#### 4.5 Service Layer Integration (SQL First)
- **Direct SQLC Usage**: Services directly use `*sql.DB` dependency
- **Query Access**: Access SQLC queries via `db.Queries()` method
- **Error Handling**: Define error types in service layer files
- **Prefer SQLC Generated Types**: Always use SQLC-generated types instead of creating custom structs
- **Example Service Pattern**:
  ```go
  type UserService struct {
      db *sql.DB
  }
  
  func NewUserService(db *sql.DB) *UserService {
      return &UserService{
          db: db,
      }
  }
  
  func (s *UserService) GetUserByID(id uuid.UUID) (*sqlc.UserProfile, error) {
      ctx := context.Background()
      user, err := s.db.Queries().GetUserByID(ctx, id)
      if err != nil {
          return nil, ErrUserNotFound
      }
      return &user, nil
  }
  
  // For complex operations, call database functions
  func (s *UserService) FollowUser(followerID, followingID uuid.UUID) error {
      ctx := context.Background()
      _, err := s.db.Queries().FollowUser(ctx, sqlc.FollowUserParams{
          FollowerID:  followerID,
          FollowingID: followingID,
      })
      return err
  }
  ```

#### 4.6 Service Layer Type Management (Mandatory)
- **NEVER Create Custom Structs in Service Layer**: Do not define custom structs for data that can be represented by SQLC-generated types
- **Use SQLC Generated Types**: Always prefer SQLC-generated Row types (`sqlc.GetPostByIDRow`, `sqlc.GetUserProfileRow`, etc.) for return values
- **Database Function Results**: When calling database functions that return complex data:
  - Use `s.db.Pool().Query()` for direct SQL execution when SQLC cannot generate proper types
  - Scan results into primitive types, then construct SQLC-generated structs
  - Return SQLC-generated types from service methods for consistency
- **Type Conversion**: If data type conversion is needed (e.g., `int32` to `int64`), perform it during scanning
- **Consistent API**: Ensure all service methods return the same SQLC types for similar operations

#### 4.7 Database Operations Best Practices
- **Simple Operations**: Use SQLC-generated queries for any operations that can be accomplished with a single SQL statement (including multi-table joins)
- **Complex Operations**: Use database functions for operations requiring multiple SQL statements or complex business logic that cannot be expressed in a single query
- **Transactions**: Database functions automatically handle transactions
- **Pagination**: Always implement pagination for list endpoints using LIMIT/OFFSET
- **Type Safety**: Leverage SQLC's compile-time type checking
- **Null Handling**: Use `sql.NullString`, `sql.NullTime`, etc. for nullable columns

#### 4.8 Data Type Mapping Standards
- **PostgreSQL Arrays**: Map to Go slices automatically
- **JSON Columns**: Use `[]byte` or custom types with proper scanning
- **UUIDs**: Use `github.com/google/uuid.UUID` type
- **Timestamps**: Use `time.Time` for NOT NULL, `sql.NullTime` for nullable
- **Custom Types**: Define type overrides in `sqlc.yaml` when needed

#### 4.9 Named Parameters Standards (pgx NamedArgs)
- **Mandatory Usage**: All SQL statements MUST use pgx NamedArgs instead of positional parameters ($1, $2, etc.)
- **Improved Readability**: Named parameters make SQL queries more self-documenting and maintainable
- **Reduced Errors**: Eliminates parameter order mistakes common with positional parameters
- **Enhanced Debugging**: Easier to trace parameter values during debugging

**SQLC Query Format** (Use named parameters in SQL):
```sql
-- Good: Using named parameters
-- name: GetUserByID :one
SELECT * FROM user_profiles WHERE id = @user_id;

-- name: ListPostsByUser :many
SELECT * FROM posts 
WHERE user_id = @user_id 
  AND created_at >= @start_date
  AND created_at <= @end_date
ORDER BY created_at DESC 
LIMIT @limit_ OFFSET @offset_;

-- name: CreatePost :one
INSERT INTO posts (title, content, user_id, status) 
VALUES (@title, @content, @user_id, @status) 
RETURNING *;

-- name: UpdateUserProfile :exec
UPDATE user_profiles 
SET nickname = @nickname,
    bio = @bio,
    updated_at = NOW()
WHERE id = @user_id;
```

**Bad Practice** (Avoid positional parameters):
```sql
-- Bad: Using positional parameters
-- name: GetUserByID :one
SELECT * FROM user_profiles WHERE id = $1;

-- name: UpdateUserProfile :exec
UPDATE user_profiles 
SET nickname = $2, bio = $3, updated_at = NOW()
WHERE id = $1;
```

### 5. Supabase Integration
- Use Supabase Auth for user authentication and management
- Leverage Supabase Storage for file uploads and management
- **Database Connection**: Use Supabase PostgreSQL connection string with SQLC

### 6. Development Workflow (SQL First)
1. **Design OpenAPI specification** (`api/v1/openapi.yaml`)
2. **Design database schema declaratively** (define desired final state)
3. **Create declarative schema files** in `supabase/schemas/` directory:
   - Define tables as independent SQL files directly in `supabase/schemas/` (e.g., `user_profile.sql`, `post.sql`)
   - Define functions as independent SQL files in `supabase/schemas/function/`
   - Define triggers as independent SQL files in `supabase/schemas/trigger/`
4. **Generate migrations automatically** using `supabase db diff`
5. **Review and apply migrations** locally for testing
6. **Write simple SQL queries** in `sql/queries/` directory with SQLC annotations
7. **Configure SQLC** in `sqlc.yaml` file (point to `supabase/schemas` for schema)
8. **Generate SQLC code** by running `sqlc generate`
9. **Test database operations** locally with Supabase
10. **Implement service layer** with direct SQLC usage (business logic)
11. **Register services** in `servicefx.Module`
12. **Generate API code** by `make generate` (This step is mandatory)
13. **Implement** `internal/api/v1.ServerInterface` methods (API handler methods)
14. **Write comprehensive unit tests** with database integration
15. **Integration testing** with complete workflow

### 7. Testing Standards (SQL First)
- **Unit Tests**: Test service methods using testcontainers or test databases
- **Integration Tests**: Test complete workflows including database functions and triggers
- **Database Functions Testing**: Test complex operations end-to-end
- **Test Data**: Use SQLC generated types for consistent test data creation

## SQL First Best Practices

### 1. PostgreSQL Architecture Patterns (Mandatory)

#### 1.1. UUID Primary Keys
**Always use a UUID primary key** (`uuid_generate_v4()`)

```sql
CREATE TABLE person (
  id  uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name text NOT NULL
);
```

#### 1.2. Timestamp Management
**Every table must include `created_at` and `updated_at`, with a BEFORE UPDATE trigger to maintain `updated_at`**

```sql
CREATE TABLE person (
  id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  name       text NOT NULL
);

CREATE OR REPLACE  TRIGGER set_person_updated_at
  BEFORE UPDATE ON person
  FOR EACH ROW 
  EXECUTE PROCEDURE moddatetime('updated_at');
```

#### 1.3. Foreign Key Constraints
**Foreign keys must use `ON UPDATE RESTRICT ON DELETE RESTRICT`** (unless cascade behavior is explicitly required)

```sql
CREATE TABLE pet (
  id       uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner_id uuid NOT NULL REFERENCES person(id)
                ON UPDATE RESTRICT ON DELETE RESTRICT,
  name     text NOT NULL
);
```

#### 1.4. Table Naming Conventions
**Table names are singular; many‑to‑many join tables are mechanically named `tableA_tableB`**

```sql
CREATE TABLE user_follow (
  id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  follower_id  uuid NOT NULL REFERENCES user_profile(id)
                   ON UPDATE RESTRICT ON DELETE RESTRICT,
  following_id uuid NOT NULL REFERENCES user_profile(id)
                   ON UPDATE RESTRICT ON DELETE RESTRICT,
  UNIQUE (follower_id, following_id)
);
```

#### 1.5. Object Naming and Schema Organization
**Triggers, functions, indexes, and other auxiliary objects should live in the same schema as the data tables and share a `set_` / `trg_` prefix**

```sql
-- Trigger function naming
CREATE FUNCTION set_current_timestamp_updated_at() ...

-- Trigger naming
CREATE OR REPLACE  TRIGGER set_user_profile_updated_at ...

-- Index naming
CREATE INDEX idx_user_profile_display_id ON user_profile(display_id);
```

#### 1.6. Database Function Responsibility
**Ensure database functions have single responsibility, use triggers whenever possible. For example, in a follow user feature, the user's follow_count should be updated by triggers, not included in the follow_user database function.**

```sql
-- Good: Single responsibility function
create or replace function public.follow_user(
    p_follower_id uuid,
    p_following_id uuid
)
returns void
language plpgsql
as $$
begin
    -- Only handle the core follow logic
    insert into public.user_follow (follower_id, following_id)
    values (p_follower_id, p_following_id);
    -- Count updates handled by triggers
end;
$$;

-- Separate trigger handles count updates
CREATE OR REPLACE  TRIGGER trg_update_follow_counts
  AFTER INSERT OR UPDATE OR DELETE ON user_follow
  FOR EACH ROW EXECUTE FUNCTION update_user_follow_counts();
```

#### 1.7. Database Function Return Types
**Database functions should avoid returning JSON, especially when only caring about operation success/failure and no data needs to be returned, prefer returning void.**

```sql
-- Good: void return for operation success/failure
create or replace function public.unfollow_user(
    p_follower_id uuid,
    p_following_id uuid
)
returns void
language plpgsql
as $$
begin
    delete public.user_follow 
    where follower_id = p_follower_id 
    and following_id = p_following_id;
    
    if not found then
        raise exception 'Follow relationship not found';
    end if;
end;
$$;

-- Good: simple type when data must be returned
create or replace function public.create_post(
    p_user_id uuid,
    p_title text,
    p_content text
)
returns uuid
language plpgsql
as $$
declare
    v_post_id uuid;
begin
    insert into public.post (user_id, title, content)
    values (p_user_id, p_title, p_content)
    returning id into v_post_id;
    
    return v_post_id;
end;
$$;
```

### 2. Query Organization
- **Simple Queries**: Group in separate files (e.g., `users.sql`, `posts.sql`) in `sql/queries/`
- **Database Tables**: Each table as independent SQL file in `supabase/schemas/`
- **Database Functions**: Each function as independent SQL file in `supabase/schemas/function/`
- **Triggers**: Each trigger as independent SQL file in `supabase/schemas/trigger/`
- **Meaningful Names**: Use descriptive names for queries, functions, and triggers
- **Follow PostgreSQL Naming**: Use `set_`/`trg_` prefixes for auxiliary objects

### 3. Performance Optimization
- **Query Annotations**: Use appropriate SQLC annotations (`:one` vs `:many`)
- **Database Functions**: Leverage PostgreSQL's performance for complex operations
- **Indexes**: Use database indexes for frequently queried columns
- **Pagination**: Always use `LIMIT` and `OFFSET` for pagination
- **JSON Assembly**: Use `jsonb_build_object/array` for nested data in single queries
- **Named Parameters**: pgx NamedArgs improve query plan caching and reduce parsing overhead compared to dynamic SQL string concatenation

### 4. Error Handling (SQL First)
- **Service Layer**: Handle `sql.ErrNoRows` appropriately in service methods
- **Database Functions**: Use proper exception handling with meaningful messages
- **Context Handling**: Use proper context for query cancellation
- **Function Errors**: Let database functions handle business rule violations
- **Constraint Violations**: Use database constraints for data integrity enforcement

### 5. Type Safety and Validation
- **SQLC Validation**: Leverage SQLC's compile-time SQL validation
- **Database Constraints**: Use database constraints for data integrity
- **Function Parameters**: Validate inputs in database functions
- **Null Handling**: Use appropriate null types in Go code
- **UUID Usage**: Consistent UUID usage for all primary keys


### 6. Row Level Security (RLS) Best Practices

#### 6.1 Mandatory RLS Usage
- **Prefer RLS over Application-Level Validation**: Always use PostgreSQL Row Level Security (RLS) for access control instead of manual validation in service layers
- **Enable RLS on All User Data Tables**: Every table containing user data must have RLS enabled
- **Comprehensive Policy Coverage**: Create separate policies for SELECT, INSERT, UPDATE, and DELETE operations

#### 6.2 RLS Policy Principles
- **Principle of Least Privilege**: Users should only access data they own or are explicitly authorized to access
- **Clear Policy Names**: Use descriptive policy names that explain the access control logic
- **Consistent auth.uid() Usage**: Always use `auth.uid()` function to get the current user ID
- **Performance Optimization**: Use `(select auth.uid())` wrapped in SELECT for better query plan caching

#### 6.3 Standard RLS Patterns

**Basic Ownership Pattern:**
```sql
-- Enable RLS on table
alter table my_table enable row level security;

-- Users can only access their own records
create policy "Users can view their own records" on my_table
for select
to authenticated
using ((select auth.uid()) = user_id);

create policy "Users can insert their own records" on my_table
for insert
to authenticated
with check ((select auth.uid()) = user_id);

create policy "Users can update their own records" on my_table
for update
to authenticated
using ((select auth.uid()) = user_id)
with check ((select auth.uid()) = user_id);

create policy "Users can delete their own records" on my_table
for delete
to authenticated
using ((select auth.uid()) = user_id);
```

**Complex Ownership Pattern (Many-to-Many):**
```sql
-- For association tables like post_oc
create policy "Users can create associations for their own resources" on post_oc
for insert
to authenticated
with check (
  exists (
    select 1 from post 
    where id = post_id 
      and user_id = auth.uid() 
      and deleted_at is null
  )
  and
  exists (
    select 1 from oc 
    where id = oc_id 
      and creator_id = auth.uid() 
      and deleted_at is null
  )
);
```

**Public Read Pattern:**
```sql
-- Allow public reading but restrict modifications to owners
create policy "Records are viewable by everyone" on my_table
for select
to authenticated, anon
using (true);

create policy "Users can only modify their own records" on my_table
for all
to authenticated
using ((select auth.uid()) = user_id)
with check ((select auth.uid()) = user_id);
```

#### 6.4 RLS vs Application Logic
- **Database-First Security**: RLS provides security at the database level, preventing data leaks even if application logic has bugs
- **Remove Application Validation**: Once RLS policies are in place, remove corresponding validation logic from service layers
- **Trust the Database**: Let PostgreSQL handle access control; application code should focus on business logic
- **Error Handling**: RLS violations will result in database errors; handle these gracefully in application error handlers

## Declarative Schema Best Practices

### 1. Schema Organization Principles
- **Independent Files**: Each table, function, and trigger is defined in its own SQL file
- **Modular Structure**: Organize schema files by business domain or logical grouping
- **Single Responsibility**: Each schema file should focus on a specific aspect (one table, one function, one trigger)
- **Idempotent Definitions**: Use `CREATE OR REPLACE` and `IF NOT EXISTS` for safe repeated application
- **Complete Definitions**: Include all constraints, indexes, and permissions in declarative files
- **Clear File Naming**: Use descriptive file names that clearly indicate the contained object (e.g., `user_profile.sql`, `follow_user.sql`)

### 3. Schema Definition Best Practices
- **Explicit Dependencies**: Clearly define object dependencies and load order
- **Version Comments**: Include version comments in schema files for tracking
- **Drop-and-Create Pattern**: For complex changes, use explicit drops followed by creates
- **Validation Functions**: Include schema validation functions to verify integrity
