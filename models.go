package main

import (
	"github.com/coopernurse/gorp"
	"time"
)

type Account struct {
	Id           int64
	RiqId        string
	ModifiedDate time.Time
}

type List struct {
	Id        string
	Title     string
	ListType  string    `db:"list_type"`
	TableName string    `db:"table_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SyncResult struct {
	Id             int64
	StartAt        time.Time `db:"start_at"`
	EndAt          time.Time `db:"end_at"`
	AccountsCount  int64     `db:"accounts_count"`
	ListsCount     int64     `db:"lists_count"`
	ListItemsCount int64     `db:"list_items_count"`
	ContactsCount  int64     `db:"contacts_count"`
	UsersCount     int64     `db:"users_count"`
}

type Field struct {
	Id            int64
	ListId        string `db:"list_id"`
	FieldId       string `db:"field_id"`
	Name          string
	IsMultiselect bool      `db:"is_multiselect"`
	IsEditable    bool      `db:"is_editable"`
	DataType      string    `db:"data_type"`
	ColumnName    string    `db:"column_name"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type StatusValue struct {
	Id            int64
	FieldId       int64  `db:"fields_field_id"`
	StatusValueId int64  `db:"status_value_id"`
	DisplayName   string `db:"display_name"`
}

func (l *List) PreInsert(s gorp.SqlExecutor) error {
	l.CreatedAt = time.Now().UTC()
	l.UpdatedAt = time.Now().UTC()
	return nil
}

func (l *List) PreUpdate(s gorp.SqlExecutor) error {
	l.UpdatedAt = time.Now().UTC()
	return nil
}

func (f *Field) PreInsert(s gorp.SqlExecutor) error {
	f.CreatedAt = time.Now().UTC()
	f.UpdatedAt = time.Now().UTC()
	return nil
}

func (f *Field) PreUpdate(s gorp.SqlExecutor) error {
	f.UpdatedAt = time.Now().UTC()
	return nil
}
