package hunts

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cljohnson4343/scavenge/hunts/models"
	"github.com/go-chi/chi"
)

// mockDB needs to implement the HuntDataStore interface
type mockDB struct {
	db []*models.Hunt
}

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Errorf("An error occurred: %v\n", err.Error())
	}

}

var testEnv mockDB
var mux *chi.Mux
var testDataBuffer []byte
var server *httptest.Server
var client *http.Client
var baseURL string

func TestMain(m *testing.M) {
	var err error
	testDataBuffer, err = ioutil.ReadFile("./test_data.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		panic(err)
	}

	err = json.Unmarshal(testDataBuffer, &testEnv.db)
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		panic(err)
	}

	mux = Routes(&testEnv)
	server = httptest.NewServer(mux)
	defer server.Close()

	client = server.Client()
	baseURL = fmt.Sprintf("%s/", server.URL)

	os.Exit(m.Run())
}

// This is just a mock function for testing purposes
func (env *mockDB) allHunts() ([]*models.Hunt, error) {
	return env.db, nil
}

func TestGetHuntsHandler(t *testing.T) {
	/*
		rr, err := client.Get(baseURL)
		checkErr(err, t)

		if rr.StatusCode != http.StatusOK {
			t.Errorf("Status code differs. Expected %d.\nGot %d instead.", http.StatusOK, rr.StatusCode)
		}

			var retHunts []*models.Hunt
			_, err = rr.Body.Read(resBuffer)
			if err != nil && err != io.EOF {
				checkErr(err, t)
			}
			defer rr.Body.Close()

			err = json.Unmarshal(resBuffer, &retHunts)

			if reflect.DeepEqual(retHunts, testEnv.db) != true {
				t.Errorf("Response body differs. Expected %v. \n Got %v\n", testEnv.db, retHunts)
			}
	*/
}

// This is just a mock function for testing purposes
func (env *mockDB) getHunt(hunt *models.Hunt, huntID int) error {
	log.Print("Entered getHunt")
	for _, v := range env.db {
		if v.ID == huntID {
			*hunt = *v
			copy(hunt.Items, v.Items)
			copy(hunt.Teams, v.Teams)

			return nil
		}
	}

	return errors.New("No Hunt")
}

func TestGetHuntHandler(t *testing.T) {
	/*
		log.Printf(fmt.Sprintf("%s/3", baseURL))
		rr, err := client.Get(fmt.Sprintf("%s/3", baseURL))
		checkErr(err, t)

		if rr.StatusCode != http.StatusOK {
			t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, rr.StatusCode)
		}

		if rr.ContentLength < 0 {
			t.Fatalf("response body was not returned")
		}

		var retHunts []*models.Hunt
		resBuffer := make([]byte, rr.ContentLength)
		_, err = rr.Body.Read(resBuffer)
		if err != nil && err != io.EOF {
			checkErr(err, t)
		}
		defer rr.Body.Close()

		err = json.Unmarshal(resBuffer, &retHunts)

		// test_data.json lists hunts in ascending order by id so we can
		// hard code the array index to retrieve the hunt_id == 3 hunt
		if reflect.DeepEqual(retHunts, testEnv.db[2]) != true {
			t.Errorf("Response body differs. Expected %v. \n Got %v\n", testEnv.db[2], retHunts)
		}
	*/
}

// This is just a mock function for testing purposes
func (env *mockDB) getItems(items *[]models.Item, huntID int) error {
	hunt := new(models.Hunt)
	err := env.getHunt(hunt, huntID)
	if err != nil {
		return err
	}

	for _, v := range hunt.Items {
		*items = append(*items, v)
	}

	return nil
}

// This is just a mock function for testing purposes
func (env *mockDB) getTeams(teams *[]models.Team, huntID int) error {
	return nil
}

// This is just a mock function for testing purposes
func (env *mockDB) insertHunt(hunt *models.Hunt) (int, error) {
	return 0, nil
}

// This is just a mock function for testing purposes
func (env *mockDB) insertTeam(team *models.Team, huntID int) (int, error) {
	return 0, nil
}

// This is just a mock function for testing purposes
func (env *mockDB) insertItem(item *models.Item, huntID int) (int, error) {
	return 0, nil
}

// This is just a mock function for testing purposes
func (env *mockDB) deleteHunt(huntID int) error {
	return nil
}
