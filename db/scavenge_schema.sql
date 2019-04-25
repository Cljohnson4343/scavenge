DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users_teams CASCADE;
DROP TABLE IF EXISTS media CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS hunts CASCADE;
DROP TABLE IF EXISTS users CASCADE;

/*
    This table represents a user. 

    relations:
        one to many--hunts have many teams
        one to many--hunts have many items
        one to many--hunts will have one creator(user)
*/
CREATE TABLE users (
    id                  serial,
    first_name          text NOT NULL,
    last_name           text NOT NULL,
    username            varchar(64) NOT NULL,
    joined_at           timestamp DEFAULT NOW(),
    last_visit          timestamp DEFAULT NOW(),
    image_url           varchar(2083), 
    email               text NOT NULL,
    PRIMARY KEY(id)
);
CREATE UNIQUE INDEX users_unique_lower_email_idx ON users(lower(email));
CREATE UNIQUE INDEX users_unique_username ON users(lower(username));

/* 
    This table represents a user's sessions.

    relations:
        many to one--many sessions can have relsationships with the same user

*/
CREATE TABLE user_sessions (
    session_key         uuid,
    expires             timestamp NOT NULL,
    created_at          timestamp DEFAULT NOW(),
    user_id             int NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (session_key)
);

/*
    This table represents a scavenger hunt game. 'hunts' contains
    the meta info for a hunt. The location stored in this table
    is not stored in 'locations' because this data is meta data
    that is not associated with a specific team while playing 
    a hunt. 

    relations:
        one to many--hunts have many teams
        one to many--hunts have many items
        one to many--hunts will have one creator(user)
*/
CREATE TABLE hunts (
    id              serial,
    name            varchar(255) NOT NULL,
    max_teams       smallint  NOT NULL CONSTRAINT positive_num_teams CHECK (max_teams > 1),
    start_time      timestamp NOT NULL,
    end_time        timestamp NOT NULL,
    latitude        real NOT NULL,
    longitude       real NOT NULL,
    location_name   varchar(80),
    created_at      timestamp DEFAULT NOW(),
    creator_id      int NOT NULL,
    CONSTRAINT hunt_with_same_name UNIQUE(name),
    PRIMARY KEY(id),
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
);

/*
    This table represents a team for a specific hunt.

    relations:
        many to one--teams can have the same hunt
        one to many--teams can have many locations
        one to many--teams can have many media rows
        many to many--multiple teams can have relationships with multiple users
*/
CREATE TABLE teams (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL CHECK (length(name) > 0),
    CONSTRAINT teams_in_same_hunt_name UNIQUE(hunt_id, name),
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE
);

/*
    This is a joining table represents the many to many relationship between the 
    users and teams tables.

*/
CREATE TABLE users_teams (
    team_id         int NOT NULL,
    user_id         int NOT NULL,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

/*
    This table contains the route info for a team during a
    specific hunt. Each entry represents a location from 
    a team while the team was playing a hunt.

    relations:
        many to one--locations can have the same team
        one to one--locations can have a single media row
*/ 
CREATE TABLE locations (
    id              serial,
    team_id         int NOT NULL,
    latitude        real NOT NULL,
    longitude       real NOT NULL,
    time_stamp      timestamp NOT NULL,
    CONSTRAINT team_same_loc_and_time UNIQUE(time_stamp, team_id),
    PRIMARY KEY(id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

/*
    This table is used to store the items for each hunt. This 
    table does not contain any team specific info, that
    should be stored in the 'item_results' table.

    relations:
        many to one--items can have the same hunt
*/
CREATE TABLE items (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL CHECK (length(name) > 0),
    points          int DEFAULT 1 CHECK (points > 0),
    CONSTRAINT items_in_same_hunt_name UNIQUE(hunt_id, name),
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE 
);

/*
    This table is used to store the team specific media info.
    This table will be how a client can tell if a team has found 
    a specific item. Each row represents a media file associated
    with a specific team. If an item_id is provided, then that
    team has "found" that item. There will be an associated 'locations'
    entry for each row.

    relations:
        many to one--media rows can have the same team
        one to one--media rows will have one location
        one to one--media rows will at most one item
*/ 
CREATE TABLE media (
    id              serial,
    team_id         int NOT NULL,
    item_id         int,
    location_id     int NOT NULL,
    url             varchar(2083) NOT NULL CHECK (length(url) > 3),
    PRIMARY KEY(id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

CREATE INDEX items_huntid_asc ON items(hunt_id ASC);
CREATE INDEX teams_huntid_asc ON teams(hunt_id ASC);
CREATE INDEX media_teams_and_loc_asc ON media(team_id ASC, location_id ASC);
CREATE INDEX loc_teamid_asc ON locations(team_id ASC, time_stamp ASC);
