package main

import (
	"fmt"
	"net/http"
	"sync"
	"html/template"
)

func WriteTemplate(w http.ResponseWriter) (int, error) {
	return w.Write([]byte(`
<html>
	<head>
		<title> Chat app </title>
		<p> hi there this is mat ryer here </p>
	</head>
</html>`))
}

func firstWay() {
	http.ListenAndServe("127.0.0.1:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteTemplate(w)
	}))
}

func secondWay() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		WriteTemplate(w)
	})
	http.ListenAndServe(":7002", nil)
}

type handler struct {
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	WriteTemplate(w)
}

func thirdWay() {
	handlerStruct := handler{
	}
	http.ListenAndServe(":9001", handlerStruct)
}

func fourthWay() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/apis", func(writer http.ResponseWriter, request *http.Request) {
		WriteTemplate(writer)
	})
	http.ListenAndServe(":6002", serverMux)
}

type handlerFuncType http.HandlerFunc

func fifthWay() {
	hFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteTemplate(w)
	})
	http.ListenAndServe(":6002", hFunc)
}

//s can be a http.handler which would then mean that
func wrapper(s string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(s)
		WriteTemplate(w)
	})
}
func sixthWay() {
	http.ListenAndServe(":9000", wrapper("a"))
}

type templating struct {
	fileName string
	once     *sync.Once
	templ    *template.Template
	User     User
}

func (t templating) ServeHTTP(w http.ResponseWriter, r *http.Request) {
			fmt.Print(t.templ.Execute(w, &t.User))

}

type User struct {
	Name string
}

func seventhWithTemplate() {
	t := template.Must(template.New("index.gohtml").ParseFiles("./template/index.gohtml"))
	templating := templating{
		fileName: "index.gohtml",
		templ:    t,
		User: User{
			Name: "akkshay",
		},
	}
	http.ListenAndServe(":9006", templating)
}

func main() {
	//firstWay()
	//secondWay()
	//thirdWay()
	//fourthWay()
	//fifthWay()
	//sixthWay()
	seventhWithTemplate()
}
