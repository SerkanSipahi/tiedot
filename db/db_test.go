package db

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	TEST_DATA_DIR = "/tmp/tiedot_test"
)

func touchFile(dir, filename string) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(path.Join(dir, filename), make([]byte, 0), 0600); err != nil {
		panic(err)
	}
}

func TestOpenEmptyDB(t *testing.T) {
	os.RemoveAll(TEST_DATA_DIR)
	defer os.RemoveAll(TEST_DATA_DIR)
	db, err := OpenDB(TEST_DATA_DIR)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create("a"); err != nil {
		t.Fatal(err)
	}
	if len(db.cols) != 1 {
		t.Fatal(db.cols)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenCloseDB(t *testing.T) {
	os.RemoveAll(TEST_DATA_DIR)
	defer os.RemoveAll(TEST_DATA_DIR)
	if err := os.MkdirAll(TEST_DATA_DIR, 0700); err != nil {
		t.Fatal(err)
	}
	touchFile(TEST_DATA_DIR+"/ColA", "dat")
	touchFile(TEST_DATA_DIR+"/ColA", "a!b!c")
	if err := os.MkdirAll(TEST_DATA_DIR+"/ColB", 0700); err != nil {
		panic(err)
	}
	db, err := OpenDB(TEST_DATA_DIR)
	if err != nil {
		t.Fatal(err)
	}
	if db.path != TEST_DATA_DIR || db.cols["ColA"] == nil || db.cols["ColB"] == nil {
		t.Fatal(db.cols)
	}
	colA := db.cols["ColA"]
	colB := db.cols["ColB"]
	if colA.indexPaths["a!b!c"][0] != "a" || colA.indexPaths["a!b!c"][1] != "b" || colA.indexPaths["a!b!c"][2] != "c" {
		t.Fatal(colA.indexPaths)
	}
	if colA.hts["a!b!c"] == nil {
		t.Fatal(colA.hts)
	}
	if len(colB.indexPaths) != 0 || len(colB.hts) != 0 {
		t.Fatal(colB)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestColCrud(t *testing.T) {
	os.RemoveAll(TEST_DATA_DIR)
	defer os.RemoveAll(TEST_DATA_DIR)
	if err := os.MkdirAll(TEST_DATA_DIR, 0700); err != nil {
		t.Fatal(err)
	}
	db, err := OpenDB(TEST_DATA_DIR)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.AllCols()) != 0 {
		t.Fatal(db.AllCols())
	}
	// Create
	if err := db.Create("a"); err != nil {
		t.Fatal(err)
	}
	if db.Create("a") == nil {
		t.Fatal("Did not error")
	}
	if err := db.Create("b"); err != nil {
		t.Fatal(err)
	}
	// Get all names & use
	if allNames := db.AllCols(); len(allNames) != 2 || !(allNames[0] == "a" && allNames[1] == "b") {
		t.Fatal(allNames)
	}
	if db.Use("a") == nil || db.Use("b") == nil || db.Use("abcde") != nil {
		t.Fatal(db.cols)
	}
	// Rename
	if db.Rename("a", "a") == nil {
		t.Fatal("Did not error")
	}
	if db.Rename("a", "b") == nil {
		t.Fatal("Did not error")
	}
	if db.Rename("abc", "b") == nil {
		t.Fatal("Did not error")
	}
	if err := db.Rename("a", "c"); err != nil {
		t.Fatal(err)
	}
	if err := db.Rename("b", "d"); err != nil {
		t.Fatal(err)
	}
	// Rename - verify
	if allNames := db.AllCols(); len(allNames) != 2 || !(allNames[0] == "c" && allNames[1] == "d") {
		t.Fatal(allNames)
	}
	if db.Use("c") == nil || db.Use("d") == nil || db.Use("a") != nil {
		t.Fatal(db.cols)
	}
	// Truncate
	if db.Truncate("a") == nil {
		t.Fatal("Did not error")
	}
	if err := db.Truncate("c"); err != nil {
		t.Fatal(err)
	}
	if err := db.Truncate("d"); err != nil {
		t.Fatal(err)
	}
	// Truncate - verify
	if allNames := db.AllCols(); len(allNames) != 2 || !(allNames[0] == "c" && allNames[1] == "d") {
		t.Fatal(allNames)
	}
	if db.Use("c") == nil || db.Use("d") == nil || db.Use("a") != nil {
		t.Fatal(db.cols)
	}
	// Drop
	if db.Drop("a") == nil {
		t.Fatal("Did not error")
	}
	if err := db.Drop("c"); err != nil {
		t.Fatal(err)
	}
	if allNames := db.AllCols(); len(allNames) != 1 || allNames[0] != "d" {
		t.Fatal(allNames)
	}
	if db.Use("d") == nil {
		t.Fatal(db.cols)
	}
	if err := db.Drop("d"); err != nil {
		t.Fatal(err)
	}
	if allNames := db.AllCols(); len(allNames) != 0 {
		t.Fatal(allNames)
	}
	if db.Use("d") != nil {
		t.Fatal(db.cols)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDumpDB(t *testing.T) {
	os.RemoveAll(TEST_DATA_DIR)
	os.RemoveAll(TEST_DATA_DIR + "bak")
	defer os.RemoveAll(TEST_DATA_DIR)
	defer os.RemoveAll(TEST_DATA_DIR + "bak")
	if err := os.MkdirAll(TEST_DATA_DIR, 0700); err != nil {
		t.Fatal(err)
	}
	db, err := OpenDB(TEST_DATA_DIR)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create("a"); err != nil {
		t.Fatal(err)
	} else if err := db.Create("b"); err != nil {
		t.Fatal(err)
	}
	id1, err := db.Use("a").Insert(map[string]interface{}{"whatever": "1"})
	if err != nil {
		t.Fatal(err)
	} else if err := db.Dump(TEST_DATA_DIR + "bak"); err != nil {
		t.Fatal(err)
	}
	// Open the new database
	db2, err := OpenDB(TEST_DATA_DIR + "bak")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if allCols := db2.AllCols(); !(allCols[0] == "a" && allCols[1] == "b") {
		t.Fatal(allCols)
	}
	if doc, err := db2.Use("a").Read(id1); err != nil || doc["whatever"].(string) != "1" {
		t.Fatal(doc, err)
	}
	if err := db2.Close(); err != nil {
		t.Fatal(err)
	}
}
