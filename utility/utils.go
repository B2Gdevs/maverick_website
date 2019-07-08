package utility

import (
	"io/ioutil"
	"net/http"
)

func Remove(index int, slice []interface{}) []interface{} {
	return append(slice[0:index], slice[index:len(slice)-1]...)
}

func CheckContentType(filePath string) string {
	clientFile, _ := ioutil.ReadFile(filePath) // or get your file from a file system
	return http.DetectContentType(clientFile)
}
