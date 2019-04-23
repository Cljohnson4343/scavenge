package db

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
)

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

var ()

var stmtMap = map[string]*sql.Stmt{}

var scriptMap = map[string]string{
	"itemSelect":             itemSelectScript,
	"itemDelete":             itemDeleteScript,
	"itemInsert":             itemInsertScript,
	"itemsSelect":            itemsSelectScript,
	"teamSelect":             teamSelectScript,
	"teamDelete":             teamDeleteScript,
	"teamInsert":             teamInsertScript,
	"teamsSelect":            teamsSelectScript,
	"teamsWithHuntIDSelect":  teamsWithHuntIDSelectScript,
	"huntSelect":             huntSelectScript,
	"huntDelete":             huntDeleteScript,
	"huntInsert":             huntInsertScript,
	"huntsSelect":            huntsSelectScript,
	"locationsForTeamSelect": locationsForTeamScript,
}

func initStatements(db *sql.DB) error {
	var err error

	for k, script := range scriptMap {
		stmtMap[k], err = db.Prepare(script)
		if err != nil {
			return err
		}
	}

	return nil
}
