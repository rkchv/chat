package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/rkchv/chat/lib/db"

	domain "github.com/rkchv/chat/internal/domain/chat"
	"github.com/rkchv/chat/internal/repository"
)

const (
	createdColumn     = "created_at"
	idColumn          = "id"
	usersChatIdColumn = "chat_id"
	usersUserIdColumn = "user_id"
)

var _ repository.Repository = (*repo)(nil)

type repo struct {
	conn db.Client
}

func New(conn db.Client) repository.Repository {
	instance := &repo{conn: conn}

	return instance
}

func (r *repo) Save(ctx context.Context, chat *domain.Chat) error {
	insert, _, err := sq.Insert("chat.chats").
		Columns(createdColumn).
		Values(sq.Expr("now()")).
		Suffix(fmt.Sprintf("RETURNING %s", idColumn)).
		ToSql()
	if err != nil {
		return err
	}

	q := db.Query{Name: "repository.postgres.Save", QueryRaw: insert}

	err = r.conn.DB().QueryRow(ctx, q).Scan(&chat.Id)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	err := r.conn.DB().ReadCommitted(ctx, func(ctx context.Context) error {
		psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		sql, args, err := psql.Delete("chat.chat_users").
			Where(sq.Eq{usersChatIdColumn: id}).
			ToSql()
		if err != nil {
			return err
		}

		q := db.Query{Name: "repository.postgres.Delete/chat_users", QueryRaw: sql}
		_, err = r.conn.DB().Exec(ctx, q, args...)
		if err != nil {
			return err
		}

		sql, args, err = psql.Delete("chat.chats").
			Where(sq.Eq{idColumn: id}).
			ToSql()
		if err != nil {
			return err
		}

		q = db.Query{Name: "repository.postgres.Delete/chats", QueryRaw: sql}
		_, err = r.conn.DB().Exec(ctx, q, args...)

		return err
	})

	return err
}

func (r *repo) Get(ctx context.Context, chatId int64) (*domain.Chat, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select(idColumn).
		From("chat.chats").
		Where(sq.Eq{idColumn: chatId}).
		ToSql()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrChatNotFound
		}
		return nil, err
	}

	var chat domain.Chat
	err = r.conn.DB().QueryRow(ctx, db.Query{Name: "repository.postgres.Get", QueryRaw: sql}, args...).Scan(&chat.Id)
	if err != nil {
		return nil, err
	}

	sql, args, err = psql.Select(usersUserIdColumn).
		From("chat.chat_users").
		Where(sq.Eq{usersChatIdColumn: chatId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.conn.DB().Query(ctx, db.Query{Name: "repository.postgres.Get", QueryRaw: sql}, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIds, err := pgx.CollectRows(rows, pgx.RowToStructByName[int64])
	if err != nil {
		return nil, err
	}

	chat.UserIds = userIds
	return &chat, nil
}

func (r *repo) Update(ctx context.Context, chat *domain.Chat) error {
	err := r.conn.DB().ReadCommitted(ctx, func(ctx context.Context) error {
		psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		sql, args, err := psql.Delete("chat.chat_users").
			Where(sq.Eq{usersChatIdColumn: chat.Id}).
			ToSql()
		if err != nil {
			return err
		}

		q := db.Query{Name: "repository.postgres.Update/chat_users", QueryRaw: sql}
		_, err = r.conn.DB().Exec(ctx, q, args...)
		if err != nil {
			return err
		}

		insertQuery := psql.Insert("chat.chat_users").Columns(usersChatIdColumn, usersUserIdColumn)

		for _, userId := range chat.UserIds {
			insertQuery = insertQuery.Values(chat.Id, userId)
		}

		query, args, err := insertQuery.ToSql()
		if err != nil {
			return err
		}

		q.QueryRaw = query
		_, err = r.conn.DB().Exec(ctx, q, args...)

		return err
	})

	return err
}
