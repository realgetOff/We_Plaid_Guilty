CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_type AS ENUM('guest', 'standard', 'api42');

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	username VARCHAR(20) UNIQUE NOT NULL,
	email TEXT UNIQUE,
	password_hash TEXT,
	type user_type DEFAULT 'standard',
	CONSTRAINT guest_auth CHECK (
		(type = 'guest' AND email IS NULL AND password_hash IS NULL)
		OR
		(type = 'api42' AND email IS NOT NULL AND password_hash IS NULL)
		OR
		(type = 'standard' AND email IS NOT NULL AND password_hash IS NOT NULL)
	)
	--created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE friendship_status AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE IF NOT EXISTS friends (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	requester_id UUID NOT NULL,
	addressee_id UUID NOT NULL,
	status friendship_status DEFAULT 'pending',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

	-- Foreign keys linking back to your users table's UUID
	CONSTRAINT fk_requester
		FOREIGN KEY(requester_id) 
		REFERENCES users(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_addressee
		FOREIGN KEY(addressee_id) 
		REFERENCES users(id)
		ON DELETE CASCADE,

	-- Prevent a user from sending a friend request to themselves
	CONSTRAINT check_not_self_friend
		CHECK (requester_id != addressee_id),

	-- Prevent duplicate requests between the same two users in the same direction
	CONSTRAINT unique_friendship
		UNIQUE (requester_id, addressee_id)
);

CREATE TABLE IF NOT EXISTS profiles (
	id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, -- using the UUID from users as the primary key for the profile
	display_name VARCHAR(32),
	avatar_url VARCHAR(256), -- URL to the avatar image file
	color VARCHAR(7) DEFAULT '#000000', -- Allows hexadecimal RGB colours like #008800 to be stored
	font VARCHAR(6) DEFAULT 'normal' -- bold, italic, normal
);

-- CREATE EXTENSION IF NOT EXISTS pg_cron;

-- SELECT cron.schedule(
-- 	'guest-cleanup', '0 */4 * * *', $$
-- 	DELETE FROM users
-- 	WHERE type = 'guest'
-- 	AND created_at < NOW() - INTERVAL '4 hours';
-- $$)