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
	var extension string
	var classifier string
	var suffix string
	var xpath string

	vars := mux.Vars(req)
	group := strings.Replace(vars["group"], ".", "/", -1)
	version := vars["version"]

	// http://maven.apache.org/ref/3.0/maven-repository-metadata/repository-metadata.html

	// Parse the top-level maven-manifest.xml
	if strings.ToLower(version) == "latest" || strings.ToLower(version) == "release" {
		// Latest version so we first need to pull the main maven-metadata.xml
		object = "maven-metadata.xml"
		url := "http://" + vars["host"] + "/" + "repository/" + vars["repo"] + "/" + group + "/" + vars["artifact"] + "/" + object

		resp, _ := fetchobject(url)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		xpath = "/metadata/versioning/" + strings.ToLower(version)
		path := xmlpath.MustCompile(xpath)
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

	// And now lets build up the correct object to refer to
	xpath = "/metadata/versioning/snapshotVersions/snapshotVersion"
	var xpathadd = ""

	// Did we pass in the extension?
	if val, ok := vars["extension"]; ok {
		extension = val
		xpathadd = "extension=\"" + val + "\""
	} else {
		extension = "war"
	}

	// Did we pass in the classifier?

	if val, ok := vars["classifier"]; ok {
		classifier = "-" + val
		xpathadd = xpathadd + " and classifier=\"" + val + "\""
		// xmlpath.v2 is broken, it doesnt support multiple predicates despite claiming
		// it does, https://github.com/go-xmlpath/xmlpath/issues/24. So we just
		// match on classifier if it exists.
		xpathadd = "classifier=\"" + val + "\""
	} else {
		classifier = ""
	}

	// setup the fetch suffix
	suffix = classifier + "." + extension
	if len(xpathadd) > 0 {
		xpathadd = "[" + xpathadd + "]"
	}

	xpath = xpath + xpathadd + "/value"
	// xpath = "//snapshotVersion" + xpathadd + "/value"

	// And now parse that using xpath
	log.Printf("[resolver] searching using xpath %s", xpath)

	// Read the response body into an xml object
	x := bytes.NewReader(body)
	root, err := xmlpath.Parse(x)
	if err != nil {
		log.Fatal(err)
	}
	path := xmlpath.MustCompile(xpath)
	value, _ := path.String(root)
	log.Printf("[resolver] matched: %s%s", value, suffix)

	// So just fetch that object
	object = version + "/" + vars["artifact"] + "-" + value + suffix
	url = "http://" + vars["host"] + "/" + "repository/" + vars["repo"] + "/" + group + "/" + vars["artifact"] + "/" + object
	resp, _ = fetchobject(url)
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)

	w.Header().Set("ETag", resp.Header.Get("ETag"))
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.Itoa(binary.Size(body)))

	r := render.New()
	r.Data(w, resp.StatusCode, body)
}
