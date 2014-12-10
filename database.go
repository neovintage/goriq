package main

import (
	"bitbucket.org/pkg/inflect"
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	dbmap *gorp.DbMap
	trans *gorp.Transaction
)

type DatabaseConfig struct {
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
	Options    map[string]string
	Migrations string
}

type CreateTableView struct {
	TableName     string `db:"table_name"`
	ColumnName    string `db:"column_name"`
	DataType      string `db:"data_type"`
	IsMultiselect bool   `db:"is_multiselect"`
}

func (dc DatabaseConfig) connectionString() string {
	u := url.URL{
		Scheme: "postgres",
	}

	u.Host = dc.Host
	if dc.Port != "" {
		u.Host = u.Host + ":" + dc.Port
	}
	u.Path = dc.Database
	u.User = url.UserPassword(dc.Username, dc.Password)

	v := url.Values{}
	for key, value := range dc.Options {
		v.Set(key, value)
	}
	u.RawQuery = v.Encode()

	return u.String()
}

func initDB() {
	db, err := sql.Open("postgres", config.Database.connectionString())
	assert(err)
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	// Table - ORM mappings
	dbmap.AddTableWithName(List{}, "lists")
	dbmap.AddTableWithName(Field{}, "fields").SetKeys(true, "Id")
	dbmap.AddTableWithName(StatusValue{}, "status_values").SetKeys(true, "Id")
	dbmap.AddTableWithName(SyncResult{}, "sync_results").SetKeys(true, "Id")

	trans, err = dbmap.Begin()
	assert(err)
}

func tableize(s string) string {
	// this block is solely for the opportunity list.  lame
	re := regexp.MustCompile("([0-9]+\\. )")
	s = re.ReplaceAllString(s, "")

	s = inflect.Underscore(s)
	re = regexp.MustCompile("\\W")
	s = re.ReplaceAllString(s, "")

	// Special Cases - this is a mega pos
	switch s {
	case "30_days_events":
		s = "thirty_days_events"
	case "7_days_events":
		s = "seven_day_events"
	}

	return s
}

func (v *CreateTableView) translateRiqDataType() string {
	switch v.DataType {
	case "User", "Contact", "Text":
		return "character varying(255)"
	case "Date":
		return "date"
	case "DateTime":
		return "timestamp without time zone"
	case "Numeric":
		return "numeric(20, 5)"
	case "List":
		if !v.IsMultiselect {
			return "character varying(255)"
		}
	}
	return ""
}

func createTable(riqListId string, listName string) {
	// TODO really need to check for dim and snp tables not just one
	var exists bool
	trans.SelectOne(
		&exists,
		fmt.Sprintf(`select exists(
            select 1
            from pg_catalog.pg_class c
            JOIN   pg_catalog.pg_namespace n ON n.oid = c.relnamespace
            WHERE  n.nspname = 'public'
            AND    c.relname = '%s'
        ) as exists`, listName+"_dim"),
	)

	var views []CreateTableView

	if exists {
		// TODO Handle making updates to the table
		log.Println("Table Exists:", listName+"_dim")
	} else {
		// Create a new table
		var statements []string
		trans.Select(&views, "select table_name, column_name, data_type, is_multiselect from lists inner join fields on lists.id = fields.list_id and fields.column_name != 'status' where lists.id = $1", riqListId)
		for _, view := range views {
			datatype := view.translateRiqDataType()
			statements = append(statements, view.ColumnName+" "+datatype)
		}
		statements = append(statements, `
                    list_item_id character varying(255),
                    name character varying(255),
                    account_id character varying(255),
                    list_id character varying(255),
                    created_at timestamp without time zone,
                    modified_at timestamp without time zone`)
		// Create the dimension table and then the snapshot
		stmt := "create table " + listName + "_dim ( " + strings.Join(statements, ",\n") + ")"
		log.Println(stmt)
		_, err := trans.Exec(stmt)
		log.Println("transaction problme:", err)

		var names []string
		trans.Select(&names, "select display_name from fields inner join status_values s on fields.id = s.fields_field_id where fields.list_id = $1 and fields.column_name = 'status'", riqListId)
		log.Println("names:", names)
		for i, s := range names {
			names[i] = tableize(s) + " timestamp without time zone"
		}
		names = append(names, `list_id character varying(255), account_id character varying(255), list_item_id character varying(255)`)
		stmt = "create table " + listName + "_snp (" + strings.Join(names, ",\n") + ")"
		log.Println(stmt)
		_, err = trans.Exec(stmt)
		log.Println("transaction problem:", err)
	}

}

// This will get the structure of the system into the tables.
// saveList does not actually add any of the list items into the database
func saveList(riqList RiqList) {
	var list List
	err := trans.SelectOne(&list, "select * from lists where id = $1", riqList.Id)
	if err != nil {
		// Means that we couldn't find the list
		list = List{
			Id:        riqList.Id,
			Title:     riqList.Title,
			ListType:  riqList.ListType,
			TableName: tableize(riqList.Title),
		}
		err = trans.Insert(&list)
		assert(err)
	} else {
		// Update all of the stuff relative to the list
		list.Title = riqList.Title
		list.ListType = riqList.ListType
		trans.Update(&list)
	}

	for _, riqField := range riqList.Fields {
		var field Field
		err = dbmap.SelectOne(&field, "select * from fields where field_id = $1 and list_id = $2", riqField.Id, list.Id)
		if err == nil {
			// TODO update all of the field options.  We do not handle this case right now
		}

		if err != nil {
			field = Field{
				ListId:        list.Id,
				FieldId:       riqField.Id,
				Name:          riqField.Name,
				IsMultiselect: riqField.IsMultiSelect,
				IsEditable:    riqField.IsEditable,
				DataType:      riqField.DataType,
				ColumnName:    tableize(riqField.Name),
			}
			err = trans.Insert(&field)
			if field.DataType == "List" {
				for _, opt := range riqField.ListOptions {
					status_value, _ := strconv.ParseInt(opt["id"].(string), 10, 64)
					statusValue := StatusValue{
						FieldId:       field.Id,
						StatusValueId: status_value,
						DisplayName:   opt["display"].(string),
					}
					err = trans.Insert(&statusValue)
					if err != nil {
						log.Println("Error inserting status value:", err)
					}
				}
			}
		}
	}

	createTable(riqList.Id, list.TableName)
}

func saveListItems(riqListItem RiqListItem) {
	// TODO finally get list items need to populate them

}
