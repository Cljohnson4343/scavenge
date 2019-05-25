package db

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
)

var db *sql.DB

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
	govalidator.TagMap["positive"] = govalidator.Validator(func(str string) bool {
		v, err := strconv.Atoi(str)
		if err != nil {
			return false
		}

		return v > 0
	})
	govalidator.TagMap["isNil"] = govalidator.Validator(func(str string) bool {
		return false
	})
	govalidator.CustomTypeTagMap.Set("startTimeBeforeEndTime", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := context.(type) {
		case HuntDB:
			return v.StartTime.Before(v.EndTime)
		}

		return false
	}))
	govalidator.CustomTypeTagMap.Set("isZeroTime", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := i.(type) {
		case time.Time:
			return v.IsZero()
		}
		return false
	}))
	govalidator.CustomTypeTagMap.Set("timeNotPast", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := i.(type) {
		case time.Time:
			return v.After(time.Now())
		}
		return false
	}))
	govalidator.CustomTypeTagMap.Set("timePast", govalidator.CustomTypeValidator(func(i interface{}, context interface{}) bool {
		switch v := i.(type) {
		case time.Time:
			return v.Before(time.Now())
		}
		return false
	}))
}

var stmtMap = map[string]*sql.Stmt{}

var scriptMap = map[string]string{
	"huntInvitationsByUserID": huntInvitationsByUserIDScript,
	"huntInvitationDelete":    huntInvitationDeleteScript,
	"huntInvitationInsert":    huntInvitationInsertScript,
	"huntGetByCreatorAndName": huntGetByCreatorAndNameScript,
	"huntSelect":              huntSelectScript,
	"huntDelete":              huntDeleteScript,
	"huntInsert":              huntInsertScript,
	"huntsSelect":             huntsSelectScript,
	"itemSelect":              itemSelectScript,
	"itemDelete":              itemDeleteScript,
	"itemInsert":              itemInsertScript,
	"itemsSelect":             itemsSelectScript,
	"locationsForTeam":        locationsForTeamScript,
	"locationInsert":          locationInsertScript,
	"locationDelete":          locationDeleteScript,
	"mediaMetasForTeam":       mediaMetasForTeamScript,
	"mediaMetaInsert":         mediaMetaInsertScript,
	"mediaMetaDelete":         mediaMetaDeleteScript,
	"permissionInsert":        permissionInsertScript,
	"permissionsForUser":      permissionsForUserScript,
	"playersGetForHunt":       playersGetForHuntScript,
	"roleInsert":              roleInsertScript,
	"roleRemove":              roleRemoveScript,
	"rolesDeleteByRegex":      rolesDeleteByRegexScript,
	"rolesForUser":            rolesForUserScript,
	"sessionInsert":           sessionInsertScript,
	"sessionGetForUser":       sessionGetForUserScript,
	"sessionGet":              sessionGetScript,
	"sessionDelete":           sessionDeleteScript,
	"teamSelect":              teamSelectScript,
	"teamDelete":              teamDeleteScript,
	"teamInsert":              teamInsertScript,
	"teamsSelect":             teamsSelectScript,
	"teamsWithHuntIDSelect":   teamsWithHuntIDSelectScript,
	"teamPoints":              teamPointsScript,
	"teamAddPlayer":           teamAddPlayerScript,
	"teamRemovePlayer":        teamRemovePlayerScript,
	"teamGetPlayers":          teamGetPlayersScript,
	"userInsert":              userInsertScript,
	"userGet":                 userGetScript,
	"userGetByUsername":       userGetByUsernameScript,
	"userDelete":              userDeleteScript,
}

func initStatements(database *sql.DB) error {
	var err error

	db = database

	for k, script := range scriptMap {
		stmtMap[k], err = database.Prepare(script)
		if err != nil {
			return err
		}
	}

	return nil
}
