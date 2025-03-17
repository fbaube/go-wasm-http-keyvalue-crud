//go:generate go tool wit-bindgen-go generate --world component --out gen ./wit
package main

// "incomplete results" from wrpc
// https://github.com/bytecodealliance/wrpc/blob/ca379336f5124a109358459e04d2b18587f87aee/crates/transport/src/invoke.rs#L208

import (
	"fmt"
	"io"
	"net/http"
	// S "strings"
	"encoding/json"

	// "database/sql"
	// _ "embed"
	// "embed"

	// Lots of fails...
	// _ "github.com/mattn/go-sqlite3"
	// _ "modernc.org/sqlite"
	// _ "gitlab.com/cznic/sqlite"
	// _ "github.com/ncruces/go-sqlite3/driver"
	// _ "github.com/ncruces/go-sqlite3/embed"


	// And then a desperation move...
	// SAVE THESE FOR LATER 
	// _ "github.com/fbaube/go-sqlite3-for-tinygo-wasm"
	// _ "github.com/fbaube/go-sqlite3-for-tinygo-wasm/driver"
	// _ "github.com/fbaube/go-sqlite3-for-tinygo-wasm/embed"

	// wasihttp does not like the new Go ServeMux, but it does 
	// work okay with this third-party router. Its README states:
	// In contrast to the default mux of Go's net/http package,
	// this router supports variables in the routing pattern
	// and matches against the request method. 
	"github.com/julienschmidt/httprouter"

	// For the keyvalue capability, we use bindings
	// for the wasi:keyvalue/store interface.
	store "github.com/wasmCloud/go/examples/component/http-keyvalue-crud/gen/wasi/keyvalue/store"
	// In the end we did not need this ugly hack, because when
	// things are working right, the compiler does locate gen:
	// store "github.com/fbaube/wc_go_http-keyvalue-crud_gen_store" 

	"go.bytecodealliance.org/cm"
	"go.wasmcloud.dev/component/net/wasihttp"
	"go.wasmcloud.dev/component/log/wasilog"
	"log/slog"	
)

/*
// EMBEDDED DATABASE
// https://github.com/mattn/go-sqlite3/issues/968
// https://github.com/mattn/go-sqlite3/pull/1188

//go:embed m5.db
var theDB []byte

//go:embed m5.db
var fsDB embed.FS 
*/

// Types for JSON validation.
type CheckRequest struct {
	Value string `json:"value"`
}

type CheckResponse struct {
	Valid   bool   `json:"valid"`
	Length  int    `json:"length,omitempty"`
	Message string `json:"message,omitempty"`
}

var router *httprouter.Router
var logger *slog.Logger

// init establishes the routes & methods for our K/V operations.
func init() {

        logger = wasilog.ContextLogger("DERF")
//	logger := slog.New(wasilog.DefaultOptions().NewHandler())
	logger.Info("Logging is initialized")
	logger.Info("Logging", "Name", "fubar", "Number", 42)
/*	
	_, e1 := sql.Open("sqlite3", "file://m5.db?mode=ro")
	_, e2 := sql.Open("sqlite3", "file://fsDB/m5.db?mode=ro")
	if e1 != nil { logger.Error("e1: " + e1.Error()) }
	if e2 != nil { logger.Error("e2: " + e2.Error()) }
	db, _ := sql.Open("sqlite3", "file:m5.db")
	var version int 
	db.QueryRow(`SELECT sqlite_version()`).Scan(&version)

	// This one seems to work, but then there's no way to convert
	// the []byte into a DB unless you write it to disk and read 
	// it in as a "file:/// URL, but this is not allowed (yet).
	logger.Info("theDB", "len", len(theDB))

	// This one works! This means we can almost definitely 
	// read it with SQLite, IF we can find a driver that 
	// can be compiled in by tinygo. 
	// 
	// func (f FS) ReadFile(name string) ([]byte, error)
	ff, _ := fsDB.Open("m5.db")
	fi, _ := ff.Stat()
	logger.Info("fsDB/db", "len", fi.Size())
	// if ee != nil { logger.Info("fsDB/db", "err", ee.Error()) }
*/
	router = httprouter.New()
	router.GET   ("/", 	    hINDEX)
	router.GET   ("/crud/:key", hGET)
	router.POST  ("/crud/:key", hPOST)
	router.DELETE("/crud/:key", hDELETE)
	wasihttp.Handle(router)
}

