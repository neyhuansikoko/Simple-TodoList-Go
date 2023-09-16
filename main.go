package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	ID       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	TaskType int32              `bson:"taskType"`
	IsDone   bool               `bson:"isDone"`
}

type ListViewModel struct {
	DocTitle  string
	ListTitle string
	TaskType  int32
	Tasks     []Task
}

func (model ListViewModel) SetTodayActive() string {
	if model.TaskType == 0 {
		return "btn-dark"
	} else {
		return "btn-outline-dark"
	}
}

func (model ListViewModel) SetWorkActive() string {
	if model.TaskType == 1 {
		return "btn-dark"
	} else {
		return "btn-outline-dark"
	}
}

var coll *mongo.Collection
var tmpl *template.Template

func handleTodayList(w http.ResponseWriter, r *http.Request) {
	var results []Task
	cursor, err := coll.Find(context.TODO(), bson.D{{Key: "taskType", Value: 0}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	todayList := ListViewModel{
		DocTitle:  "Today Task",
		ListTitle: time.Now().Format("02 Jan 2006"),
		TaskType:  0,
		Tasks:     results,
	}

	if err = tmpl.ExecuteTemplate(w, "list", todayList); err != nil {
		panic(err)
	}
}

func handleWorkList(w http.ResponseWriter, r *http.Request) {
	var results []Task
	cursor, err := coll.Find(context.TODO(), bson.D{{Key: "taskType", Value: 1}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	workList := ListViewModel{
		DocTitle:  "Work Task",
		ListTitle: "Work Task",
		TaskType:  1,
		Tasks:     results,
	}

	if err = tmpl.ExecuteTemplate(w, "list", workList); err != nil {
		panic(err)
	}
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	taskId := r.FormValue("id")
	taskObjectId, err := primitive.ObjectIDFromHex(taskId)

	if err != nil {
		panic(err)
	}

	taskType, err := strconv.Atoi(r.FormValue("taskType"))

	if err != nil {
		panic(err)
	}

	taskStatus := r.FormValue("isDone")

	_, err = coll.UpdateByID(
		context.TODO(),
		taskObjectId,
		bson.D{{Key: "$set", Value: bson.D{{Key: "isDone", Value: !(taskStatus == "true")}}}},
	)

	if err != nil {
		panic(err)
	}

	redirectToPage(w, r, taskType)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	taskType, err := strconv.Atoi(r.FormValue("taskType"))

	if err != nil {
		panic(err)
	}

	taskName := r.FormValue("name")

	_, err = coll.InsertOne(context.TODO(), bson.D{{Key: "name", Value: taskName}, {Key: "taskType", Value: taskType}})

	if err != nil {
		panic(err)
	}

	redirectToPage(w, r, taskType)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	taskType, err := strconv.Atoi(r.FormValue("taskType"))

	if err != nil {
		panic(err)
	}

	_, err = coll.DeleteMany(context.TODO(), bson.D{{Key: "isDone", Value: true}})

	if err != nil {
		panic(err)
	}

	redirectToPage(w, r, taskType)
}

func redirectToPage(w http.ResponseWriter, r *http.Request, taskType int) {
	switch taskType {
	case 0:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	case 1:
		http.Redirect(w, r, "/work", http.StatusSeeOther)
	}
}

func main() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	uri := os.Getenv("DB_URI")

	fmt.Println("Connecting to DB...")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Connected to DB")
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll = client.Database("todolistDB").Collection("tasks")

	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("public/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	tmpl = template.Must(template.ParseGlob("views/*.html"))
	template.Must(tmpl.ParseGlob("views/partials/*.html"))

	r.HandleFunc("/", handleTodayList).Methods("GET")
	r.HandleFunc("/work", handleWorkList).Methods("GET")
	r.HandleFunc("/check", handleCheck).Methods("POST")
	r.HandleFunc("/submit", handleSubmit).Methods("POST")
	r.HandleFunc("/delete", handleDelete).Methods("POST")

	fmt.Println("Starting server. Listening at port:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
