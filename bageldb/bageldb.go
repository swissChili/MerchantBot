package bageldb

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"time"
)

type User struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	UID       int           `bson:"UID"`
	Bagels    int           `bson:"Bagels,omitempty"`
	CheckedIn time.Time     `bson:"CheckedIn,omitempty"`
}

func UpdateUserBagels(db *mgo.Session, uid int, amount int, sender int) bool {
	dbsession := db.Copy()
	defer dbsession.Close()

	c := dbsession.DB("merchantbot").C("users")

	//var user User
	fmt.Println("Updating User Bagels")
	var results *User

	c.Find(bson.M{"UID": uid}).One(&results)
	fmt.Println("USER BAGELS:", GetUserBagels(db, sender), "SEND AMOUNT: ", amount)
	fmt.Println("===========================================")
	fmt.Println("USER B TYPE:", reflect.TypeOf(GetUserBagels(db, sender)))
	fmt.Println("USER B TYPE:", reflect.TypeOf(amount))
	if GetUserBagels(db, sender) >= amount {
		fmt.Println("nil")
		if results != nil {
			originalBagels := results.Bagels
			fmt.Println("Bagels:", originalBagels)

			newBagels := originalBagels + amount
			fmt.Println("New Bagels:", newBagels)
			writeToDb(db, uid, newBagels)

		} else {
			// set up endpoint
			writeToDb(db, uid, amount)
			fmt.Println("Recipient endpoint not set up, initializing")
		}

		senderOldBagels := GetUserBagels(db, sender)
		senderNewbagels := senderOldBagels - amount
		fmt.Println("\n\n#### SENDER NEW BAGELS", senderNewbagels)
		return writeToDb(db, sender, senderNewbagels)

	} else {
		fmt.Println("EVALUATED FALASE: UPDATEUSERBAGELS()")
		return false
	}

}

// returns 0 if it has been over 23 hours hours since
// last work time
func checkTimeSinceWork(t time.Time) int {
	diff := time.Since(t)
	hrs := int(diff.Hours())
	fmt.Println("~~ HRS ~~", hrs)
	if hrs >= 23 {
		return 0
	} else {
		return 23 - hrs
	}
}

// returns 0 on success, n if not enough time has passed
// where n is the # of hours remaining
// returns -1 uppon error
func Work(db *mgo.Session, uid int) int {
	dbsession := db.Copy()
	defer dbsession.Close()

	c := dbsession.DB("merchantbot").C("users")

	fmt.Println("User's UIDS:")
	var results *User
	// sets a variable to the result from the database
	fmt.Println("Right before DB check")
	c.Find(bson.M{"UID": uid}).One(&results)
	fmt.Println("Right after DB check")

	if results != nil {
		fmt.Println("UID from db:", results.UID)
		fmt.Println("UID from call:", uid)
		if reflect.TypeOf(results.CheckedIn) == reflect.TypeOf(time.Now()) {
			checkTime := checkTimeSinceWork(results.CheckedIn)
			if checkTime == 0 {
				// write to the db with the new DateTime
				newBagels := GetUserBagels(db, uid) + 10

				user := User{
					UID:       uid,
					Bagels:    newBagels,
					CheckedIn: time.Now(),
				}
				err := c.Update(bson.M{"UID": uid}, &user)
				fmt.Println("Passed final if and c.Update")
				if err != nil {
					fmt.Println(err)
				}
			}
			return checkTime
		} else {
			return createNewEndpointForWork(db, uid)
		}
	} else {
		return createNewEndpointForWork(db, uid)
	}
}

func createNewEndpointForWork(db *mgo.Session, uid int) int {
	dbsession := db.Copy()
	defer dbsession.Close()
	c := dbsession.DB("merchantbot").C("users")

	newBagels := GetUserBagels(db, uid) + 10

	// create a new endpoint with the current DateTime
	user := User{
		UID:       uid,
		Bagels:    newBagels,
		CheckedIn: time.Now(),
	}

	err := c.Insert(user)

	if err != nil {
		return -1
	} else {
		return 0
	}
}

func writeToDb(db *mgo.Session, uid int, amount int) bool {
	dbsession := db.Copy()
	defer dbsession.Close()

	c := dbsession.DB("merchantbot").C("users")

	fmt.Println("User's UIDS:")
	var results []User
	// sets a variable to the result from the database
	c.Find(bson.M{"UID": uid}).All(&results)
	for i := 0; i < len(results); i++ {
		fmt.Println(results[i].UID)
	}

	// creates the new object that will be used to update
	// or create the new database entry

	if len(results) == 0 || results == nil {
		fmt.Println("No users found!")
		fmt.Println("Making endpoint")
		objId := bson.NewObjectId()
		newUserData := User{
			ID:     objId,
			UID:    uid,
			Bagels: amount,
		}

		// creates the new database entry
		err := c.Insert(newUserData)

		if err != nil {
			fmt.Println("Error on add:", err)
		} else {
			fmt.Println("User was successfully added")
		}
		c.Find(bson.M{"UID": uid}).All(&results)
	}
	// updates the database entry
	userData := User{
		UID:    uid,
		Bagels: amount,
	}
	c.Update(bson.M{"UID": uid}, &userData)

	c.Find(bson.M{"UID": uid}).All(&results)

	fmt.Println("UID:", results[0].UID)
	fmt.Println("Updated Bagels:", results[0].Bagels)

	if results[0].Bagels == amount {
		fmt.Println("BAGELS FROM WRITE TO DB:", results[0].Bagels)
		return true
	} else {
		fmt.Println("BAGELS FROM WRITE TO DB:", results[0].Bagels)
		return false
	}

}

func GetUserBagels(db *mgo.Session, uid int) int {
	c := db.DB("merchantbot").C("users")

	fmt.Println("Checking User's Bagels")
	var results []User
	// sets a variable to the result from the database
	c.Find(bson.M{"UID": uid}).All(&results)

	if results == nil {
		return 0
	} else {
		return results[0].Bagels
	}

}

// nicely converts everything to an interface
func (q *User) ConvertToInterface() interface{} {
	return map[string]interface{}{
		"UID":    q.Bagels,
		"Bagels": q.UID,
	}
}
