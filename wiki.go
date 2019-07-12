package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

type PageList struct {
	Fields []*Page
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "tmpl/frontpage.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var frontPage = "FrontPage"

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile("data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile("data/" + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func getAllPages() []string {
	var files []string

	root := "data/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		fileName := info.Name()
		if strings.Contains(fileName, ".txt") && !strings.Contains(fileName, frontPage) {
			files = append(files, fileName[:len(fileName)-4])
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return files
}

func viewFrontPageHandler(w http.ResponseWriter, r *http.Request) {
	pages := getAllPages()
	var listOfPages []*Page
	if pages != nil {
		for _, title := range pages {
			p, err := loadPage(title)
			if err == nil {
				listOfPages = append(listOfPages, p)
			}
		}
	}
	data := PageList{Fields: listOfPages}
	err := templates.ExecuteTemplate(w, "frontpage.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	if title == frontPage {
		http.Redirect(w, r, "/view/"+frontPage, http.StatusFound)
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("title")
	fmt.Println(body)
	if len(body) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	body = strings.ReplaceAll(body, " ", "")
	fmt.Println(body)

	http.Redirect(w, r, "/edit/"+body, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/", viewFrontPageHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
