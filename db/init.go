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
}

var ()

var stmtMap = map[string]*sql.Stmt{}

var scriptMap = map[string]string{
	"itemSelect":            itemSelectScript,
	"itemDelete":            itemDeleteScript,
	"itemInsert":            itemInsertScript,
	"itemsSelect":           itemsSelectScript,
	"teamSelect":            teamSelectScript,
	"teamDelete":            teamDeleteScript,
	"teamInsert":            teamInsertScript,
	"teamsSelect":           teamsSelectScript,
	"teamsWithHuntIDSelect": teamsWithHuntIDSelectScript,
	/*	"huntSelect":            huntSelectScript,
		"huntDelete":            huntDeleteScript,
		"huntInsert":            huntInsertScript,
		"huntsSelect":           huntsSelectScript,
	*/
}

func initStatements(db *sql.DB) error {
	var err error
	/*
		// items statements
		itemSelectStmnt, err = db.Prepare(itemSelectScript)
		if err != nil {
			return err
		}

		itemInsertStmnt, err = db.Prepare(itemInsertScript)
		if err != nil {
			return err
		}

		itemDeleteStmnt, err = db.Prepare(itemDeleteScript)
		if err != nil {
			return err
		}

		itemsSelectStmnt, err = db.Prepare(itemsSelectScript)
		if err != nil {
			return err
		}
	*/
	for k, script := range scriptMap {
		stmtMap[k], err = db.Prepare(script)
		if err != nil {
			return err
		}
	}
	/*
		teamInsertStmnt, err = db.Prepare(teamInsertScript)
		if err != nil {
			return err
		}

		teamDeleteStmnt, err = db.Prepare(teamDeleteScript)
		if err != nil {
			return err
		}

		teamsSelectStmnt, err = db.Prepare(teamsSelectScript)
		if err != nil {
			return err
		}

		teamsWithHuntIDSelectStmnt, err = db.Prepare(teamsWithHuntIDSelectScript)
		if err != nil {
			return err
		}
		/*
			// hunts statements
			huntSelectStmnt, err = db.Prepare(huntSelectScript)
			if err != nil {
				return err
			}

			huntInsertStmnt, err = db.Prepare(huntInsertScript)
			if err != nil {
				return err
			}

			huntDeleteStmnt, err = db.Prepare(huntDeleteScript)
			if err != nil {
				return err
			}

			huntsSelectStmnt, err = db.Prepare(huntsSelectScript)
			if err != nil {
				return err
			}
	*/
	return nil
}
