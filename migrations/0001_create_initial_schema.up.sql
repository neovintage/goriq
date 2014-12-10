CREATE TABLE sync_results (
  id                bigserial primary key,
  start_at          timestamp without time zone,
  end_at            timestamp without time zone,
  accounts_count    integer,
  lists_count       smallint,
  list_items_count  integer,
  contacts_count    integer,
  users_count       integer
);

CREATE TABLE lists (
  id                character varying(255),
  title             character varying(255),
  list_type         character varying(255),
  table_name        character varying(255),
  created_at        timestamp without time zone,
  updated_at        timestamp without time zone
);

CREATE TABLE fields (
  id                bigserial primary key,
  list_id           character varying(255),
  field_id          character varying(255),
  name              character varying(255),
  is_multiselect    boolean,
  is_editable       boolean,
  data_type         character varying(255),
  column_name       character varying(255),
  created_at        timestamp without time zone,
  updated_at        timestamp without time zone
);

CREATE TABLE status_values (
  id                bigserial primary key,
  -- This is actually fields.id not fields.field_id
  fields_field_id   integer,
  status_value_id   integer,
  display_name      character varying(255)
);

CREATE TABLE accounts (
  id                bigserial primary key,
  riq_id            character varying(255),
  modified_date     timestamp without time zone,
  name              character varying(255)
);

CREATE TABLE contacts (
  id                bigserial primary key,
  riq_id            character varying(255)
);
