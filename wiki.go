package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// Page : Wiki Article Page
type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html", "templates/home.html", "templates/new.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	filename := "articles/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "articles/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	saveErr := p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
	if saveErr != nil {
		fmt.Printf("Save Failed: %s\n", saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("save: %s\n", p.Title)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")
	// fmt.Println()
	p := &Page{Title: title, Body: []byte(body)}
	saveErr := p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
	if saveErr != nil {
		fmt.Printf("Create Failed: %s\n", saveErr.Error())
		http.Error(w, saveErr.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("save: %s\n", p.Title)
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	newPage := &Page{Title: "Title", Body: []byte("Enter text here...")}
	t, _ := template.ParseFiles("templates/new.html")
	t.Execute(w, newPage)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	var title string
	if p != nil {
		title = p.Title
	} else {
		title = "Home"
	}
	fmt.Printf("%s: %s\n", tmpl, title)
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		fmt.Printf("Render Failed: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Printf("Invalid Path: %s\n", r.URL.Path)
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // Title will be second subexpression
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/edit/new", newHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/", homeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
