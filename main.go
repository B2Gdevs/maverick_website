package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"maverick_website/models"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var host = "localhost"
var port = ":8110"

// parse files and have templated html files created.
var temps, _ = template.ParseGlob("./templates/*.html")

// Connect to database and get collection to write and read to/from
var databaseConnection = "mongodb://localhost:27017"
var client = ConnectMongo(databaseConnection)
var collection = client.Database("Maverick").Collection("Website")

// Static folder
var staticLoc = "./static"
var staticURL = "./resources"

func copyFile(src string, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		fmt.Printf("Something wrong with %s.", src)
		return err
	}

	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	err = copyContent(src, dst)
	return err
}

func copyContent(src string, dst string) error {
	fout, err := os.Create(dst)
	defer fout.Close()
	if err != nil {
		fmt.Printf("Couldn't create file. %s", dst)
		log.Println(err)
	}

	fin, err := os.Open(src)
	if err != nil {
		fmt.Printf("Couldn't open file. %s", src)
		log.Println(err)
	}
	defer fin.Close()

	_, err = io.Copy(fout, fin)
	if err != nil {
		fmt.Printf("Could not copy %s to %s", src, dst)
		log.Println(err)
	}
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("http://" + host + port + "/Media")
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error occured trying to get media in handler")
		log.Println(err)
	}
	defer res.Body.Close()
	Media := []models.Media{}
	json.Unmarshal(body, &Media)
	data := struct {
		Data []models.Media
	}{
		Data: Media,
	}

	temps.ExecuteTemplate(w, "index.html", data)
}

func GetAllMedia(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, bson.M{})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	}
	media := []models.Media{}
	temp := models.Media{}
	for cur.Next(ctx) {
		cur.Decode(&temp)
		media = append(media, temp)
	}

	json.NewEncoder(w).Encode(media)
}

func GetAllVideos(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, bson.M{"filetype": "video"})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	}
	media := []models.Media{}
	temp := models.Media{}
	for cur.Next(ctx) {
		cur.Decode(&temp)
		media = append(media, temp)
	}

	json.NewEncoder(w).Encode(media)
}

func GetAllPhotos(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, bson.M{"filetype": "photo"})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	}
	media := []models.Media{}
	temp := models.Media{}
	for cur.Next(ctx) {
		cur.Decode(&temp)
		media = append(media, temp)
	}

	json.NewEncoder(w).Encode(media)
}

func AddVideo(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := models.Media{}
	json.NewDecoder(r.Body).Decode(&data)
	data.IsVideo = true
	temp := data.FilePath
	data.FilePath = path.Join(staticURL, filepath.Base(data.FilePath))
	if err := copyFile(temp, path.Join(staticLoc, filepath.Base(data.FilePath))); err != nil {
		log.Println("error when copying file.")
		log.Println(err)
	}

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		fmt.Println("yolo")
		log.Println(err)
	}
}

func AddPhoto(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := models.Media{}
	json.NewDecoder(r.Body).Decode(&data)
	data.IsPhoto = true
	temp := data.FilePath
	data.FilePath = path.Join(staticURL, filepath.Base(data.FilePath))
	if err := copyFile(temp, path.Join(staticLoc, filepath.Base(data.FilePath))); err != nil {
		log.Println("error when copying file.")
		log.Println(err)
	}
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		fmt.Println("yolo")
		log.Println(err)
	}
}

func ConnectMongo(connection string) *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection))
	if err != nil {
		log.Println(err)
	}
	return client
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	r.HandleFunc("/Media", GetAllMedia).Methods("GET")
	r.HandleFunc("/Videos", GetAllVideos).Methods("GET")
	r.HandleFunc("/Videos", AddVideo).Methods("POST")
	r.HandleFunc("/Photos", GetAllPhotos).Methods("GET")
	r.HandleFunc("/Photos", AddPhoto).Methods("POST")
	r.PathPrefix("/resources/").Handler(http.StripPrefix("/resources", http.FileServer(http.Dir(staticLoc))))

	log.Fatal(http.ListenAndServe(host+port, r))

}
