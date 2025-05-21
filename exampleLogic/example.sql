-- auto-generated definition
create table sales_order
(
    id             int auto_increment
        primary key,
    order_id       varchar(100) default ''                not null comment 'SO号',
    status         tinyint      default 0                 not null comment '订单状态',
    customer_name  varchar(30)  default ''                not null comment '客户姓名',
    address        varchar(500) default ''                not null comment '收货地址',
    created_at     datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at     datetime     default CURRENT_TIMESTAMP null comment '修改时间',
    deleted_at     datetime                               null,
    create_by      int unsigned default 0                 not null comment '创建人id',
    create_by_name varchar(20)  default ''                not null,
    update_by      int unsigned default 0                 not null comment '修改人id',
    update_by_name varchar(20)  default ''                not null
)
    comment '销售订单';


-- auto-generated definition
create table sales_order_detail
(
    id             int auto_increment
        primary key,
    order_id       varchar(100)   default ''                not null comment 'order_id',
    sku_code       varchar(100)   default ''                not null comment 'sku_code',
    product_name   varchar(100)   default ''                not null comment 'sku名称',
    brand_name     varchar(100)   default ''                not null comment '品牌',
    model_type     varchar(100)   default ''                not null comment '型号',
    order_quantity decimal(10, 2) default 0.00              not null comment '订单数量',
    created_at     datetime       default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at     datetime       default CURRENT_TIMESTAMP null comment '修改时间',
    deleted_at     datetime                                 null,
    create_by      int unsigned   default 0                 not null comment '创建人id',
    create_by_name varchar(20)    default ''                not null,
    update_by      int unsigned   default 0                 not null comment '修改人id',
    update_by_name varchar(20)    default ''                not null
)
    comment '销售订单明细';

