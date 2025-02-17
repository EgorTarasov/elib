create table if not exists documents(
    id integer not null unique,
    name text not null,
    page_cnt int not null,
    folder_id int,
    foreign key (folder_id) references folders(id)
);


create table if not exists folders(
    id integer not null unique,
    name text,
    parent int,
    foreign key (parent) references folders(id)
);
-- select * from folders;
-- select * from folders;
-- drop table documents;