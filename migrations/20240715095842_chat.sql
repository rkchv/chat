-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA chat;
CREATE TABLE chat.chats
(
    id integer primary key,
    created_at timestamp(0) default CURRENT_TIMESTAMP
);
CREATE TABLE chat.chat_users
(
    chat_id int references chat.chats(id),
    user_id bigint not null,
    user_name text not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA chat CASCADE;
-- +goose StatementEnd
