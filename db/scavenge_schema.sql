
DROP TABLE IF EXISTS item_results CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS hunts CASCADE;

/*
    This table represents a scavenger hunt game. 'hunts' contains
    the meta info for a hunt. The location stored in this table
    is not stored in 'locations' because this data is meta data
    that is not associated with a specific team while playing 
    a hunt. 

    relations:
        one to many--hunts have many teams
        one to many--hunts have many items
*/
CREATE TABLE hunts (
    id              serial,
    name            varchar(255) NOT NULL,
    max_teams       smallint  NOT NULL CONSTRAINT positive_num_teams CHECK (max_teams > 1),
    start_time      timestamp NOT NULL,
    /* @TODO see about constraining end_time to be after start_time */
    end_time        timestamp NOT NULL,
    latitude        real NOT NULL,
    longitude       real NOT NULL,
    location_name   varchar(80),
    created_at      timestamp DEFAULT NOW(),
    PRIMARY KEY(id)
);

/*
    This table represents a team for a specific hunt.

    relations:
        many to one--teams can have the same hunt
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
    This table contains the route info for a team during a
    specific hunt. Each entry represents a location from 
    a team while the team was playing a hunt.

    relations:
        many to one--locations can have the same team
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
    This table is used to store the team specific item info i.e
    whether the team has any media associated with a specific item.
    this table will be how a client can tell if a team has found 
    a specific item. Each row represents a team's finding of an
    item in a specific hunt. There will be an associated 'locations'
    entry for each row.

    relations:
        many to one--item results can have the same team
        one to one--item results will have one location
        one to one--item results will have one item
*/ 
CREATE TABLE item_results (
    id              serial,
    team_id         int NOT NULL,
    item_id         int NOT NULL,
    location_id     int NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);

CREATE INDEX items_huntid_asc ON items(hunt_id ASC);
CREATE INDEX teams_huntid_asc ON teams(hunt_id ASC);
CREATE INDEX item_results_teams_asc ON item_results(team_id ASC);
CREATE INDEX locations_teamid_asc ON locations(team_id ASC, time_stamp ASC);