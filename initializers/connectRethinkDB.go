package initializers

import (
	"log"
	"os"

	r "github.com/dancannon/gorethink"
)

var session *r.Session

// Connect initializes a new RethinkDB session and returns an error if there was a problem connecting.
func ConnectRethinkDB()  {
	var err error
	session, err = r.Connect(r.ConnectOpts{
		Address:  "localhost:28015",
		Database: "test",
	})
	if err != nil {
		log.Fatal("Failed to connect to the Database! \n", err.Error())
		os.Exit(1)
	}


	// Create tables and indexes here

	log.Println("✅ Connected to RethinkDB")

}

// Close terminates the RethinkDB session
func Close() error {
	return session.Close()
}

// GetSession returns the RethinkDB session
func GetSession() *r.Session {
	return session
}