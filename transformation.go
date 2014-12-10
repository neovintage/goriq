package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	RiqApi = "https://api.relateiq.com/v2"
)

type RiqAccount struct {
	Id           string
	ModifiedDate int64
	Name         string
	FieldValues  map[string]interface{}
}

type RiqContact struct {
	Id           string
	ModifiedDate int64
	Properties   map[string]interface{}
}

type RiqList struct {
	Id           string
	ModifiedDate int64
	Title        string
	ListType     string
	Fields       []RiqListField
}

type RiqListField struct {
	Id            string
	Name          string
	ListOptions   []map[string]interface{}
	IsMultiSelect bool
	IsEditable    bool
	DataType      string
}

type RiqListItem struct {
	Id           string
	ModifiedDate int64
	CreatedDate  int64
	ListId       string
	AccountId    string
	ContactIds   []string
	Name         string
	// This is friggin gross.  thanks relateiq
	FieldValues map[string][]map[string]interface{}
}

type RiqUser struct {
	Id    string
	Name  string
	Email string
}

func (a Auth) buildUrl(endpoint string, vals url.Values) *url.URL {
	uri, _ := url.Parse(RiqApi + endpoint)
	uri.User = url.UserPassword(a.Username, a.Password)
	if vals != nil {
		uri.RawQuery = vals.Encode()
	}
	return uri
}

func getRequest(uri url.URL) *http.Response {
	log.Println(uri.String())
	resp, err := http.Get(uri.String())
	if err != nil {
		assert(err)
	}
	return resp
}

func getResource(endpoint string, vals url.Values) interface{} {

	return nil
}

func getAccount(accId string) RiqAccount {
	uri := config.Auth.buildUrl("/accounts/"+accId, nil)
	resp := getRequest(*uri)
	var acc RiqAccount
	json.NewDecoder(resp.Body).Decode(&acc)
	resp.Body.Close()
	return acc
}

func getAccounts(fn func(RiqAccount)) {
	uri := config.Auth.buildUrl("/accounts", nil)
	resp := getRequest(*uri)

	// Don't need the accounts object.  we can just accept a function and then invoke it
	var data map[string][]json.RawMessage
	accounts := []RiqAccount{}
	json.NewDecoder(resp.Body).Decode(&data)
	resp.Body.Close()
	for _, obj := range data["objects"] {
		var acc RiqAccount
		_ = json.Unmarshal(obj, &acc)
		accounts = append(accounts, acc)
	}
}

func (a Auth) getLists(fn func(RiqList)) {
	v := url.Values{}
	v.Set("_limit", "50")
	uri := config.Auth.buildUrl("/lists", v)
	uri.RawQuery = v.Encode()

	resp := getRequest(*uri)
	var data map[string][]json.RawMessage
	json.NewDecoder(resp.Body).Decode(&data)
	for _, obj := range data["objects"] {
		var list RiqList
		_ = json.Unmarshal(obj, &list)
		fn(list)
	}
}

func (a Auth) getListItems(listId string, fn func(RiqListItem)) {
	// Only going to do the Account management one for now since
	// we need to test this out
	if listId != "53125b6ce4b0b6fbc0fd7ef9" {
		return
	}

	v := url.Values{}
	v.Set("_limit", "50")
	var lastRunEndTime time.Time
	err := trans.SelectOne(&lastRunEndTime, "select end_at from sync_results order by end_at desc limit 1")
	if err != nil {
		lastRunEndTime = time.Date(2000, time.January, 1, 1, 0, 0, 0, time.UTC)
	}
	v.Set("modifiedDate", strconv.FormatInt(lastRunEndTime.Unix()*1000, 10))
	start := 0
	v.Set("_start", strconv.Itoa(start))

	for {
		uri := config.Auth.buildUrl("/lists/"+listId+"/listitems", v)
		resp := getRequest(*uri)
		//b, _ := httputil.DumpResponse(resp, true)
		//log.Println(string(b))
		var data map[string][]json.RawMessage
		json.NewDecoder(resp.Body).Decode(&data)
		for _, obj := range data["objects"] {
			var listItem RiqListItem
			json.Unmarshal(obj, &listItem)
			fn(listItem)
		}

		//if len(data["objects"]) < 50 {
		break
		//} else {
		//start += 50
		//v.Set("_start", string(start))
		//}
	}
}

func (a Auth) getContacts(ids []string) {
	v := url.Values{}
	if ids != nil && len(ids) > 0 {
		v.Set("_ids", strings.Join(ids, ","))
		v.Set("_limit", strconv.Itoa(len(ids)))
	}
	uri := config.Auth.buildUrl("/contacts", v)
	resp := getRequest(*uri)
	var data map[string][]json.RawMessage
	json.NewDecoder(resp.Body).Decode(&data)
	for _, obj := range data["objects"] {
		var contact RiqContact
		json.Unmarshal(obj, &contact)
	}
}

func (a Auth) getUser(id string) RiqUser {
	uri := config.Auth.buildUrl("/users/"+id, nil)
	resp := getRequest(*uri)
	var user RiqUser
	json.NewDecoder(resp.Body).Decode(&user)
	resp.Body.Close()
	return user
}

func doTransformation() {
	initDB()
	config.Auth.getLists(saveList)
	var listIds []string
	dbmap.Select(&listIds, "select id from lists")
	for _, listId := range listIds {
		config.Auth.getListItems(listId, saveListItems)
	}
	assert(trans.Commit())
	dbmap.Db.Close()
}
