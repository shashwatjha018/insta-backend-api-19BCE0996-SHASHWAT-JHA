/*Made by Shashwat Jha*/

package main

//Importing all the necessary packages from the standard library
import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client  //Initiliazing the mongo client

//**************** MODELS START ************************

type User struct {
	Id	    primitive.ObjectID 		`json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string 				`json:"name,omitempty" bson:"name,omitempty"`
	Email      string				`json:"email,omitempty" bson:"email,omitempty"`
	Password   string				`json:"password,omitempty" bson:"password,omitempty"`
}

type Post struct {
	Id	   				 primitive.ObjectID 		`json:"_id,omitempty" bson:"_id,omitempty"`
	Caption      			  string 			    `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL    			  string				`json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	PostedTimestamp  	 primitive.DateTime		    `json:"_postedtimestamp,omitempty" bson:"_postedtimestamp,omitempty"`
	UserId 				 	  string 		        `json:"userid,omitempty" bson:"userid,omitempty"`
}

//**************** MODELS END ************************



//************** CONTROLLERS START *******************

func HomePage(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET"{
		fmt.Fprintf(w, "Instagram homepage endpoint hit!")
	}else{
		message := "Method not found"
		fmt.Fprint(w, message)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

//-------------------USER CONTROLLERS START----------------------------------

func CreateUser(w http.ResponseWriter, r *http.Request){  //POST USER
	if r.Method == "POST"{
		w.Header().Add("content-type", "application/json")
		var user User
		json.NewDecoder(r.Body).Decode(&user)
		collection := client.Database("Insta-API").Collection("user")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//Password Hashing
		data := []byte(user.Password)
		b := md5.Sum(data)
		user.Password = hex.EncodeToString(b[:])
		//Setting Cookie/Session
		result, _ := collection.InsertOne(ctx, user)
		//fmt.Println(result.InsertedID.(primitive.ObjectID).String())
		expiration := time.Now().Add(365*24*time.Hour)
		cookie := http.Cookie{Name: "user_id", Value: result.InsertedID.(primitive.ObjectID).Hex(), Expires: expiration}
		http.SetCookie(w, &cookie)
		json.NewEncoder(w).Encode(result) 
	}else{
		message := "Method not found. Enter user details via POST"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

func UserList(w http.ResponseWriter, r *http.Request){ //GET ALL THE USER (Extra Functionality added)
	if r.Method == "GET"{
		w.Header().Add("content-type", "application/json")
		var users []User
		collection := client.Database("Insta-API").Collection("user")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx){
			var user User
			cursor.Decode(&user)
			users = append(users, user)
		}
		if err := cursor.Err(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(users)
		
	}else{
		message := "Method not found. GET request is only acceptable"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

func GetUser(w http.ResponseWriter, r *http.Request){ //GET USER BY ID
	if r.Method == "GET"{
		w.Header().Add("content-type", "application/json")
		params := strings.TrimPrefix(r.URL.Path, "/users/")
		//fmt.Fprintln(w, params)
		id, _ := primitive.ObjectIDFromHex(params)
		var user User
		collection := client.Database("Insta-API").Collection("user")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := collection.FindOne(ctx, User{Id: id}).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(user)		
	}else{
		message := "Method not found. GET request is only acceptable"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

//-------------------USER CONTROLLERS END-------------------------------------------


//-------------------POSTS(INSTA) CONTROLLERS START----------------------------------
func CreatePost(w http.ResponseWriter, r *http.Request){ //POST INSTA POSTS
	if r.Method == "POST"{
		w.Header().Add("content-type", "application/json")
		var post Post
		json.NewDecoder(r.Body).Decode(&post)
		collection := client.Database("Insta-API").Collection("post")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		dt := time.Now();
		post.PostedTimestamp = primitive.NewDateTimeFromTime(dt)

		//Getting value of user_id
		for _, cookie := range r.Cookies() {
			//fmt.Fprint(w, cookie.Value)
			post.UserId = cookie.Value
		}
		result, _ := collection.InsertOne(ctx, post)
		json.NewEncoder(w).Encode(result) 
	}else{
		message := "Method not found. Enter post details via POST"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

func GetPost(w http.ResponseWriter, r *http.Request){ //GET INSTA POST BY ID
	if r.Method == "GET"{
		w.Header().Add("content-type", "application/json")
		params := strings.TrimPrefix(r.URL.Path, "/posts/")
		//fmt.Fprintln(w, params)
		id, _ := primitive.ObjectIDFromHex(params)
		var post Post
		collection := client.Database("Insta-API").Collection("post")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := collection.FindOne(ctx, User{Id: id}).Decode(&post)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(post)		
	}else{
		message := "Method not found. GET request is only acceptable"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}

func GetUserPost(w http.ResponseWriter, r *http.Request){ //GET ALL THE POST OF THE USERS
	if r.Method == "GET"{
		w.Header().Add("content-type", "application/json")
		var posts []Post
		params := strings.TrimPrefix(r.URL.Path, "/posts/users/")
		fmt.Fprintln(w, params)
		collection := client.Database("Insta-API").Collection("post")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//O/P With Pagination
		limit := 2
		page, begin := Pagination(r, limit)
		fmt.Printf("Current Page: %d, Begin: %d\n", page, begin)
		skip := int64(begin)
		limits := int64(2)
		options := options.Find()
		options.SetSkip(skip)
		options.SetLimit(limits)
		cursor, err := collection.Find(ctx, Post{UserId: params}, options)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)
		count := 0
		for cursor.Next(ctx){
			count += 1
			var post Post
			cursor.Decode(&post)
			posts = append(posts, post)
		}
		if err := cursor.Err(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `"}`))
			return
		}
		fmt.Fprintln(w,"Displayed results:",count)
		json.NewEncoder(w).Encode(posts)
		
	}else{
		message := "Method not found. GET request is only acceptable"
		fmt.Fprint(w, message)
		fmt.Fprint(w, http.StatusNotFound)
		fmt.Printf("%d", http.StatusFound)
	}
	time.Sleep(1 * time.Second)
}
//-------------------POSTS(INSTA) CONTROLLERS END----------------------------------


//************** CONTROLLERS END **************************************************


//--------------- PAGINATION FUNCTION START ---------------------------------------
func Pagination(r *http.Request, limit int) (int, int) {
	keys := r.URL.Query()
	if keys.Get("page") == ""{
		return 1, 0
	}
	page, _ := strconv.Atoi(keys.Get("page"))
	if page < 1{
		return 1, 0
	}
	begin := (limit * page) - limit
	return page, begin
}
//--------------- PAGINATION FUNCTION END ---------------------------------------

//--------------- ALL URL ROUTERS START ------------------------------------------
func handleRequests(){
	go http.HandleFunc("/", HomePage)
	go http.HandleFunc("/users", CreateUser)
	go http.HandleFunc("/userlist", UserList)
	go http.HandleFunc("/users/", GetUser)
	go http.HandleFunc("/posts", CreatePost)
	go http.HandleFunc("/posts/", GetPost)
	go http.HandleFunc("/posts/users/", GetUserPost)
	go log.Fatal(http.ListenAndServe(":8081", nil))

}
//--------------- ALL URL ROUTERS END ------------------------------------------

//--------------- MAIN FUNCTION STARTS ------------------------------------------
func main() {
	fmt.Println("Starting the API now")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	handleRequests()
}
//--------------- MAIN FUNCTION ENDS ------------------------------------------