func hINDEX(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w,
     `{"message":"GET,POST,DELETE to /crud/<key> (w JSON payload for POSTs)"}`)

     fmt.Fprintf(w,
     "\nTry these: \n\n" +
     "curl -X POST localhost:8000/crud/mario -d '{\"itsa\": \"me\", \"woo\": \"hoo\"}' \n" +
     "curl localhost:8000/crud/mario \n" +
     "curl -X DELETE localhost:8000/crud/mario")
}

func hPOST(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
	key := ps.ByName("key")

	// Check the request for a valid JSON body and assign it to var "value".
	// The user will set the value via JSON payload:
	// curl -X POST 'localhost:8000/crud/key' -d '{"foo": "bar", "woo": "hoo"}'
	var req CheckRequest
	defer r.Body.Close()
	value, err := io.ReadAll(r.Body)
	if err != nil {
		errResponseJSON(w, http.StatusBadRequest,
			"io.ReadAll(r.Body): " + err.Error())
		return
	}
	if err := json.Unmarshal(value, &req); err != nil {
		errResponseJSON(w, http.StatusBadRequest,
				 "Request.json.Unmarshal: " + err.Error())
		return
	}

	// Opens the keyvalue bucket.
	// NOTE: wasm-tools 1.127.0 does not allow the assignment of
	// a return value here and requires: store.Open("default")
	kvStore := store.Open("default")
	fmt.Printf("store.Open: result: <%T> %#v \n", kvStore, kvStore)
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	} 

	// Converts the value to a byte array.
	valueBytes := []byte(value)

	// Converts the byte array to the Component Model's cm.List type.
	valueList := cm.ToList(valueBytes)

	// Set the value for the key in the current bucket and handle any errors.
	store.Bucket.Set(*kvStore.OK(), key, valueList)
	// store.Bucket.Set(key, valueList)
	kvSet := store.Bucket.Set(*kvStore.OK(), key, valueList)
	if kvSet.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, kvSet.Err().String())
		return
	} 

	// Confirms set, returning key and value in JSON body.
	kvSetMessage := fmt.Sprintf("Set %s", key)
	kvSetResponse := fmt.Sprintf(`{"message":"%s", "value":"%s"}`,
		kvSetMessage, value)
	fmt.Fprintln(w, kvSetResponse)

}

func hGET(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
	key := ps.ByName("key")

	// Opens the keyvalue bucket.
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	}

	// Gets the value for the defined key.
	kvGet, kvGetErr, kvGetIsErr := store.Bucket.Get(*kvStore.OK(), key).Result()

	// Returns and reports that key does not exist if no value is found.
	if kvGet.Value().Len() == 0 {
		errResponseJSON(w, http.StatusBadRequest, key+": does not exist")
		return
	}
	// Handles get errors other than non-existent key
	if kvGetIsErr {
		errResponseJSON(w, http.StatusBadRequest, kvGetErr.String())
		return
	}

	// Use cm.LiftString to convert the byte value into
	// a string, taking the data and len as arguments.
	kvGetJSON := cm.LiftString[string](kvGet.Value().Data(), kvGet.Value().Len())

	// Returns key and value in JSON body.
	kvGetMessage := fmt.Sprintf("Got %s", key)
	kvGetResponse := fmt.Sprintf(`{"message":"%s", "value":"%s"}`,
		kvGetMessage, kvGetJSON)
	fmt.Fprintln(w, kvGetResponse)

}

func hDELETE(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
	key := ps.ByName("key")

	// Opens the keyvalue bucket.
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	}
	// Returns and reports that key does not exist if no value is found.
	kvGet, _, _ := store.Bucket.Get(*kvStore.OK(), key).Result()
	if kvGet.Value().Len() == 0 {
		errResponseJSON(w, http.StatusBadRequest, key+": does not exist")
		return
	}
	// Deletes the entry for the provided key.
	kvDel := store.Bucket.Delete(*kvStore.OK(), key)

	if kvDel.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, kvDel.Err().String())
		return
	}
	// Confirms delete in JSON body.
	kvDelMessage := fmt.Sprintf("Deleted %s", key)
	kvDelResponse := fmt.Sprintf(`{"message":"%s"}`, kvDelMessage)
	fmt.Fprintln(w, kvDelResponse)

}

// JSON validation handling.
func errResponseJSON(w http.ResponseWriter, code int, message string) {
	msg, _ := json.Marshal(CheckResponse{Valid: false, Message: message})
	http.Error(w, string(msg), code)
	w.Header().Set("Content-Type", "application/json")
}

// Since we don't run this program like a CLI, `func main` is empty. 
// Instead, we call handler functions when an HTTP request is received.
func main() {}
