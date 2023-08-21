package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"netflix-watchlist/controller/mongodb"
	"netflix-watchlist/model"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const movieCollName = "movies"

func homePage() {
	doc := model.Movies{Name: "KGF", Rating: 5}
	result, err := mongodb.MovieColl.InsertOne(context.TODO(), doc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", result.InsertedID)
}
func HomePage(w http.ResponseWriter, r *http.Request) {
	homePage()
	w.Write([]byte("Welcome to the HomePage!"))
}

func getAllMovies() []primitive.M {
	filter := bson.D{}
	cur, err := mongodb.MovieColl.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Error on Finding all the documents", err)
		panic(err)
	}
	defer cur.Close(context.Background())
	var result []primitive.M
	for cur.Next(context.Background()) {
		var doc bson.M
		err := cur.Decode(&doc)
		if err != nil {
			panic(err)
		}
		result = append(result, doc)
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
	return result
}

func GetAllMovies(w http.ResponseWriter, r *http.Request) {
	result := getAllMovies()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getMovie(id primitive.ObjectID) primitive.M {
	filter := bson.M{"_id": id}
	var result primitive.M
	err := mongodb.MovieColl.FindOne(context.Background(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No document found")
		return result
	}
	if err != nil {
		panic(err)
	}
	return result
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	result := getMovie(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createMovie(movie model.Movies) *mongo.InsertOneResult {
	result, err := mongodb.MovieColl.InsertOne(context.Background(), movie)
	if err != nil {
		fmt.Println("Error on inserting new Movie", err)
		panic(err)
	}
	return result
}

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie model.Movies
	_ = json.NewDecoder(r.Body).Decode(&movie)
	if movie.Name == "" {
		fmt.Println("name is a required field")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "name is a required field")
		return
	}
	result := createMovie(movie)
	json.NewEncoder(w).Encode(result)
}

// delete a movie
func deleteMovie(id primitive.ObjectID) *mongo.DeleteResult {
	filter := bson.M{"_id": id}
	result, err := mongodb.MovieColl.DeleteOne(context.Background(), filter)
	// condition if delete count is 0
	if result.DeletedCount == 0 {
		fmt.Println("No document found")
		return result
	}
	if err != nil {
		fmt.Println("Error on deleting one Movie", err)
		panic(err)
	}
	return result
}

func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	result := deleteMovie(id)
	json.NewEncoder(w).Encode(result)
}

func updateMovieRating(id primitive.ObjectID, rating int) *mongo.UpdateResult {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"rating": rating}}
	result, err := mongodb.MovieColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println("Error on updating one Movie", err)
		panic(err)
	}
	return result
}

func UpdateMovieRating(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var movie model.Movies
	_ = json.NewDecoder(r.Body).Decode(&movie)
	if movie.Rating == 0 {
		http.Error(w, "Rating is missing or invalid", http.StatusBadRequest)
		return
	}
	result := updateMovieRating(id, movie.Rating)
	json.NewEncoder(w).Encode(result)
}

func getAllDirectors() []primitive.M {
	matchStage := bson.D{{"$match", bson.D{}}}
	lookupStage := bson.D{
		{"$lookup", bson.D{
			{"from", "countries"}, // Replace with actual countries collection name
			{"localField", "nationalityId"},
			{"foreignField", "_id"},
			{"as", "nationalityDetails"},
		}},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 1},                // Include _id field if needed
			{"name", 1},               // Include other fields you need
			{"nationalityDetails", 1}, // Include nationalityDetails field
		}},
	}
	cursor, err := mongodb.DirectorColl.Aggregate(context.Background(), mongo.Pipeline{matchStage, lookupStage, projectStage})
	if err != nil {
		fmt.Println("Error on Aggregation", err)
		panic(err)
	}
	defer cursor.Close(context.Background())
	var result []bson.M
	if err := cursor.All(context.Background(), &result); err != nil {
		fmt.Println("Error decoding cursor", err)
		panic(err)
	}
	return result
}

func GetAllDirectors(w http.ResponseWriter, r *http.Request) {
	result := getAllDirectors()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getDirector(id primitive.ObjectID) primitive.M {
	matchStage := bson.D{{"$match", bson.D{{"_id", id}}}}
	lookupStage := bson.D{
		{"$lookup", bson.D{
			{"from", "countries"}, // Replace with actual countries collection name
			{"localField", "nationalityId"},
			{"foreignField", "_id"},
			{"as", "nationalityDetails"},
		}},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 1},                // Include _id field if needed
			{"name", 1},               // Include other fields you need
			{"nationalityDetails", 1}, // Include nationalityDetails field
		}},
	}
	cursor, err := mongodb.DirectorColl.Aggregate(context.Background(), mongo.Pipeline{matchStage, lookupStage, projectStage})
	if err != nil {
		fmt.Println("Error on Aggregation", err)
		panic(err)
	}
	defer cursor.Close(context.Background())
	var result bson.M
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&result); err != nil {
			fmt.Println("Error decoding cursor", err)
			panic(err)
		}
	} else {
		fmt.Println("No document found")
	}
	return result
}

func GetDirector(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	result := getDirector(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createDirector(director model.Directors) *mongo.InsertOneResult {
	result, err := mongodb.DirectorColl.InsertOne(context.Background(), director)
	if err != nil {
		fmt.Println("Error on inserting new Director", err)
		panic(err)
	}
	return result
}

func CreateDirector(w http.ResponseWriter, r *http.Request) {
	var director model.Directors
	_ = json.NewDecoder(r.Body).Decode(&director)
	if director.Name == "" {
		fmt.Println("name is a required field")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "name is a required field")
		return
	}
	countryId := director.NationalityID
	if countryId.IsZero() {
		fmt.Println("country is a required field")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "country is a required field")
		return
	}
	country := getCountry(countryId)
	if country == nil {
		fmt.Println("country not found")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "country not found")
		return
	}
	director.NationalityID = countryId
	result := createDirector(director)
	json.NewEncoder(w).Encode(result)
}

func getAllCountries() []primitive.M {
	filter := bson.D{}
	cur, err := mongodb.CountryColl.Find(context.Background(), filter)
	if err != nil {
		fmt.Println("Error on Finding all the documents", err)
		panic(err)
	}
	defer cur.Close(context.Background())
	var result []primitive.M
	for cur.Next(context.Background()) {
		var doc bson.M
		err := cur.Decode(&doc)
		if err != nil {
			panic(err)
		}
		result = append(result, doc)
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
	return result
}

func GetAllCountries(w http.ResponseWriter, r *http.Request) {
	result := getAllCountries()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getCountry(id primitive.ObjectID) primitive.M {
	filter := bson.M{"_id": id}
	var result primitive.M
	err := mongodb.CountryColl.FindOne(context.Background(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No document found")
		return result
	}
	if err != nil {
		panic(err)
	}
	return result
}

func GetCountry(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	result := getCountry(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createCountry(country model.Countries) *mongo.InsertOneResult {
	result, err := mongodb.CountryColl.InsertOne(context.Background(), country)
	if err != nil {
		fmt.Println("Error on inserting new Country", err)
		panic(err)
	}
	return result
}

func CreateCountry(w http.ResponseWriter, r *http.Request) {
	var country model.Countries
	_ = json.NewDecoder(r.Body).Decode(&country)
	if country.Name == "" {
		fmt.Println("name is a required field")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "name is a required field")
		return
	}
	result := createCountry(country)
	json.NewEncoder(w).Encode(result)
}
