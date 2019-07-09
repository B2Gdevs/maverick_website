package utility

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Remove(index int, slice []interface{}) []interface{} {
	return append(slice[0:index], slice[index:len(slice)-1]...)
}

func CheckContentType(filePath string) string {
	clientFile, _ := ioutil.ReadFile(filePath) // or get your file from a file system
	return http.DetectContentType(clientFile)
}

func GetDebugParams(debugPath string) (map[string]string, error) {
	jsonData := make(map[string]string)
	fin, err := os.Open(debugPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	jsonBytes, err := ioutil.ReadAll(fin)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	json.Unmarshal([]byte(jsonBytes), &jsonData)

	return jsonData, nil
}

func GetEnv(key string, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	return value
}
