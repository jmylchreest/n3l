package controllers

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/xmlpath.v1"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func fetchobject(url string) (resp *http.Response, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}

	log.Printf("[client] requesting: %s", url)
	resp, err = netClient.Get(url)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Request for %s failed. Error: %s", url, err)
	}

	return resp, err
}

func Fetch(w http.ResponseWriter, req *http.Request) {
	var object string

	vars := mux.Vars(req)
	group := strings.Replace(vars["group"], ".", "/", -1)
	version := vars["version"]
	ext := vars["extension"]

	if "latest" == strings.ToLower(version) {
		// Latest version so we first need to pull the main maven-metadata.xml
		object = "maven-metadata.xml"
		url := "http://" + vars["host"] + "/" + "repository/" + vars["repo"] + "/" + group + "/" + vars["artifact"] + "/" + object

		resp, _ := fetchobject(url)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		path := xmlpath.MustCompile("/metadata/versioning/latest")
		x := bytes.NewReader(body)
		root, err := xmlpath.Parse(x)
		if err != nil {
			log.Fatal(err)
		}

		if value, ok := path.String(root); ok {
			version = value
		}
	}

	// Now we have the version, we can resolve the object
	object = version + "/maven-metadata.xml"

	// Latest version so we first need to pull the main maven-metadata.xml
	url := "http://" + vars["host"] + "/" + "repository/" + vars["repo"] + "/" + group + "/" + vars["artifact"] + "/" + object
	resp, _ := fetchobject(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	path := xmlpath.MustCompile("/metadata/versioning/snapshotVersions/snapshotVersion/value")
	x := bytes.NewReader(body)
	root, err := xmlpath.Parse(x)
	if err != nil {
		log.Fatal(err)
	}

	value, _ := path.String(root)
	log.Print("[resolver] matched version:", value)

	// Now fetch the war file, we presume value matches all versions, probably incorrect but right for us
	object = version + "/" + vars["artifact"] + "-" + value + "." + ext

	url = "http://" + vars["host"] + "/" + "repository/" + vars["repo"] + "/" + group + "/" + vars["artifact"] + "/" + object
	resp, _ = fetchobject(url)
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)

	w.Header().Set("ETag", resp.Header.Get("ETag"))
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.Itoa(binary.Size(body)))

	r := render.New()
	r.Data(w, http.StatusOK, body)
}
