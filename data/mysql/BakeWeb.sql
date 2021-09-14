#创建数据库
use bakeweb;

#创建配置表
drop table if exists t_config;

create table t_config(
    configId integer not null auto_increment,
    name varchar(128) not null,
    value varchar(10240) not null,
    createTime timestamp not null default CURRENT_TIMESTAMP,
    modifyTime timestamp not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
    primary key( configId )
)engine=innodb default charset=utf8mb4 auto_increment = 10001;

alter table t_config add index nameIndex(name);