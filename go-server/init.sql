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

CREATE TABLE IF NOT EXISTS profiles (
	id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, -- using the UUID from users as the primary key for the profile
	display_name VARCHAR(32),
	avatar_url VARCHAR(256), -- URL to the avatar image file
	color VARCHAR(7) DEFAULT '#000000', -- Allows hexadecimal RGB colours like #008800 to be stored
	font VARCHAR(6) DEFAULT 'normal' -- bold, italic, normal
);


-- INSERTION INTO FRIENDS
-- INSERT INTO friends (requester_id, addressee_id, status) 
-- VALUES (
--     'user_a_uuid', 
--     'user_b_uuid', 
--     'pending'
-- );

-- GETTING FRIENDS
-- SELECT 
--     users.id, 
--     users.username, 
--     users.is_online 
-- FROM users
-- JOIN friendships ON (users.id = friendships.requester_id OR users.id = friendships.addressee_id)
-- WHERE 
--     (friendships.requester_id = 'user_a_uuid' OR friendships.addressee_id = 'user_a_uuid')
--     AND users.id != 'user_a_uuid'
--     AND friendships.status = 'accepted';

-- REMOVING FRIENDS
-- DELETE FROM friendships 
-- WHERE (requester_id = 'user_a_uuid' AND addressee_id = 'user_b_uuid')
--    OR (requester_id = 'user_b_uuid' AND addressee_id = 'user_a_uuid');

-- if you already had a DB populated before friend requests used status = 'accepted'
-- in queries, existing rows may still be 'pending'. To keep them as accepted friends
-- after pulling this schema behavior, run manually once:
--   UPDATE friends SET status = 'accepted';
