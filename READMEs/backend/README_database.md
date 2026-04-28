# Backend - Database - ft_transcendence · We Plaid Guilty

---

## Description

This project uses PostgreSQL (henceforth referred to as pSQL) as the relational data store, to access the database with the golang backend, we use a driver / library called `pgx` to send SQL queries to handle users, profiles, and the friends list.

### Why "pgx"?

The standard way to communicate with a relational database in golang is using "database/sql". Here however, we opted for pgx to take advantage of a couple different pSQL features.
Specifically, performance, type safety and a simpler concurrency in regards to goroutines.

### Connection Management

We utilize Connection Pools to ensure the backend can handle multiple concurrent game sessions without exhausting database resources. This ensures that every drawing submission and chat message is persisted reliably and quickly.

## Contents

The database itself contains three tables:
- Registered users | `users`
	This table contains an id primary key (a UUID generated with the "uuid-ossp" extension), a user's username, their hashed password, the type of user (defined by a pSQL enum), and their time of creation.
	There's a constraint that ensures that standard users have to be inserted alongside their email and hashed password, whereas guests shouldn't have a password associated with them, and 42api users have to be inserted with their 42 email.	
- Profile details | `profiles`
	This table references a users primary key, so when a user is deleted their profile is removed in tandem.
	It contains their display name, their selected chat color, and the font that their name displays with.
- A user's friends | `friends`
	This is a table where every entry is bidirectional, ensuring that if personA is a friend of personB, then personB is still counted as a friend of personA.
	When a user is deleted, all their associated `friends` entries are deleted.

## Necessary variables

Certain .env variables are necessary for the connection to the DB.
They are:
- DB_USER
	The user that can read and write to tables in the database.
	ex: `admin`
- DB_HOST
	The address on which the DB is hosted.
	ex: `localhost` or `db`
- DB_PORT
	The opened port allowing for communication on the DB_HOST address.
	ex: `6432`
- DB_PASSWORD
	The password needed to connect to the database.
	ex: `password`, `changeme`
- DB_NAME
	The name of the database.
	ex: `transcendence`

Seeing as we use ansible-vault to handle secrets / environment variables instead of a standard `.env` file, the way to modify these variables is as follows:
`ansible-vault edit ansible/group_vars/all/vault.yml`

