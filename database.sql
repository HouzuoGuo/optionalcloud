drop database optionalcloud;
create database optionalcloud;
use optionalcloud;

create table users(
    username varchar(10) primary key,
    password varchar(10)
);
insert into users values('test', 'a');
