CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rooms (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	room_code VARCHAR(6) UNIQUE NOT NULL,
	status VARCHAR(50) DEFAULT 'create',
	settings JSONB NOT NULL DEFAULT '{}',
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	username VARCHAR(20) UNIQUE NOT NULL,
	is_guest BOOLEAN DEFAULT FALSE,
	is_online BOOLEAN DEFAULT FALSE
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