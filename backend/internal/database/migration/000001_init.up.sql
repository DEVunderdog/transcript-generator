create table "encryption_keys" (
    id serial primary key,
    public_key text not null,
    private_key bytea not null,
    is_active bool default false,
    purpose varchar(100) not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

create table "users" (
    id serial primary key,
    email varchar (255) not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

create table "api_keys" (
    id serial primary key,
    user_id int unique not null,
    credential bytea not null,
    signature bytea not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

create table "file_registry" (
    id serial primary key,
    user_id int not null,
    file_name varchar(100) not null,
    object_key varchar(150) null,
    lock_status bool not null,
    upload_status varchar(20) not null,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz default current_timestamp
);

alter table "api_keys" add constraint "fk_user_api_keys" foreign key ("user_id") references "users" ("id") on update cascade on delete restrict;

alter table "file_registry" add constraint "fk_user_file_registry" foreign key ("user_id") references "users" ("id") on update cascade on delete restrict;

create unique index idx_unique_filename on "file_registry" ("file_name", "user_id");

create index idx_email on "users" ("email");

create index idx_keys_active on "encryption_keys" ("is_active");

create index idx_keys_purpose on "encryption_keys" ("purpose");

create index idx_file_registry_user_id on "file_registry" ("user_id");

create index idx_api_keys_user_id on "api_keys" ("user_id");

create index idx_file_registry on "file_registry" ("lock_status", "upload_status");