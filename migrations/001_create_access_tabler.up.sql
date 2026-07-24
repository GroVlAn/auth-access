CREATE TABLE role
(
    id UUID PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text,
    is_default boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permission
(
    id UUID PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text,
    is_default boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions
(
    role_id UUID REFERENCES role(id) ON DELETE CASCADE NOT NULL,
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE NOT NULL,
    PRIMARY KEY(role_id, permission_id)
);

CREATE TABLE user_roles
(
    role_id UUID REFERENCES role(id) ON DELETE CASCADE NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY(role_id, user_id)
)

CREATE INDEX idx_role_permissions_role
ON role_permissions(role_id);

CREATE INDEX idx_role_permissions_permission
ON role_permissions(permission_id);

CREATE INDEX idx_user_roles_role
ON user_roles(role_id);

CREATE INDEX idx_user_roles_user
ON user_roles(user_id);

ALTER TABLE roles
ADD CONSTRAINT roles_name_key UNIQUE(name);