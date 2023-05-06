create database if not exists logindb;
use logindb;
create table if not exists LoginInfo (
ID VARCHAR(20) NOT NULL,
PW VARCHAR(20) NOT NULL,
PRIMARY KEY(ID)
);
insert into LoginInfo values ("root", "1234");
insert into LoginInfo values ("guest", 1234);