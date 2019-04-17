DROP TABLE IF EXISTS hunts CASCADE;
CREATE TABLE hunts (
    id              serial,
    title           varchar(255) NOT NULL,
    max_teams       smallint  NOT NULL CONSTRAINT positive_num_teams CHECK (max_teams > 0),
    start_time      timestamp NOT NULL,
    /* @TODO see about constraining end_time to be after start_time */
    end_time        timestamp NOT NULL,
    latitude        real NOT NULL,
    longitude       real NOT NULL,
    location_name   varchar(80),
    PRIMARY KEY(id)
);

/*
    one to many: Hunt has many teams
*/
DROP TABLE IF EXISTS teams CASCADE;
CREATE TABLE teams (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL CHECK (length(name) > 0),
    CONSTRAINT teams_in_same_hunt_name UNIQUE(hunt_id, name),
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE

);

/*
    one to many: Hunt has many items
*/
DROP TABLE IF EXISTS items CASCADE;
CREATE TABLE items (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL CHECK (length(name) > 0),
    points          smallint CHECK (points > 0),
    CONSTRAINT items_in_same_hunt_name UNIQUE(hunt_id, name),
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE 
);