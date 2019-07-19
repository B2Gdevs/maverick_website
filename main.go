package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"maverick_website/models"
	"maverick_website/utility"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	jsonData, _ = utility.GetDebugParams("app_config.json")
	host        = utility.GetEnv("HOST", jsonData["HOST"])
	port        = utility.GetEnv("PORT", jsonData["PORT"])
	mongoPass   = utility.GetEnv("MONGOPASS", jsonData["MONGOPASS"])

	databaseConnection = fmt.Sprintf("mongodb+srv://Ben:%s@maverick-yqcfp.mongodb.net/test?retryWrites=true&w=majority", url.QueryEscape(mongoPass))
	client             = ConnectMongo(databaseConnection)
	collection         = client.Database("Maverick").Collection("Website")

	sess, _ = session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			utility.GetEnv("AWS_ACCESS_KEY_ID", jsonData["AWS_ACCESS_KEY_ID"]),
			utility.GetEnv("AWS_SECRET_ACCESS_KEY", jsonData["AWS_SECRET_ACCESS_KEY"]),
			""),
	})
	uploader = s3manager.NewUploader(sess)

	temps, _ = template.ParseGlob("./templates/*.html")

	// Static folder
	staticLoc = utility.GetEnv("S3URL", jsonData["S3URL"])
)

func uploadFile(file multipart.File, fileHandle *multipart.FileHeader) {
	uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utility.GetEnv("S3BUCKET", jsonData["S3BUCKET"])),
		Key:    aws.String("/media/" + fileHandle.Filename),
		Body:   file,
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	var res *http.Response
	var err error
	if host == "localhost" {
		res, err = http.Get("http://" + host + ":" + port + "/Media")
	} else {
		res, err = http.Get(host + "/Media")
	}
	if err == nil {
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
	} else {
		log.Println("error in handler: ")
		log.Print(err)
	}

}

func GetAllMedia(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	// returns nil in cur if ip isn't whitelisted at the database.
	cur, err := collection.Find(nil, bson.M{})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	} else {
		media := []models.Media{}
		temp := models.Media{}
		for cur.Next(ctx) {
			cur.Decode(&temp)
			media = append([]models.Media{temp}, media...)
		}

		json.NewEncoder(w).Encode(media)
	}
}

func GetAllVideos(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, bson.M{"filetype": "video"})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	} else {
		media := []models.Media{}
		temp := models.Media{}
		for cur.Next(ctx) {
			cur.Decode(&temp)
			media = append([]models.Media{temp}, media...)
		}
		json.NewEncoder(w).Encode(media)
	}
}

func GetAllPhotos(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cur, err := collection.Find(ctx, bson.M{"filetype": "photo"})
	defer cur.Close(ctx)
	if err != nil {
		fmt.Print("error getting files")
		log.Println(err)
	} else {
		media := []models.Media{}
		temp := models.Media{}
		for cur.Next(ctx) {
			cur.Decode(&temp)
			media = append([]models.Media{temp}, media...)
		}
		json.NewEncoder(w).Encode(media)
	}
}

func AddMedia(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024*1024) // 16MB
	r.ParseMultipartForm(16 << 20)
	file, handle, err := r.FormFile("CustomFile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("POST request had no file attached."))
	} else {
		defer file.Close()
		fileBytes, _ := ioutil.ReadAll(file)

		if !filetype.IsVideo(fileBytes) && !filetype.IsImage(fileBytes) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("POST request had no image or video."))
		} else {
			data := models.Media{
				Title:       r.Form["Title"][0],
				Description: r.Form["Description"][0],
				FilePath:    "http://" + path.Join(staticLoc, handle.Filename),
				IsPhoto:     filetype.IsImage(fileBytes),
				IsVideo:     filetype.IsVideo(fileBytes),
				Date:        time.Now().Format("01-02-2006"),
			}
			err = ioutil.WriteFile("temp", fileBytes, 777)
			if err != nil {
				log.Println(err)
			}
			f, _ := os.Open("temp")
			uploadFile(f, handle)
			f.Close()
			os.Remove("temp")

			_, err = collection.InsertOne(ctx, data)
			if err != nil {
				log.Println(err)
			}
		}

	}

}

func DeleteMedia(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := models.Media{}
	json.NewDecoder(r.Body).Decode(&data)
	collection.DeleteOne(ctx, bson.M{"filepath": data.FilePath})
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
	r.HandleFunc("/Media", DeleteMedia).Methods("DELETE")
	r.HandleFunc("/Media", AddMedia).Methods("POST")
	r.HandleFunc("/Videos", GetAllVideos).Methods("GET")
	r.HandleFunc("/Photos", GetAllPhotos).Methods("GET")
	r.PathPrefix("/resources/").Handler(http.StripPrefix("/resources", http.FileServer(http.Dir("./static"))))

	log.Fatal(http.ListenAndServe(":"+port, r))

}
