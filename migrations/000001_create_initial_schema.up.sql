-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Roles Table: Stores user roles (e.g., Admin, User, Uploader)
CREATE TABLE "roles" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar(50) UNIQUE NOT NULL
);

-- Permissions Table: Stores granular permissions
CREATE TABLE "permissions" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "code" varchar(100) UNIQUE NOT NULL -- e.g., 'manga:create', 'users:read'
);

-- Roles_Permissions Junction Table: Links roles to their permissions
CREATE TABLE "roles_permissions" (
  "role_id" uuid NOT NULL REFERENCES "roles" ("id") ON DELETE CASCADE,
  "permission_id" uuid NOT NULL REFERENCES "permissions" ("id") ON DELETE CASCADE,
  PRIMARY KEY ("role_id", "permission_id")
);

-- Users Table: Stores user information
CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "username" varchar(50) UNIQUE NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "password_hash" varchar NOT NULL,
  "role_id" uuid,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE SET NULL
);

-- Seed initial data
-- 1. Create Roles
INSERT INTO roles (id, name) VALUES
('a4198182-a398-4244-9635-5b58f3286d79', 'Admin'),
('d29a08e3-40a2-43e5-8f6a-1e6ca3a2f7a9', 'User');

-- 2. Create Permissions
INSERT INTO permissions (code) VALUES
('manga:create'), ('manga:read'), ('manga:update'), ('manga:delete'),
('chapters:upload'),
('users:read'), ('users:manage-roles');

-- 3. Assign Permissions to Roles
-- Admin gets all permissions
INSERT INTO roles_permissions (role_id, permission_id)
SELECT 'a4198182-a398-4244-9635-5b58f3286d79', id FROM permissions;

-- User gets basic reading permissions
INSERT INTO roles_permissions (role_id, permission_id)
SELECT 'd29a08e3-40a2-43e5-8f6a-1e6ca3a2f7a9', id FROM permissions WHERE code IN ('manga:read');