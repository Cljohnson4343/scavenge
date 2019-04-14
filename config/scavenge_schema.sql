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
CREATE TABLE teams (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE

);

/*
    one to many: Hunt has many items
*/
CREATE TABLE items (
    id              serial,
    hunt_id         int NOT NULL,
    name            varchar(255) NOT NULL,
    points          smallint CHECK (points > 0),
    PRIMARY KEY(id),
    FOREIGN KEY (hunt_id) REFERENCES hunts(id) ON DELETE CASCADE 
);