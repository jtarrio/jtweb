package genericsql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/engine"
)

const commentIdMask int64 = 0x5555555555555555

func int64ToCommentId(id int64) comments.CommentId {
	return comments.CommentId(strconv.FormatInt(id^commentIdMask, 36))
}

func commentIdToInt64(id comments.CommentId) (int64, error) {
	out, err := strconv.ParseInt(string(id), 36, 64)
	if err != nil {
		return 0, err
	}
	return out ^ commentIdMask, nil
}

type notFoundError struct {
	postId comments.PostId
}

func notFound(postId comments.PostId) error {
	return &notFoundError{postId: postId}
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("unknown post [%s]", e.postId)
}

type GenericSqlEngine struct {
	db *sql.DB
}

func NewGenericSqlEngine(db *sql.DB) engine.Engine {
	return &GenericSqlEngine{db: db}
}

func (e *GenericSqlEngine) GetConfig(postId comments.PostId) (*engine.Config, error) {
	tx, err := e.startTx()
	if err != nil {
		return nil, err
	}
	cfg, err := tx.getConfig(postId)
	tx.finishTx(err)
	return cfg, err
}

func (e *GenericSqlEngine) SetConfig(newConfig, oldConfig *engine.Config) error {
	tx, err := e.startTx()
	if err != nil {
		return err
	}
	current, err := tx.getConfig(newConfig.PostId)
	if _, ok := err.(*notFoundError); !ok {
		return tx.finishTx(err)
	}
	if (current == nil) != (oldConfig != nil) || current != oldConfig {
		return tx.finishTx(fmt.Errorf("old configuration is different from expected for post [%s]", newConfig.PostId))
	}
	return tx.finishTx(tx.setConfig(newConfig))
}

func (e *GenericSqlEngine) Load(postId comments.PostId) ([]engine.Comment, error) {
	tx, err := e.startTx()
	if err != nil {
		return nil, err
	}
	cmts, err := tx.load(postId)
	tx.finishTx(err)
	return cmts, err
}

func (e *GenericSqlEngine) Add(newComment *engine.NewComment) (*engine.Comment, error) {
	tx, err := e.startTx()
	if err != nil {
		return nil, err
	}
	cmt, err := tx.add(newComment)
	tx.finishTx(err)
	return cmt, err
}

type etx struct {
	tx *sql.Tx
}

func (e *GenericSqlEngine) startTx() (*etx, error) {
	tx, err := e.db.BeginTx(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return &etx{tx: tx}, nil
}

func (tx *etx) finishTx(err error) error {
	if err != nil {
		tx.tx.Rollback()
		return err
	}
	return tx.tx.Commit()
}

func (tx *etx) getConfig(postId comments.PostId) (*engine.Config, error) {
	stmt, err := tx.tx.Prepare(`SELECT PostId, State, StateOverride FROM Posts WHERE PostId = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rowPostId string
	var rowState int
	var rowStateOverride sql.NullInt16
	err = stmt.QueryRow(string(postId)).Scan(&rowPostId, &rowState, &rowStateOverride)
	if err == sql.ErrNoRows {
		return nil, notFound(postId)
	}
	cfg := &engine.Config{
		PostId: comments.PostId(rowPostId),
		State:  engine.CommentState(rowState),
	}
	if rowStateOverride.Valid {
		cfg.State = engine.CommentState(rowStateOverride.Int16)
	}
	return cfg, nil
}

func (tx *etx) setConfig(newConfig *engine.Config) error {
	stmt, err := tx.tx.Prepare(`UPDATE Posts SET StateOverride = ? WHERE PostId = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newConfig.State, newConfig.PostId)
	return err
}

func (tx *etx) load(postId comments.PostId) ([]engine.Comment, error) {
	stmt, err := tx.tx.Prepare(`SELECT PostId, CommentId, Author, Date, Text FROM Comments WHERE PostId = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []engine.Comment{}
	for rows.Next() {
		var rowPostId string
		var rowCommentId int64
		var rowAuthor string
		var rowWhen time.Time
		var rowText string
		if err := rows.Scan(&rowPostId, &rowCommentId, &rowAuthor, &rowWhen, &rowText); err != nil {
			return nil, err
		}
		out = append(out, engine.Comment{
			PostId:    comments.PostId(rowPostId),
			CommentId: int64ToCommentId(rowCommentId),
			Author:    rowAuthor,
			When:      rowWhen,
			Text:      comments.Markdown(rowText),
		})
	}
	return out, nil
}

func (tx *etx) add(newComment *engine.NewComment) (*engine.Comment, error) {
	stmt, err := tx.tx.Prepare(`INSERT INTO Comments (PostId, Author, Date, Text) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(newComment.PostId, newComment.Author, newComment.When, newComment.Text)
	if err != nil {
		return nil, err
	}
	newId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &engine.Comment{
		PostId:    newComment.PostId,
		CommentId: int64ToCommentId(newId),
		Author:    newComment.Author,
		When:      newComment.When,
		Text:      newComment.Text,
	}, nil
}
