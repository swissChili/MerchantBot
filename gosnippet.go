
database := session.DB("go_mongo")
collection := database.C("names")
/* collection.DropCollection()
collection = database.C("names") */
nameForQuery := QueryName{FirstName: first, LastName: last}
// query1 := collection.Find(bson.M{"myname": bson.M{"FirstName": first, "LastName": last}})
param := nameForQuery.ConvertToInterface()
query1 := collection.Find(bson.M{"myname": param})
count, _ := query1.Count()




name := Name{Id: bson.NewObjectId(), MyName: QueryName{FirstName: first, LastName: last}}
add_err := collection.Insert(name)  
if add_err != nil {
	fmt.Println("Error on add:", add_err)
} else {
	fmt.Println("Name was successfully added")
}
